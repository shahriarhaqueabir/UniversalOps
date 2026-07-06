# Hawkward TUI

Operations Platform for the Terminal — SysOps, NetOps, SecOps, DevOps, AI Ops.

## Build and Run

```bash
# Build
go build -o hawkward.exe ./cmd/hawkward

# Run
./hawkward.exe

# Test
go test ./...
```

## Project Structure

- `cmd/hawkward/`: Entry point.
- `internal/`: Core logic layers (sysops, netops, secops, devops, aiops).
- `internal/ui/`: Root TUI components and routing.
- `internal/common/`: Shared utilities, themes, and styles.

## Coding Standards

- Refer to [@.ai-style-rules.md](.ai-style-rules.md) for detailed coding conventions.
- Follow Bubble Tea's The Elm Architecture (TEA) for TUI components.
- Use Lip Gloss for all styling; avoid hardcoded ANSI escapes.
- Ensure all new features are covered by unit tests in `*_test.go` files.
- Log session events using the `common` logger.
