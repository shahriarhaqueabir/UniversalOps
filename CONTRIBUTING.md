# Contributing to Hawkward

Thank you for considering contributing to Hawkward! This document outlines the process for contributing to the project.

## Code of Conduct

Be respectful, constructive, and inclusive. We're all here to build something useful.

## Getting Started

1. Fork the repository
2. Clone your fork: `git clone https://github.com/your-username/hawkward.git`
3. Create a branch: `git checkout -b feat/your-feature-name`
4. Install Go 1.26.4+
5. Run tests: `go test ./...`

## Development Workflow

### Building

```bash
go build -o hawkward.exe ./cmd/hawkward
```

### Testing

```bash
go test ./...
go vet ./...
```

### Linting

```bash
golangci-lint run ./...
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
Scopes: `ui`, `sysops`, `netops`, `secops`, `devops`, `aiops`, `common`, `docs`

## Project Structure

See [ARCHITECTURE.md](docs/ARCHITECTURE.md) for a detailed overview.

Key directories:
- `cmd/hawkward/` — Entry point
- `internal/` — Application logic organized by ops layer
- `internal/ui/` — TUI components and routing
- `internal/common/` — Shared utilities and styles

## Coding Standards

- Follow the Standard Go Project Layout
- Use Bubble Tea's Model-View-Update pattern
- Use Lip Gloss for all styling; avoid hardcoded ANSI escapes
- Use `common` formatters and styles instead of inline hex colors
- Test coverage should remain at or above 80%

## Security

- Never commit secrets, tokens, or credentials
- Use environment variables for configuration
- All command execution must go through the sandbox layer
- Report vulnerabilities by opening an issue

## Questions?

Open an issue or start a discussion. We're happy to help!
