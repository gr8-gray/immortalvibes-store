# web — Immortal Vibes storefront

SvelteKit storefront for [theimmortalvibes.com](https://theimmortalvibes.com). See the repo-root
[`README.md`](../README.md) for the full architecture (storefront → in-app proxy → Worker → Go API).

## Develop

```sh
npm install
npm run dev
```

## Build

```sh
npm run build
```

Preview the production build with `npm run preview`.

## Test

```sh
npm run build && npm run e2e:local
```

Do not run `npm run e2e:live` locally — it targets the production site and creates a real
Stripe-pending order. It's reserved for the post-deploy verification step in CI.

## Environment

Copy `.env.example` to `.env` and fill in `PUBLIC_API_URL`.
