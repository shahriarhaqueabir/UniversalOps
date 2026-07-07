# Hawkward GUI Overhaul — Wails v2 Desktop Application

> **Theme**: From TUI (Bubble Tea) to native cross-platform desktop GUI
> **Goal**: Reuse all existing Go backend packages, wrap them in Wails v2 Go bindings, and build a modern React/TypeScript frontend with shadcn/ui
> **Timeline**: 5 sprints (~10 weeks)
> **Go**: 1.26.4 | **GUI Framework**: Wails v2 | **Frontend**: React 18 + TypeScript + Vite + Tailwind CSS v4 + shadcn/ui

---

## Why Wails v2

| Feature | Benefit |
|---------|---------|
| **Go backend** | Reuse ALL existing packages — sysops, netops, secops, devops, aiops, pipeline, forecast, alerts, timeseries, charts (chart lib becomes shared utility, TUI-specific chart rendering is replaced by Recharts) |
| **Web frontend** | React + TypeScript + Tailwind + shadcn/ui for modern, responsive, composable UI |
| **Real-time events** | Built-in Wails Events system pushes live data from Go to frontend (dashboard ticks, alert firings, log tailing) |
| **Native features** | File dialogs, system tray, native menus, notifications — all via Wails runtime |
| **Charts** | Recharts for rich interactive SVG charts (replaces terminal Unicode charts) |
| **Network vis** | vis-network or cytoscape.js for interactive network topology (replaces ASCII art) |
| **Cross-platform** | Windows (WebView2), macOS (WebKit), Linux (WebKit) — single Go binary + frontend assets |

---

## Project Structure

```
AllOpsFull/
├── cmd/
│   └── hawkward-gui/
│       └── main.go                  # Wails app entry point — init App, start window
├── internal/
│   ├── common/                      # REUSE EXISTING (unchanged except removing Bubble Tea dep)
│   │   ├── types.go                 # Keep: SystemStats, Screen enum (for backend ref), shared structs
│   │   ├── formatters.go            # Keep: pretty-print, Markdown formatting
│   │   ├── platform.go             # Keep: OS detection
│   │   ├── sandbox.go              # Keep: sandboxed command interface
│   │   ├── pipeline.go             # Keep: data pipeline runner
│   │   ├── timeseries.go           # Keep: ring-buffer time-series store
│   │   ├── forecast.go             # Keep: linear regression + exponential smoothing
│   │   ├── alerts.go               # Keep: alert engine with flap detection
│   │   ├── metrics.go              # Keep: metric aggregation
│   │   ├── logger.go               # Keep: logging system
│   │   ├── theme.go                # Keep: palette (for backend chart colors)
│   │   ├── styles.go               # REMOVE/ARCHIVE: Lipgloss styles (TUI-specific)
│   │   └── charts/                 # ARCHIVE: TUI Unicode charts (replaced by Recharts)
│   ├── sysops/                      # REUSE EXISTING (collector, cpu, disk, memory, processes, system)
│   ├── netops/                      # REUSE EXISTING (ping, dns, portscan, traceroute, connections, interfaces)
│   ├── secops/                      # REUSE EXISTING (firewall, users, defender, listening, events, tasks)
│   ├── devops/                      # REUSE EXISTING (shell, logtail, filebrowser, processes, services)
│   ├── aiops/                       # REUSE EXISTING (ollama, reporting, state_query)
│   ├── ui/                          # ARCHIVE: Bubble Tea TUI (all model/update/view files)
│   └── app/                         # NEW: Wails app bindings — Go structs exposed to frontend JS
│       ├── App.go                   # Root app struct: startup, lifecycle, event bus init
│       ├── Dashboard.go            # Dashboard data: system stats, health summary, sparkline data
│       ├── SysOps.go               # CPU, memory, disk, processes, system info
│       ├── NetOps.go               # Ping, DNS, port scan, traceroute, connections, interfaces
│       ├── SecOps.go               # Firewall rules, user audit, listening ports, defender, events
│       ├── DevOps.go               # Shell executor, log tailer, file browser, process manager, services
│       ├── AIOps.go                # Ollama chat, report generation, state query
│       ├── Pipeline.go             # Pipeline CRUD, execution, status
│       ├── Alerts.go               # Alert rules CRUD, fired alerts, silence/ack
│       ├── Logs.go                 # Log aggregation, filtering, export
│       ├── Events.go               # Real-time event emitter (Wails Events for live push)
│       └── Settings.go             # App settings, theme config, refresh rates
├── frontend/                        # NEW: React + TypeScript + Vite
│   ├── src/
│   │   ├── main.tsx                # Entry: render <App />, init Wails runtime
│   │   ├── App.tsx                 # Root: routing, layout shell, event listeners
│   │   ├── components/
│   │   │   ├── ui/                 # shadcn/ui primitives (button, card, dialog, input, select, table, tabs, badge, etc.)
│   │   │   ├── layout/            # Sidebar, TopBar, ContentArea, StatusBar
│   │   │   ├── dashboard/         # Dashboard: health cards, gauges, sparklines, alerts bar, forecast snippets
│   │   │   ├── sysops/            # SysOps: CPU gauge, memory gauge, disk gauge, process table, system info cards, forecast charts
│   │   │   ├── netops/            # NetOps: ping panel, DNS lookup, port scanner, traceroute, connection table, interface list
│   │   │   ├── secops/            # SecOps: firewall table, user table, listening ports, defender status, security events
│   │   │   ├── devops/            # DevOps: command terminal, log viewer, file browser tree, process manager, services list
│   │   │   ├── aiops/             # AIOps: chat panel, report viewer, anomaly display
│   │   │   ├── network/           # Network topology designer (vis-network)
│   │   │   ├── charts/            # Recharts wrappers: LineChart, BarChart, AreaChart, Gauge, Sparkline
│   │   │   ├── dialogs/           # Export dialog, Settings dialog, About dialog
│   │   │   └── logs/              # Log viewer with search, filter, export
│   │   ├── hooks/                  # Custom React hooks
│   │   │   ├── useEvents.ts       # Wails Events listener hook
│   │   │   ├── useBackend.ts      # Typed wrappers around Wails backend calls
│   │   │   ├── useTheme.ts        # Theme toggle + persistence
│   │   │   └── useInterval.ts     # Polling interval hook
│   │   ├── lib/
│   │   │   ├── types.ts           # Shared TypeScript interfaces mirroring Go types
│   │   │   ├── formatters.ts      # Date, bytes, percentage formatters
│   │   │   ├── constants.ts       # Screen names, refresh rates, limits
│   │   │   └── utils.ts           # cn() classname merger, misc utilities
│   │   └── styles/
│   │       ├── globals.css        # Tailwind directives, CSS variables, theme tokens
│   │       └── transitions.css    # Page transitions, animation keyframes
│   ├── package.json
│   ├── vite.config.ts
│   ├── tailwind.config.ts
│   ├── tsconfig.json
│   ├── tsconfig.node.json
│   └── index.html
├── wails.json                       # Wails project configuration
└── go.mod                           # Updated: add Wails v2 dependency, remove Bubble Tea/Lipgloss
```

