// web/e2e/storefront.spec.ts
//
// The buyer's journey, walked end to end and stopped at the exact moment
// Stripe takes over:
//
//   /shop (planet row) -> set landing page -> product page -> size pick ->
//   add to cart -> cart drawer -> /checkout -> address form ->
//   CONTINUE TO PAYMENT -> Stripe Payment Element mounts -> STOP.
//
// THE HARD RULE: no spec here may ever fill card details or submit the
// payment form. The suite asserts that Stripe's element arrived (the
// "redirect point" of this embedded flow) and goes no further. On the live
// project this still creates a *pending* order row — it is deliberately
// labelled so the owner can recognise and ignore it — but no PaymentIntent
// is ever confirmed and no money can move.
//
// The same specs run under two projects (see playwright.config.ts):
//   local — vite preview + network-level mocks of the Go API. Hermetic.
//   live  — the deployed site, real API, real Stripe key, same hard stop.
import { test, expect, type Page } from '@playwright/test';

// Catalog facts the assertions key on. These mirror web/src/lib/types/shop
// (the single source) — restated here as plain strings because the e2e
// suite deliberately tests the built app from the outside, not the module
// graph from the inside.
const TEE = {
  slug: 'warped-reality-tee',
  name: 'Warped Reality Tee',
  priceId: 'price_1TTl9AHy1AKk8SUWeLENS5Bu',
  size: 'M',
};

const isLive = () => test.info().project.name === 'live';

// ─── Local-gate API mocks ──────────────────────────────────────────────────
// The local build cannot reach the Go API (the proxy has no PROXY_SECRET in
// preview, so upstream answers 403 and every loader falls back to the static
// catalog — that fallback is itself part of what we're testing). Cart and
// checkout are browser-initiated, so we answer them at the network edge with
// the same shapes the Go API returns.
const FAKE_CART = {
  token: 'e2e-cart-token',
  line_items: [
    {
      price_id: TEE.priceId,
      product_id: 'prod_e2e_tee',
      name: TEE.name,
      image_url: '/photos/_drop/tee-front.png',
      currency: 'usd',
      amount: 3000,
      quantity: 1,
      size: TEE.size,
    },
  ],
};

async function mockApiForLocal(page: Page) {
  if (isLive()) return;

  // Layout hydration asks for the current cart on every first mount.
  await page.route('**/api/cart', (route) => {
    if (route.request().method() === 'POST') {
      return route.fulfill({ json: FAKE_CART });
    }
    return route.fulfill({ json: { token: '', line_items: [] } });
  });

  // Checkout session: hand back a client_secret that satisfies Stripe.js's
  // client-side shape check. The element will mount, try to talk to Stripe
  // with the dummy key, and show its own error state — but by then the app
  // has already done everything we claim it does, and the suite has stopped.
  await page.route('**/api/checkout', (route) =>
    route.fulfill({ json: { client_secret: 'pi_3E2ELocal000000_secret_e2elocalfake', order_id: 'e2e-order-local' } })
  );

  // Shipping estimates debounce off the address fields; answer instantly.
  await page.route('**/api/shipping/estimate', (route) =>
    route.fulfill({ json: { rate: { amount: 500, currency: 'usd', name: 'E2E Ground' } } })
  );
}

test.describe('storefront purchase journey (stops at Stripe)', () => {
  test('shop page renders the mission planet row', async ({ page }) => {
    await mockApiForLocal(page);
    await page.goto('/shop');

    await expect(page.locator('.select-label')).toHaveText('MISSION SELECT');
    // The Warped Reality set planet links to the set landing page, and the
    // two standalone missions link straight to their product pages.
    await expect(page.locator(`a[href="/shop/warped-reality-set"]`)).toBeVisible();
    await expect(page.locator('.mission-label .name', { hasText: 'Racerback Tanktop' })).toBeVisible();
    await expect(page.locator('.mission-label .name', { hasText: 'Phantom Basketball Shorts' })).toBeVisible();
  });

  test('set landing page lists the three pieces with catalog names', async ({ page }) => {
    await mockApiForLocal(page);
    await page.goto('/shop/warped-reality-set');

    for (const name of ['Warped Reality Tee', 'Warped Reality Beanie', 'Warped Reality Sweatpants']) {
      await expect(page.locator('.piece-name', { hasText: name })).toBeVisible();
    }
    await expect(page.locator(`a[href="/shop/${TEE.slug}"]`)).toBeVisible();
  });

  test('product page: size selection, add to cart, drawer, checkout, Stripe — then stop', async ({ page }) => {
    await mockApiForLocal(page);

    // ── Product page with a size selector ──
    await page.goto(`/shop/${TEE.slug}`);
    await expect(page.getByRole('heading', { name: TEE.name })).toBeVisible();

    const sizeGroup = page.getByRole('radiogroup', { name: 'Select size' });
    await expect(sizeGroup).toBeVisible();
    await sizeGroup.getByText(TEE.size, { exact: true }).click();

    // ── Add to cart -> drawer shows the line ──
    const addToCartRequest = page.waitForRequest(
      (req) => req.url().includes('/api/cart') && req.method() === 'POST'
    );
    await page.getByRole('button', { name: 'ADD TO CART' }).click();
    await addToCartRequest;

    await expect(page.locator('.item-title', { hasText: TEE.name })).toBeVisible();
    await expect(page.locator('.item-size', { hasText: TEE.size })).toBeVisible();

    // ── Checkout page with the order summary ──
    await page.getByRole('button', { name: 'PROCEED TO CHECKOUT' }).click();
    await expect(page).toHaveURL(/\/checkout$/);
    await expect(page.locator('.summary-item', { hasText: TEE.name })).toBeVisible();

    // ── Address form. On live this order is loudly labelled as a test. ──
    await page.locator('#email').fill('storefront-e2e@theimmortalvibes.com');
    await page.getByPlaceholder('Full name').fill('E2E GATE - DO NOT FULFILL');
    await page.getByPlaceholder('Address line 1').fill('1 Test Harness Way');
    await page.getByPlaceholder('City').fill('New Orleans');
    await page.getByPlaceholder('State').fill('LA');
    await page.getByPlaceholder('ZIP / Postal code').fill('70115');

    // ── The Stripe handoff: session created, Payment Element mounts ──
    const checkoutRequest = page.waitForRequest(
      (req) => req.url().includes('/api/checkout') && req.method() === 'POST'
    );
    await page.getByRole('button', { name: 'CONTINUE TO PAYMENT' }).click();
    await checkoutRequest;

    // Stripe has taken over when its element iframe attaches inside the
    // mount node. This is the line the suite never crosses: we assert the
    // iframe exists and deliberately interact with nothing inside it.
    await expect(
      page.locator('.payment-element-mount iframe[src*="js.stripe.com"]').first()
    ).toBeAttached({ timeout: 30_000 });

    // HARD STOP. No card fields, no PAY NOW, no confirmPayment — ever.
  });
});
