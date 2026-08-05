package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/immortalvibes/api/models"
	"github.com/immortalvibes/api/shippo"
	"github.com/immortalvibes/api/store"
	stripe "github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/webhook"
)

// ShipperClient rate-shops and purchases shipping labels.
type ShipperClient interface {
	RateShop(ctx context.Context, to shippo.Address) (rateID string, err error)
	BuyLabel(ctx context.Context, rateID string) (trackingNumber, carrier, labelURL, cost string, err error)
}

// WebhookKV is the cart-clearing subset of CartKV.
type WebhookKV interface {
	DeleteCart(ctx context.Context, token string) error
}

// WebhookStock decrements stock after a purchase.
type WebhookStock interface {
	DecrementStock(ctx context.Context, productID, variant string, qty int) error
}

// WebhookOrderStore reads and updates orders for the webhook.
type WebhookOrderStore interface {
	GetOrderByPaymentIntent(ctx context.Context, paymentIntentID string) (*store.OrderRow, error)
	UpdateOrderStatus(ctx context.Context, id, status string) error
	UpdateOrderShipping(ctx context.Context, id, trackingNumber, carrier, labelURL string) error
}

// EmailSender dispatches transactional email.
type EmailSender interface {
	SendOrderConfirmation(ctx context.Context, toEmail, orderID string, totalAmount int64, currency string) error
	SendShippingLabel(ctx context.Context, ownerEmail, orderID, labelURL, trackingNum, carrier, labelCost string, orderTotal int64, currency string, items []models.LineItem) error
	SendTrackingUpdate(ctx context.Context, customerEmail, orderID, trackingNum, carrier string) error
	SendShippingFailure(ctx context.Context, ownerEmail, orderID, customerEmail, shippingAddr, errMsg string) error
}

// OrderNotifier pushes order alerts (e.g. to ntfy) as a backstop for the
// transactional emails, which can fail silently. Best-effort: a notify error
// never fails order processing.
type OrderNotifier interface {
	NotifyOrderSuccess(ctx context.Context, orderID, total, shipTo, items, tracking, carrier string) error
	NotifyOrderFailure(ctx context.Context, orderID, email, stage, errMsg string) error
}

// WebhookHandler handles POST /api/webhooks/stripe.
type WebhookHandler struct {
	secret     string
	kv         WebhookKV
	stock      WebhookStock
	db         WebhookOrderStore
	emailer    EmailSender
	shipper    ShipperClient
	ownerEmail string
	notifier   OrderNotifier
}

// NewWebhookHandler constructs a WebhookHandler.
func NewWebhookHandler(
	secret string,
	kv WebhookKV,
	stock WebhookStock,
	db WebhookOrderStore,
	emailer EmailSender,
	shipper ShipperClient,
	ownerEmail string,
	notifier OrderNotifier,
) *WebhookHandler {
	return &WebhookHandler{
		secret:     secret,
		kv:         kv,
		stock:      stock,
		db:         db,
		emailer:    emailer,
		shipper:    shipper,
		ownerEmail: ownerEmail,
		notifier:   notifier,
	}
}

// HandleWebhook handles POST /api/webhooks/stripe.
func (h *WebhookHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	const maxBodyBytes = 65536
	body, err := io.ReadAll(io.LimitReader(r.Body, maxBodyBytes))
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}

	event, err := webhook.ConstructEventWithOptions(body, r.Header.Get("Stripe-Signature"), h.secret, webhook.ConstructEventOptions{
		IgnoreAPIVersionMismatch: true,
	})
	if err != nil {
		http.Error(w, "invalid signature", http.StatusBadRequest)
		return
	}

	switch event.Type {
	case "payment_intent.succeeded":
		h.handlePaymentIntentSucceeded(w, r, event)
	default:
		w.WriteHeader(http.StatusOK)
	}
}

type paymentIntentObject struct {
	ID       string            `json:"id"`
	Amount   int64             `json:"amount"`
	Currency string            `json:"currency"`
	Metadata map[string]string `json:"metadata"`
}

