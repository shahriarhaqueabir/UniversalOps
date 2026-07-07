# Hawkward Overhaul Plan — v3: Interactive Visual Operations Platform

**Design philosophy**: A terminal-based operations platform that rivals desktop monitoring tools in interactivity, visual richness, and real-time capability. Squib-inspired aesthetics + netscanner-grade network tools + hackingtool's search/discovery + Grafana-style dashboards — all in a keyboard-first TUI.

**Status**: Active — Sprint 1 in progress
**Target release**: v3.0.0
**Go version**: 1.26.4
**Module**: `hawkward` (import prefix `github.com/shahriarhaqueabir/AllOpsFull/...`)

---

## Core Architecture Decisions

### Visualization Engine (New: `internal/common/charts/`)
Build a pure-Lip-Gloss chart library using Unicode block + braille characters:
- **No external charting dependencies** — single binary constraint preserved
- **Braille-based line charts** (`⣀⣤⣶⣿`) for smooth curves in terminal (4×8 resolution per cell)
- **Block-element bar charts** (`▁▂▃▄▅▆▇█`) for stacked and grouped bars
- **Unicode gauge system** with color gradients and thresholds
- **Sparkline v2** — enhanced compact trend lines with min/max markers and trend arrows (↑↓)
- **Auto-scaling axis** — dynamic range calculation for all chart types
- **Theme-aware** — all chart colors from common.Palette, extended with chart-specific fields

### Data Pipeline (enhance `internal/common/`)
- **Time-series storage** — ring buffer per metric with `internal/common/timeseries.go`
- **Streaming data model** — push-based updates from ops layers to view layer
- **Forecast engine** — simple linear regression + exponential smoothing (stdlib `math` only) in `internal/common/forecast.go`
- **Anomaly detection v2** — threshold + statistical outlier detection (existing `internal/aiops/ollama.go` + new)

### Card System (New: `internal/ui/cards.go`)
Replace tab-based text views with a composable card system:
- **CardGrid** — responsive grid layout with auto-column-count
- **Card** base struct with title, body, footer, status, focus states
- Specialized card types: `GaugeCard`, `ChartCard`, `TableCard`, `StatusCard`
- `DetailPanel` — expandable overlay for drill-down data
- `SplitPane` — resizable horizontal/vertical split with drag handle

### Dashboard Landing Page (New: `internal/ui/dashboard.go`)
Replaces `internal/ui/mainmenu.go` as the default home screen:
- CPU/MEM/DISK health gauges with trend arrows and forecast overlays
- 5 operation category cards with sparklines and counts
- Active alerts section with flap detection
- Quick-jump footer with key hints
- ASCII art header (netscanner-inspired)

### Global Command Palette (New: `internal/ui/commandpalette.go`)
Accessible via `/` from any screen (replaces current `internal/ui/keys.go` filter binding):
- Fuzzy search across all operations by name, description, keyword, tag
- Result list with actions — Enter to navigate, `?` for description
- Search history (↑↓ cycles through recent searches)
- Tag filtering (`t` key) — hackingtool-inspired
- Recommend mode (`r` key) — natural language → matching operations

### Interaction Enhancements
- **Selectable charts** — ← → cursor movement over data points with value tooltip
- **Card focus** — Tab/shift-tab moves focus between cards, Enter for drill-down
- **Drill-down overlays** — Enter on any card/cell opens detailed view overlay
- **Split-pane layout** — `|` key toggles split pane; `[` `]` resizes
- **Context menu** — `ctrl+space` opens context menu for focused item

---

## Sprint 1: Visualization Engine (Foundation)

**Goal**: Build the chart/graph/gauge/card system that all views depend on. No visible UI changes yet.
**Dependencies**: None — isolated new package additions.
**Parallel lanes**: A (Charts), F (Config) can start simultaneously.

### 1.1 Chart Component Library

| Component | File | Unicode Technique | Concrete API |
|-----------|------|-------------------|--------------|
| **LineChart** | `internal/common/charts/line.go` | Braille (⣀⣤⣶⣿) | `NewLineChart(config ChartConfig)`, `AddSeries(name string, data []float64, color lipgloss.Color)`, `SetForecast(data []float64)`, `SetThreshold(value float64, label string)`, `CursorAt(index int) string`, `Render() string` |
| **BarChart** | `internal/common/charts/bar.go` | Block (█▇▆▅▄▃▂▁) | `NewBarChart(config ChartConfig)` (horizontal/vertical), `AddBar(label string, value float64, color lipgloss.Color)`, `AddGroup(label string, values []float64, colors []lipgloss.Color)`, `SortByValue(asc bool)`, `Render() string` |
| **AreaChart** | `internal/common/charts/area.go` | Braille + color fill | `NewAreaChart(config ChartConfig)`, stacked area support, gradient fill via character density, `Render() string` |
| **Gauge** | `internal/common/charts/gauge.go` | Block (▓▒░) | `NewGauge(config ChartConfig)`, `SetValue(pct float64)`, `SetThresholds(warn, crit float64)`, `SetLabel(text string)`, `ShowPercent(show bool)`, `Render() string` |
| **Sparkline v2** | `internal/common/charts/sparkline.go` | Block (▁▂▃▄▅▆▇█) | `NewSparkline(width int)`, `SetValues(data []float64)`, `ShowMinMax(show bool)`, `ShowTrend(show bool)`, `ShowLabels(show bool)`, `Render() string` |
| **HeatMap** | `internal/common/charts/heatmap.go` | Colored blocks (▪) | `NewHeatMap(rows, cols int)`, `SetCell(row, col int, value float64)`, `SetRowLabel(row int, label string)`, `SetColLabel(col int, label string)`, `ColorScheme(scheme HeatScheme)`, `Render() string` |
| **NumericDisplay** | `internal/common/charts/number.go` | Large digits w/ styling | `NewNumericDisplay(config ChartConfig)`, `SetValue(value float64, unit string)`, `SetTrend(dir TrendDirection, pct float64)`, `SetColor(color lipgloss.Color)`, `Render() string` |

**Chart configuration model** (`internal/common/charts/config.go`):
```go
type ChartConfig struct {
    Width, Height            int
    Title                    string
    ShowLegend, ShowGrid     bool
    AutoScale                bool
    MinValue, MaxValue       float64  // fixed scale when !AutoScale
    Colors                   []lipgloss.Color
    SeriesNames              []string
    XLabels                  []string  // for bar charts, time labels for line charts
    FormatFn                 func(float64) string  // value formatter
}
```

