package worker

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/go-github/v84/github"

	"github.com/fimskiy/evil-merge-detector/app/internal/ghclient"
	"github.com/fimskiy/evil-merge-detector/app/internal/store"
	"github.com/fimskiy/evil-merge-detector/internal/models"
	"github.com/fimskiy/evil-merge-detector/internal/scanner"
)

type PRJob struct {
	Owner          string
	Repo           string
	CloneURL       string
	HeadSHA        string
	HeadRef        string
	PRNumber       int
	InstallationID int64
	AppID          int64
	PrivateKey     []byte
	DB             *store.Store
	Pro            bool
}

func ScanPR(job PRJob) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	client, err := ghclient.ForInstallation(job.AppID, job.InstallationID, job.PrivateKey)
	if err != nil {
		log.Printf("error creating GitHub client for %s/%s: %v", job.Owner, job.Repo, err)
		return
	}

	runID, err := ghclient.CreateCheckRun(ctx, client, job.Owner, job.Repo, job.HeadSHA)
	if err != nil {
		log.Printf("error creating check run for %s/%s: %v", job.Owner, job.Repo, err)
		return
	}

	if job.DB != nil && !job.Pro {
		count, err := job.DB.MonthlyScansCount(ctx, job.InstallationID)
		if err != nil {
			log.Printf("monthly scan count for %s/%s: %v", job.Owner, job.Repo, err)
		} else if count >= 50 {
			log.Printf("free plan limit reached for installation %d", job.InstallationID)
			limitErr := fmt.Errorf("monthly scan limit reached (50/month on Free plan). Upgrade to Pro at https://evil-merge-detector.fly.dev/#pricing")
			if err := ghclient.FailCheckRun(ctx, client, job.Owner, job.Repo, runID, limitErr); err != nil {
				log.Printf("error failing check run: %v", err)
			}
			return
		}
	}

	result, scanErr := runScan(ctx, job)

	if scanErr != nil {
		log.Printf("scan error for %s/%s: %v", job.Owner, job.Repo, scanErr)
		if err := ghclient.FailCheckRun(ctx, client, job.Owner, job.Repo, runID, scanErr); err != nil {
			log.Printf("error failing check run: %v", err)
		}
		return
	}

	if err := ghclient.UpdateCheckRun(ctx, client, job.Owner, job.Repo, runID, result); err != nil {
		log.Printf("error updating check run for %s/%s: %v", job.Owner, job.Repo, err)
	}

	if job.DB != nil {
		rec := store.ScanRecord{
			InstallationID: job.InstallationID,
			Owner:          job.Owner,
			Repo:           job.Repo,
			PRNumber:       job.PRNumber,
			HeadSHA:        job.HeadSHA,
			EvilMerges:     result.EvilMerges,
			TotalMerges:    result.TotalMerges,
			DurationMs:     result.ScanDuration.Milliseconds(),
		}
		if err := job.DB.SaveScan(ctx, rec); err != nil {
			log.Printf("saving scan record for %s/%s: %v", job.Owner, job.Repo, err)
		}
	}
}

func runScan(ctx context.Context, job PRJob) (*models.ScanResult, error) {
	tmpDir, err := os.MkdirTemp("", "evilmerge-*")
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			log.Printf("cleanup %s: %v", tmpDir, err)
		}
	}()

	if err := ghclient.Clone(ctx, job.AppID, job.InstallationID, job.PrivateKey, job.CloneURL, job.HeadRef, tmpDir); err != nil {
		return nil, err
	}

	s := scanner.New()
	return s.Scan(ctx, models.ScanOptions{
		RepoPath: tmpDir,
		Branch:   job.HeadRef,
	})
}

func NewGitHubClient(appID, installationID int64, privateKey []byte) (*github.Client, error) {
	return ghclient.ForInstallation(appID, installationID, privateKey)
}
