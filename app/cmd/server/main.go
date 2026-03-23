package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/fimskiy/evil-merge-detector/app/internal/config"
	"github.com/fimskiy/evil-merge-detector/app/internal/store"
	"github.com/fimskiy/evil-merge-detector/app/internal/webhook"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx := context.Background()

	var db *store.Store
	if cfg.DatabaseURL != "" {
		db, err = store.New(ctx, cfg.DatabaseURL)
		if err != nil {
			log.Fatalf("database: %v", err)
		}
		defer db.Close()

		if err := db.Migrate(ctx); err != nil {
			log.Fatalf("migration: %v", err)
		}
		log.Println("database connected")
	} else {
		log.Println("DATABASE_URL not set, running without database")
	}

	mux := http.NewServeMux()
	mux.Handle("/webhook", webhook.New(cfg, db))
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	log.Printf("starting server on :%s", cfg.Port)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
