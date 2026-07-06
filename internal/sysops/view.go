package sysops

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
)

// View renders the SysOps dashboard.
func (m *Model) View(width, height int, fallback *common.SystemStats) string {
	// Use fallback stats if we haven't collected our own yet
	stats := m.stats
	if stats == nil {
		stats = fallback
	}

	var b strings.Builder

	// Show workflow report if active
	if m.showReport && m.workflowReport != "" {
		b.WriteString(common.PanelTitle.Render("📊 System Health Report"))
		b.WriteString("\n\n")
		b.WriteString(common.Output.Render(m.workflowReport))
		b.WriteString("\n")
		b.WriteString(common.Muted.Render("[r] back to dashboard  [esc] back  [1-5] jump"))
		return common.Panel.Width(width - 4).Render(b.String())
	}

	// Title
	b.WriteString(common.PanelTitle.Render("🖥 System Operations Dashboard"))
	b.WriteString("\n")

	// Tab bar
	tabs := []string{" Overview ", " Processes ", " System Info "}
	for i, tab := range tabs {
		tabStyle := lipgloss.NewStyle().Padding(0, 1)
		if i == m.tabIndex {
			tabStyle = tabStyle.
				Foreground(lipgloss.Color(common.CurrentPalette().Primary)).
				Bold(true).
				Border(lipgloss.NormalBorder(), false, false, true, false).
				BorderForeground(lipgloss.Color(common.CurrentPalette().Primary))
		} else {
			tabStyle = tabStyle.Foreground(lipgloss.Color(common.CurrentPalette().Muted))
		}
		b.WriteString(tabStyle.Render(fmt.Sprintf("[%d]%s", i+1, tab)))
	}
	b.WriteString("\n\n")

	// Determine which tab to show
	tab := m.tabIndex
	if stats == nil {
		b.WriteString("Collecting data... Press [r] to refresh.\n")
		return common.Panel.Width(width - 4).Render(b.String())
	}

	switch tab {
	case 0:
		m.renderOverview(&b, stats)
	case 1:
		m.renderProcesses(&b)
	case 2:
		m.renderSystemInfo(&b)
	}

	// Help footer
	b.WriteString("\n")
	b.WriteString(common.Muted.Render("[Tab/←→] switch tabs  [r] refresh  [R] report  [esc] back  [1-5] jump"))

	return common.Panel.Width(width - 4).Render(b.String())
}

func (m *Model) renderOverview(b *strings.Builder, stats *common.SystemStats) {
	// CPU section
	b.WriteString(common.PanelTitle.Render("CPU"))
	b.WriteString("\n")
	cpuBar := renderBar(stats.CPUPercent, 100, 30)
	cpuColor := common.GetHealthColor(stats.CPUPercent)
	b.WriteString(fmt.Sprintf("  Usage:   %s %s %s\n",
		lipgloss.NewStyle().Foreground(cpuColor).Render(cpuBar),
		common.FormatPercent(stats.CPUPercent),
		common.RenderSparkline(stats.CPUHistory, 12)))
	b.WriteString("\n")

	// Memory section
	b.WriteString(common.PanelTitle.Render("Memory"))
	b.WriteString("\n")
	memBar := renderBar(stats.MemoryUsed, 100, 30)
	memColor := common.GetHealthColor(stats.MemoryUsed)
	b.WriteString(fmt.Sprintf("  Usage:   %s %s %s\n",
		lipgloss.NewStyle().Foreground(memColor).Render(memBar),
		common.FormatPercent(stats.MemoryUsed),
		common.RenderSparkline(stats.MemHistory, 12)))
	b.WriteString(fmt.Sprintf("  Total:   %s\n", common.FormatBytes(stats.MemoryTotal)))
	b.WriteString(fmt.Sprintf("  Used:    %s\n", common.FormatBytes(uint64(stats.MemoryUsedGB*1024*1024*1024))))
	b.WriteString("\n")

	// Disk section
	b.WriteString(common.PanelTitle.Render("Disk"))
	b.WriteString("\n")
	diskBar := renderBar(stats.DiskUsed, 100, 30)
	diskColor := common.GetHealthColor(stats.DiskUsed)
	b.WriteString(fmt.Sprintf("  Usage:   %s %s\n",
		lipgloss.NewStyle().Foreground(diskColor).Render(diskBar),
		common.FormatPercent(stats.DiskUsed)))
	b.WriteString(fmt.Sprintf("  Free:    %s\n", common.FormatBytes(stats.DiskFree)))
	b.WriteString("\n")

	// Quick summary
	b.WriteString(common.PanelTitle.Render("Summary"))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("  Processes: %s\n",
		common.Value.Render(
			fmt.Sprintf("%d", stats.ProcessCount))))
	b.WriteString(fmt.Sprintf("  Uptime:    %s\n",
		lipgloss.NewStyle().Foreground(lipgloss.Color(common.CurrentPalette().Info)).Render(stats.Uptime)))
}

func (m *Model) renderProcesses(b *strings.Builder) {
	procs, err := GetTopProcesses(15)
	if err != nil {
		b.WriteString(fmt.Sprintf("Error: %v\n", err))
		return
	}

	b.WriteString(common.PanelTitle.Render("Top Processes (by CPU)"))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("  %-6s %-25s %8s %8s %s\n",
		"PID", "Name", "CPU%", "MEM(MB)", "Status"))
	b.WriteString("  " + strings.Repeat("─", 65) + "\n")

	for _, p := range procs {
		cpuStr := fmt.Sprintf("%.1f", p.CPU)
		memStr := fmt.Sprintf("%.0f", p.Memory)
		status := p.Status
		if status == "" {
			status = "?"
		}

		b.WriteString(fmt.Sprintf("  %-6d %-25s %8s %8s %s\n",
			p.PID, common.TruncateString(p.Name, 24), cpuStr, memStr, status))
	}
}

func (m *Model) renderSystemInfo(b *strings.Builder) {
	info, err := GetSystemInfo()
	if err != nil {
		b.WriteString(fmt.Sprintf("Error: %v\n", err))
		return
	}

	b.WriteString(common.PanelTitle.Render("System Information"))
	b.WriteString("\n")

	fields := []struct {
		label string
		value string
	}{
		{"Hostname", info.Hostname},
		{"OS", info.OS},
		{"Platform", fmt.Sprintf("%s %s", info.Platform, info.PlatformVersion)},
		{"Kernel", fmt.Sprintf("%s (%s)", info.KernelVersion, info.KernelArch)},
		{"Uptime", common.FormatUptime(info.UptimeSeconds)},
		{"Processes", fmt.Sprintf("%d", info.ProcessCount)},
	}

	if info.Virtualization != "" {
		fields = append(fields, struct{ label, value string }{"Virtualization", info.Virtualization})
	}

	for _, f := range fields {
		b.WriteString(fmt.Sprintf("  %-15s %s\n",
			common.Label.Render(f.label+":"),
			common.Value.Render(f.value)))
	}

	b.WriteString("\n")
	b.WriteString(common.Muted.Render(
		"  System information collected via gopsutil (cross-platform)"))
}

// renderBar renders a simple ASCII progress bar.
func renderBar(value, max, width float64) string {
	if max == 0 {
		return strings.Repeat("░", int(width))
	}
	filled := int((value / max) * width)
	if filled > int(width) {
		filled = int(width)
	}
	empty := int(width) - filled
	return strings.Repeat("█", filled) + strings.Repeat("░", empty)
}
