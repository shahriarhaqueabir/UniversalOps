# Hawkward — Workspace Memory

## Active Session
- **Sprint 21.5: Dashboard Functionality, Prometheus Integration, CI Hardening** ✅ DONE

## Completed Work

| ID | Description | Status |
|----|-------------|--------|
| T-01 → T-18 | Sprints 1–20 | ✅ |
| T-19 | Dashboard Quick Diag + Generate Briefing backend actions | ✅ |
| T-20 | Dashboard UI: dynamic red flags, diag/briefing overlays | ✅ |
| T-21 | Prometheus metrics exporter (port 9210) | ✅ |
| T-22 | TestPing Linux CI skip logic improved | ✅ |
| T-23 | README badges fixed for private repo | ✅ |
| T-24 | npm lockfile synced (typescript reinstalled) | ✅ |
| T-25 | Compute Logic Analysis now dynamic with real data | ✅ |

## Git State (Current)
- **Branch**: `main` — has uncommitted changes
- **HEAD**: `b060cee` — `fix: missing comma in wails.json blocking wails build`
- **Tags**: `v1.3.0` (pushed), `v1.2.0`, `v1.1.1`, `v1.1.0`, `v1.0.0`

## Fixes Applied This Session

| # | Issue | File | Fix |
|---|-------|------|-----|
| 1 | Dashboard "Quick Diag" button just navigated to SysOps | `Dashboard.tsx`, `Dashboard.go` | Added `RunQuickDiag()` backend call + results overlay |
| 2 | Dashboard "Generate Briefing" button just navigated to AIOps | `Dashboard.tsx`, `Dashboard.go` | Added `GenerateDashboardBriefing()` backend call + briefing overlay |
| 3 | Compute Logic Analysis was 100% hardcoded text | `Dashboard.tsx` | Added `computeRedFlags()` that uses real pipeline data |
| 4 | TestPing fails on Linux CI | `netops_test.go`, `ping.go` | Added `exec.LookPath` check, fixed Linux ping flags (`-c`/`-W` vs `-n`) |
| 5 | Typescript not installed in node_modules | `package.json` | Reinstalled typescript (`npm install typescript@~5.7.2`) |
| 6 | No Prometheus metrics endpoint | `metrics_exporter.go`, `App.go` | Added minimal HTTP server on :9210 with `/metrics`, wired into tick loop |
| 7 | `GetAlerts()` not exposed on AlertEngine | `alerts.go` | Added `GetAlerts()` method for dashboard briefing |
| 8 | README badges unclear for private repo | `README.md` | Added private repo notice, added platform badge |

## Library Inventory

### Installed (Frontend)
| Library | Version | Purpose |
|---------|---------|---------|
| `@radix-ui/*` (10 packages) | latest | Accessible UI primitives |
| `@tanstack/react-query` | v5.101+ | Data fetching & caching |
| `@tanstack/react-table` | v8.21+ | Table rendering |
| `@tanstack/react-virtual` | v3.14+ | Virtual scrolling |
| `zustand` | v5 | State management |
| `recharts` | v3 | Charts |
| `lucide-react` | v1.23+ | Icons |
| `motion` | v12 | Animations |
| `sonner` | v2 | Toast notifications |
| `class-variance-authority` | v0.7+ | Class variants |
| `date-fns` | v4 | Date formatting |
| `nanoid` | v5 | ID generation |
| `tailwindcss` | v4 | Styling |
| `Tailwind v4 Vite plugin` | v4.3+ | Tailwind integration |

### Installed (Go Backend)
| Library | Purpose |
|---------|---------|
| `gopsutil/v4` | System metrics (CPU, memory, disk, processes) |
| `miekg/dns` | DNS lookups |
| `golang.org/x/net` | ICMP ping, network operations |
| `modernc.org/sqlite` | Embedded SQLite database |
| `ollama/ollama/api` | AI chat via Ollama |
| `rs/zerolog` | Structured logging |
| `prometheus/client_golang` | **NEW** — Prometheus metrics exposition |

### Pending (P4)
| Library | Why Deferred |
|---------|-------------|
| `google/gopacket` | Needs Npcap runtime on Windows (libpcap on Linux) |

## Dashboard Architecture

### Data Flow
1. **Initial load**: `useQuery` calls `Dashboard.GetDashboardData()` → reads `DataPipeline` metrics → populates KPI cards, hero section
2. **Live updates**: `useEvents('metrics')` receives 3-second ticks from `App.collectAndEmit()` → updates DashboardData and CPU timeline
3. **Quick Diag**: Button calls `Dashboard.RunQuickDiag()` → returns categorized status (pass/warn/fail) with messages
4. **Briefing**: Button calls `Dashboard.GenerateDashboardBriefing()` → returns sections (CPU, Memory, Disk, Network, Alerts)
5. **Red Flags**: Computed client-side from DashboardData — dynamic based on actual metric values

### Button Behavior (Fixed)
- **QUICK DIAGNOSTIC**: NOW calls `Dashboard.RunQuickDiag()` → shows overlay with results. NOT just navigation.
- **GENERATE BRIEFING**: NOW calls `Dashboard.GenerateDashboardBriefing()` → shows overlay with briefing sections. NOT just navigation.

## Known Remaining Issues
1. **Release pipeline**: `softrops/action-gh-release` with `GITHUB_TOKEN` can't create releases on private repos. Needs PAT in `GH_TOKEN` secret.
2. **TestPing**: May fail on Linux CI (no root for raw sockets + ping binary needs CAP_NET_RAW) — improved skip logic.
3. **TestInsertLogAndQuery**: Flaky in full suite — SQLite DB leaks between tests.
4. **Recharts ResponsiveContainer**: Stderr warnings in jsdom tests — cosmetic (jsdom has no layout).
5. **Frontend test coverage**: Missing for TopBar, NetOps, SecOps, DevOps, AIOps, NetworkDesign — P3.
6. **gopacket**: P4 — needs Npcap runtime on Windows.
7. **"Cannot find module" LSP diagnostics for wailsjs**: Stale — packages exist and build succeeds.
8. **NetworkDesign**: Hardcoded seed devices on first visit — BY DESIGN (example topology).
9. **npm ci on Windows**: EPERM errors on `lightningcss.win32-x64-msvc.node` due to file locks — CI uses Linux, unaffected.
