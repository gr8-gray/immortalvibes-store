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
	"github.com/immortalvibes/api/store"
)

// stubStockStore implements handlers.StockStore for tests.
type stubStockStore struct {
	stock         map[string]int
	variantStocks map[string]map[string]int // productID → variant → count
}

func newStubStockStore() *stubStockStore {
	return &stubStockStore{
		stock:         map[string]int{},
		variantStocks: map[string]map[string]int{},
	}
}

func (s *stubStockStore) SetStock(ctx context.Context, productID string, count int) error {
	s.stock[productID] = count
	return nil
}

func (s *stubStockStore) GetStock(ctx context.Context, productID string) (int, error) {
	return s.stock[productID], nil
}

func (s *stubStockStore) SetVariantStock(ctx context.Context, productID, variant string, count int) error {
	if s.variantStocks[productID] == nil {
		s.variantStocks[productID] = map[string]int{}
	}
	s.variantStocks[productID][variant] = count
	return nil
}

func (s *stubStockStore) GetVariantStocks(ctx context.Context, productID string) ([]store.VariantStockRow, error) {
	var out []store.VariantStockRow
	for v, c := range s.variantStocks[productID] {
		out = append(out, store.VariantStockRow{Variant: v, StockCount: c})
	}
	return out, nil
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

// TestAdminSetVariantStock verifies PUT with "variant" field upserts a variant row.
func TestAdminSetVariantStock(t *testing.T) {
	ss := newStubStockStore()
	h := handlers.NewAdminHandler(ss)

	r := chi.NewRouter()
	r.Put("/api/admin/products/{id}/stock", h.SetStock)

	body := handlers.SetStockRequest{Count: 3, Variant: "L"}
	b, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/api/admin/products/prod_xyz/stock", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if ss.variantStocks["prod_xyz"]["L"] != 3 {
		t.Errorf("variant L stock = %d, want 3", ss.variantStocks["prod_xyz"]["L"])
	}
	// Legacy product_stock must not be touched.
	if ss.stock["prod_xyz"] != 0 {
		t.Errorf("legacy stock = %d, want 0 (untouched)", ss.stock["prod_xyz"])
	}
}

// TestAdminGetProductStock verifies GET returns all variant rows + total.
func TestAdminGetProductStock(t *testing.T) {
	ss := newStubStockStore()
	_ = ss.SetVariantStock(context.Background(), "prod_xyz", "S", 2)
	_ = ss.SetVariantStock(context.Background(), "prod_xyz", "M", 5)
	ss.stock["prod_xyz"] = 7 // legacy total

	h := handlers.NewAdminHandler(ss)

	r := chi.NewRouter()
	r.Get("/api/admin/products/{id}/stock", h.GetProductStock)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/products/prod_xyz/stock", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var resp handlers.GetStockResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.ProductID != "prod_xyz" {
		t.Errorf("product_id = %q, want prod_xyz", resp.ProductID)
	}
	if len(resp.Variants) != 2 {
		t.Errorf("variants len = %d, want 2", len(resp.Variants))
	}
}