---

## Phases of Work

### Phase 1: Scaffold & Backend Bindings (Sprint 1)

**Objective**: Project skeleton, Wails init, all Go bindings, frontend scaffold with routing shell.

#### Steps

1. **Initialize Wails project**
   - Run `wails init -n hawkward-gui -t react-ts` in a temp dir, then copy the skeleton in
   - Create `cmd/hawkward-gui/main.go` with standard Wails app setup
   - Create `internal/app/` package directory structure

2. **Create `internal/app/App.go` — Root app struct**
   ```go
   // App is the main application struct bound to the frontend.
   // Wails calls exported methods directly from JavaScript.
   type App struct {
       ctx        context.Context
       eventBus   *Events
       settings   *Settings
       logger     *Logs
       alerts     *Alerts
       pipeline   *Pipeline
       sysOps     *SysOps
       netOps     *NetOps
       secOps     *SecOps
       devOps     *DevOps
       aiOps      *AIOps
       dashboard  *Dashboard
   }
   func NewApp() *App { ... }
   func (a *App) Startup(ctx context.Context) { ... }
   func (a *App) Shutdown(ctx context.Context) { ... }
   ```

3. **Create binding files** — One struct per ops layer, each with exported methods:
   - `Dashboard.go` — `GetSystemStats()`, `GetHealthSummary()`, `GetSparklineData()`
   - `SysOps.go` — `GetCPU()`, `GetMemory()`, `GetDisk()`, `GetProcesses()`, `GetSystemInfo()`
   - `NetOps.go` — `Ping(host)`, `DNSLookup(host)`, `PortScan(host, ports)`, `Traceroute(host)`, `GetConnections()`, `GetInterfaces()`
   - `SecOps.go` — `GetFirewallRules()`, `GetUsers()`, `GetListeningPorts()`, `GetDefenderStatus()`, `GetSecurityEvents()`
   - `DevOps.go` — `ExecCommand(cmd)`, `TailLog(path)`, `BrowseFiles(path)`, `GetProcesses()`, `GetServices()`
   - `AIOps.go` — `Chat(message)`, `GenerateReport(layer)`, `QueryState(query)`
   - `Pipeline.go` — `CreatePipeline(config)`, `RunPipeline(id)`, `GetPipelineStatus(id)`, `ListPipelines()`
   - `Alerts.go` — `GetAlertRules()`, `CreateAlertRule(rule)`, `GetFiredAlerts()`, `AcknowledgeAlert(id)`
   - `Logs.go` — `GetLogs(filter)`, `ExportLogs(format)`
   - `Events.go` — Event emitter using `runtime.EventsEmit()` for real-time pushes
   - `Settings.go` — `GetSettings()`, `UpdateSettings(settings)`, `GetTheme()`, `SetTheme(theme)`

