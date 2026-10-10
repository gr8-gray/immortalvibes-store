package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/immortalvibes/api/handlers"
	"github.com/immortalvibes/api/models"
	"github.com/immortalvibes/api/shippo"
	"github.com/immortalvibes/api/store"
	stripe "github.com/stripe/stripe-go/v76"
)

// fakeShippo is a test double for handlers.ShippoEstimator.
type fakeShippo struct {
	amount   float64 // dollars, as Shippo returns
	currency string  // defaults to USD
	err      error
	calls    int
	lastTo   shippo.Address
}

func (f *fakeShippo) EstimateRate(ctx context.Context, to shippo.Address) (*shippo.RateEstimate, error) {
	f.calls++
	f.lastTo = to
	if f.err != nil {
		return nil, f.err
	}
	currency := f.currency
	if currency == "" {
		currency = "USD"
	}
	return &shippo.RateEstimate{Provider: "USPS", Service: "Ground Advantage", Amount: f.amount, Currency: currency}, nil
}

// fakeCheckoutStore is a test double for handlers.CheckoutStore.
type fakeCheckoutStore struct {
	saved    []store.OrderRow
	variants map[string][]store.VariantStockRow
}

func (s *fakeCheckoutStore) GetVariantStocks(ctx context.Context, productID string) ([]store.VariantStockRow, error) {
	return s.variants[productID], nil
}

func (s *fakeCheckoutStore) SaveOrder(ctx context.Context, o store.OrderRow) error {
	s.saved = append(s.saved, o)
	return nil
}

type checkoutFixture struct {
	h         *handlers.CheckoutHandler
	shippo    *fakeShippo
	db        *fakeCheckoutStore
	catalog   *fakeCatalog
	discounts *fakeDiscounts
	kv        *inMemoryKV
	piAmount  []int64 // amounts passed to Stripe
	piParams  []*stripe.PaymentIntentParams
}

const cartSubtotal = int64(2 * 3500) // 2 x $35.00

func newCheckoutFixture(sh *fakeShippo) *checkoutFixture {
	kv := newInMemoryKV()
	kv.carts["tok"] = &models.Cart{
		Token: "tok",
		LineItems: []models.LineItem{
			{PriceID: "price_tee35", Name: "Tee", Size: "M", Currency: "usd", Amount: 3500, Quantity: 2},
		},
	}
	cat := newFakeCatalog()
	cat.prices["price_tee35"] = handlers.CatalogPrice{PriceID: "price_tee35", ProductID: "prod_tee", ProductName: "Tee", Currency: "usd", UnitAmount: 3500}
	fx := &checkoutFixture{shippo: sh, db: &fakeCheckoutStore{}, catalog: cat, discounts: newFakeDiscounts(), kv: kv}
	fx.h = handlers.NewCheckoutHandler("sk_test_dummy", kv, fx.db, cat, sh, fx.discounts)
	fx.h.SetPaymentIntentFunc(func(p *stripe.PaymentIntentParams) (*stripe.PaymentIntent, error) {
		fx.piAmount = append(fx.piAmount, *p.Amount)
		fx.piParams = append(fx.piParams, p)
		return &stripe.PaymentIntent{ID: "pi_test", ClientSecret: "pi_test_secret"}, nil
	})
	return fx
}

func checkoutBody(extra map[string]any) []byte {
	body := map[string]any{
		"cart_token":    "tok",
		"email":         "buyer@example.com",
		"shipping_name": "Buyer",
		"line1":         "1 Test St",
		"city":          "Austin",
		"state":         "TX",
		"postal_code":   "78701",
		"country":       "US",
	}
	for k, v := range extra {
		body[k] = v
	}
	b, _ := json.Marshal(body)
	return b
}

func (fx *checkoutFixture) do(t *testing.T, extra map[string]any) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/checkout", bytes.NewReader(checkoutBody(extra)))
	w := httptest.NewRecorder()
	fx.h.Checkout(w, req)
	return w
}

