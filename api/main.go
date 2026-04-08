package main

import (
	"log"
	"net/http"

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

	log.Printf("listening on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, router); err != nil {
		log.Fatalf("server: %v", err)
	}
}
