# Hawkward Operations Platform

![Version](https://img.shields.io/badge/version-v1.3.0-7c6cff)
![Go](https://img.shields.io/badge/Go-1.26.5-00ADD8)
![License](https://img.shields.io/badge/license-MIT-2ea44f)
![Platform](https://img.shields.io/badge/platform-Windows%20|%20macOS%20|%20Linux-6366f1)

> ⚠️ **Private Repository**: Release assets are built via CI but the repository is private.
> Download the latest release from the [Releases](https://github.com/shahriarhaqueabir/AllOpsFull/releases) page (requires access).

A premium native desktop operations platform — **no programming skills required**.

```
  ╔══════════════════════════════════════════════╗
  ║  HAWKWARD — Operations Intelligence Suite    ║
  ║  System Monitoring · Network Tools · SecOps  ║
  ║  DevOps Pipeline · AI Assistant              ║
  ╚══════════════════════════════════════════════╝
```

---

## Screenshots

| Dashboard | SysOps | NetOps |
|---|---|---|
| Real-time health gauges, CPU timeline, alert indicators | Per-core CPU, memory, disk monitoring, process management | Ping, DNS, connections, interfaces, traceroute, bandwidth |

*(Screenshots coming soon — run the app to see the interface)*

---

## Download & Run (No Programming Skills)

### Option 1: Standalone EXE (Quickest)

1. Go to the **[Releases page](https://github.com/shahriarhaqueabir/AllOpsFull/releases)**
2. Under **"Assets"** for the latest release, click the file named:
   - **Windows**: `hawkward-v1.3.0-windows-amd64.exe` (~14 MB)
   - **macOS**: `hawkward-v1.3.0-darwin-universal`
   - **Linux**: `hawkward-v1.3.0-linux-amd64`
3. **Double-click** the downloaded file to launch

> ⚠️ **Windows SmartScreen**: The first time you run it, Windows may show "Windows protected your PC". Click **"More info"** → **"Run anyway"**. This happens because the binary is not yet code-signed (common for open-source tools).

### Option 2: NSIS Installer (Windows only)

If you prefer a proper installer with Start Menu shortcuts:

1. On the Releases page, download `hawkward-v1.3.0-windows-amd64-installer.exe`
2. Double-click the installer
3. Follow the setup wizard

### System Requirements

| Requirement | Minimum | Recommended |
|-------------|---------|-------------|
| OS | Windows 10, macOS 12+, Ubuntu 22.04+ | Windows 11, macOS 14+ |
| RAM | 256 MB | 1 GB |
| Disk | 50 MB free | 100 MB free |
| CPU | Any x86-64 | Any x86-64 |
| Admin Rights | Not required (some network features may be limited) | For full functionality |

### What Gets Created

- **Database**: `hawkward.db` — SQLite file created in the directory where you run the app
- **Log file**: `hawkward-gui.log` — created alongside the database for debugging
- **No registry changes** (standalone EXE) or minimal registry (NSIS installer)

### First Launch

1. The Dashboard appears with live CPU, Memory, and Disk gauges
2. Navigate tabs using the sidebar:
   - **SysOps** — system monitoring & process management
   - **NetOps** — ping, DNS, ports, traceroute
   - **SecOps** — firewall, users, defender, events
   - **DevOps** — terminal, services, file browser
   - **AIOps** — AI chat, reports, anomaly detection (requires Ollama)
3. Settings gear icon (bottom-left) lets you toggle theme and adjust refresh rates

---

## Features

- **Dashboard**: Real-time health gauges, KPI cards with sparklines, CPU timeline chart, and alert indicators.
- **SysOps**: Detailed CPU (per-core), Memory (used/available/swap), and Disk monitoring with process management (search, kill).
- **NetOps**: Continuous ICMP ping, DNS lookup (A/AAAA/MX/NS/TXT/CNAME), port scanning, connection matrix, interface monitoring, and traceroute.
- **SecOps**: Firewall rule management (enable/disable), user audit, listening ports, Windows Defender status, and security event monitoring.
- **DevOps**: Integrated terminal with dangerous-command protection, PowerShell workflow runner, Windows service management (start/stop), and file browser with preview.
- **AIOps**: AI-powered operations chat (Ollama), report generation (security, performance, compliance, topology), and anomaly detection with auto-scan.
- **Network Designer**: Interactive canvas for designing network topologies with drag-drop devices, connections, properties panel, save/load to file, and web export.
- **Log Viewer**: Virtual-scrolled log aggregator with search, level filtering, expandable detail, and auto-follow mode.

## Architecture

```
┌─────────────────────────────────────────────────┐
│              Wails v2 Desktop App                │
│  ┌───────────────────────────────────────────┐  │
│  │        Go Backend (Wails Bindings)         │  │
│  │  Dashboard · SysOps · NetOps · SecOps     │  │
│  │  DevOps · AIOps · Pipeline · Alerts       │  │
│  │                                           │  │
│  │  DataPipeline (3s tick) → SQLite (WAL)    │  │
│  │  AlertEngine → Events → Wails runtime     │  │
│  └───────────────────────────────────────────┘  │
│                        │                        │
│              Wails Bridge (IPC)                 │
│                        │                        │
│  ┌───────────────────────────────────────────┐  │
│  │     React + TypeScript + Vite (Frontend)   │  │
│  │  Squib Design System · Recharts · Lucide   │  │
│  │  Tailwind v4 · Radix UI                    │  │
│  └───────────────────────────────────────────┘  │
└─────────────────────────────────────────────────┘
```

- **Backend**: Go (Wails v2 bindings with gopsutil/v4, miekg/dns, modernc.org/sqlite)
- **Frontend**: React (TypeScript, Vite, Tailwind v4, Recharts, Lucide React)
- **Design System**: Squib-inspired dark theme with Inter & JetBrains Mono typography
- **Database**: SQLite with WAL mode, buffered writes, auto-prune (7-day retention)

## Development

### Prerequisites

- Go 1.26+
- Node.js & npm
- Wails CLI (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`)
- **IMPORTANT**: Always build with `wails build`, never `go build` (Wails embeds the frontend dist via `//go:embed`)

### Commands

```bash
# Start development mode (hot-reload frontend + backend)
wails dev

# Build production binary (NOT go build)
wails build

# Build with NSIS installer (Windows)
wails build -nsis

# Run tests
go test ./...
cd cmd/hawkward-gui/frontend && npm test
```

### Project Structure

```
AllOpsFull/
├── main.go                    # Wails entry point (//go:embed frontend)
├── internal/
│   ├── app/                   # Wails bindings (Dashboard, SysOps, NetOps, etc.)
│   ├── common/                # Shared: Pipeline, Storage, Alerts, Sandbox, Types
│   ├── sysops/                # CPU, Memory, Disk, Process monitoring
│   ├── netops/                # Ping, DNS, PortScan, Traceroute, Connections
│   ├── secops/                # Firewall, Defender, Users, Events
│   ├── devops/                # Shell, FileBrowser, Services
│   └── aiops/                 # Ollama client, Report generation
├── cmd/hawkward-gui/frontend/ # React + TypeScript + Vite frontend
├── docs/                      # Documentation & archived plans
├── scripts/                   # Build and release scripts
└── .github/workflows/         # CI/CD: test.yml, release.yml
```

## License

MIT
