# Installation & Quick Start Manual

Welcome to **Universal-Ops**. This guide will get you from zero to a fully lit-up Operations Command Center in less than 5 minutes.

## 1. Primary Download
Download the latest pre-compiled binary for your architecture from the **[Releases](https://github.com/shahriarhaqueabir/AllOpsFull/releases)** page.

- **Windows**: `universal-ops-windows-amd64.exe`
- **Portable Mode**: No installation required. Create a folder named `UniversalOps` on your Desktop, move the `.exe` there, and run it.

---

## 2. Power-Up: The Dependency Matrix
Universal-Ops is a "Native-First" application. It works out-of-the-box for 85% of tasks, but needs these "External Power-ups" for full data density.

### A. Local AI Brain (Ollama)
Required for the **AI Ops** tab and **Technical Briefings**.
1. **Download**: [ollama.com](https://ollama.com)
2. **Install**: Run the installer and ensure the Ollama icon is visible in your system tray.
3. **Connect**: Universal-Ops will automatically detect it on launch.

### B. Hardware Sensors (LibreHardwareMonitor)
Required for **CPU/GPU Temperatures** and **Fan Speeds**.
1. **Download**: [Latest Zip from GitHub](https://github.com/LibreHardwareMonitor/LibreHardwareMonitor/releases)
2. **Run**: Extract the ZIP and run `LibreHardwareMonitor.exe` as **Administrator**.
3. **Enable WMI**: Inside LHM, go to `Options` -> `Remote Control` -> `Enable WMI Provider`.

### C. GPU Metrics (NVIDIA)
Required for real-time **GPU Utilization**.
- Simply ensure your official **NVIDIA Drivers** are installed. The app uses the built-in `nvidia-smi` utility.

---

## 3. First Launch Checklist
When you first open Universal-Ops, follow the **Enhanced Onboarding Wizard**:

1.  **Welcome**: Review the local-first privacy policy.
2.  **System Check**: The app will scan for the power-ups above. Click **"Verify Now"** for any missing items.
3.  **Baseline**: Let the app sit for 30 seconds to capture your system's "Normal" state.
4.  **Ready**: Enter the Control Center.

## 4. Troubleshooting
If the dashboard shows `N/A` for sensors:
- Ensure LibreHardwareMonitor is running as **Administrator**.
- Check that Ollama is not blocked by your firewall.
- See the [Troubleshooting Guide](./TROUBLESHOOTING.md) for more.
