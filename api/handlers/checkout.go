package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/immortalvibes/api/models"
	"github.com/immortalvibes/api/shippo"
	"github.com/immortalvibes/api/store"
	stripe "github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/coupon"
	"github.com/stripe/stripe-go/v76/paymentintent"
	"github.com/stripe/stripe-go/v76/promotioncode"
)

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
	// ShippingCost is accepted for backward compatibility with older
	// storefront builds but is NEVER used to price the order — the server
	// recomputes shipping itself (see quoteShipping). A mismatch is logged.
	ShippingCost int `json:"shipping_cost,omitempty"`
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

// CheckoutStore is the subset of store.DB needed by CheckoutHandler.
type CheckoutStore interface {
	GetVariantStocks(ctx context.Context, productID string) ([]store.VariantStockRow, error)
	SaveOrder(ctx context.Context, o store.OrderRow) error
}

// CheckoutHandler handles POST /api/checkout.
type CheckoutHandler struct {
	stripeKey string
	kv        CheckoutKV
	db        CheckoutStore
	catalog   PriceCatalog
	shipping  ShippoEstimator
	// newPaymentIntent is paymentintent.New in production; overridable in tests.
	newPaymentIntent func(*stripe.PaymentIntentParams) (*stripe.PaymentIntent, error)
}

// NewCheckoutHandler constructs a CheckoutHandler. catalog is the source of
// truth for line item pricing at checkout. shipping prices shipping
// server-side; it must be the same estimator the storefront's
// /api/shipping/estimate uses.
func NewCheckoutHandler(stripeKey string, kv CheckoutKV, db CheckoutStore, catalog PriceCatalog, shipping ShippoEstimator) *CheckoutHandler {
	stripe.Key = stripeKey
	return &CheckoutHandler{
		stripeKey:        stripeKey,
		kv:               kv,
		db:               db,
		catalog:          catalog,
		shipping:         shipping,
		newPaymentIntent: paymentintent.New,
	}
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

	// Re-resolve every line item against the catalog. Stored cart amounts are
	// never trusted; this also fills sizes for the persisted manifest (STOP 18).
	if _, err := repriceLineItems(r.Context(), h.catalog, cart.LineItems); err != nil {
		writePricingError(w, err)
		return
	}

	// Variant stock guard: for products that track variants, require a
	// variant, canonicalize it against the stock rows, and reject checkout if
	// its stock is insufficient. A variant with no stock row is untracked and
	// sellable, matching the storefront. Non-atomic (no reservation) — same
	// oversell window as before, but blocks the obvious case.
	for i := range cart.LineItems {
		li := &cart.LineItems[i]
		variantRows, err := h.db.GetVariantStocks(r.Context(), li.ProductID)
		if err != nil {
			continue // DB error — don't block checkout
		}
		size, err := canonicalVariant(variantRows, li.Size)
		if err != nil {
			writePricingError(w, &LineItemError{Name: truncateName(li.Name), Err: err})
			return
		}
		li.Size = size
		row, ok := findVariant(variantRows, li.Size)
		if ok && row.StockCount < li.Quantity {
			http.Error(w,
				fmt.Sprintf("%s (%s) has insufficient stock", li.Name, li.Size),
				http.StatusConflict,
			)
			return
		}
	}

	// Quote shipping with the same logic as /api/shipping/estimate. The
	// request's shipping_cost is informational only. If no rate is available
	// for the address, checkout fails rather than proceeding without shipping.
	shippingCents, err := h.serverShippingCents(r.Context(), req)
	if err != nil {
		log.Printf("checkout: shipping quote failed: to=%s,%s,%s err=%v", req.City, req.State, req.PostalCode, err)
		http.Error(w, "unable to calculate shipping for this address; please check the address and try again", http.StatusBadGateway)
		return
	}
	if req.ShippingCost != 0 && int64(req.ShippingCost) != shippingCents {
		log.Printf("checkout: ignoring client shipping_cost=%d (server=%d) cart=%s", req.ShippingCost, shippingCents, req.CartToken)
	}

	// Currency behavior is unchanged: shipping cents are added to the cart
	// total exactly as the client-supplied value was before.
	currency := DetectCurrency(r)
	total := cart.Total()

	// Add shipping cost to total.
	total += shippingCents

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
	pi, err := h.newPaymentIntent(piParams)
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

// serverShippingCents quotes shipping for the checkout's address using the
// same address mapping and pricing as ShippingHandler.Estimate.
func (h *CheckoutHandler) serverShippingCents(ctx context.Context, req CheckoutRequest) (int64, error) {
	to := shippo.Address{
		Name:    req.ShippingName,
		Street1: req.Line1,
		City:    req.City,
		State:   req.State,
		Zip:     req.PostalCode,
		Country: req.Country,
	}
	_, cents, err := quoteShipping(ctx, h.shipping, to)
	return cents, err
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
