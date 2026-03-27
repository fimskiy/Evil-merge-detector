package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/fimskiy/evil-merge-detector/app/internal/badge"
	"github.com/fimskiy/evil-merge-detector/app/internal/config"
	"github.com/fimskiy/evil-merge-detector/app/internal/dashboard"
	"github.com/fimskiy/evil-merge-detector/app/internal/ghclient"
	"github.com/fimskiy/evil-merge-detector/app/internal/landing"
	"github.com/fimskiy/evil-merge-detector/app/internal/notifier"
	"github.com/fimskiy/evil-merge-detector/app/internal/oauth"
	"github.com/fimskiy/evil-merge-detector/app/internal/store"
	"github.com/fimskiy/evil-merge-detector/app/internal/webhook"
	"github.com/fimskiy/evil-merge-detector/app/internal/worker"
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

	ntf := notifier.New(cfg.NotificationWebhookURL, cfg.SlackWebhookURL)

	if db != nil {
		go runHistoryScheduler(cfg, db, ntf)
	}

	oauthHandler := oauth.New(cfg.OAuthClientID, cfg.OAuthClientSecret, cfg.SessionSecret, db)

	mux := http.NewServeMux()
	mux.Handle("/webhook", webhook.New(cfg, db, ntf))
	// Pass nil interface (not nil *store.Store) when DB is unavailable.
	var badgeStore badge.ScanStore
	if db != nil {
		badgeStore = db
	}
	mux.Handle("/badge/", badge.New(badgeStore))
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/auth/github", oauthHandler.Login)
	mux.HandleFunc("/auth/callback", oauthHandler.Callback)
	mux.HandleFunc("/auth/logout", oauthHandler.Logout)
	mux.Handle("/dashboard", dashboard.New(cfg.SessionSecret, db))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			landing.Handler(w, r)
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

func runHistoryScheduler(cfg *config.Config, db *store.Store, ntf *notifier.Notifier) {
	runPendingHistoryScans(cfg, db, ntf)

	ticker := time.NewTicker(6 * time.Hour)
	defer ticker.Stop()
	for range ticker.C {
		runPendingHistoryScans(cfg, db, ntf)
	}
}

func runPendingHistoryScans(cfg *config.Config, db *store.Store, ntf *notifier.Notifier) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	installations, err := db.ListPendingFullScans(ctx)
	if err != nil {
		log.Printf("scheduler: list pending: %v", err)
		return
	}
	if len(installations) == 0 {
		return
	}
	log.Printf("scheduler: found %d installation(s) pending full scan", len(installations))

	// Limit concurrent history scans to avoid exhausting resources.
	const maxConcurrent = 3
	sem := make(chan struct{}, maxConcurrent)

	for _, inst := range installations {
		repos, err := ghclient.ListRepos(ctx, cfg.AppID, inst.InstallationID, cfg.PrivateKey)
		if err != nil {
			log.Printf("scheduler: list repos for %d: %v", inst.InstallationID, err)
			continue
		}
		for _, repo := range repos {
			job := worker.HistoryJob{
				Owner:          repo.Owner,
				Repo:           repo.Name,
				DefaultBranch:  repo.DefaultBranch,
				CloneURL:       repo.CloneURL,
				InstallationID: inst.InstallationID,
				AppID:          cfg.AppID,
				PrivateKey:     cfg.PrivateKey,
				DB:             db,
				Notifier:       ntf,
			}
			sem <- struct{}{}
			go func(j worker.HistoryJob) {
				defer func() { <-sem }()
				worker.ScanHistory(j)
			}(job)
		}
	}
}
