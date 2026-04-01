package badge_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fimskiy/evil-merge-detector/app/internal/badge"
	"github.com/fimskiy/evil-merge-detector/app/internal/store"
)

type mockStore struct {
	rec *store.ScanRecord
	err error
}

func (m *mockStore) LastScan(_ context.Context, _, _ string) (*store.ScanRecord, error) {
	return m.rec, m.err
}

func serve(t *testing.T, db badge.ScanStore, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rr := httptest.NewRecorder()
	badge.New(db).ServeHTTP(rr, req)
	return rr
}

func TestBadge_NoDB(t *testing.T) {
	rr := serve(t, nil, "/badge/owner/repo")
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "unknown") {
		t.Error("expected 'unknown' when db is nil")
	}
}

func TestBadge_Passing(t *testing.T) {
	rr := serve(t, &mockStore{rec: &store.ScanRecord{EvilMerges: 0}}, "/badge/acme/repo")
	body := rr.Body.String()
	if !strings.Contains(body, "passing") {
		t.Error("expected 'passing'")
	}
	if !strings.Contains(body, "#4c1") {
		t.Error("expected green color #4c1")
	}
}

func TestBadge_Found(t *testing.T) {
	rr := serve(t, &mockStore{rec: &store.ScanRecord{EvilMerges: 3}}, "/badge/acme/repo")
	body := rr.Body.String()
	if !strings.Contains(body, "3 found") {
		t.Errorf("expected '3 found', got: %s", body)
	}
	if !strings.Contains(body, "#e05d44") {
		t.Error("expected red color #e05d44")
	}
}

func TestBadge_DBError(t *testing.T) {
	rr := serve(t, &mockStore{err: errors.New("connection refused")}, "/badge/acme/repo")
	if !strings.Contains(rr.Body.String(), "unknown") {
		t.Error("expected 'unknown' when DB returns error")
	}
}

func TestBadge_SVGSuffix(t *testing.T) {
	rr := serve(t, &mockStore{rec: &store.ScanRecord{EvilMerges: 0}}, "/badge/acme/repo.svg")
	if !strings.Contains(rr.Body.String(), "passing") {
		t.Error("expected .svg suffix to be stripped and badge rendered correctly")
	}
}

func TestBadge_ContentType(t *testing.T) {
	rr := serve(t, nil, "/badge/a/b")
	if ct := rr.Header().Get("Content-Type"); ct != "image/svg+xml" {
		t.Errorf("Content-Type %q, want image/svg+xml", ct)
	}
}

func TestBadge_CacheControl(t *testing.T) {
	rr := serve(t, nil, "/badge/a/b")
	if cc := rr.Header().Get("Cache-Control"); cc != "no-cache, max-age=0" {
		t.Errorf("Cache-Control %q", cc)
	}
}

func TestBadge_InvalidPath(t *testing.T) {
	for _, path := range []string{"/badge/", "/badge/onlyowner"} {
		rr := serve(t, nil, path)
		if rr.Code != http.StatusBadRequest {
			t.Errorf("path %q: status %d, want 400", path, rr.Code)
		}
	}
}

func TestBadge_SVGContainsLabel(t *testing.T) {
	rr := serve(t, nil, "/badge/x/y")
	if !strings.Contains(rr.Body.String(), "evil merges") {
		t.Error("expected 'evil merges' label in SVG")
	}
}
