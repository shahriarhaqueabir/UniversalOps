# Tools & Commands Reference

> Research notes on Go libraries and platform APIs used by Universal-Ops.
> This document catalogs the data sources and fallback mechanisms across the ops stack.

## Data Source Hierarchy

1. **Native Go libraries** (preferred)
2. **Windows APIs / WMI / ETW**
3. **PowerShell** (fallback)
4. **CLI tools** (last resort)

---

## System Information

| Feature | Go Library / API | Windows Command (Fallback) |
|---------|-----------------|---------------------------|
| CPU | `gopsutil/v4/cpu` | `wmic cpu` |
| Memory | `gopsutil/mem` | `systeminfo` |
| Disk | `gopsutil/disk` | `Get-Volume` |
| Partitions | `gopsutil/disk` | `Get-Partition` |
| Host | `gopsutil/host` | `hostname` |
| Users | Windows API | `Get-LocalUser` |
| Processes | `gopsutil/process` | `tasklist` |
| Services | Windows API | `Get-Service` |
| Uptime | `gopsutil/host` | `systeminfo` |
| Sensors | LibreHardwareMonitor API | WMI |

## CPU

Metrics: Usage %, Per Core, Frequency, Load Average, Temperature, Context Switches, Interrupts

```go
cpu.Percent()
cpu.Info()
cpu.Times()
```

## Memory

```go
mem.VirtualMemory()
mem.SwapMemory()
```

Metrics: Used, Cached, Buffers, Free, Swap

## Disk

```go
disk.Partitions()
disk.Usage()
disk.IOCounters()
```

Metrics: IO, Queue Length, Read Rate, Write Rate, Filesystem, SMART

## Processes

```go
process.Processes()
process.NumFDs()
```

Per Process: PID, Parent PID, Name, CPU, Memory, Threads, Handles, Open Files, User, Executable, Start Time
Open File Descriptors: `process.NumFDs()` (used by `openfds` collector)

## Load Average

```go
load.Avg()
```

Metrics: 1-minute, 5-minute, 15-minute load averages (used by `load` collector)

## Uptime

```go
host.Uptime()
```

Metrics: System uptime in seconds (used by `uptime` collector)

## Network

```go
gopsutil/net
```

Metrics: Interfaces, Connections, Bytes, Packets, Errors, Drop Rate

### Connections
```go
net.Connections()
```

Shows: TCP, UDP, LISTEN, ESTABLISHED, TIME_WAIT

### Bandwidth
```go
net.IOCounters()
```
Compute: `(Current - Previous) / Time = Bytes/sec`

## DNS

```go
net.LookupHost()
net.LookupIP()
net.LookupTXT()
net.LookupMX()
```

Advanced: `github.com/miekg/dns` — supports DNSSEC, EDNS, Zone Transfers, Custom Queries

## Ping

Library: `github.com/go-ping/ping` (or raw ICMP via `golang.org/x/net/icmp`)

Metrics: RTT, Packet Loss, TTL, Jitter

## Traceroute

Libraries: `github.com/aeden/traceroute` or `golang.org/x/net/icmp`

## Port Scanner

Libraries: `github.com/Ullaakut/nmap` or raw sockets via `net.DialTimeout()`

## Packet Capture

Library: `github.com/google/gopacket` (P4 — needs Npcap runtime on Windows)

## Firewall (Windows)

PowerShell: `Get-NetFirewallRule` or Windows Firewall COM API

## Defender (Windows)

PowerShell: `Get-MpComputerStatus`, `Get-MpThreat`, `Get-MpPreference`

## Event Logs

Native Windows Event Log API or `golang.org/x/sys/windows`

PowerShell fallback: `Get-WinEvent`

## Services

Windows API or `golang.org/x/sys/windows/svc`

PowerShell: `Get-Service`, `Start-Service`, `Stop-Service`, `Restart-Service`

## Hardware

Library: `gopsutil`, WMI via `github.com/StackExchange/wmi`

Can retrieve: CPU, Motherboard, BIOS, GPU, RAM, Manufacturer, Serial, Model

## GPU

Windows: DXGI, DirectX, NVML, AMD ADL, Intel APIs

## Installed Software

Registry: `HKLM\Software\Microsoft\Windows\CurrentVersion\Uninstall`

PowerShell: `Get-Package`

## Registry

`golang.org/x/sys/windows/registry`

Useful: Startup Apps, Installed Programs, Policies, Explorer, Windows Settings

## AI — Ollama

REST API: `POST /api/chat`, `/api/generate`, `/api/embed`

Future: Local embeddings, RAG, Function calling, MCP, Multi-agent workflows

## Charts

Current: Recharts (v3)
Considered: Apache ECharts, Tremor, React Flow (topology)

## Terminal

`os/exec` with `github.com/creack/pty` for interactive terminals (SSH, PowerShell, Bash)

## SSH

`golang.org/x/crypto/ssh` — future: remote diagnostics

## File Browser

`os`, `filepath`, `fs` — for watching: `github.com/fsnotify/fsnotify`

## Log Viewer

`bufio.Scanner` for live updates via `fsnotify`

Parsing: JSON, CSV, Syslog, Windows Event Log, Docker logs

## Notifications (Windows)

`github.com/go-toast/toast`

## Scheduling

`github.com/robfig/cron` — automatic reports, scans, snapshots, cleanup

## SQLite

`modernc.org/sqlite` (pure Go, no CGO)

Store: Metrics history, Alerts, Reports, AI conversations, Incidents, Snapshots, User settings, Audit trail

## Pipeline Architecture

```
Collectors → Aggregator → Snapshot Store (SQLite) → Correlation Engine → AI Context Builder → Frontend
```

Each module works from a consistent, timestamped view of the system, enabling historical comparisons, incident timelines, anomaly detection, and AI-assisted root-cause analysis without duplicating collection logic.
