package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/immortalvibes/api/handlers"
	"github.com/immortalvibes/api/models"
	"github.com/immortalvibes/api/store"
)

// inMemoryKV is a test double for the KV store.
type inMemoryKV struct {
	carts map[string]*models.Cart
}

func newInMemoryKV() *inMemoryKV {
	return &inMemoryKV{carts: map[string]*models.Cart{}}
}

func (m *inMemoryKV) GetCart(ctx context.Context, token string) (*models.Cart, error) {
	c, ok := m.carts[token]
	if !ok {
		return nil, store.ErrCartNotFound
	}
	return c, nil
}

func (m *inMemoryKV) SetCart(ctx context.Context, cart *models.Cart) error {
	m.carts[cart.Token] = cart
	return nil
}

func (m *inMemoryKV) DeleteCart(ctx context.Context, token string) error {
	delete(m.carts, token)
	return nil
}

// fakeVariants serves variant stock rows by product ID. err, when set, is
// returned for every read.
type fakeVariants struct {
	rows map[string][]store.VariantStockRow
	err  error
}

func (f fakeVariants) GetVariantStocks(_ context.Context, productID string) ([]store.VariantStockRow, error) {
	return f.rows[productID], f.err
}

func TestGetCart_NoToken(t *testing.T) {
	kv := newInMemoryKV()
	h := handlers.NewCartHandler(kv, newFakeCatalog(), fakeVariants{})

	r := chi.NewRouter()
	r.Get("/api/cart/{token}", h.GetCart)

	req := httptest.NewRequest(http.MethodGet, "/api/cart/unknown-tok", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", w.Code)
	}
}

func TestPostCart_CreatesNewCart(t *testing.T) {
	kv := newInMemoryKV()
	h := handlers.NewCartHandler(kv, newFakeCatalog(), fakeVariants{})

	body := handlers.AddToCartRequest{
		PriceID:   "price_usd",
		ProductID: "prod_1",
		Name:      "Tee",
		ImageURL:  "https://r2.example.com/tee.jpg",
		Currency:  "usd",
		Amount:    2500,
		Quantity:  1,
	}
	b, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/cart", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.AddToCart(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	var cart models.Cart
	if err := json.NewDecoder(w.Body).Decode(&cart); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if cart.Token == "" {
		t.Error("expected non-empty cart token")
	}
	if len(cart.LineItems) != 1 {
		t.Fatalf("got %d line items, want 1", len(cart.LineItems))
	}
	if cart.LineItems[0].PriceID != "price_usd" {
		t.Errorf("price_id = %q, want price_usd", cart.LineItems[0].PriceID)
	}

	// Cookie must be set
	cookies := w.Result().Cookies()
	var found bool
	for _, c := range cookies {
		if c.Name == "cart_token" {
			found = true
			if c.Value != cart.Token {
				t.Errorf("cookie value %q != cart token %q", c.Value, cart.Token)
			}
		}
	}
	if !found {
		t.Error("cart_token cookie not set")
	}
}

func TestPostCart_ExistingToken_AppendsItem(t *testing.T) {
	kv := newInMemoryKV()
	h := handlers.NewCartHandler(kv, newFakeCatalog(), fakeVariants{})

	// Seed a cart
	_ = kv.SetCart(context.Background(), &models.Cart{
		Token: "existing-tok",
		LineItems: []models.LineItem{
			{PriceID: "price_tee", ProductID: "prod_1", Name: "Tee", Currency: "usd", Amount: 2500, Quantity: 1},
		},
	})

	body := handlers.AddToCartRequest{
		PriceID:   "price_usd",
		ProductID: "prod_2",
		Name:      "Hat",
		Currency:  "usd",
		Amount:    1500,
		Quantity:  2,
	}
	b, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/cart", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "cart_token", Value: "existing-tok"})
	w := httptest.NewRecorder()
	h.AddToCart(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	var cart models.Cart
	json.NewDecoder(w.Body).Decode(&cart)
	if len(cart.LineItems) != 2 {
		t.Errorf("got %d line items, want 2", len(cart.LineItems))
	}
}

func TestPutCart_UpdatesQuantity(t *testing.T) {
	kv := newInMemoryKV()
	h := handlers.NewCartHandler(kv, newFakeCatalog(), fakeVariants{})

	_ = kv.SetCart(context.Background(), &models.Cart{
		Token: "upd-tok",
		LineItems: []models.LineItem{
			{PriceID: "price_usd", ProductID: "prod_1", Name: "Tee", Currency: "usd", Amount: 2500, Quantity: 1},
		},
	})

	body := handlers.UpdateLineItemRequest{PriceID: "price_usd", Quantity: 3}
	b, _ := json.Marshal(body)

	r := chi.NewRouter()
	r.Put("/api/cart/{token}", h.UpdateCart)

	req := httptest.NewRequest(http.MethodPut, "/api/cart/upd-tok", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "cart_token", Value: "upd-tok"})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var cart models.Cart
	json.NewDecoder(w.Body).Decode(&cart)
	if cart.LineItems[0].Quantity != 3 {
		t.Errorf("quantity = %d, want 3", cart.LineItems[0].Quantity)
	}
}

