package oauth

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRandomState_unique(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 20; i++ {
		s, err := randomState()
		if err != nil {
			t.Fatalf("randomState error: %v", err)
		}
		if s == "" {
			t.Fatal("randomState returned empty string")
		}
		if seen[s] {
			t.Fatalf("randomState returned duplicate value: %s", s)
		}
		seen[s] = true
	}
}

func TestLogin_setsCookieAndRedirects(t *testing.T) {
	h := New("client-id", "client-secret", []byte("secret"), nil)
	r := httptest.NewRequest(http.MethodGet, "/auth/github", nil)
	w := httptest.NewRecorder()

	h.Login(w, r)

	resp := w.Result()
	if resp.StatusCode != http.StatusFound {
		t.Fatalf("expected 302, got %d", resp.StatusCode)
	}
	found := false
	for _, c := range resp.Cookies() {
		if c.Name != "oauth_state" {
			continue
		}
		found = true
		if !c.HttpOnly {
			t.Error("oauth_state cookie should be HttpOnly")
		}
		if !c.Secure {
			t.Error("oauth_state cookie should be Secure")
		}
		if c.SameSite != http.SameSiteLaxMode {
			t.Errorf("expected SameSite=Lax, got %v", c.SameSite)
		}
	}
	if !found {
		t.Fatal("oauth_state cookie not set")
	}
}

func TestCallback_invalidState(t *testing.T) {
	h := New("client-id", "client-secret", []byte("secret"), nil)
	r := httptest.NewRequest(http.MethodGet, "/auth/callback?state=bad&code=x", nil)
	r.AddCookie(&http.Cookie{Name: "oauth_state", Value: "good"})
	w := httptest.NewRecorder()

	h.Callback(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestCallback_missingStateCookie(t *testing.T) {
	h := New("client-id", "client-secret", []byte("secret"), nil)
	r := httptest.NewRequest(http.MethodGet, "/auth/callback?state=x&code=y", nil)
	w := httptest.NewRecorder()

	h.Callback(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}
