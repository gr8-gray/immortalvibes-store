package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/immortalvibes/api/models"
	"github.com/immortalvibes/api/store"
)

// CartKV is the interface the cart handler uses for KV access.
type CartKV interface {
	GetCart(ctx context.Context, token string) (*models.Cart, error)
	SetCart(ctx context.Context, cart *models.Cart) error
	DeleteCart(ctx context.Context, token string) error
}

// CartHandler handles cart CRUD endpoints.
type CartHandler struct {
	kv CartKV
}

// NewCartHandler constructs a CartHandler with the given KV client.
func NewCartHandler(kv CartKV) *CartHandler {
	return &CartHandler{kv: kv}
}

// AddToCartRequest is the JSON body for POST /api/cart.
type AddToCartRequest struct {
	PriceID   string `json:"price_id"`
	ProductID string `json:"product_id"`
	Name      string `json:"name"`
	ImageURL  string `json:"image_url"`
	Currency  string `json:"currency"`
	Amount    int64  `json:"amount"`
	Quantity  int    `json:"quantity"`
}

// UpdateLineItemRequest is the JSON body for PUT /api/cart/{token}.
type UpdateLineItemRequest struct {
	PriceID  string `json:"price_id"`
	Quantity int    `json:"quantity"`
}

// GetCart handles GET /api/cart/{token}.
func (h *CartHandler) GetCart(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	cart, err := h.kv.GetCart(r.Context(), token)
	if errors.Is(err, store.ErrCartNotFound) {
		http.Error(w, "cart not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "failed to get cart", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cart)
}

// AddToCart handles POST /api/cart.
// Reads cart_token cookie; creates a new cart if not found.
func (h *CartHandler) AddToCart(w http.ResponseWriter, r *http.Request) {
	var req AddToCartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	token := ""
	if c, err := r.Cookie("cart_token"); err == nil {
		token = c.Value
	}

	var cart *models.Cart
	if token != "" {
		existing, err := h.kv.GetCart(r.Context(), token)
		if err == nil {
			cart = existing
		}
	}
	if cart == nil {
		token = uuid.New().String()
		cart = &models.Cart{Token: token, LineItems: []models.LineItem{}}
	}

	// Merge: if same PriceID exists, increment quantity; else append.
	found := false
	for i, li := range cart.LineItems {
		if li.PriceID == req.PriceID {
			cart.LineItems[i].Quantity += req.Quantity
			found = true
			break
		}
	}
	if !found {
		cart.LineItems = append(cart.LineItems, models.LineItem{
			PriceID:   req.PriceID,
			ProductID: req.ProductID,
			Name:      req.Name,
			ImageURL:  req.ImageURL,
			Currency:  req.Currency,
			Amount:    req.Amount,
			Quantity:  req.Quantity,
		})
	}

	if err := h.kv.SetCart(r.Context(), cart); err != nil {
		http.Error(w, "failed to save cart", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "cart_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(7 * 24 * time.Hour),
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cart)
}

// UpdateCart handles PUT /api/cart/{token}.
// Sets the quantity for a specific price_id. Quantity 0 removes the item.
func (h *CartHandler) UpdateCart(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")

	// Verify cookie matches token to prevent cross-cart tampering.
	cookieTok := ""
	if c, err := r.Cookie("cart_token"); err == nil {
		cookieTok = c.Value
	}
	if cookieTok != token {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	cart, err := h.kv.GetCart(r.Context(), token)
	if errors.Is(err, store.ErrCartNotFound) {
		http.Error(w, "cart not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "failed to get cart", http.StatusInternalServerError)
		return
	}

	var req UpdateLineItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	priceIDToUpdate := req.PriceID
	// If no price_id in body, update the first item (convenience for single-item carts).
	if priceIDToUpdate == "" && len(cart.LineItems) > 0 {
		priceIDToUpdate = cart.LineItems[0].PriceID
	}

	updated := cart.LineItems[:0]
	for _, li := range cart.LineItems {
		if li.PriceID == priceIDToUpdate {
			if req.Quantity > 0 {
				li.Quantity = req.Quantity
				updated = append(updated, li)
			}
			// qty == 0 means remove: don't append
		} else {
			updated = append(updated, li)
		}
	}
	cart.LineItems = updated

	if err := h.kv.SetCart(r.Context(), cart); err != nil {
		http.Error(w, "failed to save cart", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cart)
}