func (h *WebhookHandler) handlePaymentIntentSucceeded(w http.ResponseWriter, r *http.Request, event stripe.Event) {
	var pi paymentIntentObject
	if err := json.Unmarshal(event.Data.Raw, &pi); err != nil {
		http.Error(w, "failed to parse payment intent", http.StatusBadRequest)
		return
	}

	cartToken := pi.Metadata["cart_token"]
	email := pi.Metadata["email"]

	order, err := h.db.GetOrderByPaymentIntent(r.Context(), pi.ID)
	if err != nil {
		log.Printf("webhook: GetOrderByPaymentIntent(%s): %v", pi.ID, err)
		// A payment succeeded but we can't match it to an order — the owner
		// needs to know. Acknowledge to prevent Stripe retries regardless.
		h.pushFailure(pi.ID, email, "order-lookup", err.Error())
		w.WriteHeader(http.StatusOK)
		return
	}

	// Idempotency: Stripe may resend this event (retries / manual resends).
	// If the order is already shipped (label bought, tracking set), there is
	// nothing more to do — ack and return so we never buy a second label.
	if order.TrackingNumber != "" {
		log.Printf("webhook: order %s already shipped (tracking %s) — skipping reprocess", order.ID, order.TrackingNumber)
		w.WriteHeader(http.StatusOK)
		return
	}

	// Decrement stock exactly once, on the first transition to complete.
	// Guarded on prior status so Stripe resends never double-decrement.
	wasComplete := order.Status == "complete"

	if err := h.db.UpdateOrderStatus(r.Context(), order.ID, "complete"); err != nil {
		log.Printf("webhook: UpdateOrderStatus(%s): %v", order.ID, err)
	}

	if !wasComplete {
		h.decrementStock(r.Context(), order)
	}

	if cartToken != "" {
		if err := h.kv.DeleteCart(r.Context(), cartToken); err != nil {
			log.Printf("webhook: DeleteCart(%s): %v", cartToken, err)
		}
	}

	if email != "" {
		if err := h.emailer.SendOrderConfirmation(r.Context(), email, order.ID, pi.Amount, pi.Currency); err != nil {
			log.Printf("webhook: SendOrderConfirmation(%s): %v", email, err)
		}
	}

	// Shipping runs synchronously before we ack Stripe. This keeps label
	// purchase from being lost to a killed goroutine (e.g. machine stop), and
	// any transient failure is reported to the owner via SendShippingFailure.
	// We still return 200: the payment is captured, so we don't want Stripe to
	// retry the whole event indefinitely — the idempotency guard above makes a
	// resend safe if the owner needs to re-trigger.
	h.processShipping(order)

	w.WriteHeader(http.StatusOK)
}

// decrementStock subtracts purchased quantities from product stock, one line
// item at a time. Stock errors are logged but never fail the webhook — a paid
// order must still ship; an out-of-sync stock count is the lesser problem.
func (h *WebhookHandler) decrementStock(ctx context.Context, order *store.OrderRow) {
	for _, li := range order.LineItems {
		if li.ProductID == "" || li.Quantity <= 0 {
			continue
		}
		if err := h.stock.DecrementStock(ctx, li.ProductID, li.Size, li.Quantity); err != nil {
			log.Printf("webhook: DecrementStock(order=%s product=%s variant=%q qty=%d): %v",
				order.ID, li.ProductID, li.Size, li.Quantity, err)
		}
	}
}