func (fx *checkoutFixture) assertCharged(t *testing.T, w *httptest.ResponseRecorder, want int64) {
	t.Helper()
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %q", w.Code, w.Body.String())
	}
	if len(fx.piAmount) != 1 || fx.piAmount[0] != want {
		t.Fatalf("PaymentIntent amounts = %v, want [%d]", fx.piAmount, want)
	}
	var resp handlers.CheckoutResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.TotalAmount != want {
		t.Errorf("response total_amount = %d, want %d", resp.TotalAmount, want)
	}
	if len(fx.db.saved) != 1 || fx.db.saved[0].TotalAmount != want {
		t.Errorf("saved orders = %+v, want one with total %d", fx.db.saved, want)
	}
}

func TestCheckout_ShippingFromServerQuote(t *testing.T) {
	for name, extra := range map[string]map[string]any{
		"field omitted":   nil,
		"field zero":      {"shipping_cost": 0},
		"field below":     {"shipping_cost": -1},
		"field above":     {"shipping_cost": 1000},
		"field unrelated": {"shipping_cost": 123},
	} {
		t.Run(name, func(t *testing.T) {
			fx := newCheckoutFixture(&fakeShippo{amount: 7.50})
			w := fx.do(t, extra)
			fx.assertCharged(t, w, cartSubtotal+750)
		})
	}
}

func TestCheckout_ShippingQuotedForCheckoutAddress(t *testing.T) {
	sh := &fakeShippo{amount: 7.50}
	fx := newCheckoutFixture(sh)
	fx.do(t, nil)
	want := shippo.Address{Name: "Buyer", Street1: "1 Test St", City: "Austin", State: "TX", Zip: "78701", Country: "US"}
	if sh.calls != 1 || sh.lastTo != want {
		t.Errorf("EstimateRate calls=%d to=%+v, want 1 call to %+v", sh.calls, sh.lastTo, want)
	}
}

func TestCheckout_ShippingErrorFailsCheckout(t *testing.T) {
	fx := newCheckoutFixture(&fakeShippo{err: errors.New("shippo: 503")})
	w := fx.do(t, map[string]any{"shipping_cost": 750})
	if w.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want %d; body = %q", w.Code, http.StatusBadGateway, w.Body.String())
	}
	if len(fx.piAmount) != 0 {
		t.Errorf("PaymentIntent created despite shipping error: %v", fx.piAmount)
	}
	if len(fx.db.saved) != 0 {
		t.Errorf("order saved despite shipping error: %+v", fx.db.saved)
	}
}

