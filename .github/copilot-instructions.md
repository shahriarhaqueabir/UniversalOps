# Copilot Instructions — OpsForAll

## Project
OpsForAll is a native desktop operations platform (Go/Wails v2 + React 19/TypeScript). 100% local, zero telemetry.

## Stack
- **Backend**: Go 1.26+, Wails v2, gopsutil, miekg/dns, modernc.org/sqlite
- **Frontend**: React 19, TypeScript, Vite, Tailwind v4, Zustand, TanStack Query, Recharts, Radix UI

## Commands
- `wails dev` — hot-reload dev server
- `wails build` — production build
- `go test ./internal/...` — backend tests
- `npm test --prefix cmd/opsforall-gui/frontend` — frontend tests
- `golangci-lint run ./...` — backend lint

## Structure
- `main.go` — entry point
- `internal/app/` — Wails bound facades
- `internal/common/` — core services (Pipeline, Storage, Alerts, Sandbox)
- `internal/{sysops,netops,secops,devops,aiops}/` — subsystems
- `cmd/opsforall-gui/frontend/` — React frontend

## Rules
- **Go**: Idiomatic Go, error wrapping, log via `common.LogInfo`
- **Frontend**: Functional components, memoize expensive renders, CSS variables for theming
- **Database**: All persistence through `ops_core.db` (SQLite)
- **No cloud deps**: Never add external API calls or telemetry
- **Tests required**: New features must include tests. Run full suite before claiming done.
