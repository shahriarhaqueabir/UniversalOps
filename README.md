# Hawkward Operations Platform

A premium native desktop operations platform built with Go + Wails v2 + React.

## Download

**No programming skills required.**

1. Go to the [Releases page](https://github.com/shahriarhaqueabir/AllOpsFull/releases)
2. Download the latest version for your OS:
   - **Windows**: `hawkward-v1.0.0-windows-amd64.exe` or the `.exe.nsis` installer
   - **macOS**: `hawkward-v1.0.0-darwin-universal`
   - **Linux**: `hawkward-v1.0.0-linux-amd64`
3. Double-click to launch

---

## Features

- **Dashboard**: Real-time health gauges, KPI cards with sparklines, and uptime history.
- **SysOps**: Detailed CPU/Memory/Disk monitoring and process management.
- **NetOps**: Advanced network tools including continuous ping, DNS lookup, port scanning, and traceroute.
- **SecOps**: Firewall rule management, user auditing, and security event monitoring.
- **DevOps**: Integrated terminal, service management, and file explorer.
- **AIOps**: AI-powered operations assistant (Ollama integration), report generation, and anomaly detection.
- **Network Designer**: Interactive canvas for designing network topologies with persistence.
- **Log Viewer**: Virtual-scrolled log aggregator with advanced filtering.

## Architecture

- **Backend**: Go (Wails v2 bindings)
- **Frontend**: React (TypeScript, Vite, Tailwind v4)
- **Design System**: Squib-inspired dark theme with Inter & JetBrains Mono typography.

## Development

### Prerequisites

- Go 1.26+
- Node.js & npm
- Wails CLI (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`)

### Commands

```bash
# Start development mode (hot-reload)
wails dev

# Build production binary
wails build

# Run tests
go test ./...
cd cmd/hawkward-gui/frontend && npm test
```

## License

MIT
