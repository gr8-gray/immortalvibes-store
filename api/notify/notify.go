// Package notify pushes order alerts to an ntfy topic so the operator gets a
// phone push on every order — a backstop for the transactional emails, which
// can fail silently. A blank URL disables notifications (no-op), so local and
// test runs need no ntfy configured.
package notify

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"
)

// Notifier posts messages to an ntfy topic URL, e.g. https://ntfy.sh/<topic>.
type Notifier struct {
	url  string
	http *http.Client
}

// New returns a Notifier for the given ntfy topic URL. An empty url yields a
// Notifier whose sends are no-ops.
func New(url string) *Notifier {
	return &Notifier{url: url, http: &http.Client{Timeout: 5 * time.Second}}
}

// push sends one ntfy message. Best-effort: the error is returned for logging
// but callers must never fail order processing on a notify error.
func (n *Notifier) push(ctx context.Context, title, body, priority, tags string) error {
	if n == nil || n.url == "" {
		return nil // notifications disabled
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, n.url, bytes.NewReader([]byte(body)))
	if err != nil {
		return err
	}
	// ntfy reads metadata from headers. Title/Tags must stay ASCII (HTTP
	// headers are Latin-1) — emoji come from the named Tags shortcodes.
	req.Header.Set("Title", title)
	req.Header.Set("Priority", priority)
	req.Header.Set("Tags", tags)
	resp, err := n.http.Do(req)
	if err != nil {
		return fmt.Errorf("ntfy request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("ntfy status %d", resp.StatusCode)
	}
	return nil
}

// NotifyOrderSuccess sends a quiet, low-priority "order came in and shipped"
// ping. Owner silence after this means everything worked.
func (n *Notifier) NotifyOrderSuccess(ctx context.Context, orderID, total, shipTo, items, tracking, carrier string) error {
	title := "IV order " + total
	body := fmt.Sprintf("Order %s\n%s\nShip to: %s\n%s %s", orderID, items, shipTo, carrier, tracking)
	return n.push(ctx, title, body, "low", "white_check_mark,package")
}

// NotifyOrderFailure sends an urgent alert that a step in fulfillment failed —
// the "expect something wrong, get to a terminal" signal.
func (n *Notifier) NotifyOrderFailure(ctx context.Context, orderID, email, stage, errMsg string) error {
	title := "IV ORDER FAILED: " + stage
	body := fmt.Sprintf("Order %s\nCustomer: %s\nStage: %s\n%s", orderID, email, stage, errMsg)
	return n.push(ctx, title, body, "urgent", "rotating_light,warning")
}
