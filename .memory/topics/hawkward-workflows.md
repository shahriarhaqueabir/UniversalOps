# Hawkward — Workflow Reports

## Per-Ops Reports
- **SysOps** — `RunHealthCheck()` → `*HealthReport`: CPU, memory, disk, processes, system info
- **NetOps** — `RunNetworkDiagnostics()` → `*NetworkReport`: Ping, DNS (miekg/dns), port scan, connections, interfaces
- **SecOps** — `RunSecurityAudit()` → `*SecurityReport`: Firewall (netsh/iptables), users, listening ports, Defender, scheduled tasks
- **DevOps** — `RunDevDiagnostics()` → `*DevReport`: Shell runner, log tailer, file browser
- **AIOps** — `GenerateEnhancedReport()` → `*EnhancedReport`: Ollama chat, report generator, text/markdown/save

## Viewing
- All wired into TUI via `R` key, displayed in full-panel report view
- Each report has `String()` (plain text) and `Markdown()` renderers
