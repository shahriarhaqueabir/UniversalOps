# Concurrency Audit — UniversalOps Go Backend

**Date:** 2026-08-16
**Scope:** `internal/` Go backend — goroutine lifecycle, channel/close discipline, mutex correctness, context propagation, error-wrapping, SQLite concurrency, collector/pipeline/event-bus hot paths, `RecoverPanic` coverage.
**Constraint:** Review only — **no code changes made.**
**Commit reviewed:** `dd68210a` (v1.6.3) + uncommitted `internal/common/logger.go` (slog migration).

---

## 1. Baseline (all green)

| Check | Command | Result | Time |
| --- | --- | --- | --- |
| Build | `go build ./...` | ✅ OK | 5.1s |
| Vet | `go vet ./...` | ✅ OK | 2.8s |
| Race tests | `go test -race ./internal/... -count=1 -timeout 180s` | ✅ 7/7 packages pass, **zero races detected** | 114.3s |

**Caveat:** the race detector only exercises paths covered by the test suite. Several findings below live in paths **not** exercised by tests (logger re-init, `UpdateStorageConfig`, live-output consumer abandonment, WMI timeout leaks, `autonomousAudit` timeout). These are identified by static analysis and are real, but were not observable in the green run.

---

## 2. Ranked Findings

Severity × Confidence. **No HIGH-severity findings** — the codebase is in good shape; the race detector is clean and the previously-flagged critical issues are fixed. Findings are MEDIUM (should fix) and LOW (hygiene/defense-in-depth).

### MEDIUM

#### M1. Unsynchronized logger globals — data race on `logger` / `logFile`

- **Evidence:** `internal/common/logger.go:13-14` (`logFile *os.File`, `logger *slog.Logger` package globals), read in `getLogger()` (`:30-34`), written in `InitLogger()` (`:76-118`), read+closed in `CloseLogger()` (`:118-123`). No mutex/atomic anywhere.
- **Trigger:** `App.UpdateLogsConfig` (`internal/app/App.go:693-710`) calls `CloseLogger()` then `InitLogger()` on the UI thread; `Shutdown` calls `CloseLogger()` (`App.go:358`). Meanwhile background goroutines (engine loop, `persistAlertsAsync`, `OnAlertResolved`, event-bus subscribers) call `LogInfo`/`LogWarn` continuously → concurrent read/write of the globals.
- **Impact:** genuine data race (Go memory-model UB). Practical consequence is usually benign (a lost log line, or a panic recovered by `RecoverPanic`), but a torn read can yield a nil logger deref or a write to a closed file. Not caught by tests because no test re-inits the logger concurrently with logging.
- **Recommendation:** guard the globals with a `sync.RWMutex` (or `atomic.Pointer[slog.Logger]`), and set `logFile = nil` after close. Consider making `multiHandler` hold a stable handler set so re-init swaps atomically.

#### M2. Event-bus hot path blocks on a synchronous SQLite write

- **Evidence:** `internal/common/events.go:119-125` — `Emit` snapshots handlers under lock then dispatches **synchronously** (`for _, h := range handlers { h(evt) }`, no `go`). `internal/app/App.go:186-195` — the sole subscriber calls `s.InsertEvent(evt, nil)` (a direct DB write) on every event.
- **Impact:** every `eventBus.Emit` (metrics ticks, alerts, engine events) performs a synchronous SQLite INSERT on the emitter's critical path. The engine loop stalls on DB I/O; under contention (batch writes, prune) this adds latency to the 2s tick. Not a correctness bug — a latency/backpressure concern.
- **Recommendation:** dispatch handlers asynchronously (bounded worker or buffered channel) or make the persistence subscriber write via the storage writer loop instead of a direct synchronous `InsertEvent`. The comment at `App.go:192-194` already notes the frontend no longer needs the event — the DB write is the only reason this handler exists.

#### M3. Live-output shell functions: unbuffered blocking send with no cancellation

