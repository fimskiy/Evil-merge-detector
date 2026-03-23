package webhook

import (
	"log"
	"net/http"

	"github.com/google/go-github/v84/github"

	"github.com/fimskiy/evil-merge-detector/app/internal/config"
	"github.com/fimskiy/evil-merge-detector/app/internal/worker"
)

type Handler struct {
	cfg *config.Config
}

func New(cfg *config.Config) http.Handler {
	return &Handler{cfg: cfg}
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

	switch e := event.(type) {
	case *github.PullRequestEvent:
		h.handlePR(e)
	}
}

func (h *Handler) handlePR(e *github.PullRequestEvent) {
	action := e.GetAction()
	if action != "opened" && action != "synchronize" && action != "reopened" {
		return
	}

	pr := e.GetPullRequest()
	job := worker.PRJob{
		Owner:          e.GetRepo().GetOwner().GetLogin(),
		Repo:           e.GetRepo().GetName(),
		CloneURL:       e.GetRepo().GetCloneURL(),
		HeadSHA:        pr.GetHead().GetSHA(),
		HeadRef:        pr.GetHead().GetRef(),
		InstallationID: e.GetInstallation().GetID(),
		AppID:          h.cfg.AppID,
		PrivateKey:     h.cfg.PrivateKey,
	}

	log.Printf("scanning PR #%d in %s/%s (%.7s)", pr.GetNumber(), job.Owner, job.Repo, job.HeadSHA)
	go worker.ScanPR(job)
}
