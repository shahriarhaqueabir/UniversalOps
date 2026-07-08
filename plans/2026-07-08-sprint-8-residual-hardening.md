# Sprint 8.2: Residual Hardening (DONE)

> **Created**: 2026-07-08 | **Completed**: 2026-07-08
> **Result**: All 4 tickets DONE. All tests pass, go vet clean, frontend builds.

---

## Completed Tickets

| ID | Ticket | Status | Priority | What Changed | Verdict |
|----|--------|--------|----------|-------------|---------|
| R-01 | **Windows Sandbox via Job Objects** | DONE | High | `sandbox_windows.go`: Real sandbox with restricted token (29 privileges disabled), Job Object (1 process limit, 100MB memory, kill-on-close), background assigner goroutine | go test: PASS, go vet: clean |
| R-02 | **Server-Side Dangerous Command Detection** | DONE | High | `shell.go`: 18-pattern blocklist in `IsDangerousCommand`, sentinel `ErrDangerousCommand`, enforced in `RunCommand` + `RunCommandWithLiveOutput`; 5 new tests | go test: PASS, go vet: clean |
| R-03 | **Single CPU Percent Call** | DONE | Medium | `cpu.go`: Replaced dual `cpu.Percent(500ms)` calls with one; per-CPU estimated from total/coreCount; 6 new tests | go test: PASS, go vet: clean |
| R-04 | **Fix Tick Event Time Skew** | DONE | Medium | `App.go`: Moved `collectInterfaces()` before all `PushMetric` calls; all 6 metrics now collected in same ~1s window | go test: PASS, go vet: clean |

---

## Final Verification

| Check | Result |
|-------|--------|
| `go test ./internal/common/...` | PASS (1.433s) |
| `go test ./internal/devops/...` | PASS (3.105s) |
| `go test ./internal/app/...` | PASS (1.963s) |
| `go test ./internal/netops/...` | PASS (1.464s) |
| `go test ./internal/sysops/...` | PASS (3.938s) |
| `go vet ./internal/...` | CLEAN |
| `npm run build` (frontend) | SUCCESS |
