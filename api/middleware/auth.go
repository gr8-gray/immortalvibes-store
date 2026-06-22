package middleware

import (
	"crypto/subtle"
	"net/http"
)

// ProxyAuth validates the X-Proxy-Secret header injected by the CF Worker.
// Exempt paths bypass the secret: /health (Fly.io health checks) and the Stripe
// webhook (Stripe calls Go directly, not through the CF Worker, so it can't carry
// the proxy secret). The webhook MUST be exempted here: ProxyAuth is mounted
// globally via r.Use(), so a route-level r.With(SkipProxyAuth) cannot remove it —
// without this exemption every Stripe webhook 403s and no order ever ships
// (incident 2026-06-21). The webhook is independently authenticated by Stripe's
// signature (STRIPE_WEBHOOK_SECRET) inside the handler.
func ProxyAuth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/health" || r.URL.Path == "/api/webhooks/stripe" {
				next.ServeHTTP(w, r)
				return
			}

			got := r.Header.Get("X-Proxy-Secret")
			if subtle.ConstantTimeCompare([]byte(got), []byte(secret)) != 1 {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// SkipProxyAuth is a middleware that bypasses the proxy secret check.
// Use only for Stripe webhooks — Stripe calls Go directly, not through CF Worker.
func SkipProxyAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}
