# Installation & Quick Start Manual

Welcome to **Universal-Ops**. This guide will get you from zero to a fully lit-up Operations Command Center in less than 5 minutes.

---

## 🖱️ Easy Setup (For Non-Technical Users)

### 1. Download the App
Go to the **[Releases page](https://github.com/shahriarhaqueabir/AllOpsFull/releases)** and download the latest version:

- **Windows portable (recommended):** `universal-ops-*-windows-amd64.exe`
- **Windows installer:** `universal-ops-*-windows-amd64-installer.exe` (if you prefer a proper install)

### 2. Run It
**Portable version:** Create a folder on your Desktop named `UniversalOps`, move the `.exe` there, then double-click it. No installation needed.

**Installer version:** Double-click the installer and follow the prompts.

### 3. (Optional) Install Extras
Universal-Ops works out of the box, but for extra features:

| Extra | What it adds | How to get it |
|-------|-------------|---------------|
| **Ollama** | AI analysis in the AI Ops tab | Download from [ollama.com](https://ollama.com) |
| **LibreHardwareMonitor** | CPU/GPU temperature & fan speed sensors | Download from [github.com/LibreHardwareMonitor](https://github.com/LibreHardwareMonitor/LibreHardwareMonitor) |

That's it. Everything else works automatically.

---

## 🛠️ Advanced Setup (From Source)

### Prerequisites
- **Go 1.26+**: [Download](https://go.dev/dl/)
- **Node.js 22+**: [Download](https://nodejs.org/)
- **Wails v2**: Install via `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
- **GCC (Windows)**: [MinGW](https://www.mingw-w64.org/) or TDM-GCC
- **Linux**: `sudo apt install gcc libgtk-3-dev libwebkit2gtk-4.1-dev`
- **macOS**: Xcode Command Line Tools

### Build
```powershell
git clone https://github.com/shahriarhaqueabir/AllOpsFull.git UniversalOps
cd UniversalOps
New-Item -ItemType Directory -Path data,logs,bin -Force
cd cmd/opsforall-gui/frontend
npm install
cd ../../../
go mod tidy
wails build
```

The binary will be at `build/bin/universal-ops.exe`.

---

## External Tool Detection
Universal-Ops works out-of-the-box for core functionality. Some features require external tools that the app detects automatically on launch.

| Tool | Required For | Status |
|------|-------------|--------|
| **Ollama** | AI Ops tab, Technical Briefings | Auto-detected |
| **LibreHardwareMonitor** | CPU/GPU temperatures, fan speeds | Auto-detected |
| **NVIDIA SMI** | GPU utilization monitoring | Auto-detected with NVIDIA drivers |
| **PowerShell** | System queries, Windows API fallback | Built into Windows |
| **Docker** | Container management | Auto-detected |
| **Git** | Repository operations | Auto-detected |

> Missing tools will not block startup. The app degrades gracefully and shows `N/A` for unavailable data.

---

## First Launch Checklist
When you first open Universal-Ops, follow the **Enhanced Onboarding Wizard**:

1.  **Welcome**: Review the local-first privacy policy.
2.  **System Check**: The app scans for tools and shows their detection status.
3.  **Baseline**: Let the app sit for 30 seconds to capture your system's "Normal" state.
4.  **Ready**: Enter the Control Center.

---

## Troubleshooting
If the dashboard shows `N/A` for sensors:
- Ensure LibreHardwareMonitor is running as **Administrator** with WMI Provider enabled.
- Check that Ollama is running in your system tray.
- See the [Troubleshooting Guide](./TROUBLESHOOTING.md) for more.
