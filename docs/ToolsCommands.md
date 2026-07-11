
Since you're using **Go + Wails v2 + React + TypeScript**, I'd avoid shelling out to Windows commands unless absolutely necessary. Native APIs and Go libraries are generally faster, more reliable, and easier to support across platforms.

Think of data sources in four layers:

1. **Native Go libraries** (preferred)
2. **Windows APIs / WMI / ETW**
3. **PowerShell** (fallback)
4. **CLI tools** (last resort)

---

# System Information

| Feature    | Go Library / API                    | Windows Command (Fallback)                    |
| ---------- | ----------------------------------- | --------------------------------------------- |
| CPU        | `github.com/shirou/gopsutil/v4/cpu` | `wmic cpu`, `Get-CimInstance Win32_Processor` |
| Memory     | `gopsutil/mem`                      | `systeminfo`                                  |
| Disk       | `gopsutil/disk`                     | `Get-Volume`                                  |
| Partitions | `gopsutil/disk`                     | `diskpart`, `Get-Partition`                   |
| Host       | `gopsutil/host`                     | `hostname`                                    |
| Users      | Windows API                         | `Get-LocalUser`                               |
| Processes  | `gopsutil/process`                  | `tasklist`                                    |
| Services   | Windows API                         | `Get-Service`                                 |
| Uptime     | `gopsutil/host`                     | `systeminfo`                                  |
| Sensors    | LibreHardwareMonitor API            | WMI                                           |

---

# CPU

Useful metrics

* Usage %
* Per Core
* Frequency
* Load Average
* Temperature
* Context Switches
* Interrupts

Go

```go
cpu.Percent()
cpu.Info()
cpu.Times()
```

---

# Memory

```go
mem.VirtualMemory()
mem.SwapMemory()
```

Metrics

* Used
* Cached
* Buffers
* Free
* Swap

---

# Disk

```go
disk.Partitions()
disk.Usage()
disk.IOCounters()
```

Metrics

* IO
* Queue Length
* Read Rate
* Write Rate
* Filesystem
* SMART

For SMART

Use

* smartctl
* CrystalDiskInfo SDK
* NVMe APIs

---

# Processes

```go
process.Processes()
```

Per Process

* PID
* Parent PID
* Name
* CPU
* Memory
* Threads
* Handles
* Open Files
* User
* Executable
* Start Time

Terminate

```go
proc.Kill()
```

Suspend

Windows API

Resume

Windows API

---

# Network

Go

```go
gopsutil/net
```

Metrics

Interfaces

Connections

Bytes

Packets

Errors

Drop Rate

---

Connections

```go
net.Connections()
```

Shows

TCP

UDP

LISTEN

ESTABLISHED

TIME_WAIT

---

Bandwidth

```go
net.IOCounters()
```

Compute

```
Current - Previous

÷

Time

=

Bytes/sec
```

---

# DNS

Go

```go
net.LookupHost()

net.LookupIP()

net.LookupTXT()

net.LookupMX()
```

Advanced

Use

`github.com/miekg/dns`

Supports

* DNSSEC
* EDNS
* Zone Transfers
* Custom Queries

---

# Ping

Use

`github.com/go-ping/ping`

Metrics

* RTT
* Packet Loss
* TTL
* Jitter

---

# Traceroute

Libraries

```
github.com/aeden/traceroute

```

or

```
golang.org/x/net/icmp

```

---

# Port Scanner

Libraries

```
github.com/Ullaakut/nmap

```

or

Raw sockets

```
net.DialTimeout()
```

---

# Packet Capture

For future

```
github.com/google/gopacket

```

Can power

* Wireshark-like capture
* Protocol Analysis
* ARP
* DHCP
* DNS
* HTTP

---

# Firewall

Windows

PowerShell

```
Get-NetFirewallRule
```

Or

Windows Firewall COM API

Better

Native Windows API

---

# Defender

PowerShell

```
Get-MpComputerStatus
```

```
Get-MpThreat
```

```
Get-MpPreference
```

---

# Event Logs

Native

Windows Event Log API

Go

```
golang.org/x/sys/windows

```

or

PowerShell

