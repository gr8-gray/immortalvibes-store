// web/src/lib/money.ts
//
// The single home for money rendering. Before this file existed, the exact
// same cents-to-string function was hand-copied into CartDrawer, the product
// page, and the order confirmation page — three chances for a currency symbol
// to drift out of sync. Cross-cutting values get a lib/ home; this is it.
//
// Prices travel through the app as integer cents (see Product.price_usd in
// $lib/types/shop). They only become a human string at the very last moment,
// right at the template boundary, via this function.

/**
 * Render integer cents as a display price, e.g. 3000 -> "$30.00".
 * The only currencies the store sells in today are USD and GBP; anything
 * that is not GBP falls through to the dollar sign on purpose, matching
 * the API's default currency.
 *
 * A note on the toLowerCase: the three hand-copied versions this replaced
 * had already drifted — two normalised case, one compared raw. Every caller
 * in the app passes lower-cased codes today, so the tolerant form is the
 * superset that changes nothing and survives an API that shouts "GBP".
 */
export function formatPrice(cents: number, currency: string): string {
  const symbol = currency.toLowerCase() === 'gbp' ? '£' : '$';
  return `${symbol}${(cents / 100).toFixed(2)}`;
}
