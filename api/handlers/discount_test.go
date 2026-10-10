package handlers_test

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/immortalvibes/api/handlers"
	"github.com/immortalvibes/api/models"
	stripe "github.com/stripe/stripe-go/v76"
)

// fakeDiscounts is an in-memory DiscountResolver keyed by upper-cased code.
type fakeDiscounts struct {
	codes map[string]handlers.Discount
	err   error
}

func (f *fakeDiscounts) Resolve(_ context.Context, code string) (handlers.Discount, error) {
	if f.err != nil {
		return handlers.Discount{}, f.err
	}
	d, ok := f.codes[strings.ToUpper(code)]
	if !ok {
		return handlers.Discount{}, handlers.ErrInvalidDiscount
	}
	return d, nil
}

func newFakeDiscounts() *fakeDiscounts {
	return &fakeDiscounts{codes: map[string]handlers.Discount{
		"VIBE10": {Code: "VIBE10", PromotionCodeID: "promo_vibe10", CouponID: "cpn_10", PercentOff: 10},
	}}
}

func TestDiscountAmount(t *testing.T) {
	items := []models.LineItem{
		{ProductID: "prod_tee", Amount: 3000, Quantity: 2},
		{ProductID: "prod_hat", Amount: 2000, Quantity: 1},
	}
	const shipping = int64(750) // subtotal 8000, total 8750

	cases := []struct {
		name     string
		d        handlers.Discount
		currency string
		want     int64
		wantErr  error
	}{
		{"percent covers items and shipping", handlers.Discount{PercentOff: 10}, "usd", 875, nil},
		{"fractional percent floors", handlers.Discount{PercentOff: 12.5}, "usd", 1093, nil},
		{"restricted percent excludes shipping and other items", handlers.Discount{PercentOff: 10, ProductIDs: []string{"prod_hat"}}, "usd", 200, nil},
		{"restricted to products not in cart", handlers.Discount{PercentOff: 10, ProductIDs: []string{"prod_x"}}, "usd", 0, handlers.ErrDiscountNotApplicable},
		{"amount off", handlers.Discount{AmountOff: 500, Currency: "usd"}, "usd", 500, nil},
		{"amount off currency is case-insensitive", handlers.Discount{AmountOff: 500, Currency: "USD"}, "usd", 500, nil},
		{"amount off other currency", handlers.Discount{AmountOff: 500, Currency: "eur"}, "usd", 0, handlers.ErrDiscountNotApplicable},
		{"amount off capped at eligible items", handlers.Discount{AmountOff: 5000, Currency: "usd", ProductIDs: []string{"prod_hat"}}, "usd", 2000, nil},
		{"amount off leaves minimum charge", handlers.Discount{AmountOff: 100000, Currency: "usd"}, "usd", 8700, nil},
		{"full percent leaves minimum charge", handlers.Discount{PercentOff: 100}, "usd", 8700, nil},
		{"minimum met", handlers.Discount{PercentOff: 10, MinimumAmount: 8000, MinimumCurrency: "usd"}, "usd", 875, nil},
		{"minimum excludes shipping", handlers.Discount{PercentOff: 10, MinimumAmount: 8001, MinimumCurrency: "usd"}, "usd", 0, handlers.ErrDiscountNotApplicable},
		{"minimum in other currency", handlers.Discount{PercentOff: 10, MinimumAmount: 100, MinimumCurrency: "eur"}, "usd", 0, handlers.ErrDiscountNotApplicable},
		{"no value", handlers.Discount{}, "usd", 0, handlers.ErrInvalidDiscount},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tc.d.Amount(items, shipping, tc.currency)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("err = %v, want %v", err, tc.wantErr)
			}
			if got != tc.want {
				t.Errorf("amount = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestDiscountFromPromotionCode(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	valid := func() *stripe.PromotionCode {
		return &stripe.PromotionCode{
			ID:     "promo_1",
			Code:   "VIBE10",
			Active: true,
			Coupon: &stripe.Coupon{ID: "cpn_1", Valid: true, PercentOff: 10},
		}
	}

	got, err := handlers.DiscountFromPromotionCode(valid(), now)
	if err != nil {
		t.Fatalf("valid code: %v", err)
	}
	if got.Code != "VIBE10" || got.PromotionCodeID != "promo_1" || got.CouponID != "cpn_1" || got.PercentOff != 10 {
		t.Errorf("got %+v", got)
	}

	pc := valid()
	pc.Coupon.AppliesTo = &stripe.CouponAppliesTo{Products: []string{"prod_1"}}
	pc.Restrictions = &stripe.PromotionCodeRestrictions{MinimumAmount: 5000, MinimumAmountCurrency: stripe.CurrencyUSD}
	got, err = handlers.DiscountFromPromotionCode(pc, now)
	if err != nil {
		t.Fatalf("restricted code: %v", err)
	}
	if len(got.ProductIDs) != 1 || got.MinimumAmount != 5000 || got.MinimumCurrency != "usd" {
		t.Errorf("restrictions not carried: %+v", got)
	}

	rejected := []struct {
		name   string
		mutate func(*stripe.PromotionCode)
	}{
		{"inactive", func(p *stripe.PromotionCode) { p.Active = false }},
		{"no coupon", func(p *stripe.PromotionCode) { p.Coupon = nil }},
		{"coupon invalid", func(p *stripe.PromotionCode) { p.Coupon.Valid = false }},
		{"code expired", func(p *stripe.PromotionCode) { p.ExpiresAt = now.Unix() }},
		{"coupon redeem_by passed", func(p *stripe.PromotionCode) { p.Coupon.RedeemBy = now.Unix() - 1 }},
		{"customer specific", func(p *stripe.PromotionCode) { p.Customer = &stripe.Customer{ID: "cus_1"} }},
		{"first time only", func(p *stripe.PromotionCode) {
			p.Restrictions = &stripe.PromotionCodeRestrictions{FirstTimeTransaction: true}
		}},
		{"no discount value", func(p *stripe.PromotionCode) { p.Coupon.PercentOff = 0 }},
	}
	for _, tc := range rejected {
		t.Run(tc.name, func(t *testing.T) {
			p := valid()
			tc.mutate(p)
			if _, err := handlers.DiscountFromPromotionCode(p, now); !errors.Is(err, handlers.ErrInvalidDiscount) {
				t.Fatalf("err = %v, want ErrInvalidDiscount", err)
			}
		})
	}
	if _, err := handlers.DiscountFromPromotionCode(nil, now); !errors.Is(err, handlers.ErrInvalidDiscount) {
		t.Fatalf("nil: err = %v, want ErrInvalidDiscount", err)
	}
}

// fakeRedemptions returns fixed counts.
type fakeRedemptions struct {
	coupon, promo int64
	err           error
}

func (f fakeRedemptions) CountCoupon(context.Context, string, int64) (int64, error) {
	return f.coupon, f.err
}

func (f fakeRedemptions) CountPromotionCode(context.Context, string, int64) (int64, error) {
	return f.promo, f.err
}

const promoListBody = `{"object": "list", "url": "/v1/promotion_codes", "has_more": false, "data": [{
	"id": "promo_1", "object": "promotion_code", "code": "VIBE10", "active": true, %s
	"coupon": {"id": "cpn_1", "object": "coupon", "valid": true, "percent_off": 10, %s "applies_to": {"products": []}}
}]}`

func TestStripeDiscountResolver(t *testing.T) {
	cases := []struct {
		name        string
		promoLimit  string
		couponLimit string
		redemptions handlers.RedemptionCounter
		empty       bool
		wantErr     error
	}{
		{"unlimited", "", "", nil, false, nil},
		{"unknown code", "", "", nil, true, handlers.ErrInvalidDiscount},
		{"promo limit not reached", `"max_redemptions": 5,`, "", fakeRedemptions{promo: 4}, false, nil},
		{"promo limit reached", `"max_redemptions": 5,`, "", fakeRedemptions{promo: 5}, false, handlers.ErrDiscountExhausted},
		{"coupon limit reached", "", `"max_redemptions": 2,`, fakeRedemptions{coupon: 2}, false, handlers.ErrDiscountExhausted},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			paths := withStripeServer(t, func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if tc.empty {
					_, _ = w.Write([]byte(`{"object": "list", "url": "/v1/promotion_codes", "has_more": false, "data": []}`))
					return
				}
				_, _ = w.Write([]byte(strings.Replace(strings.Replace(promoListBody, "%s", tc.promoLimit, 1), "%s", tc.couponLimit, 1)))
			})

			r := handlers.NewStripeDiscountResolver(tc.redemptions)
			d, err := r.Resolve(context.Background(), " vibe10 ")
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("err = %v, want %v", err, tc.wantErr)
			}
			if tc.wantErr == nil && d.PromotionCodeID != "promo_1" {
				t.Errorf("discount = %+v", d)
			}

			u, err := url.Parse((*paths)[0])
			if err != nil {
				t.Fatalf("parse: %v", err)
			}
			q := u.Query()
			if u.Path != "/v1/promotion_codes" || q.Get("code") != "vibe10" || q.Get("active") != "true" || q.Get("expand[0]") != "data.coupon.applies_to" {
				t.Errorf("request = %s", (*paths)[0])
			}
		})
	}
}

