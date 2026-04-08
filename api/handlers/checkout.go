package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/immortalvibes/api/models"
	"github.com/immortalvibes/api/store"
	stripe "github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/paymentintent"
)

// eurCountries is the set of ISO country codes that map to EUR.
var eurCountries = map[string]bool{
	"AT": true, "BE": true, "CY": true, "EE": true, "FI": true,
	"FR": true, "DE": true, "GR": true, "IE": true, "IT": true,
	"LV": true, "LT": true, "LU": true, "MT": true, "NL": true,
	"PT": true, "SK": true, "SI": true, "ES": true,
}

// audCountries maps to AUD.
var audCountries = map[string]bool{
	"AU": true, "NZ": true,
}

// DetectCurrency returns the ISO currency code (lowercase) based on the
// CF-IPCountry header. Defaults to "usd" for unknown or missing country.
func DetectCurrency(r *http.Request) string {
	country := r.Header.Get("CF-IPCountry")
	if country == "GB" {
		return "gbp"
	}
	if audCountries[country] {
		return "aud"
	}
	if eurCountries[country] {
		return "eur"
	}
	return "usd"
}

// CheckoutRequest is the JSON body for POST /api/checkout.
type CheckoutRequest struct {
	CartToken string `json:"cart_token"`
	Email     string `json:"email"`
}

// CheckoutResponse is returned to the SvelteKit frontend.
type CheckoutResponse struct {
	ClientSecret string `json:"client_secret"`
	OrderID      string `json:"order_id"`
	Currency     string `json:"currency"`
	TotalAmount  int64  `json:"total_amount"`
}

// CheckoutKV is the subset of CartKV needed by CheckoutHandler.
type CheckoutKV interface {
	GetCart(ctx context.Context, token string) (*models.Cart, error)
}

// CheckoutHandler handles POST /api/checkout.
type CheckoutHandler struct {
	stripeKey string
	kv        CheckoutKV
	db        *store.DB
}

// NewCheckoutHandler constructs a CheckoutHandler.
func NewCheckoutHandler(stripeKey string, kv CheckoutKV, db *store.DB) *CheckoutHandler {
	stripe.Key = stripeKey
	return &CheckoutHandler{stripeKey: stripeKey, kv: kv, db: db}
}

// Checkout handles POST /api/checkout.
// Creates a Stripe PaymentIntent and saves a pending order in Postgres.
func (h *CheckoutHandler) Checkout(w http.ResponseWriter, r *http.Request) {
	var req CheckoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.CartToken == "" {
		// Try cookie fallback
		if c, err := r.Cookie("cart_token"); err == nil {
			req.CartToken = c.Value
		}
	}
	if req.CartToken == "" {
		http.Error(w, "cart_token required", http.StatusBadRequest)
		return
	}

	cart, err := h.kv.GetCart(r.Context(), req.CartToken)
	if errors.Is(err, store.ErrCartNotFound) {
		http.Error(w, "cart not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "failed to retrieve cart", http.StatusInternalServerError)
		return
	}

	if len(cart.LineItems) == 0 {
		http.Error(w, "cart is empty", http.StatusBadRequest)
		return
	}

	currency := DetectCurrency(r)
	total := cart.Total()

	piParams := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(total),
		Currency: stripe.String(currency),
		Metadata: map[string]string{
			"cart_token": req.CartToken,
			"email":      req.Email,
		},
	}
	pi, err := paymentintent.New(piParams)
	if err != nil {
		http.Error(w, "failed to create payment intent", http.StatusInternalServerError)
		return
	}

	orderID := uuid.New().String()
	if err := h.db.SaveOrder(r.Context(), store.OrderRow{
		ID:              orderID,
		PaymentIntentID: pi.ID,
		CartToken:       req.CartToken,
		Email:           req.Email,
		Currency:        currency,
		TotalAmount:     total,
		Status:          "pending",
	}); err != nil {
		http.Error(w, "failed to save order", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(CheckoutResponse{
		ClientSecret: pi.ClientSecret,
		OrderID:      orderID,
		Currency:     currency,
		TotalAmount:  total,
	})
}
