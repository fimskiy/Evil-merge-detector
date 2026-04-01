package webhook_test

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fimskiy/evil-merge-detector/app/internal/config"
	"github.com/fimskiy/evil-merge-detector/app/internal/notifier"
	"github.com/fimskiy/evil-merge-detector/app/internal/webhook"
)

const testSecret = "test-secret"

func sign(payload []byte) string {
	mac := hmac.New(sha256.New, []byte(testSecret))
	mac.Write(payload)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func send(t *testing.T, eventType string, payload []byte) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-GitHub-Event", eventType)
	req.Header.Set("X-Hub-Signature-256", sign(payload))
	rr := httptest.NewRecorder()
	cfg := &config.Config{WebhookSecret: []byte(testSecret)}
	webhook.New(cfg, nil, notifier.New("", "")).ServeHTTP(rr, req)
	return rr
}

func TestWebhook_InvalidSignature(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-GitHub-Event", "ping")
	req.Header.Set("X-Hub-Signature-256", "sha256=badbadbadbad")
	rr := httptest.NewRecorder()
	webhook.New(&config.Config{WebhookSecret: []byte(testSecret)}, nil, nil).ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status %d, want 401", rr.Code)
	}
}

func TestWebhook_UnhandledEvent_Push(t *testing.T) {
	payload := []byte(`{"ref":"refs/heads/main","repository":{"name":"repo","owner":{"login":"acme"}}}`)
	rr := send(t, "push", payload)
	if rr.Code != http.StatusOK {
		t.Errorf("status %d, want 200", rr.Code)
	}
}

func TestWebhook_PR_Opened(t *testing.T) {
	payload := []byte(`{
		"action": "opened",
		"number": 1,
		"pull_request": {"number": 1, "head": {"sha": "abc123", "ref": "feature"}},
		"repository": {"name": "repo", "clone_url": "https://github.com/acme/repo.git", "owner": {"login": "acme"}},
		"installation": {"id": 42}
	}`)
	rr := send(t, "pull_request", payload)
	if rr.Code != http.StatusOK {
		t.Errorf("status %d, want 200", rr.Code)
	}
}

func TestWebhook_PR_Synchronize(t *testing.T) {
	payload := []byte(`{
		"action": "synchronize",
		"number": 5,
		"pull_request": {"number": 5, "head": {"sha": "def456", "ref": "fix"}},
		"repository": {"name": "repo", "clone_url": "https://github.com/acme/repo.git", "owner": {"login": "acme"}},
		"installation": {"id": 42}
	}`)
	rr := send(t, "pull_request", payload)
	if rr.Code != http.StatusOK {
		t.Errorf("status %d, want 200", rr.Code)
	}
}

func TestWebhook_PR_IgnoredAction(t *testing.T) {
	// "closed" PRs should be accepted (200) but not trigger a scan.
	payload := []byte(`{
		"action": "closed",
		"number": 1,
		"pull_request": {"number": 1, "head": {"sha": "abc123", "ref": "feature"}},
		"repository": {"name": "repo", "clone_url": "https://github.com/acme/repo.git", "owner": {"login": "acme"}},
		"installation": {"id": 42}
	}`)
	rr := send(t, "pull_request", payload)
	if rr.Code != http.StatusOK {
		t.Errorf("status %d, want 200", rr.Code)
	}
}

func TestWebhook_Installation_Created_NoDB(t *testing.T) {
	payload := []byte(`{
		"action": "created",
		"installation": {"id": 99, "account": {"login": "acme", "type": "Organization"}}
	}`)
	rr := send(t, "installation", payload)
	if rr.Code != http.StatusOK {
		t.Errorf("status %d, want 200", rr.Code)
	}
}

func TestWebhook_Installation_Deleted_NoDB(t *testing.T) {
	payload := []byte(`{
		"action": "deleted",
		"installation": {"id": 99, "account": {"login": "acme", "type": "Organization"}}
	}`)
	rr := send(t, "installation", payload)
	if rr.Code != http.StatusOK {
		t.Errorf("status %d, want 200", rr.Code)
	}
}

func TestWebhook_Marketplace_Purchased_NoDB(t *testing.T) {
	payload := []byte(`{
		"action": "purchased",
		"marketplace_purchase": {
			"account": {"id": 77, "login": "acme", "type": "Organization"},
			"plan": {"name": "pro"}
		}
	}`)
	rr := send(t, "marketplace_purchase", payload)
	if rr.Code != http.StatusOK {
		t.Errorf("status %d, want 200", rr.Code)
	}
}

func TestWebhook_Marketplace_Changed_NoDB(t *testing.T) {
	payload := []byte(`{
		"action": "changed",
		"marketplace_purchase": {
			"account": {"id": 77, "login": "acme", "type": "Organization"},
			"plan": {"name": "team"}
		}
	}`)
	rr := send(t, "marketplace_purchase", payload)
	if rr.Code != http.StatusOK {
		t.Errorf("status %d, want 200", rr.Code)
	}
}

func TestWebhook_Marketplace_Cancelled_NoDB(t *testing.T) {
	payload := []byte(`{
		"action": "cancelled",
		"marketplace_purchase": {
			"account": {"id": 77, "login": "acme", "type": "Organization"},
			"plan": {"name": "pro"}
		}
	}`)
	rr := send(t, "marketplace_purchase", payload)
	if rr.Code != http.StatusOK {
		t.Errorf("status %d, want 200", rr.Code)
	}
}
