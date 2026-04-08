package main

import (
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/immortalvibes/api/config"
	"github.com/immortalvibes/api/handlers"
	"github.com/immortalvibes/api/middleware"
)

func newRouter(cfg *config.Config) *chi.Mux {
	r := chi.NewRouter()
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(middleware.CORS("https://theimmortalvibes.com"))
	r.Use(middleware.ProxyAuth(cfg.ProxySecret))
	r.Get("/health", handlers.Health)
	return r
}
