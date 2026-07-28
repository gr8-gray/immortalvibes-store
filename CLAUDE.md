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
