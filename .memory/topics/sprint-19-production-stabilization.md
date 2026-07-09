# Sprint 19: Production Stabilization

## Goal
Fix all runtime warnings/errors from production logs, verify every UI element works, complete the release distribution flow.

## Detail

### T-01: Fix nil error wrap + non-English locale fallback
- **Files**: `internal/secops/users.go`, `internal/secops/firewall.go`
- **Fix GetUsers**: The `%!w(<nil>)` error occurs when `cmd.Output()` returns nil output + nil err — `net user` produces empty output on some systems. Fix the error wrapping logic.
- **Fix GetFirewallRules**: netsh output is locale-dependent. Add a fallback that uses `sc query` or WMIC for firewall rules when netsh fails to parse.

### T-02: Fix PowerShell JSON parsing
- **Files**: `internal/devops/services.go`, `internal/secops/events.go`
- **Root cause**: PowerShell returns JSON with `{ "Value": "Running" }` wrapped objects or numeric `-` values for fields like StartType. The `CleanJSON` function doesn't handle these.
- **Fix**: Add structural JSON cleaning before unmarshal — unwrap Value-wrapped objects, sanitize numeric literals.

### T-03: Improve GetDefenderStatus fallback handling
- **Files**: `internal/secops/defender.go`
- **Issues**: All 3 approaches fail. Need to determine if Defender is simply not installed vs. broken.
- **Fix**: Add a pre-check for `MpCmdRun.exe` existence before attempting PowerShell queries.

### T-04: Fix ListDirectory path handling
- **Files**: `internal/devops/filebrowser.go` + `DevOps.tsx` frontend
- **Issue**: The `isPathSafe` function allows absolute paths but the frontend may send empty or malformed paths.
- **Fix**: Add input validation for empty paths, ensure default path fallback works.

### T-05: Fix RunPowerShell profile path
- **Files**: `internal/devops/shell.go`
- **Issue**: Profile path resolves relative to exe directory, fails in dev mode.
- **Fix**: Add a fallback to check working directory, then continue gracefully (already does this, but log is noisy).

### T-06: Verify NetOps tabs
- **Files**: `internal/netops/`, `NetOps.tsx`
- **Action**: Run local test of Probes, Resolution, and Endpoints tabs. Check if backend returns data, check frontend rendering.

### T-07: Audit every button/tab
- **Action**: Go through every tab in every section. Verify each button calls the correct backend method and displays results.

### T-08: Release pipeline verification
- **Files**: `.github/workflows/release.yml`
- **Action**: Verify workflow builds with `wails build`, creates exe artifact, attaches to GitHub Release.

### T-09: README launch guide
- **Files**: `README.md`
- **Action**: Add download + run instructions for non-programmers.

### T-10: Build & tag v1.2.0
- **Action**: Final build verification, tag, push.
