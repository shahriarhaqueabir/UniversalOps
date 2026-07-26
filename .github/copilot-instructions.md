# Copilot Instructions — Universal-Ops

## Project
Universal-Ops is a native desktop operations platform (Go/Wails v2 + React 19/TypeScript). 100% local, zero telemetry.

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
- **Database**: All persistence through `UniversalOps.db` (SQLite)
- **No cloud deps**: Never add external API calls or telemetry
- **Tests required**: New features must include tests. Run full suite before claiming done.

## 🔄 Agent Lifecycle — Auto-Trigger (KICKOFF Phase)
Run these steps at the START of every session, before ANY other work:

### Phase 1: KICKOFF
1. Read `~/.agents/memory/state.md` — load session continuity
2. Read `~/.agents/kanban/BOARD.md` — load active tasks
3. Read `.memory/state.md` and `.memory/index.md` — load project context
4. State goal and success criteria
5. Announce: "Will recon → review → graphify → compact → memory"

### Phase 2: RECON
Before editing any file, explore the codebase first.
- Read relevant files before modifying
- Check repo memory at `/memories/repo/` for prior findings
- Search for existing patterns (`grep_search`) before writing new code

### Phase 3: REVIEW
Before claiming completion of any task:
- Run `go build ./...` and `go vet ./...`
- Run `go test ./internal/... -count=1 -timeout 120s`
- Run frontend tests
- Show actual output, not just assertions

### Phase 4: GRAPHIFY
- Connect changes to existing patterns and prior decisions
- Check `.memory/topics/` and `/memories/repo/` for related prior work
- Document dependencies

### Phase 5: COMPACT
- Write current state to `.memory/state.md`
- Persist decisions to `.memory/topics/<topic>.md`
- Signal next step clearly

### Phase 6: MEMORY
- Record After-Action Review to `~/.agents/homunculus/aar/YYYY/MM/DD/`
- Write observations to `~/.agents/homunculus/observations/YYYY-MM-DD.jsonl`
- Update kanban board
- Promote repeated patterns to instincts