func postCart(t *testing.T, h *handlers.CartHandler, body handlers.AddToCartRequest, token string) *httptest.ResponseRecorder {
	t.Helper()
	b, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/cart", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.AddCookie(&http.Cookie{Name: "cart_token", Value: token})
	}
	w := httptest.NewRecorder()
	h.AddToCart(w, req)
	return w
}

func TestPostCart_IgnoresClientPricing(t *testing.T) {
	kv := newInMemoryKV()
	h := handlers.NewCartHandler(kv, newFakeCatalog(), fakeVariants{})

	w := postCart(t, h, handlers.AddToCartRequest{
		PriceID:   "price_usd",
		ProductID: "prod_other",
		Name:      "Renamed",
		Currency:  "gbp",
		Amount:    1,
		Quantity:  2,
		Size:      "M",
	}, "")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", w.Code, w.Body)
	}

	var cart models.Cart
	if err := json.NewDecoder(w.Body).Decode(&cart); err != nil {
		t.Fatalf("decode: %v", err)
	}
	got := cart.LineItems[0]
	if got.Amount != 2500 || got.Currency != "usd" || got.ProductID != "prod_1" || got.Name != "Tee" {
		t.Errorf("line item = %+v, want catalog pricing (2500 usd prod_1 Tee)", got)
	}

	stored := kv.carts[cart.Token]
	if stored.Total() != 5000 {
		t.Errorf("stored total = %d, want 5000", stored.Total())
	}
}

func TestPostCart_RejectsInvalidRequests(t *testing.T) {
	cases := []struct {
		name     string
		body     handlers.AddToCartRequest
		catErr   error
		wantCode int
	}{
		{"missing price id", handlers.AddToCartRequest{Quantity: 1}, nil, http.StatusBadRequest},
		{"zero quantity", handlers.AddToCartRequest{PriceID: "price_usd", Quantity: 0}, nil, http.StatusBadRequest},
		{"negative quantity", handlers.AddToCartRequest{PriceID: "price_usd", Quantity: -5}, nil, http.StatusBadRequest},
		{"quantity over max", handlers.AddToCartRequest{PriceID: "price_usd", Quantity: handlers.MaxLineQuantity + 1}, nil, http.StatusBadRequest},
		{"unknown price", handlers.AddToCartRequest{PriceID: "price_gone", Quantity: 1}, nil, http.StatusConflict},
		{"catalog unavailable", handlers.AddToCartRequest{PriceID: "price_usd", Quantity: 1}, errors.New("stripe down"), http.StatusBadGateway},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			kv := newInMemoryKV()
			cat := newFakeCatalog()
			cat.err = tc.catErr
			h := handlers.NewCartHandler(kv, cat, fakeVariants{})

			w := postCart(t, h, tc.body, "")
			if w.Code != tc.wantCode {
				t.Fatalf("status = %d, want %d", w.Code, tc.wantCode)
			}
			if len(kv.carts) != 0 {
				t.Errorf("cart persisted on rejected request")
			}
		})
	}
}

func TestPostCart_MergeCannotExceedMaxQuantity(t *testing.T) {
	kv := newInMemoryKV()
	h := handlers.NewCartHandler(kv, newFakeCatalog(), fakeVariants{})
	_ = kv.SetCart(context.Background(), &models.Cart{
		Token: "tok",
		LineItems: []models.LineItem{
			{PriceID: "price_usd", Size: "M", Currency: "usd", Amount: 2500, Quantity: handlers.MaxLineQuantity},
		},
	})

	w := postCart(t, h, handlers.AddToCartRequest{PriceID: "price_usd", Size: "M", Quantity: 1}, "tok")
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", w.Code)
	}
	if q := kv.carts["tok"].LineItems[0].Quantity; q != handlers.MaxLineQuantity {
		t.Errorf("quantity = %d, want unchanged %d", q, handlers.MaxLineQuantity)
	}
}

