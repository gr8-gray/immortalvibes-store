# CLAUDE.md — Immortal Vibes

Instructions for Claude Code sessions in this repo. The codebase rules (hard constraints) and the known
gaps are imported below, so they load with this file:

@docs/lessons/RULES.md
@docs/lessons/OPEN-WORK.md

## What this is

Live e-commerce store for theimmortalvibes.com: limited clothing drops ("missions") behind an immersive
space-themed storefront. Real customers, real money, so the checkout → payment → shipping path matters most.

| Part | Stack | Hosted |
|---|---|---|
| `web/` | SvelteKit + TypeScript, Three.js, GSAP | Cloudflare Pages (project `immortalvibes-store`) |
| `api/` | Go (version in `api/go.mod`), chi v5 | Fly.io (app `immortalvibes-api`) + Fly Postgres |
| Payments / shipping / email | Stripe Payment Element · Shippo · Resend | — |
| Carts / images | Cloudflare KV · R2 | — |

The browser never calls the Go API directly. `web/src/routes/api/[...path]/+server.ts` proxies `/api/*` to
Fly and injects `X-Proxy-Secret`. There is **no** separate Cloudflare Worker anymore (removed in PR #18).

## Run and test locally

```sh
cd api && go test ./...                  # full API suite, hermetic (no live services)
cd api && go run .                       # needs env vars exported; see api/.env.example (no .env auto-load)
cd web && npm install && npm run dev
cd web && npm run build && npm run e2e:local   # Playwright vs vite preview + mocked API
cd web && npm run check                  # svelte-check; 14 pre-existing TS errors (2026-11-22), don't chase them unprompted
```

- **Never run `npm run e2e:live` locally.** It hits the production site and creates a real pending order. It
  runs only in CI: after every deploy (`verify-live`), weekly, or by manual dispatch of Storefront E2E.
- **There is no staging stack.** Production is the only live environment. Pages *preview* deployments are not
  a sandbox either: previews call the production API (`GO_API` in `web/src/routes/api/[...path]/+server.ts`),
  so a checkout on a preview creates real orders and moves real stock.

## How a change reaches production

Only one route: `.github/workflows/deploy.yml`, on push to `main`.

1. `test-api`: the whole Go suite. Everything below waits on it.
2. `deploy-api` and `deploy-pages` run **in parallel**:
   - `deploy-api` still shows green when `api/` is unchanged, but its *Deploy API to Fly* step is skipped. Otherwise it records the running image, runs
     `flyctl deploy`, and runs `.github/scripts/smoke-api.sh`. **If the smoke fails it rolls back to the
     recorded image.**
   - `deploy-pages` builds `web/` and deploys to Cloudflare Pages.
3. `verify-live` waits on `deploy-pages` **only**: the live Playwright journey against the real site.
4. `rollback` redeploys the previous commit to **Pages only** if `verify-live` fails. It does not touch the API.
   **This rollback hasn't been proven since the Worker removal; see OPEN-WORK.**

Cloudflare Pages' Git integration has production deploys disabled. An idle `github:push` record per push is
normal (RULES IV-2). Storefront E2E (`storefront-e2e.yml`) also runs `e2e-local` on PRs that touch `web/`.

**"Did it work?" checklist** (STOP 1). In the `deploy` run, every job except `rollback` ran and passed
(`rollback` is skipped on every healthy push). For API changes, open `deploy-api` and confirm the *Deploy API to
Fly* and *Smoke the API* steps both ran, not skipped. `verify-live` passed. Then look at the live site.

**Config is a second route.** `fly secrets set` creates a new production release outside the pipeline, and
the image rollback can't undo it (RULES IV-17). For a secret rotation: ask first, change one value, check
`.github/scripts/smoke-api.sh` and the site immediately, and keep the old value until the new one works.

## Shipping policy

- The store owner merges his own PRs. The previous maintainer is an optional backup reviewer.
- Every change goes through a PR with green CI. **No direct pushes to `main`.** Branch protection isn't on yet
  (OPEN-WORK), so this is a rule, not a lock.
- "Ship it" means: open the PR, wait for green CI, merge (ask first), then run the "did it work" checklist.
- The PR template's "13 Layers" table and its "Opus merge reviewer" line come from the previous maintainer's
  setup. Fill in the layers that apply. There is no automated merge reviewer.

## Test orders: never fulfil, never count

`verify-live`, `rollback` and `e2e-live` each leave a real pending order labelled **"E2E GATE - DO NOT
FULFILL"** in production. Exclude these from fulfilment, stock and every sales or conversion number (RULES IV-9).

## Secrets (names only; values never go in this repo)

- **Fly secrets** (read in `api/config/config.go`): `DATABASE_URL`, `PROXY_SECRET`, `ADMIN_SECRET`, `STRIPE_SECRET_KEY`,
  `STRIPE_WEBHOOK_SECRET`, `RESEND_API_KEY`, `SHIPPO_API_KEY`, `OWNER_EMAIL`, `FROM_NAME`, `FROM_STREET1`,
  `FROM_CITY`, `FROM_STATE`, `FROM_ZIP`, `FROM_COUNTRY` (optional, default US), `FROM_PHONE`, `CF_ACCOUNT_ID`, `CF_API_TOKEN`, `CF_KV_CARTS_ID`,
  `R2_PUBLIC_URL`, `NTFY_URL`, `NTFY_TOPIC`.
- **Cloudflare Pages env:** `PROXY_SECRET` (must match Fly's), `PUBLIC_STRIPE_PUBLISHABLE_KEY` (rotate it
  together with `STRIPE_SECRET_KEY`), `PUBLIC_API_URL`.
- **GitHub Actions:** `CLOUDFLARE_API_TOKEN`, `CLOUDFLARE_ACCOUNT_ID`, `FLY_API_TOKEN`, plus optional alert
  secrets `NTFY_URL`, `NTFY_TOPIC`, `TELEGRAM_BOT_TOKEN`, `TELEGRAM_CHAT_ID`.

## Decide on your own vs. ask first

Decide (and say what you decided): naming, structure, refactors inside files you're changing, tests,
choosing among installed dependencies.

**Ask first:**
- merging to `main` (= production deploy)
- dispatching any workflow that touches production (`deploy` drills, Storefront E2E `run_live`)
- live Stripe changes (products, prices, coupons, refunds)
- stock writes via `PUT /api/admin/products/{id}/stock`
- database migrations
- any secret or Fly/Pages config change
- DNS or Cloudflare settings
- anything that emails customers
- new dependencies
- deleting branches or data
- force push or history rewrites

## Before you say "done"

Show the evidence: test output, the CI run, the live result (STOP 9). An "all clear" counts only after the
check has been shown to find a real case (STOP 23). This repo is public, so no customer data, personal
details, account IDs or secret values in code, docs, commits or PRs (STOP 21).
