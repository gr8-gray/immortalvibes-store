package config

import (
	"errors"
	"os"
)

type Config struct {
	Port                string
	ProxySecret         string
	StripeSecretKey     string
	StripeWebhookSecret string
	AdminSecret         string
}

func Load() (*Config, error) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	proxySecret := os.Getenv("PROXY_SECRET")
	if proxySecret == "" {
		return nil, errors.New("PROXY_SECRET is required")
	}

	return &Config{
		Port:                port,
		ProxySecret:         proxySecret,
		StripeSecretKey:     os.Getenv("STRIPE_SECRET_KEY"),
		StripeWebhookSecret: os.Getenv("STRIPE_WEBHOOK_SECRET"),
		AdminSecret:         os.Getenv("ADMIN_SECRET"),
	}, nil
}
