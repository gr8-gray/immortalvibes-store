# 2026-10-10: Checkout stops trusting the browser

**What changed:** three PRs moved every money decision to the server.

- **#21:** cart line items are priced from the Stripe catalog server-side. Quantity is limited to 1–20,
  carts to 25 lines and a single currency, and variant labels are checked against stock rows.
- **#20:** shipping is quoted server-side at checkout. With no quote, checkout returns 502 rather than charging
  a guessed amount. The storefront waits for a resolved quote and shows the server's total.
- **#22:** the store always charges the catalog currency (USD). Promotion codes go through one shared resolver
  (expiry, minimum, applies-to, amount-off currency; customer-restricted and first-time-only codes are
  rejected). An invalid code at checkout returns 409. Percent codes apply to items and shipping. The UI mirrors
  the server math (`web/src/lib/pricing.ts`) for display only. Shippo calls and the HTTP server got timeouts.

**Superseded:** the original "multi-currency priced by visitor country" design. Don't revive it without a
real reason.

**Rule produced:** IV-7.

**Still open:** see [`../OPEN-WORK.md`](../OPEN-WORK.md) (async fulfilment, promo rate limiting, strict
redemption limits, variant-label allow-list).
