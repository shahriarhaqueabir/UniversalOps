# Windows Desktop E2E Tests (pywinauto)

End-to-end tests for the Universal-Ops Wails desktop application using pywinauto + UI Automation.

## Prerequisites

```bash
pip install pywinauto pytest pytest-html pytest-timeout pillow
```

## Running Tests

```bash
# Set required environment variables
export APP_PATH="E:\Projects\projectx\AllOpsFull\build\bin\universal-ops.exe"
export APP_TITLE="Universal-Ops"
export LAUNCH_TIMEOUT=15

# Run all tests
cd tests/e2e
pytest test_tabs.py -v

# Run smoke tests only
pytest test_tabs.py -m smoke -v

# Run with custom app path
APP_PATH="./build/bin/universal-ops.exe" pytest test_tabs.py -v
```

## Test Structure

```
tests/e2e/
├── config.py          # AutomationId constants, timeouts, paths
├── conftest.py        # Fixtures (app launch with isolation), failure screenshots
├── pages.py           # Page Object Model for each tab/page
└── test_tabs.py       # Tab verification tests
```

## Automation IDs

All tab elements in the React frontend have `data-automation-id` attributes that map to UIA `AutomationId`:

| Page | Prefix | Example |
|------|--------|---------|
| Main tabs | `main-tab-*` | `main-tab-netops` |
| NetOps sub-tabs | `netops-tab-*` | `netops-tab-ping` |
| SecOps sub-tabs | `secops-tab-*` | `secops-tab-firewall` |
| SysOps sub-tabs | `sysops-tab-*` | `sysops-tab-processes` |
| Logs sub-tabs | `logs-tab-*` | `logs-tab-live` |
| DevOps sub-tabs | `devops-tab-*` | `devops-tab-terminal` |
| AIOps sub-tabs | `aiops-tab-*` | `aiops-tab-anomalies` |

## Page Objects

Each page class encapsulates:
- Locators (by AutomationId, Name, Class)
- Wait helpers (`wait_visible`, `wait_gone`, `wait_window`)
- Actions (`click`, `type_text`, `get_text`)
- Artifact collection (screenshots, step traces)

## CI/CD Integration

```yaml
# .github/workflows/e2e-desktop.yml
jobs:
  e2e:
    runs-on: windows-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-python@v5
        with: { python-version: "3.11" }
      - name: Build app
        run: wails build
      - name: Run E2E
        env:
          APP_PATH: ${{ github.workspace }}\build\bin\universal-ops.exe
          APP_TITLE: "Universal-Ops"
        run: pytest tests/e2e/test_tabs.py -v
      - uses: actions/upload-artifact@v4
        if: always()
        with:
          name: e2e-artifacts
          path: tests/e2e/artifacts/
```

## Debugging

```bash
# Enable step-by-step tracing (screenshots + JSONL log)
E2E_TRACE=1 pytest tests/e2e/test_tabs.py::TestNetOpsSubtabs::test_netops_ping_probe_flow -v

# Include typed text in trace (NEVER use on credentials!)
E2E_TRACE=1 E2E_TRACE_INCLUDE_TEXT=1 pytest ...

# Inspect UIA tree at runtime
python -c "
from pywinauto import Desktop
Desktop(backend='uia').windows()
"
```

## Adding New Tests

1. Add `data-automation-id` to React component
2. Add constant to `config.py`
3. Add locator method to page object in `pages.py`
4. Write test in `test_tabs.py`