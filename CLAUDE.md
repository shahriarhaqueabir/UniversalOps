# Project Instructions: Universal-Ops

## Project Overview
Universal-Ops is a high-performance native desktop operations platform built with Go (Wails v2) and React/TypeScript. It provides a technical utility suite for systems, network, and security operations with integrated local AI.

## Core Philosophy
- **100% Local**: No data leaves the machine.
- **Mechanics-First**: High-density technical telemetry.
- **Zero Telemetry**: No cloud dependencies or tracking.

## Technical Stack
- **GUI Framework**: Wails v2 (`github.com/wailsapp/wails/v2`)
- **Backend**: Go 1.26+
  - **System Metrics**: `github.com/shirou/gopsutil/v4`
  - **Network**: `github.com/miekg/dns`
  - **Storage**: `modernc.org/sqlite` (Database: `ops_core.db`)
- **Frontend**: React 19, TypeScript, Vite, Tailwind v4
  - **State**: Zustand
  - **Data Fetching**: TanStack Query (React Query)
  - **Icons**: Lucide React
  - **Charts**: Recharts
  - **Components**: Radix UI

## Development Workflow
- **Commands**:
  - `wails dev`: Run with hot-reload (frontend + backend).
  - `wails build`: Build production binary.
  - `go test ./internal/...`: Run backend tests.
  - `npm test --prefix cmd/opsforall-gui/frontend`: Run frontend tests.
  - `golangci-lint run ./...`: Backend linting.

## Project Structure
- `main.go`: Application entry point.
- `internal/`:
  - `app/`: Wails bound facades and API logic.
  - `common/`: Core services (Pipeline, Storage, Alerts, Sandbox).
  - `sysops/`, `netops/`, `secops/`, `devops/`, `aiops/`: Functional subsystems.
- `cmd/opsforall-gui/frontend/`: Frontend source code.

## Coding Standards
- **Go**: Idiomatic Go, error wrapping, consistent logging via `common.LogInfo`.
- **Frontend**: Functional components, memoization for performance, CSS variables for theming.
- **Database**: All persistence logic must use `ops_core.db`.
