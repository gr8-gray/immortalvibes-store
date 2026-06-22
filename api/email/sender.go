package email

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/immortalvibes/api/models"
)

// Sender dispatches transactional email via the Resend API.
type Sender struct {
	apiKey   string
	fromAddr string
	http     *http.Client
}

// NewSender creates a Sender. fromAddr is the verified Resend from address,
// e.g. "orders@immortalvibes.co.uk".
func NewSender(apiKey, fromAddr string) *Sender {
	return &Sender{
		apiKey:   apiKey,
		fromAddr: fromAddr,
		http:     &http.Client{},
	}
}

type resendPayload struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html"`
}

// SendOrderConfirmation sends an HTML order confirmation to the buyer.
func (s *Sender) SendOrderConfirmation(ctx context.Context, toEmail, orderID string, totalAmount int64, currency string) error {
	subject := fmt.Sprintf("Your Immortal Vibes order %s", orderID)
	html := fmt.Sprintf(`
		<h1>Order Confirmed</h1>
		<p>Thank you for your order!</p>
		<p><strong>Order ID:</strong> %s</p>
		<p><strong>Total:</strong> %s %d</p>
		<p>We'll be in touch when your order ships.</p>
	`, orderID, currency, totalAmount)

	payload := resendPayload{
		From:    s.fromAddr,
		To:      []string{toEmail},
		Subject: subject,
		HTML:    html,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.resend.com/emails", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.http.Do(req)
	if err != nil {
		return fmt.Errorf("resend request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("resend status %d: %s", resp.StatusCode, b)
	}
	return nil
}

// SendShippingLabel emails the owner a label-ready notification with PDF link,
// the PACKING MANIFEST (what to put in the box — STOP 18), plus a financial
// summary (order total, label cost billed to Shippo, and the resulting balance)
// so the owner has a complete record of the order.
func (s *Sender) SendShippingLabel(ctx context.Context, ownerEmail, orderID, labelURL, trackingNum, carrier, labelCost string, orderTotal int64, currency string, items []models.LineItem) error {
	subject := fmt.Sprintf("[IV Order] Label ready — %s", orderID)

	cur := strings.ToUpper(currency)
	totalStr := fmt.Sprintf("$%.2f %s", float64(orderTotal)/100, cur)
	if labelCost == "" {
		labelCost = "(see Shippo)"
	}
	// Balance = order total minus the label cost (excludes Stripe fees + COGS).
	balanceStr := "—"
	if amt := parseLeadingAmount(labelCost); amt >= 0 {
		balanceStr = fmt.Sprintf("$%.2f %s", float64(orderTotal)/100-amt, cur)
	}

	html := fmt.Sprintf(`
		<h2>Label Ready</h2>
		<p><strong>Order:</strong> %s</p>
		<p><strong>Carrier:</strong> %s</p>
		<p><strong>Tracking:</strong> %s</p>
		<p><a href="%s" style="font-weight:bold">Download Label (PDF)</a></p>
		<h3 style="margin:1.25rem 0 0.5rem">📦 Pack this</h3>
		%s
		<table style="margin:1rem 0;border-collapse:collapse;font-size:0.95em">
			<tr><td style="padding:2px 12px 2px 0;color:#666">Order total (collected via Stripe)</td><td style="padding:2px 0"><strong>%s</strong></td></tr>
			<tr><td style="padding:2px 12px 2px 0;color:#666">Shipping label (billed to Shippo)</td><td style="padding:2px 0">-%s</td></tr>
			<tr><td style="padding:2px 12px 2px 0;color:#666">After label (before Stripe fees &amp; product cost)</td><td style="padding:2px 0">%s</td></tr>
		</table>
		<p style="color:#888;font-size:0.85em">The label cost is charged to your Shippo account, not Stripe — Stripe only shows the order total.</p>
		<p style="color:#888;font-size:0.85em">Please ship within 2 business days. If you received an earlier "SHIPPING FAILED" notice for this order, disregard it — the label above is valid.</p>
	`, orderID, carrier, trackingNum, labelURL, renderManifest(items), totalStr, labelCost, balanceStr)
	return s.send(ctx, ownerEmail, subject, html)
}

// renderManifest builds the packing-list HTML table (what to put in the box).
// Falls back to a clear warning if no line items were persisted, so a blank
// manifest can never be mistaken for an empty order.
func renderManifest(items []models.LineItem) string {
	if len(items) == 0 {
		return `<p style="color:#c0392b"><strong>⚠ No item manifest on this order.</strong> ` +
			`Confirm contents with the customer before shipping.</p>`
	}
	var b strings.Builder
	b.WriteString(`<table style="border-collapse:collapse;font-size:0.95em;margin:0 0 0.5rem">`)
	b.WriteString(`<tr style="text-align:left"><th style="padding:4px 16px 4px 0">Item</th>` +
		`<th style="padding:4px 16px 4px 0">Size</th><th style="padding:4px 16px 4px 0">Qty</th>` +
		`<th style="padding:4px 0">Unit</th></tr>`)
	for _, li := range items {
		size := li.Size
		if size == "" {
			size = "—"
		}
		unit := fmt.Sprintf("$%.2f", float64(li.Amount)/100)
		b.WriteString(fmt.Sprintf(
			`<tr><td style="padding:4px 16px 4px 0">%s</td><td style="padding:4px 16px 4px 0">%s</td>`+
				`<td style="padding:4px 16px 4px 0">%d</td><td style="padding:4px 0">%s</td></tr>`,
			li.Name, size, li.Quantity, unit))
	}
	b.WriteString(`</table>`)
	return b.String()
}

// parseLeadingAmount extracts the leading numeric value from a string like
// "6.74 USD". Returns -1 if no number is present.
func parseLeadingAmount(s string) float64 {
	fields := strings.Fields(s)
	if len(fields) == 0 {
		return -1
	}
	v, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return -1
	}
	return v
}

// SendTrackingUpdate emails the customer their shipment tracking info.
func (s *Sender) SendTrackingUpdate(ctx context.Context, customerEmail, orderID, trackingNum, carrier string) error {
	subject := "Your Immortal Vibes order has shipped"
	html := fmt.Sprintf(`
		<h2>Your order is on its way.</h2>
		<p><strong>Order:</strong> %s</p>
		<p><strong>Carrier:</strong> %s</p>
		<p><strong>Tracking number:</strong> %s</p>
		<p style="color:#888;font-size:0.85em">Rise Beyond the Mortal Plane.</p>
	`, orderID, carrier, trackingNum)
	return s.send(ctx, customerEmail, subject, html)
}

// SendShippingFailure alerts the owner that label creation failed.
// Includes full order details so the owner can create the label manually.
func (s *Sender) SendShippingFailure(ctx context.Context, ownerEmail, orderID, customerEmail, shippingAddr, errMsg string) error {
	subject := fmt.Sprintf("[IV Order] SHIPPING FAILED — manual label needed — %s", orderID)
	html := fmt.Sprintf(`
		<h2 style="color:#c0392b">Shipping Automation Failed</h2>
		<p>A label could not be created automatically. Please create one manually.</p>
		<p><strong>Order:</strong> %s</p>
		<p><strong>Customer email:</strong> %s</p>
		<p><strong>Ship to:</strong></p>
		<pre style="background:#f4f4f4;padding:0.75rem">%s</pre>
		<p><strong>Error:</strong> <code>%s</code></p>
		<p><a href="https://apps.goshippo.com/orders" style="font-weight:bold">Create label in Shippo</a></p>
		<p style="color:#888;font-size:0.85em">Note: shipping is retried automatically. If you also receive a "Label ready" email for this same order, the label was created for you — ignore this notice and do not buy a second label.</p>
	`, orderID, customerEmail, shippingAddr, errMsg)
	return s.send(ctx, ownerEmail, subject, html)
}

// SendDiscountCode emails a subscriber their one-time discount code.
func (s *Sender) SendDiscountCode(ctx context.Context, toEmail, code string) error {
	subject := "Your Immortal Vibes discount code"
	html := fmt.Sprintf(`
		<div style="background:#030308;color:#F0EDE6;font-family:sans-serif;padding:2.5rem;max-width:520px;margin:0 auto">
			<p style="font-size:0.6rem;letter-spacing:0.4em;color:#C8922A;text-transform:uppercase;margin:0 0 1.5rem">Immortal Vibes</p>
			<h1 style="font-size:2rem;margin:0 0 1rem;font-weight:400">Rise Beyond 10%%</h1>
			<p style="color:rgba(240,237,230,0.6);line-height:1.6;margin:0 0 2rem">
				Your discount code is ready. Apply it at checkout on your first order.
			</p>
			<div style="background:rgba(200,146,42,0.1);border:1px solid rgba(200,146,42,0.4);padding:1.5rem;text-align:center;margin:0 0 2rem">
				<p style="font-size:1.8rem;letter-spacing:0.12em;color:#C8922A;margin:0;font-weight:700">%s</p>
				<p style="font-size:0.55rem;letter-spacing:0.3em;color:rgba(200,146,42,0.5);margin:0.5rem 0 0">ENTER AT CHECKOUT</p>
			</div>
			<p style="font-size:0.75rem;color:rgba(240,237,230,0.3);line-height:1.6;margin:0">
				One use per customer. New customers only. No expiry.
			</p>
		</div>
	`, code)
	return s.send(ctx, toEmail, subject, html)
}

func (s *Sender) send(ctx context.Context, toEmail, subject, html string) error {
	payload := resendPayload{
		From:    s.fromAddr,
		To:      []string{toEmail},
		Subject: subject,
		HTML:    html,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.resend.com/emails", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.http.Do(req)
	if err != nil {
		return fmt.Errorf("resend request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("resend status %d: %s", resp.StatusCode, b)
	}
	return nil
}
