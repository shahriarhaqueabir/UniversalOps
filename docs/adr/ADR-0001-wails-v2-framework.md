# ADR-0001: Wails v2 as Desktop Framework

**Date**: 2026-07-28
**Status**: Accepted
**Deciders**: @shahriarhaqueabir

## Context

UniversalOps needed a desktop GUI framework that could bridge a Go backend (for native system operations) with a modern web frontend. The requirements were: native performance for system telemetry, cross-platform support (Windows/Linux/macOS), access to OS-level APIs (WMI, gopsutil), and a modern component-based UI layer. Traditional options like Electron were rejected for their memory overhead, while pure native GUI toolkits lacked modern frontend ergonomics.

## Decision

Use **Wails v2** as the desktop application framework. Wails provides a Go runtime that binds directly to a WebView2 (Windows) or WebKit (Linux/macOS) frontend, enabling Go backend functions to be called directly from JavaScript with zero serialization overhead beyond JSON marshalling.

## Alternatives Considered

### Alternative 1: Electron + Node.js backend
- **Pros**: Mature ecosystem, extensive documentation, large community
- **Cons**: ~150MB+ binary size, high memory usage (~200MB idle), Node.js event loop not ideal for system-level operations
- **Why not**: Memory and binary size were unacceptable for a system monitoring tool that should run alongside other tools

### Alternative 2: Tauri (Rust backend)
- **Pros**: Small binary, strong security model, Rust performance
- **Cons**: Rust learning curve for team, less mature Windows API bindings, complex FFI for WMI/gopsutil integration
- **Why not**: The Go ecosystem (gopsutil, WMI libraries) provided faster time-to-market for system operations

### Alternative 3: Native Go GUI (Fyne / Gio)
- **Pros**: Pure Go, no webview dependency
- **Cons**: Limited component ecosystems, no CSS-based styling, smaller community
- **Why not**: Would require building all UI components from scratch — unacceptable for a feature-rich dashboard

## Consequences

- **Easier**: Go backend can directly call gopsutil, WMI, and SQLite without FFI layers. Frontend can use React's ecosystem (TanStack Query, Zustand, Recharts, Radix UI). Hot-reload development with `wails dev`.
- **Harder**: Windows requires WebView2 runtime (bundled in Windows 11, available via installer on Windows 10). Linux/macOS require WebKit GTK. Binary size is ~30-50MB (Go runtime + WebView2 host).
- **Risks**: Wails v2 has a smaller community than Electron. WebView2 version differences across Windows builds may cause rendering inconsistencies.