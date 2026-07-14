Staff Engineering Code Review — Hawkward GUI (AllOpsFull)
Reviewed by 4 parallel audit agents covering: Backend Security & Go Patterns, Frontend React & Performance, IPC Wiring & Data Flow, and Cross-Cutting Concerns.

Executive Summary
Severity	Count	Theme
CRITICAL	12	Command injection, data races, field mismatches, sandbox bypass
HIGH	20	Global state races, hardcoded colors, god files, missing TS types
MEDIUM	14	Dead code, unbounded lists, logging, bundle size
LOW	15	Convention issues, minor UX, technical debt
Total	61	
Strengths noted by reviewers:

Zero panic() calls — RecoverPanic() used consistently
Zero SQL injection — all parameterized
Clean Collector → Pipeline → Forecast → Alert architecture
Comprehensive design system (10 palettes, CSS variables, reduced-motion)
Solid CI/CD on 3 OS platforms
Zero TODO/FIXME/HACK comments — excellent discipline
CRITICAL — Fix Immediately
SEC-1: Command Injection — NetOps interface names
internal/netops/actions.go — params["interface"] → exec.Command with zero sanitization. Crafted name like eth0; rm -rf / = full compromise. Fix: Whitelist against GetInterfaces() output or regex ^[a-zA-Z0-9\-_.]+$.

SEC-2: Command Injection — SecOps user management
internal/secops/security.go — Username → net user/usermod commands unsanitized. Fix: Whitelist against GetLocalUsers() output; validate ^[a-zA-Z0-9._-]+$.

SEC-3: Command Injection — SecOps KillProcess
internal/secops/response.go:KillProcess() — PID string → taskkill /PID + pid without numeric validation. Fix: strconv.Atoi(pid) with error check.

SEC-4: Command Injection — DevOps services
internal/devops/services.go — Service name → systemctl/sc commands unsanitized. Fix: Whitelist against GetServices() output.

SEC-5: Sandbox bypass — nil sandbox = unrestricted
internal/app/DevOps.go — When app.Sandbox == nil (default), all commands execute unrestricted. Fix: Initialize sandbox by default or return error when nil.

SEC-6: Destructive IsolateHost with no undo
internal/secops/response.go:IsolateHost() — Adds firewall rules blocking ALL traffic with no confirmation/rollback timer. Fix: Auto-expiry timer (60s undo window) or confirmation gate.

SEC-7: Prometheus /config endpoint exposes config unauthenticated
internal/common/metrics_exporter.go:9210 — Full app config returned to any localhost process. Fix: Basic auth/token check, or remove /config endpoint.

IPC-1: LogSummary.trend field name mismatch — AI Summary broken
internal/app/Types.go:861 sends "errorTrend", TS types/index.ts:839 declares trend. Frontend reads undefined. Fix: Rename Go json tag to json:"trend".

IPC-2: useBackend.call() returns null → silent crash
hooks/useBackend.ts:28-34 — Returns null on method resolution failure; callers cast as SomeType, then crash on property access. Fix: Throw error instead of returning null.

IPC-3: Data race on lastCPU/lastMem/lastDisk
internal/app/App.go:323-353 — evaluateAndEmit() reads/writes from both alert loop goroutine and TriggerCollector() binding goroutine without synchronization. Fix: Mutex or serialize onto single goroutine.

FE-1: Async useEffect topology sync flooding backend
AnalysisSidebar.tsx:153-161 — useEffect fires NetDesign.SetTopology on every render due to unmemoized array props. Fix: Debounce 500ms or memoize stringified inputs.

CROSS-1: strings.Title deprecated since Go 1.18
internal/app/AIOps.go:191 — Will fail on future Go versions. Fix: cases.Title(language.English).String() from golang.org/x/text/cases.

