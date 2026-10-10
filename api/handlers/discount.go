package handlers

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/immortalvibes/api/models"
	stripe "github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/paymentintent"
	"github.com/stripe/stripe-go/v76/promotioncode"
)

// minChargeCents is the smallest amount a discount may leave on an order.
// Stripe rejects USD payment intents below $0.50.
const minChargeCents = 50

var (
	// ErrInvalidDiscount means the code does not exist, is inactive or expired,
	// or carries a restriction this store cannot honor.
	ErrInvalidDiscount = errors.New("invalid discount code")
	// ErrDiscountNotApplicable means the code is valid but does not apply to
	// this order (minimum not met, currency mismatch, no eligible items).
	ErrDiscountNotApplicable = errors.New("discount not applicable")
	// ErrDiscountExhausted means the code has reached its redemption limit.
	ErrDiscountExhausted = errors.New("discount redemption limit reached")
)

// Discount is a resolved promotion code and its coupon terms.
type Discount struct {
	Code            string
	PromotionCodeID string
	CouponID        string
	// Exactly one of PercentOff and AmountOff is non-zero.
	PercentOff float64
	AmountOff  int64
	// Currency is set for AmountOff discounts.
	Currency string
	// ProductIDs restricts the discount to these products when non-empty;
	// shipping is then not discounted.
	ProductIDs []string
	// MinimumAmount, in MinimumCurrency, is the item subtotal required.
	MinimumAmount   int64
	MinimumCurrency string
}

// DiscountResolver resolves a buyer-entered code. Implementations return
// ErrInvalidDiscount or ErrDiscountExhausted (possibly wrapped) for codes that
// must not be honored.
type DiscountResolver interface {
	Resolve(ctx context.Context, code string) (Discount, error)
}

// Amount returns the discount in cents for an order with the given line
// items (already priced from the catalog), shipping and currency. Percentage
// discounts apply to the item subtotal plus shipping unless the coupon is
// restricted to specific products, in which case they apply to those items
// only. The discount never leaves less than minChargeCents to pay.
func (d Discount) Amount(items []models.LineItem, shippingCents int64, currency string) (int64, error) {
	var subtotal, eligible int64
	restricted := len(d.ProductIDs) > 0
	for _, li := range items {
		line := li.Amount * int64(li.Quantity)
		subtotal += line
		if !restricted || containsString(d.ProductIDs, li.ProductID) {
			eligible += line
		}
	}
	if !restricted {
		eligible += shippingCents
	}

	if d.MinimumAmount > 0 {
		if !strings.EqualFold(d.MinimumCurrency, currency) || subtotal < d.MinimumAmount {
			return 0, ErrDiscountNotApplicable
		}
	}
	if eligible <= 0 {
		return 0, ErrDiscountNotApplicable
	}

	var off int64
	switch {
	case d.PercentOff > 0:
		off = int64(math.Floor(float64(eligible) * d.PercentOff / 100))
	case d.AmountOff > 0:
		if !strings.EqualFold(d.Currency, currency) {
			return 0, ErrDiscountNotApplicable
		}
		off = min(d.AmountOff, eligible)
	default:
		return 0, ErrInvalidDiscount
	}

	total := subtotal + shippingCents
	return max(0, min(off, total-minChargeCents)), nil
}

func containsString(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

// RedemptionCounter counts completed orders that used a coupon or promotion
// code, stopping once limit is reached.
type RedemptionCounter interface {
	CountCoupon(ctx context.Context, couponID string, limit int64) (int64, error)
	CountPromotionCode(ctx context.Context, promotionCodeID string, limit int64) (int64, error)
}

// StripeDiscountResolver resolves codes through the Stripe PromotionCodes API.
// Coupons are not attached to payment intents, so Stripe never increments
// redemption counts; limits are enforced against counted orders instead.
// stripe.Key must be set before use.
type StripeDiscountResolver struct {
	Redemptions RedemptionCounter
	now         func() time.Time
}

// NewStripeDiscountResolver constructs the live resolver.
func NewStripeDiscountResolver(redemptions RedemptionCounter) *StripeDiscountResolver {
	return &StripeDiscountResolver{Redemptions: redemptions, now: time.Now}
}

// Resolve implements DiscountResolver.
func (s *StripeDiscountResolver) Resolve(ctx context.Context, code string) (Discount, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return Discount{}, ErrInvalidDiscount
	}
	params := &stripe.PromotionCodeListParams{
		Code:   stripe.String(code),
		Active: stripe.Bool(true),
	}
	params.Context = ctx
	params.Limit = stripe.Int64(1)
	params.AddExpand("data.coupon.applies_to")

	iter := promotioncode.List(params)
	if !iter.Next() {
		if err := iter.Err(); err != nil {
			return Discount{}, fmt.Errorf("stripe promotion code list: %w", err)
		}
		return Discount{}, ErrInvalidDiscount
	}
	pc := iter.PromotionCode()

	d, err := discountFromPromotionCode(pc, s.now())
	if err != nil {
		return Discount{}, err
	}
	if err := s.checkRedemptions(ctx, pc); err != nil {
		return Discount{}, err
	}
	return d, nil
}

