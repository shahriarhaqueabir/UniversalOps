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

## Phase 6 — Polish & Release ✅

- [x] All 5 layers wired into UI
- [x] Cross-platform build (Windows, Linux, macOS verified)
- [ ] Cross-platform testing (Linux, macOS) — *in progress*
- [x] Configurable refresh intervals
- [x] Theme system (default, dark, light, high-contrast)
- [x] Color customization persistence
- [ ] Session logging — *moved to v3 Sprint 5*
- [x] Export reports to JSON/CSV
- [ ] Release binaries via GitHub Actions — *post-v3*
- [ ] Homebrew tap for macOS — *post-v3*
- [ ] Scoop/Chocolatey for Windows — *post-v3*

## Phase 7 — v3: Interactive Visual Operations Platform 🔄

**Vision**: Transform Hawkward from a tab-based text tool into a squib-inspired interactive card-based dashboard with real-time charts, forecasting, and a global command palette.

### Sprint 1: Visualization Engine (Foundation) 🔜

| Task | Status | Files | Depends On |
|------|--------|-------|-----------|
| Chart component library | **Planned** | `internal/common/charts/*.go` (8 files) | None |
| Time-series data store | **Planned** | `internal/common/timeseries.go` | None |
| Forecast engine | **Planned** | `internal/common/forecast.go` | None |
| Card component system | **Planned** | `internal/ui/cards.go` | Charts |
| YAML config system | **Planned** | `internal/common/config.go` | None |

**Verification**: `go test ./internal/common/charts/...` + `go build ./...`

### Sprint 2: Dashboard & Layer Redesign 🔜

| Task | Status | Files | Depends On |
|------|--------|-------|-----------|
| Dashboard landing page | **Planned** | `internal/ui/dashboard.go` | Sprint 1 |
| SysOps card rewrite | **Planned** | `internal/sysops/view.go` | Sprint 1 |
| NetOps card rewrite | **Planned** | `internal/netops/view.go` | Sprint 1 |
| SecOps card rewrite | **Planned** | `internal/secops/view.go` | Sprint 1 |
| DevOps/AIOps card rewrite | **Planned** | `internal/devops/view.go`, `internal/aiops/view.go` | Sprint 1 |
| Interaction layer | **Planned** | `internal/ui/cards.go` | Sprint 1 |
| Alerting system | **Planned** | `internal/common/alerts.go` | Sprint 1 |

**Verification**: Manual walkthrough all 5 layers + dashboard navigation

### Sprint 3: Intelligence & Forecasting 🔜

| Task | Status | Files | Depends On |
|------|--------|-------|-----------|
| Predictive analytics | **Planned** | `internal/common/forecast.go` + all view.go | Sprint 2 |
| Alerting & incidents | **Planned** | `internal/common/alerts.go` | Sprint 2 |
| AI-powered insights | **Planned** | `internal/aiops/ollama.go` | Sprint 3.1 |
| Forecast visualization | **Planned** | `internal/common/charts/line.go` | Sprint 3.1 |

**Verification**: `go test ./internal/common/forecast_test.go` + visual inspection

### Sprint 4: Search & Power User Features 🔜

| Task | Status | Files | Depends On |
|------|--------|-------|-----------|
| Global command palette | **Planned** | `internal/ui/commandpalette.go` | Sprint 2 |
| Keyboard workflow engine | **Planned** | `internal/ui/keys.go` | Sprint 2 |
| Data drill-down system | **Planned** | All view.go + DetailPanel | Sprint 2 |
| Data export | **Planned** | All view.go + charts | Sprint 2 |

**Verification**: `go build ./...` + manual test of `/`, `t`, `r`, drill-down

### Sprint 5: Config, Polish & Cross-Platform 🔜

| Task | Status | Files | Depends On |
|------|--------|-------|-----------|
| Config UI | **Planned** | `internal/ui/settings.go` | Sprint 1.5 |
| Responsive layout engine | **Planned** | `internal/ui/cards.go` | All |
| Performance optimization | **Planned** | `internal/ui/root.go`, charts | All |
| Cross-platform hardening | **Planned** | `internal/secops/users.go`, `internal/aiops/ollama.go` | All |
| Session logging & audit | **Planned** | `internal/common/logger.go`, `internal/ui/logviewer.go` | All |

**Verification**: `GOOS=linux GOARCH=amd64 go build` + `go test -race ./...`

---

## v3 Milestone Summary

| Milestone | Sprint | Target Date | Key Deliverable |
|-----------|--------|-------------|----------------|
| Foundation | Sprint 1 | Week 2 | Chart library + time-series + cards build & test |
| Visual Launch | Sprint 2 | Week 6 | Dashboard + all layer views rewritten with cards |
| Intelligence | Sprint 3 | Week 7 | Forecasts, alerts, AI insights working |
| Power User | Sprint 4 | Week 8 | Command palette, macros, drill-down, export |
| Ship | Sprint 5 | Week 9 | v3.0.0 release with config, polish, cross-platform |

## Technical Debt & Improvements

- [x] Add comprehensive test coverage (≥30% for core packages)
- [ ] Benchmark and optimize render performance — *v3 Sprint 5*
- [ ] Graceful degradation on terminals without mouse support
- [x] Proper error recovery on gopsutil failures
- [ ] Cross-platform user/group info (Linux `/etc/passwd`, macOS `dscl`) — *v3 Sprint 5*
- [ ] Config system (YAML-based) — *v3 Sprint 1*
- [ ] Session logging with rotation — *v3 Sprint 5*

---

*Last updated: 2026-07-07*
