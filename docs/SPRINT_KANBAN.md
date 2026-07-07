# Hawkward Local Sprint Kanban — v3 Overhaul

This board tracks the current local multi-agent sprint for the v3 overhaul. The workspace has no Git remote, so "merge" means integrated into this working tree and verified locally.

## Sprint 1: Visualization Engine

| ID | Title | Owner | State | Scope | Merge Gate |
| --- | --- | --- | --- | --- | --- |
| V3-001 | Chart library (Line, Bar, Gauge, Sparkline, Area, HeatMap, Numeric) | Charts Engineer | **Planned** | `internal/common/charts/*` (8 files) | `go test ./internal/common/charts/...` |
| V3-002 | Time-series data store | Charts Engineer | **Planned** | `internal/common/timeseries.go` | `go test ./internal/common/... -run TimeSeries` |
| V3-003 | Forecast engine | Charts Engineer | **Planned** | `internal/common/forecast.go` | `go test ./internal/common/... -run Forecast` |
| V3-004 | Card component system | UI Architect | **Planned** | `internal/ui/cards.go` | `go test ./internal/ui/... -run Card` |
| V3-005 | YAML config system | Config Engineer | **Planned** | `internal/common/config.go` | `go build ./...` |

## Board

| ID | Title | Owner | State | Scope | Merge Gate |
| --- | --- | --- | --- | --- | --- |
| NET-001 | Network bandwidth sparklines | Worker NetOps | Running | `internal/netops/*` | `go test ./internal/netops` and full suite pass |
| AI-001 | Natural-language system queries | Worker AI/SysOps | Running | `internal/aiops/*`, `internal/sysops/*` | focused AI/SysOps tests and full suite pass |
| AI-002 | Metrics anomaly detection | Worker AI/SysOps | Running | `internal/aiops/*`, `internal/sysops/*` | deterministic anomaly tests and full suite pass |
| UX-001 | Theme and color customization | Worker UX/Common | Running | `internal/common/*`, `internal/ui/*` | UI/common tests and full suite pass |
| OPS-001 | Session logging | Worker UX/Common | Running | `internal/common/*`, `internal/ui/*` | logging tests and full suite pass |
| REL-001 | Local release automation | Coordinator | Review | `scripts/*`, `docs/*`, package templates | release scripts lint/read cleanly, full suite pass |
| QA-001 | Broader test coverage | Coordinator + all workers | Running | touched packages | coverage improves without brittle live-system tests |

## Working Agreements

- One owner per card.
- Workers keep file ownership narrow and do not revert unrelated local edits.
- Every card needs tests or a clear reason tests are not practical.
- Integration requires `gofmt`, `go test ./...`, `go build ./...`, and a coverage readout.
- Local release artifacts go under `dist/` and are not source-controlled by default.

## v3 Planning Docs

| Document | Location | Purpose |
|----------|----------|---------|
| Overhaul plan (definitive) | `.memory/topics/hawkward-overhaul-plan.md` | 5-sprint plan with concrete file changes, API designs |
| Executive summary | `plans/overhaul-v3-summary.md` | 1-page sprint/team/architecture overview |
| Card system design | `.memory/topics/hawkward-card-system.md` | Card types, layout engine, states, interaction |
| Command palette design | `.memory/topics/hawkward-command-palette.md` | Search algorithm, key bindings, operation registry |
| Architecture doc | `docs/ARCHITECTURE.md` | Updated component tree, data flow, directory layout |
| Roadmap | `docs/ROADMAP.md` | Phase 7 (v3) with sprint breakdown and status |

## Control Notes

- Active agents: NetOps bandwidth, AI/SysOps query/anomaly, UX/Common theme/session logging.
- Coordinator owns release automation and final verification.
- v3 Sprint 1 is planned but not yet started — awaiting completion of current active tasks.
- Blockers should be written here with owner and next action.
