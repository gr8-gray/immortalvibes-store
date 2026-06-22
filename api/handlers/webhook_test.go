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

// stubEmailSender records sent emails and the manifest passed to the owner.
type stubEmailSender struct {
	sent          []string             // order-confirmation recipients
	labelCalls    int                  // SendShippingLabel invocations
	labelItems    []models.LineItem    // manifest passed to the owner email
	trackingCalls int                  // SendTrackingUpdate invocations
	failureCalls  int                  // SendShippingFailure invocations
}

func (s *stubEmailSender) SendOrderConfirmation(ctx context.Context, toEmail, orderID string, totalAmount int64, currency string) error {
	s.sent = append(s.sent, toEmail)
	return nil
}

func (s *stubEmailSender) SendShippingLabel(ctx context.Context, ownerEmail, orderID, labelURL, trackingNum, carrier, labelCost string, orderTotal int64, currency string, items []models.LineItem) error {
	s.labelCalls++
	s.labelItems = items
	return nil
}

func (s *stubEmailSender) SendTrackingUpdate(ctx context.Context, customerEmail, orderID, trackingNum, carrier string) error {
	s.trackingCalls++
	return nil
}

func (s *stubEmailSender) SendShippingFailure(ctx context.Context, ownerEmail, orderID, customerEmail, shippingAddr, errMsg string) error {
	s.failureCalls++
	return nil
}

// stubShipperClient returns a fixed label without making network calls and
// counts label purchases so tests can assert idempotency (no double-buy).
type stubShipperClient struct {
	buyCalls int
}

func (s *stubShipperClient) RateShop(ctx context.Context, to shippo.Address) (string, error) {
	return "shp_stub:rate_stub_001", nil
}

func (s *stubShipperClient) BuyLabel(ctx context.Context, rateID string) (string, string, string, string, error) {
	s.buyCalls++
	return "TRACK123", "USPS", "https://shippo.example.com/label.pdf", "6.74 USD", nil
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

	h := handlers.NewWebhookHandler(secret, stubs.kv, stubs.stock, stubs.db, emailer, &stubShipperClient{}, "owner@test.com")

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

// TestWebhookFulfillmentE2E exercises the full purchase→fulfillment path on a
// paid order and asserts the operator-can-actually-ship outcome (STOP 19):
// manifest reaches the owner, customer is notified, stock decrements once, and
// a Stripe resend is idempotent (no double label, no double decrement).
func TestWebhookFulfillmentE2E(t *testing.T) {
	stubs := newWebhookStubs()
	emailer := &stubEmailSender{}
	shipper := &stubShipperClient{}
	secret := "whsec_e2e"

	// Paid order WITH a persisted manifest (STOP 18) — the items to pack.
	stubs.db.orders["ord-e2e"] = &store.OrderRow{
		ID:              "ord-e2e",
		PaymentIntentID: "pi_e2e_001",
		CartToken:       "cart-e2e",
		Email:           "buyer@example.com",
		Currency:        "usd",
		TotalAmount:     6000,
		Status:          "pending",
		ShippingName:    "Leslie Test",
		Line1:           "1 Test St",
		City:            "Corcoran",
		State:           "CA",
		PostalCode:      "93212",
		Country:         "US",
		LineItems: []models.LineItem{
			{ProductID: "prod_1", PriceID: "price_l", Name: "Immortal Light Sweatpants", Size: "L", Amount: 3000, Quantity: 2},
		},
	}
	stubs.stock.stock["prod_1"] = 10

	h := handlers.NewWebhookHandler(secret, stubs.kv, stubs.stock, stubs.db, emailer, shipper, "owner@test.com")

	payload := []byte(`{
		"type": "payment_intent.succeeded",
		"data": {"object": {"id": "pi_e2e_001",
			"metadata": {"cart_token": "cart-e2e", "email": "buyer@example.com"},
			"amount": 6000, "currency": "usd"}}
	}`)
	fire := func() *httptest.ResponseRecorder {
		sig := signWebhookPayload(t, secret, payload)
		req := httptest.NewRequest(http.MethodPost, "/api/webhooks/stripe", bytes.NewReader(payload))
		req.Header.Set("Stripe-Signature", sig)
		w := httptest.NewRecorder()
		h.HandleWebhook(w, req)
		return w
	}

	// First delivery — full fulfillment.
	if w := fire(); w.Code != http.StatusOK {
		t.Fatalf("first webhook status=%d body=%s", w.Code, w.Body.String())
	}
	o := stubs.db.orders["ord-e2e"]
	if o.Status != "complete" {
		t.Errorf("status=%q want complete", o.Status)
	}
	if o.TrackingNumber == "" {
		t.Error("tracking not set after fulfillment")
	}
	if emailer.labelCalls != 1 {
		t.Errorf("SendShippingLabel calls=%d want 1", emailer.labelCalls)
	}
	if len(emailer.labelItems) != 1 || emailer.labelItems[0].Size != "L" || emailer.labelItems[0].Quantity != 2 {
		t.Errorf("owner manifest=%+v, want 1 line (Size L, Qty 2)", emailer.labelItems)
	}
	if emailer.trackingCalls != 1 {
		t.Errorf("customer tracking emails=%d want 1", emailer.trackingCalls)
	}
	if emailer.failureCalls != 0 {
		t.Errorf("failure emails=%d want 0", emailer.failureCalls)
	}
	if stubs.stock.stock["prod_1"] != 8 {
		t.Errorf("stock=%d want 8 (10-2)", stubs.stock.stock["prod_1"])
	}
	if shipper.buyCalls != 1 {
		t.Errorf("BuyLabel calls=%d want 1", shipper.buyCalls)
	}

	// Second delivery (Stripe resend) — idempotent.
	if w := fire(); w.Code != http.StatusOK {
		t.Fatalf("resend status=%d", w.Code)
	}
	if shipper.buyCalls != 1 {
		t.Errorf("after resend BuyLabel calls=%d want 1 (no double-buy)", shipper.buyCalls)
	}
	if stubs.stock.stock["prod_1"] != 8 {
		t.Errorf("after resend stock=%d want 8 (no double decrement)", stubs.stock.stock["prod_1"])
	}
}

func TestWebhookInvalidSignature(t *testing.T) {
	stubs := newWebhookStubs()
	emailer := &stubEmailSender{}
	h := handlers.NewWebhookHandler("real_secret", stubs.kv, stubs.stock, stubs.db, emailer, &stubShipperClient{}, "owner@test.com")

	payload := []byte(`{"type":"payment_intent.succeeded"}`)

	req := httptest.NewRequest(http.MethodPost, "/api/webhooks/stripe", bytes.NewReader(payload))
	req.Header.Set("Stripe-Signature", "t=1,v1=badsig")
	w := httptest.NewRecorder()
	h.HandleWebhook(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", w.Code)
	}
}
