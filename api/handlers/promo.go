package handlers

import (
	"encoding/json"
	"net/http"

	stripe "github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/coupon"
	"github.com/stripe/stripe-go/v76/promotioncode"
)

// PromoHandler handles POST /api/promo/validate.
type PromoHandler struct{}

func NewPromoHandler() *PromoHandler { return &PromoHandler{} }

type promoValidateRequest struct {
	Code string `json:"code"`
}

type promoDiscount struct {
	Type  string `json:"type"`  // "percent_off" | "amount_off"
	Value int64  `json:"value"` // percent (0-100) or amount in cents
}

type promoValidateResponse struct {
	Valid    bool           `json:"valid"`
	CouponID string         `json:"coupon_id,omitempty"`
	Discount *promoDiscount `json:"discount,omitempty"`
	Error    string         `json:"error,omitempty"`
}

// Validate handles POST /api/promo/validate.
// It first tries to look up a Stripe PromotionCode by code, then falls back
// to a direct Coupon lookup by ID. Returns discount details on success.
func (h *PromoHandler) Validate(w http.ResponseWriter, r *http.Request) {
	var req promoValidateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Code == "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(promoValidateResponse{Valid: false, Error: "code required"})
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// Try PromotionCode first (human-readable codes like VIBE10).
	pcParams := &stripe.PromotionCodeListParams{}
	pcParams.Filters.AddFilter("code", "", req.Code)
	pcParams.Filters.AddFilter("active", "", "true")
	pcParams.Filters.AddFilter("limit", "", "1")
	iter := promotioncode.List(pcParams)

	for iter.Next() {
		pc := iter.PromotionCode()
		if pc.Coupon == nil {
			break
		}
		disc := discountFromCoupon(pc.Coupon)
		if disc == nil {
			break
		}
		json.NewEncoder(w).Encode(promoValidateResponse{
			Valid:    true,
			CouponID: pc.Coupon.ID,
			Discount: disc,
		})
		return
	}
	if err := iter.Err(); err != nil {
		json.NewEncoder(w).Encode(promoValidateResponse{Valid: false, Error: "Invalid code"})
		return
	}

	// Fall back: try treating the code as a direct Coupon ID.
	c, err := coupon.Get(req.Code, nil)
	if err != nil || !c.Valid {
		json.NewEncoder(w).Encode(promoValidateResponse{Valid: false, Error: "Invalid code"})
		return
	}
	disc := discountFromCoupon(c)
	if disc == nil {
		json.NewEncoder(w).Encode(promoValidateResponse{Valid: false, Error: "Invalid code"})
		return
	}
	json.NewEncoder(w).Encode(promoValidateResponse{
		Valid:    true,
		CouponID: c.ID,
		Discount: disc,
	})
}

func discountFromCoupon(c *stripe.Coupon) *promoDiscount {
	if c == nil {
		return nil
	}
	if c.PercentOff > 0 {
		return &promoDiscount{Type: "percent_off", Value: int64(c.PercentOff)}
	}
	if c.AmountOff > 0 {
		return &promoDiscount{Type: "amount_off", Value: c.AmountOff}
	}
	return nil
}
