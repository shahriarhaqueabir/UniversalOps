# Library & Open-Source Research — Hawkward Operations Platform

> **Date**: 2026-07-10
> **Context**: Sprint 19 — Production Stabilization
> **Goal**: Identify libraries, tools, and open-source repos that could enhance the project

## Methodology

Each candidate was evaluated against:
1. **Problem solved** — does it address a real gap or pain point?
2. **Integration effort** — drop-in replacement vs. new subsystem
3. **Version compatibility** — works with Go 1.26.5, React 19, Vite 6, Tailwind v4
4. **Risk** — breaking changes, API drift, maintenance status

---

## 1. Quick Wins (drop-in / low effort)

### 1.1 `tailwindcss-animate` (Tailwind v4 compatible)
- **Solves**: `animate-in`, `fade-in`, `zoom-in-95`, `slide-in-from-*` classes currently emulated with manual CSS keyframes
- **Why**: Reduces 50+ lines of manual CSS; gives access to Tailwind's animation utility classes natively
- **Integration**: `npm install tailwindcss-animate` + register plugin. Remove the manual `@keyframes` fallback block in `globals.css` (lines 102-203)
- **Risk**: None — pure CSS utility, no runtime impact
- **Priority**: ⭐⭐⭐ (P0)

### 1.2 `zustand` (v5.x)
- **Solves**: Cross-page state sharing without prop drilling — currently each page fetches its own data
- **Use cases**: Shared alert state across TopBar badge + Dashboard; shared settings/theme; shared Ollama status
- **Integration**: `npm install zustand`. Create standalone stores, no Provider wrapper needed
- **Risk**: Low — 2.8kB, no deps, React 19 compatible
- **Priority**: ⭐⭐ (P1)

### 1.3 `sonner` (toast notifications)
- **Solves**: Alert events are emitted but have no visual notification system — alerts appear only when user navigates to Dashboard
- **Why**: Real-time toast for alert firings, command completions, errors
- **Integration**: `npm install sonner`. Wrap `<App>` in `<Toaster />`. Subscribe to alert events
- **Risk**: Low — 5kB, SSR-safe, React 19 compatible
- **Priority**: ⭐⭐ (P1)

### 1.4 `date-fns` (v4.x)
- **Solves**: Manual `new Date().toLocaleTimeString()` and ad-hoc formatting scattered across pages
- **Why**: Consistent, locale-aware time formatting with tree-shakeable imports
- **Integration**: `npm install date-fns`. Replace manual formatting calls
- **Risk**: Low — 0 deps, v4 works with ESM
- **Priority**: ⭐ (P2)

---

## 2. Medium Effort (new capabilities, moderate integration)

### 2.1 `@tanstack/react-virtual` (v3.x)
- **Solves**: Custom virtual scrolling in Logs.tsx (current implementation has offset bugs)
- **Why**: Handles variable row heights, scroll-to-index, overscan, and measurement correctly
- **Integration**: Replace ~60 lines of custom virtual math with `<Virtualizer>` component
- **Risk**: Low — mature library, React 19 compatible
- **Priority**: ⭐⭐⭐ (P0 - bugfix)

### 2.2 `@tanstack/react-query` (v5.x)
- **Solves**: Manual `useEffect` + `setInterval` data fetching pattern repeated in every page
- **Why**: Automatic refetching, stale-while-revalidate, loading/error states, cache dedup
- **Integration**: Wrap app in `<QueryClientProvider>`. Replace `useEffect` + `call(...)` patterns
- **Risk**: Medium — requires larger refactor of all pages but can be incremental
- **Priority**: ⭐⭐ (P1)

### 2.3 `motion` (formerly framer-motion, v12.x)
- **Solves**: No page transition animations, no layout animations
- **Why**: Adds `AnimatePresence` for page transitions, `motion.div` for micro-interactions
- **Integration**: `npm install motion`. Wrap `<Suspense>` in `<AnimatePresence>`. Replace some CSS transitions
- **Risk**: Medium — 30kB, but already has 70+ animation presets. React 19 compatible
- **Priority**: ⭐ (P2)

### 2.4 `react-hot-toast` (alternative to sonner, wider ecosystem)
- **Solves**: Same as sonner but with more customization options
- **Why**: Already many existing integrations and theming examples
- **Integration**: Same pattern as sonner
- **Priority**: Same as sonner (choose one)

### 2.5 `nanoid` / `uuid`
- **Solves**: NetworkDesign's `genId()` generates sequential IDs that could collide
- **Why**: Proper unique ID generation for topology devices/connections
- **Integration**: Replace `let nextId = 1` with `nanoid()`
- **Risk**: Low — 1kB, pure function
- **Priority**: ⭐ (P3)

---

## 3. Go Backend Libraries

### 3.1 `gopacket` (google/gopacket)
- **Solves**: Packet-level network analysis — currently NetOps only has ping/DNS/ports
- **New capability**: Live packet capture, protocol analysis, bandwidth per-flow
- **Integration**: Requires admin/root on Windows (WinPcap/Npcap). New Go dependency
- **Risk**: High — requires Npcap runtime dependency, elevated privileges
- **Priority**: ⭐ (P3 — feature addition)