4. **Scaffold React frontend**
   - `npm create vite@latest frontend -- --template react-ts`
   - Install deps: `tailwindcss`, `@tailwindcss/vite`, `lucide-react`, `recharts`, `vis-network`, `@radix-ui/*` primitives
   - Initialize shadcn/ui (`npx shadcn@latest init`)
   - Add shadcn components: button, card, dialog, input, select, table, tabs, badge, separator, scroll-area, sheet, dropdown-menu, switch, toast

5. **Create layout shell**
   - `Sidebar.tsx` — Collapsible sidebar with navigation links (Dashboard, SysOps, NetOps, SecOps, DevOps, AIOps), icon + label
   - `TopBar.tsx` — Current screen title, time, theme toggle, search quick-access
   - `ContentArea.tsx` — Router for active screen
   - `App.tsx` — Composes layout, initializes Wails runtime, registers global event listeners

6. **Wire up routing**
   - `useState<Screen>` in App.tsx drives which component renders in ContentArea
   - Each ops layer gets a placeholder page component

7. **Configure `wails.json`**
   ```json
   {
     "name": "hawkward-gui",
     "output": "../build/bin/hawkward-gui.exe",
     "frontend:install": "npm install",
     "frontend:build": "npm run build",
     "frontend:dev:watcher": "npm run dev",
     "frontend:dev:serverUrl": "http://localhost:5173",
     "author": {
       "name": "Hawkward Team",
       "email": "dev@hawkward.io"
     }
   }
   ```

8. **Update `go.mod`**
   - Add `github.com/wailsapp/wails/v2` dependency
   - Remove Bubble Tea (`charm.land/bubbletea/v2`), Lipgloss (`charm.land/lipgloss/v2`), and Bubbles (`github.com/charmbracelet/bubbles`)
   - Keep: `gopsutil/v4`, `miekg/dns`, `golang.org/x/net`

9. **Archive TUI code**
   - Move `internal/ui/` → `internal/ui_legacy/` (keep for reference, not compiled)
   - Remove TUI-specific styles from `internal/common/` (Lipgloss styles), but keep shared types, formatters, pipeline, timeseries, forecast, alerts

10. **First build test**
    - `wails build` — should compile, open a window with the sidebar layout and placeholder pages

**Deliverables**: Running Wails app with sidebar navigation, empty pages for each ops layer, Go backend bindings returning mock data

---

### Phase 2: Dashboard + SysOps Pages (Sprint 2)

**Objective**: Real dashboard with live data, health cards, gauges, sparklines, alerts. SysOps page with full system monitoring.

#### Go Backend Work

1. **`Dashboard.go`** — Implement `GetSystemStats()` calling `sysops.Collector.Collect()`, returning `SystemStats` with history. Implement `GetHealthSummary()` returning aggregated health (CPU/Memory/Disk status, anomaly count, alert count).
2. **`SysOps.go`** — Implement `GetCPU()`, `GetMemory()`, `GetDisk()`, `GetProcesses()`, `GetSystemInfo()` by wrapping existing `internal/sysops/` functions.
3. **`Alerts.go`** — Implement `GetFiredAlerts()` calling `common/alerts.go`
4. **`Events.go`** — Implement `StartDashboardTicker(interval)` that emits `dashboard:tick` events with fresh `SystemStats` every N seconds.

#### Frontend Work

1. **Dashboard page** (`DashboardPage.tsx`)
   - Health cards row (CPU, Memory, Disk) — each a shadcn Card with a Gauge chart and status indicator
   - Sparkline row for recent CPU/Memory history (Recharts AreaChart with transparent fill)
   - Alerts bar — scrolling list of active alerts (badge with severity color)
   - Quick-action buttons (Run Pipeline, View Logs, Open Settings)
   - Forecast snippet — next-hour predicted values for CPU/Memory

2. **SysOps page** (`SysOpsPage.tsx`)
   - CPU section: radial gauge (Recharts custom), per-core breakdown table, usage history chart
   - Memory section: progress bar gauge, used/total/available breakdown, swap info
   - Disk section: per-disk cards with usage bars, read/write throughput sparklines
   - Processes table: sortable columns (PID, Name, CPU%, Mem%, Status), search filter
   - System info card: OS, hostname, uptime, kernel version, platform

