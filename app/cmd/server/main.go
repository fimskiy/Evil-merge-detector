package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/fimskiy/evil-merge-detector/app/internal/config"
	"github.com/fimskiy/evil-merge-detector/app/internal/dashboard"
	"github.com/fimskiy/evil-merge-detector/app/internal/oauth"
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

	oauthHandler := oauth.New(cfg.OAuthClientID, cfg.OAuthClientSecret, cfg.SessionSecret, db)

	mux := http.NewServeMux()
	mux.Handle("/webhook", webhook.New(cfg, db))
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/auth/github", oauthHandler.Login)
	mux.HandleFunc("/auth/callback", oauthHandler.Callback)
	mux.HandleFunc("/auth/logout", oauthHandler.Logout)
	mux.Handle("/dashboard", dashboard.New(cfg.SessionSecret, db))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/dashboard", http.StatusFound)
			return
		}
		http.NotFound(w, r)
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
