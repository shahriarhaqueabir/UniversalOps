# Contributing to UniversalOps

Thank you for your interest in making UniversalOps better! We aim to build the most robust, local-first operations platform for Windows workstation telemetry.

## Core Philosophy
1.  **Local-First**: No data leaves the machine. No telemetry to us, no cloud sync.
2.  **Native-First**: Use built-in system tools (WMI, PowerShell, gopsutil) before adding external dependencies.
3.  **High Density**: Information-rich, professional UIs. No "padding" or wasted space.

## Table of Contents
- [Development Environment](#development-environment)
- [Architecture Overview](#architecture-overview)
- [Contribution Process](#contribution-process)
- [Testing Standards](#testing-standards)
- [Code Quality](#code-quality)
- [Commit Conventions](#commit-conventions)
- [Database Migrations](#database-migrations)
- [Security Considerations](#security-considerations)
- [Project Structure](#project-structure)
- [Coding Standards](#coding-standards)

---

## Development Environment
- **Go**: 1.26+
- **Node.js**: 22+ (using npm)
- **Wails CLI**: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
- **GCC (Windows)**: MinGW or TDM-GCC
- **Linux**: `sudo apt install gcc libgtk-3-dev libwebkit2gtk-4.1-dev`
- **macOS**: Xcode Command Line Tools

### Running the App
```bash
# Frontend dependencies
cd cmd/opsforall-gui/frontend && npm install

# Start Dev Mode (Hot Reload for Go & React)
wails dev

# Production build
wails build
```

---

## Architecture Overview

UniversalOps uses a **goroutine-per-collector** architecture:

```
React Frontend ←→ Wails IPC ←→ Go App Layer ←→ Engine Loop
                                                    ├── Data Pipeline → SQLite
                                                    ├── Alert Engine (in-memory + SQLite)
                                                    ├── System Collectors (goroutine-per-metric)
                                                    └── Hawk AI → Ollama (local LLM)
```

Key architectural decisions are documented in [docs/adr/](./docs/adr/). Read them before making significant changes.

---

## Contribution Process
1.  **Fork** the repo and create your branch from `main`.
2.  **Implement** your changes following the technical style.
3.  **Test** thoroughly (see Testing Standards below).
4.  **Submit** a PR with a clear description of the value added.

> 📖 **New to the codebase?** Read the [Developing Guide](docs/developing.md) for the full annotated project tree, git strategy, debugging tips, and environment reference.

---

## Testing Standards

Every PR must pass existing tests and include new ones for added features.

### Backend Tests
```bash
# Run all backend tests
go test ./internal/... -count=1 -timeout 120s

# Run a specific package with verbose output
go test ./internal/sysops -count=1 -timeout 30s -v

# Run with coverage
go test ./internal/... -count=1 -timeout 120s -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Frontend Tests
```bash
cd cmd/opsforall-gui/frontend

# Run all tests
npx vitest run

# Run in watch mode during development
npx vitest

# Run with coverage
npx vitest run --coverage
```

### E2E Tests (Windows only)
```bash
python tests/e2e/run_e2e_tests.py
```

### CI Checks (run these before pushing)
```bash
# Backend
go build ./...
go vet ./...

# Frontend
cd cmd/opsforall-gui/frontend
npx tsc -b      # build mode — root tsconfig is project references, --noEmit is a false green
npm run lint
npx vitest run
```

### Coverage Targets
- **Backend**: Aim for 60%+ coverage on new packages
- **Frontend**: Aim for 80%+ coverage on new components
- **Critical paths** (alert engine, data pipeline): 80%+ coverage

---

## Code Quality

- **Go**: `golangci-lint run ./...` — run before pushing
- **Frontend**: ESLint + TypeScript strict mode
- **No `any` types** in TypeScript — use proper types or `unknown`
- **No `panic()`** in Go — return errors and handle them
- **No `fmt.Print*`** in production code — use `common.LogInfo` / `common.LogError`

---

## Commit Conventions
We use [Conventional Commits](https://www.conventionalcommits.org/):
- `feat(sysops): add temperature tracking`
- `fix(ui): align grid cards on mobile`
- `docs(readme): update installation steps`
- `refactor(engine): extract alert evaluation`
- `test(common): add pipeline backpressure tests`
- `ci: add Windows ARM64 build target`

### Commit Message Format
```
<type>(<scope>): <description>

[optional body with details about why the change was made]
```

---

## Database Migrations

UniversalOps uses SQLite (`universalops.db`) with schema migrations managed in `internal/common/storage.go`.

### Adding a Migration
1. Add a new migration function in `internal/common/storage.go`
2. Register it in the migration list with a version number
3. Include both `UP` (apply) and `DOWN` (rollback) steps
4. Test against a copy of the production database

### Migration Guidelines
- Never modify an existing migration — add a new one
- Use `ALTER TABLE` for additive changes (new columns, indexes)
- Use transactions for multi-step migrations
- Test rollback before deploying

---

## Security Considerations

- **All processing must remain local** — never send telemetry data to external services
- **Never commit secrets or credentials** — use environment variables for configuration
- **Command execution** must go through the sandbox layer (`internal/common/sandbox.go`)
- **Report vulnerabilities** by emailing the maintainer (see `git log` for contact) — do not open public issues
- **SQL injection**: All SQL queries must use parameterized statements (prepared statements)
- **Path traversal**: Validate all file paths from user input before reading/writing

---

## Project Structure

```
UniversalOps/
├── main.go                  ← Entry point (Wails)
├── internal/
│   ├── app/                 ← Wails bound facades (API layer)
│   ├── common/              ← Shared: Pipeline, Storage, Alerts, Sandbox, Engine
│   ├── sysops/              ← CPU/Memory/Disk/Sensors monitoring
│   ├── netops/              ← DNS, ICMP, Port scanning
│   ├── secops/              ← Identity, Firewall, Privilege tracking
│   ├── devops/              ← Shell, Services, File operations, Docker
│   └── aiops/               ← Ollama integration, RCA, Anomaly detection
├── cmd/opsforall-gui/
│   └── frontend/            ← React 19 + Vite + Tailwind v4
├── docs/
│   ├── adr/                 ← Architecture Decision Records
│   ├── ARCHITECTURE.md      ← System design
│   └── ...                  ← Other documentation
├── data/                    ← Runtime data (baselines, modelfiles)
├── logs/                    ← Application logs
└── tests/e2e/               ← End-to-end test suite
```

---

## Coding Standards

- **Go**: Follow standard Go project layout and idiomatic patterns. Use `common.LogInfo` for logging. Wrap errors with `fmt.Errorf("context: %w", err)`.
- **Frontend**: React + TypeScript with Tailwind v4 and Radix UI. Functional components only. Memoize expensive renders with `React.memo` and `useMemo`.
- **Database**: Use `universalops.db` for all persistent storage logic. All queries through `internal/common/storage.go`.
- **Naming**: `PascalCase` for exports, `camelCase` for unexported, `snake_case` for filenames.
