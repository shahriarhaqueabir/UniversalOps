package devops

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
)

// View renders the DevOps dashboard.
func (m *Model) View(width, height int, _ *common.SystemStats) string {
	var b strings.Builder

	// Show workflow report if active
	if m.showReport && m.workflowReport != "" {
		b.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color(common.CurrentPalette().Primary)).
			Bold(true).
			Render("📊 Development Operations Report"))
		b.WriteString("\n\n")
		b.WriteString(common.Output.Render(m.workflowReport))
		b.WriteString("\n")
		b.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color(common.CurrentPalette().Muted)).
			Render("[r] back to dashboard  [esc] back  [1-5] jump"))
		return common.Panel.Width(width - 4).Render(b.String())
	}

	// Title
	b.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color(common.CurrentPalette().Primary)).
		Bold(true).
		Render("⚙️ Development Operations"))
	b.WriteString("\n")

	// Tab bar
	tabs := []string{" Shell ", " Logs ", " Files ", " Processes ", " Services "}
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
		m.renderShellTab(&b)
	case 1:
		m.renderLogsTab(&b)
	case 2:
		m.renderFilesTab(&b)
	case 3:
		m.renderProcessesTab(&b)
	case 4:
		m.renderServicesTab(&b)
	}

	// Help footer
	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color(common.CurrentPalette().Muted)).
		Render("[Tab/←→] switch tabs  [r] clear  [R] report  [esc] back  [1-5] jump"))

	return common.Panel.Width(width - 4).Render(b.String())
}

// renderShellTab renders the shell command interface.
func (m *Model) renderShellTab(b *strings.Builder) {
	b.WriteString(common.PanelTitle.Render("Shell Command"))
	b.WriteString("\n")

	// Input line
	prompt := lipgloss.NewStyle().Foreground(lipgloss.Color(common.CurrentPalette().Secondary)).Bold(true).Render("$ ")
	inputDisplay := m.input
	if inputDisplay == "" {
		inputDisplay = common.Muted.Render("type a command and press Enter...")
	} else {
		inputDisplay = lipgloss.NewStyle().Foreground(lipgloss.Color(common.CurrentPalette().Text)).Render(inputDisplay)
	}
	b.WriteString("  " + prompt + inputDisplay + "\n\n")

	// Output
	if m.err != nil {
		b.WriteString(common.Error.Render(fmt.Sprintf("  Error: %v", m.err)))
		b.WriteString("\n")
	} else if m.output != "" {
		b.WriteString(common.Output.Render(m.output))
		b.WriteString("\n")
	} else {
		b.WriteString(common.Muted.Render(
			"  No output yet. Run a command to see results.\n"))
	}

	// History hint
	if len(m.results) > 0 {
		b.WriteString("\n")
		b.WriteString(common.Muted.Render(
			fmt.Sprintf("  %d command(s) in history. Press [r] to clear.", len(m.results))))
	}
}

// renderLogsTab renders the log viewer interface.
func (m *Model) renderLogsTab(b *strings.Builder) {
	b.WriteString(common.PanelTitle.Render("Log Viewer"))
	b.WriteString("\n")

	// Instructions
	b.WriteString(common.Muted.Render(
		"  Enter a file path to tail (last 50 lines), or use search:/path:pattern"))
	b.WriteString("\n\n")

	// Input line
	prompt := lipgloss.NewStyle().Foreground(lipgloss.Color(common.CurrentPalette().Info)).Bold(true).Render("path> ")
	inputDisplay := m.input
	if inputDisplay == "" {
		inputDisplay = common.Muted.Render("e.g. /var/log/syslog")
	} else {
		inputDisplay = lipgloss.NewStyle().Foreground(lipgloss.Color(common.CurrentPalette().Text)).Render(inputDisplay)
	}
	b.WriteString("  " + prompt + inputDisplay + "\n\n")

	// Output
	if m.err != nil {
		b.WriteString(common.Error.Render(fmt.Sprintf("  Error: %v", m.err)))
		b.WriteString("\n")
	} else if m.output != "" {
		b.WriteString(common.Output.Render(m.output))
		b.WriteString("\n")
	} else {
		b.WriteString(common.Muted.Render(
			"  No file loaded. Enter a path and press Enter.\n"))
	}
}

// renderFilesTab renders the file browser interface.
func (m *Model) renderFilesTab(b *strings.Builder) {
	b.WriteString(common.PanelTitle.Render("File Browser"))
	b.WriteString("\n")

	// Instructions
	b.WriteString(common.Muted.Render(
		"  Enter a directory path to browse, or read:/path to view a file"))
	b.WriteString("\n\n")

	// Input line
	prompt := lipgloss.NewStyle().Foreground(lipgloss.Color(common.CurrentPalette().Warning)).Bold(true).Render("dir> ")
	inputDisplay := m.input
	if inputDisplay == "" {
		inputDisplay = common.Muted.Render("e.g. . or /home")
	} else {
		inputDisplay = lipgloss.NewStyle().Foreground(lipgloss.Color(common.CurrentPalette().Text)).Render(inputDisplay)
	}
	b.WriteString("  " + prompt + inputDisplay + "\n\n")

	// Output
	if m.err != nil {
		b.WriteString(common.Error.Render(fmt.Sprintf("  Error: %v", m.err)))
		b.WriteString("\n")
	} else if m.output != "" {
		b.WriteString(common.Output.Render(m.output))
		b.WriteString("\n")
	} else {
		b.WriteString(common.Muted.Render(
			"  No directory loaded. Enter a path and press Enter.\n"))
	}
}

