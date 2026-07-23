# Installation & Quick Start Manual

Welcome to **Universal-Ops**. This guide will get you from zero to a fully lit-up Operations Command Center in less than 5 minutes.

## 1. Primary Download
Download the latest pre-compiled binary for your architecture from the **[Releases](https://github.com/shahriarhaqueabir/AllOpsFull/releases)** page.

- **Windows**: `universal-ops-windows-amd64.exe`
- **Portable Mode**: No installation required. Create a folder named `UniversalOps` on your Desktop, move the `.exe` there, and run it.

---

## 2. External Tool Detection
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

## 3. First Launch Checklist
When you first open Universal-Ops, follow the **Enhanced Onboarding Wizard**:

1.  **Welcome**: Review the local-first privacy policy.
2.  **System Check**: The app scans for tools and shows their detection status.
3.  **Baseline**: Let the app sit for 30 seconds to capture your system's "Normal" state.
4.  **Ready**: Enter the Control Center.

---

## 4. Troubleshooting
If the dashboard shows `N/A` for sensors:
- Ensure LibreHardwareMonitor is running as **Administrator** with WMI Provider enabled.
- Check that Ollama is running in your system tray.
- See the [Troubleshooting Guide](./TROUBLESHOOTING.md) for more.