- **Evidence:** `internal/devops/shell.go` — `RunCommandWithLiveOutput`, `RunPowerShellWithLiveOutput`, and `RunGitBashWithLiveOutput` each spawn two reader goroutines doing `output <- line` on an **unbuffered** channel with no select/cancellation.
- **Impact:** if the consumer of `output` stops reading (frontend disconnect, caller abandons the stream), the reader goroutines block forever on the send; `c.Wait()` returns, `wg.Wait()` never completes, and the function never returns — hanging the caller (UI thread).
- **Recommendation:** add a done/quit channel and `select { case output <- line: case <-done: return }`, or buffer the channel, or close `output` from the consumer side. Ensure the consumer always drains or signals.

#### M4. `RunNetworkDiagnostics` has no overall timeout; `pingICMP` has no read deadline

- **Evidence:** `internal/netops/workflows.go:22-118` — 6 goroutines + `wg.Wait()` with no `context.WithTimeout`/`time.After`. `internal/netops/ping.go:63+` — `pingICMP` uses `icmp.ListenPacket` with **no read deadline** (no timeout tokens in the file).
- **Impact:** on Windows (the target platform) `Ping` routes to `pingExec` → `ping.exe`, which has built-in timeouts, so the practical hang risk is low there. On Linux/macOS builds, a silently-dropped ICMP packet blocks `conn.Read` forever → `wg.Wait()` never returns → caller hangs.
- **Recommendation:** wrap the whole diagnostic in a `context.WithTimeout` and pass it down; add a read deadline to `pingICMP`.

#### M5. `persistAlertsAsync` missing `defer tx.Rollback()` and ignores insert errors

- **Evidence:** `internal/app/App.go:965-991` — `tx, err := storage.Begin()`; loop calls `storage.InsertAlert(..., tx)` **ignoring the returned error**; `tx.Commit()` error is logged but the tx is never rolled back. This is the **only** transaction in the codebase missing the `defer tx.Rollback()` pattern (cf. `insertMetricsBatch`/`insertLogsBatch` in `storage.go`).
- **Impact:** on a failed commit (or a mid-loop insert failure that aborts the tx), the connection can be left in a transaction state. With `SetMaxOpenConns(2)`, a leaked transaction slot is expensive. Runs on every alert emit.
- **Recommendation:** add `defer tx.Rollback()` and check `InsertAlert` errors (abort the batch on failure).

#### M6. `autonomousAudit` timeout doesn't cancel the workflow; `auditRunning` released early

- **Evidence:** `internal/common/engine.go:437-490` — on `case <-time.After(30 * time.Second)` (`:477`) the function returns, the deferred `auditRunning = false` runs (`:455`), but the inner goroutine calling `e.invoker.TriggerWorkflow(workflowID)` (`:466-469`) is **still running**.
- **Impact:** (a) the 30s timeout protects the caller but does not cancel the underlying Ollama workflow — it keeps executing; (b) `auditRunning` is cleared while the old audit is still live, so a new audit can start concurrently, defeating the "no concurrent audits" guard that fixed the prior spike-re-entry finding.
- **Recommendation:** track the running workflow (e.g., a cancelable context or a `sync.Once`/generation counter) so the guard is not released until the underlying work actually finishes, or cancel the workflow on timeout.

### LOW

#### L1. `metricsLoggerLoop` not tracked in a WaitGroup

- **Evidence:** `internal/common/storage.go:167` — `go s.metricsLoggerLoop()` is started outside `writerWg`/`pruneWg`; `Close()` (`:187-196`) waits only `pruneWg` and `writerWg`.
- **Impact:** benign — the loop reads `closeCh` and `InsertMetric` checks the closed flag under `mu`. But `Close()` can return while the loop is mid-iteration.
- **Recommendation:** add it to a WaitGroup for a clean shutdown guarantee.

#### L2. `OnAlertResolved` callback invoked while holding `ae.mu`