func (h *WebhookHandler) processShipping(order *store.OrderRow) {
	toAddr := shippo.Address{
		Name:    order.ShippingName,
		Street1: order.Line1,
		Street2: order.Line2,
		City:    order.City,
		State:   order.State,
		Zip:     order.PostalCode,
		Country: order.Country,
		Email:   order.Email,
	}

	rateID, err := h.shipper.RateShop(context.Background(), toAddr)
	if err != nil {
		log.Printf("webhook: RateShop(%s): %v", order.ID, err)
		h.notifyShippingFailure(order, err.Error())
		return
	}

	trackingNum, carrier, labelURL, labelCost, err := h.shipper.BuyLabel(context.Background(), rateID)
	if err != nil {
		log.Printf("webhook: BuyLabel(%s): %v", order.ID, err)
		h.notifyShippingFailure(order, err.Error())
		return
	}

	if err := h.db.UpdateOrderShipping(context.Background(), order.ID, trackingNum, carrier, labelURL); err != nil {
		log.Printf("webhook: UpdateOrderShipping(%s): %v", order.ID, err)
	}

	ownerNotified := true
	if err := h.emailer.SendShippingLabel(context.Background(), h.ownerEmail, order.ID, labelURL, trackingNum, carrier, labelCost, order.TotalAmount, order.Currency, order.LineItems); err != nil {
		log.Printf("webhook: SendShippingLabel: %v", err)
		// The owner's manifest email failed — exactly the "owner didn't get
		// notified" case. Alert loudly instead of a quiet success ping.
		ownerNotified = false
		h.pushFailure(order.ID, order.Email, "owner-email", err.Error())
	}

	if err := h.emailer.SendTrackingUpdate(context.Background(), order.Email, order.ID, trackingNum, carrier); err != nil {
		log.Printf("webhook: SendTrackingUpdate: %v", err)
	}

	// Backstop: push a quiet success ping so the owner (and Eric) know an order
	// landed and shipped even if the emails silently failed. Skipped when the
	// owner email failed above — that already fired an urgent alert.
	if ownerNotified {
		h.pushSuccess(order, trackingNum, carrier)
	}
}

func (h *WebhookHandler) notifyShippingFailure(order *store.OrderRow, errMsg string) {
	addr := fmt.Sprintf("%s\n%s\n%s, %s %s\n%s",
		order.ShippingName, order.Line1, order.City, order.State, order.PostalCode, order.Country)
	if err := h.emailer.SendShippingFailure(context.Background(), h.ownerEmail, order.ID, order.Email, addr, errMsg); err != nil {
		log.Printf("webhook: SendShippingFailure: %v", err)
	}
	h.pushFailure(order.ID, order.Email, "shipping", errMsg)
}

// pushSuccess sends the quiet order-received/shipped ntfy. Best-effort.
func (h *WebhookHandler) pushSuccess(order *store.OrderRow, trackingNum, carrier string) {
	if h.notifier == nil {
		return
	}
	total := fmt.Sprintf("$%.2f %s", float64(order.TotalAmount)/100, strings.ToUpper(order.Currency))
	shipTo := strings.TrimSpace(fmt.Sprintf("%s, %s %s", order.City, order.State, order.Country))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := h.notifier.NotifyOrderSuccess(ctx, order.ID, total, shipTo, summarizeItems(order.LineItems), trackingNum, carrier); err != nil {
		log.Printf("webhook: NotifyOrderSuccess(%s): %v", order.ID, err)
	}
}

// pushFailure sends the urgent order-problem ntfy. Best-effort.
func (h *WebhookHandler) pushFailure(orderID, email, stage, errMsg string) {
	if h.notifier == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := h.notifier.NotifyOrderFailure(ctx, orderID, email, stage, errMsg); err != nil {
		log.Printf("webhook: NotifyOrderFailure(%s): %v", orderID, err)
	}
}

// summarizeItems renders a one-line item summary for a push notification,
// e.g. "2x Immortal Light Sweatpants (L), 1x WR Tee".
func summarizeItems(items []models.LineItem) string {
	if len(items) == 0 {
		return "(no manifest)"
	}
	parts := make([]string, 0, len(items))
	for _, li := range items {
		name := li.Name
		if name == "" {
			name = li.ProductID
		}
		if li.Size != "" {
			name = fmt.Sprintf("%s (%s)", name, li.Size)
		}
		parts = append(parts, fmt.Sprintf("%dx %s", li.Quantity, name))
	}
	return strings.Join(parts, ", ")
}
