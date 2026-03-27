package webhook

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/google/go-github/v84/github"

	"github.com/fimskiy/evil-merge-detector/app/internal/config"
	"github.com/fimskiy/evil-merge-detector/app/internal/ghclient"
	"github.com/fimskiy/evil-merge-detector/app/internal/notifier"
	"github.com/fimskiy/evil-merge-detector/app/internal/store"
	"github.com/fimskiy/evil-merge-detector/app/internal/worker"
)

type Handler struct {
	cfg      *config.Config
	db       *store.Store
	notifier *notifier.Notifier
}

func New(cfg *config.Config, db *store.Store, ntf *notifier.Notifier) http.Handler {
	return &Handler{cfg: cfg, db: db, notifier: ntf}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	payload, err := github.ValidatePayload(r, h.cfg.WebhookSecret)
	if err != nil {
		http.Error(w, "invalid signature", http.StatusUnauthorized)
		return
	}

	event, err := github.ParseWebHook(github.WebHookType(r), payload)
	if err != nil {
		http.Error(w, "cannot parse payload", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)

	ctx := r.Context()
	switch e := event.(type) {
	case *github.PullRequestEvent:
		h.handlePR(r, e)
	case *github.InstallationEvent:
		h.handleInstallation(ctx, e)
	case *github.MarketplacePurchaseEvent:
		h.handleMarketplace(ctx, e)
	}
}

func (h *Handler) handlePR(r *http.Request, e *github.PullRequestEvent) {
	action := e.GetAction()
	if action != "opened" && action != "synchronize" && action != "reopened" {
		return
	}

	pr := e.GetPullRequest()
	installationID := e.GetInstallation().GetID()

	var pro bool
	if h.db != nil {
		if inst, err := h.db.GetInstallation(r.Context(), installationID); err == nil {
			pro = inst.Plan == "pro"
		}
	}

	job := worker.PRJob{
		Owner:          e.GetRepo().GetOwner().GetLogin(),
		Repo:           e.GetRepo().GetName(),
		CloneURL:       e.GetRepo().GetCloneURL(),
		HeadSHA:        pr.GetHead().GetSHA(),
		HeadRef:        pr.GetHead().GetRef(),
		PRNumber:       pr.GetNumber(),
		InstallationID: installationID,
		AppID:          h.cfg.AppID,
		PrivateKey:     h.cfg.PrivateKey,
		DB:             h.db,
		Notifier:       h.notifier,
		Pro:            pro,
	}

	log.Printf("scanning PR #%d in %s/%s (%.7s)", pr.GetNumber(), job.Owner, job.Repo, job.HeadSHA)
	go worker.ScanPR(job)
}

func (h *Handler) handleInstallation(ctx context.Context, e *github.InstallationEvent) {
	if h.db == nil {
		return
	}

	inst := e.GetInstallation()
	account := inst.GetAccount()

	switch e.GetAction() {
	case "created":
		if err := h.db.UpsertInstallation(ctx, store.Installation{
			InstallationID: inst.GetID(),
			AccountLogin:   account.GetLogin(),
			AccountType:    account.GetType(),
			Plan:           "free",
		}); err != nil {
			log.Printf("upsert installation %d: %v", inst.GetID(), err)
		}
		go h.triggerHistoryScan(inst.GetID())
	case "deleted":
		if err := h.db.DeleteInstallation(ctx, inst.GetID()); err != nil {
			log.Printf("delete installation %d: %v", inst.GetID(), err)
		}
	}
}

func (h *Handler) triggerHistoryScan(installationID int64) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	repos, err := ghclient.ListRepos(ctx, h.cfg.AppID, installationID, h.cfg.PrivateKey)
	if err != nil {
		log.Printf("list repos for install %d: %v", installationID, err)
		return
	}
	const maxConcurrent = 3
	sem := make(chan struct{}, maxConcurrent)

	for _, repo := range repos {
		job := worker.HistoryJob{
			Owner:          repo.Owner,
			Repo:           repo.Name,
			DefaultBranch:  repo.DefaultBranch,
			CloneURL:       repo.CloneURL,
			InstallationID: installationID,
			AppID:          h.cfg.AppID,
			PrivateKey:     h.cfg.PrivateKey,
			DB:             h.db,
			Notifier:       h.notifier,
		}
		sem <- struct{}{}
		go func(j worker.HistoryJob) {
			defer func() { <-sem }()
			worker.ScanHistory(j)
		}(job)
	}
}

func (h *Handler) handleMarketplace(ctx context.Context, e *github.MarketplacePurchaseEvent) {
	if h.db == nil {
		return
	}

	purchase := e.GetMarketplacePurchase()
	account := purchase.GetAccount()
	plan := purchase.GetPlan().GetName()

	switch e.GetAction() {
	case "purchased", "changed":
		if err := h.db.UpsertInstallation(ctx, store.Installation{
			InstallationID: account.GetID(),
			AccountLogin:   account.GetLogin(),
			AccountType:    account.GetType(),
			Plan:           plan,
		}); err != nil {
			log.Printf("marketplace upsert %s: %v", account.GetLogin(), err)
		}
		log.Printf("marketplace: %s %s → plan %s", e.GetAction(), account.GetLogin(), plan)
	case "cancelled":
		if err := h.db.UpdatePlan(ctx, account.GetID(), "free"); err != nil {
			log.Printf("marketplace cancel %s: %v", account.GetLogin(), err)
		}
		log.Printf("marketplace: cancelled %s", account.GetLogin())
	}
}
