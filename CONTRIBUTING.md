# Contributing to Universal-Ops

Thank you for your interest in making Universal-Ops better! We aim to build the most robust, local-first operations platform for Windows workstation telemetry.

## Core Philosophy
1.  **Local-First**: No data leaves the machine. No telemetry to us, no cloud sync.
2.  **Native-First**: Use built-in system tools (WMI, PowerShell, gopsutil) before adding external dependencies.
3.  **High Density**: Information-rich, professional UIs. No "padding" or wasted space.

## Development Environment
- **Go**: 1.26.5+
- **Node.js**: 20+ (using npm)
- **Wails CLI**: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`

### Running the App
```bash
# Frontend dependencies
cd cmd/opsforall-gui/frontend && npm install

# Start Dev Mode (Hot Reload for Go & React)
wails dev
```

## Contribution Process
1.  **Fork** the repo and create your branch from `main`.
2.  **Implement** your changes following the technical style.
3.  **Test** thoroughly (see Testing section).
4.  **Submit** a PR with a clear description of the value added.

## Testing Standards
Every PR must pass existing tests and include new ones for added features.
```bash
# Backend
go test ./...

# Frontend
cd cmd/opsforall-gui/frontend && npm test
```

## Commit Conventions
We use [Conventional Commits](https://www.conventionalcommits.org/):
- `feat(sysops)`: add temperature tracking
- `fix(ui)`: align grid cards on mobile
- `docs(readme)`: update installation steps

## Project Structure

```
AllOpsFull/
├── main.go                  ← Entry point (Wails)
├── internal/
│   ├── app/                 ← Wails bound facades
│   ├── common/              ← Shared utilities, storage, pipeline
│   ├── sysops/              ← CPU/Memory/Disk monitoring
│   ├── netops/              ← Network operations
│   ├── secops/              ← Security operations
│   ├── devops/              ← Shell/services/file ops
│   └── aiops/               ← AI operations (Ollama)
├── cmd/opsforall-gui/
│   └── frontend/            ← React + Vite + Tailwind
└── docs/                    ← Documentation
```

## Coding Standards

- **Go**: Follow standard Go project layout and idiomatic patterns. Use `common.LogInfo` for logging.
- **Frontend**: React + TypeScript with Tailwind v4 and Radix UI.
- **Database**: Use `universalops.db` for all persistent storage logic.

## Security

- All processing must remain local.
- Never commit secrets or credentials.
- Command execution must go through the sandbox layer.
- Report vulnerabilities by opening an issue.
