# OpsForAll Operations Platform

[![Version](https://img.shields.io/badge/version-v1.3.0-7c6cff)](https://github.com/shahriarhaqueabir/AllOpsFull/releases)
[![Go](https://img.shields.io/badge/Go-1.26.5-00ADD8)](https://go.dev/)
[![React](https://img.shields.io/badge/React-19-61DAFB)](https://react.dev/)
[![License](https://img.shields.io/badge/license-MIT-2ea44f)](LICENSE)
[![Platform](https://img.shields.io/badge/platform-Windows%20|%20macOS%20|%20Linux-6366f1)](#provisioning)

A high-density operational intelligence suite for systems, network, and security auditing. **100% Private. 100% Local. Zero Cloud.**

> ⚠️ **Release Access**: Assets are built via CI. Access the latest binaries on the [Releases](https://github.com/shahriarhaqueabir/AllOpsFull/releases) page.

```
  ── PHILOSOPHY ────────────────────────────────────────────────────────────
  OpsForAll is a private operations studio built for technical depth. 
  It provides a dense ecosystem of system, security, and network tools 
  that run entirely on your local hardware.

  • 100% LOCAL: Your data never leaves your hardware.
  • TECHNICAL: Deep access to kernel-level metrics.
  • PRIVATE: Zero telemetry. Zero cloud. Zero accounts.
  ──────────────────────────────────────────────────────────────────────────
```

## Why OpsForAll?

OpsForAll replaces fragmented CLI tools with a unified, high-performance native dashboard. It's designed for SREs, SysAdmins, and security enthusiasts who require **instant telemetry** without the latency or privacy risks of cloud-based monitoring.

---

---

## Technical Substrates

| Dashboard | SysOps | NetOps |
|---|---|---|
| Real-time health gauges, CPU timeline, alert indicators | Per-core CPU, memory, disk monitoring, process management | Ping, DNS, connections, interfaces, traceroute, bandwidth |

---

## Provisioning

### Option 1: Standalone EXE (Quickest)

1. Go to the **[Releases page](https://github.com/shahriarhaqueabir/AllOpsFull/releases)**
2. Under **"Assets"** for the latest release, click the file named:
   - **Windows**: `opsforall-v1.3.0-windows-amd64.exe` (~14 MB)
   - **macOS**: `opsforall-v1.3.0-darwin-universal`
   - **Linux**: `opsforall-v1.3.0-linux-amd64`
3. **Double-click** the downloaded file to launch

> ⚠️ **Windows SmartScreen**: The first time you run it, Windows may show "Windows protected your PC". Click **"More info"** → **"Run anyway"**. This happens because the binary is not yet code-signed.

### Option 2: NSIS Installer (Windows only)

If you prefer a proper installer with Start Menu shortcuts:

1. On the Releases page, download `opsforall-v1.3.0-windows-amd64-installer.exe`
2. Double-click the installer
3. Follow the setup wizard

---

## Quick Reference Manual

### 🖥️ SYSTEM SUBSTRATE
- **Core Health**: Kernel-level audit of CPU, RAM, and processes.
- **Compute Audit**: Distinction between Physical Cores and Logical Threads.
- **Load Saturation**: Real-time Saturation Index (Load Avg relative to core count).
- **Process Management**: Runtime audit with PID tracking and impact assessment.

### 🌐 NETWORK MATRIX
- **ICMP Triage**: Internet reachability, jitter calculation, and latency history.
- **DNS Audit**: Multi-resolver benchmark and cache poisoning verification.
- **Port Scan**: Concurrent TCP connection attempts (Optimized ~200ms).
- **Routing**: Sequential ICMP TTL-incrementing probes (Traceroute).
- **Bandwidth**: Real-time throughput monitoring with historical sparklines.

### 🛡️ SECURITY MATRIX
- **Firewall Intelligence**: Cross-reference active rules with open listeners.
- **Identity Audit**: Enumerate local accounts with elevated privileges.
- **Listener Audit**: Flag processes listening on external interfaces (0.0.0.0).
- **Endpoint Intel**: Real-time Windows Defender and security event monitoring.

### 🧠 LOCAL INTELLIGENCE
- **Ollama Pipeline**: Integrated local-only LLM processing.
- **Heuristic Synthesis**: Statistical analysis of log data into technical briefings.
- **Anomaly Triage**: Detection of metric deviations via rolling window history.
- **Secret Masking**: Automated redaction of sensitive tokens from all reports.

---

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
│  │  Ops Core Design · Recharts · Lucide       │  │
│  │  Tailwind v4 · Radix UI                    │  │
│  └───────────────────────────────────────────┘  │
└─────────────────────────────────────────────────┘
```

- **Backend**: Go (Wails v2 bindings with gopsutil/v4, miekg/dns, modernc.org/sqlite)
- **Frontend**: React (TypeScript, Vite, Tailwind v4, Recharts, Lucide React)
- **Database**: `ops_core.db` — SQLite with WAL mode and 7-day retention.

---

## Development

### Prerequisites

- Go 1.26+
- Node.js & npm
- Wails CLI (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`)

### Commands

```bash
# Start development mode
wails dev

# Build production binary
wails build

# Run tests
go test ./internal/...
npm test --prefix cmd/opsforall-gui/frontend
```

## License

MIT
