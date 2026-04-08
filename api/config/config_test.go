package config_test

import (
	"os"
	"testing"

	"github.com/immortalvibes/api/config"
)

func TestLoad_defaults(t *testing.T) {
	os.Unsetenv("PORT")
	os.Setenv("PROXY_SECRET", "test-secret")
	os.Unsetenv("STRIPE_SECRET_KEY")
	os.Unsetenv("STRIPE_WEBHOOK_SECRET")
	os.Unsetenv("ADMIN_SECRET")
	defer os.Unsetenv("PROXY_SECRET")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port != "8080" {
		t.Errorf("expected default port 8080, got %q", cfg.Port)
	}
}

func TestLoad_missingRequired(t *testing.T) {
	os.Setenv("PORT", "8080")
	os.Unsetenv("PROXY_SECRET")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error for missing PROXY_SECRET, got nil")
	}
}

func TestLoad_allSet(t *testing.T) {
	os.Setenv("PORT", "9090")
	os.Setenv("PROXY_SECRET", "test-secret")
	os.Setenv("STRIPE_SECRET_KEY", "sk_test_123")
	os.Setenv("STRIPE_WEBHOOK_SECRET", "whsec_123")
	os.Setenv("ADMIN_SECRET", "admin-secret")
	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("PROXY_SECRET")
		os.Unsetenv("STRIPE_SECRET_KEY")
		os.Unsetenv("STRIPE_WEBHOOK_SECRET")
		os.Unsetenv("ADMIN_SECRET")
	}()

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port != "9090" {
		t.Errorf("expected port 9090, got %q", cfg.Port)
	}
	if cfg.ProxySecret != "test-secret" {
		t.Errorf("expected proxy secret, got %q", cfg.ProxySecret)
	}
}
