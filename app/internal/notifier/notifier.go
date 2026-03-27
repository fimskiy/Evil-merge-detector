package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Notifier struct {
	webhookURL string
	slackURL   string
	client     *http.Client
}

func New(webhookURL, slackURL string) *Notifier {
	return &Notifier{
		webhookURL: webhookURL,
		slackURL:   slackURL,
		client:     &http.Client{Timeout: 10 * time.Second},
	}
}

func (n *Notifier) Enabled() bool {
	return n.webhookURL != "" || n.slackURL != ""
}

type Finding struct {
	File     string `json:"file"`
	Severity string `json:"severity"`
	Detail   string `json:"detail"`
}

type EvilMergeEvent struct {
	Owner      string    `json:"owner"`
	Repo       string    `json:"repo"`
	PRNumber   int       `json:"pr_number,omitempty"`
	HeadSHA    string    `json:"head_sha"`
	EvilMerges int       `json:"evil_merges"`
	RepoURL    string    `json:"repo_url"`
	Findings   []Finding `json:"findings"`
}

func (n *Notifier) Notify(ctx context.Context, ev EvilMergeEvent) error {
	var firstErr error
	if n.webhookURL != "" {
		if err := n.postJSON(ctx, n.webhookURL, ev); err != nil && firstErr == nil {
			firstErr = fmt.Errorf("webhook: %w", err)
		}
	}
	if n.slackURL != "" {
		if err := n.postSlack(ctx, ev); err != nil && firstErr == nil {
			firstErr = fmt.Errorf("slack: %w", err)
		}
	}
	return firstErr
}

func (n *Notifier) postJSON(ctx context.Context, url string, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := n.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("status %d", resp.StatusCode)
	}
	return nil
}

func (n *Notifier) postSlack(ctx context.Context, ev EvilMergeEvent) error {
	var text string
	if ev.PRNumber > 0 {
		prURL := fmt.Sprintf("https://github.com/%s/%s/pull/%d", ev.Owner, ev.Repo, ev.PRNumber)
		text = fmt.Sprintf(":warning: *Evil merge detected* in <%s|%s/%s #%d> — %d suspicious change(s)",
			prURL, ev.Owner, ev.Repo, ev.PRNumber, ev.EvilMerges)
	} else {
		repoURL := fmt.Sprintf("https://github.com/%s/%s", ev.Owner, ev.Repo)
		text = fmt.Sprintf(":warning: *Evil merge detected* in <%s|%s/%s> (history scan) — %d suspicious change(s)",
			repoURL, ev.Owner, ev.Repo, ev.EvilMerges)
	}
	payload := map[string]any{
		"text": fmt.Sprintf("Evil merge detected in %s/%s", ev.Owner, ev.Repo),
		"blocks": []map[string]any{
			{"type": "section", "text": map[string]string{"type": "mrkdwn", "text": text}},
		},
	}
	return n.postJSON(ctx, n.slackURL, payload)
}
