# UniversalOps — Quick Start Guide

Get from zero to a fully operational Command Center in **under 2 minutes**.

---

## 🪟 Windows

### Option A: One-Line Install (Recommended)
Open **PowerShell** and run:

```powershell
irm https://raw.githubusercontent.com/shahriarhaqueabir/UniversalOps/main/install.ps1 | iex
```

This downloads the latest version, creates a `Desktop\UniversalOps` folder, and launches the app.

### Option B: Manual Download
1. Go to the [Releases page](https://github.com/shahriarhaqueabir/UniversalOps/releases)
2. Download `universal-ops-*-windows-amd64.exe`
3. Create a folder on your Desktop named `UniversalOps`
4. Move the `.exe` there and double-click it

---

## 🐧 Linux

```bash
curl -fsSL https://raw.githubusercontent.com/shahriarhaqueabir/UniversalOps/main/install.sh | bash
```

Or download `universal-ops-*-linux-amd64` from [Releases](https://github.com/shahriarhaqueabir/UniversalOps/releases), make it executable (`chmod +x`), and run it.

---

## 🍎 macOS

```bash
curl -fsSL https://raw.githubusercontent.com/shahriarhaqueabir/UniversalOps/main/install.sh | bash
```

Or download `universal-ops-*-darwin-universal` from [Releases](https://github.com/shahriarhaqueabir/UniversalOps/releases).

---

## 🎯 First Launch

1. **Welcome Screen** — Review the local-first privacy policy
2. **System Check** — The app scans for optional tools (Ollama, LibreHardwareMonitor, etc.)
3. **Baseline Capture** — Let the app run for 30 seconds to establish your system's normal state
4. **Ready** — You're in the Command Center

### What to explore first:

| Tab | What to look for |
|-----|-----------------|
| **Dashboard** | Real-time health score, resource summary, active alerts |
| **SysOps** | CPU cores, memory pressure, disk usage, top processes |
| **NetOps** | Ping latency, DNS resolution, port scanning |
| **SecOps** | Firewall rules, listening services, user accounts |
| **DevOps** | Sandboxed terminal, file explorer, service control |
| **AIOps** | AI analyst (requires [Ollama](https://ollama.com)) |

---

## 🔧 Optional Extras

| Tool | What it adds | Install |
|------|-------------|---------|
| **Ollama** | AI-powered analysis in AIOps tab | [ollama.com](https://ollama.com) |
| **LibreHardwareMonitor** | CPU/GPU temperature sensors | [LHM GitHub](https://github.com/LibreHardwareMonitor/LibreHardwareMonitor) |

The app works without these — missing features show `N/A` gracefully.

---

## 📚 Next Steps

- [Full Installation Guide](docs/INSTALL.md) — advanced setup, build from source
- [User Guide](docs/USER_GUIDE.md) — deep dive into every feature
- [Troubleshooting](docs/TROUBLESHOOTING.md) — fix common issues
- [Architecture Overview](docs/ARCHITECTURE.md) — how it all works