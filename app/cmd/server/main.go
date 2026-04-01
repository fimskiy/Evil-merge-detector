package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/fimskiy/evil-merge-detector/app/internal/badge"
	"github.com/fimskiy/evil-merge-detector/app/internal/billing"
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

	if cfg.StripeSecretKey != "" && db != nil {
		bh := billing.New(cfg.StripeSecretKey, cfg.StripePriceMonthly, cfg.StripePriceYearly, cfg.StripeWebhookSecret, cfg.SessionSecret, db)
		mux.HandleFunc("/billing/checkout", bh.Checkout)
		mux.HandleFunc("/billing/portal", bh.Portal)
		mux.HandleFunc("/billing/webhook", bh.Webhook)
	}
	// Pass nil interface (not nil *store.Store) when DB is unavailable.
	var badgeStore badge.ScanStore
	if db != nil {
		badgeStore = db
	}
	mux.Handle("/badge/", badge.New(badgeStore))
	mux.HandleFunc("/og-image.png", landing.OGImageHandler)
	mux.HandleFunc("/privacy", landing.PrivacyHandler)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/auth/github", oauthHandler.Login)
	mux.HandleFunc("/auth/callback", oauthHandler.Callback)
	mux.HandleFunc("/auth/logout", oauthHandler.Logout)
	mux.Handle("/dashboard", dashboard.New(cfg.SessionSecret, db, cfg.StripeSecretKey != ""))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			landing.Handler(w, r)
			return
		}
		http.NotFound(w, r)
	})

	rl := newRateLimiter(100, time.Minute)
	go rl.cleanup(5 * time.Minute)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      secureHeaders(rl.middleware(mux)),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		log.Printf("starting server on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	<-stop
	log.Println("shutting down...")
	shutCtx, shutCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutCancel()
	if err := srv.Shutdown(shutCtx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}

// rateLimiter is a fixed-window per-IP rate limiter with no external dependencies.
type rateLimiter struct {
	mu      sync.Mutex
	clients map[string]*bucket
	max     int
	window  time.Duration
}

type bucket struct {
	count   int
	resetAt time.Time
}

func newRateLimiter(max int, window time.Duration) *rateLimiter {
	return &rateLimiter{clients: make(map[string]*bucket), max: max, window: window}
}

func (rl *rateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := time.Now()
	b, ok := rl.clients[ip]
	if !ok || now.After(b.resetAt) {
		rl.clients[ip] = &bucket{count: 1, resetAt: now.Add(rl.window)}
		return true
	}
	if b.count >= rl.max {
		return false
	}
	b.count++
	return true
}

func (rl *rateLimiter) cleanup(interval time.Duration) {
	t := time.NewTicker(interval)
	defer t.Stop()
	for range t.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, b := range rl.clients {
			if now.After(b.resetAt) {
				delete(rl.clients, ip)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *rateLimiter) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Prefer CF-Connecting-IP when behind Cloudflare.
		ip := r.Header.Get("CF-Connecting-IP")
		if ip == "" {
			ip, _, _ = net.SplitHostPort(r.RemoteAddr)
			if ip == "" {
				ip = r.RemoteAddr
			}
		}
		if !rl.allow(ip) {
			http.Error(w, "too many requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func secureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
		w.Header().Set("Content-Security-Policy",
			"default-src 'self'; "+
				"script-src 'self' 'unsafe-inline' https://www.googletagmanager.com; "+
				"style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; "+
				"font-src https://fonts.gstatic.com; "+
				"img-src 'self' data: https://evilmerge.dev; "+
				"connect-src 'self' https://www.google-analytics.com https://region1.google-analytics.com https://stats.g.doubleclick.net; "+
				"frame-ancestors 'none'")
		next.ServeHTTP(w, r)
	})
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
	var wg sync.WaitGroup

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
			wg.Add(1)
			go func(j worker.HistoryJob) {
				defer wg.Done()
				defer func() { <-sem }()
				worker.ScanHistory(j)
			}(job)
		}
	}
	wg.Wait()
}
