# Shipping Automation — Implementation Plan
_Immortal Vibes · 2026-04-13_
_Spec: docs/superpowers/specs/2026-04-13-shipping-automation-design.md_

---

## Task 1 — Config: add 8 new env vars
**File:** `api/config/config.go`

Add to `Config` struct:
```go
ShippoAPIKey      string
ShippoFromName    string
ShippoFromStreet1 string
ShippoFromCity    string
ShippoFromState   string
ShippoFromZip     string
ShippoFromCountry string
OwnerEmail        string
```

Add to `Load()`:
```go
ShippoAPIKey:      os.Getenv("SHIPPO_API_KEY"),
ShippoFromName:    os.Getenv("SHIPPO_FROM_NAME"),
ShippoFromStreet1: os.Getenv("SHIPPO_FROM_STREET1"),
ShippoFromCity:    os.Getenv("SHIPPO_FROM_CITY"),
ShippoFromState:   os.Getenv("SHIPPO_FROM_STATE"),
ShippoFromZip:     os.Getenv("SHIPPO_FROM_ZIP"),
ShippoFromCountry: getEnv("SHIPPO_FROM_COUNTRY", "US"),
OwnerEmail:        os.Getenv("OWNER_EMAIL"),
```

Add to missing check (all except ShippoFromCountry which has a default):
`SHIPPO_API_KEY`, `SHIPPO_FROM_NAME`, `SHIPPO_FROM_STREET1`, `SHIPPO_FROM_CITY`, `SHIPPO_FROM_STATE`, `SHIPPO_FROM_ZIP`, `OWNER_EMAIL`

---

## Task 2 — DB layer: migration, OrderRow, new methods, updated SELECTs
**File:** `api/store/db.go`

### 2a. migrate() — append after existing CREATE TABLE statements:
```sql
ALTER TABLE orders ADD COLUMN IF NOT EXISTS shipping_name  TEXT;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS line1          TEXT;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS line2          TEXT;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS city           TEXT;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS state          TEXT;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS postal_code    TEXT;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS country        TEXT;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS tracking_number TEXT;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS carrier        TEXT;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS label_url      TEXT;
```

### 2b. OrderRow — add fields:
```go
ShippingName   string
Line1          string
Line2          string
City           string
State          string
PostalCode     string
Country        string
TrackingNumber string
Carrier        string
LabelURL       string
```

### 2c. SaveOrder — update INSERT to include 7 address fields:
```sql
INSERT INTO orders (id, payment_intent_id, cart_token, email, currency, total_amount, status,
                    shipping_name, line1, line2, city, state, postal_code, country, created_at)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,NOW())
```
Add params: `o.ShippingName, o.Line1, o.Line2, o.City, o.State, o.PostalCode, o.Country`

### 2d. New method UpdateOrderShipping:
```go
func (d *DB) UpdateOrderShipping(ctx context.Context, id, trackingNumber, carrier, labelURL string) error {
    _, err := d.db.ExecContext(ctx, `
        UPDATE orders SET tracking_number=$2, carrier=$3, label_url=$4 WHERE id=$1
    `, id, trackingNumber, carrier, labelURL)
    return err
}
```

### 2e. Update GetOrder + GetOrderByPaymentIntent:
Both SELECTs must include all 10 new columns. Update SELECT list and Scan calls to include:
`shipping_name, line1, line2, city, state, postal_code, country, tracking_number, carrier, label_url`
Scan into corresponding OrderRow fields (use `sql.NullString` for nullable fields or scan to `string` directly — nullable cols default to empty string).

Use `&o.ShippingName, &o.Line1, &o.Line2, &o.City, &o.State, &o.PostalCode, &o.Country, &o.TrackingNumber, &o.Carrier, &o.LabelURL` in Scan.

---

## Task 3 — Shippo client: new package
**File:** `api/shippo/client.go` (create)

