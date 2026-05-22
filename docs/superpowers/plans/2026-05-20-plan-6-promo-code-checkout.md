# Plan 6 — Promo Code Checkout
_Immortal Vibes · Queued 2026-05-20 — plan next session with Sebo_

## Problem
VIBE10 promo code is distributed via email subscribe flow but is unredeemable at checkout.
Zero promo support in checkout UI, Go handler, or Stripe integration.

## Scope
1. Promo input field in checkout UI (`src/routes/checkout/`)
2. Validation endpoint on Go API (`/api/promo/validate`) — Stripe coupon lookup
3. `CheckoutRequest` struct updated with `discount_code` field
4. PaymentIntent creation passes discount to Stripe
5. Discount reflected in order summary before payment

## Estimate
~2-3 hours implementation

## Status
QUEUED — plan with Sebo before executing
