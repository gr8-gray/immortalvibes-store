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
