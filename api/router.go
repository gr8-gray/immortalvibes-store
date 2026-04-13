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
	r.Get("/api/cart/{token}", cartHandler.GetCart)
	r.Post("/api/cart", cartHandler.AddToCart)
	r.Put("/api/cart/{token}", cartHandler.UpdateCart)

	// Checkout
	checkoutHandler := handlers.NewCheckoutHandler(cfg.StripeSecretKey, kv, db)
	r.Post("/api/checkout", checkoutHandler.Checkout)

	// Orders
	ordersHandler := handlers.NewOrdersHandler(db)
	r.Get("/api/order/{id}", ordersHandler.GetOrder)

	// Stripe webhook (not behind ProxyAuth — Stripe calls this directly)
	emailSender := email.NewSender(cfg.ResendAPIKey, "orders@immortalvibes.co.uk")
	shippoClient := shippo.NewClient(cfg.ShippoAPIKey, shippo.Address{
		Name:    cfg.ShippoFromName,
		Street1: cfg.ShippoFromStreet1,
		City:    cfg.ShippoFromCity,
		State:   cfg.ShippoFromState,
		Zip:     cfg.ShippoFromZip,
		Country: cfg.ShippoFromCountry,
	})
	webhookHandler := handlers.NewWebhookHandler(cfg.StripeWebhookSecret, kv, db, db, emailSender, shippoClient, cfg.OwnerEmail)
	r.With(apimiddleware.SkipProxyAuth).Post("/api/webhooks/stripe", webhookHandler.HandleWebhook)

	// Admin (behind AdminAuth)
	adminHandler := handlers.NewAdminHandler(db)
	r.With(apimiddleware.AdminAuth(cfg.AdminSecret)).Put("/api/admin/products/{id}/stock", adminHandler.SetStock)

	return r
}
