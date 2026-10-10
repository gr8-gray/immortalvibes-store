package handlers

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/immortalvibes/api/models"
	"github.com/immortalvibes/api/store"
	stripe "github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/price"
)

const (
	// MaxLineQuantity bounds the quantity of a single cart line item.
	MaxLineQuantity = 20
	// MaxCartLines bounds the number of distinct line items in a cart.
	MaxCartLines = 25
	// maxVariantLen bounds the length of a variant label.
	maxVariantLen = 64
)

var (
	// ErrPriceUnavailable means the price does not exist or cannot be sold
	// (inactive, recurring, non-per-unit billing, or an inactive product).
	ErrPriceUnavailable = errors.New("price unavailable")
	// ErrInvalidQuantity means a line item quantity is outside [1, MaxLineQuantity].
	ErrInvalidQuantity = errors.New("invalid quantity")
	// ErrMixedCurrency means the cart contains prices in more than one currency.
	ErrMixedCurrency = errors.New("mixed currencies")
	// ErrInvalidVariant means the variant label is malformed.
	ErrInvalidVariant = errors.New("invalid variant")
	// ErrCartFull means the cart already holds MaxCartLines line items.
	ErrCartFull = errors.New("cart full")
)

// LineItemError attributes a validation error to a stored line item so the
// buyer can be told which item to remove.
type LineItemError struct {
	Name string
	Err  error
}

func (e *LineItemError) Error() string { return fmt.Sprintf("line item %q: %v", e.Name, e.Err) }
func (e *LineItemError) Unwrap() error { return e.Err }

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

// CachedPriceCatalog memoizes successful lookups from an underlying catalog
// for a fixed TTL. Failures are never cached, so a price that becomes
// unavailable stops selling within one TTL.
type CachedPriceCatalog struct {
	next PriceCatalog
	ttl  time.Duration
	now  func() time.Time

	mu      sync.Mutex
	entries map[string]cachedPrice
}

type cachedPrice struct {
	price   CatalogPrice
	expires time.Time
}

// NewCachedPriceCatalog wraps next with a TTL cache.
func NewCachedPriceCatalog(next PriceCatalog, ttl time.Duration) *CachedPriceCatalog {
	return &CachedPriceCatalog{next: next, ttl: ttl, now: time.Now, entries: map[string]cachedPrice{}}
}

// Lookup implements PriceCatalog.
func (c *CachedPriceCatalog) Lookup(ctx context.Context, priceID string) (CatalogPrice, error) {
	now := c.now()
	c.mu.Lock()
	e, ok := c.entries[priceID]
	c.mu.Unlock()
	if ok && now.Before(e.expires) {
		return e.price, nil
	}

	cp, err := c.next.Lookup(ctx, priceID)
	if err != nil {
		return CatalogPrice{}, err
	}
	c.mu.Lock()
	c.entries[priceID] = cachedPrice{price: cp, expires: now.Add(c.ttl)}
	c.mu.Unlock()
	return cp, nil
}

// validQuantity reports whether q is an acceptable line item quantity.
func validQuantity(q int) bool {
	return q >= 1 && q <= MaxLineQuantity
}

// normalizeVariant trims a client-supplied variant label and rejects
// oversized values. Labels are not checked against variant stock rows: a
// variant without a row is untracked and sellable, so membership cannot be
// enforced without rejecting legitimate selections. MaxCartLines bounds what
// distinct labels can do to a cart.
func normalizeVariant(s string) (string, error) {
	s = strings.TrimSpace(s)
	if len(s) > maxVariantLen {
		return "", ErrInvalidVariant
	}
	return s, nil
}

// findVariant returns the row for variant, if present.
func findVariant(rows []store.VariantStockRow, variant string) (store.VariantStockRow, bool) {
	for _, row := range rows {
		if row.Variant == variant {
			return row, true
		}
	}
	return store.VariantStockRow{}, false
}

// applyCatalogPrice overwrites every pricing and identity field of li with
// the catalog's values. Size is left alone: it carries the buyer's variant
// selection and is part of the line item's identity in the cart.
func applyCatalogPrice(li *models.LineItem, cp CatalogPrice) {
	li.ProductID = cp.ProductID
	li.Name = cp.ProductName
	li.Currency = cp.Currency
	li.Amount = cp.UnitAmount
}

// repriceLineItems re-resolves every line item against the catalog, replacing
// any stored pricing, and enforces quantity and line bounds and a single
// currency. Empty sizes are filled from the price nickname so the persisted
// manifest is readable. It mutates items in place and returns the cart
// currency. Per-item failures are returned as *LineItemError.
func repriceLineItems(ctx context.Context, catalog PriceCatalog, items []models.LineItem) (string, error) {
	if len(items) > MaxCartLines {
		return "", ErrCartFull
	}
	resolved := make(map[string]CatalogPrice, len(items))
	var currency string
	for i := range items {
		li := &items[i]
		fail := func(err error) (string, error) {
			return "", &LineItemError{Name: li.Name, Err: err}
		}
		if li.PriceID == "" {
			return fail(ErrPriceUnavailable)
		}
		if !validQuantity(li.Quantity) {
			return fail(ErrInvalidQuantity)
		}
		cp, ok := resolved[li.PriceID]
		if !ok {
			var err error
			if cp, err = catalog.Lookup(ctx, li.PriceID); err != nil {
				return fail(err)
			}
			resolved[li.PriceID] = cp
		}
		if currency == "" {
			currency = cp.Currency
		} else if cp.Currency != currency {
			return fail(ErrMixedCurrency)
		}
		applyCatalogPrice(li, cp)
		if li.Size == "" {
			li.Size = cp.Nickname
		}
	}
	return currency, nil
}
