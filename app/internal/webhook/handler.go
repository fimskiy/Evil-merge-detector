package webhook

import (
	"context"
	"log"
	"net/http"
	"sync"
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
	default:
		log.Printf("webhook: unhandled event type %T", e)
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
	var wg sync.WaitGroup

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
		wg.Add(1)
		go func(j worker.HistoryJob) {
			defer wg.Done()
			defer func() { <-sem }()
			worker.ScanHistory(j)
		}(job)
	}
	wg.Wait()
}

func (h *Handler) handleMarketplace(ctx context.Context, e *github.MarketplacePurchaseEvent) {
	if h.db == nil {
		return
	}

	purchase := e.GetMarketplacePurchase()
	account := purchase.GetAccount()
	plan := purchase.GetPlan().GetName()
	login := account.GetLogin()

	// Marketplace events carry account.id (user/org ID), not installation.id.
	// Look up the real installation by login to update its plan.
	inst, err := h.db.GetInstallationByLogin(ctx, login)
	if err != nil {
		log.Printf("marketplace: installation not found for %s: %v", login, err)
		return
	}

	switch e.GetAction() {
	case "purchased", "changed":
		if err := h.db.UpdatePlan(ctx, inst.InstallationID, plan); err != nil {
			log.Printf("marketplace upsert %s: %v", login, err)
		}
		log.Printf("marketplace: %s %s → plan %s", e.GetAction(), login, plan)
	case "cancelled":
		if err := h.db.UpdatePlan(ctx, inst.InstallationID, "free"); err != nil {
			log.Printf("marketplace cancel %s: %v", login, err)
		}
		log.Printf("marketplace: cancelled %s", login)
	}
}
