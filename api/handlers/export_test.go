package handlers

import (
	"time"

	stripe "github.com/stripe/stripe-go/v76"
)

// Test-only exports for the external handlers_test package.
var (
	RepriceLineItems       = repriceLineItems
	CatalogPriceFromStripe = catalogPriceFromStripe
)

// SetCatalogClock replaces the cache's clock.
func SetCatalogClock(c *CachedPriceCatalog, now func() time.Time) { c.now = now }

// SetPaymentIntentFunc swaps the Stripe PaymentIntent creator so tests can
// inspect the charged amount without calling Stripe.
func (h *CheckoutHandler) SetPaymentIntentFunc(f func(*stripe.PaymentIntentParams) (*stripe.PaymentIntent, error)) {
	h.newPaymentIntent = f
}

// DiscountFromPromotionCode exposes discountFromPromotionCode.
var DiscountFromPromotionCode = discountFromPromotionCode