### 1.2 Time-Series Data Store

| Task | File(s) | Concrete API/Detail |
|------|---------|---------------------|
| Ring buffer type | `internal/common/timeseries.go` | `type TimeSeries struct { Name, Unit string; Data []DataPoint; MaxSize, Head, Count int }` |
| DataPoint struct | `internal/common/timeseries.go` | `type DataPoint struct { Timestamp time.Time; Value float64 }` |
| Multi-metric store | `internal/common/timeseries.go` | `type TimeSeriesStore map[string]*TimeSeries` — `Add(name string, value float64)`, `Get(name string) []DataPoint`, `Latest(name string) (DataPoint, bool)` |
| Rolling window aggregation | `internal/common/timeseries.go` | `func (ts *TimeSeries) WindowStats(window int) WindowStats` — min, max, avg, p50, p95, p99 |
| WindowStats struct | `internal/common/timeseries.go` | `type WindowStats struct { Min, Max, Avg, P50, P95, P99 float64; Count int; Trend TrendDirection }` |
| Save/load to JSON | `internal/common/timeseries.go` | `func (store TimeSeriesStore) Save(path string) error`, `func LoadTimeSeries(path string) (TimeSeriesStore, error)` for session continuity |

**Integration points with existing code**:
- `internal/sysops/collector.go`: `CollectAllStats()` will push CPU%, memory, disk% to TimeSeriesStore via a new `PushStatsToStore(store)` call
- `internal/ui/root.go`: `RootModel` gains a `timeSeriesStore *common.TimeSeriesStore` field, initialized in `NewRootModel()`
- Tick handler in `root.go` updates the store after `CollectStats()` before building history

### 1.3 Forecast Engine

| Task | File(s) | Concrete API/Detail |
|------|---------|---------------------|
| Linear regression | `internal/common/forecast.go` | `func LinearRegression(data []DataPoint) (slope, intercept float64, rSquared float64)` — stdlib math only |
| Exponential smoothing | `internal/common/forecast.go` | `func ExponentialSmoothing(data []DataPoint, alpha float64, steps int) []DataPoint` — Holt-Winters simplified |
| Moving average | `internal/common/forecast.go` | `func SMA(data []DataPoint, window int) []DataPoint`, `func EMA(data []DataPoint, alpha float64) []DataPoint` |
| Trend direction | `internal/common/forecast.go` | `type TrendDirection int` with constants `TrendRising`, `TrendFalling`, `TrendStable`; `func CalculateTrend(data []DataPoint) TrendDirection` |
| Forecast display helpers | `internal/common/forecast.go` | `func TimeToThreshold(data []DataPoint, threshold float64) (time.Duration, bool)` — "CPU projected to reach 90% in 12m" |
| Forecast visualization | `internal/common/charts/line.go` | Forecast overlay via `LineChart.SetForecast(data []DataPoint)` — dashed/predicted region with confidence interval band |

**Integration with existing common.Palette**: Forecast colors use `p.ChartLine3` (prediction) and a dimmed version for confidence band.

### 1.4 Enhanced Card Component System

| Task | File(s) | Concrete API/Detail |
|------|---------|---------------------|
| Card base struct | `internal/ui/cards.go` | `type Card struct { ID string; Title string; Body string; Footer string; Status CardStatus; Focusable bool; focused bool }` |
| Card status enum | `internal/ui/cards.go` | `type CardStatus int` with `CardLoading`, `CardReady`, `CardError`, `CardStreaming` |
| CardGrid layout | `internal/ui/cards.go` | `type CardGrid struct { Cards []*Card; Columns int; Gap int }`, `func (g *CardGrid) Render(width int) string` — auto-column based on width |
| GaugeCard | `internal/ui/cards.go` | Embedded `Card` + `Gauge` — `NewGaugeCard(title string, gauge *charts.Gauge) *Card` |
| ChartCard | `internal/ui/cards.go` | Embedded `Card` + `LineChart`/`BarChart` — `NewChartCard(title string, chart interface{ Render() string }) *Card` |
| TableCard | `internal/ui/cards.go` | Embedded `Card` + sortable table — `NewTableCard(title string, headers []string, rows [][]string) *Card`, `SortByColumn(col int, asc bool)` |
| StatusCard | `internal/ui/cards.go` | Compact card with icon + status + sparkline — `NewStatusCard(icon, title, value string, sparkline string, status CardStatus) *Card` |
| DetailPanel | `internal/ui/cards.go` | Expandable overlay — `type DetailPanel struct { Title string; Content string; onClose func() }`, renders as layered overlay |
| SplitPane | `internal/ui/cards.go` | Resizable horizontal/vertical split — `type SplitPane struct { Left, Right tea.Model; Ratio float64; Vertical bool }`, keyboard resizable with `[`/`]` |

**Verification**: `go build ./...` + `go test ./internal/common/charts/...`

### 1.5 Configuration System (Parallel Lane F)

| Task | File(s) | Concrete API/Detail |
|------|---------|---------------------|
| Config struct | `internal/common/config.go` | `type Config struct { Theme ThemeName; RefreshInterval time.Duration; KeyBindings map[string]string; Defaults DefaultConfig }` |
| YAML load/save | `internal/common/config.go` | `func LoadConfig() (*Config, error)`, `func SaveConfig(cfg *Config) error` to `~/.config/hawkward/config.yaml` |
| Config directory | `internal/common/config.go` (reuses `ConfigDir()`) | Uses existing `common.ConfigDir()` which returns `~/.config/hawkward/` |
| Per-layer intervals | `internal/common/config.go` | `type LayerConfig struct { SysOps, NetOps, SecOps, DevOps, AIOps time.Duration }` |

