# Codebase Research & Analysis

## Project Overview
- **Name**: Hawkward Operations Platform (AllOpsFull)
- **Type**: Native desktop app (Wails v2 - Go backend + React/TypeScript/Vite frontend)
- **Version**: 1.1.1
- **License**: MIT

## Tech Stack
- **Backend**: Go 1.26.5 with Wails v2, gopsutil/v4, miekg/dns, modernc.org/sqlite
- **Frontend**: React 19 + TypeScript + Vite 6 + Tailwind v4 + Recharts 3 + Lucide React
- **UI Components**: Radix UI (Avatar, Dialog, Dropdown Menu, Scroll Area, Select, Separator, Slider, Switch, Tabs, Tooltip)
- **Styling**: Squib-inspired dark theme with CSS custom properties, Tailwind v4 utility classes
- **State**: React hooks (useState, useEffect, useCallback, useMemo, useRef)

## Architecture
- **Backend pattern**: Wails Bindings — Go structs exposed to JS via `window.go.*`
- **Frontend pattern**: Page-based routing through sidebar navigation
- **Data flow**: Go backend collects metrics via tick loop (default 3s) → pushes to DataPipeline → emits runtime events to frontend
- **Persistence**: SQLite (WAL mode) via modernc.org/sqlite, buffered writes, 7-day auto-prune

## Internal Packages (Go)
- `internal/app/` — Wails binding facades (Dashboard, SysOps, NetOps, SecOps, DevOps, AIOps, PipelineAPI, AlertAPI, Logs)
- `internal/common/` — Shared: DataPipeline, AlertEngine, Storage (SQLite), Logger, ForecastEngine, Sandbox, Theme, Types
- `internal/sysops/` — CPU, Memory, Disk, Process monitoring via gopsutil
- `internal/netops/` — Ping, DNS, PortScan, Traceroute, Connections, Interfaces via miekg/dns + golang.org/x/net
- `internal/secops/` — Firewall (netsh), Defender (powershell), Users, Events, Listening ports, Tasks
- `internal/devops/` — Shell (sandboxed), FileBrowser, Services (powershell), LogTail, Workflows
- `internal/aiops/` — Ollama client (REST API), Report generation, State query, Anomaly detection

## Frontend Structure
- `src/App.tsx` — Root layout: Sidebar + TopBar + MainContent
- `src/pages/` — Dashboard, SysOps, NetOps, SecOps, DevOps, AIOps, NetworkDesign, Logs, Settings
- `src/components/` — layout (Sidebar, TopBar, MainContent), ui (ErrorBoundary), dialogs (ConfirmDialog, ExportDialog, SettingsDialog), dashboard (AlertBadge, HealthCard), charts (AreaChart, Gauge, MiniSparkline), network, logs
- `src/hooks/` — useBackend (Wails bridge calls), useEvents (runtime event subscription), useTheme (dark/light)
- `src/types/` — Full TypeScript mirrors of Go binding types
- `src/lib/` — utils (cn function), constants

## State of the Art Assessment

### Strengths
1. **Modern frontend stack**: React 19 + Vite 6 + Tailwind v4 with latest versions
2. **Type safety**: Full TypeScript coverage with mirrored Go types
3. **Performance**: Virtual scrolling in Logs, code-splitting (lazy routes), memo'd components
4. **UX**: Squib-inspired dark theme, consistent design language, responsive layout
5. **Error handling**: ErrorBoundary wrapper, ConfirmDialog for destructive actions, sandboxed terminal
6. **Data pipeline**: Well-designed time-series ingestion with forecasting and trend detection
7. **Alert engine**: Rule-based alerting with flap detection

### Areas for Improvement
1. **CSS variable naming**: Uses kebab-case in CSS (`--color-bg`) but Tailwind doesn't map to utilities natively without `theme()` references — need to use CSS vars via `var(...)` function calls in classes
2. **Accessibility**: Missing ARIA labels on some interactive elements, no focus indicators visible for keyboard nav
3. **Responsiveness**: Some hardcoded widths (`w-[280px]` sidebar) that could collapse better on small screens
4. **Loading states**: Some pages show loading skeleton but others (Dashboard) show a single text message
5. **Empty states**: Some tables handle empty state gracefully, others could improve
6. **Animation**: Use of `animate-in` classes that require Tailwind plugin (possibly from `tailwindcss-animate`)
7. **Backend resilience**: Go backend has error wrapping and fallbacks but some edge cases could be hardened

## Standard Practices vs Current Implementation

### What matches modern standards:
- ✅ Component-level code splitting with `React.lazy`
- ✅ TypeScript strict with full type definitions
- ✅ CSS custom properties for theming (dark/light)
- ✅ Recharts for charting with responsive containers
- ✅ Radix UI primitives for accessibility
- ✅ Error boundaries
- ✅ Debounced/throttled data fetching
- ✅ Virtual scrolling for large lists

### What could be improved:
- ⚠️ Some inline styles remain that could use Tailwind utility classes more consistently
- ⚠️ CSS variable fallthrough: `bg-[var(--color-panel)]` pattern instead of proper Tailwind theme config
- ⚠️ No focus-visible styles visible on interactive elements
- ⚠️ Dashboard loading state is a simple text string, not a skeleton
- ⚠️ Some hardcoded values vs CSS variables for colors

## Documentation References
- Wails v2: https://wails.io/docs/
- Tailwind v4: https://tailwindcss.com/docs
- Recharts: https://recharts.org/en-US/guide
- Radix UI: https://www.radix-ui.com/primitives
- Lucide Icons: https://lucide.dev/icons
- gopsutil: https://github.com/shirou/gopsutil
- modernc.org/sqlite: https://pkg.go.dev/modernc.org/sqlite
