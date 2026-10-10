package handlers_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

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
	if len(*paths) != 1 {
		t.Fatalf("requests = %v, want exactly one", *paths)
	}
	u, err := url.Parse((*paths)[0])
	if err != nil {
		t.Fatalf("parse request URI: %v", err)
	}
	if u.Path != "/v1/prices/price_1" || u.Query().Get("expand[0]") != "product" {
		t.Errorf("request = %s, want GET /v1/prices/price_1 with expand[0]=product", (*paths)[0])
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

func TestRepriceLineItems_AttributesErrorsToLineItem(t *testing.T) {
	items := []models.LineItem{
		{PriceID: "price_usd", Name: "Tee", Quantity: 1},
		{PriceID: "", Name: "Legacy Hat", Quantity: 1},
	}
	_, err := handlers.RepriceLineItems(context.Background(), newFakeCatalog(), items)
	var lie *handlers.LineItemError
	if !errors.As(err, &lie) {
		t.Fatalf("err = %v, want *LineItemError", err)
	}
	if lie.Name != "Legacy Hat" || !errors.Is(err, handlers.ErrPriceUnavailable) {
		t.Errorf("err = %v, want Legacy Hat / ErrPriceUnavailable", err)
	}
}

func TestRepriceLineItems_LooksUpEachPriceOnce(t *testing.T) {
	cat := newFakeCatalog()
	items := []models.LineItem{
		{PriceID: "price_usd", Size: "S", Quantity: 1},
		{PriceID: "price_usd", Size: "M", Quantity: 1},
		{PriceID: "price_hat", Quantity: 1},
	}
	if _, err := handlers.RepriceLineItems(context.Background(), cat, items); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cat.lookups != 2 {
		t.Errorf("lookups = %d, want 2", cat.lookups)
	}
}

func TestRepriceLineItems_RejectsTooManyLines(t *testing.T) {
	items := make([]models.LineItem, handlers.MaxCartLines+1)
	for i := range items {
		items[i] = models.LineItem{PriceID: "price_usd", Quantity: 1}
	}
	if _, err := handlers.RepriceLineItems(context.Background(), newFakeCatalog(), items); !errors.Is(err, handlers.ErrCartFull) {
		t.Fatalf("err = %v, want ErrCartFull", err)
	}
}

func TestCachedPriceCatalog(t *testing.T) {
	inner := newFakeCatalog()
	now := time.Unix(0, 0)
	c := handlers.NewCachedPriceCatalog(inner, time.Minute)
	handlers.SetCatalogClock(c, func() time.Time { return now })
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		if _, err := c.Lookup(ctx, "price_usd"); err != nil {
			t.Fatalf("lookup %d: %v", i, err)
		}
	}
	if inner.lookups != 1 {
		t.Errorf("lookups within TTL = %d, want 1", inner.lookups)
	}

	now = now.Add(time.Minute)
	if _, err := c.Lookup(ctx, "price_usd"); err != nil {
		t.Fatalf("lookup after expiry: %v", err)
	}
	if inner.lookups != 2 {
		t.Errorf("lookups after expiry = %d, want 2", inner.lookups)
	}

	for i := 0; i < 2; i++ {
		if _, err := c.Lookup(ctx, "price_gone"); !errors.Is(err, handlers.ErrPriceUnavailable) {
			t.Fatalf("err = %v, want ErrPriceUnavailable", err)
		}
	}
	if inner.lookups != 4 {
		t.Errorf("failed lookups cached: lookups = %d, want 4", inner.lookups)
	}
}

func TestRepriceLineItems_TruncatesErrorName(t *testing.T) {
	items := []models.LineItem{{PriceID: "", Name: strings.Repeat("n", 500), Quantity: 1}}
	_, err := handlers.RepriceLineItems(context.Background(), newFakeCatalog(), items)
	var lie *handlers.LineItemError
	if !errors.As(err, &lie) {
		t.Fatalf("err = %v, want *LineItemError", err)
	}
	if len(lie.Name) != 80 {
		t.Errorf("name length = %d, want 80", len(lie.Name))
	}
}