**Files changed in Sprint 1**:
```
NEW:  internal/common/charts/config.go       — ChartConfig, shared constants
NEW:  internal/common/charts/line.go          — LineChart (braille)
NEW:  internal/common/charts/bar.go           — BarChart (block)
NEW:  internal/common/charts/area.go          — AreaChart (braille + fill)
NEW:  internal/common/charts/gauge.go         — Gauge (block, horiz/vert)
NEW:  internal/common/charts/sparkline.go     — Sparkline v2 (enhanced)
NEW:  internal/common/charts/heatmap.go       — HeatMap (density grid)
NEW:  internal/common/charts/number.go        — NumericDisplay (large digits)
NEW:  internal/common/charts/chart_test.go    — All chart tests
NEW:  internal/common/timeseries.go           — Ring buffer, TimeSeriesStore, WindowStats
NEW:  internal/common/forecast.go             — LinearRegression, ExponentialSmoothing, TrendDirection
NEW:  internal/common/config.go               — YAML Config struct, load/save
NEW:  internal/ui/cards.go                    — Card, CardGrid, GaugeCard, ChartCard, TableCard, StatusCard, DetailPanel, SplitPane
MOD:  internal/common/styles.go               — Add chart-card styled primitives
MOD:  internal/common/theme.go                — Extend Palette with ChartLine1-3, ChartGrid, CardBg
MOD:  internal/ui/root.go                     — Add timeSeriesStore, card registry references
MOD:  internal/common/types.go                — Add TrendDirection, CardStatus types
```

---

## Sprint 2: Dashboard & Layer Redesign

**Goal**: Replace all text-based views with interactive chart-rich card layouts. This is the biggest visual change.
**Dependencies**: Sprint 1 (charts, timeseries, cards, forecast)
**Parallel lanes**: B (UI) only after Sprint 1A

### 2.1 Dashboard Landing Page (Replaces Main Menu)

A Grafana-inspired "home" screen showing the health of the entire system at a glance:

```
┌─────────────────────────────────────────────────────────────────────┐
│  ╔═══╗  ░ HAWKWARD v2                       Fri 2026-07-07 14:32  │
│  ║   ║  Terminal Operations Platform                                │
│  ╚═══╝                                                             │
│                                                                     │
│  ┌──────────────┐ ┌──────────────┐ ┌──────────────┐                │
│  │  CPU          │ │  MEMORY      │ │  DISK        │                │
│  │  ▓▓▓▓▓▓▓░░░  │ │  ▓▓▓▓▓░░░░░  │ │  ▓▓▓▓░░░░░░  │                │
│  │  73.2%        │ │  52.8%       │ │  38.1%       │                │
│  │  ↑ 2.1% / 5m  │ │  → 0.3% / 5m │ │  ↓ 0.1% / 5m │                │
│  │  📈 forecast  │ │  📈 forecast │ │  📈 forecast │                │
│  └──────────────┘ └──────────────┘ └──────────────┘                │
│                                                                     │
│  ┌──────────────┐ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐│
│  │ 🖥 SYSOPS    │ │ 🌐 NETOPS    │ │ 🔒 SECOPS    │ │ ⚙ DEVOPS    ││
│  │ 45 processes │ │ 12 conns     │ │ 54 rules     │ │ 3 sessions   ││
│  │ ⣀⣤⣶⣿   │ │ ⣀⣤⣶⣿   │ │ ⣀⣤⣶⣿   │ │ ⣀⣤⣶⣿   ││
│  └──────────────┘ └──────────────┘ └──────────────┘ └──────────────┘│
│                                                                     │
│  Active alerts: 2  │  Anomalies: 1  │  Uptime: 12d 4h  │  [?] Help │
└─────────────────────────────────────────────────────────────────────┘
```

| Task | File(s) | Concrete API/Detail |
|------|---------|---------------------|
| Create DashboardModel | `internal/ui/dashboard.go` | `type DashboardModel struct { gauges [3]*charts.Gauge; opCards [5]*Card; alerts *AlertView; header string }` |
| ASCII art header | `internal/ui/dashboard.go` | netscanner-style identity banner, rendered with lipgloss.Center alignment |
| System health gauge cards | `internal/ui/dashboard.go` | 3x `GaugeCard` — CPU, MEM, DISK with forecasts via `forecast.TimeToThreshold()` |
| Operation category cards | `internal/ui/dashboard.go` | 5x `StatusCard` — icon, sparkline, metric count, status indicator |
| Active alerts section | `internal/ui/dashboard.go` | Flap-detected anomalies from `internal/common/alerts.go` |
| Wire into root model | `internal/ui/root.go` | Add `ScreenDashboard` constant to `common.Screen` enum; make it the new default home |
| Main menu demotion | `internal/ui/mainmenu.go` | Simplified — becomes accessible via a nav drawer or as secondary screen |

### 2.2 SysOps View — Chart-Rich System Monitor

| Task | File(s) | Concrete Change |
|------|---------|-----------------|
| Rewrite renderOverview | `internal/sysops/view.go` | Replace `renderBar()` + `RenderSparkline()` calls with `charts.Gauge`, `charts.Sparkline`, `charts.NumericDisplay` for CPU/MEM/DISK |
| Rewrite renderProcesses | `internal/sysops/view.go` | Replace fixed process table with `TableCard` — sortable, expandable rows (Enter → `DetailPanel`) |
| Rewrite renderSystemInfo | `internal/sysops/view.go` | Replace text grid with info `Card` grid, add `NumericDisplay` for uptime/process count |
| Add CPU core breakdown | `internal/sysops/view.go` | New `renderCPUCores()` — `BarChart` of per-core usage via `gopsutil/v4/cpu.Percent(0, true)` |
| Add memory breakdown | `internal/sysops/view.go` | New `renderMemoryDetail()` — stacked `BarChart` showing used/cached/available/free |
| Add process detail | `internal/sysops/view.go` | Enter on a process row → `DetailPanel` with threads, handles, memory, path |

**Existing code reference**: Current `renderOverview` (line 75), `renderProcesses` (line 121), `renderSystemInfo` (line 147) in `internal/sysops/view.go` will be fully rewritten to use card components.

### 2.3 NetOps View — Network Operations Center

| Task | File(s) | Concrete Change |
|------|---------|-----------------|
| Rewrite renderPingTab | `internal/netops/view.go` (line 70) | Latency `LineChart` card, stats summary card, RTT history from `m.PingResult.History` |
| Rewrite renderDNSTab | `internal/netops/view.go` (line 135) | DNS record type `Card` grid with expandable lists (lines 163-218 become card content) |
| Rewrite renderPortScanTab | `internal/netops/view.go` (line 221) | Progress `Gauge` card at top, `TableCard` for results with open/closed color coding |
| Rewrite renderConnectionsTab | `internal/netops/view.go` (line 268) | `TableCard` with state color coding, summary `StatusCard` for LISTEN/ESTABLISHED/TIME_WAIT counts |
| Rewrite renderTracerouteTab | `internal/netops/view.go` (line 346) | Hop-by-hop latency `BarChart` card with per-hop latency bars |
| Rewrite renderInterfacesTab | `internal/netops/view.go` (line 397) | Throughput `LineChart` for RX/TX, per-interface `StatusCard` with sparklines |
| New: connection heat map | `internal/netops/view.go` | `HeatMap` of port density (local port range × remote count) |

