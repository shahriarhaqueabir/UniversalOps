package ui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
)

// MainMenu is the application's home screen with category selection.
type MainMenu struct {
	items  []common.MenuItem
	cursor int
	width  int
	height int
}

// NewMainMenu creates a new main menu.
func NewMainMenu() *MainMenu {
	items := []common.MenuItem{
		{
			Title:       "System Operations",
			Description: "CPU, memory, disk, processes, system info",
			Screen:      common.ScreenSysOps,
			Key:         "1",
		},
		{
			Title:       "Network Operations",
			Description: "Ping, traceroute, port scan, DNS, connections",
			Screen:      common.ScreenNetOps,
			Key:         "2",
		},
		{
			Title:       "Security Operations",
			Description: "Firewall, users, listening ports, defender, tasks",
			Screen:      common.ScreenSecOps,
			Key:         "3",
		},
		{
			Title:       "Development Operations",
			Description: "Command runner, log tailer, file browser",
			Screen:      common.ScreenDevOps,
			Key:         "4",
		},
		{
			Title:       "AI Operations",
			Description: "Local AI assistant, report generation",
			Screen:      common.ScreenAIOps,
			Key:         "5",
		},
	}

	return &MainMenu{
		items:  items,
		cursor: 0,
	}
}

// Navigate returns the selected screen.
func (m *MainMenu) Navigate() common.Screen {
	return m.items[m.cursor].Screen
}

// SetSize updates the menu dimensions.
func (m *MainMenu) SetSize(w, h int) {
	m.width = w
	m.height = h
}

// Update handles keyboard events for the menu.
func (m *MainMenu) Update(msg tea.Msg) (bool, common.Screen) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		case "enter", " ", "space":
			return true, m.items[m.cursor].Screen
		case "1":
			return true, common.ScreenSysOps
		case "2":
			return true, common.ScreenNetOps
		case "3":
			return true, common.ScreenSecOps
		case "4":
			return true, common.ScreenDevOps
		case "5":
			return true, common.ScreenAIOps
		}
	}
	return false, common.ScreenMainMenu
}

// Render returns the menu view.
func (m *MainMenu) Render() string {
	var b strings.Builder

	// Title
	b.WriteString(TitleStyle.Render("🚀 Hawkward Operations Platform"))
	b.WriteString("\n")
	b.WriteString(SubtitleStyle.Render("Terminal-based multi-layer operations dashboard"))
	b.WriteString("\n\n")

	// Menu items
	for i, item := range m.items {
		cursor := "  "
		if i == m.cursor {
			cursor = "▸ "
		}

		itemStyle := MenuItemStyle
		if i == m.cursor {
			itemStyle = MenuSelectedStyle
		}

		line := strings.Builder{}
		line.WriteString(MenuKeyStyle.Render("[" + item.Key + "]"))
		line.WriteString(itemStyle.Render(item.Title))

		b.WriteString(cursor)
		b.WriteString(line.String())
		b.WriteString("\n")
		b.WriteString("     ")
		b.WriteString(MenuDescStyle.Render(item.Description))
		b.WriteString("\n\n")
	}

	// Footer hints
	b.WriteString(Divider)
	b.WriteString("\n")
	b.WriteString(SubtitleStyle.Render(
		"Press [?] for help  |  [↑/↓] or [k/j] to navigate  |  [1-5] to jump  |  [q] to quit",
	))

	return AppStyle.Render(b.String())
}
