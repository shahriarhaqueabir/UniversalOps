# NET-001: Bandwidth Sparklines

## Context
The goal is to provide real-time network bandwidth monitoring with visual sparklines in the NetOps view.

## Current State
- `internal/netops/interfaces.go` has rate calculation logic but uses `time.Sleep`.
- `InterfaceInfo` struct has `RXHistory` and `TXHistory` fields.
- `mergeInterfaceBandwidthHistory` helper exists to persist history across updates.
- No sparklines are rendered in `internal/netops/view.go`.

## Proposed Change
1. Refactor `GetInterfaces` to accept previous counters and return new counters + rates, avoiding `time.Sleep`.
2. Update `Model` in `netops` to store last counters.
3. Use `common.RenderSparkline` in `netops/view.go` to display trends.
4. Ensure the refresh interval (default 1s or 2s) is respected.

## Risks
- Blocking the TUI if metrics collection hangs.
- Counter wrap-around (handled by `counterDelta`).
- High CPU usage if polling too frequently.
