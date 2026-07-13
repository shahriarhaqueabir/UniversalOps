# Hawkward — Known Issues & Pitfalls

## Build/Dev
- Go binary not in `$PATH` — always use full path `/e/Program\ Files/Go/bin/go`
- Path has spaces — requires escaping in shell

## AIOps
- Depends on Ollama running at `localhost:11434`
- App degrades gracefully but does not start Ollama automatically

## SecOps
- Defender and scheduled tasks use PowerShell — may fail in locked-down environments
- WMI/CIM fallback implemented for Defender and scheduled tasks

## Display
- Large rule/connection/task sets truncated: 100 firewall rules, 20 connections, 20 tasks — intentional but undocumented

## NetOps
- Ping on Windows skips raw ICMP (needs admin) and uses ping.exe directly — resolved Sprint 16
- DNS falls back to system resolver when public DNS servers (8.8.8.8, 1.1.1.1) are blocked — resolved Sprint 16
- Connection process names now resolved from PIDs via tasklist on Windows — resolved Sprint 16

## Testing
- GUI testing via Vitest + RTL for React components; Go unit tests for backend workflows

## Resolved (Sprint 14)
- ~NetOps Bandwidth tab had no content~ — Fixed: Added Recharts AreaChart showing RX/TX Mbps per interface
- ~SysOps Disk data not displayed~ — Fixed: Added Disk card with partition usage bars in Overview tab
- ~Settings GitHub/Docs links were `href="#"`~ — Fixed: Wired to real GitHub URLs with target=_blank
- ~Settings not persisted~ — Fixed: localStorage save/load for interval, pingCount, dnsTimeout
- ~ANSI escape codes in Terminal~ — Fixed: stripAnsi() applied to all terminal output
- ~Logs column headers not sticky~ — Fixed: Headers moved to direct child of scroll container

## Code Quality
- (none — all previously identified dead code has been resolved)
