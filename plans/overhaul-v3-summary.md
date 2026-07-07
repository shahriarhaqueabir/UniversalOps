# Hawkward Ops v3 — Overhaul Executive Summary

> **Theme**: Squib-inspired card-based interactive dashboard + real-time forecasting + global command palette
> **Timeline**: 9 weeks (3 sprints visible, 5 sprints total)
> **Go**: 1.26.4 | **TUI**: Bubble Tea v2 + Lip Gloss v2 | **Binary**: Single, zero external runtime deps

---

## Sprint Overview — What Ships When

| Sprint | Title | What Ships | New Files | Value |
|--------|-------|-----------|-----------|-------|
| **Sprint 1** (Wk 1-2) | Visualization Engine | Chart library, time-series store, forecast engine, card system, config | ~13 new files | Foundation — no visible change yet |
| **Sprint 2** (Wk 3-6) | Dashboard & Redesign | Dashboard landing page, all 5 layer views rewritten with cards/charts | ~3 new, ~15 modified | **Biggest visible change** |
| **Sprint 3** (Wk 7) | Intelligence | Predictive forecasts, alerting with flap detection, AI insights | ~2 new, ~8 modified | Smart predictions |
| **Sprint 4** (Wk 8) | Power User | Global command palette (`/`), keyboard macros, drill-down, export | ~1 new, ~10 modified | Speed & discoverability |
| **Sprint 5** (Wk 9) | Polish & Cross-Platform | Config UI, responsive layout, performance, session logging | ~2 new, ~12 modified | Production readiness |

**Total**: ~21 new files, ~45 modified files

---

## Team Structure — Agent Ownership

| Lane | Agent Role | Owns | Key Files |
|------|-----------|------|-----------|
| **A** | Charts Engineer | `internal/common/charts/*`, `timeseries.go`, `forecast.go` | ~8 new chart files |
| **B** | UI Architect | `dashboard.go`, `cards.go`, all `view.go` rewrites | UI layer + all ops views |
| **C** | Intelligence Lead | `forecast.go` (predictions), `alerts.go`, `alertview.go` | Prediction + alerting |
| **D** | AI Integrator | `ollama.go` (insight prompts) | AI summarization |
| **E** | UX Engineer | `commandpalette.go`, `keys.go`, `help.go` | Search + navigation |
| **F** | Config Engineer | `config.go` (independent lane) | YAML config |
| **G** | Platform Lead | `settings.go`, `logger.go`, cross-platform fixes | Polish + hardening |
| **Coordinator** | Planning & Docs | Plans, architecture docs, sprint kanban | All docs |

---

## Key Architectural Decisions

1. **Charts are pure Lip Gloss + Unicode** — No external charting deps. Braille (⣀⣤⣶⣿) for line charts, block chars (▁▂▃▄▅▆▇█) for bars and gauges.
2. **Time-series store is a ring buffer** — Fixed-size, O(1) push, rolling window aggregation (min/max/avg/p50/p95/p99).
3. **Dashboard replaces main menu** — `ScreenDashboard` becomes the default home screen; main menu is demoted to a nav drawer.
4. **Cards replace tabs** — Each ops layer view migrates from tab-based text to a card grid with focus navigation.
5. **Global command palette via `/`** — Fuzzy search accessible from any screen (replaces the current per-screen `/` filter).
6. **Forecast engine uses stdlib `math` only** — Linear regression + exponential smoothing. No external ML deps.
7. **Config lives at `~/.config/hawkward/config.yaml`** — YAML format, reloadable at runtime.
8. **All new code follows existing patterns** — `tea.Model`/`Update()`/`View()`, `common.LogInfo` logging, `common.Palette` theming.

---

## Quick-Start for New Agents

### Entry points to understand first
```bash
cmd/hawkward/main.go              # App entry point
internal/ui/root.go               # RootModel — navigation, routing, tick loop
internal/common/types.go          # Screen enum, SystemStats, shared types
internal/common/theme.go          # Palette struct, theme management
internal/common/styles.go         # Common styled primitives (PanelTitle, Value, etc.)
```

### Build & test
```bash
go build -o hawkward.exe ./cmd/hawkward
go test ./...
go vet ./...
```

### Current sprint kanban
See `docs/SPRINT_KANBAN.md` for active cards, owners, and merge gates.

### Critical conventions
- New packages go under `internal/` (private) — never `pkg/` for app code
- Every new file needs a corresponding `*_test.go`
- All chart colors come from `common.Palette` — never hardcode colors
- Follow TEA (The Elm Architecture): Model → Update → View
- Log operations via `common.LogInfo("Op: %s", action)`

---

## File Change Quick Reference

### New packages to create
```
internal/common/charts/    — 8 files (config, line, bar, area, gauge, sparkline, heatmap, number)
internal/common/           — 3 new files (timeseries.go, forecast.go, config.go)
internal/ui/               — 6 new files (dashboard.go, cards.go, commandpalette.go, settings.go, logviewer.go, alertview.go)
internal/common/           — 1 new file (alerts.go — Sprint 2)
```

### Critical existing files to modify
```
internal/ui/root.go        — ScreenDashboard, command palette routing, config init
internal/ui/styles.go      — Card, chart, dashboard style primitives
internal/ui/statusbar.go   — Alert count pulse
internal/ui/keys.go        — Rebindable keys, macros, bookmarks
internal/common/theme.go   — 10 themes, chart-specific Palette fields
internal/*/view.go         — 5 view files to rewrite with cards
```

---

*Last updated: 2026-07-07*
*Owner: Planning & Documentation Lead*
