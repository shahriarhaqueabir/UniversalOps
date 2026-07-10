# Hawkward — Workspace Memory

## Active Session
- **Sprint 24: Quality Polish & Documentation Sync** 🔄 IN PROGRESS
- **Sprint 23: Comprehensive Codebase Audit, Settings Verification, CI Pipeline Hardening** ✅ DONE

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
| T-26 | Default AI model set to `agentic-coder` (was `llama3.2`) | ✅ |
| T-27 | `healthColor()` inverted logic fixed in Dashboard | ✅ |
| T-28 | Settings page version corrected (1.1.1 → 1.3.0) | ✅ |
| T-29 | eslint-plugin-react-hooks lockfile synced (5.2.0 → 7.1.1) | ✅ |
| T-30 | CLAUDE.md updated with agentic-coder model reference | ✅ |
| T-31 | Library audit: all recommended libs already present | ✅ |
| T-32 | node_modules untracked from git (removed 9,147 tracked files) | ✅ |
| T-33 | CI test.yml: setcap CAP_NET_RAW added for TestPing | ✅ |
| T-34 | release.yml: GITHUB_TOKEN → GH_TOKEN fallback for private repo | ✅ |
| T-35 | index.html font load: Inter/JetBrains Mono → Geist (consistent with globals.css) | ✅ |
| T-36 | Settings audit: all 7 pages verified for settings wiring | ✅ |
| T-37 | All Go tests (7/7) and Frontend tests (27/27) passing | ✅ |

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
