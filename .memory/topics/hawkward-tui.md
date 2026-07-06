# Hawkward — TUI Layer

## Components (all in `internal/ui/`)
- `root.go` — Root Bubble Tea model, message routing
- `mainmenu.go` — Main menu model and view
- `help.go` — Help overlay (`?` key)
- `keys.go` — Key binding definitions
- `onboarding.go` — 5-step first-run wizard
- `statusbar.go` — Bottom status bar with layer/version info
- `styles.go` — UI-specific lipgloss styles

## Status Bar
- Shows current ops layer, version, and connection indicator
- Placeholder for "Degraded Mode" indicator (not yet implemented)
