# Hawkward — Onboarding & Persistence

## First-Run Wizard
- 5-step onboarding wizard shown on first launch
- Returning users go directly to main menu

## Persistence
- Marker file at `%APPDATA%/hawkward/.onboarded` (Windows) or `~/.config/hawkward/.onboarded` (Linux)
- Functions: `common.IsOnboarded()`, `common.MarkOnboarded()`, `common.ClearOnboarded()`

## TUI Key Bindings
| Key | Action |
|-----|--------|
| `1`-`5` | Switch ops layer (SysOps/NetOps/SecOps/DevOps/AIOps) |
| `R` | Generate report for current layer |
| `r` | Refresh/clear current view |
| `?` | Help overlay |
| `q` / `Ctrl+C` | Quit |
| `Tab` / `Shift+Tab` | Cycle tabs within layer |
| `Esc` | Go back |
