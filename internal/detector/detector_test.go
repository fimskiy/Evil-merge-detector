package detector_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/evilmerge-dev/evil-merge-detector/internal/detector"
	"github.com/evilmerge-dev/evil-merge-detector/internal/models"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// setupTestRepo creates a minimal test repo with one evil merge and one clean merge.
func setupTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	cmds := [][]string{
		{"git", "init", "-b", "main"},
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

	if err := iter.ForEach(func(c *object.Commit) error {
		if c.NumParents() >= 2 {
			if evilMerge == nil {
				evilMerge = c // Most recent merge (evil)
			} else if cleanMerge == nil {
				cleanMerge = c // Older merge (clean)
			}
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	d := detector.New()

	// Test clean merge
	if cleanMerge == nil {
		t.Fatal("clean merge not found")
	}
	report, err := d.AnalyzeMerge(context.Background(), cleanMerge)
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
	report, err = d.AnalyzeMerge(context.Background(), evilMerge)
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

// setupEvilMergeWith creates a repo where `fileName` is unchanged in both
// parent branches but the merge commit sets its content to `evilContent`.
// This triggers the highest-severity ("unchanged in both branches") detection path.
func setupEvilMergeWith(t *testing.T, fileName, evilContent string) string {
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
	writeFile := func(name string, content []byte) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(dir, name), content, 0644); err != nil {
			t.Fatal(err)
		}
	}

	run("init", "-b", "main")
	run("config", "user.email", "test@test.com")
	run("config", "user.name", "Test")

	// Initial commit with the file we'll poison
	writeFile(fileName, []byte("original content"))
	writeFile("other.txt", []byte("other"))
	run("add", "-A")
	run("commit", "-m", "initial")

	// Feature branch that doesn't touch fileName
	run("checkout", "-b", "feature")
	writeFile("feature.txt", []byte("feature"))
	run("add", "feature.txt")
	run("commit", "-m", "feature commit")

	// Merge on main — then amend with evil content
	run("checkout", "main")
	run("merge", "feature", "--no-ff", "-m", "Merge feature")
	writeFile(fileName, []byte(evilContent))
	run("add", fileName)
	run("commit", "--amend", "-m", "Merge feature (evil)")

	return dir
}

func findLatestMerge(t *testing.T, dir string) *object.Commit {
	t.Helper()
	repo, err := git.PlainOpen(dir)
	if err != nil {
		t.Fatal(err)
	}
	iter, err := repo.Log(&git.LogOptions{})
	if err != nil {
		t.Fatal(err)
	}
	var merge *object.Commit
	_ = iter.ForEach(func(c *object.Commit) error {
		if c.NumParents() == 2 && merge == nil {
			merge = c
		}
		return nil
	})
	if merge == nil {
		t.Fatal("no merge commit found")
	}
	return merge
}

func TestDetector_SensitivePath_ViteConfig(t *testing.T) {
	dir := setupEvilMergeWith(t, "vite.config.js", "export default { plugins: [] }")
	d := detector.New()
	report, err := d.AnalyzeMerge(context.Background(), findLatestMerge(t, dir))
	if err != nil {
		t.Fatal(err)
	}
	if len(report.EvilChanges) == 0 {
		t.Fatal("expected evil changes")
	}
	for _, ec := range report.EvilChanges {
		if ec.FilePath == "vite.config.js" && ec.Severity != models.SeverityCritical {
			t.Errorf("vite.config.js should be CRITICAL, got %s", ec.Severity)
		}
	}
}

func TestDetector_SensitivePath_CIWorkflow(t *testing.T) {
	// .github/workflows/ path should trigger CRITICAL
	dir := t.TempDir()
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	writeFile := func(name string, content string) {
		t.Helper()
		full := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	run("init", "-b", "main")
	run("config", "user.email", "test@test.com")
	run("config", "user.name", "Test")

	writeFile(".github/workflows/ci.yml", "name: CI\non: push")
	writeFile("main.go", "package main")
	run("add", "-A")
	run("commit", "-m", "initial")

	run("checkout", "-b", "feature")
	writeFile("feature.txt", "feature")
	run("add", "feature.txt")
	run("commit", "-m", "feature")

	run("checkout", "main")
	run("merge", "feature", "--no-ff", "-m", "Merge feature")
	writeFile(".github/workflows/ci.yml", "name: CI\non: push\n# evil step added")
	run("add", ".github/workflows/ci.yml")
	run("commit", "--amend", "-m", "Merge feature (evil)")

	d := detector.New()
	report, err := d.AnalyzeMerge(context.Background(), findLatestMerge(t, dir))
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, ec := range report.EvilChanges {
		if strings.Contains(ec.FilePath, "workflows") {
			found = true
			if ec.Severity != models.SeverityCritical {
				t.Errorf(".github/workflows/ change should be CRITICAL, got %s", ec.Severity)
			}
		}
	}
	if !found {
		t.Error("CI workflow change not detected")
	}
}

func TestDetector_BinaryFile(t *testing.T) {
	binaryContent := "PNG\x89\x00\x00\x00binary data here"
	dir := setupEvilMergeWith(t, "image.png", binaryContent)

	d := detector.New()
	report, err := d.AnalyzeMerge(context.Background(), findLatestMerge(t, dir))
	if err != nil {
		t.Fatal(err)
	}
	if len(report.EvilChanges) == 0 {
		t.Fatal("expected evil changes for binary file modification")
	}
	for _, ec := range report.EvilChanges {
		if ec.FilePath == "image.png" {
			if ec.Severity != models.SeverityCritical {
				t.Errorf("binary evil change should be CRITICAL, got %s", ec.Severity)
			}
			if !strings.Contains(ec.Detail, "binary") {
				t.Errorf("detail should mention binary, got: %q", ec.Detail)
			}
		}
	}
}

func TestDetector_LongLine(t *testing.T) {
	// 600-char line — well above the 500-char threshold
	longLine := strings.Repeat("x", 600)
	evilContent := "const a = 1;\n" + longLine + "\nconst b = 2;"
	dir := setupEvilMergeWith(t, "app.js", evilContent)

	d := detector.New()
	report, err := d.AnalyzeMerge(context.Background(), findLatestMerge(t, dir))
	if err != nil {
		t.Fatal(err)
	}
	if len(report.EvilChanges) == 0 {
		t.Fatal("expected evil changes")
	}
	for _, ec := range report.EvilChanges {
		if ec.FilePath == "app.js" {
			if ec.Severity != models.SeverityCritical {
				t.Errorf("long line should escalate to CRITICAL, got %s", ec.Severity)
			}
			if !strings.Contains(ec.Detail, "chars") {
				t.Errorf("detail should mention line length in chars, got: %q", ec.Detail)
			}
		}
	}
}

func TestDetector_SuspiciousJSContent_Function(t *testing.T) {
	evilContent := `const x = new Function("return process.env")();`
	dir := setupEvilMergeWith(t, "utils.js", evilContent)

	d := detector.New()
	report, err := d.AnalyzeMerge(context.Background(), findLatestMerge(t, dir))
	if err != nil {
		t.Fatal(err)
	}
	if len(report.EvilChanges) == 0 {
		t.Fatal("expected evil changes")
	}
	for _, ec := range report.EvilChanges {
		if ec.FilePath == "utils.js" {
			if ec.Severity != models.SeverityCritical {
				t.Errorf("Function() constructor should escalate to CRITICAL, got %s", ec.Severity)
			}
			if !strings.Contains(ec.Detail, "Function()") {
				t.Errorf("detail should mention Function(), got: %q", ec.Detail)
			}
		}
	}
}

func TestDetector_SuspiciousJSContent_Eval(t *testing.T) {
	evilContent := `function load(code) { eval(code); }`
	dir := setupEvilMergeWith(t, "loader.ts", evilContent)

	d := detector.New()
	report, err := d.AnalyzeMerge(context.Background(), findLatestMerge(t, dir))
	if err != nil {
		t.Fatal(err)
	}
	for _, ec := range report.EvilChanges {
		if ec.FilePath == "loader.ts" {
			if ec.Severity != models.SeverityCritical {
				t.Errorf("eval() should escalate to CRITICAL, got %s", ec.Severity)
			}
			if !strings.Contains(ec.Detail, "eval()") {
				t.Errorf("detail should mention eval(), got: %q", ec.Detail)
			}
		}
	}
}

func TestDetector_SuspiciousContent_NoFalsePositive_NonJS(t *testing.T) {
	// eval( in a .py file should NOT trigger the JS-specific check
	evilContent := "result = eval(expression)"
	dir := setupEvilMergeWith(t, "script.py", evilContent)

	d := detector.New()
	report, err := d.AnalyzeMerge(context.Background(), findLatestMerge(t, dir))
	if err != nil {
		t.Fatal(err)
	}
	for _, ec := range report.EvilChanges {
		if ec.FilePath == "script.py" {
			if strings.Contains(ec.Detail, "eval()") {
				t.Error("eval() pattern should not trigger for .py files")
			}
		}
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
	if err := iter.ForEach(func(c *object.Commit) error {
		if c.NumParents() == 1 && nonMerge == nil {
			nonMerge = c
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	d := detector.New()
	_, err = d.AnalyzeMerge(context.Background(), nonMerge)
	if err == nil {
		t.Error("should return error for non-merge commit")
	}
}
