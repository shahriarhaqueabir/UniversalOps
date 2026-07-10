# Project Instructions: Hawkward GUI

## Tech Stack
- **Backend**: Go 1.26.5
- **Frontend**: React 19 + TypeScript + Vite 6 (Tailwind v4)
- **GUI Framework**: Wails v2 (`github.com/wailsapp/wails/v2`)
- **Ops Libraries**: `gopsutil/v4`, `miekg/dns`, `golang.org/x/net`, `modernc.org/sqlite`, `ollama/ollama/api`

## Build & Run
- **Build**: `wails build` (NOT `go build` — Wails embed + assets require `wails build`)
- **Dev**: `wails dev`
- **Frontend Dev**: `cd cmd/hawkward-gui/frontend && npm run dev`
- **Test (Go)**: `go test ./internal/... -count=1`
- **Test (Frontend)**: `cd cmd/hawkward-gui/frontend && npm test`
- **Lint (Frontend)**: `cd cmd/hawkward-gui/frontend && npm run lint`
- **TypeScript Check**: `cd cmd/hawkward-gui/frontend && npx tsc --noEmit`
- **Frontend Build**: `cd cmd/hawkward-gui/frontend && npm run build`
- **Go Vet**: `go vet ./...`

## Code Style
- **Pattern**: Wails Bindings (Go) + React Hooks (Frontend).
- **Subsystems**: Bound in `main.go` from `internal/app/`.
- **Styling**: Squib-inspired design system in `globals.css`. All colors via CSS variables — NO hardcoded colors (no `bg-white/5`, `hover:bg-white/5`, inline hex colors).
- **Icons**: Lucide React.
- **Charts**: Recharts.
- **CSS Variable for hover**: `hover:bg-[var(--color-sidebar-hover)]` (not `hover:bg-white/5`)
- **Naming**: `PascalCase` for exported Go symbols, `camelCase` for internal. `camelCase` for TS/JS.
- **Logging**: Use `common.LogInfo` in Go; `console.log` or a dedicated hook in Frontend.

## Testing
- **Go**: Unit tests in `*_test.go` using standard `testing` package.
- **Frontend**: Vitest + React Testing Library (RTL).
- **ESLint**: eslint 10 + typescript-eslint 8 + react-hooks v7 — `eslint.config.js` disables `react-hooks/set-state-in-effect` and `react-hooks/incompatible-library` for Wails patterns.

## Project Structure
- `main.go`: Entry point (Wails v2 embed + runtime).
- `cmd/hawkward-gui/frontend/`: React + Vite frontend.
- `internal/common/`: Shared utilities, types, logging, DataPipeline, Storage (SQLite), AlertEngine.
- `internal/{sysops,netops,secops,devops,aiops}/`: Modular domain logic.
- `internal/app/`: Wails bindings (AIOps.go, App.go, Dashboard.go, DevOps.go, NetOps.go, SecOps.go, SysOps.go, Types.go, Logs.go, Pipeline.go, Alerts.go, Events.go).
- `docs/`: Project documentation.
- `scripts/`: Platform-specific build and release scripts.
- `.github/workflows/`: CI/CD (test.yml, release.yml).

## Key Libraries (Frontend)
- **@tanstack/react-query v5** — data fetching
- **@tanstack/react-table v8** — table rendering
- **@tanstack/react-virtual v3** — virtual scrolling
- **zustand v5** — state management
- **@radix-ui/*** — accessible UI primitives (collapsible, dialog, dropdown, progress, scroll-area, select, separator, slider, switch, tabs, toggle, tooltip, avatar)
- **recharts v3** — charts
- **date-fns v4** — date formatting
- **class-variance-authority + clsx + tailwind-merge** — class management
- **sonner** — toast notifications
- **motion** — animations
- **lucide-react** — icons
- **Tailwind v4** — styling

## Key Libraries (Go Backend)
- **gopsutil/v4** — system metrics (CPU, memory, disk, processes)
- **miekg/dns** — DNS lookups
- **golang.org/x/net** — ICMP ping, network operations
- **modernc.org/sqlite** — embedded database
- **ollama/ollama/api** — AI chat
- **rs/zerolog** — structured logging
- **prometheus/client_golang** — Prometheus metrics exposition (port 9210, /metrics endpoint)

## Dashboard Actions
- **QUICK DIAGNOSTIC**: Calls `Dashboard.RunQuickDiag()` backend method — NOT just a navigation link
- **GENERATE BRIEFING**: Calls `Dashboard.GenerateDashboardBriefing()` backend method — NOT just a navigation link
- **Compute Logic Analysis**: Dynamic red flags computed from real pipeline data (not hardcoded)

## Known Issues
1. **App.go version** must match `wails.json` productVersion and `package.json` version (currently v1.3.0).
2. **AIOps model sync**: `ollama.go` defaults to `llama3.2` — set `OLLAMA_MODEL` env var to override. Falls back to first available model if default not found.
3. **TestPing** may fail on Linux CI (no root for raw sockets + ping binary CAP_NET_RAW) — test has lazy skip logic.
4. **TestInsertLogAndQuery** is flaky when run in full suite — SQLite DB leaks between tests.
5. **Recharts ResponsiveContainer** stderr in jsdom tests — cosmetic, jsdom has no layout.
6. **Prometheus** — Installed (`client_golang` v1.23.2), HTTP endpoint on :9210/metrics, wired into tick loop.
7. **gopacket** ⏸️ P4 — needs Npcap runtime on Windows.
8. Missing frontend test coverage: TopBar, NetOps, SecOps, DevOps, AIOps, NetworkDesign — P3.
9. **Private repo**: Release creation via `GITHUB_TOKEN` may fail — needs PAT in `GH_TOKEN` secret.

## CI Pipeline
- **test.yml**: Runs on push to main/develop and PRs to main. Steps: npm ci → npm run build → go vet → go test → tsc --noEmit → npm test → npm run lint.
- **release.yml**: Runs on `v*` tags. Builds Windows (exe + NSIS), Linux, macOS binaries and uploads as release assets.
- **ESLint**: Uses eslint 10, react-hooks v7 (not v5), jiti v2 as dev dep.
