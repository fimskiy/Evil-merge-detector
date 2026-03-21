package detector_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/fimskiy/evil-merge-detector/internal/detector"
	"github.com/fimskiy/evil-merge-detector/internal/models"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// setupTestRepo creates a minimal test repo with one evil merge and one clean merge.
func setupTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	cmds := [][]string{
		{"git", "init"},
		{"git", "config", "user.email", "test@test.com"},
		{"git", "config", "user.name", "Test"},
	}

	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("cmd %v failed: %v\n%s", args, err, out)
		}
	}

	// Helper to run git commands
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v failed: %v\n%s", args, err, out)
		}
	}

	writeFile := func(name, content string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Initial commit
	writeFile("base.txt", "base content")
	writeFile("shared.txt", "shared content")
	run("add", "-A")
	run("commit", "-m", "Initial commit")

	// === Clean merge ===
	run("checkout", "-b", "clean-feature")
	writeFile("new.txt", "new feature")
	run("add", "new.txt")
	run("commit", "-m", "Add new feature")

	run("checkout", "main")
	run("merge", "clean-feature", "--no-ff", "-m", "Merge clean-feature")

	// === Evil merge: modify untouched file ===
	run("checkout", "-b", "feature-x")
	writeFile("base.txt", "modified by feature-x")
	run("add", "base.txt")
	run("commit", "-m", "Feature X changes base.txt")

	run("checkout", "main")

	// Do the merge
	run("merge", "feature-x", "--no-ff", "-m", "Merge feature-x")

	// Now amend the merge to sneak in evil change
	writeFile("shared.txt", "EVIL CHANGE - not in any branch")
	run("add", "shared.txt")
	run("commit", "--amend", "-m", "Merge feature-x (evil)")

	return dir
}

func TestDetector_CleanMerge(t *testing.T) {
	dir := setupTestRepo(t)

	repo, err := git.PlainOpen(dir)
	if err != nil {
		t.Fatal(err)
	}

	// Find the clean merge (first merge commit in history)
	iter, err := repo.Log(&git.LogOptions{})
	if err != nil {
		t.Fatal(err)
	}

	var cleanMerge *object.Commit
	var evilMerge *object.Commit

	iter.ForEach(func(c *object.Commit) error {
		if c.NumParents() >= 2 {
			if evilMerge == nil {
				evilMerge = c // Most recent merge (evil)
			} else if cleanMerge == nil {
				cleanMerge = c // Older merge (clean)
			}
		}
		return nil
	})

	d := detector.New()

	// Test clean merge
	if cleanMerge == nil {
		t.Fatal("clean merge not found")
	}
	report, err := d.AnalyzeMerge(cleanMerge)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.EvilChanges) != 0 {
		t.Errorf("clean merge should have 0 evil changes, got %d: %+v", len(report.EvilChanges), report.EvilChanges)
	}

	// Test evil merge
	if evilMerge == nil {
		t.Fatal("evil merge not found")
	}
	report, err = d.AnalyzeMerge(evilMerge)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.EvilChanges) == 0 {
		t.Error("evil merge should have at least 1 evil change")
	}

	// Check that shared.txt is flagged
	found := false
	for _, ec := range report.EvilChanges {
		if ec.FilePath == "shared.txt" {
			found = true
			if ec.Severity < models.SeverityWarning {
				t.Errorf("shared.txt should be at least WARNING severity, got %s", ec.Severity)
			}
		}
	}
	if !found {
		t.Error("shared.txt should be detected as evil change")
	}
}

func TestDetector_NonMergeCommit(t *testing.T) {
	dir := setupTestRepo(t)

	repo, err := git.PlainOpen(dir)
	if err != nil {
		t.Fatal(err)
	}

	iter, err := repo.Log(&git.LogOptions{})
	if err != nil {
		t.Fatal(err)
	}

	// Find a non-merge commit
	var nonMerge *object.Commit
	iter.ForEach(func(c *object.Commit) error {
		if c.NumParents() == 1 && nonMerge == nil {
			nonMerge = c
		}
		return nil
	})

	d := detector.New()
	_, err = d.AnalyzeMerge(nonMerge)
	if err == nil {
		t.Error("should return error for non-merge commit")
	}
}
