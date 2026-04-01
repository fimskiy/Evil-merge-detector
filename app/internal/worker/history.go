package worker

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/evilmerge-dev/evil-merge-detector/app/internal/ghclient"
	"github.com/evilmerge-dev/evil-merge-detector/app/internal/notifier"
	"github.com/evilmerge-dev/evil-merge-detector/app/internal/store"
	"github.com/evilmerge-dev/evil-merge-detector/internal/models"
	"github.com/evilmerge-dev/evil-merge-detector/internal/scanner"
)

type HistoryJob struct {
	Owner          string
	Repo           string
	DefaultBranch  string
	CloneURL       string
	InstallationID int64
	AppID          int64
	PrivateKey     []byte
	DB             *store.Store
	Notifier       *notifier.Notifier
	// cloneFn overrides ghclient.Clone in tests.
	cloneFn func(ctx context.Context, appID, installationID int64, privateKey []byte, repoURL, branch, destDir string) error
}

func ScanHistory(job HistoryJob) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	log.Printf("history scan: %s/%s (installation %d)", job.Owner, job.Repo, job.InstallationID)

	tmpDir, err := os.MkdirTemp("", "evilmerge-history-*")
	if err != nil {
		log.Printf("history scan mktemp %s/%s: %v", job.Owner, job.Repo, err)
		return
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			log.Printf("cleanup %s: %v", tmpDir, err)
		}
	}()

	branch := job.DefaultBranch
	if branch == "" {
		branch = "main"
	}

	cloneFn := job.cloneFn
	if cloneFn == nil {
		cloneFn = ghclient.Clone
	}
	if err := cloneFn(ctx, job.AppID, job.InstallationID, job.PrivateKey, job.CloneURL, branch, tmpDir); err != nil {
		log.Printf("history scan clone %s/%s: %v", job.Owner, job.Repo, err)
		return
	}

	s := scanner.New()
	result, err := s.Scan(ctx, models.ScanOptions{
		RepoPath: tmpDir,
		Branch:   branch,
	})
	if err != nil {
		log.Printf("history scan %s/%s: %v", job.Owner, job.Repo, err)
		return
	}

	log.Printf("history scan %s/%s: %d/%d evil merges", job.Owner, job.Repo, result.EvilMerges, result.TotalMerges)

	if job.DB != nil {
		if err := job.DB.MarkFullScanned(ctx, job.InstallationID); err != nil {
			log.Printf("history scan mark scanned %s/%s: %v", job.Owner, job.Repo, err)
		}
	}

	if job.DB != nil {
		rec := store.ScanRecord{
			InstallationID: job.InstallationID,
			Owner:          job.Owner,
			Repo:           job.Repo,
			HeadSHA:        "history",
			EvilMerges:     result.EvilMerges,
			TotalMerges:    result.TotalMerges,
			DurationMs:     result.ScanDuration.Milliseconds(),
		}
		if err := job.DB.SaveScan(ctx, rec); err != nil {
			log.Printf("history scan save %s/%s: %v", job.Owner, job.Repo, err)
		}
	}

	if result.EvilMerges > 0 && job.Notifier != nil && job.Notifier.Enabled() {
		ev := notifier.EvilMergeEvent{
			Owner:      job.Owner,
			Repo:       job.Repo,
			HeadSHA:    "history",
			EvilMerges: result.EvilMerges,
			RepoURL:    "https://github.com/" + job.Owner + "/" + job.Repo,
		}
		for _, r := range result.Reports {
			for _, ec := range r.EvilChanges {
				ev.Findings = append(ev.Findings, notifier.Finding{
					File:     ec.FilePath,
					Severity: ec.Severity.String(),
					Detail:   ec.Detail,
				})
			}
		}
		if err := job.Notifier.Notify(ctx, ev); err != nil {
			log.Printf("history scan notify %s/%s: %v", job.Owner, job.Repo, err)
		}
	}
}