### 2.4 SecOps View — Security Dashboard

| Task | File(s) | Concrete Change |
|------|---------|-----------------|
| Rewrite firewall view | `internal/secops/view.go` | `Card` grid per firewall rule with allow/deny `StatusCard` badges (uses existing `secops.Model.FirewallRules`) |
| Rewrite users view | `internal/secops/view.go` | `TableCard` for user list, `DetailPanel` for group membership (uses `secops.Model.UserResults`) |
| Rewrite listening view | `internal/secops/view.go` | Port `HeatMap` colored by risk + process attribution `TableCard` (uses `secops.Model.ListeningPorts`) |
| Rewrite defender view | `internal/secops/view.go` | Defender status `GaugeCard` with last scan time, threat count (uses `secops.Model.DefenderStatus`) |
| Rewrite tasks view | `internal/secops/view.go` | `Card` grid of scheduled tasks with status + next run time (uses `secops.Model.TaskResults`) |

### 2.5 DevOps & AIOps Views

| Task | File(s) | Concrete Change |
|------|---------|-----------------|
| Rewrite DevOps view | `internal/devops/view.go` | Shell session `StatusCard`s, log tail view with search highlight using `TableCard` |
| Rewrite AIOps view | `internal/aiops/view.go` | Chat message `Card` layout, report preview `Card`, Ollama status `Card` |

### 2.6 Interaction Layer (cards.go enhancements)

| Task | File(s) | Concrete Change |
|------|---------|-----------------|
| Chart cursor navigation | `internal/ui/cards.go` + `internal/common/charts/line.go` | ← → keys trigger `LineChart.CursorAt(index int) string` for tooltip |
| Card focus ring | `internal/ui/cards.go` | `CardGrid.NextCard()/PrevCard()` — Tab/shift-tab moves focus; focused card has highlighted border |
| Drill-down overlay | `internal/ui/cards.go` | Enter on card → `DetailPanel` overlay that captures input until `esc` dismisses |
| Split-pane support | `internal/ui/cards.go` | `|` key on focused card toggles `SplitPane`; `[` `]` adjust ratio |
| Context menu | `internal/ui/cards.go` | `ctrl+space` opens contextual action list for focused card |

**Verification**: `go build ./...` + manual walkthrough all 5 layers + dashboard + card navigation

**Files changed in Sprint 2**:
```
NEW:  internal/ui/dashboard.go               — DashboardModel, health gauges, operation cards, ASCII header
NEW:  internal/common/alerts.go               — Alert/incident types, flap detection, alert history ring buffer
MOD:  internal/ui/root.go                     — ScreenDashboard routing, DashboardModel init
MOD:  internal/common/types.go                — Add ScreenDashboard to Screen enum
MOD:  internal/ui/mainmenu.go                 — Simplified, secondary role
MOD:  internal/ui/styles.go                   — Dashboard, card grid, gauge card primitives
MOD:  internal/ui/statusbar.go                — Alert count pulse, forecast indicators in status text
MOD:  internal/ui/keys.go                     — Add card navigation keys (Tab, ← →, |, ctrl+space)
MOD:  internal/ui/help.go                     — Context-sensitive help for card navigation
MOD:  internal/sysops/view.go                 — Full rewrite with cards, charts, forecasts
MOD:  internal/netops/view.go                 — Full rewrite with cards, progress, throughput charts
MOD:  internal/secops/view.go                 — Full rewrite with status cards, port heatmap
MOD:  internal/devops/view.go                 — Full rewrite with session cards, tail highlights
MOD:  internal/aiops/view.go                  — Full rewrite with chat cards, insight panels
```

---

## Sprint 3: Intelligence & Forecasting

**Goal**: Make the platform predictive, not just reactive. Forecast trends, predict anomalies, surface insights.
**Dependencies**: Sprint 1 (forecast engine, chart overlays) + Sprint 2 (dashboard overlay hooks)
**Parallel lanes**: C (Intelligence) after Sprint 1A-1B; D (AI Insights) after Sprint 3.1

### 3.1 Predictive Analytics Engine

| Task | Files | Concrete Change |
|------|-------|-----------------|
| CPU trend forecast | `internal/sysops/view.go` + `internal/common/forecast.go` | Call `forecast.LinearRegression(store.Get("cpu"))` → "CPU projected to reach 90% in 12m" |
| Memory exhaustion predictor | `internal/sysops/view.go` + `internal/common/forecast.go` | `forecast.TimeToThreshold(store.Get("mem"), 95.0)` → "Memory full in 2h 34m" |
| Disk full predictor | `internal/sysops/view.go` + `internal/common/forecast.go` | Based on disk usage rate over last 24h → "Disk full in 3.2 days" |
| Network bandwidth trend | `internal/netops/view.go` + `internal/common/forecast.go` | "Traffic up 12% week-over-week" via trend slope |
| Anomaly prediction | `internal/common/forecast.go` + `internal/aiops/ollama.go` | Combined threshold + statistical outlier → "Anomaly likely in next 5m (confidence: 78%)" |

### 3.2 Alerting & Incidents

| Task | File(s) | Concrete API/Detail |
|------|---------|---------------------|
| Flap detection | `internal/common/alerts.go` | `type FlapDetector struct { threshold int; recovery int; consecutive int }` — N consecutive misses → alert, M replies → recover |
| Alert struct | `internal/common/alerts.go` | `type Alert struct { ID string; Severity AlertSeverity; Message string; Timestamp time.Time; ResolvedAt *time.Time }` |
| Alert severity | `internal/common/alerts.go` | `type AlertSeverity int` — `AlertInfo`, `AlertWarning`, `AlertCritical` |
| Alert history | `internal/common/alerts.go` | Ring buffer of alerts, `AddAlert(a Alert)`, `Unresolved() []Alert` |
| Status bar pulse | `internal/ui/statusbar.go` | Pulsing alert count with severity color via `m.stats.AnomalyCount` → animated indicator |
| Alert detail view | `internal/ui/alertview.go` | Scrollable alert timeline overlay, `type AlertView struct { Alerts []Alert; selected int }` |
| Alert persistence | `internal/common/alerts.go` | Save/load alert history to JSON in `ConfigDir()` |

