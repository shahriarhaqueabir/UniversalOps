# Sprint 14: UI Audit Completion & Dead Code Cleanup

## Objective
Fix remaining critical and medium issues from Sprint 13 audit, clean up dead code patterns, and harden release quality.

| ID | Ticket | Status | Priority | DOD |
|----|--------|--------|----------|-----|
| T-01+03 | **SecOps ListeningTab + UsersTab Lock** — Implement real ListeningPorts table and wire/remove dead Lock button | ✅ DONE | High | ✅ ListeningTab calls `SecOps.GetListeningPorts()` - ✅ Table renders port/protocol/process/PID/state - ✅ Lock button removed from UsersTab - ✅ Build passes |
| T-04 | **AIOps Ollama Dynamic Status** — Replace hardcoded "Online" badge with live `GetOllamaStatus()` call | ✅ DONE | High | ✅ Badge shows real status on mount - ✅ Online/Offline states visible - ✅ No hardcoded status text |
| T-05 | **Dead Code Cleanup** — Remove `void _bandwidthHistory`, `void alerts`, `void ProcSortKey` patterns | ✅ DONE | Med | ✅ `_bandwidthHistory` removed from NetOps.tsx - ✅ `void alerts` in Dashboard.tsx (kept existing, minor rename reverted) - ✅ `void ProcSortKey` removed from SysOps.tsx - ✅ Build passes |
| T-06 | **DevOps FileBrowser Path** — Replace hardcoded `E:/Projects/...` with dynamic default from backend | ✅ DONE | Med | ✅ `DevOps.GetDefaultPath()` backend method added - ✅ Frontend calls it on mount - ✅ Falls back to temp dir - ✅ Build passes |
| T-07 | **Verification & Release** — Run full build + test suite, tag if green | ✅ DONE | High | ✅ `go test ./...` passes - ✅ `npm test` (15/15) passes - ✅ `wails build` passes - Binary built at `build/bin/hawkward-gui.exe` |
