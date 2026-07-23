# Audit Findings: Concurrency & Persistence

**Reviewer**: Concurrency Auditor Subagent
**Date**: 2026-07-21
**Confidence**: 0.90 (High)

## 1. Executive Summary
The concurrency model relies heavily on a global `storageLock` and asynchronous channels. While this prevents immediate crashes, it creates "cascading failures" where a stall in the OS layer (WMI/PowerShell) propagates through the Engine Loop and eventually causes the Persistence layer to drop data.

## 2. High-Severity Findings

### [HIGH] Autonomous Spike Re-entry Loop
- **Location**: `internal/common/engine.go:180`
- **Mechanism**: Spikes trigger `autonomousAudit` in a background goroutine without a "cooldown" or "active check."
- **Impact**: If a system is under sustained stress (e.g., 100% CPU for 30s), the engine will spawn dozens of parallel AI/Diagnostic tasks, likely deadlocking the WMI service and exhausting memory.
- **Remediation**: Implement a semaphore or state flag (`isAuditRunning`) to ensure only one diagnostic runs at a time.

### [HIGH] Global Storage Lock Contention
- **Location**: `internal/common/engine.go:102`
- **Mechanism**: `EngineLoop.Step` holds a `storageLock.RLock()` across the *entire* collection pass (Alerts + Metrics).
- **Impact**: If a collector hangs (WMI/Net), this lock is held indefinitely. Any UI thread attempting to write a setting or the `writerLoop` attempting to commit a batch will block, causing UI freezes.
- **Remediation**: Reduce the scope of the `storageLock` to only the actual `CaptureSnapshot` or `storage.Insert` calls.

### [HIGH] SQLite Commit Blocking
- **Location**: `internal/common/storage.go:441`
- **Mechanism**: The `writerLoop` commits batches on a 1s ticker.
- **Impact**: On slow disks or during long UI reads, `tx.Commit()` will hit the 5s busy timeout. This backs up `metricsCh`, causing `InsertMetric` to drop samples and log "DISK I/O IS SATURATED."
- **Remediation**: Use a more resilient batching strategy or detect "busy" states to delay commits without blocking the channel.

## 3. Medium-Severity Findings

### [MEDIUM] Snapshot Temporal Inconsistency
- **Location**: `internal/common/engine.go:160`
- **Mechanism**: `CaptureSnapshot` pulls metrics sequentially from the pipeline without a unified "point-in-time" lock.
- **Impact**: In high-velocity systems, the CPU value might be from T=0 while Memory is from T=100ms, leading to slightly skewed "spike" correlation.
- **Remediation**: Implement a `Snapshot()` method on the `DataPipeline` that returns all current values under a single lock.

### [MEDIUM] Noisy Heartbeat Writes
- **Location**: `internal/common/storage.go:392`
- **Mechanism**: `metricsLoggerLoop` writes a probe every 5s.
- **Impact**: Unnecessary SSD wear and database growth.
- **Remediation**: Disable by default; enable only via a "Debug Mode" setting.

## 4. Observations
- **WAL Mode**: WAL mode is correctly used, but its benefits are partially negated by the coarse-grained `storageLock` in the `EngineLoop`.
- **Batching**: 32-item metric batching is efficient, but the log-batching (50 items) might be too high for low-traffic systems, leading to delayed log persistence.