// renderProcessesTab renders the process manager interface.
func (m *Model) renderProcessesTab(b *strings.Builder) {
	b.WriteString(common.PanelTitle.Render("Process Manager"))
	b.WriteString("\n")
	b.WriteString(common.Muted.Render(
		"  Commands: list, kill:<pid>, restart:<pid>"))
	b.WriteString("\n\n")

	prompt := lipgloss.NewStyle().Foreground(lipgloss.Color(common.CurrentPalette().Danger)).Bold(true).Render("process> ")
	inputDisplay := m.input
	if inputDisplay == "" {
		inputDisplay = common.Muted.Render("list")
	} else {
		inputDisplay = lipgloss.NewStyle().Foreground(lipgloss.Color(common.CurrentPalette().Text)).Render(inputDisplay)
	}
	b.WriteString("  " + prompt + inputDisplay + "\n\n")

	if m.err != nil {
		b.WriteString(common.Error.Render(fmt.Sprintf("  Error: %v", m.err)))
		b.WriteString("\n")
	} else if m.output != "" {
		b.WriteString(common.Output.Render(m.output))
		b.WriteString("\n")
	} else {
		b.WriteString(common.Muted.Render(
			"  No process data loaded. Press Enter to list top processes.\n"))
	}
}

// renderServicesTab renders the service status dashboard.
func (m *Model) renderServicesTab(b *strings.Builder) {
	b.WriteString(common.PanelTitle.Render("Service Status"))
	b.WriteString("\n")
	b.WriteString(common.Muted.Render(
		"  Press Enter to refresh service status."))
	b.WriteString("\n\n")

	prompt := lipgloss.NewStyle().Foreground(lipgloss.Color(common.CurrentPalette().Secondary)).Bold(true).Render("services> ")
	inputDisplay := m.input
	if inputDisplay == "" {
		inputDisplay = common.Muted.Render("list")
	} else {
		inputDisplay = lipgloss.NewStyle().Foreground(lipgloss.Color(common.CurrentPalette().Text)).Render(inputDisplay)
	}
	b.WriteString("  " + prompt + inputDisplay + "\n\n")

	if m.err != nil {
		b.WriteString(common.Error.Render(fmt.Sprintf("  Error: %v", m.err)))
		b.WriteString("\n")
	} else if m.output != "" {
		b.WriteString(common.Output.Render(m.output))
		b.WriteString("\n")
	} else {
		b.WriteString(common.Muted.Render(
			"  No service data loaded. Press Enter to list services.\n"))
	}
}

// formatLines formats log lines with line numbers.
func formatLines(lines []string, path string) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("File: %s (%d lines)\n", path, len(lines)))
	b.WriteString(strings.Repeat("─", 40) + "\n")
	for i, line := range lines {
		b.WriteString(fmt.Sprintf("%5d | %s\n", i+1, line))
	}
	return b.String()
}

// formatFileEntries formats a directory listing.
func formatFileEntries(entries []FileEntry, path string) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Directory: %s (%d entries)\n", path, len(entries)))
	b.WriteString(strings.Repeat("─", 50) + "\n")
	b.WriteString(fmt.Sprintf("  %-30s %10s %s\n", "Name", "Size", "Modified"))
	b.WriteString("  " + strings.Repeat("─", 50) + "\n")
	for _, e := range entries {
		icon := "📄"
		if e.IsDir {
			icon = "📁"
		}
		name := e.Name
		if len(name) > 28 {
			name = name[:27] + "…"
		}
		b.WriteString(fmt.Sprintf("  %s %-30s %10s %s\n",
			icon, name, e.Size, e.ModTime.Format("2006-01-02 15:04")))
	}
	return b.String()
}

func formatProcesses(processes []ProcessEntry) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Processes: %d shown\n", len(processes)))
	b.WriteString(strings.Repeat("─", 76) + "\n")
	b.WriteString(fmt.Sprintf("  %-8s %-24s %8s %10s %s\n", "PID", "Name", "CPU%", "Memory", "Status"))
	b.WriteString("  " + strings.Repeat("─", 72) + "\n")
	for _, proc := range processes {
		status := proc.Status
		if status == "" {
			status = "unknown"
		}
		b.WriteString(fmt.Sprintf("  %-8d %-24s %8.1f %9.1fM %s\n",
			proc.PID,
			common.TruncateString(proc.Name, 23),
			proc.CPU,
			proc.Memory,
			status))
	}
	return b.String()
}

func formatServices(services []ServiceEntry) string {
	var b strings.Builder
	running := 0
	stopped := 0
	for _, service := range services {
		switch strings.ToLower(service.Status) {
		case "running", "active":
			running++
		case "stopped", "inactive", "failed":
			stopped++
		}
	}

	b.WriteString(fmt.Sprintf("Services: %d shown | Running: %d | Stopped/Inactive: %d\n", len(services), running, stopped))
	b.WriteString(strings.Repeat("─", 92) + "\n")
	b.WriteString(fmt.Sprintf("  %-28s %-12s %-12s %s\n", "Name", "Status", "Start", "Display Name"))
	b.WriteString("  " + strings.Repeat("─", 88) + "\n")
	for _, service := range services {
		b.WriteString(fmt.Sprintf("  %-28s %-12s %-12s %s\n",
			common.TruncateString(service.Name, 27),
			common.TruncateString(service.Status, 11),
			common.TruncateString(service.StartType, 11),
			common.TruncateString(service.DisplayName, 34)))
	}
	return b.String()
}
