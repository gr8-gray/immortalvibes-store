package main

import (
	"log"
	"net/http"
	"time"

	"github.com/immortalvibes/api/config"
	"github.com/immortalvibes/api/store"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	db, err := store.Open(cfg.DBUrl)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer db.Close()

	kv := store.NewKVClient(
		"https://api.cloudflare.com",
		cfg.CFAccountID,
		cfg.CFKVCartsID,
		cfg.CFAPIToken,
	)

	router := newRouter(cfg, db, kv)

	// WriteTimeout must exceed the slowest handler: the payment webhook can
	// create a shipment and buy a label (shippo.purchaseTimeout each).
	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      150 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
	log.Printf("listening on :%s", cfg.Port)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("server: %v", err)
	}
}
