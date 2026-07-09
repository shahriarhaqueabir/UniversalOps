# Hawkward — Workspace Memory

## Active Session
- **Sprint 18: Release Hardening** — ✅ DONE

## Completed This Session

| ID | Status | Description |
|----|--------|-------------|
| C-11 | ✅ DONE | NetOps Enhancement Loop 1 — T-01 complete (full tests), Ping Jitter + Chart added, DNS custom resolver support added |
| C-12 | ✅ DONE | SecOps Enhancement Loop 2 — T-02 complete (full tests), Firewall high-risk flagging, External listener detection |
| C-13 | ✅ DONE | DevOps Enhancement Loop 3 — FileBrowser breadcrumbs + binary safety, Service sc-fallback, DevOps tests added |
| C-14 | ✅ DONE | SysOps Enhancement Loop 4 — Physical/Logical core detection, Load Saturation Index, Memory/Disk/System tests added |
| C-15 | ✅ DONE | AIOps Enhancement Loop 5 — Ollama model discovery, Chat suggested prompts, AIOps mock tests added |
| C-16 | ✅ DONE | T-03 complete — MinGW installed, `go test -race ./...` passed across all modules |
| C-17 | ✅ DONE | T-04, T-05, T-06 complete — Version bumped to v1.1.1, final verification passed, git tag v1.1.1 created |
| C-18 | ✅ DONE | Dashboard & Logs Enhancement Loop 6 — Dashboard drill-downs, App layer unit tests added |

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
