package handlers

import (
	"context"
	"errors"
	"fmt"

	"github.com/immortalvibes/api/models"
	stripe "github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/price"
)

// MaxLineQuantity bounds the quantity of a single cart line item.
const MaxLineQuantity = 20

var (
	// ErrPriceUnavailable means the price does not exist or cannot be sold
	// (inactive, recurring, non-per-unit billing, or an inactive product).
	ErrPriceUnavailable = errors.New("price unavailable")
	// ErrInvalidQuantity means a line item quantity is outside [1, MaxLineQuantity].
	ErrInvalidQuantity = errors.New("invalid quantity")
	// ErrMixedCurrency means the cart contains prices in more than one currency.
	ErrMixedCurrency = errors.New("mixed currencies")
)

// CatalogPrice is the authoritative, sellable view of a single price.
type CatalogPrice struct {
	PriceID     string
	ProductID   string
	ProductName string
	Nickname    string
	Currency    string
	UnitAmount  int64
}

// PriceCatalog resolves price IDs to authoritative pricing. Implementations
// return ErrPriceUnavailable (possibly wrapped) for prices that must not be sold.
type PriceCatalog interface {
	Lookup(ctx context.Context, priceID string) (CatalogPrice, error)
}

// StripePriceCatalog is the live PriceCatalog backed by the Stripe Prices API.
// stripe.Key must be set before use.
type StripePriceCatalog struct{}

// Lookup implements PriceCatalog.
func (StripePriceCatalog) Lookup(ctx context.Context, priceID string) (CatalogPrice, error) {
	params := &stripe.PriceParams{}
	params.Context = ctx
	params.AddExpand("product")

	p, err := price.Get(priceID, params)
	if err != nil {
		var serr *stripe.Error
		if errors.As(err, &serr) && serr.Code == stripe.ErrorCodeResourceMissing {
			return CatalogPrice{}, fmt.Errorf("%w: %s", ErrPriceUnavailable, priceID)
		}
		return CatalogPrice{}, fmt.Errorf("stripe price get %s: %w", priceID, err)
	}
	return catalogPriceFromStripe(p)
}

// catalogPriceFromStripe validates that p is sellable as a one-off, per-unit
// item and converts it.
func catalogPriceFromStripe(p *stripe.Price) (CatalogPrice, error) {
	switch {
	case p == nil, p.Deleted, !p.Active:
		return CatalogPrice{}, ErrPriceUnavailable
	case p.Type != stripe.PriceTypeOneTime:
		return CatalogPrice{}, ErrPriceUnavailable
	case p.BillingScheme != stripe.PriceBillingSchemePerUnit:
		return CatalogPrice{}, ErrPriceUnavailable
	case p.CustomUnitAmount != nil, p.UnitAmount <= 0:
		return CatalogPrice{}, ErrPriceUnavailable
	case p.Product == nil, p.Product.Deleted, !p.Product.Active:
		return CatalogPrice{}, ErrPriceUnavailable
	}
	return CatalogPrice{
		PriceID:     p.ID,
		ProductID:   p.Product.ID,
		ProductName: p.Product.Name,
		Nickname:    p.Nickname,
		Currency:    string(p.Currency),
		UnitAmount:  p.UnitAmount,
	}, nil
}

// validQuantity reports whether q is an acceptable line item quantity.
func validQuantity(q int) bool {
	return q >= 1 && q <= MaxLineQuantity
}

// applyCatalogPrice overwrites every pricing and identity field of li with the
// catalog's values. Size is only filled when empty, since it carries the
// buyer's variant selection rather than pricing.
func applyCatalogPrice(li *models.LineItem, cp CatalogPrice) {
	li.ProductID = cp.ProductID
	li.Name = cp.ProductName
	li.Currency = cp.Currency
	li.Amount = cp.UnitAmount
	if li.Size == "" {
		li.Size = cp.Nickname
	}
}

// repriceLineItems re-resolves every line item against the catalog, replacing
// any stored pricing, and enforces quantity bounds and a single currency. It
// mutates items in place and returns the cart currency.
func repriceLineItems(ctx context.Context, catalog PriceCatalog, items []models.LineItem) (string, error) {
	var currency string
	for i := range items {
		li := &items[i]
		if li.PriceID == "" {
			return "", ErrPriceUnavailable
		}
		if !validQuantity(li.Quantity) {
			return "", ErrInvalidQuantity
		}
		cp, err := catalog.Lookup(ctx, li.PriceID)
		if err != nil {
			return "", err
		}
		if currency == "" {
			currency = cp.Currency
		} else if cp.Currency != currency {
			return "", ErrMixedCurrency
		}
		applyCatalogPrice(li, cp)
	}
	return currency, nil
}
