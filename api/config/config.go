package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port                string
	ProxySecret         string
	StripeSecretKey     string
	StripeWebhookSecret string
	AdminSecret         string
	R2PublicURL         string
	CFAccountID         string
	CFKVCartsID         string
	CFAPIToken          string
	DBUrl               string
	ResendAPIKey        string
}

func Load() (*Config, error) {
	c := &Config{
		Port:                getEnv("PORT", "8080"),
		ProxySecret:         os.Getenv("PROXY_SECRET"),
		StripeSecretKey:     os.Getenv("STRIPE_SECRET_KEY"),
		StripeWebhookSecret: os.Getenv("STRIPE_WEBHOOK_SECRET"),
		AdminSecret:         os.Getenv("ADMIN_SECRET"),
		R2PublicURL:         os.Getenv("R2_PUBLIC_URL"),
		CFAccountID:         os.Getenv("CF_ACCOUNT_ID"),
		CFKVCartsID:         os.Getenv("CF_KV_CARTS_ID"),
		CFAPIToken:          os.Getenv("CF_API_TOKEN"),
		DBUrl:               os.Getenv("DATABASE_URL"),
		ResendAPIKey:        os.Getenv("RESEND_API_KEY"),
	}

	var missing []string
	if c.ProxySecret == "" {
		missing = append(missing, "PROXY_SECRET")
	}
	if c.StripeSecretKey == "" {
		missing = append(missing, "STRIPE_SECRET_KEY")
	}
	if c.StripeWebhookSecret == "" {
		missing = append(missing, "STRIPE_WEBHOOK_SECRET")
	}
	if c.AdminSecret == "" {
		missing = append(missing, "ADMIN_SECRET")
	}
	if c.R2PublicURL == "" {
		missing = append(missing, "R2_PUBLIC_URL")
	}
	if c.CFAccountID == "" {
		missing = append(missing, "CF_ACCOUNT_ID")
	}
	if c.CFKVCartsID == "" {
		missing = append(missing, "CF_KV_CARTS_ID")
	}
	if c.CFAPIToken == "" {
		missing = append(missing, "CF_API_TOKEN")
	}
	if c.DBUrl == "" {
		missing = append(missing, "DATABASE_URL")
	}
	if c.ResendAPIKey == "" {
		missing = append(missing, "RESEND_API_KEY")
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("missing required env vars: %v", missing)
	}
	return c, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
