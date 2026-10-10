package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

// PromoHandler handles POST /api/promo/validate.
type PromoHandler struct {
	discounts DiscountResolver
}

// NewPromoHandler constructs a PromoHandler backed by the same resolver
// checkout uses, so a code accepted here is accepted at checkout.
func NewPromoHandler(discounts DiscountResolver) *PromoHandler {
	return &PromoHandler{discounts: discounts}
}

type promoValidateRequest struct {
	Code string `json:"code"`
}

type promoDiscount struct {
	Type  string  `json:"type"`  // "percent_off" | "amount_off"
	Value float64 `json:"value"` // percent (0-100) or amount in cents
	// ProductIDs restricts the discount to these products; when set,
	// shipping is not discounted.
	ProductIDs []string `json:"product_ids,omitempty"`
	// MinimumAmount is the item subtotal, in cents, the order must reach.
	MinimumAmount int64 `json:"minimum_amount,omitempty"`
}

type promoValidateResponse struct {
	Valid    bool           `json:"valid"`
	CouponID string         `json:"coupon_id,omitempty"`
	Discount *promoDiscount `json:"discount,omitempty"`
	Error    string         `json:"error,omitempty"`
}

// Validate handles POST /api/promo/validate. Order-specific conditions
// (minimum amount, eligible items) are returned for display and enforced at
// checkout.
func (h *PromoHandler) Validate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req promoValidateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Code == "" {
		json.NewEncoder(w).Encode(promoValidateResponse{Error: "code required"})
		return
	}

	d, err := h.discounts.Resolve(r.Context(), req.Code)
	switch {
	case err == nil:
	case errors.Is(err, ErrInvalidDiscount):
		json.NewEncoder(w).Encode(promoValidateResponse{Error: "Invalid code"})
		return
	case errors.Is(err, ErrDiscountExhausted):
		json.NewEncoder(w).Encode(promoValidateResponse{Error: "This code is no longer available"})
		return
	default:
		log.Printf("promo: resolve failed: %q", err.Error())
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(promoValidateResponse{Error: "Could not validate code. Try again."})
		return
	}

	disc := &promoDiscount{ProductIDs: d.ProductIDs, MinimumAmount: d.MinimumAmount}
	if d.PercentOff > 0 {
		disc.Type, disc.Value = "percent_off", d.PercentOff
	} else {
		disc.Type, disc.Value = "amount_off", float64(d.AmountOff)
	}
	json.NewEncoder(w).Encode(promoValidateResponse{Valid: true, CouponID: d.CouponID, Discount: disc})
}
