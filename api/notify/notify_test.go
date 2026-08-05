package notify

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// A blank URL must make sends no-ops (local/test runs need no ntfy).
func TestNotifierDisabledIsNoop(t *testing.T) {
	n := New("")
	if err := n.NotifyOrderSuccess(context.Background(), "ord-1", "$30.00 USD", "Testville, CA US", "1x Tee", "TRACK1", "USPS"); err != nil {
		t.Fatalf("disabled success = %v, want nil", err)
	}
	if err := n.NotifyOrderFailure(context.Background(), "ord-1", "buyer@example.com", "shipping", "boom"); err != nil {
		t.Fatalf("disabled failure = %v, want nil", err)
	}
}

func TestNotifierSuccessPost(t *testing.T) {
	var gotTitle, gotPriority, gotTags, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotTitle = r.Header.Get("Title")
		gotPriority = r.Header.Get("Priority")
		gotTags = r.Header.Get("Tags")
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	n := New(srv.URL)
	if err := n.NotifyOrderSuccess(context.Background(), "ord-42", "$30.00 USD", "Testville, CA US", "2x Tee (L)", "TRACK9", "USPS"); err != nil {
		t.Fatalf("success = %v", err)
	}
	if gotPriority != "low" {
		t.Errorf("priority=%q want low", gotPriority)
	}
	if !strings.Contains(gotTitle, "$30.00") {
		t.Errorf("title=%q want it to include the amount", gotTitle)
	}
	if !strings.Contains(gotBody, "ord-42") || !strings.Contains(gotBody, "TRACK9") {
		t.Errorf("body=%q missing order id / tracking", gotBody)
	}
	if gotTags == "" {
		t.Errorf("tags header empty")
	}
}

func TestNotifierFailureIsUrgent(t *testing.T) {
	var gotPriority, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPriority = r.Header.Get("Priority")
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	n := New(srv.URL)
	if err := n.NotifyOrderFailure(context.Background(), "ord-99", "buyer@example.com", "shipping", "shippo down"); err != nil {
		t.Fatalf("failure = %v", err)
	}
	if gotPriority != "urgent" {
		t.Errorf("priority=%q want urgent", gotPriority)
	}
	if !strings.Contains(gotBody, "shipping") || !strings.Contains(gotBody, "shippo down") {
		t.Errorf("body=%q missing stage/error", gotBody)
	}
}
