# immortalvibes-store — Codebase Guide

Monorepo for theimmortalvibes.com, a streetwear drop store. A SvelteKit storefront
(Cloudflare Pages) talks through a Cloudflare Worker proxy to a Go API (Fly.io)
that handles cart, checkout (Stripe), order persistence (Postgres), stock,
shipping labels (Shippo), and email (Resend). Product drops are "Missions" —
content and stock flips ship as commits to `web/`.

## Directory map

| Path | What it is |
|---|---|
| `api/` | Go 1.26 API service (`github.com/immortalvibes/api`). Deployed to Fly.io (`api/fly.toml`, Dockerfile). Packages: `handlers/` (HTTP incl. `webhook.go` Stripe fulfillment), `store/` (Postgres `db.go` + Cloudflare KV `kv.go` carts), `shippo/` (label client), `email/` (Resend sender), `config/`, `middleware/`, `models/` |
| `web/` | SvelteKit 5 storefront → Cloudflare Pages (`immortalvibes-store` project) |
| `worker/` | Cloudflare Worker proxy in front of the api (holds `PROXY_SECRET`) |
| `drop/`, `photos/` | Multi-MB product/marketing images — see clone trap below |
| `deploy.sh` | Manual Pages deploy shortcut (build `web/` + wrangler pages deploy) |
| `.github/workflows/deploy.yml` | CI: `test-api` gate → deploy Worker → deploy Pages |
| `.github/workflows/storefront-e2e.yml` | CI: storefront Playwright gates (separate from the api gate — see below) |

## Deploy path

- Push to `main` → `deploy.yml`: `test-api` (`go test ./...` in `api/`) must pass,
  then Worker deploys, then Pages. PRs to `main` run tests only.
- The Go api itself deploys to Fly.io separately (`fly deploy` from `api/`) —
  deploy.yml does NOT deploy the api. Test failures still block the CF surfaces.

## STOP 18/19 — order manifest + E2E fulfillment (do not regress)

From the missing-manifest incident (commit `da74dab`):

- **STOP 18**: Orders must be self-contained — line items persist on the order
  (`line_items` JSONB) at checkout; the owner "Label ready" email renders a
  packing manifest. Never rely on the cart surviving past payment.
- **STOP 19**: The purchase→fulfillment path is covered by
  `api/handlers/webhook_test.go: TestWebhookFulfillmentE2E` — manifest reaches
  owner, customer notified, stock decrements exactly once, Stripe webhook
  resends are idempotent. Keep this test passing and hermetic; extend it when
  touching checkout, webhook, stock, or email code.

## web/ storefront map (wave-2)

Features are colocated under `web/src/routes/`; product facts and other
cross-cutting values have exactly one home each under `web/src/lib/`:

| Home | Owns |
|---|---|
| `lib/types/shop.ts` | The catalog: `MOCK_PRODUCTS`, `SLUG` (canonical slugs), `WR_SET`, `MISSION_LABELS`, `productBySlug`. Every page derives product names/slugs/mission envs from here — never restate them. |
| `lib/live-products.ts` | Go-API wire types + the live merge (`fetchLiveProducts`, `mergeLiveProduct`). Both shop loaders share it. Catalog owns visuals/identity; API owns prices/stock. |
| `lib/money.ts` | `formatPrice` — the only cents→string renderer. |
| `lib/stores/cart.ts` | Cart state + the only GoCart→CartItem mapping (`setFromGoCart`, `makeVariantId` dedup key). Layout hydration, add-to-cart, and qty updates all funnel through it. |
| `lib/stores/transition.ts` | Page-transition state, `MISSION_ORDER`, `MISSION_ACCENT` (keyed off `SLUG`). |

Known duplicates left on purpose (documented in place): set-page blurbs/labels
are distinct marketing copy, not catalog descriptions; checkout's inline `$`
totals are USD-only summary math; the Fly.io API URL appears in both
`web/src/routes/api/[...path]/+server.ts` and `worker/wrangler.toml` because
they are separate deployables; `SetPicker`'s `setId` prop type restates the
set-id literal as a type annotation only. Tailwind v3 scans ALL of `web/src`
— including comments — for class-like tokens; never write a bare utility word
(`hidden`, `table`, `inline`, `lowercase`, …) in a comment or you'll grow the
CSS bundle.

## Storefront E2E (Playwright, `web/e2e/`)

- `cd web && npm run e2e:local` — hermetic gate: builds nothing itself; runs
  `vite preview` (port 4173, host 127.0.0.1 — beware the IPv6 black hole) over
  the last build with the Go API mocked at the network edge. Run
  `npm run build` first.
- `cd web && npm run e2e:live` — same journey against theimmortalvibes.com
  (`E2E_TARGET=live` in CI). Creates one labelled pending order per run.
- **The suite hard-stops at the Stripe Payment Element.** It must never fill
  card details, click PAY NOW, or confirm a payment — local or live.
- CI: `storefront-e2e.yml` runs the local gate on PRs/pushes touching `web/`;
  the live gate runs weekly + on manual dispatch only. It is deliberately a
  separate workflow from `deploy.yml`'s `test-api` job — do not merge them.

## Known traps

- **Stale remote handle**: older local checkouts (e.g. the Desktop one) point
  origin at `gr8gray253/immortalvibes-store` — a pre-rename handle. GitHub
  redirects it, but the canonical remote is `gr8-gray/immortalvibes-store`.
- **Huge binaries**: `drop/` and `photos/` hold 5–7 MB images; full clones are
  slow. Prefer `git clone --filter=blob:none` (+ sparse-checkout of `api/`,
  `web/`, `.github/` as needed).
- **Oversell guard**: `DecrementStock` is called once per line item on the first
  transition to `complete`, guarded against Stripe resends. Don't move that
  call without re-running the E2E test.
- Tests need no live credentials — httptest + `t.Setenv` fakes only. Keep it
  that way, or gate live tests behind a `t.Skip` on a missing env var.
