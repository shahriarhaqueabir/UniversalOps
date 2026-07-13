# OpsForAll — Architecture

## Go Module
- **Module name**: `github.com/shahriarhaqueabir/AllOpsFull`
- **Import prefix**: `github.com/shahriarhaqueabir/AllOpsFull/...`
- **Entry point**: `cmd/opsforall-gui/main.go` (Wails app)
- **Go version**: 1.26.5

## Package Layout

```
cmd/opsforall-gui/main.go          — Entry point (Wails app)
internal/
  common/                         — Shared types, formatters, platform, sandbox
    formatters.go                 — Pretty-printing, colorized output
    platform.go                   — OS detection
    sandbox.go                    — Sandboxed command interface (unified)
    sandbox_windows.go            — Windows: NoInheritHandles + HideWindow
    sandbox_linux.go              — Linux: CLONE_NEWNET/NS/PID/USER namespaces
    types.go                      — Shared structs
  sysops/                         — System operations
    collector.go, cpu.go, disk.go, memory.go, processes.go, system.go
  netops/                         — Network operations
    connections.go, dns.go, interfaces.go, ping.go, portscan.go
  secops/                         — Security operations
    defender.go, firewall.go, listening.go, tasks.go, users.go
  devops/                         — Development operations
    filebrowser.go, logtail.go, shell.go
  aiops/                          — AI operations (Ollama)
    ollama.go, reporting.go
  app/                            — Wails app bindings
    App.go, Dashboard.go, ... 
```

## Key Design Decisions
- No external CLI dependencies — all ops use Go native libs + `exec.Command`
- No Docker, no web UI, no data egress — fully local desktop app
- Reports: each ops layer has `Run*()` function + report struct with `String()` and `Markdown()` methods
- GUI: Wails v2 (Go + React/TypeScript frontend with Tailwind v4 + Recharts)
- TUI-era code (Bubble Tea, Lip Gloss) has been fully removed — this is a GUI-only application