func TestShippingEstimate_UsesSharedQuote(t *testing.T) {
	h := handlers.NewShippingHandler(&fakeShippo{amount: 7.499}, shippo.Address{})
	body := []byte(`{"street1":"1 Test St","city":"Austin","state":"TX","zip":"78701"}`)
	w := httptest.NewRecorder()
	h.Estimate(w, httptest.NewRequest(http.MethodPost, "/api/shipping/estimate", bytes.NewReader(body)))
	var resp struct {
		Rate *struct {
			Amount int `json:"amount"`
		} `json:"rate"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil || resp.Rate == nil {
		t.Fatalf("decode: %v body=%q", err, w.Body.String())
	}
	if resp.Rate.Amount != 750 {
		t.Errorf("estimate amount = %d, want 750", resp.Rate.Amount)
	}
}

func TestCheckout_ChargesCatalogPriceNotStoredAmount(t *testing.T) {
	fx := newCheckoutFixture(&fakeShippo{amount: 7.50})
	fx.kv.carts["tok"].LineItems[0].Amount = 1

	w := fx.do(t, nil)
	fx.assertCharged(t, w, cartSubtotal+750)
	if got := fx.db.saved[0].LineItems[0].Amount; got != 3500 {
		t.Errorf("persisted unit amount = %d, want 3500", got)
	}
}

func TestCheckout_UnavailablePriceFailsCheckout(t *testing.T) {
	fx := newCheckoutFixture(&fakeShippo{amount: 7.50})
	delete(fx.catalog.prices, "price_tee35")

	w := fx.do(t, nil)
	if w.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409", w.Code)
	}
	if len(fx.piAmount) != 0 || len(fx.db.saved) != 0 {
		t.Errorf("payment intents = %v, orders = %d; want none", fx.piAmount, len(fx.db.saved))
	}
}

func TestCheckout_TrackedVariants(t *testing.T) {
	cases := []struct {
		name     string
		size     string
		stock    int
		wantCode int
		wantSize string
	}{
		{"canonicalized and in stock", "m", 5, http.StatusOK, "M"},
		{"sold out after canonicalization", "m", 0, http.StatusConflict, ""},
		{"missing variant", "", 5, http.StatusConflict, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fx := newCheckoutFixture(&fakeShippo{amount: 7.50})
			fx.kv.carts["tok"].LineItems[0].Size = tc.size
			fx.db.variants = map[string][]store.VariantStockRow{"prod_tee": {{Variant: "M", StockCount: tc.stock}}}

			w := fx.do(t, nil)
			if w.Code != tc.wantCode {
				t.Fatalf("status = %d, want %d: %s", w.Code, tc.wantCode, w.Body)
			}
			if tc.wantCode == http.StatusOK && fx.db.saved[0].LineItems[0].Size != tc.wantSize {
				t.Errorf("persisted size = %q, want %q", fx.db.saved[0].LineItems[0].Size, tc.wantSize)
			}
		})
	}
}

func TestCheckout_ChargesCatalogCurrency(t *testing.T) {
	fx := newCheckoutFixture(&fakeShippo{amount: 7.50})
	req := httptest.NewRequest(http.MethodPost, "/api/checkout", bytes.NewReader(checkoutBody(nil)))
	req.Header.Set("CF-IPCountry", "GB")
	w := httptest.NewRecorder()
	fx.h.Checkout(w, req)

	fx.assertCharged(t, w, cartSubtotal+750)
	if got := *fx.piParams[0].Currency; got != "usd" {
		t.Errorf("payment intent currency = %q, want usd", got)
	}
	if got := fx.db.saved[0].Currency; got != "usd" {
		t.Errorf("order currency = %q, want usd", got)
	}
}

func TestCheckout_RejectsShippingQuoteInOtherCurrency(t *testing.T) {
	fx := newCheckoutFixture(&fakeShippo{amount: 7.50, currency: "GBP"})
	w := fx.do(t, nil)
	if w.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want 502", w.Code)
	}
	if len(fx.piAmount) != 0 {
		t.Errorf("payment intent created: %v", fx.piAmount)
	}
}

func TestCheckout_PercentDiscountAppliesAfterShipping(t *testing.T) {
	fx := newCheckoutFixture(&fakeShippo{amount: 7.50})
	w := fx.do(t, map[string]any{"discount_code": "vibe10"})

	total := cartSubtotal + 750
	fx.assertCharged(t, w, total-total/10)
	md := fx.piParams[0].Metadata
	if md["discount_code"] != "VIBE10" || md["coupon_id"] != "cpn_10" || md["promotion_code_id"] != "promo_vibe10" {
		t.Errorf("metadata = %v", md)
	}
}

func TestCheckout_DiscountFailures(t *testing.T) {
	cases := []struct {
		name     string
		code     string
		err      error
		wantCode int
	}{
		{"unknown code", "NOPE", nil, http.StatusConflict},
		{"exhausted", "VIBE10", handlers.ErrDiscountExhausted, http.StatusConflict},
		{"stripe failure", "VIBE10", errors.New("stripe down"), http.StatusBadGateway},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fx := newCheckoutFixture(&fakeShippo{amount: 7.50})
			fx.discounts.err = tc.err
			w := fx.do(t, map[string]any{"discount_code": tc.code})
			if w.Code != tc.wantCode {
				t.Fatalf("status = %d, want %d", w.Code, tc.wantCode)
			}
			if len(fx.piAmount) != 0 || len(fx.db.saved) != 0 {
				t.Errorf("payment intents = %v, orders = %d; want none", fx.piAmount, len(fx.db.saved))
			}
		})
	}
}

func TestCheckout_DiscountNotApplicable(t *testing.T) {
	fx := newCheckoutFixture(&fakeShippo{amount: 7.50})
	fx.discounts.codes["BIG"] = handlers.Discount{Code: "BIG", PercentOff: 10, MinimumAmount: 100000, MinimumCurrency: "usd"}
	w := fx.do(t, map[string]any{"discount_code": "BIG"})
	if w.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409", w.Code)
	}
}
