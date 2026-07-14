# OpsForAll — Workspace Memory

## Project Identity
- **Name**: OpsForAll
- **Persona**: Nostalgic Professional (Reliable, high-performance universal utility).
- **Core Value**: Universal deep system visibility and automation for everyone.

## Active Session
- **Sprint 29: DevOps Foundation & Actionability** 🔄 IN PROGRESS
- **Sprint 28: Layer-Driven Code Review** ✅ DONE

## Completed Work

| ID | Description | Status |
|----|-------------|--------|
| T-01 → T-37 | Sprints 1–23 | ✅ |
| T-38 | Fix AIOps type mismatch & build errors | ✅ |
| T-39 | Clean temp diagnostic files | ✅ |
| T-40 | Wire SecOps Scheduled Tasks | ✅ |
| T-41 | Wire DevOps Tool Detection | ✅ |
| T-42 | Wire AIOps Session Management | ✅ |
| T-43 | Wire Dashboard Timeline Summary | ✅ |
| T-44 | Dashboard Forecast Chart integration | ✅ |
| T-45 | AI Event Explanation on Dashboard | ✅ |
| T-46 | Backend Topology Analysis & Inventory | ✅ |
| T-47 | Alert Rules Management UI | ✅ |
| T-48 | Process Tree View with PPID support | ✅ |
| T-49 | Arbitrary Log Explorer with Tail/Search | ✅ |
| T-50 | New UI tests for AIOps, SecOps, DevOps | ✅ |

## Git State (Current)
- **Branch**: `main` ✅ pushed to origin
- **HEAD**: `237ec56` — `chore: untrack node_modules from git`
- **Tags**: `v1.3.0` (pushed), `v1.2.0`, `v1.1.1`, `v1.1.0`, `v1.0.0`
- **Push commits**: `e2b6a0b` (CI hardening) + `237ec56` (untrack node_modules)

## Fixes Applied This Session

| # | Issue | File | Fix |
|---|-------|------|-----|
| 1 | `healthColor()` inverted — low CPU showed red, high CPU showed green | `Dashboard.tsx` | Reversed: <80% → success(green), >=80% → warning(amber), >=90% → danger(red) |
| 2 | Settings DEFAULT_APP_INFO version hardcoded as `1.1.1` | `Settings.tsx` | Updated to `1.3.0` |
| 3 | Default AI model was `llama3.2` (not found locally) | `ollama.go` | Changed to `agentic-coder` (Qwopus3.5-9B-Coder based) |
| 4 | CLAUDE.md referenced old `llama3.2` model | `CLAUDE.md` | Updated known issue #2 |
| 5 | eslint-plugin-react-hooks@5.2.0 had peer dep conflict with eslint@10 | lockfile | Synced lockfile to 7.1.1 (supports eslint ^10) |
| 6 | node_modules (9,147 files) tracked in git, causing bloat + CI CRLF issues | `.gitignore`, `git rm --cached` | Untracked node_modules — CI uses `npm ci` |
| 7 | CI test.yml missing setcap for ping binary | `.github/workflows/test.yml` | Added `sudo setcap cap_net_raw+ep /bin/ping` |
| 8 | release.yml GITHUB_TOKEN fails on private repo | `.github/workflows/release.yml` | Changed to `GH_TOKEN || GITHUB_TOKEN` |
| 9 | index.html loaded Inter/JetBrains Mono but CSS uses Geist | `index.html` | Switched to Geist + Geist Mono (matches globals.css) |

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
12. **Semantic diff for Docs & Plans**: Archived `plans/2025-05-22-sprint-24-quality-polish.md` — 6 stale TODOs closed. Archived `docs/ToolsCommands.md` raw research notes — formatted into proper reference doc.
13. **Doc fixes applied 2026-07-12**: Updated ARCHITECTURE.md/STANDARDS.md dates, fixed USER_GUIDE.md (agentic-coder model, removed onboarding wizard), archived stale plan, cleaned ToolsCommands.md, updated state.md, fixed main.go (Hawkward→OpsForAll, path), removed 25 unused TS imports, added missing diagColor/diagIcon functions in Dashboard.tsx, fixed unicode corruption in Dashboard.tsx, fixed 3 missing type imports.
14. **Full-layer test sprint 2026-07-12 — 6 parallel agents**:
   - Backend: 140/140 Go tests pass, vet clean ✅
   - TypeScript: tsc clean, build clean (3217 modules) ✅
   - Vitest: 60/60 pass (was 59/60, fixed Dashboard test) ✅
   - Lint: 0 errors, 0 warnings, 45 files ✅
   - Dead code audit: 4 TS items + entire `theme.go` (415 lines) + 6 orphan Go funcs flagged
   - E2E audit: Fixed wails.json (stale frontend:dir path → blocked all Wails builds), fixed test.yml/release.yml CI paths. Wails build now passes. 8k+ hawkward refs remain as known debt.
