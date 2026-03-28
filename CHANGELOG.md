# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.9] - 2026-03-28

### Security

- HTTP security headers middleware (X-Frame-Options, X-Content-Type-Options, HSTS, Referrer-Policy, CSP)
- Per-IP rate limiting (100 req/min, CF-Connecting-IP aware)
- OAuth state cookie SameSite=Lax; session cookie SameSite=Strict
- `crypto/rand` error checked in `randomState()` — predictable state prevented
- Graceful shutdown on SIGTERM/SIGINT with 30s drain; HTTP IdleTimeout=120s

### Fixed

- Goroutine leak in history scan — missing `wg.Wait()` could write to destroyed channel
- Marketplace webhook used `account.GetID()` (user ID) instead of installation ID — Pro plan was never applied
- Dashboard template rendered to `w` before error check; now uses `bytes.Buffer`
- `TotalMerges` counted the commit that triggered the scan limit
- Removed dead `worker.NewGitHubClient()`
- Webhook switch missing `default` — unknown event types now logged
- Annual pricing "Save 20%" showed $56 instead of $67

### Added

- Dashboard notice when GitHub App is not installed
- GDPR: Data Controller name in Privacy Policy; one-click consent revoke on /privacy
- Accessibility: focus trap on cookie banner, `aria-modal`, `aria-pressed`, `aria-hidden` on cursor span
- Tests for `session`, `oauth`, `dashboard` packages
- OG image at `/og-image.png`
- Landing page SEO: OG tags, Twitter Card, schema.org JSON-LD, canonical URL

### Changed

- Domain: `evilmerge.dev`
- Landing page redesigned with coming-soon overlay

## [0.1.5] - 2026-03-27

### Added

**CLI**
- `--since-tag` / `--until-tag` — scan commits between git tags instead of dates
- `--ignore-bots` — skip merge commits authored by known bots (dependabot, renovate, github-actions[bot], etc.)
- `--exclude` / `--include` — filter reported findings by glob pattern (supports `**`)
- `--output FILE` — write results to file instead of stdout
- `--workers N` — parallel analysis of merge commits
- `--verbose` — print per-commit progress to stderr for large repositories
- `.evilmerge.yml` — project-level config file (`fail-on`, `ignore-bots`, `exclude`, `include`, `output`)
- `.evilmerge-ignore` — allowlist file for commit hashes (7–40 hex chars) and author names/emails

**Detector**
- Long-line detection — flags lines pushed far right (>500 chars), CRITICAL severity; matches the horizontal obfuscation technique from the real incident
- Suspicious JS/TS content patterns — `eval()` and `Function()` constructor in `.js`/`.ts`/`.mjs`/`.cjs`/`.jsx`/`.tsx` files
- Expanded sensitive path patterns — build configs (`webpack.config.*`, `vite.config.*`, `rollup.config.*`), CI workflows (`.github/workflows/`)
- Binary file detection — null-byte check, reported as CRITICAL

**GitHub App**
- Status badge endpoint `GET /badge/{owner}/{repo}` — SVG badge showing `passing` / `N found` / `unknown`
- Slack and webhook notifications on evil merge findings (`SLACK_WEBHOOK_URL`, `NOTIFICATION_WEBHOOK_URL`)
- Scheduled full-history scan — runs on installation and every 30 days; detects past incidents in repos that connected after the fact

**Integrations**
- GitLab CI template (`examples/gitlab-ci.yml`)
- Bitbucket Pipelines template (`examples/bitbucket-pipelines.yml`)
- Pre-receive hook for self-hosted git servers (`examples/pre-receive`)

## [0.1.0] - 2026-03-23

### Added
- CLI utility `evilmerge` for detecting evil merges in Git repositories
- Detection algorithm comparing merge tree against merge-base and both parents
- Severity classification: CRITICAL, WARNING, INFO
- Sensitive file pattern detection (`.env`, `auth`, `crypto`, `password`, etc.)
- Text output with colored severity indicators
- JSON output for CI/CD integration
- SARIF 2.1.0 output (`--format=sarif`) for GitHub Code Scanning integration
- `--commit` flag for detailed single-commit inspection with line-level diffs
- `context.Context` propagation throughout scanner and detector with `--timeout` flag and Ctrl+C cancellation support
- Filtering by date (`--since`, `--until`), branch (`--branch`), severity (`--severity`)
- Commit limit (`--limit`) for large repositories
- `--fail-on` flag for CI pipelines (exit code 1 on findings)
- GitHub Action composite wrapper (`uses: fimskiy/Evil-merge-detector@v1`) with annotations, job summary, and SARIF upload support
- GoReleaser configuration for cross-platform builds (Linux, macOS, Windows — amd64/arm64)
- GitHub Actions CI/CD pipeline (test, lint, shellcheck, release)
