# Hawkward — Architecture

## Go Module
- **Module name**: `hawkward`
- **Import prefix**: `hawkward/...`
- **Entry point**: `cmd/hawkward-gui/main.go` (Wails app)
- **Go version**: 1.26.4

## Package Layout

```
cmd/hawkward/main.go          — Entry point
internal/
  common/                     — Shared types, styles, formatters, platform, sandbox
    formatters.go             — Pretty-printing, colorized output
    platform.go               — OS detection
    sandbox.go                — Sandboxed command interface (unified)
    sandbox_windows.go        — Windows: NoInheritHandles + HideWindow
    sandbox_linux.go          — Linux: CLONE_NEWNET/NS/PID/USER namespaces
    styles.go                 — Lipgloss style definitions
    types.go                  — Shared structs
  sysops/                     — System operations
    collector.go, cpu.go, disk.go, memory.go, processes.go, system.go
    model.go, update.go, view.go, workflows.go
  netops/                     — Network operations
    connections.go, dns.go, interfaces.go, ping.go, portscan.go
    model.go, update.go, view.go, workflows.go
  secops/                     — Security operations
    defender.go, firewall.go, listening.go, tasks.go, users.go
    model.go, update.go, view.go, workflows.go
  devops/                     — Development operations
    filebrowser.go, logtail.go, shell.go
    model.go, update.go, view.go, workflows.go
  aiops/                      — AI operations (Ollama)
    ollama.go, reporting.go
    model.go, update.go, view.go, workflows.go
```

## Key Design Decisions
- No external CLI dependencies — all ops use Go native libs + `exec.Command`
- No Docker, no web UI, no data egress — fully local terminal app
- Reports: each ops layer has `Run*()` function + report struct with `String()` and `Markdown()` methods
- GUI: Wails v2 (Go + React/TypeScript frontend with Tailwind v4 + Recharts)
