package email

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

// SendShippingLabel emails the owner a label-ready notification with PDF link.
func (s *Sender) SendShippingLabel(ctx context.Context, ownerEmail, orderID, labelURL, trackingNum, carrier string) error {
	subject := fmt.Sprintf("[IV Order] Label ready — %s", orderID)
	html := fmt.Sprintf(`
		<h2>Label Ready</h2>
		<p><strong>Order:</strong> %s</p>
		<p><strong>Carrier:</strong> %s</p>
		<p><strong>Tracking:</strong> %s</p>
		<p><a href="%s" style="font-weight:bold">Download Label (PDF)</a></p>
		<p style="color:#888;font-size:0.85em">Please ship within 2 business days.</p>
	`, orderID, carrier, trackingNum, labelURL)
	return s.send(ctx, ownerEmail, subject, html)
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

// SendShippingFailure alerts the owner that Shippo label creation failed.
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
		<p><a href="https://app.goshippo.com" style="font-weight:bold">Create label at goshippo.com</a></p>
	`, orderID, customerEmail, shippingAddr, errMsg)
	return s.send(ctx, ownerEmail, subject, html)
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