### 3.3 AI-Powered Insights (Leveraging Existing Ollama)

| Task | Files | Concrete Change |
|------|-------|-----------------|
| Stats-to-text | `internal/aiops/ollama.go` | New `SummarizeSystemState(stats *SystemStats) string` — calls existing Ollama chat API with prompt template |
| Trend explanation | `internal/aiops/ollama.go` | New `ExplainTrend(metric string, data []DataPoint) string` — "Memory usage rising due to Chrome (PID 1234)" |
| Recommended actions | `internal/aiops/ollama.go` | New `GetRecommendations(stats *SystemStats, trends []TrendDirection) string` — "Consider closing 3 idle processes to free 2GB" |
| Forecast narrative | `internal/aiops/ollama.go` | New `NarrateForecast(metric string, forecast []DataPoint) string` — "At this rate, disk will be full Saturday morning" |

**Note**: All AI functions use the existing `internal/aiops/ollama.go` HTTP client to `localhost:11434/api/chat`. No new dependencies.

### 3.4 Forecast Visualization

| Task | Files | Concrete Change |
|------|-------|-----------------|
| Forecast overlay | `internal/common/charts/line.go` | `LineChart.SetForecast(data []DataPoint)` — renders predicted values with different braille glyphs or color |
| Confidence band | `internal/common/charts/line.go` | Shaded area using dimmed characters around forecast line |
| Threshold marker | `internal/common/charts/line.go` | `SetThreshold(threshold float64, label string)` — vertical dashed line at projected crossing point |
| What-if toggle | `internal/common/forecast.go` | `WhatIfProjection(data []DataPoint, slopeModifier float64) []DataPoint` — "If trend continues..." view |

**Verification**: `go test ./internal/common/forecast_test.go` + `go test ./internal/common/... -run Forecast` + visual inspection

**Files changed in Sprint 3**:
```
NEW:  internal/ui/alertview.go                — Alert timeline overlay view
MOD:  internal/common/alerts.go               — (created Sprint 2) Add flap detection, alert history
MOD:  internal/common/forecast.go             — Add prediction narrative strings, what-if projection
MOD:  internal/common/charts/line.go          — Add forecast overlay, confidence interval band, threshold marker
MOD:  internal/aiops/ollama.go                — Add SummarizeSystemState, ExplainTrend, GetRecommendations, NarrateForecast
MOD:  internal/sysops/view.go                 — Wire CPU/MEM/DISK forecasts into gauge cards
MOD:  internal/netops/view.go                 — Wire bandwidth forecast into throughput chart
MOD:  internal/ui/statusbar.go                — Pulse animation for alert count
MOD:  internal/ui/dashboard.go                — Forecast overlay on health gauges, alert section
```

---

## Sprint 4: Search, Navigation & Power User Features

**Goal**: Make the platform fast to navigate and deeply searchable.
**Dependencies**: Sprint 2 (all views exist in card form)
**Parallel lanes**: E (Power User) after Sprint 2B

### 4.1 Global Command Palette

| Task | File(s) | Concrete API/Detail |
|------|---------|---------------------|
| Text input overlay | `internal/ui/commandpalette.go` | `/` opens search bar at top of screen; text input with real-time filtering |
| Fuzzy search | `internal/ui/commandpalette.go` | `func SearchOps(query string, ops []Operation) []Operation` — match by name, description, keyword, tag |
| Operation registry | `internal/ui/commandpalette.go` | `var OperationRegistry []Operation` — registered at init by each ops layer |
| Operation struct | `internal/ui/commandpalette.go` | `type Operation struct { Name, Description string; Tags []string; Screen Screen; Action func() tea.Cmd }` |
| Result rendering | `internal/ui/commandpalette.go` | List of matched ops with tag badges; Enter navigates there |
| Search history | `internal/ui/commandpalette.go` | `var searchHistory []string` — ↑↓ cycles through recent searches |
| Tag filtering | `internal/ui/commandpalette.go` | `t` key in search → `TagFilter` overlay → filter by tag (hackingtool-inspired) |
| Recommend mode | `internal/ui/commandpalette.go` | `r` key → `Recommend(query string) []Operation` — natural language → matching ops |

**Key binding change**: `/` currently maps to `key.WithHelp("/", "filter")` bound to the `Filter` action. In v3, `/` opens the command palette on *any* screen, not just as a filter. The existing filter per-screen becomes a secondary action.

### 4.2 Keyboard Workflow Engine

| Task | File(s) | Concrete Change |
|------|---------|-----------------|
| Rebindable keys | `internal/common/config.go` + `internal/ui/keys.go` | Add `KeyBindings map[string]string` to `Config`; `RebuildKeyMap(cfg *Config) KeyMap` |
| Vim-mode enhancements | `internal/ui/keys.go` | Add `gg` (top), `G` (bottom), `/` (search-forward), `n`/`N` (next/prev result) |
| Screen bookmarks | `internal/ui/keys.go` | Add `BookmarkCurrent()`, `GoToBookmark(int)` — `ctrl+s` saves, `ctrl+b` shows list |
| Macro recording | `internal/ui/keys.go` | `RecordMacro()`, `PlayMacro()`, `StopMacro()` — records key sequence, stores in `[]string` |
| Context-sensitive help | `internal/ui/help.go` | `HelpOverlay.ShowForScreen(screen Screen)` — only shows shortcuts relevant to current screen |

### 4.3 Data Drill-Down System

| Task | Files | Concrete Change |
|------|-------|-----------------|
| Expandable table rows | `internal/ui/cards.go` | `TableCard.ExpandRow(index int)` — Enter on row slides open a detail pane below |
| Process detail | `internal/sysops/view.go` | `renderProcessDetail(pid int)` → `DetailPanel` with threads, handles, memory map, open files |
| Connection detail | `internal/netops/view.go` | `renderConnDetail(conn Connection)` → DNS resolution, process name, bandwidth |
| Port scan detail | `internal/netops/view.go` | `renderPortDetail(port int)` → banner grab, service version |
| User detail | `internal/secops/view.go` | `renderUserDetail(username string)` → groups, privileges, last logon |
| Log tail detail | `internal/devops/view.go` | `renderLogContext(lineNum int)` → full context around matched line in file |

### 4.4 Data Export