func TestStripeDiscountResolver_LimitedCodeNeedsCounter(t *testing.T) {
	withStripeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(strings.Replace(strings.Replace(promoListBody, "%s", `"max_redemptions": 1,`, 1), "%s", "", 1)))
	})
	_, err := handlers.NewStripeDiscountResolver(nil).Resolve(context.Background(), "VIBE10")
	if err == nil || errors.Is(err, handlers.ErrInvalidDiscount) {
		t.Fatalf("err = %v, want a configuration error", err)
	}
}

func TestStripeRedemptionCounter(t *testing.T) {
	var query string
	withStripeServer(t, func(w http.ResponseWriter, r *http.Request) {
		query = r.URL.Query().Get("query")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"object": "search_result", "url": "/v1/payment_intents/search", "has_more": false, "data": [
			{"id": "pi_1", "object": "payment_intent"}, {"id": "pi_2", "object": "payment_intent"}, {"id": "pi_3", "object": "payment_intent"}
		]}`))
	})

	n, err := handlers.StripeRedemptionCounter{}.CountPromotionCode(context.Background(), "promo_1", 2)
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	if n != 2 {
		t.Errorf("count = %d, want 2 (stops at limit)", n)
	}
	if want := "status:'succeeded' AND metadata['promotion_code_id']:'promo_1'"; query != want {
		t.Errorf("query = %q, want %q", query, want)
	}

	if _, err := (handlers.StripeRedemptionCounter{}).CountCoupon(context.Background(), "x' OR status:'canceled", 1); err == nil {
		t.Error("expected quoted value to be rejected")
	}
}
