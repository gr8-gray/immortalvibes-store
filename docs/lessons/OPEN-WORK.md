# Open work and known gaps

What's unfinished, and which safety nets aren't proven yet. Each item says its status, so nobody has to
guess whether something works. Update this file when an item changes. Last full review: 2026-11-22.

## Safety nets: status

| Item | Status | What closes it |
|---|---|---|
| Pages `rollback` job | **Unproven.** It pointed at the deleted `worker/package-lock.json` (fixed in PR #23). No drill has run since July. | Merge #23, then (ask first) run Actions → `deploy` → *Run workflow* from `main` only (a dispatch from another branch deploys that branch to production) with `rollback_drill = true`. It redeploys current HEAD through the rollback path, re-runs the live journey, and leaves one test order (don't fulfil it). |
| API image rollback in `deploy-api` | **Unproven.** Never drilled. Fires on any failure after the running image is captured (the Fly deploy or the smoke test). | Ask first, then dispatch from `main` with `api_rollback_drill = true`. |
| Branch protection on `main` | **Off.** "PR + green CI, no direct push" is a convention. | Settings → Branches: require a PR and the `test-api` check. Don't require `e2e-local` yet: it's path-filtered to `web/`, so on API-only or docs PRs it never reports and the PR could never merge. |
| Uptime monitoring | **The owner has none.** The only monitor runs on the previous maintainer's own machine. | Add a monitor (any uptime service) probing `theimmortalvibes.com` **and** `www.theimmortalvibes.com` (RULES IV-10). |
| Deploy and order alerts | `deploy.yml` sends ntfy/Telegram deploy alerts, and the API sends ntfy order alerts, through secrets configured by the previous maintainer. **They may reach him, not the owner.** | Re-point `NTFY_*` + `TELEGRAM_*` (GitHub Actions) and `NTFY_*` (Fly) at the owner's own channel during the account handoff. |
| Leftover Cloudflare Worker | The repo removed it in #18. **Not verified** whether a deployed Worker still exists in the Cloudflare account. | Check Workers & Pages in the dashboard. If one exists with no route, delete it. If it has a route, that's a second production path (RULES IV-2), so remove the route first. |

## Code follow-ups (from the 2026-10-10 checkout hardening)

1. **Fulfil asynchronously.** The Stripe webhook buys the label and sends email synchronously, which can exceed
   Stripe's wait. Acknowledge first, then fulfil from a durable record. The email client also needs a timeout.
   This is the money path, so it gets its own PR and its STOP 19 test.
2. **Rate-limit `/api/promo/validate`.**
3. **Strict promo redemption limits** need a reservation record. That's a schema change, so ask first. Today
   limits are best-effort via Stripe search.
4. **Allow-list variant labels** (look-alike Unicode characters can slip through today).
5. **Confirm the Stripe account's country supports Search** (it isn't available in every country). The
   redemption checks rely on it.

## Product and measurement

- **Conversion is measured wrong by default.** PaymentIntents in `requires_payment_method` are visitors who
  reached the payment form (RULES IV-8). Measure sessions reaching `/checkout` versus paid orders, and exclude
  test orders (RULES IV-9).
- **Size-and-colour combinations** aren't a stock dimension (stock is per product + one variant label).
- `npm run check` has 14 pre-existing TypeScript errors (measured 2026-11-22; Svelte component exports, GSAP tween types). They don't
  break anything. Clean up when convenient.

## Repo housekeeping

- **3 stale remote branches**, all from PRs that are already merged (squash merges, so git doesn't show them as
  merged): `feat/phantom-dual-axis-stock`, `fix/planet-product-overlays`, `fix/vite-dev-allowedhosts`. Safe to
  delete (ask first). **Don't merge any of them**: they predate later work on `main`, and merging would revert it
  (RULES IV-12). `fix/pages-rollback-cache-path` is open PR #23, so don't delete it.
- **Commit metadata in existing history.** Some older commits carry a personal email as author or committer.
  Changing that means rewriting public history (a force push), so it needs a deliberate decision. New commits
  use the GitHub noreply address.
- `wrangler.toml` at the root names the project `immortalvibes`. CI deploys with
  `--project-name immortalvibes-store`, and that flag is what counts.
