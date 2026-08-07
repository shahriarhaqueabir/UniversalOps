# UniversalOps Testing Guide

Canonical reference for all testing practices, conventions, and infrastructure in the UniversalOps project.

---

## Table of Contents

1. [Overview](#overview)
2. [Go Backend Tests](#go-backend-tests)
3. [Frontend Tests](#frontend-tests)
4. [Integration Tests](#integration-tests)
5. [Fuzz Tests](#fuzz-tests)
6. [Benchmarks](#benchmarks)
7. [E2E Tests](#e2e-tests)
8. [Running Tests](#running-tests)
9. [Coverage Targets](#coverage-targets)
10. [CI Pipeline](#ci-pipeline)
11. [Test Checklist (PRs)](#test-checklist-prs)

---

## Overview

| Layer | Framework | Location | Coverage Target |
|-------|-----------|----------|-----------------|
| Go unit | `testing` + `testify` | `internal/**/*_test.go` | ≥60% |
| Go integration | `testing` + build tag | `internal/**/*_integration_test.go` | ≥60% |
| Go fuzz | `testing/fuzz` | `internal/**/*_test.go` | N/A |
| Go benchmarks | `testing.B` | `internal/**/*_bench_test.go` | N/A |
| React unit | Vitest + Testing Library | `frontend/src/**/*.test.*` | ≥60% |
| E2E (desktop) | pywinauto + pytest | `tests/e2e/` | Smoke ≥100% |

---

## Go Backend Tests

### Assertion Library

Use **`github.com/stretchr/testify`** for all assertions. Import the `assert` and `require` packages:

```go
import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestSomething(t *testing.T) {
    result := DoSomething()
    assert.NotNil(t, result)
    assert.Equal(t, 42, result.Value)
    assert.NoError(t, result.Err)

    // require stops test on failure (no subsequent nil-dereference panics)
    require.NotNil(t, result.SubObject)
    assert.True(t, result.SubObject.IsValid)
}
```

> **Why testify?** It's the most widely adopted Go assertion library (74k+ GitHub stars), zero external dependencies, compatible with `testing.T`, and doesn't force a BDD structure on tests.

### Naming Conventions

| Pattern | Example |
|---------|---------|
| Test files | `*_test.go` — same package as source |
| Test functions | `TestFeatureName` — descriptive PascalCase |
| Sub-tests | `t.Run("case description", ...)` — lowercase, spaces OK |
| Table-driven | Anonymous structs inside `TestXxx` |
| Benchmarks | `BenchmarkFeatureName` |
| Fuzz targets | `FuzzParseInput` |

### Table-Driven Test Pattern

```go
func TestFormatBytes(t *testing.T) {
    tests := []struct {
        name   string
        input  uint64
        want   string
    }{
        {"zero", 0, "0 B"},
        {"kilobyte", 1024, "1.00 KB"},
        {"megabyte", 1048576, "1.00 MB"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := FormatBytes(tt.input)
            assert.Equal(t, tt.want, got)
        })
    }
}
```

### Mock Strategy

- **Prefer interfaces** for testability. Define small interfaces near their usage site.
- **Use manual mock structs** for simple cases (embed the interface, override methods you need).
- For complex mocks, use **`testify/mock`**:

```go
type MockStorage struct {
    mock.Mock
}

func (m *MockStorage) Get(key string) (string, error) {
    args := m.Called(key)
    return args.String(0), args.Error(1)
}
```

### What to Test

| Priority | What | Example |
|----------|------|---------|
| 🔴 Critical | Pure business logic | `FormatBytes`, `ParsePortRange`, anomaly detection |
| 🟠 High | Facade methods | Wails-bound `GetCPUInfo`, `ScanPorts` |
| 🟡 Medium | Error paths | Nil pipeline, closed storage, timeout |
| 🟢 Low | Trivial getters/setters | Skip unless they contain logic |

---

## Frontend Tests

### Stack

- **Runner**: Vitest (configured in `vitest.config.ts`)
- **DOM**: jsdom
- **Assertions**: `@testing-library/jest-dom` (custom matchers like `toBeInTheDocument`)
- **User events**: `@testing-library/user-event`
- **Accessibility**: `vitest-axe` (see below)

### Component Test Pattern

```tsx
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, it, expect, vi } from 'vitest'
import { HealthBadge } from './HealthBadge'

describe('HealthBadge', () => {
  it('renders healthy state', () => {
    render(<HealthBadge status="healthy" />)
    expect(screen.getByText('Healthy')).toBeInTheDocument()
  })

  it('renders critical state with danger color', () => {
    render(<HealthBadge status="critical" />)
    const badge = screen.getByText('Critical')
    expect(badge).toHaveClass('text-red-500') // or check style
  })

  it('calls onClick when clicked', async () => {
    const onClick = vi.fn()
    render(<HealthBadge status="warning" onClick={onClick} />)
    await userEvent.click(screen.getByRole('button'))
    expect(onClick).toHaveBeenCalledTimes(1)
  })
})
```

### Store Test Pattern (Zustand)

```ts
import { describe, it, expect, beforeEach } from 'vitest'
import { useConfigStore } from './useConfigStore'

describe('useConfigStore', () => {
  beforeEach(() => {
    useConfigStore.setState(useConfigStore.getInitialState())
  })

  it('sets polling interval', () => {
    useConfigStore.getState().setPollingInterval(5000)
    expect(useConfigStore.getState().pollingInterval).toBe(5000)
  })

  it('has correct defaults', () => {
    const state = useConfigStore.getState()
    expect(state.pollingInterval).toBe(3000)
    expect(state.theme).toBe('system')
  })
})
```

### Hook Test Pattern

```ts
import { renderHook, act } from '@testing-library/react'
import { describe, it, expect, vi } from 'vitest'
import { useWidgetRefresh } from './useWidgetRefresh'

describe('useWidgetRefresh', () => {
  beforeEach(() => {
    vi.useFakeTimers()
  })

  it('calls onRefresh at interval', () => {
    const onRefresh = vi.fn()
    renderHook(() => useWidgetRefresh(onRefresh, 1000))
    expect(onRefresh).toHaveBeenCalledTimes(1) // initial
    act(() => { vi.advanceTimersByTime(1000) })
    expect(onRefresh).toHaveBeenCalledTimes(2)
  })
})
```

### Accessibility Testing

```bash
npm install --save-dev vitest-axe
```

```tsx
import { render } from '@testing-library/react'
import { axe, toHaveNoViolations } from 'vitest-axe'
expect.extend(toHaveNoViolations)

it('has no accessibility violations', async () => {
  const { container } = render(<Sidebar />)
  const results = await axe(container)
  expect(results).toHaveNoViolations()
})
```

Priority components for a11y testing: `Sidebar`, `TopBar`, `ConfirmDialog`, `ErrorBoundary`, `SearchInput`.

### Snapshot Tests

Use snapshots for components that change infrequently (layout, navigation):

```tsx
it('renders consistently', () => {
  const { container } = render(<TopBar title="UniversalOps" />)
  expect(container).toMatchSnapshot()
})
```

Review snapshot diffs carefully in PRs. Update with `npx vitest run --update`.

---

## Integration Tests

Integration tests are gated behind the `integration` build tag and use real (or file-based) dependencies rather than mocks.

### File naming

```
feature_integration_test.go   // must end with the `_integration_test.go` suffix
```

### Build tag

Every integration test file starts with:

```go
//go:build integration

package common_test
```

### What to Cover

| Area | Scope |
|------|-------|
| Storage | SQLite CRUD, WAL mode, concurrent reads/writes, backup/restore |
| EventBus | Multi-subscriber, concurrent delivery, unsubscribe mid-stream |
| Pipeline | Full data flow: push → persist → query → forecast |
| Workflows | Multi-step orchestration: create → execute → rollback → resume |
| App facades | Bound methods under real conditions (started pipeline, open storage) |

### Running Integration Tests

```bash
# All integration tests
go test -tags=integration ./internal/... -count=1 -timeout 180s

# Single package
go test -tags=integration ./internal/common/... -count=1 -v
```

Integration tests are **not** run in the default `go test ./...` — only with `-tags=integration`.

---

## Fuzz Tests

Go 1.26+ native fuzzing for input-heavy code.

### When to Add Fuzz Targets

| Code | Why |
|------|-----|
| `netops/actions.go` — `ParsePortRange` | User-provided port strings |
| `netops/dns.go` — DNS query inputs | Network input parsing |
| `secops/audit.go` — log line parsing | Arbitrary file content |
| `aiops/analyst.go` — prompt construction | LLM injection surface |

### Pattern

```go
func FuzzParsePortRange(f *testing.F) {
    f.Add("80-443")
    f.Add("1-65535")
    f.Add("0")
    f.Fuzz(func(t *testing.T, input string) {
        result := ParsePortRange(input)
        // Should never panic — always return valid or error
        if result != nil {
            assert.LessOrEqual(t, result.Start, result.End)
        }
    })
}
```

### Running Fuzz Tests

```bash
# Run indefinitely (until crash or Ctrl+C)
go test -fuzz=FuzzParsePortRange ./internal/netops/ -fuzztime=30s

# Regression on existing corpus
go test -fuzz=FuzzParsePortRange ./internal/netops/ -run=FuzzParsePortRange
```

---

## Benchmarks

### When to Add Benchmarks

| Code | Why |
|------|-----|
| `pipeline.go` — `PushMetric` + `GetTimeSeries` | Called at every collection tick |
| `events.go` — EventBus dispatch | Concurrency overhead |
| `storage.go` — bulk insert + query | Database bottleneck |
| `anomalies.go` — z-score calculation | Math-heavy, called per metric |

### Pattern

```go
func BenchmarkPipelinePush(b *testing.B) {
    dp := NewDataPipeline(DefaultCollectionConfig())
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        dp.PushMetric("cpu.percent", "%", float64(i))
    }
}

func BenchmarkEventBusPublish(b *testing.B) {
    bus := NewEventBus()
    bus.Subscribe("test", func(Event) {})
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        bus.Publish(Event{Type: "test", Data: "payload"})
    }
}
```

### Running Benchmarks

```bash
# Run all benchmarks
go test -bench=. ./internal/common/... -benchmem

# Compare with benchstat
go test -bench=. -count=5 -benchmem ./internal/common/... > /tmp/old.txt
# ... make changes ...
go test -bench=. -count=5 -benchmem ./internal/common/... > /tmp/new.txt
benchstat /tmp/old.txt /tmp/new.txt
```

### Regression Threshold

If a benchmark regresses by >10% in CI, the pipeline should warn. Track benchmark history over releases.

---

## E2E Tests

### Stack

- **Framework**: pywinauto (UIA backend) + pytest
- **Runner**: `python run_e2e_tests.py` (or `pytest tests/e2e/ -v` directly)
- **Windows-only**: Tests require the built `universal-ops.exe` binary

### Running Locally

```bash
# Build first
wails build -platform windows/amd64 -o build/bin/universal-ops.exe

# Run E2E tests
python run_e2e_tests.py

# Or with more verbosity
cd tests/e2e
pytest test_tabs.py test_pipeline.py -v --tb=long
```

### Page Objects

UI interactions are defined in `tests/e2e/pages.py`. When selectors break (common after Wails rebuilds), use the inspector:

```bash
cd tests/e2e
python inspect_uia.py   # dumps full UIA tree for debugging
```

### Smoke Tests (Must-Pass)

Tests marked `@smoke` are the minimum bar for any PR. They should be fast and reliable:
- App launches without crash
- Dashboard loads
- Settings opens

### Debugging Failures

1. Check `tests/e2e/artifacts/` for FAIL screenshots
2. Run `python inspect_uia.py` to see current UIA structure
3. Update selectors in `pages.py` or `config.py` if automation IDs changed
4. Run just the failing test: `pytest test_tabs.py::TestMainTabs -v`

---

## Running Tests

### Quick Reference

```bash
# ── Go Backend ──
go test ./internal/... -count=1 -timeout 120s        # Unit tests only
go test -race ./internal/... -count=1 -timeout 180s   # With race detection
go test -tags=integration ./internal/... -count=1     # + integration tests
go test -bench=. ./internal/common/... -benchmem      # Benchmarks
go test -fuzz=FuzzXxx ./internal/netops/ -fuzztime=30s # Fuzz (runs until crash)

# ── Coverage ──
go test ./internal/... -coverprofile=coverage.out -covermode=atomic
go tool cover -html=coverage.out -o coverage.html     # Visual coverage report

# ── Frontend ──
npm test --prefix cmd/opsforall-gui/frontend                    # Unit tests
npx vitest run --coverage --prefix cmd/opsforall-gui/frontend   # With coverage
npx tsc -b --prefix cmd/opsforall-gui/frontend                 # Type check (build mode — root tsconfig is project references)
npm run lint --prefix cmd/opsforall-gui/frontend                 # Lint

# ── E2E (Windows) ──
python run_e2e_tests.py
```

### VS Code Tasks

The project includes VS Code tasks for all common test operations:

| Task | Description |
|------|-------------|
| `🧪 Test All Backend` | `go test ./internal/...` |
| `🧪 Test Backend Package` | Pick a specific package |
| `🧪 Test Backend with Coverage` | Coverage + HTML report |
| `🧪 Test with Race Detection` | `go test -race ./internal/...` |
| `🧪 Test Backend + Integration` | `go test -tags=integration` |
| `🧪 Test Frontend` | `vitest run` |
| `🧪 Test Frontend with Coverage` | `vitest run --coverage` |

---

## Coverage Targets

| Scope | Target | Threshold | Action if Missed |
|-------|--------|-----------|------------------|
| Project coverage | ≥60% | ±5% | Review new code, add tests |
| Patch coverage | ≥60% | ±5% | Add tests for changed code |
| Critical packages (`common/`, `aiops/`) | ≥70% | ±5% | Priority escalation |

Coverage is measured via:
- **Go**: `go test -coverprofile=coverage.out -covermode=atomic`
- **Frontend**: `npx vitest run --coverage`
- **Aggregation**: Codecov (uploaded in CI)

---

## CI Pipeline

The full CI pipeline runs on every push to `main`/`develop` and every PR to `main`:

```
Checkout → Setup Go + Node → Install deps → Build →
  Vet → Lint (golangci-lint) →
  🧪 Go Unit Tests + Coverage →
  🧪 Go Race Detection →
  🧪 Frontend Tests + Coverage →
  🔍 TypeScript Check →
  🔍 ESLint →
  📦 Wails Build →
  🐍 E2E Tests (smoke=required, full=optional) →
  🔒 npm audit →
  📤 Upload Coverage (Codecov) →
  📎 Archive Test Artifacts (on failure)
```

### Gating Rules

| Gate | Enforced | Exception |
|------|----------|-----------|
| Go build + vet | ✅ Yes | — |
| Go tests pass | ✅ Yes | — |
| Frontend tests pass | ✅ Yes | — |
| TypeScript compiles | ✅ Yes | — |
| ESLint clean | ✅ Yes | — |
| Go race detection | ✅ Yes | — |
| E2E smoke tests | ✅ Yes | — |
| E2E full suite | ⚠️ Warning | Known flaky tests documented |
| Codecov upload | ⚠️ Advisory | Comment on PR only |

---

## Test Checklist (PRs)

Every PR must pass this checklist before merge:

```markdown
### ✅ Testing Checklist
- [ ] Backend unit tests pass
- [ ] Frontend unit tests pass
- [ ] TypeScript compiles (`tsc -b` — build mode, not `--noEmit` which is a false green)
- [ ] ESLint passes
- [ ] New backend code has ≥60% coverage
- [ ] New frontend code has tests (component + store + hook as applicable)
- [ ] Integration tests updated if storage/workflow/event code changed
- [ ] Race detection run (`go test -race ./internal/...`)
- [ ] E2E smoke tests pass (or failure is documented)
- [ ] Accessibility checked (axe scan on new/changed components)
```

---

## Evolving This Guide

This guide should be updated when:
- A new testing tool or framework is adopted
- Coverage targets change
- CI pipeline steps are added or removed
- New test patterns emerge that should be standardized

Update this file in the same PR that introduces the change.
