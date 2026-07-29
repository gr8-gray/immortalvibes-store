// web/playwright.config.ts
//
// Storefront E2E — two separate gates over the same specs:
//
//   local : builds nothing itself, but boots `vite preview` over the last
//           `npm run build` output and mocks the Go API at the network edge.
//           Fully hermetic — no live services touched, no secrets needed.
//   live  : runs the same user journey against the deployed site
//           (https://theimmortalvibes.com). Read-mostly, but the checkout
//           step does create a clearly-labelled pending order — see
//           e2e/storefront.spec.ts for the hard stop before payment.
//
// Select a gate with:  npx playwright test --project=local  (or live)
// CI sets E2E_TARGET=live for the live job so no preview server is booted.
import { defineConfig, devices } from '@playwright/test';

const LIVE_BASE = process.env.E2E_LIVE_BASE_URL ?? 'https://theimmortalvibes.com';
const LOCAL_BASE = 'http://127.0.0.1:4173';

export default defineConfig({
  testDir: './e2e',
  timeout: 90_000,
  expect: { timeout: 15_000 },
  fullyParallel: false, // the journey specs share a cart; keep them ordered
  retries: process.env.CI ? 1 : 0,
  reporter: process.env.CI ? [['list'], ['html', { open: 'never' }]] : [['list']],
  use: {
    trace: 'retain-on-failure',
    // The storefront is animation-heavy (planet shaders, page transitions).
    // Honouring reduced-motion keeps headless runs deterministic without
    // exercising a code path real users can't reach.
    contextOptions: { reducedMotion: 'reduce' },
  },
  projects: [
    { name: 'local', use: { ...devices['Desktop Chrome'], baseURL: LOCAL_BASE } },
    { name: 'live', use: { ...devices['Desktop Chrome'], baseURL: LIVE_BASE } },
  ],
  // The preview server only exists for the local gate. `reuseExistingServer`
  // stays false on purpose: a stale dev server once served pre-fix code and
  // burned a session — better to fail loudly on a busy port than test the
  // wrong build. The Stripe key is a syntactically-valid dummy; Stripe.js
  // accepts any pk_-prefixed string at construction time and the suite stops
  // before anything would validate it server-side.
  webServer:
    process.env.E2E_TARGET === 'live'
      ? undefined
      : {
          command: 'npm run preview -- --port 4173 --strictPort --host 127.0.0.1',
          url: `${LOCAL_BASE}/shop`,
          reuseExistingServer: false,
          timeout: 120_000,
          env: { PUBLIC_STRIPE_PUBLISHABLE_KEY: 'pk_test_e2e_local_dummy_key_0000' },
        },
});
