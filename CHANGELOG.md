# Changelog

All notable changes to Universal-Ops will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
