# Hawkward Local Sprint Kanban

This board tracks the current local multi-agent sprint. The workspace has no Git remote, so "merge" means integrated into this working tree and verified locally.

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

## Control Notes

- Active agents: NetOps bandwidth, AI/SysOps query/anomaly, UX/Common theme/session logging.
- Coordinator owns release automation and final verification.
- Blockers should be written here with owner and next action.