3. **Real-time updates**
   - `useEvents('dashboard:tick')` hook receives live SystemStats
   - Components update without polling — CPU gauge needle moves, sparklines extend, alerts bar refreshes

4. **Chart components** (`/components/charts/`)
   - `GaugeChart.tsx` — Radial gauge for CPU/Memory/Disk percentage (Recharts Pie with filled arc)
   - `SparklineChart.tsx` — Mini AreaChart for history preview
   - `LineChart.tsx` — Full-size Recharts LineChart with tooltip, legend, responsive
   - `AreaChart.tsx` — Stacked area for memory/disk breakdown
   - `BarChart.tsx` — For process comparison, per-core CPU

**Deliverables**: Live dashboard with real-time gauges and sparklines. Full SysOps page with process management.

---

### Phase 3: NetOps + SecOps Pages (Sprint 3)

**Objective**: Network operations tools + network topology visualization. Security operations page.

#### Go Backend Work

1. **`NetOps.go`** — Wire up all existing `internal/netops/` functions:
   - `Ping(host string, count int)` returns ping results with RTT stats
   - `DNSLookup(host string, recordType string)` returns DNS records
   - `PortScan(host string, ports []int)` returns open/closed/filtered
   - `Traceroute(host string)` returns hop-by-hop results
   - `GetConnections()` returns active network connections
   - `GetInterfaces()` returns network interfaces with IP, MAC, status
   - `StartBandwidthMonitor()` emits `net:bandwidth` events with real-time throughput

2. **`SecOps.go`** — Wire up existing `internal/secops/`:
   - `GetFirewallRules()` returns parsed Windows firewall rules
   - `GetUsers()` returns local user accounts with metadata
   - `GetListeningPorts()` returns listening TCP/UDP ports
   - `GetDefenderStatus()` returns Windows Defender state, last scan, threats
   - `GetSecurityEvents()` returns recent security event log entries
   - `StartSecurityMonitor()` emits `sec:event` for real-time security alerts

#### Frontend Work

1. **NetOps page** (`NetOpsPage.tsx`)
   - Ping tool: host input, count slider, start button, live results table (RTT, TTL, status), stats summary
   - DNS lookup: host input, record type selector (A, AAAA, MX, NS, TXT, CNAME), results table
   - Port scanner: host input, port range inputs, scan button, results color-coded (open/closed/filtered)
   - Traceroute: host input, start button, hop table with RTT bars
   - Connection table: sortable/filterable table of active connections
   - Interface list: cards showing interface name, IP, MAC, status, speed

2. **Network Topology** (`NetworkTopologyPage.tsx` or integrated tab)
   - Use `vis-network` React wrapper
   - Auto-discover nodes: gateways, DNS servers, known hosts from connection table + arp cache
   - Interactive: drag nodes, zoom, pan, click for details
   - Color-coded: local vs remote, active vs inactive, Windows vs Linux

3. **Bandwidth monitor** (widget on dashboard or NetOps page)
   - Real-time line chart (download/upload Mbps)
   - Time window selector (30s, 1m, 5m, 15m)
   - Top connections by bandwidth usage

4. **SecOps page** (`SecOpsPage.tsx`)
   - Firewall rules table: filterable by direction, action, protocol
   - User accounts table: name, group, disabled/locked status, last logon
   - Listening ports: table with port, protocol, process, PID, state
   - Defender status card: real-time protection, last quick scan, threat count
   - Security events feed: scrolling log with severity badges, filter by event type

**Deliverables**: Full NetOps toolset with interactive topology map. SecOps page with firewall, user, and event management.

---

### Phase 4: DevOps + AIOps + Network Design (Sprint 4)

**Objective**: Developer tools, AI chat interface, network topology designer.

#### Go Backend Work

1. **`DevOps.go`** — Wire up existing `internal/devops/`:
   - `ExecCommand(cmd string, args []string)` — sandboxed command execution, streams output via events
   - `TailLog(path string, lines int)` — tail a log file, emits `devops:logline` events
   - `BrowseFiles(path string)` — directory listing, navigate up/down
   - `GetProcesses()` — process list with CPU/Memory
   - `GetServices()` — service list with status (running/stopped)

2. **`AIOps.go`** — Wire up existing `internal/aiops/`:
   - `Chat(message string)` — send message to Ollama, stream response via events
   - `GenerateReport(layer string, format string)` — generate ops report using AI
   - `QueryState(query string)` — semantic query of system state

