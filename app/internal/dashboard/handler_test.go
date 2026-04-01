package dashboard

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/evilmerge-dev/evil-merge-detector/app/internal/session"
)

var testSecret = []byte("test-secret-32-bytes-long-enough!")

func TestServeHTTP_redirectsWhenNoSession(t *testing.T) {
	h := New(testSecret, nil)
	r := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, r)

	if w.Code != http.StatusFound {
		t.Fatalf("expected 302 redirect, got %d", w.Code)
	}
	if loc := w.Header().Get("Location"); loc != "/auth/github" {
		t.Fatalf("expected redirect to /auth/github, got %s", loc)
	}
}

func TestServeHTTP_rendersWithValidSession(t *testing.T) {
	h := New(testSecret, nil)

	// Build a request with a valid signed session cookie
	setCookieResp := httptest.NewRecorder()
	session.Set(setCookieResp, &session.Data{GitHubLogin: "alice", GitHubID: 1}, testSecret)
	cookies := setCookieResp.Result().Header["Set-Cookie"]

	r := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	r.Header["Cookie"] = cookies
	w := httptest.NewRecorder()

	h.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "alice") {
		t.Error("expected login 'alice' in dashboard HTML")
	}
}

func TestServeHTTP_wrongSecretRedirects(t *testing.T) {
	// Cookie signed with different secret should be rejected
	setCookieResp := httptest.NewRecorder()
	session.Set(setCookieResp, &session.Data{GitHubLogin: "alice"}, []byte("other-secret-value-here-is-long!"))
	cookies := setCookieResp.Result().Header["Set-Cookie"]

	h := New(testSecret, nil)
	r := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	r.Header["Cookie"] = cookies
	w := httptest.NewRecorder()

	h.ServeHTTP(w, r)

	if w.Code != http.StatusFound {
		t.Fatalf("expected 302, got %d", w.Code)
	}
}
