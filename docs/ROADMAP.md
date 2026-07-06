# Hawkward Roadmap

## Phase 1 — Foundation ✅

- [x] Project scaffolding (Go module layout)
- [x] Architecture document
- [x] Development standards
- [x] Root model with screen routing
- [x] Main menu with keyboard navigation
- [x] Help overlay system
- [x] Onboarding wizard (5-step)
- [x] Status bar with system health
- [x] SysOps layer — CPU, memory, disk, processes, system info
- [x] Lip Gloss styling system
- [x] gopsutil integration for cross-platform metrics

## Phase 2 — NetOps Layer ✅

- [x] Network interface discovery and monitoring
- [x] ICMP ping tool with live results (`ping.exe` fallback on Windows)
- [x] DNS lookup tool (A, AAAA, MX, NS via `miekg/dns`)
- [x] TCP port scanner (`net.DialTimeout`)
- [x] Connection table (TCP/UDP via `netstat -ano`)
- [x] Traceroute implementation
- [x] Network bandwidth monitoring with sparklines

## Phase 3 — SecOps Layer ✅

- [x] Firewall rule viewer (`netsh advfirewall`)
- [x] Local user and group audit (`net user` / `net localgroup`)
- [x] Listening port scanner with process attribution (`netstat` + `tasklist`)
- [x] Windows Defender status (`Get-MpComputerStatus`)
- [x] Scheduled tasks viewer (`Get-ScheduledTask`)
- [x] Security event log reader

## Phase 4 — DevOps Layer ✅

- [x] Command runner with output display
- [x] Log tailer with pattern search
- [x] File browser (directory listing + file reading)
- [x] Process manager (view, kill, restart)
- [x] Service status dashboard

## Phase 5 — AI Ops Layer ✅

- [x] Ollama integration for local AI (`localhost:11434/api/chat`)
- [x] Chat interface with message history
- [x] Automated report generation (text + markdown)
- [x] Report export to file
- [x] Natural language query of system state
- [x] Anomaly detection from metrics patterns

## Phase 6 — Polish & Release 🔜

- [x] All 5 layers wired into UI
- [x] Cross-platform build (Windows, Linux, macOS verified)
- [ ] Cross-platform testing (Linux, macOS)
- [x] Configurable refresh intervals
- [x] Theme system (default, dark, light, high-contrast)
- [x] Color customization persistence
- [ ] Session logging
- [x] Export reports to JSON/CSV
- [ ] Release binaries via GitHub Actions
- [ ] Homebrew tap for macOS
- [ ] Scoop/Chocolatey for Windows

## Technical Debt & Improvements

- [x] Add comprehensive test coverage (≥30% for core packages)
- [ ] Benchmark and optimize render performance
- [ ] Graceful degradation on terminals without mouse support
- [x] Proper error recovery on gopsutil failures
- [ ] Cross-platform user/group info (Linux `/etc/passwd`, macOS `dscl`)
