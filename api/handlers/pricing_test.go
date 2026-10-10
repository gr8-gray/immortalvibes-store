package handlers_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/immortalvibes/api/handlers"
	"github.com/immortalvibes/api/models"
	stripe "github.com/stripe/stripe-go/v76"
)

// fakeCatalog is an in-memory PriceCatalog. Unknown price IDs return
// ErrPriceUnavailable; err, when set, is returned for every lookup.
type fakeCatalog struct {
	prices  map[string]handlers.CatalogPrice
	err     error
	lookups int
}

func newFakeCatalog() *fakeCatalog {
	return &fakeCatalog{prices: map[string]handlers.CatalogPrice{
		"price_usd": {PriceID: "price_usd", ProductID: "prod_1", ProductName: "Tee", Nickname: "M", Currency: "usd", UnitAmount: 2500},
		"price_tee": {PriceID: "price_tee", ProductID: "prod_1", ProductName: "Tee", Nickname: "L", Currency: "usd", UnitAmount: 2500},
		"price_hat": {PriceID: "price_hat", ProductID: "prod_2", ProductName: "Hat", Currency: "usd", UnitAmount: 1500},
		"price_eur": {PriceID: "price_eur", ProductID: "prod_1", ProductName: "Tee", Currency: "eur", UnitAmount: 2300},
	}}
}

func (f *fakeCatalog) Lookup(_ context.Context, priceID string) (handlers.CatalogPrice, error) {
	f.lookups++
	if f.err != nil {
		return handlers.CatalogPrice{}, f.err
	}
	cp, ok := f.prices[priceID]
	if !ok {
		return handlers.CatalogPrice{}, handlers.ErrPriceUnavailable
	}
	return cp, nil
}

func TestRepriceLineItems_OverwritesStoredPricing(t *testing.T) {
	items := []models.LineItem{
		{PriceID: "price_usd", ProductID: "prod_x", Name: "Other", Currency: "gbp", Amount: 1, Quantity: 2, Size: "S"},
		{PriceID: "price_hat", Amount: 0, Quantity: 1},
	}
	currency, err := handlers.RepriceLineItems(context.Background(), newFakeCatalog(), items)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if currency != "usd" {
		t.Errorf("currency = %q, want usd", currency)
	}

	want := []models.LineItem{
		{PriceID: "price_usd", ProductID: "prod_1", Name: "Tee", Currency: "usd", Amount: 2500, Quantity: 2, Size: "S"},
		{PriceID: "price_hat", ProductID: "prod_2", Name: "Hat", Currency: "usd", Amount: 1500, Quantity: 1},
	}
	for i := range want {
		if items[i] != want[i] {
			t.Errorf("item %d = %+v, want %+v", i, items[i], want[i])
		}
	}

	cart := models.Cart{LineItems: items}
	if got := cart.Total(); got != 6500 {
		t.Errorf("total = %d, want 6500", got)
	}
}

func TestRepriceLineItems_FillsEmptySizeFromNickname(t *testing.T) {
	items := []models.LineItem{{PriceID: "price_usd", Quantity: 1}}
	if _, err := handlers.RepriceLineItems(context.Background(), newFakeCatalog(), items); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if items[0].Size != "M" {
		t.Errorf("size = %q, want M", items[0].Size)
	}
}

func TestRepriceLineItems_Rejects(t *testing.T) {
	upstream := errors.New("stripe down")
	cases := []struct {
		name    string
		items   []models.LineItem
		catErr  error
		wantErr error
	}{
		{"zero quantity", []models.LineItem{{PriceID: "price_usd", Quantity: 0}}, nil, handlers.ErrInvalidQuantity},
		{"negative quantity", []models.LineItem{{PriceID: "price_usd", Quantity: -3}}, nil, handlers.ErrInvalidQuantity},
		{"quantity over max", []models.LineItem{{PriceID: "price_usd", Quantity: handlers.MaxLineQuantity + 1}}, nil, handlers.ErrInvalidQuantity},
		{"empty price id", []models.LineItem{{Quantity: 1}}, nil, handlers.ErrPriceUnavailable},
		{"unknown price", []models.LineItem{{PriceID: "price_gone", Quantity: 1}}, nil, handlers.ErrPriceUnavailable},
		{"mixed currency", []models.LineItem{{PriceID: "price_usd", Quantity: 1}, {PriceID: "price_eur", Quantity: 1}}, nil, handlers.ErrMixedCurrency},
		{"upstream failure", []models.LineItem{{PriceID: "price_usd", Quantity: 1}}, upstream, upstream},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cat := newFakeCatalog()
			cat.err = tc.catErr
			_, err := handlers.RepriceLineItems(context.Background(), cat, tc.items)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("err = %v, want %v", err, tc.wantErr)
			}
		})
	}
}

