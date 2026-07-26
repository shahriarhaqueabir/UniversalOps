# Developing UniversalOps

> Deep development guide for contributors. Covers git workflow, environment setup, debugging, project tree, and PR best practices.

---

## 📋 Table of Contents

1. [Git & Branch Strategy](#1-git--branch-strategy)
2. [Environment Setup](#2-environment-setup)
3. [Project Tree (Annotated)](#3-project-tree-annotated)
4. [Development Workflow](#4-development-workflow)
5. [Debugging & Profiling](#5-debugging--profiling)
6. [PR Best Practices](#6-pr-best-practices)
7. [Environment Variables Reference](#7-environment-variables-reference)

---

## 1. Git & Branch Strategy

### Branch Naming

Use descriptive, hyphenated names with a type prefix:

| Pattern | Example |
|---------|---------|
| `feat/<description>` | `feat/gpu-monitoring` |
| `fix/<description>` | `fix/dns-timeout` |
| `refactor/<description>` | `refactor/collector-pipeline` |
| `docs/<description>` | `docs/api-contracts` |
| `chore/<description>` | `chore/update-deps` |

### Commit Messages

We follow **Conventional Commits** with optional emoji prefixes for visual scanning:

| Type | Emoji | Example |
|------|-------|---------|
| `feat` | ✨ | `✨ feat(sysops): add per-core temperature tracking` |
| `fix` | 🐛 | `🐛 fix(netops): handle DNS timeout gracefully` |
| `refactor` | ♻️ | `♻️ refactor(common): extract ring buffer to shared type` |
| `docs` | 📝 | `📝 docs(readme): update installation steps` |
| `test` | 🧪 | `🧪 test(aiops): add anomaly detection edge cases` |
| `chore` | 🔧 | `🔧 chore(deps): bump gopsutil to v4.24` |
| `ci` | 👷 | `👷 ci: add Windows E2E test workflow` |
| `perf` | ⚡ | `⚡ perf(app): memoize expensive chart re-renders` |
| `style` | 🎨 | `🎨 style(frontend): align grid cards on mobile` |

Format:
```
<emoji> <type>(<scope>): <description>

[optional body]
[optional footer]
```

### Pull Request Flow

1. Create a branch from `main`
2. Implement changes with incremental commits
3. Run full test suite locally
4. Push and open PR with the [PR template](../.github/PULL_REQUEST_TEMPLATE.md)
5. Ensure CI passes on all three OS targets
6. Squash-merge to `main`

---

## 2. Environment Setup

### Prerequisites

| Tool | Version | Install |
|------|---------|---------|
| Go | 1.26+ | [go.dev](https://go.dev/dl/) |
| Node.js | 20+ (LTS) | [nodejs.org](https://nodejs.org/) |
| Wails CLI | v2 | `go install github.com/wailsapp/wails/v2/cmd/wails@latest` |
| MinGW (Windows) | — | Via [Chocolatey](https://chocolatey.org/): `choco install mingw` |
| golangci-lint | latest | `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest` |

### macOS / Linux Extras

| Tool | Why |
|------|-----|
| GTK 3 / WebKit2GTK | Required by Wails on Linux |
| Xcode Command Line Tools | Required by Wails on macOS |
| `setcap` (Linux) | Required for unprivileged `ping` in NetOps |

### Quick Start

```bash
# Clone and enter
git clone https://github.com/shahriarhaqueabir/UniversalOps.git
cd UniversalOps

# Install frontend dependencies
cd cmd/opsforall-gui/frontend
npm install
cd ../..

# Verify Go build
go build ./...

# Start dev server (hot-reload for Go + React)
wails dev
```

The dev server runs at `http://localhost:5173` and proxies API calls to the Go backend.

---

## 3. Project Tree (Annotated)

```
UniversalOps/
│
├── main.go                         # Wails app entry point — bootstrap, binding, lifecycle
├── main_stub.go                    # Platform stub for cross-compilation
├── main_windows.go                 # Windows-specific startup (hide console, etc.)
│
├── internal/
│   ├── app/                        # Wails bound facades (API layer)
│   │   ├── App.go                  #   Core app — engine lifecycle, metric emission
│   │   ├── SysOps.go               #   System operations facade
│   │   ├── NetOps.go               #   Network operations facade
│   │   ├── SecOps.go               #   Security operations facade
│   │   ├── DevOps.go               #   DevOps automation facade
│   │   ├── AIOps.go                #   AI/Ollama operations facade
│   │   ├── Dashboard.go            #   Dashboard aggregation facade
│   │   ├── Alerts.go               #   Alert management facade
│   │   ├── Pipeline.go             #   Telemetry pipeline facade
│   │   ├── Logs.go                 #   Log viewer facade
│   │   ├── Events.go               #   Event timeline facade
│   │   ├── Timeline.go             #   Event analysis facade
│   │   ├── Collectors.go           #   Collector configuration facade
│   │   ├── Knowledge.go            #   Knowledge base facade
│   │   ├── Reports.go              #   Report generation facade
│   │   ├── Workflow.go             #   Workflow engine facade
│   │   ├── Environment.go          #   Environment detection facade
│   │   ├── Types.go                #   API type definitions (frontend contract)
│   │   └── *_test.go               #   Facade unit tests
│   │
│   ├── common/                     # Core services & shared infrastructure
│   │   ├── engine.go               #   Main engine loop (parallel collector lanes)
│   │   ├── collector.go            #   Collector interface & registry
│   │   ├── pipeline.go             #   Time-series data pipeline (ring buffer)
│   │   ├── storage.go              #   SQLite storage layer (async batch writer)
│   │   ├── alerts.go               #   Alert rule engine & evaluation
│   │   ├── sandbox.go              #   Command sandbox for secure execution
│   │   ├── capability.go           #   Capability registry & auto-detection
│   │   ├── logging.go              #   Structured logging helpers
│   │   ├── config.go               #   Configuration management
│   │   └── ...
│   │
│   ├── sysops/                     # System operations (CPU, Memory, Disk, GPU)
│   ├── netops/                     # Network operations (DNS, Ping, Ports, Connections)
│   ├── secops/                     # Security operations (Firewall, Defender, Identity)
│   ├── devops/                     # DevOps (Terminal, File Browser, Services)
│   ├── aiops/                      # AI operations (Ollama, Anomaly Detection, Reporting)
│   │   └── mcp/                    #   In-process MCP tools (24 tools, no external transport)
│   └── tools/                      # Shared utility tools
│
├── cmd/
│   └── opsforall-gui/
│       └── frontend/               # React 19 + TypeScript + Vite + Tailwind v4
│           ├── src/
│           │   ├── main.tsx        #   React entry point
│           │   ├── App.tsx         #   Root component with router
│           │   ├── components/     #   Reusable UI components (Radix UI, Recharts)
│           │   ├── pages/          #   Page-level components per module
│           │   ├── stores/         #   Zustand state stores
│           │   ├── hooks/          #   Custom React hooks
│           │   ├── types/          #   TypeScript type definitions
│           │   └── lib/            #   Utility functions
│           ├── index.html
│           ├── vite.config.ts
│           └── package.json
│
├── docs/                           # Documentation (see docs/readme.md)
├── scripts/                        # Build & release scripts
├── tests/e2e/                      # End-to-end tests (Python/Playwright)
├── build/                          # Build artifacts (output directory)
└── data/                           # Runtime data (baselines, modelfiles)
```

---

## 4. Development Workflow

### Typical Change Cycle

1. **Branch**: `git checkout -b feat/my-feature`
2. **Implement**: Make changes, commit incrementally
3. **Build**: `go build ./...` — catch compilation errors early
4. **Lint**: `golangci-lint run ./...` — follow Go idioms
5. **Backend Tests**: `go test ./internal/... -count=1 -timeout 120s`
6. **Frontend Build**: `cd cmd/opsforall-gui/frontend && npx tsc --noEmit`
7. **Frontend Tests**: `npm test --prefix cmd/opsforall-gui/frontend`
8. **ESLint**: `npm run lint --prefix cmd/opsforall-gui/frontend`
9. **Push & PR**: Push branch, open PR with template

### Important Commands

```bash
# Dev server (hot-reload backend + frontend)
wails dev

# Production build
wails build

# Race detection testing
go test -race ./internal/...

# Specific package tests with verbose output
go test ./internal/sysops/... -v -count=1

# Frontend coverage
npm test --prefix cmd/opsforall-gui/frontend -- --coverage

# Full verification (build + vet + test + lint)
# Use the "Verify All" task in VS Code
```

### Hot Reload Tips

- `wails dev` watches both Go and frontend files
- Go changes trigger automatic binary rebuild (~2–5s)
- Frontend changes are instant via Vite HMR
- To reload frontend without backend restart, save a `.tsx` file
- If the Wails dev server feels slow, run `wails dev -loglevel Error` for quieter logs

---

## 5. Debugging & Profiling

### Backend Debugging

```bash
# Run with verbose Wails logging
wails dev -loglevel Debug

# Run backend tests with race detection
go test -race ./internal/... -v

# Print collector metrics to stdout (add temporary LogInfo calls)
# Metrics are logged via common.LogInfo in engine.go
```

### Frontend Debugging

- Open browser DevTools (F12) at `http://localhost:5173`
- React DevTools extension works for component inspection
- TanStack Query DevTools show cache state and request status
- Zustand store state can be inspected via React DevTools

### SQLite Debugging

The database file lives at `universalops.db` in the working directory:
```bash
# Install SQLite CLI
# Quick inspection
sqlite3 universalops.db ".tables"
sqlite3 universalops.db "SELECT * FROM metrics LIMIT 10;"
```

### Profiling

```bash
# CPU profile (30s)
go test -bench=. -cpuprofile=cpu.out ./internal/...
go tool pprof cpu.out

# Memory profile
go test -bench=. -memprofile=mem.out ./internal/...
go tool pprof mem.out
```

---

## 6. PR Best Practices

### Before Opening a PR

- [ ] Run `go build ./...` and `go vet ./...` — zero errors
- [ ] Run `go test ./internal/... -count=1 -timeout 120s` — all green
- [ ] Run `npx tsc --noEmit` in frontend directory — no type errors
- [ ] Run `npm test --prefix cmd/opsforall-gui/frontend` — all green
- [ ] Run `golangci-lint run ./...` — no lint warnings
- [ ] Test on Windows (primary target) — everything works

### PR Size Guidelines

| Size | Lines Changed | Review Time | Notes |
|------|--------------|-------------|-------|
| Tiny | < 30 | Fast | Single concern, docs, trivial fixes |
| Small | 30–150 | Normal | Most PRs should be this size |
| Medium | 150–500 | Slower | Break into smaller PRs if possible |
| Large | 500+ | Very slow | Strongly prefer splitting |

### Review Checklist

- Does the code follow Go idioms and project conventions?
- Are there new cloud dependencies or telemetry? (should be none)
- Are tests included for new functionality?
- Do existing tests still pass?
- Are error messages wrapped with context?
- Is the frontend using functional components and proper memoization?

---

## 7. Environment Variables Reference

| Variable | Default | Description |
|----------|---------|-------------|
| `OLLAMA_HOST` | `http://localhost:11434` | Ollama server URL |
| `OLLAMA_MODEL` | `universalops` | AI model name for Hawk analyst |
| `UNIVERSALOPS_DB_PATH` | `./universalops.db` | SQLite database file location |
| `UNIVERSALOPS_LOG_LEVEL` | `info` | Log level (debug, info, warn, error) |
| `UNIVERSALOPS_COLLECT_INTERVAL` | `3s` | Telemetry collection interval |
| `UNIVERSALOPS_DISABLE_COLLECTORS` | — | Comma-separated list of disabled collectors |

---

> **Next**: Read the [Contributing Guide](../CONTRIBUTING.md) for the contribution process and coding standards.
