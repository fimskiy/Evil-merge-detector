package admin

import (
	"fmt"
	"net/http"

	"github.com/google/go-github/v84/github"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/fimskiy/evil-merge-detector/app/internal/config"
	"github.com/fimskiy/evil-merge-detector/app/internal/ghclient"
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
