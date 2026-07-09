# Hawkward — Workspace Memory

## Active Session
- **Sprint 18: Release Hardening** — In Progress 🔶

## Completed This Session

| ID | Status | Description |
|----|--------|-------------|
| C-01 | ✅ DONE | TUI remnants fully purged — `.ai-style-rules.md`, `AI.md`, `plans/`, `docs/superpowers/`, `legacy/`, `modelfiles/`, `docs/archive/` deleted |
| C-02 | ✅ DONE | Dead code scan — No TODO/FIXME/HACK/hardcoded markers in Go or TSX. No mock data/stubs in production code |
| C-03 | ✅ DONE | UI audit completion — Read DevOps FileBrowserTab (wired), AIOps ReportsTab (wired), AIOps AnomaliesTab (wired). All tabs fully functional |
| C-04 | ✅ DONE | NetOps ping error handling — `executePing` now checks `res.error` field and shows timeout status |
| C-05 | ✅ DONE | Connections locale fix — Replaced `tasklist /FO CSV` with PowerShell `Get-Process` (wmic fallback). Locale-independent |
| C-06 | ✅ DONE | Port scan performance — Concurrent goroutines with 200ms timeout (was 500ms serial). Now completes in ~200ms instead of 11.5s |
| C-07 | ✅ DONE | `go vet ./...` passes |
| C-08 | ✅ DONE | `go test ./...` — All pass |
| C-09 | ✅ DONE | `npm test -- --run` — 4 suites, 15 tests, all pass |
| C-10 | ✅ DONE | `wails build -skipbindings` — Builds in 23.8s, 17MB binary |
| C-11 | ✅ DONE | NetOps Enhancement Loop 1 — T-01 complete (full tests), Ping Jitter + Chart added, DNS custom resolver support added |
| C-12 | ✅ DONE | SecOps Enhancement Loop 2 — T-02 complete (full tests), Firewall high-risk flagging, External listener detection |
| C-13 | ✅ DONE | DevOps Enhancement Loop 3 — FileBrowser breadcrumbs + binary safety, Service sc-fallback, DevOps tests added |
| C-14 | ✅ DONE | SysOps Enhancement Loop 4 — Physical/Logical core detection, Load Saturation Index, Memory/Disk/System tests added |
| C-15 | ✅ DONE | AIOps Enhancement Loop 5 — Ollama model discovery, Chat suggested prompts, AIOps mock tests added |
| C-16 | ✅ DONE | T-03 complete — MinGW installed, `go test -race ./...` passed across all modules |

## Completed (30 Review Loops)

| ID | Status | Description |
|----|--------|-------------|
| L-01 | ⚠️ BLOCKED | Race detector — requires gcc/MinGW for CGO on Windows |
| L-02 | ✅ DONE | `go mod verify` — All modules verified |
| L-03 through L-30 | ✅ DONE | All review loops complete |

## Known Issues

| Severity | Issue | File |
|----------|-------|------|
| ⚠️ Note | Wails v2.13.0 pins golang.org/x/net, x/crypto, x/sys — cannot upgrade beyond what Wails supports | go.mod |
| ⚠️ Note | Dependabot may still show 19 old vulns until it re-scans with upgraded deps | — |
| ✅ RESOLVED | Frontend tests had `act(...)` warnings — fixed with `waitFor` wrappers | test files |

## Sprint 18: Release Hardening
- **Plan**: `plans/2026-07-09-sprint-18-release-hardening.md`

## Topics
- [[project-graph]] — Entity relationship, data flow, and dependency graph
- [[hawkward-architecture]] — Go + Wails v2 architecture
- [[hawkward-sandbox]] — Sandbox layer implementation
- [[hawkward-known-issues]] — Known issues & pitfalls