| Task | Files | Concrete API/Detail |
|------|-------|---------------------|
| Export view to JSON | All `view.go` files | `func (m *Model) ExportJSON() (string, error)` on each ops model |
| Export view to CSV | All `view.go` files | `func (m *Model) ExportCSV() (string, error)` on each ops model |
| Chart as ASCII | `internal/common/charts/*.go` | `func (lc *LineChart) ExportASCII() string` — renders chart to static string for sharing |
| Dashboard snapshot | `internal/ui/dashboard.go` | `func (d *DashboardModel) Snapshot() string` — captures all cards + charts as ASCII |
| Scheduled export | `internal/common/timeseries.go` | Periodic export via configurable interval in config |

**Verification**: `go build ./...` + manual test of `/`, `t`, `r`, drill-down on each layer

**Files changed in Sprint 4**:
```
NEW:  internal/ui/commandpalette.go           — CommandPaletteModel, text input, SearchOps, OperationRegistry
MOD:  internal/ui/root.go                     — Command palette integration, key rerouting when open
MOD:  internal/ui/keys.go                     — Rebindable keys, vim-mode, bookmarks, macro recording
MOD:  internal/ui/help.go                     — Context-sensitive help per screen
MOD:  internal/ui/cards.go                    — Expandable table rows, TableCard.SortByColumn
MOD:  internal/common/config.go               — KeyBindings in Config struct
MOD:  internal/common/types.go                — Add Operation struct for registry
MOD:  All view.go files                       — Add ExportJSON/ExportCSV methods
MOD:  internal/common/charts/*.go             — Add ExportASCII methods
```

---

## Sprint 5: Config, Polish & Cross-Platform

**Goal**: Persistent settings, responsive layouts, platform hardening, performance optimization.
**Dependencies**: All previous sprints
**Parallel lanes**: G (Polish) after all predecessors

### 5.1 YAML Config System (completed files from Sprint 1.5)

| Task | File(s) | Concrete Change |
|------|---------|-----------------|
| Config UI | `internal/ui/settings.go` | In-app settings editor — theme picker, key rebind UI, interval sliders |
| Config integration | `internal/ui/root.go` | `LoadConfig()` on startup, `SaveConfig()` on changes |

### 5.2 Responsive Layout Engine

| Task | File(s) | Concrete Change |
|------|---------|-----------------|
| Auto column count | `internal/ui/cards.go` | `CardGrid.AutoColumns(width int) int` — narrow → 1 col, medium → 2, wide → 3+ |
| Min card width | `internal/ui/cards.go` | Cards never squash below 30-character minimum; wrap to new column |
| Chart width adaptation | `internal/common/charts/*` | Charts accept `Width` in `ChartConfig` and auto-scale to available width |
| Collapse mode | `internal/ui/cards.go` | In very narrow terminals (<80 cols), cards become compact sparklines with no title |

### 5.3 Performance Optimization

| Task | File(s) | Concrete Change |
|------|---------|-----------------|
| Render profiling | `internal/ui/root.go` | Add `renderTime` tracking per layer; log slow renders via `common.LogWarn` |
| Throttled updates | `internal/ui/root.go` | Skip `View()` if render took longer than refresh interval in previous tick |
| Lazy chart rendering | `internal/common/charts/*` | Cache rendered output; only re-render when `SetValues`/`AddSeries` changes data |
| History pruning | `internal/common/timeseries.go` | Auto-prune data points older than configured max age (default: 1 hour) |

### 5.4 Cross-Platform & Hardening

| Task | File(s) | Concrete Change |
|------|---------|-----------------|
| Linux user parsing | `internal/secops/users.go` | Parse `/etc/passwd` and `/etc/group` for user/group audit (currently Windows-only via `net.exe`) |
| macOS `dscl` support | `internal/secops/users.go` | Execute `dscl . list /Users` as fallback for macOS |
| Ollama graceful degradation | `internal/aiops/ollama.go` | Retry with backoff (3 retries, exponential), show offline status card if unreachable |
| Cross-platform build | `go.mod` / `scripts/` | Ensure `GOOS=linux GOARCH=amd64 go build` and `GOOS=darwin GOARCH=amd64 go build` succeed |

### 5.5 Session Logging & Audit

| Task | File(s) | Concrete Change |
|------|---------|-----------------|
| Log rotation | `internal/common/logger.go` | Existing `InitLogger()` enhanced with max size (10MB default), auto-rotate |
| Log levels | `internal/common/logger.go` | Existing: `LogInfo`, `LogWarn`, `LogError` — all present |
| In-app log viewer | `internal/ui/logviewer.go` | `type LogViewer struct { entries []LogEntry; search string; filter Level }` — tail session log in TUI |
| Operation audit trail | Internal all `update.go` files | Each ops layer's `update.go` calls `common.LogInfo("Operation: %s (duration: %v)", action, duration)` |

**Verification**: `go build ./...` + `GOOS=linux GOARCH=amd64 go build` + `go test -race ./...` + manual config test

**Files changed in Sprint 5**:
```
NEW:  internal/ui/settings.go                — In-app config editor
NEW:  internal/ui/logviewer.go               — In-app session log viewer
MOD:  internal/ui/cards.go                   — Responsive layout, auto-column, collapse mode
MOD:  internal/ui/root.go                    — Config load on startup, render throttling
MOD:  internal/common/config.go              — (completed) Config UI integration
MOD:  internal/common/logger.go              — Log rotation, max size
MOD:  internal/common/charts/*.go            — Lazy rendering, width adaptation
MOD:  internal/common/timeseries.go          — Auto-pruning, periodic export
MOD:  internal/secops/users.go               — Linux /etc/passwd, macOS dscl
MOD:  internal/aiops/ollama.go               — Retry with backoff
MOD:  All update.go files                    — Audit trail logging
```

---

## Dependency Graph

```
Sprint 1 (Visualization Engine)
  ├─ 1.1-1.4 Chart Library + TimeSeries + Forecast + Cards
  └─ 1.5 Config System (independent — lane F)
       │
       ├──────────────────────┐
       ▼                      ▼
Sprint 2 (UI Redesign)     Sprint 3 (Intelligence)
  ├─ 2.1 Dashboard            ├─ 3.1 Predictions
  ├─ 2.2-2.5 Layer Views      ├─ 3.2 Alerting
  └─ 2.6 Interaction Layer    ├─ 3.3 AI Insights
                               └─ 3.4 Forecast Viz
       │                      │
       └──────────┬───────────┘
                  ▼
          Sprint 4 (Power User)
            ├─ 4.1 Command Palette
            ├─ 4.2 Keyboard Engine
            ├─ 4.3 Drill-Down
            └─ 4.4 Export
                  │
                  ▼
          Sprint 5 (Polish)
            ├─ 5.1 Config UI
            ├─ 5.2 Responsive
            ├─ 5.3 Performance
            ├─ 5.4 Cross-Platform
            └─ 5.5 Session Logging
```