- **Evidence:** `internal/common/alerts.go:263-264` (`Evaluate` takes `ae.mu.Lock()`) and `:329-330` (`ae.OnAlertResolved(ae.alerts[idx])` inside the locked region).
- **Impact:** safe today only because `App.go:173-183` wires it to spawn an async goroutine. Any future synchronous handler that re-enters `AlertEngine` deadlocks.
- **Recommendation:** snapshot the resolved alerts and invoke the callback after releasing the lock.

#### L3. Trust worker goroutine has no quit channel

- **Evidence:** `internal/sysops/processes.go:55-62` — `trustQueue` (cap 500) + `trustWorkerOnce.Do(func() { go func() { for path := range trustQueue {...} }() })`. The queue is never closed.
- **Impact:** the worker leaks on shutdown (idle-blocked on the channel read — benign, but untracked).
- **Recommendation:** close `trustQueue` (or add a quit channel) during shutdown.

#### L4. `CapabilityRegistry.Shutdown()` never called

- **Evidence:** `internal/common/capability.go` — `Shutdown()` (`:176-183`) safely closes `stopCh`, but no call site exists (grep: no `.Shutdown()` on the registry). The background refresh goroutine (`:130-168`) checks `stopCh` between tools.
- **Impact:** `stopCh` is never closed in practice; the goroutine is finite (loops through `allIds` then exits), so no leak — but the cancellation path is dead code.
- **Recommendation:** call `Shutdown()` from `App.Shutdown`, or remove the dead path.

#### L5. WMI timeout doesn't cancel the underlying `wmi.Query` goroutine

- **Evidence:** `internal/common/wmi.go:20-39` — `WMIQueryWithTimeout`/`WMIQueryNamespaceWithTimeout` spawn a goroutine and `select` on `ctx.Done` vs `errChan`. The goroutine running `wmi.Query` is not cancellable.
- **Impact:** the caller is protected, but if the WMI provider truly hangs, each timed-out call leaks a goroutine until the query returns (or forever).
- **Recommendation:** accept as a known limitation of the `wmi` library, or document it; consider a bounded worker pool for WMI calls.

#### L6. `PingMultiTarget` — unbounded goroutines, no `RecoverPanic`

- **Evidence:** `internal/netops/ping_multi.go:72-80` — one goroutine per target with `defer wg.Done()` but **no** `defer common.RecoverPanic()` (contrast `portscan.go:92-118` which has it).
- **Impact:** if `pingOneTarget` panics, `wg.Done` is skipped and `wg.Wait()` hangs forever. Target count is small in practice, but the inconsistency is a latent hang.
- **Recommendation:** add `defer common.RecoverPanic()` (and consider a semaphore for large target lists).

#### L7. `metrics_exporter.go` — dead code behind build tag + stale version

- **Evidence:** `internal/common/metrics_exporter.go:1` (`//go:build prometheus` — not built by default) and `:45` (`"version": "1.4.0"` — current is v1.6.3). `metricsPort` (`:19`) read without sync (`:33`, `:94`).
- **Impact:** dead code unless built with the tag; the `/health` endpoint would report a stale version.
- **Recommendation:** update the version constant or delete the file if the exporter is unused.

#### L8. `TrackedCommand` and `AutoExpireWG` are dead code

