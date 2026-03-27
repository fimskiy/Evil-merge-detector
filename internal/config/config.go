package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config holds the content of .evilmerge.yml.
type Config struct {
	FailOn     string   `yaml:"fail-on"`
	IgnoreBots bool     `yaml:"ignore-bots"`
	Exclude    []string `yaml:"exclude"`
	Include    []string `yaml:"include"`
	Output     string   `yaml:"output"`
}

// Ignore holds the parsed .evilmerge-ignore file.
type Ignore struct {
	Hashes  []string // commit hash prefixes to skip (7–40 hex chars)
	Authors []string // author names or emails to skip
}

// Load reads .evilmerge.yml from repoPath. Returns empty config if file not found.
func Load(repoPath string) (*Config, error) {
	if repoPath == "" {
		repoPath = "."
	}
	data, err := os.ReadFile(filepath.Join(repoPath, ".evilmerge.yml"))
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, fmt.Errorf("reading .evilmerge.yml: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing .evilmerge.yml: %w", err)
	}
	return &cfg, nil
}

// LoadIgnore reads .evilmerge-ignore from repoPath.
// Lines starting with '#' are comments.
// Lines of 7–40 hex chars are treated as commit hash prefixes.
// All other lines are author names or emails.
func LoadIgnore(repoPath string) (*Ignore, error) {
	if repoPath == "" {
		repoPath = "."
	}
	f, err := os.Open(filepath.Join(repoPath, ".evilmerge-ignore"))
	if err != nil {
		if os.IsNotExist(err) {
			return &Ignore{}, nil
		}
		return nil, fmt.Errorf("reading .evilmerge-ignore: %w", err)
	}
	defer func() { _ = f.Close() }()

	var ig Ignore
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if len(line) >= 7 && len(line) <= 40 && isHexString(line) {
			ig.Hashes = append(ig.Hashes, strings.ToLower(line))
		} else {
			ig.Authors = append(ig.Authors, line)
		}
	}
	return &ig, sc.Err()
}

func isHexString(s string) bool {
	for _, r := range s {
		if (r < '0' || r > '9') && (r < 'a' || r > 'f') && (r < 'A' || r > 'F') {
			return false
		}
	}
	return true
}
