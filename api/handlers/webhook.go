package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/immortalvibes/api/shippo"
	"github.com/immortalvibes/api/store"
	stripe "github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/webhook"
)

// ShipperClient rate-shops and purchases shipping labels.
type ShipperClient interface {
	RateShop(ctx context.Context, to shippo.Address) (rateID string, err error)
	BuyLabel(ctx context.Context, rateID string) (trackingNumber, carrier, labelURL string, err error)
}

// WebhookKV is the cart-clearing subset of CartKV.
type WebhookKV interface {
	DeleteCart(ctx context.Context, token string) error
}

// WebhookStock decrements stock after a purchase.
type WebhookStock interface {
	DecrementStock(ctx context.Context, productID string, qty int) error
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
	SendShippingLabel(ctx context.Context, ownerEmail, orderID, labelURL, trackingNum, carrier string) error
	SendTrackingUpdate(ctx context.Context, customerEmail, orderID, trackingNum, carrier string) error
	SendShippingFailure(ctx context.Context, ownerEmail, orderID, customerEmail, shippingAddr, errMsg string) error
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
) *WebhookHandler {
	return &WebhookHandler{
		secret:     secret,
		kv:         kv,
		stock:      stock,
		db:         db,
		emailer:    emailer,
		shipper:    shipper,
		ownerEmail: ownerEmail,
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
		// Acknowledge to prevent Stripe retries for orders we can't find.
		w.WriteHeader(http.StatusOK)
		return
	}

	if err := h.db.UpdateOrderStatus(r.Context(), order.ID, "complete"); err != nil {
		log.Printf("webhook: UpdateOrderStatus(%s): %v", order.ID, err)
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

	// Shipping runs in a goroutine — Stripe gets a fast 200 immediately.
	orderCopy := *order
	go h.processShipping(&orderCopy)

	w.WriteHeader(http.StatusOK)
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
	}

	rateID, err := h.shipper.RateShop(context.Background(), toAddr)
	if err != nil {
		log.Printf("webhook: RateShop(%s): %v", order.ID, err)
		h.notifyShippingFailure(order, err.Error())
		return
	}

	trackingNum, carrier, labelURL, err := h.shipper.BuyLabel(context.Background(), rateID)
	if err != nil {
		log.Printf("webhook: BuyLabel(%s): %v", order.ID, err)
		h.notifyShippingFailure(order, err.Error())
		return
	}

	if err := h.db.UpdateOrderShipping(context.Background(), order.ID, trackingNum, carrier, labelURL); err != nil {
		log.Printf("webhook: UpdateOrderShipping(%s): %v", order.ID, err)
	}

	if err := h.emailer.SendShippingLabel(context.Background(), h.ownerEmail, order.ID, labelURL, trackingNum, carrier); err != nil {
		log.Printf("webhook: SendShippingLabel: %v", err)
	}

	if err := h.emailer.SendTrackingUpdate(context.Background(), order.Email, order.ID, trackingNum, carrier); err != nil {
		log.Printf("webhook: SendTrackingUpdate: %v", err)
	}
}

func (h *WebhookHandler) notifyShippingFailure(order *store.OrderRow, errMsg string) {
	addr := fmt.Sprintf("%s\n%s\n%s, %s %s\n%s",
		order.ShippingName, order.Line1, order.City, order.State, order.PostalCode, order.Country)
	if err := h.emailer.SendShippingFailure(context.Background(), h.ownerEmail, order.ID, order.Email, addr, errMsg); err != nil {
		log.Printf("webhook: SendShippingFailure: %v", err)
	}
}