```go
package shippo

// Client makes Shippo REST API calls.
type Client struct {
    apiKey   string
    fromAddr Address
    http     *http.Client
}

type Address struct {
    Name    string
    Street1 string
    City    string
    State   string
    Zip     string
    Country string
}

// NewClient constructs a Shippo client with a fixed from-address.
func NewClient(apiKey string, from Address) *Client

// RateShop creates a shipment and returns the object_id of the cheapest valid rate.
// Parcel is fixed: 12×9×1 in, 8 oz.
func (c *Client) RateShop(ctx context.Context, to Address) (rateID string, err error)

// BuyLabel purchases a label for the given rate ID.
// Returns tracking number, carrier name, and PDF label URL.
func (c *Client) BuyLabel(ctx context.Context, rateID string) (trackingNumber, carrier, labelURL string, err error)
```

Internal helpers:
- `shipmentRequest` / `shipmentResponse` JSON structs matching Shippo API
- `transactionRequest` / `transactionResponse` JSON structs
- Rate selection: iterate `response.Rates`, parse `Amount` as float64, pick minimum where `ObjectStatus == "VALID"`
- Auth header: `Authorization: ShippoToken <apiKey>`
- Both calls use `"async": false`

---

## Task 4 — Email: 3 new methods
**File:** `api/email/sender.go`

### SendShippingLabel (to owner)
```go
func (s *Sender) SendShippingLabel(ctx context.Context, ownerEmail, orderID, labelURL, trackingNum, carrier string) error
```
Subject: `[IV Order] Label ready — <orderID>`
HTML: order ID, carrier, tracking number, label PDF link as `<a href="...">Download Label</a>`, reminder to ship within 2 business days.

### SendTrackingUpdate (to customer)
```go
func (s *Sender) SendTrackingUpdate(ctx context.Context, customerEmail, orderID, trackingNum, carrier string) error
```
Subject: `Your Immortal Vibes order has shipped`
HTML: order ID, carrier, tracking number. Keep brand voice — brief, clean.

### SendShippingFailure (to owner)
```go
func (s *Sender) SendShippingFailure(ctx context.Context, ownerEmail, orderID, customerEmail, shippingAddr, errMsg string) error
```
Subject: `[IV Order] SHIPPING FAILED — manual label needed — <orderID>`
HTML: order ID, customer email, full shipping address block, raw error, link to goshippo.com dashboard.

---

## Task 5 — Webhook: new interfaces + Shippo sequence
**File:** `api/handlers/webhook.go`

### New interface ShippoClient:
```go
type ShippoClient interface {
    RateShop(ctx context.Context, to shippo.Address) (rateID string, err error)
    BuyLabel(ctx context.Context, rateID string) (trackingNumber, carrier, labelURL string, err error)
}
```

### Extend WebhookOrderStore interface:
Add:
```go
UpdateOrderShipping(ctx context.Context, id, trackingNumber, carrier, labelURL string) error
```

### Extend EmailSender interface:
Add:
```go
SendShippingLabel(ctx context.Context, ownerEmail, orderID, labelURL, trackingNum, carrier string) error
SendTrackingUpdate(ctx context.Context, customerEmail, orderID, trackingNum, carrier string) error
SendShippingFailure(ctx context.Context, ownerEmail, orderID, customerEmail, shippingAddr, errMsg string) error
```

### Extend WebhookHandler struct:
Add fields:
```go
shippo     ShippoClient
ownerEmail string
```

### Update NewWebhookHandler signature:
Add `shippoClient ShippoClient, ownerEmail string` params.

### Update handlePaymentIntentSucceeded:
After `UpdateOrderStatus("complete")` succeeds, add:

```go
// Ship it — non-fatal block
go func() {
    toAddr := shippo.Address{
        Name:    order.ShippingName,
        Street1: order.Line1,
        City:    order.City,
        State:   order.State,
        Zip:     order.PostalCode,
        Country: order.Country,
    }
    rateID, err := h.shippo.RateShop(context.Background(), toAddr)
    if err != nil {
        h.notifyShippingFailure(order, err.Error())
        return
    }
    trackingNum, carrier, labelURL, err := h.shippo.BuyLabel(context.Background(), rateID)
    if err != nil {
        h.notifyShippingFailure(order, err.Error())
        return
    }
    if err := h.db.UpdateOrderShipping(context.Background(), order.ID, trackingNum, carrier, labelURL); err != nil {
        log.Printf("webhook: UpdateOrderShipping(%s): %v", order.ID, err)
    }
    if err := h.emailer.SendShippingLabel(context.Background(), h.ownerEmail, order.ID, labelURL, trackingNum, carrier); err != nil {
        log.Printf("webhook: SendShippingLabel: %v", err)
    }
    if err := h.emailer.SendTrackingUpdate(context.Background(), order.Email, order.ID, trackingNum, carrier); err != nil {
        log.Printf("webhook: SendTrackingUpdate: %v", err)
    }
}()
```

