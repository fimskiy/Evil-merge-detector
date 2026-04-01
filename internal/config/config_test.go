package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/evilmerge-dev/evil-merge-detector/internal/config"
)

func TestLoad_NoFile(t *testing.T) {
	dir := t.TempDir()
	cfg, err := config.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.FailOn != "" || cfg.IgnoreBots || len(cfg.Exclude) != 0 || len(cfg.Include) != 0 || cfg.Output != "" {
		t.Errorf("expected zero-value config for missing file, got %+v", cfg)
	}
}

func TestLoad_Valid(t *testing.T) {
	dir := t.TempDir()
	content := `
fail-on: warning
ignore-bots: true
exclude:
  - "*.lock"
  - "dist/**"
include:
  - "src/**"
output: results.sarif
`
	write(t, dir, ".evilmerge.yml", content)

	cfg, err := config.Load(dir)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.FailOn != "warning" {
		t.Errorf("fail-on: got %q, want %q", cfg.FailOn, "warning")
	}
	if !cfg.IgnoreBots {
		t.Error("ignore-bots: expected true")
	}
	if len(cfg.Exclude) != 2 {
		t.Errorf("exclude: got %d patterns, want 2", len(cfg.Exclude))
	}
	if len(cfg.Include) != 1 || cfg.Include[0] != "src/**" {
		t.Errorf("include: got %v, want [src/**]", cfg.Include)
	}
	if cfg.Output != "results.sarif" {
		t.Errorf("output: got %q, want %q", cfg.Output, "results.sarif")
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, ".evilmerge.yml", "{ invalid: yaml :::")
	_, err := config.Load(dir)
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestLoadIgnore_NoFile(t *testing.T) {
	dir := t.TempDir()
	ig, err := config.LoadIgnore(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(ig.Hashes) != 0 || len(ig.Authors) != 0 {
		t.Errorf("expected empty ignore for missing file, got %+v", ig)
	}
}

func TestLoadIgnore_Parsing(t *testing.T) {
	dir := t.TempDir()
	content := `
# a comment

abc1234
DEADBEEF12345678
evil@contractor.com
Evil Contractor
`
	write(t, dir, ".evilmerge-ignore", content)

	ig, err := config.LoadIgnore(dir)
	if err != nil {
		t.Fatal(err)
	}

	if len(ig.Hashes) != 2 {
		t.Fatalf("hashes: got %d, want 2 — %v", len(ig.Hashes), ig.Hashes)
	}
	if len(ig.Authors) != 2 {
		t.Fatalf("authors: got %d, want 2 — %v", len(ig.Authors), ig.Authors)
	}
	// hashes must be stored lowercase
	for _, h := range ig.Hashes {
		for _, r := range h {
			if r >= 'A' && r <= 'F' {
				t.Errorf("hash %q contains uppercase", h)
			}
		}
	}
}

func TestLoadIgnore_TooShortHex(t *testing.T) {
	// 6-char hex — below 7-char threshold, should be treated as author not hash
	dir := t.TempDir()
	write(t, dir, ".evilmerge-ignore", "abc123\n")

	ig, err := config.LoadIgnore(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(ig.Hashes) != 0 {
		t.Error("6-char hex should not be treated as a hash")
	}
	if len(ig.Authors) != 1 {
		t.Error("6-char hex should be treated as an author")
	}
}

func write(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}
