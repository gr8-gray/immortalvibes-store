# Immortal Vibes

E-commerce storefront for [theimmortalvibes.com](https://theimmortalvibes.com), built as three
deployables behind a Cloudflare edge:

- **`web/`** — SvelteKit storefront (Cloudflare Pages).
- **`worker/`** — Cloudflare Worker that fronts the storefront's `/api/*` calls, adds rate
  limiting/session handling, and proxies through to the Go API.
- **`api/`** — Go REST API (Fly.io) — products, cart, checkout (Stripe), orders, shipping (Shippo),
  admin, email (Resend), order-alert notifications (ntfy).

## Architecture

```
browser → Cloudflare Pages (web/)
            → web/src/routes/api/[...path]/+server.ts   (in-app proxy, injects PROXY_SECRET)
              → Cloudflare Worker (worker/)               (rate limit, sessions, KV carts)
                → Go API (api/, Fly.io)                    (products, checkout, orders)
```

The storefront never calls the Go API directly. Every `/api/*` request from the browser is
proxied server-side through `web/src/routes/api/[...path]/+server.ts`, which attaches the shared
`PROXY_SECRET` before forwarding to the Worker. The Worker re-checks that secret, applies rate
limiting/session handling, then forwards to the Go API on Fly.io.

## Local development

### API (`api/`)

```sh
cd api
go run .
```

Config is read directly via `os.Getenv` (`api/config/config.go`) — there is **no** `.env`
auto-loading. Export the required variables in your shell (or use `direnv`/similar) before
running. See `api/.env.example` for the full list.

### Storefront (`web/`)

```sh
cd web
npm install
npm run dev
```

### Worker (`worker/`)

```sh
cd worker
npm install
npx wrangler dev
```

Copy `worker/.dev.vars.example` to `worker/.dev.vars` and set `PROXY_SECRET` to match the value
used by the storefront's proxy route.

## Testing

```sh
cd web
npm run build && npm run e2e:local
```

**Never** run the `live` Playwright project (`npm run e2e:live`) outside of CI/production — it
targets theimmortalvibes.com directly and creates a real Stripe-pending order.

Go API tests:

```sh
cd api
go test ./...
```

## Deploy

`.github/workflows/deploy.yml` runs on push to `main`:

1. **Test gate** — the full Go API test suite must pass before anything deploys.
2. **Deploy Worker** — `wrangler deploy` + rotate the Worker's `PROXY_SECRET`.
3. **Deploy Pages** — build the storefront and deploy to Cloudflare Pages.
4. **Post-deploy live verification** — runs the full live customer journey (`web/e2e/storefront.spec.ts`,
   `live` project) against the production site immediately after deploy (add-to-cart → checkout →
   Stripe Payment Element, then stops before paying). If the live journey fails, the workflow
   **auto-rolls back** the Pages deployment to the last known-good SHA and redeploys it.

This deploy-then-verify-then-rollback loop is the part of this project worth reading closely if
you're evaluating CD design — see the comments in `deploy.yml` for the specifics (test gate,
rollback drill dispatch input, live verification).

## Stack

- SvelteKit + TypeScript (storefront)
- Cloudflare Workers + KV + R2 (edge proxy, sessions, carts, images)
- Cloudflare Pages (storefront hosting)
- Go (API)
- Fly.io (API hosting)
- Stripe (payments), Shippo (shipping), Resend (email), ntfy (order alerts)
- Playwright (E2E, local + live projects)

## License

MIT (source code). See [`LICENSE`](./LICENSE) — brand photography/model images are excluded.
