# Universal-Ops: High-Performance Native Operations Dashboard for Windows

[![Version](https://img.shields.io/badge/version-v1.3.1-7c6cff)](https://github.com/shahriarhaqueabir/AllOpsFull/releases)
[![License](https://img.shields.io/badge/license-MIT-2ea44f)](LICENSE)
[![Local-First](https://img.shields.io/badge/Privacy-100%25%20Local-blue)](docs/ARCHITECTURE.md)
[![AI-Powered](https://img.shields.io/badge/AI-Ollama%20Integrated-orange)](docs/INSTALL.md)

**Universal-Ops** is a native desktop studio designed for SREs, developers, and security enthusiasts who require instant, high-density telemetry without the latency or privacy risks of cloud-based monitoring.

> [!IMPORTANT]
> **100% Private. 100% Local. Zero Cloud.** Your data never leaves your hardware.

---

## 🖥️ The Command Center Experience
Universal-Ops replaces fragmented CLI tools with a unified, high-performance dashboard. It bridges the gap between raw kernel metrics and actionable intelligence using integrated local AI.

### 🧩 Core Substrates

| Layer | Focus | Key Capabilities |
| :--- | :--- | :--- |
| **SysOps** | Infrastructure | Per-core CPU, Memory pressure, Disk throughput, Thermal envelopes. |
| **NetOps** | Connectivity | Real-time ICMP Jitter, DNS Audit, Port Scanning, Traffic Attribution. |
| **SecOps** | Hardening | Firewall intelligence, Identity Audit, Elevated Privilege tracking. |
| **DevOps** | Automation | Service orchestration, Sandboxed Terminal, Process Impact analysis. |
| **AIOps** | Intelligence | Autonomous Root Cause Analysis (RCA) and Technical Briefings via **Ollama**. |

---

## 🚀 Quick Start
Ready to build your observation command center?

1.  **Download**: Get the latest `.exe` from [Releases](https://github.com/shahriarhaqueabir/AllOpsFull/releases).
2.  **Initialize**: Follow the **Interactive Onboarding Wizard** to link local tools.
3.  **Optimize**: Connect [Ollama](https://ollama.com) for local AI briefings.

**Detailed Installation Guide**: [Read the Manual](./docs/INSTALL.md)

---

## 🧠 Local Intelligence (Hawk)
Universal-Ops features **Hawk**, a built-in AI analyst powered by local LLMs.
- **Statistical Grounding**: No hallucination. Hawk analyzes summarized telemetry data.
- **Root Cause Analysis**: Jump from a CPU spike anomaly directly to a technical briefing.
- **Contextual Awareness**: 32k context window for long-horizon system event correlation.

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
- Join the discussion in [Issues](https://github.com/shahriarhaqueabir/AllOpsFull/issues).

---

## 📄 License
Distributed under the MIT License. See `LICENSE` for more information.

---
*Developed for professionals who value speed, privacy, and technical depth.*