3. **`Pipeline.go`** — Full implementation:
   - `CreatePipeline(name, config)` — define a new data pipeline
   - `RunPipeline(id)` — execute a pipeline, stream progress
   - `GetPipelineStatus(id)` — check execution state
   - `ListPipelines()` — all defined pipelines
   - `DeletePipeline(id)` — remove a pipeline

#### Frontend Work

1. **DevOps page** (`DevOpsPage.tsx`)
   - Command terminal: input field, output display with ANSI stripping, command history, Ctrl+C support
   - Log tailer: file path input/select, streaming log output, search/filter, pause/resume, clear
   - File browser: tree view of filesystem, click to navigate, file info panel, open-in-explorer
   - Process manager: same table as SysOps but with kill/restart actions
   - Services list: service name, status, start type, start/stop/restart actions

2. **AIOps page** (`AIOpsPage.tsx`)
   - Chat interface: message bubbles, streaming response (typewriter effect), system prompt selector
   - Report generator: layer selector (sysops/netops/secops/devops), format selector, generate button, preview pane
   - Anomaly display: list of detected anomalies with severity, timestamp, description, AI explanation

3. **Network Design tool** (`NetworkDesignerPage.tsx`)
   - Drag & drop canvas using `vis-network` in edit mode
   - Node palette: router, switch, firewall, server, client, cloud, custom
   - Edge configuration: bandwidth, latency, protocol, status
   - Export topology as JSON/PNG
   - Auto-layout algorithms (hierarchical, force-directed, circular)

4. **Pipeline management** (tab in DevOps or standalone page)
   - Pipeline list with status badges, last run time, duration
   - Create/edit pipeline dialog (name, config JSON, schedule)
   - Pipeline run history with timestamps, duration, success/failure
   - Start/stop/delete actions

**Deliverables**: Terminal, file browser, log tailer, AI chat, network topology designer.

---

### Phase 5: Polish & Dialogs (Sprint 5)

**Objective**: Export, settings, about, session logging, error handling, edge cases.

#### Go Backend Work

1. **`Logs.go`** — Full implementation:
   - `GetLogs(filter LogFilter)` — query logs with level, time range, source filters
   - `ExportLogs(format string, filter LogFilter)` — export to CSV, JSON, Markdown
   - `ClearLogs()` — rotate/archive log file

2. **`Settings.go`** — Full implementation:
   - `GetSettings()` — current app settings
   - `UpdateSettings(settings)` — save settings to config file
   - `GetTheme()` — active theme name
   - `SetTheme(name)` — switch theme, emit `app:theme-changed`
   - `GetRefreshRate()` / `SetRefreshRate(ms)` — dashboard tick interval

3. **App lifecycle hardening**:
   - Graceful shutdown: stop all goroutines, save state, flush logs
   - Error recovery: panic handler, restart tickers on crash
   - Session logging: timestamps + activity log to `~/.config/hawkward/session.log`

#### Frontend Work

1. **Export Dialog** (`ExportDialog.tsx`)
   - Format selector: CSV, JSON, Markdown, PDF (via `html2pdf.js` or similar)
   - Scope selector: current view, all data, time range
   - Preview pane before export
   - Wails runtime file save dialog for destination
   - Progress indicator for large exports

2. **Settings Dialog** (`SettingsDialog.tsx`)
   - Theme selector: Dark, Light, System (with preview)
   - Refresh rate slider: 1s, 2s, 5s, 10s, 30s (for dashboard ticks)
   - AI config: Ollama endpoint URL, model selector, system prompt
   - Pipeline defaults: max concurrent, timeout
   - Logging level: Debug, Info, Warn, Error
   - Reset to defaults button

3. **About Dialog** (`AboutDialog.tsx`)
   - App name, version, build date
   - Go version, Wails version
   - License information
   - Link to repository

4. **Session logging**
   - Activity log panel accessible from top bar
   - Shows: screen navigation, operations performed, errors encountered
   - Search and filter
   - Export session log

5. **Error handling**
   - Global error boundary React component
   - Toast notifications for async errors (shadcn toast)
   - Network/backend connection lost indicator
   - Retry mechanism for failed backend calls
   - Loading states on all data-fetching components (skeleton loaders)

6. **Accessibility & UX polish**
   - Keyboard shortcuts: Ctrl+K for command palette, Ctrl+, for settings, Escape to close dialogs
   - Focus trapping in modals
   - Tooltips on icon-only buttons
   - Empty states for all data views
   - Responsive layout adjustments (sidebar collapse on narrow windows)

**Deliverables**: Production-ready polish, export to multiple formats, settings persistence, comprehensive error handling.

---

## Design System

### Framework Stack

