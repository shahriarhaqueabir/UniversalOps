package ui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
)

// HelpOverlay displays keyboard shortcuts and general help.
type HelpOverlay struct {
	visible bool
}

// NewHelpOverlay creates a new help overlay.
func NewHelpOverlay() *HelpOverlay {
	return &HelpOverlay{visible: false}
}

// Toggle shows or hides the help overlay.
func (h *HelpOverlay) Toggle() {
	h.visible = !h.visible
}

// Visible returns whether the help is currently shown.
func (h *HelpOverlay) Visible() bool {
	return h.visible
}

// Hide hides the help overlay.
func (h *HelpOverlay) Hide() {
	h.visible = false
}

// HandleKey returns true if this key event was consumed by the help overlay.
func (h *HelpOverlay) HandleKey(msg tea.KeyPressMsg) bool {
	if !h.visible {
		return false
	}
	if msg.String() == "?" || msg.String() == "esc" || msg.String() == "q" {
		h.visible = false
		return true
	}
	return true // consume all keys while help is visible
}

// Render returns the help overlay content.
func (h *HelpOverlay) Render() string {
	if !h.visible {
		return ""
	}

	helpContent := strings.Builder{}
	helpContent.WriteString(OnboardingTitleStyle.Render("⌨ Keyboard Shortcuts"))
	helpContent.WriteString("\n\n")

	sections := []struct {
		title string
		keys  [][2]string // [key, description]
	}{
		{
			title: "Navigation",
			keys: [][2]string{
				{"↑ / ↓ or k / j", "Move cursor up/down"},
				{"Enter or Space", "Select item"},
				{"Esc", "Go back / close help"},
				{"Tab / Shift+Tab", "Next / previous tab"},
			},
		},
		{
			title: "Quick Jumps",
			keys: [][2]string{
				{"1", "System Operations (SysOps)"},
				{"2", "Network Operations (NetOps)"},
				{"3", "Security Operations (SecOps)"},
				{"4", "Development Operations (DevOps)"},
				{"5", "AI Operations (AI Ops)"},
			},
		},
		{
			title: "Reports",
			keys: [][2]string{
				{"R", "Generate combined report for current layer"},
				{"G", "Generate text report (AIOps)"},
				{"M", "Generate markdown report (AIOps)"},
				{"S", "Save report to disk (AIOps)"},
			},
		},
		{
			title: "Actions",
			keys: [][2]string{
				{"q", "Quit application"},
				{"?", "Toggle this help screen"},
				{"t", "Cycle UI themes"},
				{"r", "Refresh current view"},
				{"/", "Filter/search"},
			},
		},
	}

	for _, section := range sections {
		helpContent.WriteString(PanelTitleStyle.Render(section.title))
		helpContent.WriteString("\n")
		for _, pair := range section.keys {
			helpContent.WriteString("  ")
			helpContent.WriteString(HelpKeyStyle.Render(pair[0]))
			helpContent.WriteString("  ")
			helpContent.WriteString(HelpDescStyle.Render(pair[1]))
			helpContent.WriteString("\n")
		}
		helpContent.WriteString("\n")
	}

	helpContent.WriteString(SubtitleStyle.Render("Press [?], [Esc], or [q] to close"))

	return PanelStyle.Render(helpContent.String())
}
