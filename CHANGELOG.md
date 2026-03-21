# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- CLI utility `evilmerge` for detecting evil merges in Git repositories
- Detection algorithm comparing merge tree against merge-base and both parents
- Severity classification: CRITICAL, WARNING, INFO
- Sensitive file pattern detection (`.env`, `auth`, `crypto`, `password`, etc.)
- Text output with colored severity indicators
- JSON output for CI/CD integration
- Filtering by date (`--since`, `--until`), branch (`--branch`), severity (`--severity`)
- Commit limit (`--limit`) for large repositories
- `--fail-on` flag for CI pipelines (exit code 1 on findings)
- GoReleaser configuration for cross-platform builds
- GitHub Actions CI/CD pipeline (test, lint, release)
