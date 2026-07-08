# Contributing to Hawkward

Thank you for considering contributing to Hawkward! This document outlines the process for contributing to the project.

## Code of Conduct

Be respectful, constructive, and inclusive. We're all here to build something useful.

## Prerequisites

- **Go** 1.26+
- **Node.js** 20+ and **npm**
- **Wails CLI**: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`

## Getting Started

1. Fork the repository
2. Clone your fork: `git clone https://github.com/your-username/hawkward.git`
3. Create a branch: `git checkout -b feat/your-feature-name`

## Development Workflow

### Quick Start

```bash
cd AllOpsFull
wails dev        # development with hot-reload
wails build      # production binary
```

### Testing

```bash
go test ./internal/...                                          # backend tests
cd cmd/hawkward-gui/frontend && npm test                        # frontend tests
cd cmd/hawkward-gui/frontend && npx tsc --noEmit                # type check
```

### Linting

```bash
golangci-lint run ./...
cd cmd/hawkward-gui/frontend && npm run lint
```

## Pull Request Guidelines

- Follow the existing code style (run `gofmt` and `goimports`)
- Write tests for new functionality
- Update documentation where applicable
- Keep commits small and focused
- Use [Conventional Commits](https://www.conventionalcommits.org/) format

### Commit Format

```
<type>(<scope>): <description>

[optional body]
```

Types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`
Scopes: `sysops`, `netops`, `secops`, `devops`, `aiops`, `common`, `frontend`, `docs`

## Project Structure

See [ARCHITECTURE.md](docs/ARCHITECTURE.md) for a detailed overview.

Key directories:

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
├── cmd/hawkward-gui/
│   └── frontend/            ← React + Vite + Tailwind
├── scripts/                 ← Build & release scripts
└── docs/                    ← Documentation
```

## Coding Standards

- **Go**: Follow standard Go project layout, error wrapping, idiomatic Go patterns. Use `common.LogInfo` for logging.
- **Frontend**: React + TypeScript with `camelCase` naming, Tailwind v4 utility classes, Lucide React icons, Recharts for charts, Radix UI primitives.
- Follow existing patterns in each subsystem rather than introducing new styles.

## Security

- Never commit secrets, tokens, or credentials
- Use environment variables for configuration
- All command execution must go through the sandbox layer
- Report vulnerabilities by opening an issue

## Questions?

Open an issue or start a discussion. We're happy to help!
