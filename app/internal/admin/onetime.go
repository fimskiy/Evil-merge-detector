package admin

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/google/go-github/v84/github"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/fimskiy/evil-merge-detector/app/internal/config"
	"github.com/fimskiy/evil-merge-detector/app/internal/ghclient"
	"github.com/fimskiy/evil-merge-detector/app/internal/store"
)

// oneTimeToken gates a temporary maintenance endpoint. Delete this file
// (and its route in main.go) once it's been run — see the Sep 2026 Fly
// billing investigation for why: GitHub auto-creates a check_suite on every
// push for apps with Checks:write, regardless of webhook event subscriptions,
// which was keeping the Fly machines from scaling to zero.
const oneTimeToken = "5d1de85715a0193a68ae93c02fdfc1af"

// DisableCheckSuiteAutotrigger disables GitHub's automatic check_suite
// creation for every repo across every installation.
func DisableCheckSuiteAutotrigger(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("token") != oneTimeToken {
			http.NotFound(w, r)
			return
		}
		ctx := r.Context()

		pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer pool.Close()

		rows, err := pool.Query(ctx, `SELECT installation_id FROM installations`)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		var installationIDs []int64
		for rows.Next() {
			var id int64
			if err := rows.Scan(&id); err != nil {
				rows.Close()
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			installationIDs = append(installationIDs, id)
		}
		rows.Close()

		for _, instID := range installationIDs {
			client, err := ghclient.ForInstallation(cfg.AppID, instID, cfg.PrivateKey)
			if err != nil {
				fmt.Fprintf(w, "installation %d: client: %v\n", instID, err)
				continue
			}

			repos, err := ghclient.ListRepos(ctx, cfg.AppID, instID, cfg.PrivateKey)
			if err != nil {
				fmt.Fprintf(w, "installation %d: listing repos: %v\n", instID, err)
				continue
			}

			for _, repo := range repos {
				owner, name := repo.Owner, repo.Name
				_, _, err := client.Checks.SetCheckSuitePreferences(ctx, owner, name, github.CheckSuitePreferenceOptions{
					AutoTriggerChecks: []*github.AutoTriggerCheck{
						{AppID: github.Ptr(cfg.AppID), Setting: github.Ptr(false)},
					},
				})
				if err != nil {
					fmt.Fprintf(w, "%s/%s: %v\n", owner, name, err)
					continue
				}
				fmt.Fprintf(w, "%s/%s: auto_trigger_checks disabled\n", owner, name)
			}
		}
	}
}

// RepoScanStats reports which installation owns a repo and its scan volume,
// for investigating unexpectedly heavy scan traffic (?owner=x&repo=y).
func RepoScanStats(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("token") != oneTimeToken {
			http.NotFound(w, r)
			return
		}
		owner, repo := r.URL.Query().Get("owner"), r.URL.Query().Get("repo")
		if owner == "" || repo == "" {
			http.Error(w, "owner and repo query params required", http.StatusBadRequest)
			return
		}
		ctx := r.Context()

		pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer pool.Close()

		var installationID int64
		var accountLogin, accountType, plan string
		var installedAt time.Time
		err = pool.QueryRow(ctx, `
			SELECT i.installation_id, i.account_login, i.account_type, i.plan, i.installed_at
			FROM scans s JOIN installations i ON i.installation_id = s.installation_id
			WHERE s.owner = $1 AND s.repo = $2
			LIMIT 1
		`, owner, repo).Scan(&installationID, &accountLogin, &accountType, &plan, &installedAt)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "installation %d: account=%s (%s), plan=%s, installed_at=%s\n",
			installationID, accountLogin, accountType, plan, installedAt.Format(time.RFC3339))

		var total int
		var first, last time.Time
		if err := pool.QueryRow(ctx, `
			SELECT COUNT(*), MIN(scanned_at), MAX(scanned_at) FROM scans WHERE owner = $1 AND repo = $2
		`, owner, repo).Scan(&total, &first, &last); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "%s/%s: %d scans, first=%s, last=%s\n",
			owner, repo, total, first.Format(time.RFC3339), last.Format(time.RFC3339))

		var last24h, lastHour int
		if err := pool.QueryRow(ctx, `
			SELECT COUNT(*) FILTER (WHERE scanned_at >= NOW() - INTERVAL '24 hours'),
			       COUNT(*) FILTER (WHERE scanned_at >= NOW() - INTERVAL '1 hour')
			FROM scans WHERE owner = $1 AND repo = $2
		`, owner, repo).Scan(&last24h, &lastHour); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "%s/%s: %d scans in last 24h, %d in last hour\n", owner, repo, last24h, lastHour)
	}
}

// VerifyRepoScanClaim proves ClaimRepoScan can't be raced against the real
// production database: fires many concurrent claims for a throwaway repo and
// checks exactly one wins, then cleans up its own rows.
func VerifyRepoScanClaim(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("token") != oneTimeToken {
			http.NotFound(w, r)
			return
		}
		ctx := r.Context()

		db, err := store.New(ctx, cfg.DatabaseURL)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer db.Close()
		if err := db.Migrate(ctx); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		const owner, repo = "verify-claim-test", "repo"
		// Cleanup must outlive the request context - a disconnected client
		// shouldn't leave a stale claim row behind for the next run.
		defer func() {
			cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := db.DeleteRepoScanClaim(cleanupCtx, owner, repo); err != nil {
				log.Printf("verify-repo-scan-claim: cleanup: %v", err)
			}
		}()

		claimed, err := db.ClaimRepoScan(ctx, owner, repo, time.Minute)
		if err != nil {
			fmt.Fprintf(w, "FAIL: first claim errored: %v\n", err)
			return
		}
		fmt.Fprintf(w, "first claim: %v (want true)\n", claimed)

		claimed, err = db.ClaimRepoScan(ctx, owner, repo, time.Minute)
		if err != nil {
			fmt.Fprintf(w, "FAIL: second claim errored: %v\n", err)
			return
		}
		fmt.Fprintf(w, "second claim within window: %v (want false)\n", claimed)

		if err := db.DeleteRepoScanClaim(ctx, owner, repo); err != nil {
			fmt.Fprintf(w, "FAIL: cleanup before concurrency test: %v\n", err)
			return
		}

		const n = 50
		var wins int
		var wg sync.WaitGroup
		var mu sync.Mutex
		var firstErr error
		for range n {
			wg.Add(1)
			go func() {
				defer wg.Done()
				claimed, err := db.ClaimRepoScan(ctx, owner, repo, time.Minute)
				if err != nil {
					mu.Lock()
					if firstErr == nil {
						firstErr = err
					}
					mu.Unlock()
					return
				}
				if claimed {
					mu.Lock()
					wins++
					mu.Unlock()
				}
			}()
		}
		wg.Wait()

		if firstErr != nil {
			fmt.Fprintf(w, "FAIL: concurrent claim errored: %v\n", firstErr)
			return
		}
		fmt.Fprintf(w, "concurrent claims: %d/%d winners (want exactly 1)\n", wins, n)
		if wins == 1 {
			fmt.Fprintln(w, "PASS")
		} else {
			fmt.Fprintln(w, "FAIL")
		}
	}
}
