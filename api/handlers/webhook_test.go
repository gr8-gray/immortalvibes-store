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
	labelErr      error                // if set, SendShippingLabel returns it
}

func (s *stubEmailSender) SendOrderConfirmation(ctx context.Context, toEmail, orderID string, totalAmount int64, currency string) error {
	s.sent = append(s.sent, toEmail)
	return nil
}

func (s *stubEmailSender) SendShippingLabel(ctx context.Context, ownerEmail, orderID, labelURL, trackingNum, carrier, labelCost string, orderTotal int64, currency string, items []models.LineItem) error {
	s.labelCalls++
	s.labelItems = items
	return s.labelErr
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

// stubFailingShipper rate-shops fine but fails to buy a label, exercising the
// shipping-failure notification path.
type stubFailingShipper struct{}

func (s *stubFailingShipper) RateShop(ctx context.Context, to shippo.Address) (string, error) {
	return "shp_stub:rate_fail", nil
}

func (s *stubFailingShipper) BuyLabel(ctx context.Context, rateID string) (string, string, string, string, error) {
	return "", "", "", "", fmt.Errorf("shippo unavailable")
}

// stubNotifier records order push notifications.
type stubNotifier struct {
	successCalls  int
	failureCalls  int
	lastFailStage string
}

func (n *stubNotifier) NotifyOrderSuccess(ctx context.Context, orderID, total, shipTo, items, tracking, carrier string) error {
	n.successCalls++
	return nil
}

func (n *stubNotifier) NotifyOrderFailure(ctx context.Context, orderID, email, stage, errMsg string) error {
	n.failureCalls++
	n.lastFailStage = stage
	return nil
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

func (s *stubStockStore) DecrementStock(ctx context.Context, productID, variant string, qty int) error {
	key := productID
	if variant != "" {
		key = productID + ":" + variant
	}
	if s.stock[key] < qty {
		// Fall back to product-level key for legacy carts.
		if s.stock[productID] < qty {
			return store.ErrInsufficientStock
		}
		s.stock[productID] -= qty
		return nil
	}
	s.stock[key] -= qty
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

	h := handlers.NewWebhookHandler(secret, stubs.kv, stubs.stock, stubs.db, emailer, &stubShipperClient{}, "owner@test.com", &stubNotifier{})

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

	notifier := &stubNotifier{}
	h := handlers.NewWebhookHandler(secret, stubs.kv, stubs.stock, stubs.db, emailer, shipper, "owner@test.com", notifier)

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
	if notifier.successCalls != 1 {
		t.Errorf("success notifications=%d want 1", notifier.successCalls)
	}
	if notifier.failureCalls != 0 {
		t.Errorf("failure notifications=%d want 0", notifier.failureCalls)
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
	if notifier.successCalls != 1 {
		t.Errorf("after resend success notifications=%d want 1 (no duplicate)", notifier.successCalls)
	}
}

// TestWebhookVariantDecrement verifies that a line item with Size="L" decrements
// the variant stock key (prod_1:L) and leaves other variants untouched.
func TestWebhookVariantDecrement(t *testing.T) {
	stubs := newWebhookStubs()
	emailer := &stubEmailSender{}
	shipper := &stubShipperClient{}
	secret := "whsec_var"

	stubs.db.orders["ord-var-001"] = &store.OrderRow{
		ID:              "ord-var-001",
		PaymentIntentID: "pi_var_001",
		CartToken:       "cart-var",
		Email:           "buyer@example.com",
		Currency:        "usd",
		TotalAmount:     6000,
		Status:          "pending",
		ShippingName:    "Test User",
		Line1:           "1 Test St",
		City:            "Testville",
		State:           "CA",
		PostalCode:      "90210",
		Country:         "US",
		LineItems: []models.LineItem{
			{ProductID: "prod_1", PriceID: "price_l", Name: "WR Tee", Size: "L", Amount: 3000, Quantity: 2},
		},
	}
	// Seed variant stocks: L=3, M=5, other variants untouched.
	stubs.stock.stock["prod_1:L"] = 3
	stubs.stock.stock["prod_1:M"] = 5

	h := handlers.NewWebhookHandler(secret, stubs.kv, stubs.stock, stubs.db, emailer, shipper, "owner@test.com", &stubNotifier{})

	payload := []byte(`{"type":"payment_intent.succeeded","data":{"object":{"id":"pi_var_001","metadata":{"cart_token":"cart-var","email":"buyer@example.com"},"amount":6000,"currency":"usd"}}}`)
	sig := signWebhookPayload(t, secret, payload)
	req := httptest.NewRequest(http.MethodPost, "/api/webhooks/stripe", bytes.NewReader(payload))
	req.Header.Set("Stripe-Signature", sig)
	w := httptest.NewRecorder()
	h.HandleWebhook(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}

	// L: 3→1 (decremented by qty 2)
	if stubs.stock.stock["prod_1:L"] != 1 {
		t.Errorf("variant L stock=%d want 1", stubs.stock.stock["prod_1:L"])
	}
	// M: untouched
	if stubs.stock.stock["prod_1:M"] != 5 {
		t.Errorf("variant M stock=%d want 5 (untouched)", stubs.stock.stock["prod_1:M"])
	}
	// Manifest email must still have Size=L
	if len(emailer.labelItems) != 1 || emailer.labelItems[0].Size != "L" {
		t.Errorf("owner manifest=%+v, want Size L", emailer.labelItems)
	}
}

// TestWebhookLegacyCartFallback verifies that a line item with Size="" falls
// back to the product_stock (legacy) path.
func TestWebhookLegacyCartFallback(t *testing.T) {
	stubs := newWebhookStubs()
	emailer := &stubEmailSender{}
	shipper := &stubShipperClient{}
	secret := "whsec_legacy"

	stubs.db.orders["ord-leg-001"] = &store.OrderRow{
		ID:              "ord-leg-001",
		PaymentIntentID: "pi_leg_001",
		CartToken:       "cart-leg",
		Email:           "buyer@example.com",
		Currency:        "usd",
		TotalAmount:     3000,
		Status:          "pending",
		ShippingName:    "Test User",
		Line1:           "1 Test St",
		City:            "Testville",
		State:           "CA",
		PostalCode:      "90210",
		Country:         "US",
		LineItems: []models.LineItem{
			// No Size — old cart that didn't carry size.
			{ProductID: "prod_legacy", PriceID: "price_usd", Name: "Old Tee", Amount: 3000, Quantity: 1},
		},
	}
	stubs.stock.stock["prod_legacy"] = 7 // legacy product_stock

	h := handlers.NewWebhookHandler(secret, stubs.kv, stubs.stock, stubs.db, emailer, shipper, "owner@test.com", &stubNotifier{})

	payload := []byte(`{"type":"payment_intent.succeeded","data":{"object":{"id":"pi_leg_001","metadata":{"cart_token":"cart-leg","email":"buyer@example.com"},"amount":3000,"currency":"usd"}}}`)
	sig := signWebhookPayload(t, secret, payload)
	req := httptest.NewRequest(http.MethodPost, "/api/webhooks/stripe", bytes.NewReader(payload))
	req.Header.Set("Stripe-Signature", sig)
	w := httptest.NewRecorder()
	h.HandleWebhook(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}

	if stubs.stock.stock["prod_legacy"] != 6 {
		t.Errorf("legacy stock=%d want 6", stubs.stock.stock["prod_legacy"])
	}
}

func TestWebhookInvalidSignature(t *testing.T) {
	stubs := newWebhookStubs()
	emailer := &stubEmailSender{}
	h := handlers.NewWebhookHandler("real_secret", stubs.kv, stubs.stock, stubs.db, emailer, &stubShipperClient{}, "owner@test.com", &stubNotifier{})

	payload := []byte(`{"type":"payment_intent.succeeded"}`)

	req := httptest.NewRequest(http.MethodPost, "/api/webhooks/stripe", bytes.NewReader(payload))
	req.Header.Set("Stripe-Signature", "t=1,v1=badsig")
	w := httptest.NewRecorder()
	h.HandleWebhook(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", w.Code)
	}
}

// TestWebhookShippingFailureNotifies verifies that when label purchase fails,
// the owner gets both the failure email and an urgent push, and no success
// push fires.
func TestWebhookShippingFailureNotifies(t *testing.T) {
	stubs := newWebhookStubs()
	emailer := &stubEmailSender{}
	notifier := &stubNotifier{}
	secret := "whsec_shipfail"

	stubs.db.orders["ord-fail"] = &store.OrderRow{
		ID:              "ord-fail",
		PaymentIntentID: "pi_fail_001",
		Email:           "buyer@example.com",
		Currency:        "usd",
		TotalAmount:     3000,
		Status:          "pending",
		ShippingName:    "Test User",
		Line1:           "1 Test St",
		City:            "Testville",
		State:           "CA",
		PostalCode:      "90210",
		Country:         "US",
		LineItems:       []models.LineItem{{ProductID: "prod_1", Name: "Tee", Size: "L", Amount: 3000, Quantity: 1}},
	}
	stubs.stock.stock["prod_1:L"] = 5

	h := handlers.NewWebhookHandler(secret, stubs.kv, stubs.stock, stubs.db, emailer, &stubFailingShipper{}, "owner@test.com", notifier)

	payload := []byte(`{"type":"payment_intent.succeeded","data":{"object":{"id":"pi_fail_001","metadata":{"email":"buyer@example.com"},"amount":3000,"currency":"usd"}}}`)
	sig := signWebhookPayload(t, secret, payload)
	req := httptest.NewRequest(http.MethodPost, "/api/webhooks/stripe", bytes.NewReader(payload))
	req.Header.Set("Stripe-Signature", sig)
	w := httptest.NewRecorder()
	h.HandleWebhook(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	if emailer.failureCalls != 1 {
		t.Errorf("shipping-failure emails=%d want 1", emailer.failureCalls)
	}
	if notifier.failureCalls != 1 || notifier.lastFailStage != "shipping" {
		t.Errorf("failure pushes=%d stage=%q want 1 stage=shipping", notifier.failureCalls, notifier.lastFailStage)
	}
	if notifier.successCalls != 0 {
		t.Errorf("success pushes=%d want 0 on shipping failure", notifier.successCalls)
	}
}

// TestWebhookOwnerEmailFailureNotifies verifies that a successful label but a
// failed owner manifest email fires an urgent push (not a quiet success),
// since that is the "owner didn't get notified" case.
func TestWebhookOwnerEmailFailureNotifies(t *testing.T) {
	stubs := newWebhookStubs()
	emailer := &stubEmailSender{labelErr: fmt.Errorf("resend 500")}
	notifier := &stubNotifier{}
	secret := "whsec_owneremailfail"

	stubs.db.orders["ord-oe"] = &store.OrderRow{
		ID:              "ord-oe",
		PaymentIntentID: "pi_oe_001",
		Email:           "buyer@example.com",
		Currency:        "usd",
		TotalAmount:     3000,
		Status:          "pending",
		ShippingName:    "Test User",
		Line1:           "1 Test St",
		City:            "Testville",
		State:           "CA",
		PostalCode:      "90210",
		Country:         "US",
		LineItems:       []models.LineItem{{ProductID: "prod_1", Name: "Tee", Size: "L", Amount: 3000, Quantity: 1}},
	}
	stubs.stock.stock["prod_1:L"] = 5

	h := handlers.NewWebhookHandler(secret, stubs.kv, stubs.stock, stubs.db, emailer, &stubShipperClient{}, "owner@test.com", notifier)

	payload := []byte(`{"type":"payment_intent.succeeded","data":{"object":{"id":"pi_oe_001","metadata":{"email":"buyer@example.com"},"amount":3000,"currency":"usd"}}}`)
	sig := signWebhookPayload(t, secret, payload)
	req := httptest.NewRequest(http.MethodPost, "/api/webhooks/stripe", bytes.NewReader(payload))
	req.Header.Set("Stripe-Signature", sig)
	w := httptest.NewRecorder()
	h.HandleWebhook(w, req)

	if notifier.failureCalls != 1 || notifier.lastFailStage != "owner-email" {
		t.Errorf("failure pushes=%d stage=%q want 1 stage=owner-email", notifier.failureCalls, notifier.lastFailStage)
	}
	if notifier.successCalls != 0 {
		t.Errorf("success pushes=%d want 0 when owner email failed", notifier.successCalls)
	}
}

// TestWebhookOrderLookupFailureNotifies verifies that a payment with no
// matching order fires an urgent push so the owner can reconcile.
func TestWebhookOrderLookupFailureNotifies(t *testing.T) {
	stubs := newWebhookStubs()
	emailer := &stubEmailSender{}
	notifier := &stubNotifier{}
	secret := "whsec_nolookup"

	h := handlers.NewWebhookHandler(secret, stubs.kv, stubs.stock, stubs.db, emailer, &stubShipperClient{}, "owner@test.com", notifier)

	// No order seeded — GetOrderByPaymentIntent returns ErrOrderNotFound.
	payload := []byte(`{"type":"payment_intent.succeeded","data":{"object":{"id":"pi_missing","metadata":{"email":"ghost@example.com"},"amount":3000,"currency":"usd"}}}`)
	sig := signWebhookPayload(t, secret, payload)
	req := httptest.NewRequest(http.MethodPost, "/api/webhooks/stripe", bytes.NewReader(payload))
	req.Header.Set("Stripe-Signature", sig)
	w := httptest.NewRecorder()
	h.HandleWebhook(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status=%d want 200 (ack to stop retries)", w.Code)
	}
	if notifier.failureCalls != 1 || notifier.lastFailStage != "order-lookup" {
		t.Errorf("failure pushes=%d stage=%q want 1 stage=order-lookup", notifier.failureCalls, notifier.lastFailStage)
	}
}
