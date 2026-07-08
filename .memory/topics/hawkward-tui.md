# Hawkward — TUI Layer (Legacy — Removed)

> **Removed in Sprint 11/Release Pipeline**: The Bubble Tea TUI layer was superseded by the Wails v2 GUI.
> Legacy code archived in `legacy/tui/` has been deleted. All functionality migrated to the React/TypeScript frontend.

## Historical Components (formerly in `internal/ui/`)
- `root.go` — Root Bubble Tea model, message routing
- `mainmenu.go` — Main menu model and view
- `help.go` — Help overlay (`?` key)
- `keys.go` — Key binding definitions
- `onboarding.go` — 5-step first-run wizard
- `statusbar.go` — Bottom status bar with layer/version info
- `styles.go` — UI-specific lipgloss styles

## Status Bar (Historical)
- Showed current ops layer, version, and connection indicator
