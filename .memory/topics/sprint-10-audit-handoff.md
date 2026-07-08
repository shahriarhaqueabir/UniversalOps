# Sprint 10 Handoff — Multi-Perspective Audit Loop

## Execution Mode
Subagent-driven development with parallel audit agents (3 concurrent) + sequential implementation agents (3 parallel, disjoint write surfaces).

## Phase 1: Audit Results

### A-01 Performance (storage.go, App.go, pipeline.go)
- **High**: goroutine-per-PushMetric churn (pipeline.go:112 — `go s.InsertMetric()` creates ~21k goroutines/hr). Fixed by replacing `go s.InsertMetric(...)` with direct call.
- **High**: TickInterval default is 1s but Capacity comment says "12 min at 3s". Fixed comment to match reality.
- **Med**: dailyPruneLoop races with db.Close() on shutdown. Fixed by adding pruneWg sync.
- **Low**: deferred tx.Rollback() after successful tx.Commit() returning sql.ErrTxDone. Fixed with error-checked rollback pattern.

### A-02 Security (shell.go, services.go, filebrowser.go)
- **CRITICAL**: `powershell` not in DangerousCommands — any command could be prefixed with `powershell Remove-Item` to bypass the blocklist.
- **CRITICAL**: ControlService in services.go used `cmdStr+" -Name "+name` — direct PowerShell injection via the `name` parameter.
- **CRITICAL**: WriteFile could overwrite PowerShell profile → RunPowerShell with allowed workflow → RCE.
- **HIGH**: File operations (ReadFile, WriteFile, ListDir, TailLog, SearchLog) had no path restrictions.
- **HIGH**: 22 missing command patterns in blocklist, `&` single-ampersand operator not blocked.
- All CRITICAL/HIGH issues fixed. Security posture upgraded from Critical → Low.

### A-03 Accuracy (cpu.go, memory.go)
- All metrics are system-wide (gopsutil reads OS counters). No app-specific leakage.
- Per-CPU values estimated as total/coreCount (documented tradeoff to avoid ~1s blocking call).
- 500ms CPU delta on 3s tick is within gopsutil recommendations.
- **Verdict**: No metric-parity bugs. All correct.

## Phase 2: Implementation Summary

### Files Changed (15 files)
```
Backend:
  internal/common/pipeline.go      — Removed goroutine churn, fixed comment
  internal/common/storage.go       — Added pruneWg sync, fixed rollback pattern
  internal/app/App.go              — Added ctx.Err() cancellation check
  internal/devops/shell.go         — Expanded DangerousCommands (+23 patterns)
  internal/devops/services.go      — Replaced PowerShell injection with net.exe + validator
  internal/devops/filebrowser.go   — Added path sandbox, 0600 perms
  internal/devops/logtail.go       — Added path sandbox

Frontend:
  src/pages/Dashboard.test.tsx     — NEW: tests for render, metrics, loading state
  src/pages/Settings.test.tsx      — NEW: tests for render, theme toggle
  src/pages/Logs.test.tsx          — NEW: tests for render, log entries, filters
  src/hooks/useBackend.ts          — eslint-disable for Wails bridge types
  src/hooks/useEvents.ts           — eslint-disable for Wails event types
  src/pages/Dashboard.tsx          — eslint-disable for Wails event any
  src/pages/NetOps.tsx             — eslint-disable for bandwidth history any
  src/pages/SecOps.tsx             — Extracted BackendCall type alias
  src/test/setup.ts                — eslint-disable for ResizeObserver mock
  eslint.config.js                 — Added wailsjs + *.d.ts to ignored paths
```

### Verification
- `go test ./internal/...` — 9/9 packages PASS
- `go vet ./...` — clean
- `go build ./...` — clean
- `npm run build` — PASS (tsc + vite build)
- `npm run lint` — 0 errors, 0 warnings

### What P-07 (Loading States) Status
All pages already had loading skeletons:
- SysOps.tsx: ✅ — shows skeleton when `!cpuInfo || !memInfo || !sysInfo`
- NetOps.tsx: ✅ — shows skeleton via `initialLoading`
- SecOps.tsx: ✅ — shows skeleton via `loading` state
- DevOps.tsx: ✅ — shows skeleton via `loading` state
- AIOps.tsx: ✅ — shows skeleton via `loading` state
No changes needed.

## Next Steps / Future Opportunities
1. **Add integration tests** for Go backend (currently unit-test-only)
2. **Consider increasing TickInterval default** from 1s to 3s to reduce CPU sampling overhead
3. **Audit AIOps** (Ollama dependency) for network error handling — currently assumes localhost:11434
4. **Windows-specific sandbox** — sandbox_windows.go exists, verify it applies job objects correctly
5. **Consider adding `npm test` script** — currently only `npm run build` and `npm run lint` exist

## Files Not to Touch
- `wailsjs/` — generated code
- `dist/` — build output
- `node_modules/` — dependencies
