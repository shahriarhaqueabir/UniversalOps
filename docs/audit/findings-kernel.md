# Audit Findings: Kernel Specialist (OS Ingestion)

**Reviewer**: Kernel Specialist Subagent
**Date**: 2026-07-21
**Confidence**: 0.95 (High)

## 1. Executive Summary
The OS ingestion layer is functional but highly susceptible to "blocking hangs" caused by synchronous WMI and PowerShell calls. In restricted Windows environments or during system stress, these calls will cause the Engine Loop to miss ticks and the Wails UI to freeze.

## 2. High-Severity Findings

### [CRITICAL] Synchronous Lock-Bound OS Calls
- **Location**: `internal/sysops/processes.go:180` (`getTrustInfo`)
- **Mechanism**: The `UpdateProcessSnapshot` loop calls `getTrustInfo` synchronously for every new process. Each call spawns a PowerShell process with a 3s timeout.
- **Impact**: If 10 new processes appear, the Engine Loop hangs for up to 30 seconds. Since the Engine Loop is the source for Wails events, the UI will freeze completely.
- **Remediation**: Move `getTrustInfo` to an asynchronous worker queue or perform it *after* the critical telemetry snapshot is released to the UI.

### [CRITICAL] Missing WMI Timeouts
- **Location**: `internal/sysops/gpu.go:42, 153`
- **Mechanism**: Calls `wmi.Query` and `wmi.QueryNamespace` without a Go `context`. 
- **Impact**: If the WMI service or `LibreHardwareMonitor` provider hangs (common on Windows), the Go routine will never return, causing a permanent telemetry "flatline" or UI hang.
- **Remediation**: Use `wmi.QueryWithContext` with a 2-5s timeout.

## 3. Medium-Severity Findings

### [HIGH] PID Recycling Cache Race
- **Location**: `internal/sysops/processes.go:63-100`
- **Mechanism**: Concurrent access to `procCache` and `lastSnapshot` without persistent locking across the entire snapshot pass.
- **Impact**: Inconsistent telemetry data or map-access panics if manual refresh and engine loop tick overlap.
- **Remediation**: Wrap the entire snapshot logic in a single mutex or use a copy-on-write strategy.

### [HIGH] Capability Registry Registry Lock Contention
- **Location**: `internal/common/capability.go:137`
- **Mechanism**: The background `RefreshBatch` holds the global registry lock while performing sequential `--version` binary executions.
- **Impact**: Calls to `IsAvailable()` from the UI will block for up to 30s during startup.
- **Remediation**: Use a read-lock for lookup and only write-lock when updating individual tool status.

## 4. Observations
- **Observer Effect**: High frequency of PowerShell spawning (`gpu.go`, `capability.go`) adds measurable CPU load to the system being monitored.
- **Silent Failures**: WMI failures return 0.0 metrics without error propagation, misleading AI analysis.