Add private helper:
```go
func (h *WebhookHandler) notifyShippingFailure(order *store.OrderRow, errMsg string) {
    addr := fmt.Sprintf("%s\n%s\n%s, %s %s\n%s",
        order.ShippingName, order.Line1, order.City, order.State, order.PostalCode, order.Country)
    if err := h.emailer.SendShippingFailure(context.Background(), h.ownerEmail, order.ID, order.Email, addr, errMsg); err != nil {
        log.Printf("webhook: SendShippingFailure: %v", err)
    }
}
```

Note: Shippo calls run in a goroutine so Stripe gets a fast 200. The webhook already returns before the goroutine completes.

---

## Task 6 — Checkout handler: address fields
**File:** `api/handlers/checkout.go`

### Extend CheckoutRequest:
```go
ShippingName string `json:"shipping_name"`
Line1        string `json:"line1"`
Line2        string `json:"line2"`
City         string `json:"city"`
State        string `json:"state"`
PostalCode   string `json:"postal_code"`
Country      string `json:"country"`
```

### Add address validation after cart validation:
```go
if req.ShippingName == "" || req.Line1 == "" || req.City == "" || req.State == "" || req.PostalCode == "" || req.Country == "" {
    http.Error(w, "shipping address required", http.StatusBadRequest)
    return
}
```

### Update SaveOrder call to include address fields:
Pass all 7 address fields from `req` into `store.OrderRow`.

---

## Task 7 — Frontend: address form
**File:** `web/src/routes/checkout/+page.svelte`

Add an address section above the Stripe Payment Element. Fields:
- Full name (shipping_name) — required
- Address line 1 (line1) — required
- Address line 2 (line2) — optional
- City — required
- State (state) — required
- ZIP / Postal code (postal_code) — required
- Country (country) — required, default "US"

Style: match existing checkout form inputs (Lunar White bg, Inter font, same border/focus styles).

Include all 7 fields in the POST body to `/api/checkout`.

---

## Task 8 — main.go: wire everything
**File:** `api/main.go`

1. Import `api/shippo`
2. Construct Shippo client after config load:
```go
shippoClient := shippo.NewClient(cfg.ShippoAPIKey, shippo.Address{
    Name:    cfg.ShippoFromName,
    Street1: cfg.ShippoFromStreet1,
    City:    cfg.ShippoFromCity,
    State:   cfg.ShippoFromState,
    Zip:     cfg.ShippoFromZip,
    Country: cfg.ShippoFromCountry,
})
```
3. Pass `shippoClient` and `cfg.OwnerEmail` to `NewWebhookHandler`.

---

## Task 9 — Fly secrets: set all new vars
Run after code is deployed:

```bash
fly secrets set \
  SHIPPO_API_KEY=<from Shippo dashboard> \
  SHIPPO_FROM_NAME="<owner name>" \
  SHIPPO_FROM_STREET1="<owner street>" \
  SHIPPO_FROM_CITY="<owner city>" \
  SHIPPO_FROM_STATE="<2-letter state>" \
  SHIPPO_FROM_ZIP="<zip>" \
  OWNER_EMAIL="immortalvibesco@gmail.com" \
  --app immortalvibes-api
```

**Prerequisite:** Owner must sign up at goshippo.com and retrieve API key before this step.

Verify after deploy:
```bash
fly secrets list --app immortalvibes-api
curl https://immortalvibes-api.fly.dev/health
```

---

## Execution Order

Tasks 1–3 have no dependencies — can be done in parallel.
Task 4 depends on Task 3 (needs Sender struct).
Task 5 depends on Tasks 2, 3, 4 (needs interfaces + types).
Task 6 depends on Task 2 (needs OrderRow address fields).
Task 7 is frontend-only, independent.
Task 8 depends on Tasks 1, 3, 5.
Task 9 is last — deploy then set secrets.
