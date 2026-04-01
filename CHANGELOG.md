# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.9] - 2026-03-28

### Security

**Server**
- HTTP security headers middleware — X-Frame-Options, X-Content-Type-Options, HSTS, Referrer-Policy, Content-Security-Policy
- Per-IP rate limiting — fixed-window 100 req/min, CF-Connecting-IP aware, background cleanup
- Graceful shutdown on SIGTERM/SIGINT with 30s drain; HTTP IdleTimeout=120s

**Auth**
- OAuth state cookie: explicit SameSite=Lax; session cookie: SameSite=Strict in both Set and Clear
- `crypto/rand` error now checked in `randomState()` — prevented predictable zero-byte state token

### Fixed

**GitHub App**
- Goroutine leak in history scan — missing `wg.Wait()` could write to a destroyed semaphore channel
- Marketplace webhook used `account.GetID()` (GitHub user ID) instead of installation ID — Pro plan was never applied to the correct record
- Webhook event switch missing `default` — unknown event types now logged instead of silently dropped

**Dashboard**
- Template rendered directly to `w` before error check; switched to `bytes.Buffer` — partial HTML with status 200 on error is no longer possible

**Scanner**
- `TotalMerges` counter incremented before limit check — last unanalyzed commit was included in the count; moved after the check

**Landing**
- Annual pricing "Save 20%" displayed $56 (33% off) instead of $67 (correct 20% off $84/yr)

**Dead code**
- Removed unused exported `worker.NewGitHubClient()` and its orphaned import

### Added

**GitHub App**
- Dashboard shows an "App not installed" notice with install link when `InstallationID == 0`

**GDPR**
- Privacy Policy now identifies the Data Controller by name (GDPR Art. 13(1)(a))
- One-click "Revoke cookie consent" button on `/privacy` (GDPR Art. 7(3))

**Accessibility**
- Cookie banner captures keyboard focus on show via `requestAnimationFrame`
- `aria-modal="true"` on cookie dialog, `aria-pressed` on billing toggle, `aria-hidden="true"` on decorative cursor span

**Tests**
- New test packages: `session` (round-trip, tampered payload, Clear), `oauth` (randomState uniqueness, cookie attrs, invalid state), `dashboard` (redirect, render, wrong secret)

**Landing**
- OG image served at `/og-image.png` via embed
- SEO: Open Graph tags, Twitter Card, schema.org JSON-LD, canonical URL, improved H1

### Changed

- Domain switched to `evilmerge.dev`
- Landing page redesigned: minimal variant with coming-soon overlay, improved typography and accessibility

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
- GitHub Action composite wrapper (`uses: evilmerge-dev/Evil-merge-detector@v1`) with annotations, job summary, and SARIF upload support
- GoReleaser configuration for cross-platform builds (Linux, macOS, Windows — amd64/arm64)
- GitHub Actions CI/CD pipeline (test, lint, shellcheck, release)
