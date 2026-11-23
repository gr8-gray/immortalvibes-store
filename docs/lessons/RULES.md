# Rules for this codebase

Every rule here was learned from a real mistake on this store. Claude Code sessions treat them as hard
constraints. Each entry gives the **rule**, **why** it exists, and **how to apply** it in this repo.

- `STOP <n>` numbers are stable IDs carried over from the previous maintainer's ledger. Workflow comments
  and the PR template cite them (STOP 9, 19, 21), so don't renumber.
- `IV-<n>` are lessons specific to this store.
- Longer stories behind the rules are in [`incidents/`](./incidents/).

Add a rule when a mistake is corrected and its root cause is understood. Keep it short.

---

## Shipping and deploys

### STOP 1: Merged is not deployed
- **Why:** changes have been reported "live" while the deploy job was skipped, failing, or running old code.
- **How:** after a merge, open the `deploy` run. Confirm `test-api` passed, `deploy-pages` and `verify-live`
  passed, and, for `api/` changes, `deploy-api` actually deployed (it skips when `api/` is unchanged) and its
  **Smoke the API** step passed. A skipped job is not a passed job.

### STOP 28 (candidate): A safety net isn't real until it has been seen to fail
- **Why:** the Pages `rollback` job pointed at a deleted `worker/package-lock.json` for weeks. It looked
  configured and would have failed the one time it was needed.
- **How:** prove a check or rollback works with its drill (`workflow_dispatch`: `rollback_drill` for Pages,
  `api_rollback_drill` for the API) before trusting it. See OPEN-WORK for current status.

### IV-2: Keep exactly one production deploy route
- **Why:** the Cloudflare Pages Git integration kept deploying the site while the GitHub Actions token was
  dead, so `verify-live` and `rollback` were silently skipped for 3 weeks. See
  [incident](./incidents/2026-10-02-silent-second-deploy-route.md).
- **How:** production deploys happen only through `.github/workflows/deploy.yml`. Pages Git integration has
  production deploys **disabled**. An idle `github:push` record per push is normal; one with a non-idle
  build stage means the second route came back. Never deploy by hand around the pipeline.

### IV-17: Config changes are a production deploy that the pipeline doesn't cover
- **Why:** required env vars crash-looped the API on boot, and missing `FROM_*` Fly secrets silently broke
  shipping. `fly secrets set` creates a new release with the same image, so the image rollback can't undo it.
- **How:** secret or config changes are ask-first. Change one value at a time, then immediately run
  `.github/scripts/smoke-api.sh` (or watch `fly logs`) and check `/api/products` and a shipping quote. Keep
  the old value until the new one is verified. New optional integrations must not be required at boot.

### IV-20: Know which Cloudflare account you're deploying to
- **Why:** this project has lived in two Cloudflare accounts, and early deploys went to the wrong one.
- **How:** CI uses the `CLOUDFLARE_ACCOUNT_ID` secret. For any manual `wrangler` command, run
  `wrangler whoami` first and confirm the account that owns the `theimmortalvibes.com` zone.

## Money path (checkout, orders, fulfilment)

### STOP 18: An order persists its own contents at creation
- **Why:** an order once paid with no record of what was bought, because the cart was deleted on payment and
  the order row held only an amount. The owner got a shipping label and no packing list. See
  [incident](./incidents/2026-06-21-fulfilment-chain.md).
- **How:** `orders.line_items` (item, size, qty, unit price) is written at checkout. Never make fulfilment
  depend on cart/session state, which is cleared on payment.

### STOP 19: Every customer-facing change keeps an E2E test of the money→shipment path
- **Why:** the same incident passed every unit test. Nothing asserted the owner could actually fulfil.
- **How:** `TestWebhookFulfillmentE2E` (Go) and `web/e2e/storefront.spec.ts` (Playwright) must stay green
  and must assert outcomes the owner depends on (manifest in the owner email, stock decremented once,
  customer notified). There is **no staging stack**: run `go test ./...` and `npm run e2e:local` locally.
  The `live` project runs only in CI or by manual dispatch.

