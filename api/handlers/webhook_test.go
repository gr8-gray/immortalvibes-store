package handlers_test

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/immortalvibes/api/handlers"
	"github.com/immortalvibes/api/models"
	"github.com/immortalvibes/api/shippo"
	"github.com/immortalvibes/api/store"
)

// webhookStubs aggregates all dependencies the webhook handler needs.
type webhookStubs struct {
	kv    *inMemoryKV
	stock *stubStockStore
	db    *stubOrderStore
	emails []string
}

func newWebhookStubs() *webhookStubs {
	return &webhookStubs{
		kv:    newInMemoryKV(),
		stock: newStubStockStore(),
		db:    newStubOrderStore(),
	}
}

// stubEmailSender records sent emails.
type stubEmailSender struct {
	sent []string
}

func (s *stubEmailSender) SendOrderConfirmation(ctx context.Context, toEmail, orderID string, totalAmount int64, currency string) error {
	s.sent = append(s.sent, toEmail)
	return nil
}

func (s *stubEmailSender) SendShippingLabel(ctx context.Context, ownerEmail, orderID, labelURL, trackingNum, carrier string) error {
	return nil
}

func (s *stubEmailSender) SendTrackingUpdate(ctx context.Context, customerEmail, orderID, trackingNum, carrier string) error {
	return nil
}

func (s *stubEmailSender) SendShippingFailure(ctx context.Context, ownerEmail, orderID, customerEmail, shippingAddr, errMsg string) error {
	return nil
}

// stubShippoClient returns a fixed label without making network calls.
type stubShippoClient struct{}

func (s *stubShippoClient) RateShop(ctx context.Context, to shippo.Address) (string, error) {
	return "rate_stub_001", nil
}

func (s *stubShippoClient) BuyLabel(ctx context.Context, rateID string) (string, string, string, error) {
	return "TRACK123", "USPS", "https://shippo.example.com/label.pdf", nil
}

func signWebhookPayload(t *testing.T, secret string, payload []byte) string {
	t.Helper()
	ts := time.Now().Unix()
	sig := computeStripeSignature(secret, ts, payload)
	return fmt.Sprintf("t=%d,v1=%s", ts, sig)
}

func computeStripeSignature(secret string, ts int64, payload []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	fmt.Fprintf(mac, "%d.", ts)
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

func (s *stubStockStore) DecrementStock(ctx context.Context, productID string, qty int) error {
	if s.stock[productID] < qty {
		return store.ErrInsufficientStock
	}
	s.stock[productID] -= qty
	return nil
}

func (s *stubOrderStore) GetOrderByPaymentIntent(ctx context.Context, paymentIntentID string) (*store.OrderRow, error) {
	for _, o := range s.orders {
		if o.PaymentIntentID == paymentIntentID {
			return o, nil
		}
	}
	return nil, store.ErrOrderNotFound
}

func (s *stubOrderStore) UpdateOrderStatus(ctx context.Context, id, status string) error {
	if o, ok := s.orders[id]; ok {
		o.Status = status
	}
	return nil
}

func TestWebhookPaymentIntentSucceeded(t *testing.T) {
	stubs := newWebhookStubs()
	emailer := &stubEmailSender{}
	secret := "whsec_test_secret"

	// Seed a pending order that matches the payment intent.
	stubs.db.orders["ord-wh-001"] = &store.OrderRow{
		ID:              "ord-wh-001",
		PaymentIntentID: "pi_webhook_001",
		CartToken:       "cart-wh-tok",
		Email:           "buyer@example.com",
		Currency:        "usd",
		TotalAmount:     2500,
		Status:          "pending",
	}

	// Seed a cart to verify it gets cleared.
	_ = stubs.kv.SetCart(context.Background(), &models.Cart{
		Token:     "cart-wh-tok",
		LineItems: []models.LineItem{{PriceID: "price_usd", ProductID: "prod_1", Quantity: 1, Amount: 2500}},
	})

	// Seed stock.
	stubs.stock.stock["prod_1"] = 10

	h := handlers.NewWebhookHandler(secret, stubs.kv, stubs.stock, stubs.db, emailer, &stubShippoClient{}, "owner@test.com")

	payload := []byte(`{
		"type": "payment_intent.succeeded",
		"data": {
			"object": {
				"id": "pi_webhook_001",
				"metadata": {
					"cart_token": "cart-wh-tok",
					"email": "buyer@example.com"
				},
				"amount": 2500,
				"currency": "usd"
			}
		}
	}`)

	sig := signWebhookPayload(t, secret, payload)

	req := httptest.NewRequest(http.MethodPost, "/api/webhooks/stripe", bytes.NewReader(payload))
	req.Header.Set("Stripe-Signature", sig)
	w := httptest.NewRecorder()
	h.HandleWebhook(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 — body: %s", w.Code, w.Body.String())
	}

	// Cart should be cleared.
	_, err := stubs.kv.GetCart(context.Background(), "cart-wh-tok")
	if err != store.ErrCartNotFound {
		t.Error("expected cart to be deleted after payment")
	}

	// Email should have been sent.
	if len(emailer.sent) != 1 || emailer.sent[0] != "buyer@example.com" {
		t.Errorf("emails sent = %v, want [buyer@example.com]", emailer.sent)
	}
}

func TestWebhookInvalidSignature(t *testing.T) {
	stubs := newWebhookStubs()
	emailer := &stubEmailSender{}
	h := handlers.NewWebhookHandler("real_secret", stubs.kv, stubs.stock, stubs.db, emailer, &stubShippoClient{}, "owner@test.com")

	payload := []byte(`{"type":"payment_intent.succeeded"}`)

	req := httptest.NewRequest(http.MethodPost, "/api/webhooks/stripe", bytes.NewReader(payload))
	req.Header.Set("Stripe-Signature", "t=1,v1=badsig")
	w := httptest.NewRecorder()
	h.HandleWebhook(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", w.Code)
	}
}
