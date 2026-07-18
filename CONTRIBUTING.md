# Contributing to OpsForAll

Thank you for considering contributing to OpsForAll! This document outlines the process for contributing to the project.

## Code of Conduct

Be respectful, constructive, and inclusive.

## Prerequisites

- **Go** 1.26+
- **Node.js** 20+ and **npm**
- **Wails CLI**: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`

## Getting Started

1. Fork the repository
2. Clone your fork: `git clone https://github.com/your-username/opsforall.git`
3. Create a branch: `git checkout -b feat/your-feature-name`

## Development Workflow

### Quick Start

```bash
wails dev        # development with hot-reload
wails build      # production binary
```

### Testing

```bash
go test ./internal/...                                          # backend tests
npm test --prefix cmd/opsforall-gui/frontend                    # frontend tests
npx tsc --noEmit --project cmd/opsforall-gui/frontend           # type check
```

### Linting

```bash
golangci-lint run ./...
npm run lint --prefix cmd/opsforall-gui/frontend
```

## Pull Request Guidelines

- Follow the existing technical style
- Write tests for new functionality
- Update documentation where applicable
- Use [Conventional Commits](https://www.conventionalcommits.org/) format

### Commit Format

```
<type>(<scope>): <description>
```

Types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`
Scopes: `sysops`, `netops`, `secops`, `devops`, `aiops`, `common`, `frontend`, `docs`

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
- **Database**: Use `allopsfull.db` for all persistent storage logic.

## Security

- All processing must remain local.
- Never commit secrets or credentials.
- Command execution must go through the sandbox layer.
- Report vulnerabilities by opening an issue.
