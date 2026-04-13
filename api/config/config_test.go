package config_test

import (
	"os"
	"testing"

	"github.com/immortalvibes/api/config"
)

func setAllEnv(t *testing.T) {
	t.Helper()
	t.Setenv("PROXY_SECRET", "test-secret")
	t.Setenv("STRIPE_SECRET_KEY", "sk_test_123")
	t.Setenv("STRIPE_WEBHOOK_SECRET", "whsec_123")
	t.Setenv("ADMIN_SECRET", "admin-secret")
	t.Setenv("R2_PUBLIC_URL", "https://r2.example.com")
	t.Setenv("CF_ACCOUNT_ID", "acct123")
	t.Setenv("CF_KV_CARTS_ID", "ns123")
	t.Setenv("CF_API_TOKEN", "tok123")
	t.Setenv("DATABASE_URL", "postgres://localhost/test")
	t.Setenv("RESEND_API_KEY", "re_123")
	t.Setenv("SHIPPO_API_KEY", "shippo_test")
	t.Setenv("SHIPPO_FROM_NAME", "Test Sender")
	t.Setenv("SHIPPO_FROM_STREET1", "123 Test St")
	t.Setenv("SHIPPO_FROM_CITY", "Testville")
	t.Setenv("SHIPPO_FROM_STATE", "CA")
	t.Setenv("SHIPPO_FROM_ZIP", "90210")
	t.Setenv("OWNER_EMAIL", "owner@example.com")
}

func TestLoad_defaults(t *testing.T) {
	setAllEnv(t)
	os.Unsetenv("PORT")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port != "8080" {
		t.Errorf("expected default port 8080, got %q", cfg.Port)
	}
}

func TestLoad_missingRequired(t *testing.T) {
	// Don't set any env vars — all required fields missing.
	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error for missing required env vars, got nil")
	}
}

func TestLoad_allSet(t *testing.T) {
	setAllEnv(t)
	t.Setenv("PORT", "9090")

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