| Layer | Technology |
|-------|-----------|
| **Desktop framework** | Wails v2 |
| **UI library** | React 18+ with TypeScript |
| **Build tool** | Vite 6 |
| **CSS framework** | Tailwind CSS v4 |
| **Component library** | shadcn/ui (Radix primitives) |
| **Charts** | Recharts |
| **Network topology** | vis-network (via `vis-network-react` or wrapper) |
| **Icons** | Lucide React |
| **Type checking** | TypeScript 5.x (strict mode) |

### Color System (Squib-inspired)

```css
/* Dark mode (default) */
:root {
  --background: #0f172a;    /* slate-900 */
  --card: #1e293b;          /* slate-800 */
  --popover: #1e293b;       /* slate-800 */
  --primary: #38bdf8;       /* sky-400 */
  --primary-foreground: #0f172a;
  --secondary: #334155;     /* slate-700 */
  --secondary-foreground: #f8fafc;
  --accent: #334155;        /* slate-700 */
  --accent-foreground: #f8fafc;
  --success: #4ade80;       /* green-400 */
  --warning: #fbbf24;       /* amber-400 */
  --danger: #f87171;        /* red-400 */
  --info: #60a5fa;          /* blue-400 */
  --muted: #94a3b8;         /* slate-400 */
  --muted-foreground: #64748b;
  --text: #f8fafc;          /* slate-50 */
  --text-secondary: #cbd5e1;/* slate-300 */
  --border: #334155;        /* slate-700 */
  --ring: #38bdf8;          /* sky-400 */
  --radius: 0.5rem;
}

/* Light mode */
.light {
  --background: #f8fafc;    /* slate-50 */
  --card: #ffffff;
  --popover: #ffffff;
  --primary: #0284c7;       /* sky-600 */
  --primary-foreground: #ffffff;
  --secondary: #e2e8f0;     /* slate-200 */
  --secondary-foreground: #0f172a;
  --accent: #e2e8f0;        /* slate-200 */
  --accent-foreground: #0f172a;
  --success: #16a34a;       /* green-600 */
  --warning: #d97706;       /* amber-600 */
  --danger: #dc2626;        /* red-600 */
  --info: #2563eb;          /* blue-600 */
  --muted: #64748b;         /* slate-500 */
  --muted-foreground: #94a3b8;
  --text: #0f172a;          /* slate-900 */
  --text-secondary: #334155;/* slate-700 */
  --border: #e2e8f0;        /* slate-200 */
  --ring: #0284c7;          /* sky-600 */
}
```

### Typography

| Element | Size | Weight | Family |
|---------|------|--------|--------|
| Page title | text-3xl (30px) | font-bold | Inter / system sans-serif |
| Section heading | text-xl (20px) | font-semibold | Inter |
| Card title | text-base (16px) | font-medium | Inter |
| Body | text-sm (14px) | font-normal | Inter |
| Small / metric | text-xs (12px) | font-medium | Inter |
| Monospace (logs/terminal) | text-sm | font-mono | JetBrains Mono / Fira Code |

### Layout

```
┌─────────────────────────────────────────────────┐
│  TopBar (screen title, time, theme toggle, ...) │
├──────────┬──────────────────────────────────────┤
│          │                                       │
│ Sidebar  │  ContentArea                          │
│ (collap- │  (active screen component)            │
│  sible)  │                                       │
│          │                                       │
│ Icons +  │                                       │
│ Labels   │                                       │
│          │                                       │
├──────────┴──────────────────────────────────────┤
│  StatusBar (alerts count, last refresh, ...)     │
└─────────────────────────────────────────────────┘
```

### Component Architecture

```
App.tsx
├── ThemeProvider (context)
├── EventsProvider (context — Wails event bus)
├── TopBar
│   ├── ScreenTitle
│   ├── SearchInput (quick nav)
│   ├── ThemeToggle
│   └── SettingsButton
├── Sidebar
│   ├── NavItem (Dashboard)
│   ├── NavItem (SysOps)
│   ├── NavItem (NetOps)
│   ├── NavItem (SecOps)
│   ├── NavItem (DevOps)
│   ├── NavItem (AIOps)
│   └── CollapseButton
├── ContentArea
│   └── [active screen component]
│       ├── DashboardPage
│       │   ├── HealthCards row
│       │   ├── SparklineSection
│       │   ├── ForecastSection
│       │   └── AlertsBar
│       ├── SysOpsPage
│       │   ├── CPUSection (GaugeChart + LineChart)
│       │   ├── MemorySection (GaugeChart + breakdown)
│       │   ├── DiskSection (per-disk cards)
│       │   ├── ProcessTable
│       │   └── SystemInfoCard
│       ├── NetOpsPage (tabs: Ping, DNS, Port Scan, Traceroute, Connections, Topology)
│       ├── SecOpsPage (tabs: Firewall, Users, Ports, Defender, Events)
│       ├── DevOpsPage (tabs: Terminal, Logs, Files, Processes, Services)
│       ├── AIOpsPage (tabs: Chat, Reports, Anomalies)
│       └── NetworkDesignerPage (vis-network canvas)
└── StatusBar
    ├── AlertCount
    ├── LastRefresh
    └── ConnectionStatus
```