func TestRepriceLineItems_MaxQuantityAllowed(t *testing.T) {
	items := []models.LineItem{{PriceID: "price_usd", Quantity: handlers.MaxLineQuantity}}
	if _, err := handlers.RepriceLineItems(context.Background(), newFakeCatalog(), items); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCatalogPriceFromStripe(t *testing.T) {
	valid := func() *stripe.Price {
		return &stripe.Price{
			ID:            "price_1",
			Active:        true,
			Type:          stripe.PriceTypeOneTime,
			BillingScheme: stripe.PriceBillingSchemePerUnit,
			Currency:      stripe.CurrencyUSD,
			UnitAmount:    4200,
			Nickname:      "XL",
			Product:       &stripe.Product{ID: "prod_1", Name: "Tee", Active: true},
		}
	}

	got, err := handlers.CatalogPriceFromStripe(valid())
	if err != nil {
		t.Fatalf("valid price: unexpected error: %v", err)
	}
	want := handlers.CatalogPrice{PriceID: "price_1", ProductID: "prod_1", ProductName: "Tee", Nickname: "XL", Currency: "usd", UnitAmount: 4200}
	if got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}

	cases := []struct {
		name   string
		mutate func(*stripe.Price)
	}{
		{"inactive", func(p *stripe.Price) { p.Active = false }},
		{"deleted", func(p *stripe.Price) { p.Deleted = true }},
		{"recurring", func(p *stripe.Price) { p.Type = stripe.PriceTypeRecurring }},
		{"tiered", func(p *stripe.Price) { p.BillingScheme = stripe.PriceBillingSchemeTiered }},
		{"custom amount", func(p *stripe.Price) { p.CustomUnitAmount = &stripe.PriceCustomUnitAmount{} }},
		{"zero amount", func(p *stripe.Price) { p.UnitAmount = 0 }},
		{"negative amount", func(p *stripe.Price) { p.UnitAmount = -1 }},
		{"no product", func(p *stripe.Price) { p.Product = nil }},
		{"inactive product", func(p *stripe.Price) { p.Product.Active = false }},
		{"deleted product", func(p *stripe.Price) { p.Product.Deleted = true }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := valid()
			tc.mutate(p)
			if _, err := handlers.CatalogPriceFromStripe(p); !errors.Is(err, handlers.ErrPriceUnavailable) {
				t.Fatalf("err = %v, want ErrPriceUnavailable", err)
			}
		})
	}

	if _, err := handlers.CatalogPriceFromStripe(nil); !errors.Is(err, handlers.ErrPriceUnavailable) {
		t.Fatalf("nil price: err = %v, want ErrPriceUnavailable", err)
	}
}

// withStripeServer points the Stripe API backend at handler for the duration
// of the test.
func withStripeServer(t *testing.T, handler http.HandlerFunc) *[]string {
	t.Helper()
	var paths []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.RequestURI())
		handler(w, r)
	}))
	t.Cleanup(srv.Close)

	prevKey := stripe.Key
	prevBackend := stripe.GetBackend(stripe.APIBackend)
	stripe.Key = "sk_test_catalog"
	stripe.SetBackend(stripe.APIBackend, stripe.GetBackendWithConfig(stripe.APIBackend, &stripe.BackendConfig{
		URL:               stripe.String(srv.URL),
		MaxNetworkRetries: stripe.Int64(0),
		LeveledLogger:     &stripe.LeveledLogger{Level: stripe.LevelNull},
	}))
	t.Cleanup(func() {
		stripe.Key = prevKey
		stripe.SetBackend(stripe.APIBackend, prevBackend)
	})
	return &paths
}

func TestStripePriceCatalog_Lookup(t *testing.T) {
	paths := withStripeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id": "price_1", "object": "price", "active": true, "type": "one_time",
			"billing_scheme": "per_unit", "currency": "usd", "unit_amount": 4200, "nickname": "XL",
			"product": {"id": "prod_1", "object": "product", "name": "Tee", "active": true}
		}`))
	})

	got, err := handlers.StripePriceCatalog{}.Lookup(context.Background(), "price_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := handlers.CatalogPrice{PriceID: "price_1", ProductID: "prod_1", ProductName: "Tee", Nickname: "XL", Currency: "usd", UnitAmount: 4200}
	if got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}
	if len(*paths) != 1 || !strings.Contains((*paths)[0], "expand") {
		t.Errorf("requests = %v, want one request expanding product", *paths)
	}
}

func TestStripePriceCatalog_LookupErrors(t *testing.T) {
	cases := []struct {
		name        string
		status      int
		body        string
		unavailable bool
	}{
		{"missing price", http.StatusNotFound, `{"error": {"type": "invalid_request_error", "code": "resource_missing", "message": "No such price"}}`, true},
		{"server error", http.StatusInternalServerError, `{"error": {"type": "api_error", "message": "boom"}}`, false},
		{"inactive price", http.StatusOK, `{"id": "price_1", "object": "price", "active": false, "type": "one_time", "billing_scheme": "per_unit", "currency": "usd", "unit_amount": 100, "product": {"id": "prod_1", "active": true}}`, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			withStripeServer(t, func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			})

			_, err := handlers.StripePriceCatalog{}.Lookup(context.Background(), "price_1")
			if err == nil {
				t.Fatal("expected error")
			}
			if got := errors.Is(err, handlers.ErrPriceUnavailable); got != tc.unavailable {
				t.Errorf("errors.Is(ErrPriceUnavailable) = %v, want %v (err: %v)", got, tc.unavailable, err)
			}
		})
	}
}
