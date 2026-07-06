package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/immortalvibes/api/models"
	"github.com/immortalvibes/api/store"
	stripe "github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/coupon"
	"github.com/stripe/stripe-go/v76/paymentintent"
	"github.com/stripe/stripe-go/v76/price"
	"github.com/stripe/stripe-go/v76/promotioncode"
)

// enrichSizes resolves each line item's variant/size from its Stripe price
// nickname, best-effort — a failed lookup leaves Size empty and never blocks
// checkout. Mutates the slice in place.
// Only fills Size when it is already empty (legacy carts); never overwrites a
// frontend-provided value.
func enrichSizes(items []models.LineItem) {
	for i := range items {
		if items[i].Size != "" {
			continue // frontend already set the size — trust it
		}
		if items[i].PriceID == "" {
			continue
		}
		p, err := price.Get(items[i].PriceID, nil)
		if err == nil && p != nil {
			items[i].Size = p.Nickname
		}
	}
}

// manifestSummary builds a compact one-line summary for Stripe PI metadata,
// e.g. "2x Immortal Light Sweatpants (L); 1x Tee (M)". Capped at 480 chars
// (Stripe metadata values max 500).
func manifestSummary(items []models.LineItem) string {
	parts := make([]string, 0, len(items))
	for _, li := range items {
		s := fmt.Sprintf("%dx %s", li.Quantity, li.Name)
		if li.Size != "" {
			s += " (" + li.Size + ")"
		}
		parts = append(parts, s)
	}
	out := strings.Join(parts, "; ")
	if len(out) > 480 {
		out = out[:477] + "..."
	}
	return out
}

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
	CartToken    string `json:"cart_token"`
	Email        string `json:"email"`
	ShippingName string `json:"shipping_name"`
	Line1        string `json:"line1"`
	Line2        string `json:"line2"`
	City         string `json:"city"`
	State        string `json:"state"`
	PostalCode   string `json:"postal_code"`
	Country      string `json:"country"`
	DiscountCode string `json:"discount_code,omitempty"`
	ShippingCost int    `json:"shipping_cost,omitempty"`
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

	if req.ShippingName == "" || req.Line1 == "" || req.City == "" || req.State == "" || req.PostalCode == "" || req.Country == "" {
		http.Error(w, "shipping address required", http.StatusBadRequest)
		return
	}

	// Resolve sizes so the persisted manifest is human-readable (STOP 18).
	enrichSizes(cart.LineItems)

	// Variant stock guard: reject checkout if any line item's variant stock is
	// known and insufficient. Non-atomic (no reservation) — same oversell window
	// as before, but blocks the obvious case.
	for _, li := range cart.LineItems {
		if li.ProductID == "" || li.Size == "" {
			continue
		}
		variantRows, err := h.db.GetVariantStocks(r.Context(), li.ProductID)
		if err != nil {
			continue // DB error — don't block checkout
		}
		for _, v := range variantRows {
			if v.Variant == li.Size {
				if v.StockCount < li.Quantity {
					http.Error(w,
						fmt.Sprintf("%s (%s) has insufficient stock", li.Name, li.Size),
						http.StatusConflict,
					)
					return
				}
				break
			}
		}
	}

	currency := DetectCurrency(r)
	total := cart.Total()

	// Add shipping cost to total.
	total += int64(req.ShippingCost)

	// Apply discount if a code was provided.
	var appliedCouponID string
	if req.DiscountCode != "" {
		resolvedCouponID, discountedTotal, ok := resolveDiscount(req.DiscountCode, total)
		if ok {
			appliedCouponID = resolvedCouponID
			total = discountedTotal
		}
		// If the code is invalid we silently continue at full price —
		// the frontend validates first; this is just a safety net.
	}

	piParams := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(total),
		Currency: stripe.String(currency),
		Metadata: map[string]string{
			"cart_token":    req.CartToken,
			"email":         req.Email,
			"discount_code": req.DiscountCode,
			"coupon_id":     appliedCouponID,
			// Manifest summary so the owner can see contents on the Stripe
			// dashboard and it survives in the webhook event (STOP 18).
			"items": manifestSummary(cart.LineItems),
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
		ShippingName:    req.ShippingName,
		Line1:           req.Line1,
		Line2:           req.Line2,
		City:            req.City,
		State:           req.State,
		PostalCode:      req.PostalCode,
		Country:         req.Country,
		LineItems:       cart.LineItems,
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

// resolveDiscount tries to resolve a human-readable promo code or direct
// coupon ID via Stripe, then returns the coupon ID, discounted total, and ok.
func resolveDiscount(code string, total int64) (couponID string, discountedTotal int64, ok bool) {
	// Try PromotionCode first.
	pcParams := &stripe.PromotionCodeListParams{}
	pcParams.Filters.AddFilter("code", "", code)
	pcParams.Filters.AddFilter("active", "", "true")
	pcParams.Filters.AddFilter("limit", "", "1")
	iter := promotioncode.List(pcParams)
	for iter.Next() {
		pc := iter.PromotionCode()
		if pc.Coupon != nil {
			return applyStripeCoupon(pc.Coupon, total, pc.Coupon.ID)
		}
	}
	if iter.Err() != nil {
		return "", total, false
	}

	// Fall back to direct Coupon ID.
	c, err := coupon.Get(code, nil)
	if err != nil || !c.Valid {
		return "", total, false
	}
	return applyStripeCoupon(c, total, c.ID)
}

func applyStripeCoupon(c *stripe.Coupon, total int64, id string) (string, int64, bool) {
	if c == nil {
		return "", total, false
	}
	if c.PercentOff > 0 {
		discount := int64(float64(total) * float64(c.PercentOff) / 100.0)
		return id, total - discount, true
	}
	if c.AmountOff > 0 {
		discounted := total - c.AmountOff
		if discounted < 0 {
			discounted = 0
		}
		return id, discounted, true
	}
	return "", total, false
}
