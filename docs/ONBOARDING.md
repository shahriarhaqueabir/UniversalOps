# Onboarding Guide: OpsForAll
 Operations Platform

## Overview
OpsForAll
 is a high-performance, local-first operations platform designed for system administrators, security analysts, and DevOps engineers. It provides real-time telemetry across five specialized layers (Sys, Net, Sec, Dev, AI) using a Go-based backend and a React-based GUI via Wails.

## Tech Stack
| Layer | Technology | Version |
|-------|-----------|---------|
| Backend | Go | 1.26.5 |
| Frontend | React + TypeScript + Vite | v19 / Tailwind v4 |
| Bridge | Wails v2 | v2.13.0 |
| Database | SQLite | - |
| Monitoring | gopsutil/v4 | v4.x |

## Architecture
OpsForAll
 follows a **Modular Monolith** architecture with a **Facade-Bridge** pattern:
1. **Core Modules** (`internal/*ops/`): Pure domain logic, platform-specific implementations, and high-performance collection.
2. **Common Layer** (`internal/common/`): Shared infrastructure including the Time-Series Data Pipeline, Alert Engine, and Persistent Storage.
3. **Facade Layer** (`internal/app/`): Wails-bound structs that map internal domain models to frontend-safe API types.
4. **GUI Layer** (`cmd/opsforall-gui/frontend/`): React application using Tailwind v4 for a premium, high-density dashboard.

## Key Entry Points
- **Backend Entry**: `main.go` — Wails bootstrap and module binding.
- **Frontend Entry**: `cmd/opsforall-gui/frontend/src/main.tsx` — React root.
- **Telemetry Loop**: `internal/app/App.go` (`collectAndEmit`) — Centralized polling logic.
- **API Definition**: `internal/app/Types.go` — Source of truth for frontend/backend data contracts.

## Directory Map
- `internal/app/` → Wails Facades & Type Mappings
- `internal/sysops/` → CPU, RAM, Disk, and Process monitoring
- `internal/netops/` → DNS, Ping (Jitter), Traceroute, and Connections
- `internal/secops/` → Firewall (High-risk flagging), Defender, and Identity audits
- `internal/devops/` → Terminal wrapper, Breadcrumb File Browser, and Services
- `internal/aiops/` → Local AI analysis via Ollama, Anomaly Detection, and Reporting
- `internal/common/` → Cross-cutting concerns (Metrics Pipeline, Alerts, Storage)

## Conventions
- **Naming**: PascalCase for Go exports, camelCase for internal Go and all TypeScript.
- **Error Handling**: Standard Go wrapping: `fmt.Errorf("context: %w", err)`.
- **Testing**: Module-local unit tests (`*_test.go`) with race detection verification.
- **Frontend**: Component-based architecture with Lucide icons and Recharts.

## Common Tasks
- **Build**: `wails build` (compiles frontend and embeds in Go binary)
- **Dev (Simultaneous)**: `wails dev`
- **Run Tests**: `go test ./...`
- **Check Races**: `go test -race ./...` (Requires MinGW on Windows)
