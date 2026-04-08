package handlers_test

import (
	"net/http/httptest"
	"testing"

	"github.com/immortalvibes/api/handlers"
)

func TestDetectCurrency(t *testing.T) {
	cases := []struct {
		country  string
		expected string
	}{
		{"GB", "gbp"},
		{"US", "usd"},
		{"AU", "aud"},
		{"NZ", "aud"},  // NZ mapped to AUD (no NZD price in Stripe)
		{"DE", "eur"},
		{"FR", "eur"},
		{"IT", "eur"},
		{"ES", "eur"},
		{"XX", "usd"},  // unknown → default USD
		{"", "usd"},    // missing header → default USD
	}

	for _, tc := range cases {
		req := httptest.NewRequest("GET", "/", nil)
		if tc.country != "" {
			req.Header.Set("CF-IPCountry", tc.country)
		}
		got := handlers.DetectCurrency(req)
		if got != tc.expected {
			t.Errorf("country %q: got %q, want %q", tc.country, got, tc.expected)
		}
	}
}
