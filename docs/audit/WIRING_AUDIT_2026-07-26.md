# UniversalOps: Comprehensive Wiring Audit Report

**Date:** 2026-07-26 | **Scope:** Full-stack Go → TypeScript | **Method:** 6 parallel sub-agents

---

## Executive Summary

6 specialized sub-agents audited: Go backend wiring, frontend-backend bindings, dead code, type consistency, interface compliance, and MCP/AIOps pipeline. **Total: 27 findings** (9 🔴 Critical, 10 🟠 Warning, 8 🟡 Info).

---

## 🔴 CRITICAL Findings

### 1. `AIOps.WithTimeout()` — Wails binding will crash at runtime
**File:** `internal/app/AIOps.go:922`
**Problem:** Returns `(context.Context, context.CancelFunc)` — neither is JSON-serializable. If the frontend ever calls this, Wails panics during `json.Marshal`.
**Binding:** `AIOps.d.ts` shows `Promise<context.Context|context.CancelFunc>` — a codegen artifact.
**Fix:** Remove from facade or refactor to not expose Go internals.

### 2. MCP Server is never started — no transport layer
**Files:** `internal/aiops/mcp/server.go`, `internal/app/AIOps.go`, `internal/app/App.go`
**Problem:** `mcp.Server` is created inside `NewAIOps()` but has **no transport** (no HTTP, stdio, WebSocket). The 24 tools are only reachable through Ollama's `<function>` tag parsing in chat. No `Start()`, no `Shutdown()`, no lifecycle management.
**Fix:** Either add a transport layer or document that MCP is an internal-only function-calling abstraction.

### 3. Backend emits `metrics` + `timeline` Wails events every tick — frontend never listens
**Files:** `internal/app/App.go:153-180` (EventsEmit), frontend `App.tsx`
**Evidence:** Frontend only subscribes to `alert` events. Dashboard uses TanStack Query polling instead. Every tick, the backend serializes and emits data that is silently dropped.
**Fix:** Remove the `EventsEmit` calls for `metrics`/`timeline`, or subscribe on the frontend side.

### 4. `MetricsEvent` dead code — `buildMetricsEvent()` never called
**File:** `internal/app/App.go:787-820`
**Problem:** 34 lines of dead code. Also hardcodes `Connections: 0` which would be misleading if re-activated. The actual metrics flow uses `OnMetricsEmit` callback.

### 5. `traceroute` MCP handler silently ignores `max_ttl` parameter
**File:** `internal/aiops/mcp/server.go:195-201`
**Problem:** The input schema declares a `max_ttl` parameter with default 30, but the handler struct only parses `target`. The parameter is silently dropped.

### 6. DevOps nil-ctx panic risk in `RunCommandLive`/`RunPowerShellLive`/`RunGitBashLive`
**File:** `internal/app/DevOps.go`
**Problem:** `NewDevOps(nil, ...)` initializes with nil context. The three "Live" methods call `EventsEmit(d.ctx, ...)` inside goroutines. If any is called between `NewApp()` and `Startup()` (possible via eager frontend bindings), `d.ctx` is nil → **panic**.

### 7. `connection.count` metric (MetricConnCnt) defined but no collector registered
**Files:** `internal/common/metrics.go`, `internal/app/collectors.go`
**Problem:** `MetricConnCnt = "connection.count"` is in `DefaultMetrics` with alert thresholds, but no collector is registered in `RegisterCollectors()`. The value is computed live without ever entering the pipeline ring buffer.

### 8. Missing DB indexes on 3 tables (previously flagged, still unfixed)
**File:** `internal/common/storage.go` (migration v1)
**Problem:** `forensics`, `incidents`, `custom_workflows` tables lack timestamp indexes — all other tables have them. This was flagged in the previous audit and remains unaddressed.

### 9. `GPUData` TypeScript type missing 3 fields + `Dashboard.go` drops them
**Files:** `internal/app/Dashboard.go:126-147`, `frontend/src/types/index.ts`
**Problem:** Go `GPUInfo` has 8 fields (added `temperature`, `utilization`, `fan_speed`). TypeScript `GPUData` still expects only 5. Additionally, `Dashboard.GetDashboardData()` explicitly maps only 5 fields — the 3 telemetry fields are dropped before they even reach the JSON.

---

## 🟠 WARNING Findings

### W1. Handshake registrations all use empty `sessionID = ""` — broken DORA audit trail
**Files:** `internal/app/SysOps.go:452`, `internal/app/SecOps.go:145` (+8 more SecOps callers)
**Impact:** `LogDecision()` writes empty sessionID to `decisions_audit` table. Cannot correlate actions to user sessions.

### W2. `Workflow.go` has unused imports (`netops`, `sysops`)
**File:** `internal/app/Workflow.go:10-16`
**Problem:** `netops` and `sysops` imported but never referenced in the file.

### W3. `TrendInfo` dual definitions (int vs string) — fragile conversion layer
**Files:** `internal/common/Types.go`, `internal/app/Types.go`
**Problem:** `common.TrendInfo.Direction` is `TrendDirection` (int 0/1/-1). `app.TrendInfo.Direction` is `string` ("rising"/"falling"/"stable"). The `convertTrendInfo()` function manually maps them — any new `common.TrendInfo` field must be manually added.

### W4. `DashboardData` shows `Processes` and `Connections` but frontend renders zero visual elements for them
**Files:** `internal/app/Dashboard.go`, `frontend/Dashboard.tsx`
**Problem:** Data flows correctly to frontend but only 6 KPI cards exist (CPU, Memory, Disk, GPU, Battery, Network). Process count and connection count are invisible.