```
Get-WinEvent
```

---

# Services

Windows API

or

```
golang.org/x/sys/windows/svc
```

PowerShell

```
Get-Service

Start-Service

Stop-Service

Restart-Service
```

---

# Hardware

Libraries

```
gopsutil

```

WMI

```
github.com/StackExchange/wmi
```

Can retrieve

CPU

Motherboard

BIOS

GPU

RAM

Manufacturer

Serial

Model

---

# GPU

Windows

DXGI

DirectX

NVML

AMD ADL

Intel APIs

---

# Installed Software

Registry

```
HKLM

Software

Microsoft

Windows

CurrentVersion

Uninstall
```

or

PowerShell

```
Get-Package
```

---

# Registry

Go

```
golang.org/x/sys/windows/registry
```

Useful

Startup Apps

Installed Programs

Policies

Explorer

Windows Settings

---

# AI

Ollama

REST API

```
POST

/api/chat

/api/generate

/api/embed

```

Future

* Local embeddings
* RAG
* Function calling
* MCP
* Multi-agent workflows

---

# Charts

Current best options

React

* Recharts
* Tremor
* Apache ECharts ⭐
* React Flow (topology)

I'd replace most charts with **Apache ECharts**. It handles real-time updates, heatmaps, timelines, gauges, and large datasets far better than Recharts.

---

# Terminal

Go

```
os/exec
```

PTY

```
github.com/creack/pty
```

Needed for

Interactive terminals

SSH

PowerShell

Bash

---

# SSH

```
golang.org/x/crypto/ssh
```

Future

Remote diagnostics

---

# File Browser

```
os

filepath

fs

```

For watching changes

```
github.com/fsnotify/fsnotify
```

---

# Log Viewer

```
bufio.Scanner
```

For live updates

```
fsnotify
```

For parsing

* JSON
* CSV
* Syslog
* Windows Event Log
* Docker logs

---

# Notifications

Windows

```
github.com/go-toast/toast
```

---

# Scheduling

Cron

```
github.com/robfig/cron
```

Useful

Automatic reports

Scans

Snapshots

Cleanup

---

# SQLite

I would absolutely include a local database.

Use

```
modernc.org/sqlite
```

or

```
github.com/mattn/go-sqlite3
```

Store

* Metrics history
* Alerts
* Reports
* AI conversations
* Incidents
* Snapshots
* User settings
* Audit trail

---

# State-of-the-Art Libraries Worth Adding

| Purpose            | Library                             |
| ------------------ | ----------------------------------- |
| System metrics     | `gopsutil`                          |
| WMI access         | `StackExchange/wmi`                 |
| Packet capture     | `google/gopacket`                   |
| Nmap integration   | `Ullaakut/nmap`                     |
| DNS                | `miekg/dns`                         |
| Ping               | `go-ping/ping`                      |
| PTY terminal       | `creack/pty`                        |
| SSH                | `x/crypto/ssh`                      |
| File watching      | `fsnotify`                          |
| Cron jobs          | `robfig/cron`                       |
| SQLite             | `modernc.org/sqlite`                |
| Charts             | Apache ECharts                      |
| Topology           | React Flow                          |
| Virtualized tables | TanStack Virtual                    |
| Data tables        | AG Grid Community or TanStack Table |
| Command palette    | `cmdk`                              |
| Keyboard shortcuts | `react-hotkeys-hook`                |
| Dockable layouts   | `FlexLayout` or `GoldenLayout`      |

## Architectural recommendation

Instead of each page calling system libraries independently, build a **collector pipeline**:

```
Collectors
    │
    ├── CPU
    ├── Memory
    ├── Disk
    ├── Network
    ├── Processes
    ├── Services
    ├── Security
    ├── Logs
    │
Aggregator
    │
Snapshot Store (SQLite)
    │
Correlation Engine
    │
AI Context Builder
    │
Frontend (Dashboard, SysOps, NetOps, SecOps, etc.)
```

This architecture lets every module work from a consistent, timestamped view of the system, making historical comparisons, incident timelines, anomaly detection, and AI-assisted root-cause analysis much easier to implement without duplicating collection logic.
