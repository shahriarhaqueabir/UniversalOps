# Sprint 1: Polish and Coverage

## Board

| ID | Ticket | Status | Priority | Scope | DOD |
|----|--------|--------|----------|-------|-----|
| T-01 | **NET-001: Bandwidth Sparklines** | ✅ DONE | High | `internal/netops/` | - [x] Calculate bandwidth rates without `time.Sleep` blocking.<br>- [x] Render RX/TX sparklines in `netops` view.<br>- [x] Verify with unit tests. |
| T-02 | **UX-001: UI Theme Switching** | ✅ DONE | Medium | `internal/ui/`, `internal/common/` | - [x] Add keyboard shortcut (e.g., `T`) to cycle themes.<br>- [x] Display current theme in status bar or help.<br>- [x] Persist theme selection (if possible/requested). |
| T-03 | **QA-001: System/Net/DevOps Coverage** | ✅ DONE | High | `internal/sysops/`, `internal/netops/`, `internal/devops/` | - [x] Increase `sysops` coverage to >30%.<br>- [x] Increase `netops` coverage to >30%.<br>- [x] Increase `devops` coverage to >30%. |
| T-04 | **AI-001: Refine State Queries** | ✅ DONE | Medium | `internal/aiops/` | - [x] Add more query patterns for security and devops.<br>- [x] Improve anomaly detection logic based on more history. |
| T-05 | **REL-001: Verify Release Scripts** | ✅ DONE | Low | `scripts/` | - [x] Run `build.sh`/`build.bat` and verify binary output.<br>- [x] Verify `release.sh` logic. |

## Working Agreements
- Follow TEA pattern.
- Delegate to `handleKeyPress`.
- Use `common` styles.
- Unit tests for all logic.