### IV-7: Prices, currency, quantities and shipping are computed server-side
- **Why:** the client total used to be trusted. It's now priced from the Stripe catalog (#21), shipping is
  quoted server-side (#20), and promo codes resolve on the server (#22). See
  [incident](./incidents/2026-10-10-checkout-pricing-hardening.md).
- **How:** never accept a price, currency, discount or shipping amount from the browser. The UI mirrors
  server math (`web/src/lib/pricing.ts`) for display only. The store always charges the catalog currency
  (USD). The old multi-currency-by-country design is superseded; don't revive it.

### IV-3: Never do fulfilment work in a fire-and-forget goroutine
- **Why:** Fly scale-to-zero killed a background shipping goroutine mid-order.
- **How:** `min_machines_running = 1` in `api/fly.toml` stays. Fulfilment runs synchronously and
  idempotently, or from a durable record. A goroutine with no persistence is not a queue.

### IV-4: Check middleware order when adding an exemption
- **Why:** Stripe webhooks got 403s because the global proxy-auth middleware ran before the route's exemption.
- **How:** exempt by path at the middleware that enforces it (see `api/middleware/`), and add a test that
  calls the exempt route without the proxy secret.

### IV-5: Unshipped orders have NULL shipping columns
- **Why:** scanning `tracking`/`carrier`/`label_url` into plain strings crashed order lookups for every
  unshipped order.
- **How:** `COALESCE(col, '')` in queries (or nullable types), plus a test with an unshipped order.

### IV-6: Shippo/USPS needs a from-email and from-phone, and a billing method
- **Why:** label purchase failed until the account had a payment method and `address_from` carried email and phone.
- **How:** the `FROM_*` Fly secrets must all be set (`api/config/config.go` lists them). Surface Shippo's real
  error text in logs.

### IV-16: Check third-party JSON against a real response, not the struct you declared
- **Why:** `attributes` was declared as a list of objects, Shippo returns a list of strings, and every
  successful rate response failed to unmarshal, so the store showed "no rate".
- **How:** capture one real response before writing the struct. Decode only fields you use.

### IV-15: A backend error fix needs the frontend to show it
- **Why:** three deploy cycles were spent piping real shipping errors into JSON while the checkout page
  hardcoded "Unable to calculate".
- **How:** grep the frontend for the message before changing the backend. Fix both sides in one change.

### IV-22: The cart cookie is HttpOnly
- **Why:** carts reset on reload because JS can't read `cart_token`, and nothing fetched the server cart.
- **How:** the cart hydrates from `GET /api/cart` in the root layout. Cart mutations adopt the server response.

## Products and stock

### IV-11: Web stock is a ceiling, not the truth
- **Why:** the owner sells in person and on Instagram, so web counts drift from physical stock. Per-size
  overselling of the last units is possible.
- **How:** reconcile against the owner's physical count before drops and campaigns. Stock writes go through
  the admin endpoint (`PUT /api/admin/products/{id}/stock`, optional `variant`) and are ask-first.

### IV-13: `/shop` matches products by exact name
- **Why:** the shop loader merges the live API by exact product **name** and derives availability from
  stock > 0. A rename in the Stripe dashboard silently breaks the match.
- **How:** rename in Stripe and in `web/src/lib/types/shop.ts` together, then update `web/e2e/storefront.spec.ts`,
  which asserts product names on purpose.

### IV-14: Create the Stripe product before marking a SKU available
- **Why:** a SKU marked available with no Stripe price ID fails at add-to-cart.
- **How:** product and price exist in Stripe (ask-first: live payment change), then flip the SKU, then verify
  add-to-cart on the live site.

## Measuring the store

### IV-8: A PaymentIntent in `requires_payment_method` means "reached the payment form"
- **Why:** `POST /api/checkout` creates the PaymentIntent at CONTINUE TO PAYMENT, before any card is entered.
  Counting those as failed payments once made a working checkout look 86% broken.
- **How:** measure conversion as sessions reaching `/checkout` versus paid orders, not PaymentIntent states.

### IV-9: Test orders are never fulfilled and never counted
- **Why:** `verify-live`, `rollback` and the weekly `e2e-live` each leave a real pending order named
  "E2E GATE – DO NOT FULFILL" in production. Diagnostic orders have been created by hand too.
- **How:** exclude them from fulfilment, stock and every sales or conversion number.

### IV-10: Monitor every hostname customers use
- **Why:** `www` returned 522 for 18 days while uptime checks probed only the apex and reported 200. See
  [incident](./incidents/2026-10-01-www-522.md).
- **How:** any uptime monitor checks both `theimmortalvibes.com` and `www.theimmortalvibes.com`.

## DNS and Cloudflare

### IV-1: A proxied CNAME to `*.pages.dev` is not an attached Pages domain
- **Why:** the `www` DNS record existed and was proxied, but `www` was never attached to the Pages project, so
  Cloudflare returned 522. Today `www` 301-redirects to the apex via a zone redirect rule.
- **How:** a hostname serves from Pages only once it's listed in the Pages project's custom domains. Verify with
  `curl -I` on every hostname after any DNS change. DNS changes are ask-first.

## Frontend and visuals

### STOP 14: "The scene starts too low" means move the horizon DOWN the screen
- **Why:** on the original canvas hero, "too low" meant too much dark ground was visible, so the horizon line
  sat too HIGH. The intuitive fix (decreasing the horizon fraction) made it worse. That scene was removed in
  the homepage redesign (`67247aa`), but the rule holds for any canvas horizon scene.
- **How:** in any horizon scene, "too low" means increase the horizon's base y-fraction.

### STOP 22: Consumer pages ship perceived-quality basics
- **Why:** a store that feels slow loses sales even when it works.
- **How:** four patterns, wherever they apply: (1) skeleton loaders, never a blank flash on data-backed views;
  (2) cached reads at the right layer (SvelteKit `load` + cache headers + Cloudflare edge); (3) optimistic UI
  with rollback on error for cart actions; (4) every icon-only button has an `aria-label` and a tooltip.
  Build them with what's installed (SvelteKit, GSAP, Three.js). A new UI library is ask-first.

### IV-18: Visual fixes are verified in rendered pixels
- **Why:** a "fixed" overlay was reported live while production still served the old build. A Playwright
  screenshot of prod versus the preview showed the difference.
- **How:** compare the production page against the Pages preview (screenshot or sampled pixels), at desktop
  and mobile widths. A green deploy is not visual proof.

### IV-21: Layout bugs: check geometry, not appearance
- **Why:** a fixed off-canvas cart drawer escaped `body { overflow-x: hidden }` and scrolled the page sideways
  (fixed with `html { overflow-x: clip }`). A discount banner was painted over by the nav because of
  `top`/z-index.
- **How:** measure with `getBoundingClientRect` in Playwright at 320, 375 and 1440 px. Check for horizontal
  overflow and overlap.

### IV-23: Planet textures: equirectangular or 3D-procedural only
- **Why:** AI-generated "planet" images are disks with black corners, which show as blank patches on a
  rotating sphere. UV-based noise shows a seam.
- **How:** use equirectangular textures or 3D position-based procedural shaders (the current planets).

### STOP 24: Grep the whole codebase before removing a UI element
- **Why:** a "hidden" feature kept showing up because the same content was rendered from a second place.
- **How:** `grep -rn` the component, class or id across `web/src` before deleting it.

## Working habits

### STOP 9: Never call something done without evidence
- **How:** test output, a green CI run, or the live result, shown in the session.

### STOP 10: Query the authoritative source first
- **How:** for "what's live" or "what's set", ask the system (Fly, Cloudflare, Stripe, the database, the deploy
  run) instead of reasoning about what's probably there.

### STOP 15: Read a config file before editing it
- **How:** open `fly.toml`, `wrangler.toml`, workflow YAML or `svelte.config.js` and read the relevant
  section before changing it. Never guess keys or paths.

### STOP 23: An empty or "all clear" result counts only after a positive control
- **How:** before trusting "no matches", "no errors" or "0 failing", show the same check finding a case you
  know exists.

### STOP 25: A documented mechanism isn't one until you find it in code
- **How:** before relying on a flag, file or job a doc mentions, grep for it. Docs drift; this file can too.

### STOP 21: Public repo text never identifies private people or accounts
- **Why:** this repo is public, and commits and PR text are permanent.
- **How:** no customer data, personal emails, phone numbers, addresses, account IDs or secret values in code,
  docs, commits or PRs. Commit with your GitHub noreply email.

### IV-12: Rebase a stale branch onto `main` before merging it
- **Why:** a drop branch cut weeks earlier was missing four fulfilment fixes. A naive merge would have reverted them.
- **How:** `git rebase origin/main` (or compare `git log main..branch` and `branch..main`) before merging any
  branch older than a few days.

### IV-19: Anchor `.gitignore` patterns
- **Why:** an unanchored pattern silently ignored real asset files.
- **How:** use `/pattern` for root-only ignores. After adding assets, `git status` must show them.
