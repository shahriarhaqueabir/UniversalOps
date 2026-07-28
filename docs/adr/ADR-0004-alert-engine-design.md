# ADR-0004: In-Memory Alert Engine with SQLite Persistence

**Date**: 2026-07-28
**Status**: Accepted
**Deciders**: @shahriarhaqueabir

## Context

UniversalOps needed an alerting system that could evaluate rules against live telemetry data, detect state transitions (normal → warning → critical), suppress duplicate alerts, detect flapping (rapid state oscillations), and correlate related alerts. The system also needed to persist alert history across restarts so users don't lose context when the application is closed and reopened.

## Decision

Implement a **dual-layer alert system**: an in-memory alert engine (`AlertEngine`) for real-time evaluation, deduplication, and flap detection, backed by SQLite persistence for historical records. The engine maintains an in-memory map of active alert keys for O(1) lookup during evaluation, while a background goroutine asynchronously writes new alerts and resolution events to SQLite. On startup, persisted alerts are restored into the in-memory state so the UI reflects historical alerts immediately.

## Alternatives Considered

### Alternative 1: Fully SQLite-based alerting
- **Pros**: Single source of truth, no sync issues, queryable by external tools
- **Cons**: Every evaluation requires a database round-trip — adds ~1-5ms latency per rule. Flap detection requires reading recent history, adding more queries. SQLite write contention under high alert volume.
- **Why not**: The latency of per-evaluation database queries would slow down the 3-second evaluation cycle, especially with many rules

### Alternative 2: External monitoring service (Prometheus + Alertmanager)
- **Pros**: Battle-tested, scalable, rich alert routing
- **Cons**: Requires separate server process, violates local-first principle, over-engineered for a single-machine desktop app
- **Why not**: Completely incompatible with the "100% local, zero infrastructure" philosophy

### Alternative 3: In-memory only (no persistence)
- **Pros**: Simplest implementation, maximum performance
- **Cons**: All alert history lost on restart — users lose context of what was happening before the app was closed
- **Why not**: This was the original design and proved inadequate — users expected alert history to survive restarts

## Consequences

- **Easier**: O(1) alert key lookup during evaluation. Async persistence doesn't block the evaluation loop. Flap detection works against in-memory state. Restore from SQLite on startup provides continuity.
- **Harder**: Two sources of truth must be kept in sync (in-memory + SQLite). Restore logic must handle edge cases (partially written records, schema migrations). Async writes mean a crash between evaluation and persistence could lose the most recent alert.
- **Risks**: Memory growth if alert history accumulates without bounds. Mitigated by a configurable maximum active alert count and periodic pruning of resolved alerts older than a threshold.