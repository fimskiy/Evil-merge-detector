package scanner_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/fimskiy/evil-merge-detector/internal/models"
	"github.com/fimskiy/evil-merge-detector/internal/scanner"
)

func setupTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

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

	run("init", "-b", "main")
	run("config", "user.email", "test@test.com")
	run("config", "user.name", "Test")

	// Initial commit
	writeFile("file.txt", "initial")
	run("add", "-A")
	run("commit", "-m", "init")

	// Clean merge
	run("checkout", "-b", "clean")
	writeFile("clean.txt", "clean")
	run("add", "clean.txt")
	run("commit", "-m", "clean feature")
	run("checkout", "main")
	run("merge", "clean", "--no-ff", "-m", "Merge clean")

	// Evil merge
	run("checkout", "-b", "evil")
	writeFile("feature.txt", "feature")
	run("add", "feature.txt")
	run("commit", "-m", "evil feature")
	run("checkout", "main")
	run("merge", "evil", "--no-ff", "-m", "Merge evil")
	writeFile("file.txt", "sneaky change")
	run("add", "file.txt")
	run("commit", "--amend", "-m", "Merge evil (with sneaky change)")

	return dir
}

func TestScanner_Scan(t *testing.T) {
	dir := setupTestRepo(t)

	s := scanner.New()
	result, err := s.Scan(models.ScanOptions{
		RepoPath: dir,
	})
	if err != nil {
		t.Fatal(err)
	}

	if result.TotalMerges != 2 {
		t.Errorf("expected 2 total merges, got %d", result.TotalMerges)
	}

	if result.EvilMerges < 1 {
		t.Errorf("expected at least 1 evil merge, got %d", result.EvilMerges)
	}
}

func TestScanner_ScanWithLimit(t *testing.T) {
	dir := setupTestRepo(t)

	s := scanner.New()
	result, err := s.Scan(models.ScanOptions{
		RepoPath: dir,
		Limit:    1,
	})
	if err != nil {
		t.Fatal(err)
	}

	// With limit=1, at most 1 merge should be analyzed (reported)
	analyzed := result.EvilMerges + (result.TotalMerges - result.EvilMerges)
	_ = analyzed
	if len(result.Reports) > 1 {
		t.Errorf("expected at most 1 analyzed merge with limit=1, got %d reports", len(result.Reports))
	}
}

func TestScanner_ScanWithSeverityFilter(t *testing.T) {
	dir := setupTestRepo(t)

	s := scanner.New()
	result, err := s.Scan(models.ScanOptions{
		RepoPath:    dir,
		MinSeverity: models.SeverityCritical,
	})
	if err != nil {
		t.Fatal(err)
	}

	// All reported merges should be CRITICAL
	for _, r := range result.Reports {
		if r.MaxSeverity < models.SeverityCritical {
			t.Errorf("expected only CRITICAL merges, got %s for %s", r.MaxSeverity, r.ShortHash)
		}
	}
}