---

## Parallel Execution Lanes

| Lane | Sprint | Dependencies | Isolated Scope | Files |
|------|--------|-------------|---------------|-------|
| **A — Charts** | Sprint 1 (1.1-1.4) | None | `internal/common/charts/*`, `internal/common/timeseries.go`, `internal/common/forecast.go`, `internal/ui/cards.go` | ~12 new files |
| **B — UI** | Sprint 2 | Sprint 1A | `internal/ui/dashboard.go`, `internal/*/view.go`, `internal/common/alerts.go` | ~1 new + 8 modified |
| **C — Intelligence** | Sprint 3 | Sprint 1A-1B | `internal/common/forecast.go` (extend), `internal/common/alerts.go` (extend) | ~1 new + 6 modified |
| **D — AI Insights** | Sprint 3.3 | Sprint 3.1 | `internal/aiops/ollama.go` (extend) | 1 modified |
| **E — Power User** | Sprint 4 | Sprint 2B | `internal/ui/commandpalette.go`, `internal/ui/keys.go` | ~1 new + 8 modified |
| **F — Config** | Sprint 1 (1.5) | None | `internal/common/config.go` (independent) | 1 new |
| **G — Polish** | Sprint 5 | Sprints 1-4 | All packages | ~2 new + 12 modified |

**Maximum parallelism**: Lanes A, F can start immediately. Lanes B, C, E after A. Lanes D, G after multiple predecessors.

---

## File Change Summary

### New Packages and Files
```
internal/common/charts/          — Chart library (8 files)
  config.go, line.go, bar.go, area.go, gauge.go, sparkline.go, heatmap.go, number.go
internal/common/timeseries.go    — Ring buffer, TimeSeriesStore, WindowStats
internal/common/forecast.go      — LinearRegression, ExponentialSmoothing, TrendDirection
internal/common/config.go        — YAML Config struct, load/save
internal/common/alerts.go        — Alert/incident types, flap detection
internal/ui/dashboard.go         — Dashboard landing page (replaces main menu as default)
internal/ui/cards.go             — Card, CardGrid, GaugeCard, ChartCard, TableCard, DetailPanel, SplitPane
internal/ui/commandpalette.go    — Search overlay, tag filter, recommend mode
internal/ui/settings.go          — In-app settings editor
internal/ui/logviewer.go         — In-app session log viewer
internal/ui/alertview.go         — Alert timeline view
```

### Files to Modify Heavily
```
internal/ui/root.go              — ScreenDashboard routing, command palette integration, config load
internal/ui/styles.go            — Dashboard, card, gauge, badge, chart overlay style primitives
internal/ui/mainmenu.go          — Simplified, becomes secondary nav
internal/ui/statusbar.go         — Alert count pulse, forecast indicators
internal/ui/help.go              — Context-sensitive help sections
internal/ui/keys.go              — Dynamic key map, bookmark, macro, rebindable keys
internal/common/theme.go         — Extend Palette with CardBg, ChartLine1-3, ChartGrid fields
internal/common/styles.go        — Add chart-card styled primitives
internal/common/types.go         — Add ScreenDashboard, TrendDirection, CardStatus, Alert, Operation types

internal/sysops/view.go          — Full rewrite with gauge cards, charts, forecasts, drill-down
internal/netops/view.go          — Full rewrite with latency charts, progress gauges, throughput graphs
internal/secops/view.go          — Full rewrite with status cards, port heatmap, rule cards
internal/devops/view.go          — Full rewrite with session cards, log tail with highlights
internal/aiops/view.go           — Full rewrite with chat cards, insight cards, status cards

internal/aiops/ollama.go         — Retry with backoff, trend analysis prompts, system summarization
internal/secops/users.go         — Linux /etc/passwd, macOS dscl support
internal/common/logger.go        — Log rotation, max size enhancement
```

### Files to Create Modifications (New Logic)
```
internal/sysops/                 — CPU core breakdown, memory breakdown in view.go
internal/netops/                 — CIDR scanner (new scanner.go), connection heat map in view.go
```

---

## Verification Gates

| Gate | Command | Sprint | Expected Outcome |
|------|---------|--------|-----------------|
| Build | `go build -o hawkward.exe ./cmd/hawkward` | After every sprint | Exit code 0, binary created |
| Vet | `go vet ./...` | After every sprint | No warnings |
| Format | `gofmt -d .` | After every sprint | No diffs |
| Chart unit tests | `go test ./internal/common/charts/...` | Sprint 1 | All chart rendering tests pass |
| TimeSeries tests | `go test ./internal/common/... -run TimeSeries` | Sprint 1 | Ring buffer, window stats, persistence pass |
| Forecast tests | `go test ./internal/common/... -run Forecast` | Sprint 3 | Linear regression, smoothing, trend pass |
| Alert tests | `go test ./internal/common/... -run Alert` | Sprint 3 | Flap detection, history pass |
| Card tests | `go test ./internal/ui/... -run Card` | Sprint 2 | CardGrid layout, focus, drill-down pass |
| Command palette tests | `go test ./internal/ui/... -run CommandPalette` | Sprint 4 | Search, filtering, history pass |
| All tests | `go test ./...` | After every sprint | All pass |
| Race detector | `go test -race ./...` | After sprints 2, 4, 5 | No races |
| Manual TUI | Walk through all 5 layers + dashboard + search | After each sprint | Visual verification |
| Cross-platform | `GOOS=linux GOARCH=amd64 go build` | Sprint 5 | Linux build succeeds |
| Cross-platform | `GOOS=darwin GOARCH=amd64 go build` | Sprint 5 | macOS build succeeds |
| Coverage | `go test -coverprofile=coverage.out ./...` | After sprint 5 | ≥30% coverage per package |

---

## Visual Design Reference

### Color System (Squib-Inspired)

Extended `common.Palette` struct with chart-specific fields:

```go
type Palette struct {
    // Existing fields (Sprint 1-2 compatible)
    Primary, Secondary, Warning, Danger, Info string
    Muted, Text, Background, Border          string
    
    // Sprint 3+ additions
    ChartLine1  string  // First series color (used in LineChart, AreaChart)
    ChartLine2  string  // Second series color
    ChartLine3  string  // Third series / forecast color
    ChartGrid   string  // Grid line color
    CardBg      string  // Card background color
}
```

```
Dark Theme (squib-dark):
  Background: #0f172a (slate-900)
  Card bg:    #1e293b (slate-800)
  Primary:    #38bdf8 (sky-400)
  Secondary:  #4ade80 (green-400)
  Warning:    #fbbf24 (amber-400)
  Danger:     #f87171 (red-400)
  Info:       #a78bfa (violet-400)
  Text:       #f8fafc (slate-50)
  Muted:      #94a3b8 (slate-400)
  Border:     #334155 (slate-700)
  ChartLine1: #38bdf8 (sky-400 → CPU/primary series)
  ChartLine2: #a78bfa (violet-400 → memory/secondary series)
  ChartLine3: #fbbf24 (amber-400 → forecast/prediction)
  ChartGrid:  #1e293b (slate-800 → subtle grid)
  CardBg:     #1e293b (slate-800)

Light Theme (squib-light):
  Background: #f8fafc (slate-50)
  Card bg:    #ffffff (white)
  Primary:    #0ea5e9 (sky-500)
  Secondary:  #22c55e (green-500)
  Warning:    #eab308 (yellow-500)
  Danger:     #ef4444 (red-500)
  Info:       #8b5cf6 (violet-500)
  Text:       #0f172a (slate-900)
  Muted:      #64748b (slate-500)
  Border:     #e2e8f0 (slate-200)
  ChartLine1: #0ea5e9 (sky-500)
  ChartLine2: #8b5cf6 (violet-500)
  ChartLine3: #eab308 (yellow-500)
  ChartGrid:  #e2e8f0 (slate-200)
  CardBg:     #ffffff (white)
```

### Theme Expansion Plan
Base 4 themes (existing) → 10 themes:
1. `default` — Purple/green dark (current)
2. `dark` — Purple/green dark variant (current)
3. `light` — Light mode (current)
4. `high-contrast` — Accessibility (current)
5. `squib-dark` — Slate/sky dark (inspired by squib)
6. `squib-light` — Slate/sky light (inspired by squib)
7. `amber` — Amber-on-black retro terminal
8. `green` — Green-on-black classic terminal
9. `dracula` — Dracula-inspired
10. `nord` — Nord-inspired

### Typography Hierarchy
```
Header / Title: Bold, Primary color, larger padding
Card Title:    Bold, Text color, compact
Section:       Bold, Info color
Label:         Muted color, normal weight
Value:         Bold, Text color
Muted:         Muted color, dim
Error:         Danger color, bold
Good:          Success color (Secondary)
Warning:       Warning color
```

### Icon Set for Operation Cards
```
🖥  SysOps     → CPU, memory, disk, processes
🌐  NetOps     → Ping, DNS, ports, connections, interfaces
🔒  SecOps     → Firewall, users, defender, tasks
⚙  DevOps     → Shell, logs, files, tools
🤖  AIOps      → Chat, reports, insights
📊  Dashboard  → Overview, health, forecasts
🔍  Search     → Command palette, filter
📈  Trend      → Forecast, prediction
⚠  Alert      → Anomaly, incident
```

---

## Recommended Execution Order

```
Week 1-2:  Sprint 1 — Visualization Engine (Foundation)
           - 1.1 Chart library (Line, Bar, Gauge, Sparkline, Area, HeatMap, Numeric) — Lane A
           - 1.2 Time-series store — Lane A
           - 1.3 Forecast engine basics — Lane A
           - 1.4 Card component system — Lane A
           - 1.5 Config system (parallel) — Lane F

Week 3-4:  Sprint 2.1-2.2 — Dashboard + SysOps
           - Dashboard landing page with live gauges — Lane B
           - SysOps view with charts and forecasts — Lane B
           - First visible "wow" moment

Week 5-6:  Sprint 2.3-2.6 — Remaining Layer Views
           - NetOps, SecOps, DevOps, AIOps card rewrites — Lane B
           - Interaction layer (cursor nav, drill-down) — Lane B

Week 7:    Sprint 3 — Intelligence & Forecasting
           - 3.1 Predictive analytics on all metrics — Lane C
           - 3.2 Alerting system with flap detection — Lane C
           - 3.3 AI-powered insights via Ollama — Lane D
           - 3.4 Forecast visualization — Lane C

Week 8:    Sprint 4 — Search & Power User Features
           - 4.1 Command palette, tag filter, recommend mode — Lane E
           - 4.2 Keyboard workflow (macros, bookmarks, rebind) — Lane E
           - 4.3 Drill-down detail overlays — Lane E
           - 4.4 Data export — Lane E

Week 9:    Sprint 5 — Polish & Cross-Platform
           - 5.1 Config UI — Lane G
           - 5.2 Responsive layout — Lane G
           - 5.3 Performance optimization — Lane G
           - 5.4 Cross-platform hardening — Lane G
           - 5.5 Session logging — Lane G
```

This plan prioritizes **visible progress early** (Sprint 1 → Sprint 2 gives dramatic before/after in 3-4 weeks) while building toward the deeper intelligence features in later sprints.

---

## Risk Register

| Risk | Impact | Likelihood | Mitigation |
|------|--------|-----------|------------|
| Braille charts look poor in some terminals | Medium | Low | Add block-char fallback; test in Windows Terminal, iTerm2, Alacritty |
| gopsutil v4 breaking changes | High | Low | Pin version in go.mod; test early in Sprint 1 |
| Forecast accuracy too low for useful predictions | Medium | Medium | Use conservative framing ("trending upward"); don't claim precision |
| Card grid doesn't fit terminal widths well | Medium | Medium | Test at 80, 100, 120, 160 col widths; collapse mode at <80 |
| Command palette perf with many operations | Low | Low | Limit fuzzy search to registered operations (~50-100 max) |
| Cross-platform: wmi/netstat differences | Medium | Low | Existing tests cover; add CI matrix for Windows/Linux/macOS |

---

*Last updated: 2026-07-07*
*Owner: Planning & Documentation Lead*
*Next review: After Sprint 1 delivery*
