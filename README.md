# UniversalOps: High-Performance Native Operations Dashboard

<p align="center">
  <a href="https://github.com/shahriarhaqueabir/UniversalOps/releases">
    <img src="https://img.shields.io/badge/version-v1.3.0-7c6cff?style=flat-square" alt="Version">
  </a>
  <a href="LICENSE">
    <img src="https://img.shields.io/badge/license-MIT-2ea44f?style=flat-square" alt="License">
  </a>
  <a href="https://github.com/shahriarhaqueabir/UniversalOps/actions">
    <img src="https://img.shields.io/github/actions/workflow/status/shahriarhaqueabir/UniversalOps/.github/workflows/test.yml?branch=main&style=flat-square&label=CI" alt="CI">
  </a>
  <a href="https://github.com/shahriarhaqueabir/UniversalOps/blob/main/go.mod">
    <img src="https://img.shields.io/github/go-mod/go-version/shahriarhaqueabir/UniversalOps?style=flat-square&label=Go" alt="Go">
  </a>
  <a href="docs/ARCHITECTURE.md">
    <img src="https://img.shields.io/badge/Privacy-100%25%20Local-0066ff?style=flat-square" alt="Local-First">
  </a>
  <a href="docs/INSTALL.md">
    <img src="https://img.shields.io/badge/AI-Ollama%20Integrated-orange?style=flat-square" alt="AI-Powered">
  </a>
  <a href="https://github.com/shahriarhaqueabir/UniversalOps/issues">
    <img src="https://img.shields.io/badge/PRs-Welcome-brightgreen?style=flat-square" alt="PRs Welcome">
  </a>
  <a href="https://github.com/shahriarhaqueabir/UniversalOps/releases">
    <img src="https://img.shields.io/github/downloads/shahriarhaqueabir/UniversalOps/total?style=flat-square&label=Downloads" alt="Downloads">
  </a>
</p>

**UniversalOps** is a native desktop studio designed for SREs, developers, and security enthusiasts who require instant, high-density telemetry without the latency or privacy risks of cloud-based monitoring.

> [!IMPORTANT]
> **100% Private. 100% Local. Zero Cloud.** Your data never leaves your hardware.

---

## 🚀 One-Line Installation (Windows)
Run this command in PowerShell to instantly set up UniversalOps on your desktop:
```powershell
irm https://raw.githubusercontent.com/shahriarhaqueabir/UniversalOps/main/install.ps1 | iex
```

---

## 🎯 Mission & Vision

### **Our Mission**
To empower technical professionals with a **sovereign monitoring environment** that prioritizes speed, privacy, and technical depth. We believe that critical infrastructure data belongs on the machine that generates it—not in a third-party cloud.

### **Our Vision**
To become the definitive open-source Command Center for SREs and security researchers, bridging the gap between raw kernel metrics and actionable AI-driven intelligence.

### **Core Goals**
- **Zero Latency**: Sub-second telemetry updates without network roundtrips.
- **Privacy by Default**: 100% local data storage and AI analysis.
- **Actionable Intelligence**: Shift from "watching charts" to "solving problems" using Hawk, our local AI analyst.

---

## 🖥️ The Solution
UniversalOps replaces fragmented CLI tools and bloated cloud dashboards with a unified, high-performance substrate. It provides a single pane of glass for:

| Layer | Focus | The Problem We Solve |
| :--- | :--- | :--- |
| **SysOps** | Infrastructure | Fragmented CPU/Mem/Disk tools with poor correlation. |
| **NetOps** | Connectivity | Delayed jitter detection and lack of traffic attribution. |
| **SecOps** | Hardening | "Black box" endpoint security and opaque privilege tracking. |
| **DevOps** | Automation | Context-less shell execution and manual RCA. |
| **AIOps** | Intelligence | Cloud-based AI hallucinations and privacy data leakage. |

---

## 🖼️ Visual Gallery

### **Dashboard**
Real-time high-density telemetry including CPU thermal envelopes and memory pressure.
![Dashboard](./docs/screenshots/final_execution_1_dashboard.png)

### **Reports Center**
Structured insights and historical auditing with local SQLite persistence.
![Reports](./docs/screenshots/final_execution_2_reports.png)

### **System Operations**
Deep dive into per-core metrics, logical volumes, and thermal sensors.
![SysOps](./docs/screenshots/final_execution_3_sysops.png)

### **Network Operations**
Real-time ICMP Jitter, DNS Audit, and Port Scanning.
![NetOps](./docs/screenshots/final_execution_4_netops.png)

### **Security Operations**
Identity auditing, elevated privilege tracking, and firewall intelligence.
![SecOps](./docs/screenshots/final_execution_5_secops.png)

### **DevOps Automation**
Service orchestration and sandboxed terminal diagnostics.
![DevOps](./docs/screenshots/final_execution_6_devops.png)

### **Local Intelligence (Hawk)**
Autonomous Root Cause Analysis (RCA) via Ollama integration.
![AIOps](./docs/screenshots/final_execution_7_aiops.png)

### **Logs Viewer**
Structured log browsing with filtering, severity highlighting, and real-time tailing.
![Logs](./docs/screenshots/final_execution_8_logs.png)

### **Workflow Library**
Pre-built automation workflows with one-click execution and customizable parameters.
![Workflows](./docs/screenshots/final_execution_9_workflowlibrary.png)

---

## 🛠️ Architecture
Built for performance and portability.
- **Backend**: Go (Wails v2 bindings) using `gopsutil/v4`, `yusufpapurcu/wmi`, and `modernc.org/sqlite`.
- **Frontend**: React 19 + TypeScript + Tailwind v4 + Radix UI.
- **Storage**: SQLite with WAL (Write-Ahead Logging) for high-frequency time-series data.

**Full Architecture Deep-Dive**: [Documentation](./docs/ARCHITECTURE.md)

---

## 🤝 Contributing
We welcome contributions to enhance the Command Center.
- Review our [Contributing Guide](CONTRIBUTING.md).
- Read the [Development Guide](docs/developing.md) for setup, debugging, and the annotated project tree.
- Browse the [Documentation Hub](docs/readme.md) for all docs in one place.
- Join the discussion in [Issues](https://github.com/shahriarhaqueabir/UniversalOps/issues).
- This project adheres to the [Contributor Covenant Code of Conduct](CODE_OF_CONDUCT.md).

---

## 📄 License
Distributed under the MIT License. See `LICENSE` for more information.

---
*Developed for professionals who value speed, privacy, and technical depth.*
