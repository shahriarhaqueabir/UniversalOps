# Sprint 11: Dead Code Audit Findings

## Audit Date
2026-07-08

## Tools Used
- `go build ./...` — clean
- `go vet ./...` — clean
- `go test ./internal/...` — all 9 packages pass
- `staticcheck ./...` — primary dead code detector
- `tsc --noEmit` — clean (frontend)
- `grep` — manual pattern scans for `_ = x`, TODOs, stubs, mock data
- Import graph tracing — verified `internal/ui/` has zero production callers

## Dead Code Findings (7 items)

| # | File | Line(s) | Symbol | Type | Severity | Action |
|---|------|---------|--------|------|----------|--------|
| 1 | `internal/app/Alerts.go` | 99-104 | `func alertTime` | Unused function | High | Remove |
| 2 | `internal/app/Logs.go` | 102-120 | `func parseLogLine` | Unused function | High | Remove |
| 3 | `internal/common/charts/charts.go` | 134-139 | `func min` | Unused function (Go 1.21 builtin available) | High | Remove |
| 4 | `internal/common/charts/charts.go` | 141-146 | `func max` | Unused function (Go 1.21 builtin available) | High | Remove |
| 5 | `internal/netops/view.go` | 474-484 | `func renderBar` | Duplicate (sysops/view.go:182 is used version) | High | Remove |
| 6 | `internal/ui/root.go` | 52 | `lastNetCounters` field | Unused struct field | Medium | Remove |
| 7 | `internal/ui/root.go` | 53 | `lastNetCapture` field | Unused struct field | Medium | Remove |

## Legacy/Unused Package: `internal/ui/`

- **Zero production imports**: Not referenced by `main.go`, `internal/app/`, or any other production Go file
- **Only referenced by**: Self-tests (`internal/ui/*_test.go`)
- **Status**: Legacy Bubble Tea TUI layer, superseded by Wails GUI
- **Recommendation**: Archive or add deprecation notice

## Stub Code (partially dead)

| # | File | Line(s) | Pattern | Action |
|---|------|---------|---------|--------|
| 8 | `internal/app/Alerts.go` | 77 | `_ = rules` — RemoveRule is a no-op stub | Remove stub or implement |
| 9 | `cmd/hawkward-gui/frontend/src/lib/mockData.ts` | all | 11 functions returning empty/zero state. Still used by ProcessManager.tsx for `mockProcesses()` | Trim unused exports |

## Staticcheck Lint Issues (5 items)

| # | File | Line | Code | Issue |
|---|------|------|------|-------|
| 10 | `internal/aiops/update.go` | 141 | ST1005 | Error string capitalized: "Ollama is not available..." |
| 11 | `internal/aiops/workflows.go` | 77 | ST1005 | Error string capitalized: "Ollama is not available" |
| 12 | `internal/secops/defender.go` | 60 | ST1005 | Error string capitalized: "Windows Defender is not available..." |
| 13 | `internal/common/common_test.go` | 219 | SA4006 | `origDir` set but never used |
| 14 | `internal/sysops/workflows.go` | 110 | S1039 | Unnecessary `fmt.Sprintf` (no format verbs) |

## Previously Cleaned
- `internal/netops/ping.go`: `_ = echo` dead code confirmed removed in prior sprint
- `internal/netops/workflows.go`: unused `cmds` variable confirmed removed in prior sprint
