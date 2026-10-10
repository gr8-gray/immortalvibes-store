package handlers

import "time"

// Test-only exports for the external handlers_test package.
var (
	RepriceLineItems       = repriceLineItems
	CatalogPriceFromStripe = catalogPriceFromStripe
)

// SetCatalogClock replaces the cache's clock.
func SetCatalogClock(c *CachedPriceCatalog, now func() time.Time) { c.now = now }
