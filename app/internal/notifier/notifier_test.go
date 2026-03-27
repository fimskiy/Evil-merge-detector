package notifier_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fimskiy/evil-merge-detector/app/internal/notifier"
)

func TestEnabled(t *testing.T) {
	if notifier.New("", "").Enabled() {
		t.Error("empty notifier should not be enabled")
	}
	if !notifier.New("http://example.com", "").Enabled() {
		t.Error("webhook-only notifier should be enabled")
	}
	if !notifier.New("", "http://example.com").Enabled() {
		t.Error("slack-only notifier should be enabled")
	}
}

func TestNotify_Webhook_Payload(t *testing.T) {
	var got notifier.EvilMergeEvent
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method %s, want POST", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("Content-Type %q", ct)
		}
		json.NewDecoder(r.Body).Decode(&got)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	ev := notifier.EvilMergeEvent{
		Owner:      "acme",
		Repo:       "myrepo",
		PRNumber:   42,
		HeadSHA:    "abc1234",
		EvilMerges: 2,
		RepoURL:    "https://github.com/acme/myrepo",
		Findings: []notifier.Finding{
			{File: "vite.config.js", Severity: "CRITICAL", Detail: "modified in merge"},
		},
	}

	n := notifier.New(srv.URL, "")
	if err := n.Notify(context.Background(), ev); err != nil {
		t.Fatal(err)
	}

	if got.Owner != "acme" {
		t.Errorf("owner %q", got.Owner)
	}
	if got.Repo != "myrepo" {
		t.Errorf("repo %q", got.Repo)
	}
	if got.PRNumber != 42 {
		t.Errorf("pr_number %d", got.PRNumber)
	}
	if got.EvilMerges != 2 {
		t.Errorf("evil_merges %d", got.EvilMerges)
	}
	if len(got.Findings) != 1 || got.Findings[0].File != "vite.config.js" {
		t.Errorf("findings %+v", got.Findings)
	}
}

func TestNotify_Slack_Payload_PR(t *testing.T) {
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&body)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	ev := notifier.EvilMergeEvent{Owner: "acme", Repo: "repo", PRNumber: 7, EvilMerges: 1}
	n := notifier.New("", srv.URL)
	if err := n.Notify(context.Background(), ev); err != nil {
		t.Fatal(err)
	}

	text, _ := body["text"].(string)
	if !strings.Contains(text, "acme/repo") {
		t.Errorf("slack text missing repo: %q", text)
	}

	blocks, _ := body["blocks"].([]any)
	if len(blocks) == 0 {
		t.Fatal("expected blocks in slack payload")
	}
	block, _ := blocks[0].(map[string]any)
	blockText, _ := block["text"].(map[string]any)
	mrkdwn, _ := blockText["text"].(string)
	if !strings.Contains(mrkdwn, "pull/7") {
		t.Errorf("slack mrkdwn missing PR URL: %q", mrkdwn)
	}
}

func TestNotify_Slack_Payload_History(t *testing.T) {
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&body)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	// PRNumber == 0 → history scan message
	ev := notifier.EvilMergeEvent{Owner: "acme", Repo: "repo", PRNumber: 0, EvilMerges: 1}
	n := notifier.New("", srv.URL)
	if err := n.Notify(context.Background(), ev); err != nil {
		t.Fatal(err)
	}

	blocks, _ := body["blocks"].([]any)
	block, _ := blocks[0].(map[string]any)
	blockText, _ := block["text"].(map[string]any)
	mrkdwn, _ := blockText["text"].(string)
	if !strings.Contains(mrkdwn, "history scan") {
		t.Errorf("expected 'history scan' in slack message: %q", mrkdwn)
	}
}

func TestNotify_BothChannels(t *testing.T) {
	webhookCalled, slackCalled := false, false

	webhook := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		webhookCalled = true
		w.WriteHeader(http.StatusOK)
	}))
	defer webhook.Close()

	slack := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slackCalled = true
		w.WriteHeader(http.StatusOK)
	}))
	defer slack.Close()

	n := notifier.New(webhook.URL, slack.URL)
	ev := notifier.EvilMergeEvent{Owner: "a", Repo: "b", EvilMerges: 1}
	if err := n.Notify(context.Background(), ev); err != nil {
		t.Fatal(err)
	}
	if !webhookCalled {
		t.Error("webhook endpoint was not called")
	}
	if !slackCalled {
		t.Error("slack endpoint was not called")
	}
}

func TestNotify_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	n := notifier.New(srv.URL, "")
	ev := notifier.EvilMergeEvent{Owner: "a", Repo: "b", EvilMerges: 1}
	if err := n.Notify(context.Background(), ev); err == nil {
		t.Error("expected error on 500 response")
	}
}

func TestNotify_PRNumberOmittedWhenZero(t *testing.T) {
	var raw map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&raw)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	ev := notifier.EvilMergeEvent{Owner: "a", Repo: "b", PRNumber: 0, EvilMerges: 1}
	n := notifier.New(srv.URL, "")
	n.Notify(context.Background(), ev)

	if _, ok := raw["pr_number"]; ok {
		t.Error("pr_number should be omitted when zero (omitempty)")
	}
}
