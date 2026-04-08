package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// StockStore is the interface the admin handler uses for stock updates.
type StockStore interface {
	SetStock(ctx context.Context, productID string, count int) error
	GetStock(ctx context.Context, productID string) (int, error)
}

// AdminHandler handles admin-only endpoints.
type AdminHandler struct {
	stock StockStore
}

// NewAdminHandler constructs an AdminHandler.
func NewAdminHandler(stock StockStore) *AdminHandler {
	return &AdminHandler{stock: stock}
}

// SetStockRequest is the JSON body for PUT /api/admin/products/:id/stock.
type SetStockRequest struct {
	Count int `json:"count"`
}

// SetStockResponse is the response body.
type SetStockResponse struct {
	ProductID string `json:"product_id"`
	Count     int    `json:"count"`
}

// SetStock handles PUT /api/admin/products/{id}/stock.
func (h *AdminHandler) SetStock(w http.ResponseWriter, r *http.Request) {
	productID := chi.URLParam(r, "id")

	var req SetStockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.Count < 0 {
		http.Error(w, "count must be >= 0", http.StatusBadRequest)
		return
	}

	if err := h.stock.SetStock(r.Context(), productID, req.Count); err != nil {
		http.Error(w, "failed to set stock", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(SetStockResponse{
		ProductID: productID,
		Count:     req.Count,
	})
}
