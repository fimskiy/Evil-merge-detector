# Evil Merge Detector

[![CI](https://github.com/evilmerge-dev/Evil-merge-detector/actions/workflows/ci.yml/badge.svg)](https://github.com/evilmerge-dev/Evil-merge-detector/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/evilmerge-dev/evil-merge-detector)](https://goreportcard.com/report/github.com/evilmerge-dev/evil-merge-detector)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Buy Me a Coffee](https://img.shields.io/badge/Buy%20Me%20a%20Coffee-donate-FFDD00?logo=buymeacoffee&logoColor=black)](https://www.buymeacoffee.com/fimskiy)

Detect **evil merges** in Git repositories — merge commits that contain changes beyond conflict resolution, invisible to code review.

## What is an evil merge?

An evil merge is a merge commit that introduces changes not present in either parent branch. These changes:

- **Bypass code review** — reviewers see the branch diff, not the merge commit itself
- **Are hard to trace** — `git blame` points to the merge, not a meaningful commit
- **Can hide malicious code** — a common supply chain attack vector

```
  feature ──●──────────────────────●── main
             \                    /
              ●── evil change ───● ← merge commit contains extra code
             /
  main ─────●
```

## Installation

**Homebrew:**
```sh
brew install fimskiy/tap/evilmerge
```

**Go:**
```sh
go install github.com/evilmerge-dev/evil-merge-detector/cmd/evilmerge@latest
```

**Binary:** download from [Releases](https://github.com/evilmerge-dev/Evil-merge-detector/releases)

## Usage

```sh
# Scan current repository
evilmerge scan

# Scan a specific path
evilmerge scan /path/to/repo

# Scan a specific branch since a date
evilmerge scan --branch=main --since=2024-01-01

# Only show critical findings
evilmerge scan --severity=critical

# JSON output for scripting
evilmerge scan --format=json

# Exit with code 1 if any warnings found (for CI)
evilmerge scan --fail-on=warning

# Detailed inspection of a specific merge commit (with line diffs)
evilmerge scan --commit=a1b2c3d

# Limit scan to 60 seconds (useful for very large repositories)
evilmerge scan --timeout=60s
```

**Example output:**

```
Evil Merge Detector v0.1.9
Scanning repository: /path/to/repo (branch: main)
Analyzed 142 merge commits, found 2 evil merges (in 340ms)

SEVERITY    COMMIT                                              AUTHOR                     FILES
----------------------------------------------------------------------------------------------------------
CRITICAL    a1b2c3d Merge branch 'feature/auth'                dev@company.com            config.py
WARNING     d4e5f6a Merge branch 'hotfix/payment'              dev@company.com            utils.js

Re-run with --format=json for full details on each merge.
```

## Severity levels

| Severity | Meaning |
|----------|---------|
| **CRITICAL** | File unchanged in both branches but modified in merge; new file added in merge; change in sensitive file (`.env`, `auth`, `crypto`, etc.) |
| **WARNING** | File changed in one branch, but merge result differs from both parents |
| **INFO** | File changed in both branches (conflict zone) — likely legitimate conflict resolution, worth reviewing |

## GitHub App

The easiest way to get started — install once and every pull request in your organization is automatically scanned. No workflow changes needed.

**[Install from GitHub Marketplace →](https://github.com/marketplace/evil-merge-detector)**

- Automatic scan on every PR via GitHub Checks
- Results appear directly in the pull request
- Full history scan when you first install — catches past incidents
- Scan history available at [evilmerge.dev/dashboard](https://evilmerge.dev/dashboard)
- Slack / webhook notifications when evil merges are found

**Plans:** Free (public repos, 50 scans/month) · Pro $7/month (private repos, unlimited scans, dashboard)

**Status badge** — add to your README:

```markdown
![Evil Merges](https://evilmerge.dev/badge/owner/repo)
```

Shows `passing` (green) or `N found` (red) based on the latest scan result.

---

## GitHub Action

Add to your workflow to automatically check PRs for evil merges:

```yaml
name: Evil Merge Check
on:
  pull_request:
    types: [opened, synchronize]

jobs:
  check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      - uses: evilmerge-dev/Evil-merge-detector@v1
        with:
          fail-on: warning
```

### Action inputs

| Input | Description | Default |
|-------|-------------|---------|
| `fail-on` | Minimum severity to fail the check (`info`/`warning`/`critical`) | `warning` |
| `severity` | Minimum severity to report | `info` |
| `branch` | Branch to scan (auto-detected from PR by default) | — |
| `since` | Only scan merges after this date (YYYY-MM-DD) | — |
| `version` | Version of evilmerge to use | `latest` |
| `upload-sarif` | Upload SARIF results to Code Scanning | `false` |

### Action outputs

| Output | Description |
|--------|-------------|
| `evil-merges-found` | `true` if evil merges were found |
| `evil-merge-count` | Number of evil merge commits found |

### With SARIF upload

```yaml
- uses: evilmerge-dev/Evil-merge-detector@v1
  with:
    fail-on: warning
    upload-sarif: true
```

Findings will appear in **Security → Code scanning alerts** with severity, affected file, and commit fingerprint.

## Other CI integrations

Ready-to-use templates are in [`examples/`](examples/):

| Platform | File |
|----------|------|
| GitLab CI | [`examples/gitlab-ci.yml`](examples/gitlab-ci.yml) |
| Bitbucket Pipelines | [`examples/bitbucket-pipelines.yml`](examples/bitbucket-pipelines.yml) |
| Self-hosted git (pre-receive hook) | [`examples/pre-receive`](examples/pre-receive) |

## CI/CD Integration (manual)

If you prefer to install the binary directly:

```yaml
- name: Check for evil merges
  run: |
    go install github.com/evilmerge-dev/evil-merge-detector/cmd/evilmerge@latest
    evilmerge scan --branch=main --fail-on=warning
```

## Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--branch` | Branch to scan | current HEAD |
| `--since` | Scan commits after date (YYYY-MM-DD) | — |
| `--until` | Scan commits before date (YYYY-MM-DD) | — |
| `--since-tag` | Scan commits after this tag | — |
| `--until-tag` | Scan commits before this tag | — |
| `--severity` | Minimum severity to report: `info`, `warning`, `critical` | `info` |
| `--limit` | Max merge commits to analyze | unlimited |
| `--format` | Output format: `text`, `json`, `sarif` | `text` |
| `--fail-on` | Exit code 1 if findings at or above severity | — |
| `--commit` | Detailed inspection of a specific merge commit (hash) | — |
| `--timeout` | Maximum scan duration, e.g. `30s`, `5m` | unlimited |
| `--ignore-bots` | Skip merges by known bots (dependabot, renovate, etc.) | false |
| `--exclude` | Exclude findings in files matching this glob (repeatable) | — |
| `--include` | Only report findings in files matching this glob (repeatable) | — |
| `--output` | Write results to file instead of stdout | — |
| `--workers` | Number of parallel workers for merge analysis | 1 |
| `--verbose` | Print each analyzed commit to stderr | false |

## Configuration

Project-level defaults live in `.evilmerge.yml` at the repository root. CLI flags override file values; `exclude`/`include` patterns are additive.

```yaml
# .evilmerge.yml
fail-on: warning
ignore-bots: true
exclude:
  - "*.lock"
  - "dist/**"
include:
  - "src/**"
output: results.sarif
```

### Ignore list

Create `.evilmerge-ignore` to permanently whitelist specific commits or authors:

```
# Lines starting with # are comments

# Commit hashes (7–40 hex chars, short prefix works)
abc1234
deadbeef12345678

# Author names or emails
legacy-bot@company.com
Merge Bot
```

Anything that looks like a hex string (≥7 chars) is treated as a commit hash prefix; everything else is an author name or email.

## How it works

For each merge commit **M** with parents **P1** and **P2**:

1. Find the merge base **B** = common ancestor of P1 and P2
2. For each file in M, compare its content against B, P1, and P2
3. Flag any file where M's content cannot be explained by either parent

## License

MIT — see [LICENSE](LICENSE)
