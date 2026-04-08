package middleware

import (
	"net/http"
)

// AdminAuth returns a middleware that requires the X-Admin-Secret header
// to match the provided secret. Responds 401 if missing, 403 if wrong.
func AdminAuth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			provided := r.Header.Get("X-Admin-Secret")
			if provided == "" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			if provided != secret {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
