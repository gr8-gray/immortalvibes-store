# Immortal Vibes

E-commerce storefront for [theimmortalvibes.com](https://theimmortalvibes.com), built as two
deployables:

- **`web/`**: SvelteKit storefront (Cloudflare Pages). Its server route proxies every `/api/*` call to the API.
- **`api/`**: Go REST API (Fly.io + Postgres): products, cart, checkout (Stripe), orders, shipping (Shippo),
  admin, email (Resend), order-alert notifications (ntfy).

## Architecture

```
browser → Cloudflare Pages (web/)
            → web/src/routes/api/[...path]/+server.ts   (server-side proxy, injects PROXY_SECRET)
              → Go API (api/, Fly.io)                    (products, cart, checkout, orders, shipping)
                  ↳ Postgres (Fly) · KV carts + R2 images (Cloudflare) · Stripe · Shippo · Resend
```

The storefront never calls the Go API from the browser. Every `/api/*` request is proxied server-side by
`web/src/routes/api/[...path]/+server.ts`, which attaches the shared `PROXY_SECRET`; the API rejects
requests without it (the Stripe webhook route is exempt and verified by Stripe signature instead). An
earlier Cloudflare Worker tier between the two was removed in #18.

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

`.github/workflows/deploy.yml` runs on push to `main`. It's the only production deploy route:

1. **Test gate** (`test-api`): the full Go API test suite must pass before anything deploys.
2. **Deploy API** (`deploy-api`, only when `api/` changed): records the running image, deploys to Fly, runs
   `.github/scripts/smoke-api.sh`, and rolls back to the recorded image if the smoke test fails.
3. **Deploy Pages** (`deploy-pages`, in parallel with the API): builds the storefront and deploys to Cloudflare Pages.
4. **Post-deploy live verification** (`verify-live`, after Pages): runs the full live customer journey
   (`web/e2e/storefront.spec.ts`, `live` project) against the production site (add-to-cart → checkout →
   Stripe Payment Element, stopping before payment). If it fails, `rollback` redeploys the previous commit
   to Pages.

Both rollback paths have drill inputs (`rollback_drill`, `api_rollback_drill`) to exercise them on demand.
Current proof status of each safety net is tracked in [`docs/lessons/OPEN-WORK.md`](docs/lessons/OPEN-WORK.md).
This deploy-then-verify-then-rollback loop is the part of this project worth reading closely if
you're evaluating CD design — see the comments in `deploy.yml` for the specifics (test gate,
rollback drill dispatch input, live verification).

## Stack

- SvelteKit + TypeScript (storefront)
- Cloudflare KV + R2 (carts, images)
- Cloudflare Pages (storefront hosting)
- Go (API)
- Fly.io (API hosting)
- Stripe (payments), Shippo (shipping), Resend (email), ntfy (order alerts)
- Playwright (E2E, local + live projects)

## License

MIT (source code). See [`LICENSE`](./LICENSE) — brand photography/model images are excluded.
