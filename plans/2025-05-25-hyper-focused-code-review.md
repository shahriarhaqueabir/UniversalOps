# Sprint: Hyper-Focused Code Review (Layer-Driven)

This sprint focuses on a deep, line-by-line review of the OpsForAll codebase, organized by architectural layer. Each ticket target a specific system or component to ensure security, performance, and naming consistency (addressing the Hawkward -> OpsForAll debt).

## Layer 1: Infrastructure & Common (`internal/common`)
| ID | Ticket | Status | Priority | File Scope | DOD |
|----|--------|--------|----------|------------|-----|
| CR-01 | **Security Sandbox Audit** | 🔄 IN PROGRESS | Critical | `sandbox.go`, `sandbox_windows.go` | - [ ] Verify Integrity Level logic on Windows - [ ] Confirm recent fix for ReadOnlyFS doesn't leak privileges |
| CR-02 | **Data Pipeline & Forecasting** | 🔲 TODO | High | `pipeline.go`, `forecast.go`, `timeseries.go` | - [ ] Audit math for NaN/Inf safety in regressions - [ ] Check ring buffer concurrency locks |
| CR-03 | **Persistence & Retention** | 🔲 TODO | High | `storage.go` | - [ ] Review async batch writing for race conditions - [ ] Verify daily prune SQL efficiency |
| CR-04 | **Alert Engine & Templates** | 🔲 TODO | Med | `alerts.go` | - [ ] Audit flap detection logic - [ ] Verify string replacement safety in templates |

## Layer 2: Domain Logic (`internal/{sysops,netops,secops,devops,aiops}`)
| ID | Ticket | Status | Priority | Area | DOD |
|----|--------|--------|----------|------|-----|
| CR-05 | **NetOps: Connectivity Logic** | 🔲 TODO | High | `netops/` | - [ ] Review Ping fallback logic (Raw vs Exec) - [ ] Check for command injection in traceroute/dns |
| CR-06 | **SecOps: Vulnerability Heuristics** | 🔲 TODO | High | `secops/` | - [ ] Audit sensitive port list - [ ] Verify firewall rule parsing for non-English locales |
| CR-07 | **DevOps: Shell & Concurrency** | 🔲 TODO | Critical | `devops/shell.go` | - [ ] Audit WaitGroup fix for streaming output - [ ] Review dangerous command blocklist completeness |
| CR-08 | **AIOps: Anomaly Detection** | 🔲 TODO | Med | `aiops/` | - [ ] Review statistical threshold logic - [ ] Check Ollama client timeout handling |
| CR-09 | **SysOps: Telemetry Accuracy** | 🔲 TODO | Low | `sysops/` | - [ ] Cross-check gopsutil usage with upstream best practices |

## Layer 3: Application Bindings (`internal/app`)
| ID | Ticket | Status | Priority | File Scope | DOD |
|----|--------|--------|----------|------------|-----|
| CR-10 | **Wails Facade Audit** | 🔲 TODO | Med | `app/*.go` | - [ ] Ensure all Go methods are type-safe for JS - [ ] Check for proper error sanitization before returning to UI |
| CR-11 | **Event Bus & Tick Loop** | 🔲 TODO | High | `App.go` | - [ ] Audit goroutine leak potential in `collectAndEmit` - [ ] Check event serialization overhead |

## Layer 4: Frontend Presentation (`cmd/opsforall-gui/frontend`)
| ID | Ticket | Status | Priority | File Scope | DOD |
|----|--------|--------|----------|------------|-----|
| CR-12 | **State Management & Query Logic** | 🔲 TODO | High | `src/hooks/`, `src/stores/` | - [ ] Audit `useBackend` error handling - [ ] Verify query invalidation patterns |
| CR-13 | **Component Safety & Performance** | 🔲 TODO | Med | `src/pages/`, `src/components/` | - [ ] Audit Terminal line rendering performance - [ ] Check Radix UI accessibility compliance |

## Global Cross-Cutting Review
| ID | Ticket | Status | Priority | Focus | DOD |
|----|--------|--------|----------|-------|-----|
| CR-14 | **Naming Debt Cleanup** | 🔲 TODO | High | Entire Repo | - [ ] Identify all remaining `hawkward` references - [ ] Plan safe rename strategy for internal paths |