func TestPostCart_MergeRepricesExistingLine(t *testing.T) {
	kv := newInMemoryKV()
	h := handlers.NewCartHandler(kv, newFakeCatalog(), fakeVariants{})
	_ = kv.SetCart(context.Background(), &models.Cart{
		Token: "tok",
		LineItems: []models.LineItem{
			{PriceID: "price_usd", Size: "M", Currency: "usd", Amount: 1, Quantity: 1},
		},
	})

	w := postCart(t, h, handlers.AddToCartRequest{PriceID: "price_usd", Size: "M", Quantity: 1}, "tok")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	li := kv.carts["tok"].LineItems[0]
	if li.Quantity != 2 || li.Amount != 2500 {
		t.Errorf("line item = %+v, want quantity 2 at 2500", li)
	}
}

func TestPostCart_RejectsMixedCurrency(t *testing.T) {
	kv := newInMemoryKV()
	h := handlers.NewCartHandler(kv, newFakeCatalog(), fakeVariants{})
	_ = kv.SetCart(context.Background(), &models.Cart{
		Token:     "tok",
		LineItems: []models.LineItem{{PriceID: "price_usd", Currency: "usd", Amount: 2500, Quantity: 1}},
	})

	w := postCart(t, h, handlers.AddToCartRequest{PriceID: "price_eur", Quantity: 1}, "tok")
	if w.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409", w.Code)
	}
	if n := len(kv.carts["tok"].LineItems); n != 1 {
		t.Errorf("line items = %d, want 1", n)
	}
}

func TestPutCart_RejectsQuantityOverMax(t *testing.T) {
	kv := newInMemoryKV()
	h := handlers.NewCartHandler(kv, newFakeCatalog(), fakeVariants{})
	_ = kv.SetCart(context.Background(), &models.Cart{
		Token:     "tok",
		LineItems: []models.LineItem{{PriceID: "price_usd", Currency: "usd", Amount: 2500, Quantity: 1}},
	})

	b, _ := json.Marshal(handlers.UpdateLineItemRequest{PriceID: "price_usd", Quantity: handlers.MaxLineQuantity + 1})
	r := chi.NewRouter()
	r.Put("/api/cart/{token}", h.UpdateCart)
	req := httptest.NewRequest(http.MethodPut, "/api/cart/tok", bytes.NewReader(b))
	req.AddCookie(&http.Cookie{Name: "cart_token", Value: "tok"})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", w.Code)
	}
	if q := kv.carts["tok"].LineItems[0].Quantity; q != 1 {
		t.Errorf("quantity = %d, want unchanged 1", q)
	}
}

func putCart(t *testing.T, h *handlers.CartHandler, token string, body handlers.UpdateLineItemRequest) *httptest.ResponseRecorder {
	t.Helper()
	b, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	r := chi.NewRouter()
	r.Put("/api/cart/{token}", h.UpdateCart)
	req := httptest.NewRequest(http.MethodPut, "/api/cart/"+token, bytes.NewReader(b))
	req.AddCookie(&http.Cookie{Name: "cart_token", Value: token})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestPutCart_EmptyPriceIDTargetsOnlyThatLine(t *testing.T) {
	kv := newInMemoryKV()
	h := handlers.NewCartHandler(kv, newFakeCatalog(), fakeVariants{})
	_ = kv.SetCart(context.Background(), &models.Cart{
		Token: "tok",
		LineItems: []models.LineItem{
			{PriceID: "price_hat", Currency: "usd", Amount: 1500, Quantity: 1},
			{PriceID: "", Name: "Legacy", Quantity: 1},
		},
	})

	w := putCart(t, h, "tok", handlers.UpdateLineItemRequest{PriceID: "", Quantity: 0})
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	items := kv.carts["tok"].LineItems
	if len(items) != 1 || items[0].PriceID != "price_hat" {
		t.Errorf("line items = %+v, want only price_hat", items)
	}
}

func TestPostCart_RepeatedAddsMergeOneLine(t *testing.T) {
	cases := []struct {
		name  string
		sizes []string
		want  string
	}{
		{"no size", []string{"", "", ""}, ""},
		{"whitespace variants", []string{"M", " M", "M "}, "M"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			kv := newInMemoryKV()
			h := handlers.NewCartHandler(kv, newFakeCatalog(), fakeVariants{})
			token := ""
			for _, size := range tc.sizes {
				w := postCart(t, h, handlers.AddToCartRequest{PriceID: "price_usd", Size: size, Quantity: 1}, token)
				if w.Code != http.StatusOK {
					t.Fatalf("status = %d, want 200", w.Code)
				}
				for _, c := range w.Result().Cookies() {
					if c.Name == "cart_token" {
						token = c.Value
					}
				}
			}
			items := kv.carts[token].LineItems
			if len(items) != 1 || items[0].Quantity != len(tc.sizes) || items[0].Size != tc.want {
				t.Errorf("line items = %+v, want one line of %d with size %q", items, len(tc.sizes), tc.want)
			}
		})
	}
}

