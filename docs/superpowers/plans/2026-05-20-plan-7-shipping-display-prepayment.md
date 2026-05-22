# Plan 7 — Shipping Cost Display Pre-Payment
_Immortal Vibes · Queued 2026-05-20 — plan next session with Sebo_

## Problem
Customers pay without seeing shipping cost. Shippo is only invoked post-payment
(in webhook handler). No shipping estimate shown in cart drawer or checkout summary.

## Scope
1. Shipping estimate endpoint on Go API (`/api/shipping/estimate`) — calls Shippo rates API with address + package weight
2. Checkout UI collects shipping address early, calls estimate endpoint
3. Shipping cost displayed in order summary before PAY NOW
4. PaymentIntent amount updated to include shipping
5. Wire to existing Shippo spec (`docs/superpowers/specs/2026-04-13-shipping-automation-design.md`)

## Dependencies
- Shippo API key must be set in Fly.io secrets (SHIPPO_API_KEY)
- Existing shipping automation spec covers post-payment label creation — this plan covers pre-payment display only

## Status
QUEUED — plan with Sebo before executing
