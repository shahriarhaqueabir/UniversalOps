# Hawkward — Development Standards

## Go Conventions

### Naming

| Pattern | Convention | Example |
|---------|-----------|---------|
| Packages | lowercase, no underscores | `sysops`, `netops` |
| Exported types | PascalCase | `CPUStats`, `NetworkInterface` |
| Unexported types | camelCase | `cpuCollector`, `netScanner` |
| Exported functions | PascalCase | `GetCPUUsage()`, `ScanPorts()` |
| Unexported functions | camelCase | `collectMetrics()`, `parseOutput()` |
| Constants | PascalCase | `MaxRetries`, `DefaultTimeout` |
| Variables | camelCase | `cpuCount`, `totalMemory` |
| Files | snake_case | `cpu_usage.go`, `port_scanner.go` |

### Code Style

- Follow `gofmt` / `goimports` (enforced by CI)
- Use `golangci-lint` with strict rules
- Max line length: 120 characters
- Error handling: always check errors, wrap with context using `fmt.Errorf("context: %w", err)`
- No panics in production code
- No global state

### Imports Ordering

```go
import (
    // Standard library
    "fmt"
    "os"
    "time"

    // External dependencies
    tea "charm.land/bubbletea/v2"
    "charm.land/lipgloss/v2"
    "github.com/shirou/gopsutil/v4/cpu"

    // Internal packages
    "hawkward/internal/common"
    "hawkward/internal/ui"
)
```

### Testing

- All packages must have at least 80% test coverage
- Table-driven tests preferred
- Use `testing.T` for unit tests
- Mock external dependencies (gopsutil, OS calls) with interfaces
- Example test structure:

```go
func TestGetCPUUsage(t *testing.T) {
    tests := []struct {
        name string
        mock func()
        want float64
    }{
        {"normal usage", func() { /* setup mock */ }, 45.2},
        {"zero usage", func() { /* setup mock */ }, 0.0},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            tt.mock()
            got := GetCPUUsage()
            if got != tt.want {
                t.Errorf("GetCPUUsage() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

## Commit Conventions

We use [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <description>

[optional body]
```

### Types

| Type | Usage |
|------|-------|
| `feat` | A new feature |
| `fix` | A bug fix |
| `docs` | Documentation changes |
| `style` | Formatting, whitespace |
| `refactor` | Code restructuring |
| `test` | Adding/updating tests |
| `chore` | Build, CI, dependencies |

### Scopes

| Scope | Area |
|-------|------|
| `ui` | TUI components, styling |
| `sysops` | System operations layer |
| `netops` | Network operations layer |
| `secops` | Security operations layer |
| `devops` | DevOps layer |
| `aiops` | AI Ops layer |
| `common` | Shared utilities |
| `docs` | Documentation |

### Examples

```
feat(sysops): add CPU usage dashboard panel
fix(netops): handle DNS timeout gracefully
docs: update installation instructions
test(sysops): add table-driven tests for memory stats
```

## Branch Strategy

```
main           # Production-ready releases
├── develop    # Integration branch
│   ├── feat/*       # New features
│   ├── fix/*        # Bug fixes
│   ├── refactor/*   # Code refactoring
│   └── docs/*       # Documentation
```

- Feature branches branch from `develop`
- Pull requests merge into `develop`
- `main` is updated via release PRs from `develop`
- Branch naming: `feat/sysops-cpu-dashboard`, `fix/dns-timeout`

## Code Review Checklist

Before submitting a PR:

- [ ] Tests pass (`go test ./...`)
- [ ] Lint passes (`golangci-lint run`)
- [ ] No debug code or commented-out code
- [ ] All public functions documented with godoc comments
- [ ] Error handling is complete (no ignored errors)
- [ ] No hardcoded paths or secrets
- [ ] Cross-platform considerations addressed
- [ ] Keyboard navigation works (no dead ends)
- [ ] Terminal resize handling correct

## Performance Guidelines

- Keep message handling in `Update()` under 1ms (keyboard responsiveness)
- Offload blocking operations (DNS, ping, port scan) to goroutines
- Use `tea.Batch` to run parallel operations
- Refresh dashboards at 1-5 second intervals (configurable)
- Profile with `pprof` before optimizing

---

*Last updated: 2026-07-01*