### W5. Prometheus dependency behind build tag — `go mod tidy` will silently remove it
**Files:** `go.mod`, `internal/common/metrics_exporter.go`
**Problem:** `github.com/prometheus/client_golang` is only imported behind `//go:build prometheus`. Running `go mod tidy` without `-tags prometheus` drops it from `go.mod`.

### W6. DevOps/NetOps `toolSpec` struct defined twice (~60 lines duplicated)
**File:** `internal/app/DevOps.go` (lines ~300 and ~809)
**Problem:** `GetInstalledTools()` and `GetEnvironment()` each define the same `toolSpec` struct + `detectTool()` helper. Both need updating when tools are added.

### W7. `Ollama.CreateModel` swallows all progress callbacks
**File:** `internal/aiops/ollama.go:177-183`
**Problem:** The progress callback returns `nil` — all model creation progress is silently dropped. The frontend sets up `EventsOn('ollama:progress', ...)` but no events are ever emitted.

### W8. `SysOps.GetGPUInfo()`/`GetHardwareInfo()` lack panic recovery
**File:** `internal/app/SysOps.go`
**Problem:** WMI/reflection calls via gopsutil can panic. Most SysOps methods have `defer recover()` — these two don't.

### W9. `GetTimelineEventByID`, `GenerateReport`, `GetReport` return `*T` typed as `T` in binding
**Files:** `internal/app/Timeline.go:82`, `internal/app/Reports.go:225,300`
**Problem:** `.d.ts` declares `T` not `T | null`, but Go returns `nil` on error. Frontend will receive `null` JSON for an expected object — uncaught runtime error.

### W10. GPU collector uses non-standard hardcoded CollectorID `"gpu"` instead of a constant
**File:** `internal/app/collectors.go:~341`
**Problem:** All other collectors use `common.CollectorCPU`, `common.CollectorMem`, etc. GPU collector uses raw string — bypasses type system.

---

## 🟡 INFO Findings

### I1. `internal/netops/qlog.go` — Entire file is dead code (zero callers)
**File:** `internal/netops/qlog.go`
**Lines:** ~44 — `QLogFile`, `QLogTrace`, `QLogEvent`, `ParseQLog()` never used anywhere.

### I2. `SystemKnowledge` has no TypeScript interface — frontend can't type-check it
**File:** `frontend/src/types/index.ts`
**Impact:** If `GetSnapshot()` is ever called from frontend, response is untyped with dotted JSON keys.

### I3. `TrendDirection` type alias uses wrong values (`'up'|'down'|'stable'` vs actual `'rising'|'falling'|'stable'`)
**File:** `frontend/src/types/index.ts:162`
**Impact:** Dead type alias — never imported anywhere — but would mislead developers.

### I4. `MetricDataPoint` — dead interface with PascalCase fields that won't deserialize
**File:** `frontend/src/types/index.ts:169-172`
**Impact:** `Time`/`Value` (PascalCase) won't match Go's `timestamp`/`value` JSON output. Zero imports in frontend code.

### I5. App version drift: `wails.json` = 1.3.0, `App.go` = 1.3.1, `CHANGELOG.md` = 1.3.1
**Files:** `wails.json:12`, `internal/app/App.go:352`, `CHANGELOG.md`

### I6. `ActionPreview` — Go sends 9 fields, TypeScript declares only 6
**Files:** `internal/common/Types.go:42-53`, `frontend/src/types/index.ts:882-889`
**Missing TS fields:** `typical_values`, `workflow_id`, `steps`

### I7. `TrendDirection.String()` exists but `state_query.go` still uses hardcoded `"CPU"`/`"memory"`/`"disk"` display names
**File:** `internal/aiops/state_query.go`
**Impact:** Display names don't align with `MetricCPU = "cpu.percent"` constant convention.

### I8. `handleDNSSLookup` function name typo (double 'S')
**File:** `internal/aiops/mcp/server.go:381`

---

## Previously Flagged & Now Confirmed Fixed

| Issue | Status |
|-------|--------|
| All `.d.ts` facade methods exist in Go (140+ methods) | ✅ |
| All cross-package calls reference valid functions | ✅ |
| `Register(sessionID, action, command, params)` signatures match | ✅ |
| `LogDecision()` callers match signature (3 callers) | ✅ |
| All 14 bound structs exist on `App` and are initialized | ✅ |
| All 12 `Collector` implementations registered | ✅ |
| All build tag pairs balanced (Windows/!Windows/Prometheus) | ✅ |
| `go build ./...` compiles | ✅ (sub-agent 1 confirmed) |
| Startup/Shutdown chain properly ordered | ✅ |

---

## Recommended Fix Priority

| Priority | Finding | Effort |
|----------|---------|--------|
| **P0** | Fix `AIOps.WithTimeout()` — runtime crash risk | 5 min |
| **P0** | Add nil-ctx guard to DevOps Live methods | 10 min |
| **P1** | Remove wasteful `EventsEmit` for metrics/timeline | 10 min |
| **P1** | Fix `traceroute` max_ttl parameter extraction | 5 min |
| **P1** | Fix `DashboardData.GPU` to include temperature/utilization/fan_speed | 15 min |
| **P1** | Register `connection.count` collector | 15 min |
| **P2** | Add DB migration for missing indexes (3 tables) | 15 min |
| **P2** | Remove dead code: `qlog.go`, `buildMetricsEvent()`, `UpdatePrometheus()` | 15 min |
| **P2** | Add sessionID threading through handshake registrations | 30 min |
| **P3** | Deduplicate `toolSpec` in DevOps.go | 20 min |
| **P3** | Add `defer recover()` to unprotected SysOps WMI methods | 10 min |
| **P3** | Remove unused imports in Workflow.go | 5 min |
| **P4** | Align TypeScript types: `GPUData`, `ActionPreview`, `SystemKnowledge` | 30 min |
| **P4** | Update `wails.json` version to match App.go | 2 min |
