# Sprint 24: Quality Polish & Documentation Sync

## Goal
Align documentation with the Wails GUI architecture, expand frontend test coverage, and harden backend storage tests.

## Kanban

| ID | Ticket | Status | Priority | File Scope | DOD |
|----|--------|--------|----------|------------|-----|
| T-38 | **Doc Sync** | 🔲 TODO | High | `docs/*.md` | TUI references removed from Architecture and Standards docs. |
| T-39 | **Store Tests** | 🔲 TODO | Med | `src/stores/*.ts` | 80%+ coverage for Settings and Ollama stores. |
| T-40 | **UI Tests P1** | 🔲 TODO | Med | `src/pages/NetOps.tsx`, `TopBar.tsx` | Core interaction tests passing for TopBar and NetOps. |
| T-41 | **UI Tests P2** | 🔲 TODO | Low | `src/pages/*.tsx` | Coverage for SecOps, DevOps, AIOps, NetworkDesign. |
| T-42 | **Fix Flaky Test** | 🔲 TODO | High | `internal/common/storage_test.go` | `TestInsertLogAndQuery` passes 10/10 times in isolation and suite. |
| T-43 | **State Repair** | 🔲 TODO | High | `.memory/state.md` | State file reflects v1.3.0 GUI status, not stale v1.0.0 TUI goals. |

## Baseline Results
- **Go Tests**: 7/7 packages passed.
- **Frontend Tests**: 27/27 passed.
- **Lint**: 0 errors.