### shadcn/ui Components to Install

| Component | Usage |
|-----------|-------|
| `Button` | Actions, toolbar buttons |
| `Card` | Dashboard cards, section containers |
| `Dialog` | Settings, Export, About dialogs |
| `Input` | Text inputs (host, command, search) |
| `Select` | Dropdown selectors (record type, format) |
| `Table` | Process table, connections, firewall rules |
| `Tabs` | Page section navigation (NetOps tabs) |
| `Badge` | Status indicators (severity, service state) |
| `Separator` | Visual section dividers |
| `ScrollArea` | Scrollable content (log viewer, terminal) |
| `Sheet` | Side panel (quick settings, details) |
| `DropdownMenu` | Action menus (export, per-row actions) |
| `Switch` | Toggle controls (theme, auto-refresh) |
| `Toast` | Notifications for async operations |
| `Skeleton` | Loading placeholders |
| `Progress` | Pipeline progress, scan progress |
| `Tooltip` | Icon descriptions, truncated text |

---

## Real-Time Data Architecture

### Event Flow

```
Go Backend                          Frontend
───────────                         ────────
Ticker (every N sec) ──emit──→  useEvents('dashboard:tick')
  └─ sysops.Collector.Collect()      └─ setStats(data)
  └─ alerts.Check()                  └─ setAlerts(data)
  └─ forecast.Predict()             └─ setForecast(data)

Bandwidth Monitor ──emit──→  useEvents('net:bandwidth')
  └─ netops.GetThroughput()          └─ appendToChart(data)

Log Tailer ──emit──→  useEvents('devops:logline')
  └─ tail -f                        └─ appendToView(line)

AI Response ──emit──→  useEvents('aiops:stream')
  └─ ollama.Stream()                └─ typewriterEffect(text)

Security Events ──emit──→  useEvents('sec:event')
  └─ event log watcher              └─ toast notification
```

### Event Names Convention

| Event Name | Payload | Emitter |
|-----------|---------|---------|
| `dashboard:tick` | `SystemStats` | Dashboard ticker |
| `dashboard:health` | `HealthSummary` | Dashboard health check |
| `net:bandwidth` | `BandwidthSample` | Bandwidth monitor |
| `net:scan-progress` | `ScanProgress` | Port scanner |
| `devops:logline` | `LogLine` | Log tailer |
| `devops:cmd-output` | `CommandOutput` | Shell executor |
| `aiops:stream` | `AIMessageChunk` | Ollama streaming |
| `sec:event` | `SecurityEvent` | Event log watcher |
| `app:theme-changed` | `string` | Settings changes |
| `app:error` | `ErrorInfo` | Error recovery |

---

## Go Backend Binding Patterns

### Pattern: Method returning data synchronously

```go
// SysOps.go
package app

import "hawkward/internal/sysops"

type SysOps struct {
    collector *sysops.Collector
}

func NewSysOps() *SysOps {
    return &SysOps{collector: sysops.NewCollector()}
}

// GetCPU returns current CPU statistics.
// This method is bound to the frontend via Wails Bind().
func (s *SysOps) GetCPU() (*CPUData, error) {
    stats, err := s.collector.CollectCPU()
    if err != nil {
        return nil, fmt.Errorf("cpu collection failed: %w", err)
    }
    return &CPUData{
        Percent:    stats.Percent,
        PerCore:    stats.PerCore,
        Temp:       stats.Temperature,
        CoreCount:  stats.CoreCount,
    }, nil
}
```

### Pattern: Long-running operation with event streaming

```go
// NetOps.go
func (n *NetOps) StartPing(host string, count int) error {
    go func() {
        for i := 0; i < count; i++ {
            result, err := n.pinger.Ping(host)
            if err != nil {
                runtime.EventsEmit(n.ctx, "net:ping-result", map[string]any{
                    "error": err.Error(),
                    "seq":   i,
                    "done":  i == count-1,
                })
                return
            }
            runtime.EventsEmit(n.ctx, "net:ping-result", map[string]any{
                "seq":     i,
                "rtt_ms":  result.RTT.Milliseconds(),
                "ttl":     result.TTL,
                "success": true,
                "done":    i == count-1,
            })
            time.Sleep(time.Second)
        }
    }()
    return nil
}
```

