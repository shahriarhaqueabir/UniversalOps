# Collector Architecture — Implementation Reference

## Overview

Universal-Ops uses a modular, multi-threaded collector system. Every data source — CPU, memory, disk, network — runs in its own independent goroutine with its own schedule and lifecycle.

## Design Goals
- **Zero Lock Contention**: Sharded data stores ensure concurrent collection does not block telemetry reading.
- **Granular Control**: Every collector can be independently enabled/disabled or re-scheduled.
- **Resilient Pipeline**: Collectors use exponential backoff on OS failure and never block the main loop.
- **Efficiency**: Expensive statistical analysis is memoized and only recalculated when new metrics arrive.

### 1. Core Types (`internal/common/collector.go`)

```go
type CollectorID string

const (
    CollectorCPU    CollectorID = "cpu"
    CollectorMem    CollectorID = "memory"
    CollectorDisk   CollectorID = "disk"
    CollectorNet    CollectorID = "network"
    CollectorTemp   CollectorID = "temperature"
    CollectorProc   CollectorID = "processes"
    CollectorUptime CollectorID = "uptime"
    CollectorLoad   CollectorID = "load"
    CollectorSwap   CollectorID = "swap"
    CollectorDiskIO CollectorID = "diskio"
    CollectorOpenFD CollectorID = "openfds"
)

type MetricSample struct {
    Name  string    // e.g. "cpu.percent"
    Unit  string    // e.g. "%"
    Value float64
}

type CollectorInfo struct {
    ID              CollectorID
    Name            string
    Description     string
    DefaultInterval time.Duration
    DefaultEnabled  bool
}

type Collector interface {
    Info() CollectorInfo
    Collect(ctx context.Context) ([]MetricSample, error)
}
```

### 2. Registry (`internal/common/registry.go`)
The registry manages the state of all available collectors and provides a unified injection point for the DataPipeline.

### 3. Scheduler (`internal/common/scheduler.go`)
The scheduler spawns one goroutine per **enabled** collector.
- **Isolation**: A failing collector cannot crash the system or stall other collectors.
- **Backoff**: Automatic exponential backoff is applied if a collector returns repeated errors.

### 4. Pipeline & Memoization (`internal/common/pipeline.go`)
The `DataPipeline` acts as the sharded metric buffer.
- **Linear Regression**: Pearson R correlation and trend slopes are calculated on-ingestion and cached. 
- **Read Performance**: The UI (Dashboard) reads pre-computed stats from the cache, ensuring sub-millisecond response times even under high ingest load.

### 4. Collector Implementations (`internal/app/collectors.go`)

Eleven collectors, each wrapping existing sysops/netops calls:

| ID | Wraps | Metric Name | Unit | Default |
|----|-------|-------------|------|---------|
| `cpu` | `sysops.GetCPUStats()` | `cpu.percent` | `%` | enabled |
| `memory` | `sysops.GetMemoryStats()` | `memory.percent` | `%` | enabled |
| `disk` | `sysops.GetDiskStats()` | `disk.percent` | `%` | enabled |
| `network` | `NetOps.collectInterfaces()` | `network.rx.rate`, `network.tx.rate` | `bps` | enabled |
| `temperature` | `sensors.SensorsTemperatures()` | `cpu.temperature` | `°C` | enabled |
| `processes` | `sysops.GetSystemInfo()` | `process.count` | `count` | enabled |
| `uptime` | `sysops.GetSystemInfo()` | `system.uptime` | `s` | enabled |
| `load` | `load.Avg()` | `load.1m`, `load.5m`, `load.15m` | `load` | enabled |
| `swap` | `sysops.GetMemoryStats()` | `swap.percent` | `%` | enabled |
| `diskio` | `sysops.GetDiskIO()` | `disk.io.read`, `disk.io.write` | `bytes` | enabled |
| `openfds` | `process.NumFDs()` | `system.open_fds` | `count` | **disabled** (expensive on Windows) |

Note: **GPU** and **Battery** stay on-demand only (not tick-based) — called via existing facade methods.

Collectors are wired via a single `func RegisterCollectors(r *CollectorRegistry, app *App)` factory.

### 5. Backend Bindings (`internal/app/App.go` — additions)

```go
func (a *App) ListCollectors() []CollectorInfo   // name, enabled, interval, lastRun
func (a *App) SetCollectorEnabled(id string, enabled bool) error
func (a *App) SetCollectorInterval(id string, intervalMs int) error
func (a *App) TriggerCollector(id string) error  // one-shot manual collect
```

### 6. Alert Evaluation & Dashboard Emission

Alert evaluation and dashboard event emission become **their own scheduled operation** (separate from data collection) on a fixed 5s tick. They read from the pipeline, which is fed by all collectors independently.

### 7. Frontend (`Settings.tsx` additions)

Each collector gets a toggle + interval selector + "Collect Now" button in the Settings page:

```
┌─ Collectors ─────────────────────────────────┐
│ CPU Usage    [●] Enabled   Interval: [3s ▼]  │
│ Memory       [●] Enabled   Interval: [5s ▼]  │
│ Disk         [●] Enabled   Interval: [10s ▼] │
│ Network      [○] Disabled                    │
│ Temperature  [○] Disabled                    │
│ Processes    [●] Enabled   Interval: [10s ▼] │
└──────────────────────────────────────────────┘
```

New TS type:

```typescript
interface CollectorStatus {
  id: string
  name: string
  description: string
  enabled: boolean
  interval_ms: number
  default_interval_ms: number
  last_run: string | null
}
```

### 8. Migration Path (No Breaking Changes)

1. Add `Collector`, `CollectorRegistry`, `CollectorScheduler` to `internal/common/` — zero existing code touched
2. Add `RegisterCollectors()` + 6 collector implementations to `internal/app/collectors.go`
3. Modify `App.go`:
   - `startTickLoop()` → `startCollectorScheduler()`
   - `collectAndEmit()` **removed**
   - `evaluateAndEmit()` — new standalone function for alerts + dashboard events, runs on its own tick
   - `Startup()` calls `RegisterCollectors()` then `scheduler.Start()`
   - `Shutdown()` calls `scheduler.Stop(5s)`
4. Add 4 new binding methods to `App.go`
5. Update `Settings.tsx` with collector management UI
6. Update `types/index.ts` with `CollectorStatus`

### 9. Files Changed

| File | Action |
|------|--------|
| `internal/common/collector.go` | **NEW** — Collector interface, types, IDs |
| `internal/common/registry.go` | **NEW** — CollectorRegistry |
| `internal/common/scheduler.go` | **NEW** — CollectorScheduler |
| `internal/app/collectors.go` | **NEW** — 6 collector impls + factory |
| `internal/app/App.go` | **MODIFY** — replace tick loop, add bindings |
| `internal/app/App.go` | **MODIFY** — Startup/Shutdown |
| `internal/sysops/collector.go` | **KEEP** — still used by CPU/Mem/Disk collectors |
| `cmd/.../frontend/src/types/index.ts` | **MODIFY** — add CollectorStatus |
| `cmd/.../frontend/src/pages/Settings.tsx` | **MODIFY** — add collector toggles |
