package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/immortalvibes/api/store"
)

// StockStore is the interface the admin handler uses for stock updates.
type StockStore interface {
	SetStock(ctx context.Context, productID string, count int) error
	GetStock(ctx context.Context, productID string) (int, error)
	SetVariantStock(ctx context.Context, productID, variant string, count int) error
	GetVariantStocks(ctx context.Context, productID string) ([]store.VariantStockRow, error)
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
// When Variant is set, upserts a variant row; otherwise upserts the legacy total.
type SetStockRequest struct {
	Count   int    `json:"count"`
	Variant string `json:"variant,omitempty"`
}

// SetStockResponse is the response body.
type SetStockResponse struct {
	ProductID string `json:"product_id"`
	Variant   string `json:"variant,omitempty"`
	Count     int    `json:"count"`
}

// GetStockResponse is the response body for GET /api/admin/products/{id}/stock.
type GetStockResponse struct {
	ProductID string                  `json:"product_id"`
	Total     int                     `json:"total"`
	Variants  []store.VariantStockRow `json:"variants"`
}

// SetStock handles PUT /api/admin/products/{id}/stock.
// Body with "variant" upserts a variant row (per-size/colorway).
// Body without "variant" upserts the legacy product_stock total (back-compat).
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

	if req.Variant != "" {
		if err := h.stock.SetVariantStock(r.Context(), productID, req.Variant, req.Count); err != nil {
			http.Error(w, "failed to set variant stock", http.StatusInternalServerError)
			return
		}
	} else {
		if err := h.stock.SetStock(r.Context(), productID, req.Count); err != nil {
			http.Error(w, "failed to set stock", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(SetStockResponse{
		ProductID: productID,
		Variant:   req.Variant,
		Count:     req.Count,
	})
}

// GetProductStock handles GET /api/admin/products/{id}/stock.
// Returns all variant rows + the current total (which may be a SUM of variants).
func (h *AdminHandler) GetProductStock(w http.ResponseWriter, r *http.Request) {
	productID := chi.URLParam(r, "id")

	total, err := h.stock.GetStock(r.Context(), productID)
	if err != nil {
		http.Error(w, "failed to get stock", http.StatusInternalServerError)
		return
	}

	variants, err := h.stock.GetVariantStocks(r.Context(), productID)
	if err != nil {
		http.Error(w, "failed to get variant stocks", http.StatusInternalServerError)
		return
	}
	if variants == nil {
		variants = []store.VariantStockRow{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(GetStockResponse{
		ProductID: productID,
		Total:     total,
		Variants:  variants,
	})
}
