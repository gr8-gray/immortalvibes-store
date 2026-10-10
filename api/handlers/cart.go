package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
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
	kv       CartKV
	catalog  PriceCatalog
	variants VariantReader
}

// NewCartHandler constructs a CartHandler. catalog is the source of truth for
// line item pricing; client-supplied amounts are never stored. variants
// canonicalizes variant labels against stock rows.
func NewCartHandler(kv CartKV, catalog PriceCatalog, variants VariantReader) *CartHandler {
	return &CartHandler{kv: kv, catalog: catalog, variants: variants}
}

// AddToCartRequest is the JSON body for POST /api/cart. Name, Currency and
// Amount are accepted for compatibility with existing clients but ignored;
// they are resolved from the price catalog.
type AddToCartRequest struct {
	PriceID   string `json:"price_id"`
	ProductID string `json:"product_id"`
	Name      string `json:"name"`
	ImageURL  string `json:"image_url"`
	Currency  string `json:"currency"`
	Amount    int64  `json:"amount"`
	Quantity  int    `json:"quantity"`
	Size      string `json:"size,omitempty"` // variant label (size or colorway); empty for OS products
}

// UpdateLineItemRequest is the JSON body for PUT /api/cart/{token}.
type UpdateLineItemRequest struct {
	PriceID  string `json:"price_id"`
	Size     string `json:"size,omitempty"` // must match the size used at add-to-cart time
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

// GetCurrentCart handles GET /api/cart — returns the cart identified by the
// cart_token cookie. Used by the frontend to hydrate the cart store on page load
// (the cookie is HttpOnly so JS can't read the token directly).
// Returns an empty cart with empty token if no cookie or cart not found.
func (h *CartHandler) GetCurrentCart(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	c, err := r.Cookie("cart_token")
	if err != nil || c.Value == "" {
		json.NewEncoder(w).Encode(&models.Cart{Token: "", LineItems: []models.LineItem{}})
		return
	}

	cart, err := h.kv.GetCart(r.Context(), c.Value)
	if errors.Is(err, store.ErrCartNotFound) {
		json.NewEncoder(w).Encode(&models.Cart{Token: "", LineItems: []models.LineItem{}})
		return
	}
	if err != nil {
		http.Error(w, "failed to get cart", http.StatusInternalServerError)
		return
	}
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
	if req.PriceID == "" {
		http.Error(w, "price_id required", http.StatusBadRequest)
		return
	}
	if !validQuantity(req.Quantity) {
		writePricingError(w, ErrInvalidQuantity)
		return
	}

	size, err := normalizeVariant(req.Size)
	if err != nil {
		writePricingError(w, err)
		return
	}

	cp, err := h.catalog.Lookup(r.Context(), req.PriceID)
	if err != nil {
		writePricingError(w, err)
		return
	}
	rows, err := h.variants.GetVariantStocks(r.Context(), cp.ProductID)
	if err != nil {
		writePricingError(w, fmt.Errorf("variant stocks for %s: %w", cp.ProductID, err))
		return
	}
	if size, err = canonicalVariant(rows, size); err != nil {
		writePricingError(w, err)
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

	// Merge: if same PriceID+Size exists, increment quantity; else append.
	// Dedup key is price_id:size so different sizes are distinct line items.
	found := false
	for i := range cart.LineItems {
		li := &cart.LineItems[i]
		if li.PriceID == req.PriceID && li.Size == size {
			if !validQuantity(li.Quantity + req.Quantity) {
				writePricingError(w, ErrInvalidQuantity)
				return
			}
			li.Quantity += req.Quantity
			applyCatalogPrice(li, cp)
			found = true
			break
		}
	}
	if !found {
		if len(cart.LineItems) >= MaxCartLines {
			writePricingError(w, ErrCartFull)
			return
		}
		for _, li := range cart.LineItems {
			if li.Currency != "" && li.Currency != cp.Currency {
				writePricingError(w, ErrMixedCurrency)
				return
			}
		}
		li := models.LineItem{
			PriceID:  req.PriceID,
			ImageURL: req.ImageURL,
			Quantity: req.Quantity,
			Size:     size,
		}
		applyCatalogPrice(&li, cp)
		cart.LineItems = append(cart.LineItems, li)
	}

	if err := h.kv.SetCart(r.Context(), cart); err != nil {
		log.Printf("cart: SetCart failed: %v", err)
		http.Error(w, "failed to save cart: "+err.Error(), http.StatusInternalServerError)
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

	// Lines are matched exactly on (price_id, size), including empty values, so
	// a line with a missing price_id can still be removed.
	if req.Quantity > MaxLineQuantity {
		writePricingError(w, ErrInvalidQuantity)
		return
	}

	updated := cart.LineItems[:0]
	for _, li := range cart.LineItems {
		if li.PriceID == req.PriceID && li.Size == req.Size {
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
		log.Printf("cart: SetCart failed: %v", err)
		http.Error(w, "failed to save cart: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cart)
}

// writePricingError maps catalog and line item validation errors to HTTP
// responses. Errors attributed to a stored line item name it, so the buyer
// knows which item to remove. Unclassified errors are upstream failures and
// are logged.
func writePricingError(w http.ResponseWriter, err error) {
	var msg string
	status := http.StatusConflict
	switch {
	case errors.Is(err, ErrInvalidQuantity):
		msg, status = "invalid quantity", http.StatusBadRequest
	case errors.Is(err, ErrInvalidVariant):
		msg = "selected option is unavailable"
	case errors.Is(err, ErrPriceUnavailable):
		msg = "item unavailable"
	case errors.Is(err, ErrMixedCurrency):
		msg = "item cannot be combined with cart contents"
	case errors.Is(err, ErrCartFull):
		msg = "too many items in cart; remove some and try again"
	default:
		log.Printf("pricing: lookup failed: %q", err.Error())
		http.Error(w, "pricing temporarily unavailable", http.StatusBadGateway)
		return
	}
	var lie *LineItemError
	if errors.As(err, &lie) && lie.Name != "" {
		msg = lie.Name + ": " + msg
	}
	http.Error(w, msg, status)
}