// gatedCatalog blocks every lookup until release is closed.
type gatedCatalog struct {
	release chan struct{}
	mu      sync.Mutex
	calls   int
}

func (g *gatedCatalog) Lookup(ctx context.Context, priceID string) (handlers.CatalogPrice, error) {
	g.mu.Lock()
	g.calls++
	g.mu.Unlock()
	<-g.release
	return handlers.CatalogPrice{PriceID: priceID, Currency: "usd", UnitAmount: 100}, nil
}

func TestCachedPriceCatalog_SharesConcurrentMisses(t *testing.T) {
	g := &gatedCatalog{release: make(chan struct{})}
	c := handlers.NewCachedPriceCatalog(g, time.Minute)

	const n = 50
	var wg sync.WaitGroup
	errs := make(chan error, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := c.Lookup(context.Background(), "price_1")
			errs <- err
		}()
	}
	// Let every goroutine reach the cache before releasing the lookup.
	time.Sleep(50 * time.Millisecond)
	close(g.release)
	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			t.Fatalf("lookup: %v", err)
		}
	}
	if g.calls != 1 {
		t.Errorf("upstream calls = %d, want 1", g.calls)
	}
}

func TestCachedPriceCatalog_WaiterHonorsOwnContext(t *testing.T) {
	g := &gatedCatalog{release: make(chan struct{})}
	defer close(g.release)
	c := handlers.NewCachedPriceCatalog(g, time.Minute)

	go func() { _, _ = c.Lookup(context.Background(), "price_1") }()
	time.Sleep(20 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if _, err := c.Lookup(ctx, "price_1"); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("err = %v, want DeadlineExceeded", err)
	}
}

func TestCachedPriceCatalog_CancelledCallerDoesNotFailOthers(t *testing.T) {
	g := &gatedCatalog{release: make(chan struct{})}
	c := handlers.NewCachedPriceCatalog(g, time.Minute)

	first, cancelFirst := context.WithCancel(context.Background())
	firstErr := make(chan error, 1)
	go func() {
		_, err := c.Lookup(first, "price_1")
		firstErr <- err
	}()
	time.Sleep(20 * time.Millisecond)

	second := make(chan error, 1)
	go func() {
		_, err := c.Lookup(context.Background(), "price_1")
		second <- err
	}()
	time.Sleep(20 * time.Millisecond)

	cancelFirst()
	if err := <-firstErr; !errors.Is(err, context.Canceled) {
		t.Fatalf("first caller err = %v, want Canceled", err)
	}
	close(g.release)
	if err := <-second; err != nil {
		t.Fatalf("second caller err = %v, want nil", err)
	}
	if g.calls != 1 {
		t.Errorf("upstream calls = %d, want 1", g.calls)
	}
}

// panicOnceCatalog panics on its first lookup and succeeds afterwards.
type panicOnceCatalog struct {
	mu    sync.Mutex
	calls int
}

func (p *panicOnceCatalog) Lookup(_ context.Context, priceID string) (handlers.CatalogPrice, error) {
	p.mu.Lock()
	p.calls++
	n := p.calls
	p.mu.Unlock()
	if n == 1 {
		panic("boom")
	}
	return handlers.CatalogPrice{PriceID: priceID, Currency: "usd", UnitAmount: 100}, nil
}

func TestCachedPriceCatalog_PanicDoesNotWedgePrice(t *testing.T) {
	c := handlers.NewCachedPriceCatalog(&panicOnceCatalog{}, time.Minute)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if _, err := c.Lookup(ctx, "price_1"); err == nil || errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("first lookup err = %v, want panic converted to error", err)
	}
	if _, err := c.Lookup(ctx, "price_1"); err != nil {
		t.Fatalf("second lookup err = %v, want nil", err)
	}
}