func (s *StripeDiscountResolver) checkRedemptions(ctx context.Context, pc *stripe.PromotionCode) error {
	if s.Redemptions == nil {
		if pc.MaxRedemptions > 0 || pc.Coupon.MaxRedemptions > 0 {
			return fmt.Errorf("redemption counter not configured for limited code %s", pc.ID)
		}
		return nil
	}
	if limit := pc.MaxRedemptions; limit > 0 {
		n, err := s.Redemptions.CountPromotionCode(ctx, pc.ID, limit)
		if err != nil {
			return err
		}
		if n >= limit {
			return ErrDiscountExhausted
		}
	}
	if limit := pc.Coupon.MaxRedemptions; limit > 0 {
		n, err := s.Redemptions.CountCoupon(ctx, pc.Coupon.ID, limit)
		if err != nil {
			return err
		}
		if n >= limit {
			return ErrDiscountExhausted
		}
	}
	return nil
}

// discountFromPromotionCode validates pc against restrictions this store can
// honor and converts it. Customer-specific and first-time-only codes are
// rejected because checkout has no customer identity to check them against.
func discountFromPromotionCode(pc *stripe.PromotionCode, now time.Time) (Discount, error) {
	switch {
	case pc == nil, !pc.Active, pc.Coupon == nil, !pc.Coupon.Valid:
		return Discount{}, ErrInvalidDiscount
	case pc.ExpiresAt > 0 && now.Unix() >= pc.ExpiresAt:
		return Discount{}, ErrInvalidDiscount
	case pc.Coupon.RedeemBy > 0 && now.Unix() >= pc.Coupon.RedeemBy:
		return Discount{}, ErrInvalidDiscount
	case pc.Customer != nil:
		return Discount{}, ErrInvalidDiscount
	case pc.Restrictions != nil && pc.Restrictions.FirstTimeTransaction:
		return Discount{}, ErrInvalidDiscount
	case pc.Coupon.PercentOff <= 0 && pc.Coupon.AmountOff <= 0:
		return Discount{}, ErrInvalidDiscount
	}

	d := Discount{
		Code:            pc.Code,
		PromotionCodeID: pc.ID,
		CouponID:        pc.Coupon.ID,
		PercentOff:      pc.Coupon.PercentOff,
		AmountOff:       pc.Coupon.AmountOff,
		Currency:        string(pc.Coupon.Currency),
	}
	if pc.Coupon.AppliesTo != nil {
		d.ProductIDs = pc.Coupon.AppliesTo.Products
	}
	if r := pc.Restrictions; r != nil && r.MinimumAmount > 0 {
		d.MinimumAmount = r.MinimumAmount
		d.MinimumCurrency = string(r.MinimumAmountCurrency)
	}
	return d, nil
}

// StripeRedemptionCounter counts succeeded payment intents carrying the
// coupon_id / promotion_code_id metadata written at checkout. Stripe search is
// eventually consistent (typically under a minute), so a limit can be
// exceeded by orders completed within that window.
type StripeRedemptionCounter struct{}

// CountCoupon implements RedemptionCounter.
func (StripeRedemptionCounter) CountCoupon(ctx context.Context, couponID string, limit int64) (int64, error) {
	return countSucceeded(ctx, "coupon_id", couponID, limit)
}

// CountPromotionCode implements RedemptionCounter.
func (StripeRedemptionCounter) CountPromotionCode(ctx context.Context, promotionCodeID string, limit int64) (int64, error) {
	return countSucceeded(ctx, "promotion_code_id", promotionCodeID, limit)
}

func countSucceeded(ctx context.Context, key, value string, limit int64) (int64, error) {
	if strings.ContainsAny(value, `'"\`) {
		return 0, fmt.Errorf("unsupported metadata value %q", value)
	}
	params := &stripe.PaymentIntentSearchParams{}
	params.Context = ctx
	params.Query = fmt.Sprintf("status:'succeeded' AND metadata['%s']:'%s'", key, value)
	params.Limit = stripe.Int64(100)

	var n int64
	iter := paymentintent.Search(params)
	for n < limit && iter.Next() {
		n++
	}
	if err := iter.Err(); err != nil {
		return 0, fmt.Errorf("stripe payment intent search: %w", err)
	}
	return n, nil
}
