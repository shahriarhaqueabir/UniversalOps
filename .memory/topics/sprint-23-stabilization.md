# Sprint 23 — Stabilization: Defensive Guards, CI Fixes, Release Pipeline

## Changes Made (Committed)

### 1. Runtime Error Fix: `Cannot read properties of null (reading 'map')`

**Root cause**: Dashboard's `useEvents('metrics', ...)` callback assumed the Wails event payload always contained a complete `DashboardData` object. If the backend emitted a partial/metrics event before full initialization, `d.cpu` could be null, causing a cascade.

**Fix** (`cmd/hawkward-gui/frontend/src/pages/Dashboard.tsx`):
- Added null check in `useEvents` callback: guards `!d || !d.cpu || !d.memory || !d.disk`
- Added optional chaining (`?.`) in `computeRedFlags()` for `cpu`, `memory`, `disk` access
- Added optional chaining in `HeroSection()` for `stats.cpu?.value`, `stats.memory?.value`, `stats.disk?.value`
- Added `stats.cpu?.history` fallback in `useMemo` dependency

### 2. CI: eslint-plugin-react-hooks Resolution

**Root cause**: `package-lock.json` had `eslint-plugin-react-hooks@5.2.0` which only supports eslint `^3.0.0 - ^9.0.0`, but the project uses eslint `^10.6.0`.

**Fix**: Already resolved in previous commit (`8519ba3`). Lockfile was verified in-sync with `^7.1.1` (supports eslint 10).

### 3. CI: TestPing Linux Failure

**Root cause**: GitHub Actions Ubuntu 24.04+ runners don't have `CAP_NET_RAW` on the `ping` binary, causing `ping exec: exit status 255`.

**Fix** (`.github/workflows/test.yml`):
- Added `sudo setcap cap_net_raw+ep /bin/ping` step before Go tests
- Uses `|| true` to avoid failing if binary is at different path

### 4. Release Pipeline: Private Repo

**Root cause**: `softprops/action-gh-release` with `secrets.GITHUB_TOKEN` cannot create releases on private repos. Needs PAT.

**Fix** (`.github/workflows/release.yml`):
- Changed to `secrets.GH_TOKEN || secrets.GITHUB_TOKEN` fallback
- User must create a classic PAT with `repo` scope and save as `GH_TOKEN` secret

## Verification

| Check | Result |
|-------|--------|
| `npm test` | 27/27 ✅ |
| `npm run lint` | ✅ |
| `npx tsc --noEmit` | ✅ |
| `go test ./internal/...` | 7/7 ✅ |
| `wails build` | ✅ |

## AI Model Configuration

**Default model**: `agentic-coder` (based on `hf.co/Jackrong/Qwopus3.5-9B-Coder-GGUF:Q4_K_M`)
**Location**: `E:\Projects\ai\ollama\modelfiles\agentic-coder.Modelfile`
**Fallback**: If model not found in Ollama, uses first available model
**Override**: `OLLAMA_MODEL` env var

The `agentic-coder` model was chosen because:
1. It's fine-tuned for code/operations tasks
2. Custom modelfile has system prompt optimized for system diagnostics
3. Qwopus3.5-9B-Coder base provides good quality for the ops domain

## Dashboard Architecture (Quick Diag & Briefing)

**Quick Diagnostic** (`Dashboard.RunQuickDiag()`):
- Uses DataPipeline to get current metrics: CPU, Memory, Disk, Process count, Alerts
- Each metric categorized as `pass/warn/fail` based on thresholds
- Returns `[]DiagnosticResult` array rendered as overlay list

**Generate Briefing** (`Dashboard.GenerateDashboardBriefing()`):
- Uses DataPipeline metrics with trend analysis
- Creates `[]BriefingSection` with title, content, level (info/warning/critical)
- Includes alert section if active alerts exist
- Rendered as overlay with color-coded cards

**Platform Vitality (HeroSection)**:
- Shows health donut SVG (average of CPU, Memory, Disk)
- CPU history bars (sampled to 90 bars)
- Active uptime display
- Active alerts indicator

**Compute Logic Analysis**:
- Dynamic red flags computed client-side from real pipeline data
- Rules: CPU>90 critical, CPU>70+rising, Memory>92, Memory>80+rising, Disk>95, Disk>85
- Not hardcoded — adapts to current data

## Known Remaining Issues

1. **Release CI**: Still needs `GH_TOKEN` secret added to repository
2. **Frontend test coverage**: TopBar, NetOps, SecOps, DevOps, AIOps, NetworkDesign missing tests (P3)
3. **Recharts ResponsiveContainer**: Console stderr in jsdom (cosmetic)
4. **TestInsertLogAndQuery**: Flaky when run in full suite (SQLite DB leaks)