func TestPostCart_RejectsOversizedVariant(t *testing.T) {
	kv := newInMemoryKV()
	h := handlers.NewCartHandler(kv, newFakeCatalog(), fakeVariants{})
	w := postCart(t, h, handlers.AddToCartRequest{PriceID: "price_usd", Size: strings.Repeat("x", 65), Quantity: 1}, "")
	if w.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409", w.Code)
	}
}

func TestPostCart_RejectsLineBeyondMax(t *testing.T) {
	kv := newInMemoryKV()
	h := handlers.NewCartHandler(kv, newFakeCatalog(), fakeVariants{})
	items := make([]models.LineItem, handlers.MaxCartLines)
	for i := range items {
		items[i] = models.LineItem{PriceID: "price_usd", Size: fmt.Sprintf("v%d", i), Currency: "usd", Amount: 2500, Quantity: 1}
	}
	_ = kv.SetCart(context.Background(), &models.Cart{Token: "tok", LineItems: items})

	w := postCart(t, h, handlers.AddToCartRequest{PriceID: "price_usd", Size: "new", Quantity: 1}, "tok")
	if w.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409", w.Code)
	}
	w = postCart(t, h, handlers.AddToCartRequest{PriceID: "price_usd", Size: "v0", Quantity: 1}, "tok")
	if w.Code != http.StatusOK {
		t.Fatalf("merge into existing line: status = %d, want 200", w.Code)
	}
}

func TestPostCart_VariantLabels(t *testing.T) {
	tracked := fakeVariants{rows: map[string][]store.VariantStockRow{
		"prod_1": {{Variant: "M", StockCount: 0}, {Variant: "Black / L", StockCount: 3}},
	}}
	cases := []struct {
		name     string
		variants fakeVariants
		size     string
		wantCode int
		wantSize string
	}{
		{"exact tracked", tracked, "M", http.StatusOK, "M"},
		{"case-folded to row spelling", tracked, "black / l", http.StatusOK, "Black / L"},
		{"untracked label on tracked product", tracked, "XXL", http.StatusOK, "XXL"},
		{"empty on tracked product", tracked, "", http.StatusConflict, ""},
		{"zero-width character", tracked, "M\u200b", http.StatusConflict, ""},
		{"control character", tracked, "M\x00", http.StatusConflict, ""},
		{"empty on untracked product", fakeVariants{}, "", http.StatusOK, ""},
		{"variant store failure", fakeVariants{err: errors.New("db down")}, "M", http.StatusBadGateway, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			kv := newInMemoryKV()
			h := handlers.NewCartHandler(kv, newFakeCatalog(), tc.variants)

			w := postCart(t, h, handlers.AddToCartRequest{PriceID: "price_usd", Size: tc.size, Quantity: 1}, "")
			if w.Code != tc.wantCode {
				t.Fatalf("status = %d, want %d: %s", w.Code, tc.wantCode, w.Body)
			}
			if tc.wantCode != http.StatusOK {
				if len(kv.carts) != 0 {
					t.Error("cart persisted on rejected request")
				}
				return
			}
			var cart models.Cart
			if err := json.NewDecoder(w.Body).Decode(&cart); err != nil {
				t.Fatalf("decode: %v", err)
			}
			if got := cart.LineItems[0].Size; got != tc.wantSize {
				t.Errorf("size = %q, want %q", got, tc.wantSize)
			}
		})
	}
}
