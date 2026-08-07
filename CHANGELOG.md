# Changelog

All notable changes to Universal-Ops will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.6.0] - 2026-08-06

### Added
- **AI Sensory Awareness**: Hawk now has passive knowledge of active network interfaces (e.g. Teredo, VPNs, Wi-Fi) in every prompt.
- **DevOps MCP Tools**: Added `get_docker_summary`, `get_k8s_status`, `get_k8s_pods`, and `get_k8s_events` to Hawk's toolkit.
- **Network Discovery Tool**: Added `get_network_interfaces` tool for Hawk to inspect detailed adapter stats and rates.
- **Self-Diagnosis**: Renamed `query_logs` to `get_app_logs` and enhanced instructions for Hawk to use app logs for self-troubleshooting.
- Observation-layer UX upgrades: persistent sidebar collapse, denser alerts, clearer dashboard metadata, and improved collector/settings summaries.
- Searchability and metadata improvements: page metadata, document metadata, and SEO-friendly app tags.
- Code-splitting and bundle-size work: centralized lazy page registry, suggested route preloading, and removal of the frontend `date-fns` dependency.

### Fixed
- **Version Drift**: Consolidated version numbering into `internal/common/version.go` (v1.6.0).
- Release surface consistency: version metadata, installers, and app packaging now align on `1.6.0`.
- Frontend utility date formatting now uses local helpers, reducing bundle weight and removing an unnecessary dependency.

## [1.5.0] - 2026-07-31

### Added
- AI onboarding test coverage: `AIOps_setup_test.go` (18 tests — RAM-based model fallback, setup recommendation structure/logic, trivial-message classification, context cache TTL, modelfile presence, Ollama persona setup).
- Port scan hardening: max 1024 ports cap, range validation, dedup + fuzz tests.
- SLO tests and topology persistence (`feat(app)` hardening batch).
- Typed Wails bridge: `mockQueryReturn<T>()` removed 51 `as any` casts; stable list keys across 33 frontend files.

### Fixed
- gopsutil v4.26.6 (DiskIO IOCTL failure on FAT32/Google Drive), Prometheus client v1.24.1, Wails restored to v2.13.0.
- GC + FreeOSMemory, workflow timeout guard, AIOps request-context hardening.
- CI/CD pipeline: toolchains pinned (`wails@v2.13.0`, `golangci-lint@v1.64.8` — `.golangci.yml` is v1 config and `@latest` drifted to v2); Scoop manifest now gets the **real SHA256** (placeholder hashes would fail every install); version handling standardized (assets/URLs no longer split between `v1.5.0` and `1.5.0`); broken Homebrew formula removed from CI (macOS binaries aren't built there); `checksums.txt` no longer includes `.sha256` sidecars; `package-*.ps1` un-ignored so CI checkout contains them.
- `install.ps1` fallback URL matched CI asset naming (`universal-ops-<ver>-windows-amd64.exe`).
- TypeScript build errors resolved (`tsc -b`): onboarded result cast in `App.tsx`, optional `EventsOff` handler in `useBackend.ts`, `isPreviousData` dropped for TanStack Query v5. CI TypeScript check switched from `tsc --noEmit` (false green on the project-references root) to `tsc -b`.

### Changed
- Version bumped to 1.5.0 (wails.json, package.json, install.ps1, README badge).

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
