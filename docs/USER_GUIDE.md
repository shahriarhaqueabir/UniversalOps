# Hawkward User Guide

Welcome to Hawkward, the terminal-based operations platform. This guide covers setup, usage, and customization.

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
- **Windows / Linux / macOS**: Hawkward is cross-platform.

### Building from Source
```bash
git clone https://github.com/youruser/hawkward.git
cd hawkward
go build -o hawkward.exe ./cmd/hawkward
./hawkward.exe
```

### First Run
On the first launch, Hawkward will walk you through an onboarding wizard. This will explain the basic navigation and the five "Ops Layers."

---

## 2. AI Operations Setup

Hawkward uses **Ollama** by default for local, private AI operations.

### Installing Ollama
1. Download Ollama from [ollama.com](https://ollama.com/).
2. Install and run the Ollama application.
3. Open a terminal and pull the default model:
   ```bash
   ollama pull llama3.2
   ```

### Customizing the AI Model
By default, Hawkward looks for `llama3.2` at `http://localhost:11434`. You can change this using environment variables:

- `OLLAMA_HOST`: The URL of your Ollama server (e.g., `http://192.168.1.50:11434`).
- `OLLAMA_MODEL`: The name of the model you want to use (e.g., `mistral`, `gemma`).

**Example (Windows PowerShell):**
```powershell
$env:OLLAMA_MODEL = "mistral"
./hawkward.exe
```

### Using External AI (OpenAI/Anthropic)
Currently, Hawkward focuses on local AI for privacy and offline capability. Integration with external APIs is planned for a future release. For now, you can use **LiteLLM** or a similar proxy to expose OpenAI-compatible endpoints as an Ollama-like service if needed.

---

## 3. Features Deep Dive

### SysOps (System Operations)
- **Mechanism**: Uses the `gopsutil` library to poll system metrics directly from the OS kernel/WMI.
- **Commands**: 
  - `1`: Overview dashboard (CPU/RAM/Disk).
  - `2`: Process table (sorted by CPU).
  - `3`: Detailed host info.

### NetOps (Network Operations)
- **Mechanism**: 
  - **Ping**: Uses ICMP packets (or `ping.exe` fallback on Windows).
  - **DNS**: Direct UDP/TCP queries to your system's configured nameservers.
  - **Port Scan**: Sequential TCP connection attempts with a short timeout.

### SecOps (Security Operations)
- **Mechanism**: Orchestrates system tools like `netsh`, `netstat`, and PowerShell's `Get-MpComputerStatus` to audit security posture.

### DevOps (Development Operations)
- **Mechanism**: A safe wrapper around shell execution and file system APIs.
- **Log Tailer**: Uses a sliding window buffer to follow file updates in real-time.

### AI Ops (AI Operations)
- **Natural Language Queries**: You can ask "What is my CPU usage?" and Hawkward will parse your local system state to answer without needing an LLM if the query is deterministic. For complex questions, it routes to Ollama.
- **Anomaly Detection**: Hawkward maintains a short history of metrics. It triggers a warning in the status bar if it detects sustained high usage or sudden spikes.

---

## 4. Customization

### Theming
- Press `t` to cycle through available themes: **Default**, **Dark**, **Light**, and **High Contrast**.
- **Custom Themes**: You can define a custom palette by setting the `HAWKWARD_THEME` environment variable to a JSON path (planned).

### Refresh Interval
The dashboard refreshes every 3 seconds by default. This can be configured in the source or via future CLI flags.

---

## 5. Troubleshooting

- **AI Ops not responding**: Ensure Ollama is running and you have pulled the model (`ollama pull llama3.2`).
- **Permission Denied (SecOps)**: Some security checks (like firewall rules) require Administrator/Sudo privileges.
- **High CPU usage by Hawkward**: This can happen if the refresh interval is set too low (e.g., < 500ms).
