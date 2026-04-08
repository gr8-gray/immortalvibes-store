package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

func TestGetCart_NoToken(t *testing.T) {
	kv := newInMemoryKV()
	h := handlers.NewCartHandler(kv)

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
	h := handlers.NewCartHandler(kv)

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
	h := handlers.NewCartHandler(kv)

	// Seed a cart
	_ = kv.SetCart(context.Background(), &models.Cart{
		Token: "existing-tok",
		LineItems: []models.LineItem{
			{PriceID: "price_gbp", ProductID: "prod_1", Name: "Tee", Currency: "gbp", Amount: 2000, Quantity: 1},
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
	h := handlers.NewCartHandler(kv)

	_ = kv.SetCart(context.Background(), &models.Cart{
		Token: "upd-tok",
		LineItems: []models.LineItem{
			{PriceID: "price_usd", ProductID: "prod_1", Name: "Tee", Currency: "usd", Amount: 2500, Quantity: 1},
		},
	})

	body := handlers.UpdateLineItemRequest{Quantity: 3}
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
