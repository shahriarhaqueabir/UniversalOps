# Sprint 11: Dead Code Removal & Static Cleanup

**Status**: ✅ Plan written — ready for execution
**Baseline**: `go build` clean, `go vet` clean, `go test ./internal/...` all pass, `staticcheck` identifies 7 dead code items + 5 lint issues

## Lane Matrix

| Ticket | Can run in parallel? | Write surface | Risk | Verification |
|--------|---------------------|---------------|------|-------------|
| T-01 | yes (isolated) | 4 packages | low | `go build`, `go vet`, tests |
| T-02 | after T-01 | internal/ui | low | `go build`, tests (test file removal) |
| T-03 | yes (isolated) | 2 files | low | `go build`, tests |
| T-04 | yes (isolated) | 5 files | low | `go build`, `staticcheck` |

## Tickets

| ID | Ticket | Status | Priority | DOD |
|----|--------|--------|----------|-----|
| T-01 | **Remove dead functions** — `alertTime`, `parseLogLine`, `min`/`max` (charts), `renderBar` (netops/view), `lastNetCounters`/`lastNetCapture` fields | 🔲 TODO | High | - [ ] `alertTime` removed from `internal/app/Alerts.go` - [ ] `parseLogLine` removed from `internal/app/Logs.go` - [ ] `min`/`max` removed from `internal/common/charts/charts.go` - [ ] `renderBar` removed from `internal/netops/view.go` - [ ] `lastNetCounters`/`lastNetCapture` removed from `internal/ui/root.go` - [ ] `go build ./...` passes - [ ] `go test ./internal/...` passes |
| T-02 | **Archive TUI layer** — entire `internal/ui/` package is unused (0 callers in production code, only self-tests) | 🔲 TODO | Medium | - [ ] Confirm no imports from `internal/ui` in any production Go file - [ ] Move to `internal/ui.legacy/` or add deprecation notice - [ ] `go build ./...` passes - [ ] `go test ./internal/...` passes |
| T-03 | **Clean up stubs** — `Alerts.RemoveRule` stub (`_ = rules` pattern), `mockData.ts` fallback functions | 🔲 TODO | Low | - [ ] Remove `_ = rules` from `Alerts.go` (or implement RemoveRule properly) - [ ] Audit `mockData.ts` usage and trim unused exports - [ ] `go build ./...` passes - [ ] `npm run build` passes |
| T-04 | **Fix staticcheck lint issues** — 3 ST1005 (capitalized errors), 1 SA4006 (unused var), 1 S1039 (unnecessary `fmt.Sprintf`) | 🔲 TODO | Low | - [ ] Fix 3 error strings to lowercase - [ ] Fix `origDir` in `common_test.go` - [ ] Replace bare `fmt.Sprintf` in `sysops/workflows.go` - [ ] `staticcheck ./...` returns 0 findings |

## Implementation Order

```
T-01 ──> T-02
  │
  ├──> T-03 (independent, parallel)
  └──> T-04 (independent, parallel)
```

T-01 first (high priority dead code removal), T-02 after (depends on T-01 clean state). T-03 and T-04 can run in parallel — they touch non-overlapping files.