HIGH — Fix Before Release
#	Area	File:Line	Issue	Fix
H1	Security	devops/shell.go	Blocklist filtering bypassable via encoding/eval	Allowlist approach or sandbox
H2	Concurrency	netops/monitoring.go	Global mutable state + goroutine leak on double-start	Struct with lifecycle guard
H3	Concurrency	common/alerts.go:AlertCount	Write lock for read-only operation	RLock
H4	Concurrency	common/scheduler.go:Stop()	Goroutine leak on hung collector	Context cancellation
H5	Concurrency	aiops/ollama.go	effectiveModel global race	atomic.Value
H6	Concurrency	sysops/processes.go	processCache global without sync	sync.RWMutex
H7	Incomplete	secops/network.go	parseCertJSON is a placeholder stub	Implement or remove
H8	Incomplete	secops/endpoint.go	parseBitLockerJSON is a placeholder stub	Implement or remove
H9	Incomplete	sysops/logs.go:parseLogOutput	Only recognizes INFO level	Parse ERROR/WARN/DEBUG
H10	Security	common/sandbox_linux.go	No mount/pivot_root isolation	Implement or document limitation
H11	Type Mismatch	types/index.ts	GitStashEntry missing branch field	Add field
H12	Type Mismatch	types/index.ts	PingStats phantom lost_pct field	Compute in frontend or add to Go
H13	Type Mismatch	types/index.ts	ServiceInfo phantom port field	Remove from TS or add to Go
H14	Type Mismatch	types/index.ts	LogTimelinePoint missing total	Add field
H15	Type Mismatch	types/index.ts	LogSummary missing topMessage	Add field
H16	Architecture	app/DevOps.go	2107-line god file	Split by domain
H17	Design System	50+ frontend files	text-white hardcoded	text-[var(--color-text)]
H18	Design System	DevOps.tsx	bg-black hardcoded	bg-[var(--color-terminal-bg)]
H19	Duplication	common/alerts.go	GetAlerts()/ActiveAlerts() duplicates	Remove one
H20	Dead Code	.gitignore	Stale "hawkward" references	Update to opsforall
MEDIUM — Next Sprint
#	Area	Issue
M1	Security	devops/filebrowser.go:isPathSafe() blocks .. but not absolute paths
M2	Reliability	common/alerts.go:EmitAlert — event ring buffer silently drops oldest events
M3	Reliability	common/storage.go:batchWriteLoop — blocking send with no timeout
M4	Reliability	netops/monitoring.go:StartSyslogServer — no access control on UDP 5514
M5	Data Integrity	common/storage.go:Prune() — doesn't clean alerts/conversations tables
M6	Performance	PackageManagerTab.tsx — unbounded list (thousands of rows) without virtualization
M7	Performance	ProcessesTab.tsx — unbounded process list without virtualization
M8	Performance	LogsTab.tsx — 200 log entries without virtualization
M9	UX	netops/PortScanTab.tsx / TracerouteTab.tsx — error caught but no UI feedback
M10	Quality	DevOps.tsx:874, NetworkDesign.tsx:403 — console.log in production
M11	CSS	globals.css — duplicate .tabular-nums, orphaned spacing variables
M12	IPC	NetworkDesign.tsx — bypasses useBackend() for Save/Load IPC
M13	Performance	netops/OverviewTab.tsx — 6+ useQuery calls creating waterfall
M14	Duplication	netops/components.tsx + secops/components.tsx — duplicated SectionBriefing/MiniStat

Cross-Cutting Themes (Systemic Issues)
1. Command injection in incident response (4 CRITICAL in SecOps)
BlockIP, DisableAccount, SetFirewallRuleState concatenate user input into shell commands. The security module is the least sandboxed — response.go uses raw exec.Command while other modules use common.SandboxedCommand. Fix: validate all inputs (net.ParseIP, username regex) and sandbox all IR commands.

2. Data races in shared state (6+ CRITICAL/HIGH)
common.AllSeries() leaks *TimeSeries pointers outside RLock — callers race with PushMetric
NetOps model (lastCounters, lastCapture, etc.) has no mutex — collector goroutine races frontend reads
AIOps.effectiveModel is a package-level var with no synchronization
App.evaluateAndEmit() runs concurrently on alert tick + frontend trigger
common.activeTheme has no mutex
common.globalStorage singleton has no synchronization on 43+ call sites
3. Systematic error swallowing (15+ HIGH/MEDIUM)
The dominant pattern: result, _ := someFunc() or if err != nil { continue }. Callers get zero-value results indistinguishable from legitimate "no data." Worst offenders:

SecOps.audit.go — all 7 security data sources discard errors via _
DevOps.gitRun — all git errors silently discarded
App.DevOps.go:1867 — K8sScaleDeployment error discarded
App.Startup — storage init failure silently continues, DB features silently degrade
4. Linux sudo in GUI = guaranteed failure (2 CRITICAL in SysOps)
sudo requires a TTY which a Wails desktop app does not have. Affects: reboot, shutdown, sleep, hibernate, DNS flush, temp cleanup, pkg cache clean, system update on Linux. These are all dead code on Linux. Fix: use polkit/pkexec or D-Bus APIs.

5. Sandbox bypass across 3 modules (5 HIGH)
Git, Docker, and Kubernetes commands in DevOps, plus all incident response in SecOps, plus ARP/routing/WiFi/actions in NetOps — all use raw exec.Command instead of common.SandboxedCommand. The sandbox infrastructure exists but ~60% of OS command invocations skip it.

6. Business logic leaking into binding layer (3 HIGH)
DevOps.go is 2107 lines (2.5× next largest) with netstat parsing, tool detection, HTTP probes that belong in domain modules
SecOps.go has 300+ lines of security scoring/risk detection/recommendation logic inline
Dashboard.go generates user-facing diagnostic strings (presentation logic in Go)
7. Security stubs producing false negatives (5 HIGH in SecOps)
parseCertJSON — always returns empty (no TLS certs ever found)
parseBitLockerJSON — always shows not encrypted
parseServicesJSON — always returns empty
parseFailedLoginsJSON — always returns empty
network.go:getTLSCertificatesLinux — lists symlink filenames as cert subjects
8. Frontend issues (4 CRITICAL)
DevOps.tsx is ~1500 lines — same monolithic problem as the Go layer
// @ts-nocheck in ALL test files — tests provide zero type safety
3 zustand stores crammed into 1 file (settings, alerts, theme)
no-explicit-any globally disabled — defeats TypeScript's primary value

