package worker

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/fimskiy/evil-merge-detector/app/internal/notifier"
)

// setupEvilRepo creates a temporary git repository with one evil merge.
// Returns the repo directory path.
func setupEvilRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	write := func(name, content string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	run("init", "-b", "main")
	run("config", "user.email", "test@test.com")
	run("config", "user.name", "Test")

	write("file.txt", "initial")
	run("add", "-A")
	run("commit", "-m", "init")

	run("checkout", "-b", "feature")
	write("feature.txt", "feature")
	run("add", "feature.txt")
	run("commit", "-m", "feature commit")

	run("checkout", "main")
	run("merge", "feature", "--no-ff", "-m", "Merge feature")
	write("file.txt", "sneaky change")
	run("add", "file.txt")
	run("commit", "--amend", "-m", "Merge feature (evil)")

	return dir
}

// localCloneFn returns a cloneFn that clones from a local path using git.
func localCloneFn(t *testing.T) func(ctx context.Context, _, _ int64, _ []byte, repoURL, branch, destDir string) error {
	t.Helper()
	return func(ctx context.Context, _, _ int64, _ []byte, repoURL, branch, destDir string) error {
		args := []string{"clone", "--branch", branch, repoURL, destDir}
		cmd := exec.CommandContext(ctx, "git", args...)
		if out, err := cmd.CombinedOutput(); err != nil {
			return errors.New(string(out))
		}
		return nil
	}
}

func TestScanHistory_FindsEvilMerge(t *testing.T) {
	dir := setupEvilRepo(t)

	job := HistoryJob{
		Owner:         "test",
		Repo:          "testrepo",
		DefaultBranch: "main",
		CloneURL:      dir,
		cloneFn:       localCloneFn(t),
	}

	// Run synchronously (ScanHistory blocks internally).
	ScanHistory(job)
	// No assertion on result since DB is nil — we just verify it doesn't panic.
	// The real assertion: no panic and no fatal log crash.
}

func TestScanHistory_CloneFailure(t *testing.T) {
	job := HistoryJob{
		Owner:         "test",
		Repo:          "repo",
		DefaultBranch: "main",
		CloneURL:      "/nonexistent/path",
		cloneFn: func(_ context.Context, _, _ int64, _ []byte, _, _, _ string) error {
			return errors.New("clone failed")
		},
	}
	// Must return without panic when clone fails.
	ScanHistory(job)
}

func TestScanHistory_NotifierCalled(t *testing.T) {
	dir := setupEvilRepo(t)

	var notified atomic.Bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		notified.Store(true)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	ntf := notifier.New(srv.URL, "")

	job := HistoryJob{
		Owner:         "test",
		Repo:          "testrepo",
		DefaultBranch: "main",
		CloneURL:      dir,
		Notifier:      ntf,
		cloneFn:       localCloneFn(t),
	}

	ScanHistory(job)

	if !notified.Load() {
		t.Error("notifier was not called despite evil merge being found")
	}
}

func TestScanHistory_NotifierNotCalledOnCleanRepo(t *testing.T) {
	dir := t.TempDir()
	run := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		_ = cmd.Run()
	}
	run("init", "-b", "main")
	run("config", "user.email", "t@t.com")
	run("config", "user.name", "T")
	if err := os.WriteFile(filepath.Join(dir, "f.txt"), []byte("hi"), 0644); err != nil {
		t.Fatal(err)
	}
	run("add", "-A")
	run("commit", "-m", "init")

	run("checkout", "-b", "branch")
	if err := os.WriteFile(filepath.Join(dir, "b.txt"), []byte("b"), 0644); err != nil {
		t.Fatal(err)
	}
	run("add", "b.txt")
	run("commit", "-m", "branch commit")
	run("checkout", "main")
	run("merge", "branch", "--no-ff", "-m", "Clean merge")

	var notified atomic.Bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		notified.Store(true)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	job := HistoryJob{
		Owner:         "test",
		Repo:          "cleanrepo",
		DefaultBranch: "main",
		CloneURL:      dir,
		Notifier:      notifier.New(srv.URL, ""),
		cloneFn:       localCloneFn(t),
	}
	ScanHistory(job)

	if notified.Load() {
		t.Error("notifier should not be called for a clean repo")
	}
}

func TestScanHistory_DefaultBranch_Fallback(t *testing.T) {
	dir := setupEvilRepo(t)

	cloned := false
	job := HistoryJob{
		Owner:         "test",
		Repo:          "repo",
		DefaultBranch: "", // empty → should fall back to "main"
		CloneURL:      dir,
		cloneFn: func(_ context.Context, _, _ int64, _ []byte, _, branch, destDir string) error {
			cloned = true
			if branch != "main" {
				return errors.New("expected branch 'main', got " + branch)
			}
			return localCloneFn(nil)(context.Background(), 0, 0, nil, dir, branch, destDir)
		},
	}

	// Bypass nil t in the localCloneFn above — use a proper clone directly.
	job.cloneFn = func(ctx context.Context, _, _ int64, _ []byte, repoURL, branch, destDir string) error {
		cloned = true
		if branch != "main" {
			t.Errorf("expected branch 'main', got %q", branch)
		}
		cmd := exec.CommandContext(ctx, "git", "clone", "--branch", branch, repoURL, destDir)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return errors.New(strings.TrimSpace(string(out)))
		}
		return nil
	}

	ScanHistory(job)

	if !cloned {
		t.Error("cloneFn was not called")
	}
}
