package aiops

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
)

// View renders the AIOps dashboard.
func (m *Model) View(width, height int, stats *common.SystemStats) string {
	m.rememberStats(stats)

	var b strings.Builder

	// Title
	b.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color(common.CurrentPalette().Primary)).
		Bold(true).
		Render("🤖 AI Operations"))
	b.WriteString("\n")

	// Ollama status bar
	ollamaStatus := m.renderOllamaStatus()
	b.WriteString("  " + ollamaStatus + "\n\n")

	// Tab bar
	tabs := []string{" Chat ", " Reports "}
	for i, tab := range tabs {
		TabStyle := lipgloss.NewStyle().Padding(0, 1)
		if i == m.tabIndex {
			TabStyle = TabStyle.
				Foreground(lipgloss.Color(common.CurrentPalette().Primary)).
				Bold(true).
				Border(lipgloss.NormalBorder(), false, false, true, false).
				BorderForeground(lipgloss.Color(common.CurrentPalette().Primary))
		} else {
			TabStyle = TabStyle.Foreground(lipgloss.Color(common.CurrentPalette().Muted))
		}
		b.WriteString(TabStyle.Render(fmt.Sprintf("[%d]%s", i+1, tab)))
	}
	b.WriteString("\n\n")

	// Render active tab
	switch m.tabIndex {
	case 0:
		m.renderChatTab(&b)
	case 1:
		m.renderReportsTab(&b)
	}

	// Help footer
	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color(common.CurrentPalette().Muted)).
		Render("[Tab/←→] switch tabs  [R] generate report  [c] clear  [esc] back"))

	return common.Panel.Width(width - 4).Render(b.String())
}

// renderOllamaStatus returns a string showing Ollama availability.
func (m *Model) renderOllamaStatus() string {
	if m.ollamaStatus == nil {
		return lipgloss.NewStyle().Foreground(lipgloss.Color(common.CurrentPalette().Warning)).Render("⏳ Checking Ollama...")
	}
	if !m.ollamaStatus.Available {
		return lipgloss.NewStyle().Foreground(lipgloss.Color(common.CurrentPalette().Danger)).Render("⛔ Ollama not available (start with 'ollama serve')")
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color(common.CurrentPalette().Secondary)).Render(
		fmt.Sprintf("✅ Ollama • %s", m.ollamaStatus.Model))
}

// renderChatTab renders the AI chat interface.
func (m *Model) renderChatTab(b *strings.Builder) {
	b.WriteString(common.PanelTitle.Render("AI Chat"))
	b.WriteString("\n")

	if !m.ollamaAvailable() {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(common.CurrentPalette().Warning)).Render(
			"  Offline mode: ask about CPU, memory, disk, processes, uptime, status, or anomalies."))
		b.WriteString("\n")
	}

	// Messages
	if len(m.messages) == 0 {
		b.WriteString(common.Muted.Render(
			"  No messages yet. Try `status`, `state cpu`, or `anomalies`.\n"))
	} else {
		for _, msg := range m.messages {
			switch msg.Role {
			case "user":
				b.WriteString(fmt.Sprintf("  %s %s\n",
					lipgloss.NewStyle().Foreground(lipgloss.Color(common.CurrentPalette().Info)).Bold(true).Render("You:"),
					lipgloss.NewStyle().Foreground(lipgloss.Color(common.CurrentPalette().Text)).Render(msg.Content)))
			case "assistant":
				b.WriteString(fmt.Sprintf("  %s %s\n",
					lipgloss.NewStyle().Foreground(lipgloss.Color(common.CurrentPalette().Secondary)).Bold(true).Render("AI:"),
					lipgloss.NewStyle().Foreground(lipgloss.Color(common.CurrentPalette().Muted)).Render(msg.Content)))
			}
			b.WriteString("\n")
		}
	}

	// Input line
	b.WriteString("\n")
	prompt := lipgloss.NewStyle().Foreground(lipgloss.Color(common.CurrentPalette().Info)).Bold(true).Render("> ")
	inputDisplay := m.input
	if inputDisplay == "" {
		inputDisplay = common.Muted.Render("Type your message...")
	} else {
		inputDisplay = lipgloss.NewStyle().Foreground(lipgloss.Color(common.CurrentPalette().Text)).Render(inputDisplay)
	}
	b.WriteString("  " + prompt + inputDisplay + "\n")

	// Chat hint
	b.WriteString("\n")
	b.WriteString(common.Muted.Render(
		"  [c] clear conversation"))

	// Error display
	if m.err != nil {
		b.WriteString("\n\n")
		b.WriteString(common.Error.Render(fmt.Sprintf("  Error: %v", m.err)))
		b.WriteString("\n")
	}
}

// renderReportsTab renders the report generation interface.
func (m *Model) renderReportsTab(b *strings.Builder) {
	b.WriteString(common.PanelTitle.Render("Report Generator"))
	b.WriteString("\n")

	b.WriteString(common.Muted.Render(
		"  Enter Title|Content and press Enter to add a section."))
	b.WriteString("\n")
	b.WriteString(common.Muted.Render(
		"  Commands: generate, markdown, json, csv, save:/path, save-json:/path, save-csv:/path, clear"))
	b.WriteString("\n\n")

	// Input line
	prompt := lipgloss.NewStyle().Foreground(lipgloss.Color(common.CurrentPalette().Warning)).Bold(true).Render("section> ")
	inputDisplay := m.input
	if inputDisplay == "" {
		inputDisplay = common.Muted.Render("Title|Content")
	} else {
		inputDisplay = lipgloss.NewStyle().Foreground(lipgloss.Color(common.CurrentPalette().Text)).Render(inputDisplay)
	}
	b.WriteString("  " + prompt + inputDisplay + "\n\n")

	// Existing sections
	if len(m.reports) > 0 {
		b.WriteString(common.PanelTitle.Render("Current Sections"))
		b.WriteString("\n")
		for i, section := range m.reports {
			b.WriteString(fmt.Sprintf("  %d. %s\n", i+1,
				lipgloss.NewStyle().Foreground(lipgloss.Color(common.CurrentPalette().Secondary)).Bold(true).Render(section.Title)))
			preview := section.Content
			if len(preview) > 60 {
				preview = preview[:57] + "..."
			}
			if preview != "" {
				b.WriteString(fmt.Sprintf("     %s\n",
					common.Muted.Render(preview)))
			}
		}
		b.WriteString("\n")
	} else {
		b.WriteString(common.Muted.Render(
			"  No sections yet. Add a section above.\n\n"))
	}

	// Generated report output
	if m.output != "" {
		b.WriteString(common.PanelTitle.Render("Generated Report"))
		b.WriteString("\n")
		b.WriteString(common.Output.Render(m.output))
		b.WriteString("\n")
	}

	// Report action hints
	if len(m.reports) > 0 {
		b.WriteString(common.Muted.Render(
			fmt.Sprintf("  [g] text  [m] markdown  [j] json  [v] csv  [s] save  [c] clear  ·  %d section(s)", len(m.reports))))
		b.WriteString("\n")
	}

	// Error display
	if m.err != nil {
		b.WriteString("\n")
		b.WriteString(common.Error.Render(fmt.Sprintf("  Error: %v", m.err)))
		b.WriteString("\n")
	}
}
