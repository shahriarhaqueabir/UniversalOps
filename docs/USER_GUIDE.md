# Universal-Ops User Guide

Welcome to Universal-Ops, the premium native GUI operations platform. This guide covers setup, usage, and customization.

## Table of Contents
1. [Installation & Setup](#1-installation--setup)
2. [AI Operations Setup](#2-ai-operations-setup)
3. [Features Deep Dive](#3-features-deep-dive)
4. [Customization](#4-customization)
5. [Troubleshooting](#5-troubleshooting)

---

## 1. Installation & Setup

### Prerequisites
- **Go 1.26+**: Required for building from source.
- **Wails v2 CLI**: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
- **Node.js**: Required for frontend builds.
- **Windows / Linux / macOS**: Universal-Ops is cross-platform.

### Building from Source
```bash
git clone https://github.com/youruser/opsforall.git
cd opsforall
wails build
```
The compiled binary will be located in the `build/bin/` directory.

### First Run
Each page is accessible from the sidebar. The five ops layers (SysOps, NetOps, SecOps, DevOps, AIOps) each provide a dedicated view.

---

## 2. AI Operations Setup

Universal-Ops uses **Ollama** by default for local, private AI operations.

### Installing Ollama
1. Download Ollama from [ollama.com](https://ollama.com/).
2. Install and run the Ollama application.
3. Open a terminal and initialize the intelligence substrate:
   ```bash
   ollama create universalops -f Modelfile
   ```

### Customizing the AI Model
By default, Universal-Ops looks for `universalops` (based on Qwythos-9B) at `http://localhost:11434`.
 Falls back to the first available model if the default is not found. You can override this using environment variables or the **Settings** page within the application.

- `OLLAMA_HOST`: The URL of your Ollama server (e.g., `http://192.168.1.50:11434`).
- `OLLAMA_MODEL`: The name of the model you want to use (e.g., `mistral`, `gemma`).

**Example (Windows PowerShell):**
```powershell
$env:OLLAMA_MODEL = "mistral"
./build/bin/Universal-Ops.exe
```

### Using External AI (OpenAI/Anthropic)
Currently, Universal-Ops
 focuses on local AI for privacy and offline capability. Integration with external APIs is planned for a future release. For now, you can use **LiteLLM** or a similar proxy to expose OpenAI-compatible endpoints as an Ollama-like service if needed.

---

## 3. Features Deep Dive

### SysOps (System Operations)
- **Mechanism**: Uses `gopsutil` to poll system metrics from the OS kernel.
- **Compute Audit**: Distinguishes between **Physical Cores** and **Logical Threads** (SMT).
- **Load Saturation**: Calculates a real-time **Saturation Index** (Load Avg relative to core count) to identify system bottlenecks.
- **Process Management**: Features a runtime audit with PID tracking and resource impact assessment.

### NetOps (Network Operations)
- **Mechanism**: 
  - **Ping**: Uses ICMP packets (or `ping.exe` fallback on Windows). Now includes **Jitter** calculation and a real-time **Latency History Chart** to detect routing instability.
  - **DNS**: Direct UDP/TCP queries. Now supports **Custom Resolvers** (e.g., 8.8.8.8) to audit local cache poisoning.
  - **Port Scan**: Concurrent TCP connection attempts. Optimized for speed (~200ms for full common port scan).
  - **Traceroute**: Sequential ICMP TTL-incrementing probes to map network paths.
  - **Bandwidth**: Real-time throughput monitoring with historical sparklines.

### SecOps (Security Operations)
- **Mechanism**: Orchestrates system tools like `netsh`, `netstat`, and PowerShell's `Get-MpComputerStatus`.
- **Firewall Intelligence**: Automatically identifies **High Risk** rules (Allow + Any IP + Sensitive Ports like RDP/SSH).
- **Listener Audit**: Flags processes listening on **External Interfaces** (0.0.0.0), highlighting potential exposure.
- **Identity Audit**: Lists local accounts with administrative privileges to enforce Least Privilege.

### DevOps (Development Operations)
- **Mechanism**: A safe wrapper around shell execution and file system APIs.
- **Interactive Terminal**: Destructive commands (rm, del) are intercepted and require explicit confirmation.
- **File Explorer**: Features **Breadcrumb Navigation** and **Binary Safety** (prevents viewing corrupted/binary data as text).
- **Service Control**: Orchestrates system services with an `sc query` fallback for non-admin environments.

### AI Ops (AI Operations)
- **Local Intelligence**: Uses **Ollama** for private, offline system analysis.
- **Model Discovery**: Automatically detects and lists all available local models (Llama, Mistral, etc.).
- **Interactive Analyst**: Features **Suggested Prompts** for common operational tasks like health reviews and anomaly analysis.
- **Anomaly Detection**: Compares live metrics against a rolling window of history to identify statistical deviations.

---

## 4. Customization

### Theming
- Use the **Theme Toggle** button in the TopBar to switch between **Dark** and **Light** modes.
- Universal-Ops uses a "Squib" design system with fluid CSS variables, adapting its palette for optimal readability and focus.

### Refresh Interval
The dashboard refreshes every 3 seconds by default. This can be configured in the **Settings** page, allowing you to balance data granularity with system overhead.

---

## 5. Troubleshooting

- **AI Ops not responding**: Ensure Ollama is running and you have initialized the substrate (`ollama create universalops -f Modelfile`).
- **Permission Denied (SecOps)**: Some security checks (like firewall rules) require Administrator/Sudo privileges.
- **High CPU usage by Universal-Ops**: This can happen if the refresh interval is set too low (e.g., < 500ms).
