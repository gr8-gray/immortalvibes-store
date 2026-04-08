package middleware

import (
	"crypto/subtle"
	"net/http"
)

// ProxyAuth validates the X-Proxy-Secret header injected by the CF Worker.
// /health is exempt so Fly.io health checks work without the secret.
func ProxyAuth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/health" {
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
