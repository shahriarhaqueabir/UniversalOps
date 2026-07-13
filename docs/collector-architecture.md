# Collector Architecture — Design Proposal

## Problem

Today `collectAndEmit()` in `App.go` runs a single ticker and pulls **every** data source on every tick — CPU, memory, disk, network interfaces, temperature sensors, process count — then evaluates alerts and emits dashboard events all in one fused function. Users cannot:

- Disable subsystems they don't need (e.g., temperature on a VM)
- Set per-collector intervals (network at 30s vs CPU at 3s)
- Trigger a single collector manually

## Design

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

```go
type CollectorRegistry struct {
    mu         sync.RWMutex
    collectors map[CollectorID]*managedCollector
    pipeline   *DataPipeline
}

type managedCollector struct {
    Collector
    enabled  bool
    interval time.Duration
    lastRun  time.Time
}

func (r *CollectorRegistry) Register(c Collector)
func (r *CollectorRegistry) Enable(id CollectorID) error
func (r *CollectorRegistry) Disable(id CollectorID) error
func (r *CollectorRegistry) SetInterval(id CollectorID, d time.Duration) error
func (r *CollectorRegistry) CollectNow(id CollectorID) ([]MetricSample, error)
func (r *CollectorRegistry) Snapshot() []CollectorStatus
```

Collecting pushes directly into the pipeline via `r.pipeline.PushMetric()`.

### 3. Scheduler (`internal/common/scheduler.go`)

Replaces the single `startTickLoop()` goroutine with one goroutine per **enabled** collector:

```go
type CollectorScheduler struct {
    registry *CollectorRegistry
    quit     chan struct{}
    wg       sync.WaitGroup
    logger   *log.Logger
}

func (s *CollectorScheduler) Start()
func (s *CollectorScheduler) Stop(timeout time.Duration)
```

Each goroutine reads its collector's interval from the registry (hot-swappable via `SetInterval`), ticks independently, and pushes samples to the pipeline. When a collector is disabled mid-flight, its goroutine simply skips ticks until re-enabled.

On collector **error**, the scheduler logs + moves on (no crash, no backpressure). On repeated errors, exponential backoff is applied.

### 4. Collector Implementations (`internal/app/collectors.go`)

Six collectors, each wrapping existing sysops/netops calls:

| ID | Wraps | Metric Name | Unit |
|----|-------|-------------|------|
| `cpu` | `sysops.GetCPUStats()` | `cpu.percent` | `%` |
| `memory` | `sysops.GetMemoryStats()` | `memory.percent` | `%` |
| `disk` | `sysops.GetDiskStats()` | `disk.percent` | `%` |
| `network` | `NetOps.collectInterfaces()` | `network.rx.rate`, `network.tx.rate` | `bps` |
| `temperature` | `sensors.SensorsTemperatures()` | `cpu.temperature` | `°C` |
| `processes` | `sysops.GetSystemInfo()` | `process.count` | `count` |

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
