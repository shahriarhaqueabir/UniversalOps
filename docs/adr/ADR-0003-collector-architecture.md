# ADR-0003: Goroutine-Per-Collector Architecture

**Date**: 2026-07-28
**Status**: Accepted
**Deciders**: @shahriarhaqueabir

## Context

UniversalOps monitors multiple system domains (CPU, memory, disk, network, processes, services, sensors) simultaneously. Each domain requires independent data collection at potentially different intervals. A single-threaded collection loop would introduce latency — a slow disk query would delay CPU metric updates. The architecture needed to support independent collection cycles, graceful degradation (one collector failing shouldn't affect others), and backpressure handling.

## Decision

Use a **goroutine-per-collector** architecture where each system domain runs in its own goroutine with independent timing, exponential backoff on failure, and a shared sharded data store for results. Collectors communicate results via a channel-based pipeline that batches writes to SQLite.

## Alternatives Considered

### Alternative 1: Single-threaded polling loop
- **Pros**: Simple implementation, no concurrency concerns, deterministic ordering
- **Cons**: Slow collectors block fast ones, no isolation between domains, hard to add per-collector retry logic
- **Why not**: A slow WMI query (common on Windows) would delay all other metric updates

### Alternative 2: Worker pool with job queue
- **Pros**: Bounded concurrency, controlled resource usage
- **Cons**: Requires job scheduling logic, queue management, priority handling
- **Why not**: Over-engineered for the use case — collectors are long-running, not bursty jobs. Goroutine-per-collector is simpler and equally effective.

### Alternative 3: Timer-based event loop with async callbacks
- **Pros**: Single-threaded with async I/O where possible
- **Cons**: gopsutil and WMI calls are synchronous — they block the calling thread. Async wrappers would add complexity without benefit.
- **Why not**: The underlying system APIs are synchronous; async wrappers would just move the blocking to a goroutine anyway.

## Consequences

- **Easier**: Each collector is independently testable. Failure isolation — a crash in the disk collector doesn't affect CPU metrics. Per-collector backoff and retry. Easy to add new collectors without modifying existing ones.
- **Harder**: Must handle concurrent access to shared state (sharded data stores). Resource usage scales with number of collectors (~1 goroutine per collector, ~10 total). Need context cancellation for graceful shutdown.
- **Risks**: Goroutine leaks if collectors don't properly handle context cancellation. Mitigated by `sync.WaitGroup` tracking and a shutdown timeout in the engine loop.