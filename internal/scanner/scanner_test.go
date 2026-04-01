package scanner_test

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/evilmerge-dev/evil-merge-detector/internal/models"
	"github.com/evilmerge-dev/evil-merge-detector/internal/scanner"
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
	result, err := s.Scan(context.Background(), models.ScanOptions{
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
	result, err := s.Scan(context.Background(), models.ScanOptions{
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
	result, err := s.Scan(context.Background(), models.ScanOptions{
		RepoPath:    dir,
		MinSeverity: models.SeverityCritical,
	})
	if err != nil {
		t.Fatal(err)
	}

	for _, r := range result.Reports {
		if r.MaxSeverity < models.SeverityCritical {
			t.Errorf("expected only CRITICAL merges, got %s for %s", r.MaxSeverity, r.ShortHash)
		}
	}
}

func TestScanner_Workers(t *testing.T) {
	dir := setupTestRepo(t)
	s := scanner.New()
	opts := models.ScanOptions{RepoPath: dir}

	seq, err := s.Scan(context.Background(), opts)
	if err != nil {
		t.Fatal(err)
	}

	opts.Workers = 4
	par, err := s.Scan(context.Background(), opts)
	if err != nil {
		t.Fatal(err)
	}

	if seq.EvilMerges != par.EvilMerges {
		t.Errorf("workers=4 found %d evil merges, sequential found %d", par.EvilMerges, seq.EvilMerges)
	}
	if seq.TotalMerges != par.TotalMerges {
		t.Errorf("workers=4 total=%d, sequential total=%d", par.TotalMerges, seq.TotalMerges)
	}
}

func TestScanner_Verbose(t *testing.T) {
	dir := setupTestRepo(t)
	s := scanner.New()

	var buf bytes.Buffer
	_, err := s.Scan(context.Background(), models.ScanOptions{
		RepoPath: dir,
		Progress: &buf,
	})
	if err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "[1/") {
		t.Errorf("expected progress output like [1/N], got: %q", out)
	}
}

func TestScanner_Exclude(t *testing.T) {
	dir := setupTestRepo(t)
	s := scanner.New()

	result, err := s.Scan(context.Background(), models.ScanOptions{
		RepoPath: dir,
		Exclude:  []string{"file.txt"},
	})
	if err != nil {
		t.Fatal(err)
	}

	for _, r := range result.Reports {
		for _, ec := range r.EvilChanges {
			if ec.FilePath == "file.txt" {
				t.Error("file.txt should be excluded from findings")
			}
		}
	}
}

func TestScanner_Include(t *testing.T) {
	dir := setupTestRepo(t)
	s := scanner.New()

	// Only include .md files — there are no evil .md changes, so reports should be empty.
	result, err := s.Scan(context.Background(), models.ScanOptions{
		RepoPath: dir,
		Include:  []string{"*.md"},
	})
	if err != nil {
		t.Fatal(err)
	}

	if result.EvilMerges != 0 {
		t.Errorf("expected 0 evil merges with include=*.md, got %d", result.EvilMerges)
	}
}

func TestScanner_IgnoreHash(t *testing.T) {
	dir := setupTestRepo(t)

	cmd := exec.Command("git", "log", "--merges", "--format=%H", "-1")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		t.Fatal(err)
	}
	evilHash := strings.TrimSpace(string(out))
	if evilHash == "" {
		t.Fatal("could not find evil merge hash")
	}

	s := scanner.New()
	result, err := s.Scan(context.Background(), models.ScanOptions{
		RepoPath:      dir,
		IgnoredHashes: []string{evilHash[:7]},
	})
	if err != nil {
		t.Fatal(err)
	}

	for _, r := range result.Reports {
		if strings.HasPrefix(r.CommitHash, evilHash[:7]) {
			t.Errorf("commit %s should have been ignored", evilHash[:7])
		}
	}
}

func TestScanner_IgnoreAuthor(t *testing.T) {
	dir := t.TempDir()
	run := makeRunner(t, dir)
	writeFile := makeWriter(t, dir)

	run("init", "-b", "main")
	run("config", "user.email", "honest@dev.com")
	run("config", "user.name", "Honest Dev")

	writeFile("file.txt", "initial")
	run("add", "-A")
	run("commit", "-m", "init")

	run("checkout", "-b", "feature")
	writeFile("feature.txt", "feature")
	run("add", "feature.txt")
	run("commit", "-m", "feature commit")

	run("checkout", "main")

	// Change author to someone we'll ignore, then do the evil merge
	run("config", "user.email", "evil@contractor.com")
	run("config", "user.name", "Evil Contractor")
	run("merge", "feature", "--no-ff", "-m", "Merge feature")
	writeFile("file.txt", "sneaky change")
	run("add", "file.txt")
	run("commit", "--amend", "-m", "Merge feature (evil)")

	s := scanner.New()
	result, err := s.Scan(context.Background(), models.ScanOptions{
		RepoPath:       dir,
		IgnoredAuthors: []string{"evil@contractor.com"},
	})
	if err != nil {
		t.Fatal(err)
	}

	if result.EvilMerges != 0 {
		t.Errorf("evil merge by ignored author should not be reported, got %d", result.EvilMerges)
	}
}

func TestScanner_IgnoreBots(t *testing.T) {
	dir := t.TempDir()
	run := makeRunner(t, dir)
	writeFile := makeWriter(t, dir)

	run("init", "-b", "main")
	run("config", "user.email", "honest@dev.com")
	run("config", "user.name", "Honest Dev")

	writeFile("file.txt", "initial")
	run("add", "-A")
	run("commit", "-m", "init")

	run("checkout", "-b", "deps")
	writeFile("go.sum", "hash1")
	run("add", "go.sum")
	run("commit", "-m", "update deps")

	run("checkout", "main")

	// Bot does the merge
	run("config", "user.name", "dependabot[bot]")
	run("config", "user.email", "49699333+dependabot[bot]@users.noreply.github.com")
	run("merge", "deps", "--no-ff", "-m", "Bump dependency")
	writeFile("file.txt", "sneaky bot change")
	run("add", "file.txt")
	run("commit", "--amend", "-m", "Bump dependency (evil)")

	s := scanner.New()
	result, err := s.Scan(context.Background(), models.ScanOptions{
		RepoPath:   dir,
		IgnoreBots: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	if result.EvilMerges != 0 {
		t.Errorf("bot evil merge should be skipped with --ignore-bots, got %d", result.EvilMerges)
	}

	// Without the flag it should be detected
	result2, err := s.Scan(context.Background(), models.ScanOptions{RepoPath: dir})
	if err != nil {
		t.Fatal(err)
	}
	if result2.EvilMerges == 0 {
		t.Error("without --ignore-bots the evil merge should be detected")
	}
}

func TestScanner_SinceTag(t *testing.T) {
	dir := t.TempDir()
	run := makeRunner(t, dir)
	writeFile := makeWriter(t, dir)

	run("init", "-b", "main")
	run("config", "user.email", "test@test.com")
	run("config", "user.name", "Test")

	// Old evil merge (before tag)
	writeFile("old.txt", "initial")
	run("add", "-A")
	run("commit", "-m", "init")

	run("checkout", "-b", "old-feature")
	writeFile("other.txt", "other")
	run("add", "other.txt")
	run("commit", "-m", "old feature")
	run("checkout", "main")
	run("merge", "old-feature", "--no-ff", "-m", "Old merge")
	writeFile("old.txt", "old evil change")
	run("add", "old.txt")
	run("commit", "--amend", "-m", "Old merge (evil)")

	// Commit that the tag will point to (after the evil merge)
	writeFile("release.txt", "v1.0")
	run("add", "release.txt")
	run("commit", "-m", "release commit")
	run("tag", "v1.0")

	// New clean commit after tag
	writeFile("new.txt", "new")
	run("add", "new.txt")
	run("commit", "-m", "post-release commit")

	s := scanner.New()
	result, err := s.Scan(context.Background(), models.ScanOptions{
		RepoPath: dir,
		SinceTag: "v1.0",
	})
	if err != nil {
		t.Fatal(err)
	}

	// The evil merge is before the tag, so --since-tag v1.0 should not include it
	if result.EvilMerges != 0 {
		t.Errorf("expected 0 evil merges after tag v1.0, got %d", result.EvilMerges)
	}
}

// makeRunner returns a helper that runs git commands in dir and fails the test on error.
func makeRunner(t *testing.T, dir string) func(args ...string) {
	t.Helper()
	return func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
}

// makeWriter returns a helper that writes files in dir.
func makeWriter(t *testing.T, dir string) func(name, content string) {
	t.Helper()
	return func(name, content string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
}