15. **Renaming debt (Hawkward→OpsForAll)**: After this sprint, paths are fixed but internal branding is ~50% migrated. Known remaining: binary names in release.yml, App.go app name, DB/log filenames, Prometheus metric names, AI prompt context, config dir, CLAUDE.md, README.md.
2. **TestPing**: May fail on Linux CI (no root for raw sockets + ping binary needs CAP_NET_RAW) — improved skip logic.
3. **TestInsertLogAndQuery**: Flaky in full suite — SQLite DB leaks between tests.
4. **Recharts ResponsiveContainer**: Stderr warnings in jsdom tests — cosmetic (jsdom has no layout).
5. **Frontend test coverage**: Missing for TopBar, NetOps, SecOps, DevOps, AIOps, NetworkDesign — P3.
6. **gopacket**: P4 — needs Npcap runtime on Windows.
7. **"Cannot find module" LSP diagnostics for wailsjs**: Stale — packages exist and build succeeds.
8. **NetworkDesign**: Hardcoded seed devices on first visit — BY DESIGN (example topology).
9. **npm ci on Windows**: EPERM errors on `lightningcss.win32-x64-msvc.node` due to file locks — CI uses Linux, unaffected.
10. **node_modules previously tracked in git** ✅ RESOLVED — untracked in `237ec56`. CI will use `npm ci` as intended.
11. **CI setcap for TestPing** ✅ RESOLVED — `setcap cap_net_raw+ep /bin/ping` in test.yml

## Comprehensive Architecture Analysis (2026-07-14)
- [[comprehensive-analysis-2026-07-14]] — Full three-round analysis: code recon, docs cross-ref, 2026 ops standards. Collector architecture exemplary. Alerting needs notification routing. AI needs RAG/tool calling. Top 3 investments identified.

## Staff Engineering Code Review Audit (2026-07-14)
- [[audit-report-2026-07-14]] — 4 parallel agents, 61 findings (12 CRITICAL, 20 HIGH, 14 MEDIUM, 15 LOW). 8 cross-cutting themes: command injection in SecOps incident response, data races in shared state, systematic error swallowing, Linux sudo in GUI, sandbox bypass, business logic in bindings, security stubs, frontend systemic issues. Verification pass required before fixes.

## Settings Audit Results (Sprint 23)

| Page | Settings Used | Status |
|------|--------------|--------|
| SysOps.tsx | `refreshInterval` (5 queries: CPU, Mem, Disk, Sys, Procs) | ✅ Wired |
| NetOps.tsx | `pingCount`, `refreshInterval`, `dnsTimeout` (Ping, Connections, Interfaces) | ✅ Wired |
| AIOps.tsx | `refreshInterval` (Ollama status) | ✅ Wired |
| Logs.tsx | `refreshInterval` (log polling) | ✅ Wired |
| Settings.tsx | Control panel: theme toggle, interval select, ping slider, DNS slider, reset, about | ✅ Control panel |
| Dashboard.tsx | Event-driven via `useEvents('metrics')` + `staleTime: 10000` | ✅ No polling needed |
| SecOps.tsx | No polling (static/on-demand data — firewall, users, defender, events) | ✅ Acceptable |
| DevOps.tsx | No polling (user-driven — terminal, file browser, services) | ✅ Acceptable |

### Settings Persistence Chain
```
Settings.tsx → useSettingsStore (zustand) → localStorage → NetOps/SysOps/AIOps/Logs
                                                → PipelineAPI.UpdateSettings (Go backend tick loop)
```
All pages that poll data read `refreshInterval` from the same store. Changes take effect on next query refetch.
