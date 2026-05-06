# Shipping Automation — Design Spec
_Immortal Vibes · 2026-04-13_

## Overview

When a Stripe payment succeeds, the Go webhook handler automatically rate-shops via Shippo, buys the cheapest label, emails the owner a PDF link, and emails the customer their tracking number. If Shippo fails at any point, the order remains complete and the owner receives a failure notification with full details to create the label manually.

---

## Flow

```
Customer fills address form + pays
    → POST /api/checkout (cart_token, email, address fields)
        → Stripe PaymentIntent created
        → Order saved to Postgres (status: pending, address stored)
        → client_secret returned to frontend

Stripe fires payment_intent.succeeded webhook
    → Go: UpdateOrderStatus → "complete"
    → Go: Shippo.RateShop(to=order address, from=config, parcel=fixed)
        → pick cheapest rate by amount
    → Go: Shippo.BuyLabel(cheapest rate ID)
        → tracking_number, carrier, label_url
    → Go: db.UpdateOrderShipping(tracking_number, carrier, label_url)
    → Go: email owner (immortalvibesco@gmail.com) — label PDF link + tracking
    → Go: email customer — carrier + tracking number

    If Shippo fails at any step:
    → Go: email owner — failure alert with order ID, customer email, full address, error message
    → Log error, return 200 to Stripe (don't retry)
```

---

## Parcel Profile (fixed)

| Field | Value |
|---|---|
| Length | 12 in |
| Width | 9 in |
| Height | 1 in |
| Weight | 8 oz |
| Mass unit | oz |
| Distance unit | in |
| Template | (none — flat dimensions used) |

Owner sourcing: poly mailers from Amazon in bulk.

---

## DB Schema Changes

`ALTER TABLE orders ADD COLUMN IF NOT EXISTS` for each:

| Column | Type | Notes |
|---|---|---|
| shipping_name | TEXT | Full name on package |
| line1 | TEXT | Street address |
| line2 | TEXT | Apt/suite (optional) |
| city | TEXT | |
| state | TEXT | 2-letter for US |
| postal_code | TEXT | |
| country | TEXT | ISO 2-letter, default "US" |
| tracking_number | TEXT | Set by webhook after label purchase |
| carrier | TEXT | e.g. "USPS", "UPS" |
| label_url | TEXT | Shippo PDF URL |

---

## New Env Vars (Fly secrets)

| Var | Required | Notes |
|---|---|---|
| SHIPPO_API_KEY | yes | From Shippo dashboard |
| SHIPPO_FROM_NAME | yes | Owner name on return address |
| SHIPPO_FROM_STREET1 | yes | Owner street |
| SHIPPO_FROM_CITY | yes | Owner city |
| SHIPPO_FROM_STATE | yes | Owner state (2-letter) |
| SHIPPO_FROM_ZIP | yes | Owner zip |
| SHIPPO_FROM_COUNTRY | no | Defaults to "US" |
| OWNER_EMAIL | yes | immortalvibesco@gmail.com |

---

## Shippo REST API Calls

### Rate Shop
```
POST https://api.goshippo.com/shipments/
Authorization: ShippoToken <SHIPPO_API_KEY>
{
  "address_from": { "name", "street1", "city", "state", "zip", "country" },
  "address_to":   { "name", "street1", "street2", "city", "state", "zip", "country" },
  "parcels": [{ "length": "12", "width": "9", "height": "1", "distance_unit": "in",
                "weight": "8", "mass_unit": "oz" }],
  "async": false
}
→ response.rates[] — pick min(amount) where object_status == "VALID"
```

### Buy Label
```
POST https://api.goshippo.com/transactions/
{
  "rate": "<rate_object_id>",
  "label_file_type": "PDF",
  "async": false
}
→ tracking_number, tracking_url_provider, label_url
```

---

## Email Templates

### Owner — Label Ready
**Subject:** `[IV Order] Label ready — <order_id>`
**Body:** Order ID, customer email, carrier, tracking number, label PDF link (click to download/print), shipping address summary.

### Owner — Shippo Failure
**Subject:** `[IV Order] SHIPPING FAILED — manual label needed — <order_id>`
**Body:** Order ID, customer email, full shipping address, raw error message. Instructions: create label manually at goshippo.com.

### Customer — Tracking
**Subject:** `Your Immortal Vibes order has shipped`
**Body:** Order ID, carrier name, tracking number, tracking URL (Shippo provider URL if available).

---

## Files Changed

| File | Change |
|---|---|
| `api/config/config.go` | 8 new env vars |
| `api/store/db.go` | Migration, OrderRow, SaveOrder, UpdateOrderShipping, updated SELECTs |
| `api/handlers/checkout.go` | Address fields in CheckoutRequest + validation |
| `api/shippo/client.go` | New package — RateShop + BuyLabel |
| `api/email/sender.go` | 3 new methods |
| `api/handlers/webhook.go` | New interfaces + Shippo sequence |
| `api/main.go` | Wire new config → constructors |
| `web/src/routes/checkout/+page.svelte` | Address form above Payment Element |

---

## Prerequisites

- Owner must create a Shippo account and obtain an API key
- Owner must set all 8 Fly secrets before deploying
- Verified sender domain on Resend must be active (already wired)
