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
)

// stubStockStore implements handlers.StockStore for tests.
type stubStockStore struct {
	stock map[string]int
}

func newStubStockStore() *stubStockStore {
	return &stubStockStore{stock: map[string]int{}}
}

func (s *stubStockStore) SetStock(ctx context.Context, productID string, count int) error {
	s.stock[productID] = count
	return nil
}

func (s *stubStockStore) GetStock(ctx context.Context, productID string) (int, error) {
	return s.stock[productID], nil
}

func TestAdminSetStock(t *testing.T) {
	ss := newStubStockStore()
	h := handlers.NewAdminHandler(ss)

	r := chi.NewRouter()
	r.Put("/api/admin/products/{id}/stock", h.SetStock)

	body := handlers.SetStockRequest{Count: 25}
	b, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/api/admin/products/prod_abc/stock", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	if ss.stock["prod_abc"] != 25 {
		t.Errorf("stock = %d, want 25", ss.stock["prod_abc"])
	}
}

func TestAdminSetStock_InvalidBody(t *testing.T) {
	ss := newStubStockStore()
	h := handlers.NewAdminHandler(ss)

	r := chi.NewRouter()
	r.Put("/api/admin/products/{id}/stock", h.SetStock)

	req := httptest.NewRequest(http.MethodPut, "/api/admin/products/prod_abc/stock", bytes.NewReader([]byte(`not-json`)))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", w.Code)
	}
}

func TestAdminSetStock_NegativeCount(t *testing.T) {
	ss := newStubStockStore()
	h := handlers.NewAdminHandler(ss)

	r := chi.NewRouter()
	r.Put("/api/admin/products/{id}/stock", h.SetStock)

	body := handlers.SetStockRequest{Count: -1}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPut, "/api/admin/products/prod_abc/stock", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", w.Code)
	}
}
