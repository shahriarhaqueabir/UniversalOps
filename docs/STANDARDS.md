# Universal-Ops — Development Standards

## Go Conventions (Backend)

### Versioning & Idioms
- **Go 1.26+**: Leverage new reflection iterators (`reflect.Value.Fields()`) for system-state mapping.
- **Strict URL Parsing**: Be aware of `urlstrictcolons=1`. Ensure all hostname strings provided by users are validated before `url.Parse`.
- **Error Handling**: Use the optimized `fmt.Errorf` (automatic direct return of `errors.New` if no formatting is present).

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

- Follow `gofmt` / `goimports` (enforced by CI).
- Use `go vet` and `staticcheck` for static analysis.
- Error handling: Always check errors. Wrap with context using `fmt.Errorf("context: %w", err)`.
- **Wails Bindings**: Methods bound to the UI must be exported (`PascalCase`) and should return `(result, error)` for automatic JS Promise rejection handling.

### Imports Ordering

```go
import (
    // Standard library
    "context"
    "fmt"
    "time"

    // External dependencies
    "github.com/shirou/gopsutil/v4/cpu"
    "github.com/wailsapp/wails/v2/pkg/runtime"

    // Internal packages
    "github.com/shahriarhaqueabir/UniversalOps/internal/common"
    "github.com/shahriarhaqueabir/UniversalOps/internal/app"
)
```

## React & TypeScript Conventions (Frontend)

### React 19 Standards
- **Actions**: Prefer `useActionState` and `useOptimistic` for data mutations (e.g., AI chat input, setting overrides).
- **Transitions**: Use `startTransition(async () => ...)` for long-running IPC calls to maintain UI responsiveness.
- **Hook Primatives**: Use `use(Context)` or `use(Promise)` where applicable for cleaner async data loading.

### Naming

- **Components**: `PascalCase` (e.g., `DashboardCard.tsx`).
- **Hooks**: `camelCase` with `use` prefix (e.g., `useBackend.ts`).
- **Stores**: `camelCase` with `use` prefix and `Store` suffix (e.g., `useSettingsStore.ts`).
- **Styles**: Tailwind classes preferred. Custom CSS variables in `globals.css` must use the `--color-*` pattern.

### Styling (Tailwind v4)

- **No Hardcoded Colors**: Use CSS variables (e.g., `text-[var(--color-text)]`, `bg-[var(--color-bg)]`).
- **Design Consistency**: Follow the Squib-inspired dark theme. Use `var(--color-accent)` for primary actions.
- **States**: Use `hover:bg-[var(--color-sidebar-hover)]` instead of arbitrary opacity (e.g., `bg-white/5`).

### State Management (Zustand)

- Keep stores atomic and focused.
- Use selectors to prevent unnecessary re-renders:
  ```typescript
  const refreshInterval = useSettingsStore((s) => s.refreshInterval);
  ```

## Wails & IPC Standards

- **Events**: Use `useEvents('event-name')` hook for backend-to-frontend updates.
- **Bindings**: All Go binding calls should be wrapped in `useQuery` or `useMutation` from TanStack Query for caching and loading states.
- **Type Safety**: Run `wails generate module` to update TypeScript definitions whenever Go structs in `internal/app/` change.

## Commit Conventions

We use [Conventional Commits](https://www.conventionalcommits.org/):

### Types

| Type | Usage |
|------|-------|
| `feat` | A new feature |
| `fix` | A bug fix |
| `docs` | Documentation changes |
| `style` | Formatting, CSS variables, Tailwind |
| `refactor` | Code restructuring |
| `test` | Adding/updating tests |
| `chore` | Build, CI, dependencies |

### Scopes

| Scope | Area |
|-------|------|
| `ui` | React components, hooks, stores |
| `sysops` | System operations backend |
| `netops` | Network operations backend |
| `secops` | Security operations backend |
| `devops` | DevOps/Shell backend |
| `aiops` | AI/Ollama integration |
| `common` | Pipeline, storage, shared utils |

## Code Review Checklist

Before submitting a PR:

- [ ] Tests pass (`go test ./...` and `npm test`).
- [ ] No hardcoded hex/RGB colors in components (use CSS vars).
- [ ] No `any` types in TypeScript (unless strictly necessary for Wails generic events).
- [ ] `wails.json` version matches `App.go` and `package.json`.
- [ ] Accessibility: All interactive elements have focus states and labels.
- [ ] No debug `console.log` or `fmt.Printf` (use `common.LogInfo` or `zerolog`).

## Performance Guidelines

- **Tick Loop**: Keep the `DataPipeline` tick interval at 3 seconds to balance real-time feel with CPU overhead.
- **React Rendering**: Memoize expensive charts or large tables using `React.memo` or `useMemo`.
- **Virtualization**: Use `@tanstack/react-virtual` for any list/table exceeding 50 rows (e.g., Log Viewer, Process List).

---

*Last updated: 2026-07-12*
