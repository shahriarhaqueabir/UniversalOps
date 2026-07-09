# Hawkward — Workspace Memory

## Active Session
- **Sprint 16: v1.1.0 Release Readiness** — Active 🚧

## Completed Tasks (Sprint 16)

| ID | Status | Description |
|----|--------|-------------|
| N-01 | ✅ DONE | Ping: Skip raw ICMP on Windows — direct to ping.exe, fixed regex for Windows output format |
| N-02 | ✅ DONE | DNS: Added system resolver fallback when public DNS servers are blocked |
| N-03 | ✅ DONE | Connections: Resolve process names from PIDs via tasklist on Windows |

## Completed Tasks (Sprint 15 — Archive & Consolidation)

| ID | Status | Description |
|----|--------|-------------|
| F-01 | ✅ DONE | Feature complete — all planned ops pages implemented |
| F-02 | ✅ DONE | Full test suite — Go backend + frontend tests passing |
| F-03 | ✅ DONE | Release pipeline — GitHub Actions + NSIS installer |
| F-04 | ✅ DONE | Documentation — README, memory topics, architecture docs |
| C-02 | ✅ DONE | Dead code removed — TUI remnants, unused functions, stale deps |
| C-03 | ✅ DONE | TUI-era docs cleaned up — SPRINT_KANBAN, ONBOARDING, ROADMAP, archive files removed |

## Known Issues

| Severity | Issue | File |
|----------|-------|------|
| ⚠️ Note | Wails v2.13.0 pins golang.org/x/net, x/crypto, x/sys — cannot upgrade without breaking compat | go.mod |

## Topics
- [[project-graph]] — Entity relationship, data flow, and dependency graph
- [[hawkward-architecture]] — Go + Wails v2 architecture
- [[hawkward-sandbox]] — Sandbox layer implementation
- [[hawkward-known-issues]] — Known issues & pitfalls
