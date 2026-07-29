# Changelog

All notable changes to Universal-Ops will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.4.2] - 2026-07-29

### Fixed
- Dependabot npm groups split into production/development — major bumps get individual PRs.
- `brace-expansion` + `minimatch` npm overrides to fix vitest coverage CJS compat and eliminate high vulns.
- `golangci-lint` config fully repaired (invalid rules, skip-dirs, exclusions).
- Compilation errors in `events.go` and `baselines.go` resolved.
- Docker container query returns empty slice instead of nil.
- CI/CD pipeline hardened end-to-end (build order, lint, test, audit).

### Changed
- Dependabot configuration: npm groups separated, `update_types: ["minor", "patch"]` for safe updates.
- `install.ps1` fallback version → 1.4.2, README badge → v1.4.2.

## [1.4.1] - 2026-07-27

### Fixed
- Various compilation errors in `events.go`, `baselines.go`, `App.go` (missing imports).

### Changed
- `linter` field added to golangci-lint config.
- SLO engine tests added.
- `HealthBadge` wired into Dashboard.

## [1.4.0] - 2026-07-25

### Fixed
- CI pipeline: removed invalid `shell: bash` from golangci-lint action step (blocked all runs since July 24).
- Version numbers synced across wails.json, install scripts, and README badge.

## [1.3.1] - Unreleased

### Fixed
- DevOps page tests updated to match current tab structure (PS/Bash instead of Terminal).
- CI Go version pinned to 1.26.x to match `go.mod` requirement.
- Version drift in metrics exporter health endpoint (1.3.0 → 1.3.1).
- Build artifact (`coverage`) removed from version control.
- Machine-specific data files (`data/baseline.json`, `data/.onboarded`) removed from version control.
- .gitignore expanded to prevent re-introduction of personal scripts, dead build scripts,
  internal planning docs, and redundant documentation.

### Added
- `SECURITY.md`: Security policy and vulnerability reporting guidelines.
- `codecov.yml`: Code coverage thresholds and CI integration config.

### Removed
- Stale developer scripts (`scripts/fix-path.ps1`, `scripts/build.sh`, etc.).
- Redundant internal docs (`docs/reviewprompt.md`, `docs/SEO_STRATEGY.md`, etc.).
- `.aiexclude` (patterns already covered by `.gitignore`).
