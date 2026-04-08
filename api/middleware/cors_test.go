package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/immortalvibes/api/middleware"
)

func TestCORS_setsHeaders(t *testing.T) {
	handler := middleware.CORS("https://theimmortalvibes.com")(http.HandlerFunc(okHandler))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://theimmortalvibes.com")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	got := w.Header().Get("Access-Control-Allow-Origin")
	if got != "https://theimmortalvibes.com" {
		t.Errorf("expected CORS origin header, got %q", got)
	}
}

func TestCORS_preflight(t *testing.T) {
	handler := middleware.CORS("https://theimmortalvibes.com")(http.HandlerFunc(okHandler))
	req := httptest.NewRequest(http.MethodOptions, "/api/products", nil)
	req.Header.Set("Origin", "https://theimmortalvibes.com")
	req.Header.Set("Access-Control-Request-Method", "GET")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204 for preflight, got %d", w.Code)
	}
}

func TestCORS_unknownOriginBlocked(t *testing.T) {
	handler := middleware.CORS("https://theimmortalvibes.com")(http.HandlerFunc(okHandler))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://evil.com")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	got := w.Header().Get("Access-Control-Allow-Origin")
	if got == "https://evil.com" {
		t.Error("should not allow unknown origin")
	}
}