### 3.2 `prometheus/client_golang` (v1.x)
- **Solves**: No metrics export — pipeline data is only displayed in-GUI
- **New capability**: Expose collected metrics as Prometheus endpoint for external monitoring
- **Integration**: Add HTTP endpoint on localhost (separate port). Register collectors that wrap DataPipeline
- **Risk**: Low — well-established library. Adds ~2MB to binary
- **Priority**: ⭐ (P4 — advanced feature)

### 3.3 `rs/zerolog` (v1.x)
- **Solves**: Custom `common.LogInfo` / `common.LogWarn` logger (simple leveled logging)
- **Upgrade**: Structured JSON logging with fields, stack traces, sampling
- **Integration**: Replace `common.Logger` calls with zerolog. Keep existing log file output
- **Risk**: Low — drop-in replacement for current logging API
- **Priority**: ⭐ (P3 — nice to have)

### 3.4 `ollama` Go SDK (via REST or amriksingh/ollama-go)
- **Solves**: AIOps currently calls Ollama via raw `http.Post` / REST
- **Upgrade**: Typed client with streaming support, model management
- **Integration**: Wrap current HTTP calls with SDK client. Add streaming for chat responses
- **Risk**: Low — just a convenience wrapper
- **Priority**: ⭐ (P3)

---

## 4. Critical Issues Found (Fix Now)

### 4.1 NetworkDesign wailsjs imports
- **File**: `src/pages/NetworkDesign.tsx`
- **Issue**: Imports `SaveFileDialog`, `OpenFileDialog`, `WriteFile`, `ReadFile` from `@/wailsjs/go/app/App`
- **Risk**: If Wails bindings haven't generated these wrappers, the build will fail or runtime calls will error
- **Fix**: These are Wails auto-generated bindings — they should be present after `wails build` generates them. The current implementation uses them correctly but the fallback path (raw `window.go.*`) is needed for dev mode

### 4.2 Logs virtual scrolling offset bug
- **File**: `src/pages/Logs.tsx`
- **Issue**: `totalHeight` calculation uses simplified formula that doesn't account for multiple expanded rows simultaneously. The current design only allows one expanded row at a time, so it works, but the calculation is fragile
- **Fix**: No immediate fix needed — single-expand pattern is correct. Flag for future enhancement

### 4.3 Missing frontend tests
- **Directory**: `src/test/` only contains `setup.ts`, no actual tests
- **Issue**: vitest is fully configured but there are 0 test files
- **Fix**: Add component tests for at least ErrorBoundary, ConfirmDialog, and key utility functions

### 4.4 Dashboard alert state pattern
- **File**: `src/pages/Dashboard.tsx` line 208
- **Issue**: `const setAlerts = useState<AlertInfo[]>([])[1]` — only the setter is extracted, never the value
- **Impact**: None functionally, but wastes a state slot and is confusing
- **Fix**: Consider removing the unused state or using a ref instead

---

## 5. Open-Source Repos for Inspiration

| Repo | Why | What to learn |
|------|-----|---------------|
| `shadcn/ui` | Design system patterns | Component API design, Radix UI composition (already partially used) |
| `netdata/netdata` | Real-time monitoring UI | Dashboard layout, metric visualization patterns |
| `grafana/grafana` | Dashboard composition | Panel system, data source abstraction |
| `wailsapp/wails` | Framework updates | Version 3 migration path (when stable) |
| `getlago/lago` | Usage metering | Not directly applicable but good architecture patterns |
| `microsoft/windows-dev-box-setup-scripts` | Windows automation | PowerShell workflow scripts for DevOps |

---

## 6. Priority Ranking Summary

| Priority | Library | Effort | Category |
|----------|---------|--------|----------|
| P0 | `tailwindcss-animate` | 5 min | CSS cleanup |
| P0 | `@tanstack/react-virtual` | 30 min | Bugfix |
| P1 | `zustand` | 1 hr | Architecture |
| P1 | `sonner` | 30 min | UX |
| P1 | `@tanstack/react-query` | 2-4 hrs | Architecture |
| P2 | `date-fns` | 30 min | Consistency |
| P2 | `motion` | 1 hr | Animation |
| P3 | `nanoid` | 5 min | Correctness |
| P3 | `zerolog` | 1 hr | Logging |
| P3 | Ollama Go SDK | 30 min | Cleanup |
| P4 | `prometheus/client_golang` | 2 hrs | Export |
| P4 | `gopacket` | 4+ hrs | Feature |

---

## 7. What NOT to Add (Anti-Recommendations)

- ❌ **Redux, MobX, Recoil, Jotai** — overkill for this app's state pattern (page-local + event-driven)
- ❌ **Wails v3** — not stable enough for production migration
- ❌ **gRPC / protobuf** — would require replacing Wails bridge, massive effort for no gain
- ❌ **Docker containerization** — this is a native desktop app, containerization is counterproductive
- ❌ **Chart.js / D3.js** — Recharts is already serving the use case well; upgrade to `uplot` if performance becomes an issue
- ❌ `go-sqlite3` (CGo) — `modernc.org/sqlite` (pure Go) is already in use and correctly chosen for cross-platform builds
