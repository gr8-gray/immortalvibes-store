package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/immortalvibes/api/middleware"
)

func okHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestProxyAuth_missingHeader(t *testing.T) {
	handler := middleware.ProxyAuth("correct-secret")(http.HandlerFunc(okHandler))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestProxyAuth_wrongSecret(t *testing.T) {
	handler := middleware.ProxyAuth("correct-secret")(http.HandlerFunc(okHandler))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Proxy-Secret", "wrong-secret")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestProxyAuth_correctSecret(t *testing.T) {
	handler := middleware.ProxyAuth("correct-secret")(http.HandlerFunc(okHandler))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Proxy-Secret", "correct-secret")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestProxyAuth_healthBypassesAuth(t *testing.T) {
	handler := middleware.ProxyAuth("correct-secret")(http.HandlerFunc(okHandler))
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	// no secret header
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("/health should bypass auth, got %d", w.Code)
	}
}
