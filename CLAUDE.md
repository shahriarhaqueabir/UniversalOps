# Project Instructions: Hawkward GUI

## Tech Stack
- **Backend**: Go 1.26.4
- **Frontend**: React + TypeScript + Vite (Tailwind v4)
- **GUI Framework**: Wails v2 (`github.com/wailsapp/wails/v2`)
- **Ops Libraries**: `gopsutil/v4`, `miekg/dns`, `golang.org/x/net`

## Build & Run
- **Build**: `wails build`
- **Dev**: `wails dev`
- **Frontend Dev**: `cd cmd/hawkward-gui/frontend && npm run dev`
- **Test (Go)**: `go test ./...`
- **Test (Frontend)**: `cd cmd/hawkward-gui/frontend && npm test`

## Code Style
- **Pattern**: Wails Bindings (Go) + React Hooks (Frontend).
- **Subsystems**: Bound in `main.go` from `internal/app/`.
- **Styling**: Squib-inspired design system in `globals.css`.
- **Icons**: Lucide React.
- **Charts**: Recharts.
- **Naming**: `PascalCase` for exported Go symbols, `camelCase` for internal. `camelCase` for TS/JS.
- **Logging**: Use `common.LogInfo` in Go; `console.log` or a dedicated hook in Frontend.

## Testing
- **Go**: Unit tests in `*_test.go` using standard `testing` package.
- **Frontend**: Vitest + React Testing Library (RTL).

## Project Structure
- `cmd/hawkward/`: Entry point.
- `internal/ui/`: Root TUI components and routing.
- `internal/common/`: Shared utilities, types, and logging.
- `internal/{sysops,netops,secops,devops,aiops}/`: Modular domain logic.
- `docs/`: Project documentation.
- `scripts/`: Platform-specific build and release scripts.
