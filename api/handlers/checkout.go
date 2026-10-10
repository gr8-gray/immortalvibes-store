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
	"github.com/stripe/stripe-go/v76/paymentintent"
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
	discounts DiscountResolver
	// newPaymentIntent is paymentintent.New in production; overridable in tests.
	newPaymentIntent func(*stripe.PaymentIntentParams) (*stripe.PaymentIntent, error)
}

// NewCheckoutHandler constructs a CheckoutHandler. catalog is the source of
// truth for line item pricing at checkout. shipping prices shipping
// server-side; it must be the same estimator the storefront's
// /api/shipping/estimate uses. discounts resolves promotion codes.
func NewCheckoutHandler(stripeKey string, kv CheckoutKV, db CheckoutStore, catalog PriceCatalog, shipping ShippoEstimator, discounts DiscountResolver) *CheckoutHandler {
	stripe.Key = stripeKey
	return &CheckoutHandler{
		stripeKey:        stripeKey,
		kv:               kv,
		db:               db,
		catalog:          catalog,
		shipping:         shipping,
		discounts:        discounts,
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
	currency, err := repriceLineItems(r.Context(), h.catalog, cart.LineItems)
	if err != nil {
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
	shippingCents, err := h.serverShippingCents(r.Context(), req, currency)
	if err != nil {
		log.Printf("checkout: shipping quote failed: to=%s,%s,%s err=%v", req.City, req.State, req.PostalCode, err)
		http.Error(w, "unable to calculate shipping for this address; please check the address and try again", http.StatusBadGateway)
		return
	}
	if req.ShippingCost != 0 && int64(req.ShippingCost) != shippingCents {
		log.Printf("checkout: ignoring client shipping_cost=%d (server=%d) cart=%s", req.ShippingCost, shippingCents, req.CartToken)
	}

	// The order is charged in the catalog currency of its line items, the
	// currency every displayed price is in.
	total := cart.Total() + shippingCents

	var discount Discount
	if req.DiscountCode != "" {
		discount, err = h.discounts.Resolve(r.Context(), req.DiscountCode)
		if err != nil {
			writeDiscountError(w, err)
			return
		}
		off, err := discount.Amount(cart.LineItems, shippingCents, currency)
		if err != nil {
			writeDiscountError(w, err)
			return
		}
		total -= off
	}

	piParams := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(total),
		Currency: stripe.String(currency),
		Metadata: map[string]string{
			"cart_token": req.CartToken,
			"email":      req.Email,
			// coupon_id and promotion_code_id are counted to enforce
			// redemption limits (StripeRedemptionCounter).
			"discount_code":     discount.Code,
			"coupon_id":         discount.CouponID,
			"promotion_code_id": discount.PromotionCodeID,
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
// same address mapping and pricing as ShippingHandler.Estimate. The quote must
// be in the order currency.
func (h *CheckoutHandler) serverShippingCents(ctx context.Context, req CheckoutRequest, currency string) (int64, error) {
	to := shippo.Address{
		Name:    req.ShippingName,
		Street1: req.Line1,
		City:    req.City,
		State:   req.State,
		Zip:     req.PostalCode,
		Country: req.Country,
	}
	est, cents, err := quoteShipping(ctx, h.shipping, to)
	if err != nil {
		return 0, err
	}
	if !strings.EqualFold(est.Currency, currency) {
		return 0, fmt.Errorf("shipping quote currency %q does not match order currency %q", est.Currency, currency)
	}
	return cents, nil
}

// writeDiscountError maps discount resolution errors to HTTP responses.
// Unclassified errors are upstream failures and are logged.
func writeDiscountError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrInvalidDiscount):
		http.Error(w, "discount code is not valid", http.StatusConflict)
	case errors.Is(err, ErrDiscountExhausted):
		http.Error(w, "discount code is no longer available", http.StatusConflict)
	case errors.Is(err, ErrDiscountNotApplicable):
		http.Error(w, "discount code does not apply to this order", http.StatusConflict)
	default:
		log.Printf("discount: resolve failed: %q", err.Error())
		http.Error(w, "unable to apply discount right now; try again", http.StatusBadGateway)
	}
}
