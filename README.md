# Hawkward

> **Operations Platform for the Terminal** — SysOps, NetOps, SecOps, DevOps, AI Ops

Hawkward is a keyboard-navigable terminal user interface (TUI) that puts system, network, security, development, and AI operations at your fingertips. It provides a unified, interactive dashboard for complex administrative tasks without the need for manual command juggling or multiple terminal windows.

Built with [Go](https://go.dev/) + [Bubble Tea](https://github.com/charmbracelet/bubbletea).

---

## Quick Start

```bash
# Build from source
cd hawkward
go build -o hawkward.exe ./cmd/hawkward

# Run
./hawkward.exe
```

**Prerequisites:** [Go 1.26.4+](https://go.dev/dl/)

---

## Features

### SysOps (System Operations) — ✅ Active
- Live CPU, memory, and disk usage dashboards
- Real-time progress bars with color-coded health (green/yellow/red)
- Top processes table sorted by CPU usage
- System information (hostname, OS, kernel, uptime)
- Auto-refreshing every 3 seconds

### NetOps (Network Operations) — ✅ Active
- ICMP ping (with `ping.exe` fallback on Windows)
- DNS lookup (via `miekg/dns`)
- TCP port scanning (via `net.DialTimeout`)
- Connection table (TCP/UDP via `netstat -ano`)
- Network interface monitoring

### SecOps (Security Operations) — ✅ Active
- Firewall rule viewer (`netsh advfirewall`)
- User and group audits (`net user` / `net localgroup`)
- Listening ports with process attribution (`netstat -ano` + `tasklist`)
- Windows Defender status (`Get-MpComputerStatus`)
- Scheduled task review (`Get-ScheduledTask`)

### DevOps (Development Operations) — ✅ Active
- Shell command runner with output display
- Log tailer and pattern search
- File browser with directory listing and file reading
- Process manager with list, kill, and restart commands
- Service status dashboard

### AI Ops (AI Operations) — ✅ Active
- Local AI assistant (Ollama integration via `localhost:11434`)
- Report generation (text and markdown)
- Report export to file

---

## Keyboard Controls

| Key | Action |
|-----|--------|
| `↑/↓` or `k/j` | Navigate menus |
| `Enter` / `Space` | Select item |
| `Esc` | Go back |
| `1` | SysOps dashboard |
| `2` | NetOps dashboard |
| `3` | SecOps dashboard |
| `4` | DevOps dashboard |
| `5` | AI Ops dashboard |
| `r` | Refresh current view |
| `?` | Toggle help overlay |
| `Tab` / `Shift+Tab` | Next/previous tab |
| `q` / `Ctrl+C` | Quit |

---

## Architecture

```
hawkward/
├── cmd/hawkward/        # Entry point
│   └── main.go
├── internal/
│   ├── sysops/          # System Operations layer
│   │   ├── cpu.go       # CPU metrics (gopsutil)
│   │   ├── memory.go    # RAM metrics
│   │   ├── disk.go      # Disk usage
│   │   ├── system.go    # Host info
│   │   ├── processes.go # Process listing
│   │   ├── collector.go # Aggregated stats
│   │   ├── model.go     # SysOps model
│   │   ├── update.go    # Message handling
│   │   └── view.go      # Dashboard rendering
│   ├── netops/          # Network Operations layer
│   ├── secops/          # Security Operations layer
│   ├── devops/          # DevOps layer
│   ├── aiops/           # AI Ops layer
│   ├── ui/              # Shared UI components
│   │   ├── root.go      # Root model & routing
│   │   ├── mainmenu.go  # Home screen
│   │   ├── help.go      # Help overlay
│   │   ├── onboarding.go# First-run wizard
│   │   ├── statusbar.go # Status bar with health
│   │   ├── keys.go      # Key bindings
│   │   └── styles.go    # Lip Gloss styles
│   └── common/          # Shared utilities
│       ├── types.go     # Common types
│       ├── formatters.go# Format helpers
│       └── platform.go  # OS detection
├── docs/
│   ├── ARCHITECTURE.md  # Architecture document
│   ├── STANDARDS.md     # Development standards
│   ├── ONBOARDING.md    # Onboarding design
│   └── ROADMAP.md       # Future plans
├── go.mod
└── README.md
```

---

## Development

```bash
# Clone and enter directory
cd hawkward

# Install dependencies
go mod tidy

# Build
go build -o hawkward.exe ./cmd/hawkward

# Run
./hawkward.exe

# Lint
golangci-lint run ./...

# Test
go test ./...

# Build local release artifacts
./scripts/release.ps1 -Version dev
```

### Standards

See [docs/STANDARDS.md](docs/STANDARDS.md) for commit conventions, code style, and review checklist.

## Documentation

- **[User Guide](docs/USER_GUIDE.md)**: Comprehensive guide on setup, AI configuration, and feature details.
- **[Architecture](docs/ARCHITECTURE.md)**: Deep dive into the internal design and technologies.
- **[Standards](docs/STANDARDS.md)**: Development guidelines and code style.

---

## License

MIT
