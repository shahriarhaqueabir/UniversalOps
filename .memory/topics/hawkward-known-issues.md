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

## Testing
- No integration tests for TUI wire-up (`R` key → workflow → display)
- Unit tests cover workflow functions only, not UI message routing

## Code Quality
- `internal/netops/ping.go`: `_ = echo` dead code after ICMP reply parsing (cleanup pending)
- `internal/netops/workflows.go`: removed unused `cmds` variable in prior session