### Pattern: Frontend hook for typed events

```typescript
// hooks/useEvents.ts
import { useState, useEffect, useCallback } from 'react';

export function useBackendEvent<T>(eventName: string) {
  const [data, setData] = useState<T | null>(null);

  useEffect(() => {
    const unsubscribe = window.runtime.EventsOn(eventName, (payload: T) => {
      setData(payload);
    });
    return () => {
      window.runtime.EventsOff(eventName);
    };
  }, [eventName]);

  return data;
}
```

---

## Migration Notes

### What Changes

| File/Module | Action |
|-------------|--------|
| `cmd/hawkward/main.go` | **Replace** with `cmd/hawkward-gui/main.go` (Wails entry) |
| `internal/app/` | **New** — all Go bindings |
| `frontend/` | **New** — React + TypeScript + Vite |
| `wails.json` | **New** — Wails project config |
| `go.mod` | **Update** — add Wails, remove Bubble Tea/Lipgloss/Bubbles |
| `internal/common/styles.go` | **Archive** — Lipgloss styles (TUI-specific) |
| `internal/common/charts/` | **Archive** — Unicode terminal charts (replaced by Recharts) |
| `internal/ui/` | **Archive** — full TUI directory (all model/update/view files) |
| `CLAUDE.md` | **Update** — reflect new GUI architecture, build commands |
| All ops layers | **Reuse** as-is — no changes needed to `internal/{sysops,netops,secops,devops,aiops}/` |

### What Stays

- All `internal/common/` files **except** `styles.go` (Lipgloss-specific) — types, formatters, pipeline, timeseries, forecast, alerts, logger, platform, sandbox, metrics
- All `internal/sysops/`, `internal/netops/`, `internal/secops/`, `internal/devops/`, `internal/aiops/` — pure Go logic with no TUI dependency
- `scripts/` — build scripts updated for Wails
- `docs/` — documentation, updated for GUI

### What Ships

- **TUI binary** (`hawkward.exe`) — still buildable from `cmd/hawkward/` for terminal users
- **GUI binary** (`hawkward-gui.exe`) — new Wails-based desktop app from `cmd/hawkward-gui/`

### Coexistence Strategy

Both `cmd/hawkward/` (TUI) and `cmd/hawkward-gui/` (GUI) coexist during migration:
- Shared internal packages remain untouched — no behavioral changes
- TUI can be built with `go build -o hawkward.exe ./cmd/hawkward`
- GUI can be built with `wails build`
- After full GUI validation, TUI entry can be deprecated but kept for reference

---

## Build & Dev Commands

```bash
# Development (hot reload frontend + backend)
wails dev

# Production build
wails build

# Build for specific platform
wails build -platform windows/amd64
wails build -platform darwin/amd64
wails build -platform linux/amd64

# Frontend dev only (browser preview)
cd frontend && npm run dev

# Run Go tests (all packages including new app/)
go test ./...

# TUI still buildable
go build -o hawkward.exe ./cmd/hawkward
```

---

## Key Decisions

1. **Wails v2 over Electron** — Go backend reuse is the primary driver. Wails is smaller, faster, and uses significantly less memory than Electron.
2. **React over Vue/Svelte** — shadcn/ui is React-only. React has the largest ecosystem for charting (Recharts), visualization (vis-network), and component libraries.
3. **shadcn/ui over MUI/AntD** — Tailwind-based, copy-paste components (no dependency bloat), full control over styling, dark/light mode built-in.
4. **Recharts over D3/Chart.js** — React-native (declarative), good TypeScript support, lightweight, easy to customize with Tailwind. D3 is available for custom visualizations if needed.
5. **vis-network over cytoscape.js** — Better React integration via wrappers, more mature for network topology, built-in physics engine for auto-layout.
6. **TUI code archived, not deleted** — Reference for behavior verification during migration. Can be removed after full GUI validation.
7. **Event-driven real-time** — Wails Events system is the single channel for live data. No WebSocket, no polling from frontend (except initial load).
8. **Backend methods are synchronous by default** — Long-running operations use goroutines + Events for streaming. The frontend calls a start method and listens for events.

---

## Testing Strategy

| Layer | Tool | Approach |
|-------|------|----------|
| Go backend (existing) | `go test` | Existing tests continue to pass |
| Go app bindings | `go test` | Unit tests for binding methods with mocked ops layers |
| React components | Vitest + React Testing Library | Component render tests, hook tests |
| End-to-end | Wails dev + manual | Visual verification, cross-platform testing |
| Real-time events | Integration test | Mock Wails runtime, verify event emissions |

---

*Last updated: 2026-07-07*
*Owner: GUI Overhaul Architect*
