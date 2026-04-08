package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/immortalvibes/api/handlers"
	"github.com/immortalvibes/api/models"
)

// stubProductService implements handlers.ProductService for tests.
type stubProductService struct {
	products []models.Product
	err      error
}

func (s *stubProductService) ListProducts(ctx context.Context) ([]models.Product, error) {
	return s.products, s.err
}

func (s *stubProductService) GetProduct(ctx context.Context, id string) (*models.Product, error) {
	for _, p := range s.products {
		if p.ID == id {
			return &p, nil
		}
	}
	return nil, handlers.ErrProductNotFound
}

func TestListProducts(t *testing.T) {
	svc := &stubProductService{
		products: []models.Product{
			{
				ID:       "prod_1",
				Name:     "Tee",
				ImageURL: "https://r2.example.com/tee.jpg",
				Prices: []models.Price{
					{PriceID: "price_usd", Currency: "usd", Amount: 2500},
					{PriceID: "price_gbp", Currency: "gbp", Amount: 2000},
				},
				StockCount: 5,
			},
		},
	}

	h := handlers.NewProductsHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/products", nil)
	w := httptest.NewRecorder()
	h.ListProducts(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	var products []models.Product
	if err := json.NewDecoder(w.Body).Decode(&products); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(products) != 1 {
		t.Fatalf("got %d products, want 1", len(products))
	}
	if products[0].ID != "prod_1" {
		t.Errorf("got ID %q, want prod_1", products[0].ID)
	}
}

func TestGetProduct(t *testing.T) {
	svc := &stubProductService{
		products: []models.Product{
			{ID: "prod_1", Name: "Tee", StockCount: 3},
		},
	}

	h := handlers.NewProductsHandler(svc)

	r := chi.NewRouter()
	r.Get("/api/products/{id}", h.GetProduct)

	req := httptest.NewRequest(http.MethodGet, "/api/products/prod_1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	var product models.Product
	if err := json.NewDecoder(w.Body).Decode(&product); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if product.ID != "prod_1" {
		t.Errorf("got ID %q, want prod_1", product.ID)
	}
}

func TestGetProduct_NotFound(t *testing.T) {
	svc := &stubProductService{products: nil}
	h := handlers.NewProductsHandler(svc)

	r := chi.NewRouter()
	r.Get("/api/products/{id}", h.GetProduct)

	req := httptest.NewRequest(http.MethodGet, "/api/products/no-such", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", w.Code)
	}
}
