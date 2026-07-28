# ADR-0002: SQLite WAL for Local Storage

**Date**: 2026-07-28
**Status**: Accepted
**Deciders**: @shahriarhaqueabir

## Context

UniversalOps collects high-frequency system telemetry (CPU, memory, disk, network metrics every 3 seconds by default) and needs to persist alert history, metric snapshots, workflow definitions, and configuration. The storage layer must be local-only (no cloud), support concurrent reads from the UI while writes are in progress, and require zero infrastructure setup from the user.

## Decision

Use **SQLite with Write-Ahead Logging (WAL) mode** via `modernc.org/sqlite` (a pure Go SQLite implementation, no CGo dependency). WAL mode allows concurrent reads and writes without blocking — the UI can query historical data while the data pipeline writes new metrics.

## Alternatives Considered

### Alternative 1: BoltDB / bbolt
- **Pros**: Embedded, pure Go, no CGo, simple key-value API
- **Cons**: No SQL query capability, no indexing, manual aggregation logic required for time-series queries
- **Why not**: The need for complex queries (time-range filtering, aggregation, JOINs between alerts and metrics) made a key-value store impractical

### Alternative 2: BadgerDB
- **Pros**: High-performance LSM tree, pure Go
- **Cons**: No SQL, higher memory usage, more complex backup/restore
- **Why not**: Same query limitations as BoltDB, plus unnecessary complexity for the workload

### Alternative 3: PostgreSQL (embedded)
- **Pros**: Full SQL, mature, excellent query capabilities
- **Cons**: Requires separate server process, ~50MB+ install, not zero-infrastructure
- **Why not**: Violates the "zero infrastructure" requirement — users should not need to install a database server

## Consequences

- **Easier**: SQL queries for time-range filtering, aggregation, and JOINs. WAL mode enables concurrent read/write without locks. Pure Go implementation avoids CGo cross-compilation issues. Single file (`universalops.db`) is easy to backup, inspect, or delete.
- **Harder**: SQLite is not designed for cluster-level or multi-writer scenarios. Write-ahead logging increases disk I/O under heavy write loads. No built-in replication.
- **Risks**: Under very high-frequency writes (>100 writes/second), WAL file can grow large without proper checkpointing. Mitigated by periodic `PRAGMA wal_checkpoint(TRUNCAT)` in the data pipeline.