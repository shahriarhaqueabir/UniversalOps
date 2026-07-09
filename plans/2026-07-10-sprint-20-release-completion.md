# Sprint 20: v1.2.0 Release Completion & Pipeline Hardening

**Date**: 2026-07-10
**Status**: 🔄 IN PROGRESS

## Kanban

| ID | Ticket | Status | Priority | DOD |
|----|--------|--------|----------|-----|
| T-01 | **Commit uncommitted changes + clean git tracking** | ✅ DONE | High | `git status` shows clean working tree; `.gitignore` updated if needed |
| T-02 | **Fix CI test pipeline (`test.yml`)** | ✅ DONE | High | `go vet ./...` no longer fails in CI; frontend build step added |
| T-03 | **Verify/Create v1.2.0 GitHub Release** | 🔄 IN PROGRESS | High | Release exists at `github.com/shahriarhaqueabir/AllOpsFull/releases/tag/v1.2.0` with Windows/Linux/macOS artifacts |
| T-04 | **Fix `release.yml` Linux webkit dependency** | ✅ DONE | Medium | Added webkit2gtk package fallback (4.1 → 4.0) |
| T-05 | **Full dependency audit & update** | ✅ DONE | Medium | `govulncheck` clean, `npm audit` clean, all deps at latest |
| T-06 | **TUI remnants deep verification** | ✅ DONE | Medium | Grep for Bubble Tea/Lip Gloss/TUI: no matches found |
| T-07 | **Document remaining known issues** | 🔲 TODO | Low | Update `.memory/` with current state |

## Verification Results

| Check | Result |
|------|--------|
| Frontend tests | ✅ 28/28 pass |
| Go tests (`-race`) | ✅ All 7 packages pass |
| `go vet ./...` | ✅ Clean |
| `tsc --noEmit` | ✅ Clean |
| `npm audit` | ✅ 0 vulnerabilities |
| `govulncheck` | ✅ No vulnerabilities found |
| `go mod tidy` | ✅ Clean |

## Dependabot Notes

- 14 alerts shown on push (7 critical, 2 high, 5 moderate)
- All in `golang.org/x/crypto` (indirect dep, v0.54.0 = latest available)
- Stale scan from 2 days ago — Dependabot re-scans after push
- Actual security posture: **clean** per `govulncheck`

## Commits

- `f010014` — CI pipeline hardening + webkit fallback + version bump
- `ae44eaf` — Sprint 19 final: memory updates, UI audit fixes
