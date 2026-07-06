# Project Instructions: Hawkward TUI

## Tech Stack
- **Language**: Go 1.26.4
- **TUI Framework**: Bubble Tea v2 (`charm.land/bubbletea/v2`)
- **Styling**: Lip Gloss v2 (`charm.land/lipgloss/v2`)
- **Ops Libraries**: `gopsutil/v4`, `miekg/dns`, `golang.org/x/net`

## Build & Run
- **Build**: `go build -o hawkward.exe ./cmd/hawkward`
- **Run**: `./hawkward.exe` (or `go run ./cmd/hawkward`)
- **Test**: `go test ./...`
- **Release**: `./scripts/release.sh` or `scripts/build.bat`

## Code Style
- **Pattern**: Follow The Elm Architecture (TEA).
- **Update Loop**: Delegate keyboard events to a private `handleKeyPress` helper in `update.go`.
- **Screen Routing**: Always use the `Screen` enum in `internal/common/types.go`.
- **Styling**: Use `internal/common/` styles; avoid hardcoded colors.
- **Naming**: `PascalCase` for exported Go symbols, `camelCase` for internal. Kebab-case for file names (except `*_test.go`).
- **Logging**: Use `common.LogInfo`, `common.LogWarn`, and `common.LogError`.
- **Async**: Wrap long-running tasks in `tea.Cmd` and use `ResultMsg` structs for completions.

## Testing
- Ensure unit tests exist for all new features in `*_test.go`.
- Prefer table-driven tests (see `internal/sysops/sysops_test.go` for examples).

## Project Structure
- `cmd/hawkward/`: Entry point.
- `internal/ui/`: Root TUI components and routing.
- `internal/common/`: Shared utilities, types, and logging.
- `internal/{sysops,netops,secops,devops,aiops}/`: Modular domain logic.
- `docs/`: Project documentation.
- `scripts/`: Platform-specific build and release scripts.
