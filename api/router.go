package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/immortalvibes/api/config"
	"github.com/immortalvibes/api/email"
	"github.com/immortalvibes/api/handlers"
	apimiddleware "github.com/immortalvibes/api/middleware"
	"github.com/immortalvibes/api/shippo"
	"github.com/immortalvibes/api/store"
)

func newRouter(cfg *config.Config, db *store.DB, kv *store.KVClient) http.Handler {
	r := chi.NewRouter()
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(apimiddleware.CORS("https://theimmortalvibes.com"))
	r.Use(apimiddleware.ProxyAuth(cfg.ProxySecret))

	// Health
	r.Get("/health", handlers.Health)

	// Products
	productSvc := handlers.NewStripeProductService(cfg.StripeSecretKey, cfg.R2PublicURL, db)
	productsHandler := handlers.NewProductsHandler(productSvc)
	r.Get("/api/products", productsHandler.ListProducts)
	r.Get("/api/products/{id}", productsHandler.GetProduct)

	// Cart
	cartHandler := handlers.NewCartHandler(kv)
	r.Get("/api/cart", cartHandler.GetCurrentCart)
	r.Get("/api/cart/{token}", cartHandler.GetCart)
	r.Post("/api/cart", cartHandler.AddToCart)
	r.Put("/api/cart/{token}", cartHandler.UpdateCart)

	// Promo code validation
	promoHandler := handlers.NewPromoHandler()
	r.Post("/api/promo/validate", promoHandler.Validate)

	// Shared Shippo client
	shippoFromAddr := shippo.Address{
		Name:    cfg.FromName,
		Street1: cfg.FromStreet1,
		City:    cfg.FromCity,
		State:   cfg.FromState,
		Zip:     cfg.FromZip,
		Country: cfg.FromCountry,
		Email:   "orders@theimmortalvibes.com", // Shippo requires a non-empty from-email to buy labels.
	}
	shippoClient := shippo.NewClient(cfg.ShippoAPIKey, shippoFromAddr)

	// Shipping estimate
	shippingHandler := handlers.NewShippingHandler(shippoClient, shippoFromAddr)
	r.Post("/api/shipping/estimate", shippingHandler.Estimate)

	// Checkout
	checkoutHandler := handlers.NewCheckoutHandler(cfg.StripeSecretKey, kv, db)
	r.Post("/api/checkout", checkoutHandler.Checkout)

	// Orders
	ordersHandler := handlers.NewOrdersHandler(db)
	r.Get("/api/order/{id}", ordersHandler.GetOrder)

	// Stripe webhook — Stripe calls Go directly (no CF Worker, no proxy secret).
	// Exempted inside ProxyAuth by path (global r.Use can't be undone per-route);
	// authenticated by Stripe signature in the handler.
	emailSender := email.NewSender(cfg.ResendAPIKey, "orders@theimmortalvibes.com")
	webhookHandler := handlers.NewWebhookHandler(cfg.StripeWebhookSecret, kv, db, db, emailSender, shippoClient, cfg.OwnerEmail)
	r.Post("/api/webhooks/stripe", webhookHandler.HandleWebhook)

	// Subscribe — email capture for discount code
	subscribeHandler := handlers.NewSubscribeHandler(db, emailSender)
	r.Post("/api/subscribe", subscribeHandler.Subscribe)

	// Admin (behind AdminAuth)
	adminHandler := handlers.NewAdminHandler(db)
	r.With(apimiddleware.AdminAuth(cfg.AdminSecret)).Put("/api/admin/products/{id}/stock", adminHandler.SetStock)

	return r
}
