package handlers

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/immortalvibes/api/store"
	stripe "github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/webhook"
)

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
}

// EmailSender dispatches transactional email.
type EmailSender interface {
	SendOrderConfirmation(ctx context.Context, toEmail, orderID string, totalAmount int64, currency string) error
}

// WebhookHandler handles POST /api/webhooks/stripe.
type WebhookHandler struct {
	secret  string
	kv      WebhookKV
	stock   WebhookStock
	db      WebhookOrderStore
	emailer EmailSender
}

// NewWebhookHandler constructs a WebhookHandler.
func NewWebhookHandler(
	secret string,
	kv WebhookKV,
	stock WebhookStock,
	db WebhookOrderStore,
	emailer EmailSender,
) *WebhookHandler {
	return &WebhookHandler{
		secret:  secret,
		kv:      kv,
		stock:   stock,
		db:      db,
		emailer: emailer,
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
		// Acknowledge unknown events.
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

	// Look up the pending order.
	order, err := h.db.GetOrderByPaymentIntent(r.Context(), pi.ID)
	if err != nil {
		log.Printf("webhook: GetOrderByPaymentIntent(%s): %v", pi.ID, err)
		// Acknowledge to prevent Stripe retries for orders we can't find.
		w.WriteHeader(http.StatusOK)
		return
	}

	// Mark order complete.
	if err := h.db.UpdateOrderStatus(r.Context(), order.ID, "complete"); err != nil {
		log.Printf("webhook: UpdateOrderStatus(%s): %v", order.ID, err)
	}

	// Clear the cart from KV.
	if cartToken != "" {
		if err := h.kv.DeleteCart(r.Context(), cartToken); err != nil {
			log.Printf("webhook: DeleteCart(%s): %v", cartToken, err)
		}
	}

	// Send confirmation email (non-fatal if it fails).
	if email != "" {
		if err := h.emailer.SendOrderConfirmation(r.Context(), email, order.ID, pi.Amount, pi.Currency); err != nil {
			log.Printf("webhook: SendOrderConfirmation(%s): %v", email, err)
		}
	}

	w.WriteHeader(http.StatusOK)
}
