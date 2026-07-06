# Per-Variant Stock Tracking — Design Contract (2026-07-05)

> Trigger: owner inventory reconcile exposed that stock is per-product only; size-level
> oversell is possible. ALSO fixes a live gap: the customer's chosen size never reaches the
> backend (AddItemPayload has no size field), so order manifests rely on Stripe price
> nicknames that are per-CURRENCY, not per-size — manifest sizes are unreliable.
> Design by Cire (Fable). Build agent: implement exactly; deviations must be justified in
> the completion report.

## Core decision

**Variant = explicit string label carried from the frontend** ("S"…"3XL" for apparel,
"Earth Blue"/"Green" for the beanie). NOT per-size Stripe Prices (would mean 6 sizes × 4
currencies = 24 prices per product — owner burden, bigger refactor). Stripe stays untouched;
catalog source of truth unchanged. **Reuse the existing `LineItem.Size` field as the variant
carrier** — the manifest already renders it, so fulfillment output improves for free.

## Schema (api/store/db.go migrate() — idempotent, additive)

```sql
CREATE TABLE IF NOT EXISTS variant_stock (
    product_id  TEXT NOT NULL,
    variant     TEXT NOT NULL,
    stock_count INT  NOT NULL DEFAULT 0,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (product_id, variant)
);
```

Legacy `product_stock` stays (transition safety + old carts).

## Read/write semantics

- **GetStock(product) total** = `SUM(stock_count) FROM variant_stock WHERE product_id=$1`
  **if any variant rows exist**, else fall back to legacy `product_stock` row. Products API
  `stock_count` keeps its current meaning (total) — frontend status derivation untouched.
- **GET /api/products**: each product gains `"variants": [{"variant": "L", "stock_count": 2}, …]`
  (empty array when no variant rows — frontend treats missing/empty as "no per-size data,
  all sizes enabled", today's behavior).
- **DecrementStock(product, variant, qty)**: atomic
  `UPDATE variant_stock … WHERE product_id=$1 AND variant=$2 AND stock_count>=$3`.
  If 0 rows matched (no such variant row — e.g. legacy cart with empty Size), fall back to
  legacy product_stock decrement (current SQL). Log which path ran.
- **Admin**: `PUT /api/admin/products/{id}/stock` body gains optional `"variant"`:
  `{"count": 2, "variant": "L"}` upserts a variant row; body without `variant` keeps today's
  legacy-total behavior (back-compat, existing scripts unaffected).
  NEW: `GET /api/admin/products/{id}/stock` returns all variant rows + legacy row (owner
  reconciles need a read).

## Cart/checkout flow

1. `AddItemPayload` (web/src/lib/api.ts) + Go add-item handler + `models.LineItem` KV shape:
   size flows in at add-to-cart (`size` json field, reusing LineItem.Size). Beanie colorway
   swatch sets it too (variant labels must match `ProductVariant[]` names in shop.ts exactly).
2. Cart dedup key becomes price_id+size — frontend store `variantId` = `${price_id}:${size}`,
   Go cart merge logic keyed the same. Cart drawer displays the size/colorway per line.
3. `enrichSizes` (checkout.go): only fills `Size` from price nickname **when Size is empty**
   (legacy carts) — never overwrites a frontend-provided value.
4. **Checkout stock guard** (new, cheap): at PaymentIntent creation, SELECT variant stock for
   each line; if any line's variant row exists and stock < qty → 409 with a clear message.
   Non-atomic (no reservation) — same oversell window as today, but blocks the obvious case.
5. Webhook decrement passes `li.Size` per line (fallback path per semantics above).
   Idempotency guards unchanged.

## Frontend availability UI

- Loaders (`shop/+page.ts`, `shop/[slug]/+page.ts`): pass `variants` through the merge.
  Product-level status derivation unchanged (total > 0).
- `SizeSelector.svelte`: new `soldOut: Set<string>` prop — sizes with a variant row at 0 render
  disabled + struck-through (match existing design language: Lunar White at reduced opacity,
  no new colors). Sizes with NO variant row stay enabled (no data ≠ sold out).
- Beanie colorway swatches: same treatment when a colorway hits 0.

## Tests (STOP 19 — mandatory)

- Extend `TestWebhookFulfillmentE2E`: line item with Size="L" decrements variant (L: 3→1 on
  qty 2), other variants untouched, manifest email still asserts Size.
- New: legacy-cart fallback (Size="" decrements product_stock), checkout 409 on insufficient
  variant stock, admin variant upsert + GET, resend idempotency still holds with variant path.
- `go build` + full `go test ./...` green before deploy. Frontend `npm run build` green.

## Deploy + seed (in order)

1. Deploy API (`fly deploy` from api/ — migrations run on boot, additive/safe).
2. Seed via admin endpoint (owner's authoritative 2026-07-05 count; labels must match shop.ts):
   - WR Tee + WR Sweatpants (each): S=1, M=2, L=2, XL=2, 2XL=1, 3XL=1 (owner's XXL/XXXL → 2XL/3XL)
   - Racerback: S=2, M=3, L=2, XL=2
   - Beanie: per shop.ts variant names — blue colorway=4, green colorway=1
   - Immortal Light Sweatpants: ALL size variants = 0 (sold out; explicit rows so restock is easy)
3. Verify `GET /api/products`: totals unchanged (9/9/9/5/0) AND variants arrays present.
4. Commit + push frontend → GH Actions → CF Pages. Verify live: product page renders size
   picker with no regressions; cart line shows size; IL page still SOLD OUT.
5. Verify legacy admin PUT (no variant) still works against a throwaway value, then restore.

## Out of scope

- Stock reservation/holds (oversell window between concurrent checkouts stays)
- Size+color combined dimensions (no current product needs both)
- Per-size restock alerts
