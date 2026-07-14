# Staff Engineering Code Review Audit — 2026-07-14

## Source
`reviewandaudits/AuditReport.md` + `AuditReportRemediation.md`

## Summary
4 parallel audit agents, **61 findings** across 4 domains:
- Backend Security & Go Patterns
- Frontend React & Performance
- IPC Wiring & Data Flow
- Cross-Cutting Concerns

## Severity Breakdown
| Severity | Count | Key Themes |
|----------|-------|------------|
| CRITICAL | 12 | Command injection (4), sandbox bypass, Prometheus config exposure, IPC field mismatch, data races, async flooding, deprecated API |
| HIGH | 20 | Global state races, hardcoded colors, god files (DevOps.go 2107 lines), missing TS types, stub implementations |
| MEDIUM | 14 | Dead code, unbounded lists, logging, bundle size, reliability |
| LOW | 15 | Convention issues, minor UX, technical debt |

## Critical Findings (Immediate Action)

| ID | Area | File | Issue |
|----|------|------|-------|
| SEC-1 | Command Injection | netops/actions.go | Interface name → exec.Command unsanitized |
| SEC-2 | Command Injection | secops/security.go | Username → net user/usermod unsanitized |
| SEC-3 | Command Injection | secops/response.go | PID string → taskkill without numeric validation |
| SEC-4 | Command Injection | devops/services.go | Service name → systemctl/sc unsanitized |
| SEC-5 | Sandbox Bypass | app/DevOps.go | app.Sandbox == nil = unrestricted execution |
| SEC-6 | Destructive IR | secops/response.go | IsolateHost blocks ALL traffic, no undo timer |
| SEC-7 | Config Exposure | common/metrics_exporter.go | :9210/config returns full config unauthenticated |
| IPC-1 | Type Mismatch | Types.go:861 / types/index.ts:839 | LogSummary.trend: Go sends "errorTrend", TS expects "trend" |
| IPC-2 | Silent Crash | hooks/useBackend.ts | Returns null on method resolution failure |
| IPC-3 | Data Race | App.go:323-353 | evaluateAndEmit reads/writes lastCPU/lastMem/lastDisk from 2 goroutines |
| FE-1 | Flooding | AnalysisSidebar.tsx | useEffect fires SetTopology every render (unmemoized) |
| CROSS-1 | Deprecated | app/AIOps.go:191 | strings.Title deprecated since Go 1.18 |

## High Findings (Before Release)

Key systemic issues:
- **H1**: devops/shell.go blocklist bypassable via encoding/eval
- **H2-H6**: Concurrency races in netops, alerts, scheduler, aiops, sysops
- **H7-H9**: Stub implementations (parseCertJSON, parseBitLockerJSON, parseLogOutput)
- **H11-H15**: TypeScript/Go type mismatches (GitStashEntry, PingStats, ServiceInfo, LogTimelinePoint, LogSummary)
- **H16**: DevOps.go = 2107-line god file (2.5× next largest)
- **H17-H18**: Hardcoded colors (text-white, bg-black) in 50+ files
- **H19**: GetAlerts()/ActiveAlerts() duplication
- **H20**: Stale "hawkward" references in .gitignore

## Cross-Cutting Themes (8 Systemic Issues)

1. **Command injection in incident response** (4 CRITICAL in SecOps) — BlockIP, DisableAccount, SetFirewallRuleState concatenate user input into shell commands
2. **Data races in shared state** (6+ CRITICAL/HIGH) — TimeSeries pointers leaked, NetOps model no mutex, AIOps effectiveModel race, App evaluateAndEmit concurrent, common.activeTheme no mutex, globalStorage no sync
3. **Systematic error swallowing** (15+ HIGH/MEDIUM) — `_` discards on security audit, git, K8s, storage init
4. **Linux sudo in GUI = guaranteed failure** (2 CRITICAL) — sudo needs TTY; reboot, shutdown, sleep, hibernate, DNS flush, pkg cache clean, system update dead on Linux
5. **Sandbox bypass across 3 modules** (5 HIGH) — ~60% of OS commands skip common.SandboxedCommand
6. **Business logic in binding layer** (3 HIGH) — DevOps.go 2107 lines, SecOps.go 300+ lines scoring, Dashboard.go presentation logic
7. **Security stubs = false negatives** (5 HIGH) — parseCertJSON, parseBitLockerJSON, parseServicesJSON, parseFailedLoginsJSON all empty; getTLSCertificatesLinux lists symlinks as certs
8. **Frontend systemic issues** (4 CRITICAL) — DevOps.tsx ~1500 lines, @ts-nocheck in ALL tests, 3 zustand stores in 1 file, no-explicit-any disabled

## Remediation Protocol (from AuditReportRemediation.md)

1. **Verify every claim** against current source before fixing (audits go stale)
2. **One issue = one branch/commit** — never bundle unrelated fixes
3. **Fix not done until**: defect reproduced → fix applied → regression test passes → residual risk stated
4. **Don't silently expand scope** — if fix requires touching Y, report before proceeding
5. **Don't delete error handling** to make code compile — surface it
6. **If claim not reproducible** → report "claim not verified" with evidence, don't force-fit
7. **After each batch**: run tests + linter + build for all 3 OSes

## Next Steps

Per remediation doc: **Verification Pass First** — confirm each finding against current codebase before any fixes. Classify as:
- CONFIRMED
- PARTIALLY-CONFIRMED (inaccurate description)
- NOT-FOUND (code changed)
- DUPLICATE (same root cause as another)

Then group by root cause (cross-cutting themes) and fix in severity order.