# Hawkward GUI Overhaul v2 — Squib-Inspired Operations Platform

> **Active Sprint Plan** | Created: 2026-07-07 | Replaces: `gui-overhaul-v1.md`
>
> **Vision**: Transform the Hawkward TUI operations platform into a premium native desktop GUI
> inspired by the squib design language — featuring real-time charts, interactive network topology,
> forecasting, alerting, log viewer, and an intuitive multi-panel layout for technical operations work.

---

## Table of Contents

1. [Research Synthesis](#1-research-synthesis)
2. [Design System (Squib-Inspired)](#2-design-system-squib-inspired)
3. [Architecture Overview](#3-architecture-overview)
4. [Sprint 1: Design System Foundation + Backend Wiring](#4-sprint-1-design-system-foundation--backend-wiring)
5. [Sprint 2: Dashboard + SysOps](#5-sprint-2-dashboard--sysops)
6. [Sprint 3: NetOps (The Star)](#6-sprint-3-netops-the-star)
7. [Sprint 4: SecOps + DevOps + AIOps](#7-sprint-4-secops--devops--aiops)
8. [Sprint 5: Network Designer + Logs + Settings](#8-sprint-5-network-designer--logs--settings)
9. [Sprint 6: Polish, Performance, Packaging](#9-sprint-6-polish-performance-packaging)
10. [Frontend Component Tree](#10-frontend-component-tree)
11. [Real-Time Data Architecture](#11-real-time-data-architecture)
12. [Feature Inventory by Page](#12-feature-inventory-by-page)
13. [Testing Strategy](#13-testing-strategy)
14. [Build & Dev Commands](#14-build--dev-commands)

---

## 1. Research Synthesis

### Sources
| Repo | Stars | Key Takeaway |
|------|-------|-------------|
| **squib** (`teknetai/squib`) | — | **UI/UX design language**: dark theme with `#080a0f` bg, `#7c6cff` purple accent, Inter + JetBrains Mono typography, 252px sidebar, KPI cards with sparklines, status dot system with glow, glassmorphism topbar, slide-in drawers, toast notifications |
| **netscanner** (`Chleba/netscanner`) | — | Interactive network scanning UI: real-time port scan visualization, service detection, topology-aware scanning, bandwidth monitoring |
| **hackingtool** (`Z4nzu/hackingtool`) | 78k+ | 185-tool aggregator pattern: category-based navigation, tool status tracking, OS-aware filtering, search/tag/recommend system |

### Design Decisions from Research

| Decision | Rationale |
|----------|-----------|
| **Adopt squib's exact color tokens** | Proven dark dashboard design with 3-tier text hierarchy, consistent status colors, and purple accent identity |
| **Replace current Tailwind color scheme** | Current `#0f172a` bg + `#38bdf8` blue accent → squib's `#080a0f` bg + `#7c6cff` purple accent |
| **Add Inter + JetBrains Mono fonts** | Squib uses Inter for UI, JetBrains Mono for data — critical for tabular-nums alignment in metrics |
| **Implement squib's KPI card layout** | 4-column grid with icon, large value, delta, SVG sparkline + hover lift effect |
| **Adopt squib's drawer/modal system** | Right slide-in detail drawer (540px) with scrim + blur, centered modals (560px) with backdrop blur |
| **Category menu from hackingtool** | Ops layers as squib-style sidebar nav items with active state + badge counts |
| **Real-time scanning from netscanner** | Port scanning with progress bars, live results table, service color coding |
| **Replace inline SVGs with lucide-react** | Squib uses custom SVGs but lucide-react provides consistent icon system |

---

## 2. Design System (Squib-Inspired)

### 2.1 Color Palette

```css
/* Dark Theme (default) */
--bg:            #080a0f;    /* Deepest page background */
--bg-gradient:   #0f1120;    /* Subtle vignette */
--panel:         #0f1118;    /* Cards, sidebar, tables */
--panel-2:       #141722;    /* Secondary surfaces, inputs */
--panel-3:       #1a1e2b;    /* Hover states, chips */
--elevated:      #171a25;    /* Toasts, modals */
--border:        rgba(255,255,255,.07);
--border-2:      rgba(255,255,255,.11);
--text:          #e7e9f2;    /* Primary body text */
--text-dim:      #9298ab;    /* Secondary text */
--text-faint:    #5e6578;    /* Labels, timestamps */
--accent:        #7c6cff;    /* Primary purple */
--accent-2:      #a78bff;    /* Lighter accent */
--accent-soft:   rgba(124,108,255,.14);
--up:            #2dd4a7;    /* Healthy status */
--up-soft:       rgba(45,212,167,.13);
--degraded:      #f5b942;    /* Warning status */
--degraded-soft: rgba(245,185,66,.13);
--down:          #fb5d6b;    /* Critical status */
--down-soft:     rgba(251,93,107,.13);

/* Light Theme */
--bg:            #f4f5f9;
--bg-gradient:   #e7e9ff;
--panel:         #ffffff;
--panel-2:       #f7f8fc;
--panel-3:       #eef0f6;
--accent:        #6353f0;    /* Darker purple for light */
--up:            #11a981;
--degraded:      #d9941a;
--down:          #e23b4c;
```

### 2.2 Typography

| Token | Value |
|-------|-------|
| Font UI | `'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif` |
| Font Mono | `'JetBrains Mono', ui-monospace, SFMono-Regular, Menlo, monospace` |
| Base size | 14px, line-height 1.45, letter-spacing -0.01em |
| Tabular nums | `.tabular-nums { font-variant-numeric: tabular-nums }` on all metric values |
| Page title | 16px, weight 650, letter-spacing -0.02em |
| Page subtitle | 12px, color var(--text-faint) |
| KPI values | 27px, weight 700, letter-spacing -0.035em |
| Panel heading | 14px, weight 620, letter-spacing -0.02em |
| Table header | 11px, weight 600, uppercase, letter-spacing 0.05em |
| Table body | 13px |
| Status badge | 12.5px, weight 550 |
| Section label | 10.5px, weight 600, uppercase, letter-spacing 0.08em |
| Brand | 16.5px, weight 700, letter-spacing -0.02em |

### 2.3 Spacing & Radius Scale

| Token | Value | Used For |
|-------|-------|----------|
| `--r-sm` | 8px | Buttons, small chips |
| `--r` | 12px | KPI cards, inputs |
| `--r-lg` | 16px | Panels, hero section |
| `--r-xl` | 22px | Modals, drawers |
| Space | 24px | Page padding, section gaps |
| Grid gap | 14px | Between KPI cards |
| Grid gap | 16px | Between panels |
| Sidebar | 252px | Navigation width |
| Topbar | 60px | Header height |
| Padding card | 15px 16px | Inside KPI cards |
| Padding panel | 22px 24px | Inside panels |

### 2.4 Shadows

```css
--shadow:     '0 1px 2px rgba(0,0,0,.4), 0 8px 24px -8px rgba(0,0,0,.5)'
--shadow-lg:  '0 24px 64px -16px rgba(0,0,0,.7)'
--glow:      '0 6px 18px -6px rgba(124,108,255,.7)'
```

### 2.5 Status Indicators

- **Status dots**: 8x8px, `border-radius: 50%` with 3px `box-shadow` matching status color at `.13` opacity
- **Live ping**: CSS animation `ping 1.8s cubic-bezier(0,0,.2,1) infinite` — scales to 2.6x and fades
- **Pills/tags**: 11px, weight 600, padding 3px 10px, border-radius 20px
- **Status badges**: inline-flex, gap 7px, weight 550, 12.5px

### 2.6 Animations

| Element | Duration | Easing |
|---------|----------|--------|
| Page fade-in | 0.4s | ease |
| Drawer slide | 0.32s | cubic-bezier(.4,0,.1,1) |
| Scrim fade | 0.25s | — |
| Toast slide | 0.3s | cubic-bezier(.2,.8,.2,1) |
| KPI card hover lift | 0.18s | — |
| Gauge arc | 0.6s | — |
| All transitions | 0.12-0.15s | — |

### 2.7 Scrollbars

```css
::-webkit-scrollbar{width:6px;height:6px}
::-webkit-scrollbar-thumb{background:var(--border-2);border-radius:20px;border:2px solid transparent;background-clip:content-box}
::-webkit-scrollbar-thumb:hover{background:var(--text-faint);background-clip:content-box}
```

---

## 3. Architecture Overview

### 3.1 Current State (Before)

```
TUI: cmd/hawkward/            Bubble Tea terminal app
GUI: cmd/hawkward-gui/        React + Tailwind v4 (skeleton, mock data)
Backend: internal/app/*.go    Wails bindings (working, but not connected)
Frontend: hooks/*.ts          Stubs that log to console
Mock: lib/mockData.ts         Complex mock generators (used everywhere)
```

### 3.2 Target State (After)

```
main.go                       Wails entry point (unchanged)
internal/app/                 Go bindings → all exported methods callable from JS
internal/common/              Data pipeline, forecast engine, alerts, timeseries
internal/{sysops,netops,...}/ Domain collectors (unchanged)
cmd/hawkward-gui/frontend/    === FULLY REDESIGNED ===
  src/
    App.tsx                   Root: squib layout shell (sidebar + topbar + content)
    styles/globals.css        Squib color tokens + theme system
    hooks/
      useBackend.ts           REAL Wails runtime calls (window.go.main.App.*)
      useEvents.ts            REAL Wails EventsOn subscriptions
      useTheme.ts             Theme toggle with CSS variable switching
    lib/
      mockData.ts             Reduced: only for dev mode fallback
      utils.ts                cn(), formatters, validators
    types/index.ts            Go type mirrors (already complete)
    components/
      layout/                 Sidebar (252px), TopBar (60px glass), MainContent
      charts/                 AreaChart, Gauge, MiniSparkline, BarChart, HeatMap
      dashboard/              HealthCard, AlertBadge, KPIRow, StatusBar
      dialogs/                ExportDialog, SettingsDialog, ProcessManager, AlertDialog
      network/                DeviceNode, ConnectionLine, PropertiesPanel, MiniMap
      logs/                   LogViewer, LogFilterBar, LogDetail
      ui/                     shadcn-inspired: Button, Input, Select, Badge, Card, Tabs, Toast, Drawer, Modal
    pages/
      Dashboard.tsx           Squib-style hero + KPI cards + charts + alerts
      SysOps.tsx              Sub-tabbed: Overview, Processes, System Info, Disks
      NetOps.tsx              Sub-tabbed: Ping, DNS, PortScan, Traceroute, Connections, Interfaces, Bandwidth
      SecOps.tsx              Sub-tabbed: Firewall, Users, Ports, Defender, Events
      DevOps.tsx              Sub-tabbed: Terminal, Services, Files, Containers
      AIOps.tsx               Sub-tabbed: Chat, Reports, Anomalies
      NetworkDesign.tsx       Canvas + palette + properties + save/load
      Logs.tsx                Virtual-scrolled log viewer + filter + search + export
      Settings.tsx            Theme, alerts, pipeline, about
```

### 3.3 Data Flow

```
Go Backend (every 3s tick loop)
  │
  ├─ runtime.EventsEmit("metrics", MetricsEvent)  ← push to ALL subscribers
  ├─ runtime.EventsEmit("alert", AlertEvent)
  └─ runtime.EventsEmit("log", LogEvent)
       │
       ▼
Frontend hooks
  useEvents("metrics", handler)   ← updates dashboard state in real-time
  useEvents("alert", handler)     ← shows toast, updates badge
  useEvents("log", handler)       ← appends to log viewer
       │
       ▼
On-demand calls (user actions)
  window.go.main.App.SysOps.GetCPUInfo()   ← button click
  window.go.main.App.NetOps.Ping(host, 4)  ← start ping
       │
       ▼
Go method → returns data → React state update → re-render
```

---

## 4. Sprint 1: Design System Foundation + Backend Wiring

**Duration**: ~2 sessions | **Files touched**: ~20

### 4.1 CSS Design System Overhaul

**File**: `cmd/hawkward-gui/frontend/src/styles/globals.css`

Replace the current Tailwind v4 theme block with squib's exact color tokens:

```css
@import "tailwindcss";

@theme {
  /* Squib-inspired dark theme colors */
  --color-bg: #080a0f;
  --color-bg-gradient: #0f1120;
  --color-panel: #0f1118;
  --color-panel-2: #141722;
  --color-panel-3: #1a1e2b;
  --color-elevated: #171a25;
  --color-border: rgba(255,255,255,.07);
  --color-border-2: rgba(255,255,255,.11);
  --color-text: #e7e9f2;
  --color-text-dim: #9298ab;
  --color-text-faint: #5e6578;
  --color-accent: #7c6cff;
  --color-accent-2: #a78bff;
  --color-accent-soft: rgba(124,108,255,.14);
  --color-success: #2dd4a7;
  --color-warning: #f5b942;
  --color-danger: #fb5d6b;
  --color-sidebar: #0b1120;
  --color-sidebar-hover: rgba(124,108,255,.08);
}
```

Add theme toggle via CSS custom properties on `[data-theme="light"]`.

### 4.2 Font Loading

Add to `index.html`:
```html
<link rel="preconnect" href="https://fonts.googleapis.com">
<link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;450;500;600;650;700&family=JetBrains+Mono:wght@400;500;600&display=swap" rel="stylesheet">
```

### 4.3 Layout Component Redesign

| Component | Current | Target |
|-----------|---------|--------|
| **Sidebar** | 224px, blue active border, inline SVG icons | 252px fixed, purple accent bg-soft active, lucide-react icons, badge counts, footer with status |
| **TopBar** | Plain header with breadcrumb + theme toggle | 60px glass `backdrop-filter: blur(14px)`, breadcrumb, live status dot (pulsing), notification bell, theme toggle |
| **MainContent** | Simple flex wrapper | `<main>` with radial gradient vignette, `padding: 24px`, custom scrollbar |

### 4.4 Real Backend Wiring (CRITICAL)

**File**: `cmd/hawkward-gui/frontend/src/hooks/useBackend.ts`

Replace mock stub with actual Wails runtime calls:

```typescript
// Before: console.log mock
// After: window.go.main.App.SysOps.GetCPUInfo()

export function useBackend() {
  const call = async (method: string, ...args: any[]) => {
    // Try Wails runtime first
    try {
      const go = (window as any).go?.main?.App
      if (go) {
        // Parse "SysOps.GetCPUInfo" into go.SysOps.GetCPUInfo(...)
        const parts = method.split('.')
        let target = go
        for (const part of parts) {
          target = target[part]
        }
        return await target(...args)
      }
    } catch (err) {
      console.warn(`Backend call failed: ${method}`, err)
    }
    // Fallback to mock
    const { mockData } = await import('@/lib/mockData')
    return mockData(method)
  }
  return { call }
}
```

Update `useEvents.ts` to use `window.runtime.EventsOn` directly with cleanup.

### 4.5 Shadcn/UI Component Installation

Install and configure shadcn/ui base components:
- `Button`, `Input`, `Select`, `Badge`, `Card`, `Tabs`, `Dialog`, `Drawer`, `Toast`
- Configure with squib color tokens via CSS variables
- Each component gets squib-style styling

### 4.6 Verify

```bash
cd cmd/hawkward-gui/frontend && npm run build   # Frontend builds
go build -o hawkward-gui.exe .                   # Go builds
wails build -o hawkward-gui.exe                  # Wails builds
```

---

## 5. Sprint 2: Dashboard + SysOps

### 5.1 Dashboard Page Redesign

**File**: `cmd/hawkward-gui/frontend/src/pages/Dashboard.tsx`

Current: 5 HealthCards row + 2 AreaCharts + quick action cards + alert badge
Target: Squib-inspired layout with:

```
┌─────────────────────────────────────────────────────┐
│  Dashboard                                        │
│  System overview and key metrics at a glance       │
├─────────────────────────────────────────────────────┤
│  ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐              │
│  │ CPU  │ │ MEM  │ │ DISK │ │ NET  │  ← 4 KPI     │
│  │ 45%  │ │ 62%  │ │ 78%  │ │ 1.2  │    cards     │
│  │ ▁▃▅▇ │ │ ▁▃▅▇ │ │ ▁▃▅▇ │ │ ▁▃▅▇ │  w/ sparkline│
│  └──────┘ └──────┘ └──────┘ └──────┘              │
│                                                     │
│  ┌─────────────────────┐ ┌─────────────────────┐   │
│  │ CPU Over Time       │ │ Memory Over Time    │   │
│  │  [area chart]       │ │  [area chart]       │   │
│  │  Threshold: 90% ─── │ │  Threshold: 90% ─── │   │
│  └─────────────────────┘ └─────────────────────┘   │
│                                                     │
│  ┌─────────────────────┐ ┌─────────────────────┐   │
│  │ Active Incidents    │ │ Group Health        │   │
│  │ • CPU spike (CRIT)  │ │ SysOps   ■■■■■■■□□□│   │
│  │ • Disk 85% (WARN)   │ │ NetOps  ■■■■□□□□□□│   │
│  │ • Mem leak (INFO)   │ │ SecOps  ■■■■■■■■■■│   │
│  └─────────────────────┘ └─────────────────────┘   │
│                                                     │
│  ┌─────────────────────────────────────────────┐    │
│  │ Uptime Strip: ▁▃▅▇▅▃▁▃▅▇▅▃▁▃▅▇▅▃▁▃▅▇▅▃▁... │    │
│  └─────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────┘
```

New components needed:
- **HeroSection**: Health gauge (SVG donut) + overall status + "N alarms" badge
- **KPIRow**: 4-column grid of KPI cards with sparkline, delta, trend arrow
- **IncidentsPanel**: Left panel in two-column grid, scrollable incident list
- **GroupHealthPanel**: Right panel with per-ops-layer health bars
- **UptimeStrip**: 90 thin bars, color-coded, showing 24h uptime history

### 5.2 SysOps Page Redesign

**File**: `cmd/hawkward-gui/frontend/src/pages/SysOps.tsx`

Current: Sub-tabbed (Overview, Processes, System Info) with basic bars and tables
Target: Squib sub-tab pattern with:

#### Overview Tab
- **CPU Section**: Detailed per-core bar chart (8+ cores, color-coded), load avg 1/5/15 line, frequency, temperature, model info
- **Memory Section**: Usage bar, swap usage, process memory breakdown (pie/treemap view)
- **Disk Section**: Per-partition usage bars, I/O read/write rates, filesystem type badges
- **System Info Card**: Hostname, OS, kernel, uptime, virtualization, process count

#### Processes Tab
- Sortable/filterable table: PID, Name, CPU%, Mem%, Status, FDs
- Search bar with debounced filtering
- Kill process button with confirmation dialog
- Process detail drawer (click row → slide-in with resource usage over time)
- Color-coded CPU/Mem bars in table rows

#### System Info Tab
- Clean key-value card layout with copy-to-clipboard
- Hardware details: CPU model, cores, RAM, disks
- OS details: version, build, kernel, boot time

---

## 6. Sprint 3: NetOps (The Star)

**This is the most feature-rich page — inspired by netscanner + hackingtool.**

**File**: `cmd/hawkward-gui/frontend/src/pages/NetOps.tsx`

Current: 6 sub-tabs (Ping, DNS, PortScan, Traceroute, Connections, Interfaces)
Target: Enhanced sub-tabs with real-time visualization:

### 6.1 Ping Tab

Current: Basic ping with stats + latency chart
Enhancements:
- **Real-time continuous ping**: Start/Stop button, table scrolls with new entries, latency renered as sparkline + mini chart
- **Multi-target**: Add multiple targets simultaneously, color-coded per target
- **Latency heatmap**: Matrix of targets × time, cell color = RTT
- **Packet loss gauge**: Arc gauge showing loss %
- **Statistics**: Min/Max/Avg/StdDev/Mdev, jitter calculation
- **Export**: Copy as CSV, JSON

### 6.2 DNS Tab

Current: Lookup with record sections
Enhancements:
- **Bulk lookup**: Paste multiple domains, parallel lookups with progress
- **DNS timing**: Query time per record type, resolver info
- **Reverse DNS**: IP → hostname lookup
- **DNSsec**: Status indicator for DNSSEC validation
- **History**: Recent lookups with timestamps, re-run button

### 6.3 Port Scan Tab

Current: Range/port list scan with progress bar, result table
Enhancements:
- **Real-time streaming**: Results appear as each port is scanned (netscanner-style)
- **Scan profiles**: Quick (top 20), Common (top 100), Full (1-65535), Custom
- **Service detection**: Color-coded service badges (HTTP=green, SSH=blue, SQL=red, etc.)
- **Port visualization**: Grid heatmap (ports × scan time), or horizontal bar chart
- **Bandwidth control**: Scan speed slider (slow/thorough ↔ fast/noisy)
- **Export**: Copy open ports, save to JSON/CSV
- **Nmap-style output**: Optional text view showing raw scan output

### 6.4 Traceroute Tab

Current: Animated hop-by-hop display
Enhancements:
- **Visual route map**: Geographic approximation of hops (country flags)
- **Latency sparkline per hop**: Mini chart showing 3 probes per hop
- **ASN/IP info**: Hover for whois/organization data
- **Comparison**: Run 2 traces side-by-side to compare routes
- **Export**: Save as text/JSON

### 6.5 Connections Tab

Current: Filtered table
Enhancements:
- **Auto-refresh**: Toggle with interval selector (1s/3s/5s/10s)
- **State distribution**: Donut chart of connection states (ESTABLISHED, LISTEN, TIME_WAIT, CLOSE_WAIT)
- **Process explorer**: Group by process name, expand to see connections per process
- **Geographic map**: IP → approximate location (future)
- **Kill connection**: Right-click → terminate (admin required)

### 6.6 Interfaces Tab

Current: Expandable interface cards with sparklines
Enhancements:
- **Bandwidth graph**: RX/TX line chart with auto-scaling Y axis
- **Interface matrix**: Grid of all interfaces, each showing name, IP, speed, status dot
- **Traffic breakdown**: Protocol distribution (TCP/UDP/ICMP) per interface
- **Speed test**: On-demand throughput test between interfaces
- **ARP table**: MAC-to-IP mapping view

### 6.7 New: Bandwidth Monitor Tab

- Real-time bandwidth graph with RX/TX stacked area chart
- Top talkers by process (requires elevated permissions)
- Historical bandwidth usage (hourly/daily/weekly views)
- Alerts on bandwidth threshold breach

---

## 7. Sprint 4: SecOps + DevOps + AIOps

### 7.1 SecOps Redesign

**File**: `cmd/hawkward-gui/frontend/src/pages/SecOps.tsx`

Current: 5 sub-tabs (Firewall, Users, Ports, Defender, Events)

Enhancements per tab:

- **Firewall**: Rule table with search/filter, enable/disable toggle, rule detail drawer, profile badges (Domain/Private/Public), rule statistics (count by action: ALLOW/BLOCK)
- **Users**: User table with admin badge, enable/disable toggle, group membership tags, last login info, account status indicator
- **Listening Ports**: Port-proc mapping, PID detail, protocol badges, state donut, filter by port range
- **Defender**: Service status cards with health indicators, scan buttons (Quick/Full/Custom), signature version age, real-time protection toggle, threat history
- **Events**: Virtual-scrolled event list, filter by level/provider/time range, severity color coding, event detail slide-in, export filtered

### 7.2 DevOps Redesign

**File**: `cmd/hawkward-gui/frontend/src/pages/DevOps.tsx`

Current: Terminal, Services, File Browser tabs

Enhancements:
- **Terminal**: Built-in terminal emulator (xterm.js or simple input/output), command history, working directory display, output with syntax highlighting, Ctrl+C support, session management
- **Services**: Service table with start/stop/restart buttons, status dots, start type badges, detail drawer with service dependencies, log snippet per service
- **File Browser**: Tree + detail view, file size formatting, mode/permissions display, search, directory navigation, open-in-explorer button
- **New: Container Dashboard**: Docker/Podman container list, status, logs, start/stop (if Docker detected)

### 7.3 AIOps Redesign

**File**: `cmd/hawkward-gui/frontend/src/pages/AIOps.tsx`

Current: Chat, Reports, Anomalies tabs

Enhancements:
- **Chat**: Full chat UI with markdown rendering, copy buttons, model selector, temperature slider, context window display, streaming response via Wails events
- **Reports**: Report template gallery (health, network, security, combined), preview panel, JSON-to-formatted report rendering, export to PDF/HTML, schedule recurring reports
- **Anomalies**: Anomaly timeline chart, severity heatmap, per-metric anomaly detection, threshold adjustment UI, acknowledge/dismiss actions, trend forecast overlay

---

## 8. Sprint 5: Network Designer + Logs + Settings

### 8.1 Network Designer Redesign

**File**: `cmd/hawkward-gui/frontend/src/pages/NetworkDesign.tsx`

Current: Canvas with draggable devices, connecting lines, properties panel
Enhancements:

- **Canvas improvements**: Zoom controls (in/out/reset/fit), pan via drag on empty space, grid snap, minimap overlay (bottom-right corner)
- **Device palette**: Sidebar with drag-to-add device types (Router, Switch, Server, Workstation, Firewall, Cloud, IoT, Load Balancer)
- **Connection types**: Ethernet (solid), Fiber (dashed), Wireless (dotted) with color coding
- **Properties panel**: Right-side slide-in with fields: label, IP, subnet, MAC, status, notes, tags, custom icon
- **Persistence**: Save to JSON file via Wails file dialog, load from file, recent files list
- **Export**: Export as PNG screenshot, SVG, JSON, network config snippet
- **Auto-layout**: Force-directed or hierarchical layout algorithm (d3-force or similar)
- **Status overlay**: Live device status (from backend ping/scan), color-coded borders, alert badge on failing devices
- **Subnet grouping**: Visual grouping of devices in same subnet (dashed bounding box)
- **Annotation**: Text labels, arrows, notes on canvas

### 8.2 Log Viewer Redesign

**File**: `cmd/hawkward-gui/frontend/src/pages/Logs.tsx`

Current: Simple log list
Enhancements:

- **Virtual scrolling**: Use `@tanstack/react-virtual` for 10K+ entries with zero DOM overhead
- **Filter bar**: Level filter (INFO/WARN/ERROR), source filter (sysops/netops/secops/devops/aiops), date range, search text (debounced)
- **Log level badges**: Color-coded: INFO=blue, WARN=amber, ERROR=red
- **Detail panel**: Click row → slide-in with full message, stack trace, source metadata, timestamp
- **Auto-scroll**: Toggle follow mode (auto-scroll to bottom on new entries)
- **Export**: Copy selected, export filtered view to TXT/JSON/CSV via Wails dialog
- **Clear**: Clear visible logs button
- **Statistics**: Log volume chart (entries/min by level), source breakdown
- **Tail mode**: Live streaming from backend with indicator showing rate

### 8.3 Settings Page Redesign

**File**: `cmd/hawkward-gui/frontend/src/pages/Settings.tsx`

Current: Basic settings form
Target: Multi-section settings page:

- **Theme**: Dark/Light toggle, accent color picker (purple, blue, green, orange)
- **Collection**: Interval selector (1s/3s/5s/10s/30s), metric retention (1h/6h/24h/7d)
- **Alerts**: Alert rule editor (add/edit/delete rules), notification settings, threshold defaults
- **Pipeline**: Pipeline status, metrics per series, data age, manual refresh/reset button
- **Network**: Default scan ports, ping count, DNS resolver, timeout settings
- **AIOps**: Ollama endpoint URL, default model, max context length, temperature default
- **About**: Version info, build date, uptime, license, links

---

## 9. Sprint 6: Polish, Performance, Packaging

### 9.1 Error Handling

- Error boundaries for each page (React error boundary component)
- Backend health check with auto-reconnect UI
- Graceful degradation when backend unavailable (show cached/stale data)
- Toast notifications for all async errors

### 9.2 Performance

- Virtual scrolling for all long lists (logs, connections, processes, events)
- Memoized chart components with `React.memo` + `useMemo`
- Debounced search inputs (300ms)
- Lazy-loaded page components via `React.lazy` + `Suspense`
- Optimized re-renders via proper `useCallback` in hooks

### 9.3 Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| `Ctrl+K` or `?` | Command palette |
| `Ctrl+1-9` | Navigate to page |
| `Ctrl+D` | Dashboard |
| `Ctrl+S` | SysOps |
| `Ctrl+N` | NetOps |
| `Ctrl+,` | Settings |
| `Ctrl+Shift+E` | Export current view |
| `Escape` | Close drawer/modal |

### 9.4 Packaging

- Wails build with production frontend embed
- NSIS installer script for Windows (generated by Wails)
- First-run setup: WebView2 check, config directory creation
- Auto-update infrastructure (future)

---

## 10. Frontend Component Tree

```
App
├── Sidebar (252px)
│   ├── Brand (logo + name + tagline)
│   ├── NavSection "Operations"
│   │   ├── NavItem × 6 (ops layers + badges)
│   ├── NavSection "Tools"
│   │   ├── NavItem × 3 (Network Design, Logs, Settings)
│   └── Footer (version, health status)
├── Main Area (flex-1)
│   ├── TopBar (60px, glass)
│   │   ├── Breadcrumb (Hawkward > Page)
│   │   ├── Live Status Dot (pulsing)
│   │   ├── Notification Bell
│   │   └── Theme Toggle
│   └── Page Content (scrollable, 24px padding)
│       ├── Dashboard
│       │   ├── HeroSection (gauge + status + uptime strip)
│       │   ├── KPIRow × 4
│       │   ├── ChartsRow (AreaChart × 2)
│       │   ├── PanelsRow (IncidentsPanel + GroupHealthPanel)
│       │   └── QuickActionGrid
│       ├── SysOps
│       │   ├── TabBar (Overview | Processes | System Info)
│       │   ├── [Overview] CPUBlock, MemBlock, DiskBlock, SysInfoCard
│       │   ├── [Processes] SearchBar, ProcessTable, ProcessDrawer
│       │   └── [System Info] InfoCardGrid
│       ├── NetOps
│       │   ├── TabBar (Ping | DNS | Scan | Trace | Connections | Interfaces | Bandwidth)
│       │   ├── [Ping] TargetInput, StartStop, LiveTable, LatencyChart, StatsRow
│       │   ├── [DNS] HostnameInput, RecordCards, TimingInfo, History
│       │   ├── [PortScan] TargetInput, ProfileSelect, Progress, ResultTable, ServiceLegend
│       │   ├── [Traceroute] TargetInput, HopVisualization, MapOverlay
│       │   ├── [Connections] StateDonut, FilterBar, ConnectionTable
│       │   ├── [Interfaces] InterfaceGrid, BandwidthChart
│       │   └── [Bandwidth] RX/TX Chart, TopTalkers, HistorySelector
│       ├── SecOps (5 sub-tabs with tables, toggles, detail drawers)
│       ├── DevOps (Terminal, Services, File Browser, Containers)
│       ├── AIOps (Chat, Reports, Anomalies)
│       ├── NetworkDesign (Canvas, Palette, Properties, Toolbar, MiniMap)
│       ├── Logs (FilterBar, VirtualList, DetailPanel, Stats)
│       └── Settings (Sectioned forms, toggles, sliders, about card)
├── Global Drawer (right slide-in)
├── Global Modal (centered)
└── Toast Container (bottom-right)
```

---

## 11. Real-Time Data Architecture

### Event Flow

```
Go tick loop (3s)
  │
  ├─ Emit: "metrics" → Dashboard, SysOps, NetOps update gauges/charts
  ├─ Emit: "alert" → AlertBadge flashes, toast appears, bell badge updates
  ├─ Emit: "log" → LogViewer appends entry
  │
  Go on-demand (user actions)
  │
  ├─ Ping("8.8.8.8", 4) → returns PingResult, frontend renders table
  ├─ PortScan("192.168.1.1", [22,80,443]) → returns PortResult[]
  ├─ GetCPUInfo() → CPUInfo, renders per-core bars
  ├─ KillProcess(1234) → returns success/error, refreshes process list
  │
  Frontend hooks (typed events)
  │
  ├─ useMetrics() → { cpu, mem, disk, net } updates every 3s
  ├─ useAlerts() → { alerts[], count } updates on fire/resolve
  ├─ useLogStream() → appends to virtual list
  └─ useBackend() → proxy for all on-demand Go calls
```

### Event Names (from `internal/app/Events.go`)

| Event | Payload | Frequency | Consumers |
|-------|---------|-----------|-----------|
| `metrics` | `MetricsEvent` | Every 3s | Dashboard, SysOps |
| `alert` | `AlertEvent` | On fire/resolve | AlertBadge, Toasts, Bell |
| `log` | `LogEvent` | On write | LogViewer |
| `pipeline` | `PipelineEvent` | On status change | Settings |
| `cmd:line` | `string` | Live output | DevOps Terminal |
| `cmd:done` | `string` | On command end | DevOps Terminal |

---

## 12. Feature Inventory by Page

### Dashboard
- [x] Page exists with HealthCards + area charts + alerts
- [ ] Redesign to squib layout with hero, KPIs, incidents, group health
- [ ] Wire to real backend events (metrics event)
- [ ] Add uptime strip visualization
- [ ] Add live status indicator (pulsing dot)
- [ ] Add system info summary card

### SysOps
- [x] Page exists with basic structure
- [ ] Redesign to squib panels with proper sub-tabs
- [ ] Wire CPU, Memory, Disk, Processes, SystemInfo to real Go bindings
- [ ] Add per-core CPU bar chart
- [ ] Add memory breakdown visualization
- [ ] Add I/O stats to disk view
- [ ] Add process kill with confirmation dialog
- [ ] Add process detail drawer

### NetOps
- [x] Page exists with mock-powered tabs
- [ ] Wire Ping, DNS, PortScan, Traceroute to real Go bindings
- [ ] Add real-time continuous ping
- [ ] Add multi-target ping
- [ ] Add port scan progress bar
- [ ] Add service badge color system
- [ ] Add connection state donut chart
- [ ] Add interface bandwidth graph
- [ ] Add bandwidth monitor tab
- [ ] Add DNS history/comparison

### SecOps
- [x] Page exists (partial)
- [ ] Wire Firewall, Users, Ports, Defender, Events to real Go bindings
- [ ] Add firewall rule toggle
- [ ] Add user enable/disable
- [ ] Add security event timeline
- [ ] Add defender health cards

### DevOps
- [x] Page exists (partial)
- [ ] Wire Terminal, Services, File Browser to real Go bindings
- [ ] Add terminal emulator
- [ ] Add service start/stop/restart
- [ ] Add container dashboard

### AIOps
- [x] Page exists (partial)
- [ ] Wire Chat, Reports, Anomalies to real Go bindings
- [ ] Add streaming chat
- [ ] Add report preview
- [ ] Add anomaly timeline

### Network Design
- [x] Page exists with canvas + drag-drop
- [ ] Add persist/load via Wails file dialog
- [ ] Add minimap
- [ ] Add zoom controls
- [ ] Add device palette
- [ ] Add subnet grouping
- [ ] Add live status overlay
- [ ] Add auto-layout
- [ ] Add export (PNG, SVG, JSON)

### Logs
- [x] Page exists (basic)
- [ ] Add virtual scrolling
- [ ] Add filter bar with level/source/date/search
- [ ] Add log detail drawer
- [ ] Add export functionality
- [ ] Add log statistics chart
- [ ] Add tail/follow mode

### Settings
- [x] Page exists (basic)
- [ ] Add theme picker with accent colors
- [ ] Add collection interval config
- [ ] Add alert rule editor
- [ ] Add pipeline status
- [ ] Add about section

---

## 13. Testing Strategy

### Backend Tests (Go)
```bash
go test ./internal/common/...       # 72+ tests (already exist)
go test ./internal/sysops/...       # System operations (already exist)
go test ./internal/netops/...       # Network operations (already exist)
go test ./internal/app/...          # Wails bindings (ADD)
```

### Frontend Tests (Vitest + React Testing Library)
```bash
cd cmd/hawkward-gui/frontend
npm test                             # Component tests
```

Test targets:
1. **Component rendering**: Dashboard, HealthCard, Gauge, NetOps tabs, ProcessManager
2. **Hook tests**: useBackend, useEvents, useTheme (mock Wails runtime)
3. **Integration**: Page components with mocked backend data
4. **Accessibility**: axe-core assertions on all interactive components

### E2E (Manual via Wails build)
```bash
wails build && ./build/bin/hawkward-gui.exe
```

---

## 14. Build & Dev Commands

```bash
# Dev mode (frontend only, mock data)
cd cmd/hawkward-gui/frontend && npm run dev

# Full Wails build
wails build -o hawkward-gui.exe

# Go backend only
go build -o hawkward-gui.exe .

# Tests
go test ./...
cd cmd/hawkward-gui/frontend && npm test

# Frontend production build
cd cmd/hawkward-gui/frontend && npm run build

# Frontend checks
cd cmd/hawkward-gui/frontend && npx tsc --noEmit   # TypeScript check
cd cmd/hawkward-gui/frontend && npm run lint        # Lint

# Clean
rm -rf build/bin/hawkward-gui.exe
rm -rf cmd/hawkward-gui/frontend/dist
```

---

## Appendix A: File Change Summary

| File | Action | Sprint |
|------|--------|--------|
| `cmd/hawkward-gui/frontend/src/styles/globals.css` | Rewrite (squib tokens) | 1 |
| `cmd/hawkward-gui/frontend/index.html` | Add Inter + JetBrains Mono fonts | 1 |
| `cmd/hawkward-gui/frontend/src/hooks/useBackend.ts` | Rewrite (real Wails calls) | 1 |
| `cmd/hawkward-gui/frontend/src/hooks/useEvents.ts` | Rewrite (real EventsOn) | 1 |
| `cmd/hawkward-gui/frontend/src/hooks/useTheme.ts` | Enhance (CSS variables) | 1 |
| `cmd/hawkward-gui/frontend/src/App.tsx` | Layout update (squib shell) | 1 |
| `cmd/hawkward-gui/frontend/src/components/layout/Sidebar.tsx` | Redesign | 1 |
| `cmd/hawkward-gui/frontend/src/components/layout/TopBar.tsx` | Redesign | 1 |
| `cmd/hawkward-gui/frontend/src/components/ui/*.tsx` | Create (shadcn-inspired) | 1 |
| `cmd/hawkward-gui/frontend/src/pages/Dashboard.tsx` | Rewrite | 2 |
| `cmd/hawkward-gui/frontend/src/components/dashboard/*` | Rewrite | 2 |
| `cmd/hawkward-gui/frontend/src/pages/SysOps.tsx` | Rewrite | 2 |
| `cmd/hawkward-gui/frontend/src/pages/NetOps.tsx` | Rewrite | 3 |
| `cmd/hawkward-gui/frontend/src/pages/SecOps.tsx` | Rewrite | 4 |
| `cmd/hawkward-gui/frontend/src/pages/DevOps.tsx` | Rewrite | 4 |
| `cmd/hawkward-gui/frontend/src/pages/AIOps.tsx` | Rewrite | 4 |
| `cmd/hawkward-gui/frontend/src/pages/NetworkDesign.tsx` | Enhance | 5 |
| `cmd/hawkward-gui/frontend/src/pages/Logs.tsx` | Rewrite | 5 |
| `cmd/hawkward-gui/frontend/src/pages/Settings.tsx` | Rewrite | 5 |
| `cmd/hawkward-gui/frontend/src/lib/mockData.ts` | Reduce (dev-only) | 1 |
| `internal/app/*.go` | Minor fixes/adjustments | 1-5 |
| `main.go` | Unchanged | — |

---

## Appendix B: Dependency Plan

### Current (already installed)
- `recharts` — Area charts, line charts
- `lucide-react` — Icon library
- `clsx` / `tailwind-merge` — CSS utilities
- `react`, `react-dom` (TypeScript) — Core
- `tailwindcss` v4 — CSS framework
- `vite` — Build tool

### To Install (Sprint 1)
- `@radix-ui/react-dialog` — Modal primitives
- `@radix-ui/react-drawer` — Drawer primitives
- `@radix-ui/react-toast` — Toast notifications
- `@radix-ui/react-tabs` — Tab primitives
- `@radix-ui/react-select` — Select dropdowns
- `@radix-ui/react-switch` — Toggle switches
- `@radix-ui/react-slider` — Slider controls
- `class-variance-authority` — Component variants (shadcn-style)

### To Install (Sprint 5)
- `@tanstack/react-virtual` — Virtual scrolling for logs
- `d3-force` — Force-directed network layout (optional)

### To Install (Sprint 6)
- `@axe-core/react` — Accessibility testing
- `vitest` + `@testing-library/react` — Frontend testing

---

> **Next Step**: Begin Sprint 1 — the design system overhaul and backend wiring are the critical foundation everything else depends on. Start with `globals.css` squib token rewrite, then `useBackend.ts` real Wails wiring, then layout components, then each page in order.
