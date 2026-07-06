package ui

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines all key bindings for the application.
type KeyMap struct {
	Quit    key.Binding
	Up      key.Binding
	Down    key.Binding
	Enter   key.Binding
	Back    key.Binding
	Help    key.Binding
	Refresh key.Binding
	TabNext key.Binding
	TabPrev key.Binding
	Filter  key.Binding
	Number1 key.Binding
	Number2 key.Binding
	Number3 key.Binding
	Number4 key.Binding
	Number5 key.Binding
}

// DefaultKeyMap returns the default key bindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter", " "),
			key.WithHelp("enter", "select"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
		),
		TabNext: key.NewBinding(
			key.WithKeys("tab", "l"),
			key.WithHelp("tab", "next tab"),
		),
		TabPrev: key.NewBinding(
			key.WithKeys("shift+tab", "h"),
			key.WithHelp("S+tab", "prev tab"),
		),
		Filter: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "filter"),
		),
		Number1: key.NewBinding(
			key.WithKeys("1"),
			key.WithHelp("1", "SysOps"),
		),
		Number2: key.NewBinding(
			key.WithKeys("2"),
			key.WithHelp("2", "NetOps"),
		),
		Number3: key.NewBinding(
			key.WithKeys("3"),
			key.WithHelp("3", "SecOps"),
		),
		Number4: key.NewBinding(
			key.WithKeys("4"),
			key.WithHelp("4", "DevOps"),
		),
		Number5: key.NewBinding(
			key.WithKeys("5"),
			key.WithHelp("5", "AI Ops"),
		),
	}
}

// FullHelp returns all key bindings for the help screen.
func (k KeyMap) FullHelp() []key.Binding {
	return []key.Binding{
		k.Quit, k.Help, k.Up, k.Down,
		k.Enter, k.Back, k.Refresh,
		k.TabNext, k.TabPrev, k.Filter,
		k.Number1, k.Number2, k.Number3,
		k.Number4, k.Number5,
	}
}

// ShortHelp returns the minimal key bindings for the status bar.
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.Quit, k.Help, k.Enter, k.Back,
	}
}
