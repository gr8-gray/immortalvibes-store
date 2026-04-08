package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/immortalvibes/api/models"
	"github.com/immortalvibes/api/store"
)

// OrderStore is the interface the orders handler uses.
type OrderStore interface {
	GetOrder(ctx context.Context, id string) (*store.OrderRow, error)
}

// OrdersHandler handles order retrieval endpoints.
type OrdersHandler struct {
	db OrderStore
}

// NewOrdersHandler constructs an OrdersHandler.
func NewOrdersHandler(db OrderStore) *OrdersHandler {
	return &OrdersHandler{db: db}
}

// GetOrder handles GET /api/order/{id}.
func (h *OrdersHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	row, err := h.db.GetOrder(r.Context(), id)
	if errors.Is(err, store.ErrOrderNotFound) {
		http.Error(w, "order not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "failed to retrieve order", http.StatusInternalServerError)
		return
	}

	order := models.Order{
		ID:              row.ID,
		PaymentIntentID: row.PaymentIntentID,
		CartToken:       row.CartToken,
		Email:           row.Email,
		Currency:        row.Currency,
		TotalAmount:     row.TotalAmount,
		Status:          row.Status,
		CreatedAt:       row.CreatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}
