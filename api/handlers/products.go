package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/immortalvibes/api/models"
)

// ErrProductNotFound is returned by ProductService.GetProduct when the ID is unknown.
var ErrProductNotFound = errors.New("product not found")

// ProductService is the interface the products handler depends on.
// The real implementation (in this file) calls Stripe. Tests stub it.
type ProductService interface {
	ListProducts(ctx context.Context) ([]models.Product, error)
	GetProduct(ctx context.Context, id string) (*models.Product, error)
}

// ProductsHandler holds dependencies for product endpoints.
type ProductsHandler struct {
	svc ProductService
}

// NewProductsHandler constructs a ProductsHandler with the given service.
func NewProductsHandler(svc ProductService) *ProductsHandler {
	return &ProductsHandler{svc: svc}
}

// ListProducts handles GET /api/products.
// Returns all active Stripe products enriched with R2 image URLs and stock counts.
func (h *ProductsHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	products, err := h.svc.ListProducts(r.Context())
	if err != nil {
		http.Error(w, "failed to list products", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(products)
}

// GetProduct handles GET /api/products/{id}.
func (h *ProductsHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	product, err := h.svc.GetProduct(r.Context(), id)
	if errors.Is(err, ErrProductNotFound) {
		http.Error(w, "product not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "failed to get product", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(product)
}