- **Evidence:** `internal/common/sandbox.go:272-285` (`TrackedCommand` — defined, never called; the comment admits it can't register before `Start`), `internal/secops/response.go:19` (`AutoExpireWG` — defined, never `Add`/`Done`/`Wait`).
- **Impact:** misleading — suggests a process-tracking guarantee that doesn't exist. `SandboxedCmd.Start/Wait` (`sandbox.go:295-304`) do register/unregister correctly.
- **Recommendation:** remove or wire up; if raw `exec.Cmd` paths need shutdown cleanup, register them explicitly.

#### L9. `TimeSeries.Values()` / `GetStats()` take the write lock for reads

- **Evidence:** `internal/common/timeseries.go:57-70` (`Values` uses `ts.mu.Lock()` to cache `lastValuesView`) and `:90+` (`GetStats` uses `ts.mu.Lock()` to cache `statsCache`).
- **Impact:** readers serialize with writers and with each other on the metrics hot path. Minor — the critical sections are short.
- **Recommendation:** use `RLock` for reads and accept cache invalidation races, or guard the cache with a separate mutex.

#### L10. `multiHandler.Handle` returns on first handler error

- **Evidence:** `internal/common/logger.go:52-60` — if the file handler errors, the console handler is skipped for that record.
- **Impact:** a file write failure silently drops console output for that record.
- **Recommendation:** log the error and continue to the remaining handlers.

#### L11. `RefreshBatch` holds the registry write lock during binary execution

- **Evidence:** `internal/common/capability.go:186-190` — `RefreshBatch` takes `r.mu.Lock()` for the whole batch, including `detectVersion` which runs external binaries with 2s timeouts (`:260+`).
- **Impact:** bounded (~2s per tool) but a UI-thread call (`VerifyCapability` → `RefreshBatch`) can block on a global lock while executing a binary. Improvement over the prior 30s finding, but still a synchronous exec under lock.
- **Recommendation:** run version detection outside the lock (per-tool lock, as the background scan already does).

---

## 3. Verified-Good Patterns (no action needed)

- **Race detector:** clean across all 7 packages.
- **`storage.go`:** `closeOnce` (`:69`, `:187`), `writerWg`/`pruneWg` (`:62-63`), `defer tx.Rollback()` in `insertMetricsBatch`/`insertLogsBatch`, batched `Prune` with 100ms yield (`:1130-1180`), `getDB(tx)` helper (`:1356`), `InsertAlert` lock-avoidance comment (`:1370`).
- **`netops/discovery.go` + `portscan.go`:** bounded semaphores (64), `wg.Add/Done`, distinct-index writes — no shared-write races.
- **`AIOps.go:420-470`:** 5-goroutine concurrent history load, each writing a distinct local variable, all with `RecoverPanic`.
- **`capability.go`:** `atomic.Bool` refresh guard + `stopCh` select between tools.
- **`scheduler.go`:** `Stop(timeout)` with WaitGroup; exponential backoff.
- **`workflow.go`:** clean `RWMutex`-guarded map.
- **`ollama.go`:** `defaultClientMu` + `modelMu` properly locked.
- **`collectors.go`:** `speedCacheMu` properly locked.
- **`engine.go`:** `storageLock.RLock()` scoped to capture phase only; `auditMu`/`auditCooldown` guard (with M6 caveat).

## 4. Prior-Finding Verification (from `docs/audit/findings-concurrency.md`)

| Prior finding | Status |
| --- | --- |
| Autonomous spike re-entry | ✅ FIXED (`auditRunning` + cooldown) — residual in M6 |
| Global storage-lock contention | ✅ FIXED (RLock scoped to capture) |
| Storage drain-loop `break` bug | ✅ FIXED (labeled breaks) |
| `StopMonitoring` never closes done channel | ✅ FIXED |
| Process snapshot stalls engine | ✅ FIXED (decoupled worker) |
| Unprotected WMI calls | ✅ FIXED (timeouts present) — residual in L5 |
| Unvalidated `ReadTextFile` path | ✅ FIXED (`filepath.Abs` + `IsDir`) |
| Missing `tx.Rollback` in batch inserts | ✅ FIXED — but `persistAlertsAsync` (M5) still lacks it |

---

## 5. Suggested Fix Order

1. **M1** (logger race) — correctness; small change.
2. **M3** (live-output blocking send) — hang risk on UI thread.
3. **M6** (audit guard bypass) — re-introduces a previously-fixed class of bug.
4. **M2** (event-bus hot path) — latency on the 2s tick.
5. **M5** (tx rollback) — connection-slot hygiene.
6. **M4** (diagnostics timeout) — platform-dependent hang.
7. L1–L11 as hygiene backlog.
