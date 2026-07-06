package secops

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
)

// View renders the SecOps dashboard.
func (m *Model) View(width, height int, fallback *common.SystemStats) string {
	var b strings.Builder

	// Show workflow report if active
	if m.showReport && m.workflowReport != "" {
		b.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color(common.CurrentPalette().Primary)).
			Bold(true).
			Render("📊 Security Audit Report"))
		b.WriteString("\n\n")
		b.WriteString(common.Output.Render(m.workflowReport))
		b.WriteString("\n")
		b.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color(common.CurrentPalette().Muted)).
			Render("[r] back to dashboard  [esc] back  [1-6] jump"))
		return common.Panel.Width(width - 4).Render(b.String())
	}

	// Title
	b.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color(common.CurrentPalette().Primary)).
		Bold(true).
		Render("🔒 Security Operations Dashboard"))
	b.WriteString("\n")

	// Tab bar
	tabs := []string{" Firewall ", " Users ", " Listening ", " Defender ", " Tasks ", " Events "}
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

	// Loading state
	if m.loading {
		b.WriteString(common.Muted.Render("Collecting data... Press [r] to refresh.\n"))
		return common.Panel.Width(width - 4).Render(b.String())
	}

	if !m.ready {
		b.WriteString(common.Muted.Render("Waiting for data... Press [r] to refresh.\n"))
		return common.Panel.Width(width - 4).Render(b.String())
	}

	// Render the active tab
	switch m.tabIndex {
	case 0:
		m.renderFirewallTab(&b)
	case 1:
		m.renderUsersTab(&b)
	case 2:
		m.renderListeningTab(&b)
	case 3:
		m.renderDefenderTab(&b)
	case 4:
		m.renderTasksTab(&b)
	case 5:
		m.renderEventsTab(&b)
	}

	// Help footer
	b.WriteString("\n")
	b.WriteString(common.Muted.Render("[Tab/←→] switch tabs  [r] refresh  [R] report  [esc] back  [1-6] jump"))

	return common.Panel.Width(width - 4).Render(b.String())
}

// ---- Tab 0: Firewall ----

func (m *Model) renderFirewallTab(b *strings.Builder) {
	if m.fwErr != nil {
		b.WriteString(common.Error.Render(fmt.Sprintf("Error loading firewall rules: %v", m.fwErr)))
		b.WriteString("\n")
	}

	b.WriteString(common.PanelTitle.Render("Windows Firewall Rules"))
	b.WriteString("\n")

	if m.firewallRules == nil {
		b.WriteString(common.Muted.Render("No firewall data available.\n"))
		return
	}

	if len(m.firewallRules) == 0 {
		b.WriteString(common.Muted.Render("No firewall rules found.\n"))
		return
	}

	// Header
	headerFmt := "  %-30s %-10s %-8s %-8s %-10s %-10s %s\n"
	b.WriteString(fmt.Sprintf(headerFmt, "Name", "Direction", "Action", "Protocol", "LocalPort", "RemotePort", "Enabled"))
	b.WriteString("  " + strings.Repeat("─", 100) + "\n")

	for _, r := range m.firewallRules {
		enabled := colorBool(r.Enabled, "Yes", "No")
		b.WriteString(fmt.Sprintf(headerFmt,
			common.TruncateString(r.Name, 29),
			r.Direction,
			r.Action,
			r.Protocol,
			r.LocalPort,
			r.RemotePort,
			enabled))
	}

	b.WriteString(fmt.Sprintf("\n  %s %d rules displayed (of %d total)",
		common.Muted.Render("Total:"),
		len(m.firewallRules),
		len(m.firewallRules)))
}

// ---- Tab 1: Users ----

func (m *Model) renderUsersTab(b *strings.Builder) {
	if m.usersErr != nil {
		b.WriteString(common.Error.Render(fmt.Sprintf("Error loading users: %v", m.usersErr)))
		b.WriteString("\n")
	}

	b.WriteString(common.PanelTitle.Render("User Accounts"))
	b.WriteString("\n")

	if m.users == nil {
		b.WriteString(common.Muted.Render("No user data available.\n"))
		return
	}

	if len(m.users) == 0 {
		b.WriteString(common.Muted.Render("No user accounts found.\n"))
		return
	}

	// Users table
	headerFmt := "  %-20s %-22s %-6s %-9s %s\n"
	b.WriteString(fmt.Sprintf(headerFmt, "Username", "Full Name", "Admin", "Enabled", "Group"))
	b.WriteString("  " + strings.Repeat("─", 80) + "\n")

	for _, u := range m.users {
		admin := ""
		if u.IsAdmin {
			admin = lipgloss.NewStyle().Foreground(lipgloss.Color(common.CurrentPalette().Secondary)).Bold(true).Render("✓ Admin")
		}
		enabled := colorBool(u.IsEnabled, "✓ Yes", "✗ No")
		b.WriteString(fmt.Sprintf(headerFmt,
			common.TruncateString(u.Username, 19),
			common.TruncateString(u.FullName, 21),
			admin,
			enabled,
			u.Group))
	}

	// Groups section
	b.WriteString("\n")
	b.WriteString(common.PanelTitle.Render("Security Groups"))
	b.WriteString("\n")

	if m.groups == nil {
		b.WriteString(common.Muted.Render("No group data available.\n"))
		return
	}

	for i, g := range m.groups {
		if i > 0 && i%4 == 0 {
			b.WriteString("\n")
		}
		sep := "  "
		if i%4 == 3 || i == len(m.groups)-1 {
			sep = "\n"
		}
		marker := "•"
		b.WriteString(fmt.Sprintf("  %s %s%s", marker, g, sep))
	}
}

// ---- Tab 2: Listening Ports ----

func (m *Model) renderListeningTab(b *strings.Builder) {
	if m.lpErr != nil {
		b.WriteString(common.Error.Render(fmt.Sprintf("Error loading listening ports: %v", m.lpErr)))
		b.WriteString("\n")
	}

	b.WriteString(common.PanelTitle.Render("Listening Ports"))
	b.WriteString("\n")

	if m.listeningPorts == nil {
		b.WriteString(common.Muted.Render("No listening port data available.\n"))
		return
	}

	if len(m.listeningPorts) == 0 {
		b.WriteString(common.Muted.Render("No listening ports found.\n"))
		return
	}

	// Count by protocol
	tcpCount := 0
	udpCount := 0
	for _, p := range m.listeningPorts {
		if p.Protocol == "TCP" {
			tcpCount++
		} else {
			udpCount++
		}
	}

	b.WriteString(fmt.Sprintf("  %s TCP: %d | UDP: %d\n\n",
		common.Muted.Render("Summary:"), tcpCount, udpCount))

	headerFmt := "  %-6s %-7s %-24s %-8s %s\n"
	b.WriteString(fmt.Sprintf(headerFmt, "Proto", "Port", "Process", "PID", "State"))
	b.WriteString("  " + strings.Repeat("─", 60) + "\n")

	for _, p := range m.listeningPorts {
		procName := p.ProcessName
		if procName == "" {
			procName = fmt.Sprintf("pid:%d", p.PID)
		}
		b.WriteString(fmt.Sprintf(headerFmt,
			p.Protocol,
			fmt.Sprintf("%d", p.Port),
			common.TruncateString(procName, 23),
			fmt.Sprintf("%d", p.PID),
			p.State))
	}
}

// ---- Tab 3: Defender ----

func (m *Model) renderDefenderTab(b *strings.Builder) {
	if m.defErr != nil {
		b.WriteString(common.Error.Render(fmt.Sprintf("Error loading Defender status: %v", m.defErr)))
		b.WriteString("\n")
	}

	b.WriteString(common.PanelTitle.Render("Windows Defender Status"))
	b.WriteString("\n")

	if m.defenderStatus == nil {
		b.WriteString(common.Muted.Render("No Defender data available.\n"))
		return
	}

	status := m.defenderStatus

	// Health overview
	b.WriteString(common.PanelTitle.Render("Health Overview"))
	b.WriteString("\n")

	healthItems := []struct {
		label string
		value string
		good  bool
	}{
		{"Antivirus Enabled", boolStr(status.Enabled), status.Enabled},
		{"Real-Time Protection", boolStr(status.RealTimeProtection), status.RealTimeProtection},
		{"Cloud-Delivered Protection", boolStr(status.CloudProtection), status.CloudProtection},
		{"Signatures Up-To-Date", boolStr(status.UpToDate), status.UpToDate},
	}

	for _, item := range healthItems {
		valColor := common.HealthyColor
		if !item.good {
			valColor = common.DangerColor
		}
		b.WriteString(fmt.Sprintf("  %-28s %s\n",
			common.Label.Render(item.label+":"),
			lipgloss.NewStyle().Foreground(valColor).Bold(true).Render(item.value)))
	}

	// Signature details
	b.WriteString("\n")
	b.WriteString(common.PanelTitle.Render("Signature & Scan Info"))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("  %-28s %s\n", common.Label.Render("Signature Age:"), common.Value.Render(status.SignatureAge)))
	b.WriteString(fmt.Sprintf("  %-28s %s\n", common.Label.Render("Last Scan:"), common.Value.Render(status.LastScan)))
	b.WriteString(fmt.Sprintf("  %-28s %d days\n", common.Label.Render("Quick Scan Age:"), status.QuickScanAge))
	b.WriteString(fmt.Sprintf("  %-28s %d days\n", common.Label.Render("Full Scan Age:"), status.FullScanAge))

	// Service status
	b.WriteString("\n")
	b.WriteString(common.PanelTitle.Render("Service Status"))
	b.WriteString("\n")

	serviceItems := []struct {
		label string
		val   bool
	}{
		{"Anti-Malware Service", status.AMServiceEnabled},
		{"Anti-Spyware Service", status.AntispywareEnabled},
		{"Network Inspection System", status.NISEnabled},
	}

	for _, s := range serviceItems {
		marker := "✓ Running"
		c := common.HealthyColor
		if !s.val {
			marker = "✗ Stopped"
			c = common.DangerColor
		}
		b.WriteString(fmt.Sprintf("  %-28s %s\n",
			common.Label.Render(s.label+":"),
			lipgloss.NewStyle().Foreground(c).Render(marker)))
	}
}

// ---- Tab 4: Scheduled Tasks ----

func (m *Model) renderTasksTab(b *strings.Builder) {
	if m.tasksErr != nil {
		b.WriteString(common.Error.Render(fmt.Sprintf("Error loading scheduled tasks: %v", m.tasksErr)))
		b.WriteString("\n")
	}

	b.WriteString(common.PanelTitle.Render("Scheduled Tasks"))
	b.WriteString("\n")

	if m.scheduledTasks == nil {
		b.WriteString(common.Muted.Render("No scheduled task data available.\n"))
		return
	}

	if len(m.scheduledTasks) == 0 {
		b.WriteString(common.Muted.Render("No scheduled tasks found.\n"))
		return
	}

	// Count by status
	running := 0
	ready := 0
	disabled := 0
	for _, t := range m.scheduledTasks {
		switch t.Status {
		case "Running":
			running++
		case "Disabled":
			disabled++
		default:
			ready++
		}
	}

	b.WriteString(fmt.Sprintf("  %s Running: %d | Ready: %d | Disabled: %d\n\n",
		common.Muted.Render("Summary:"), running, ready, disabled))

	headerFmt := "  %-40s %-10s %-18s %-18s %s\n"
	b.WriteString(fmt.Sprintf(headerFmt, "Task Name", "Status", "Next Run", "Last Run", "Trigger"))
	b.WriteString("  " + strings.Repeat("─", 110) + "\n")

	for _, t := range m.scheduledTasks {
		statusColor := common.HealthyColor
		if t.Status == "Disabled" {
			statusColor = common.DangerColor
		} else if t.Status == "Running" {
			statusColor = common.WarningColor
		}
		statusStr := lipgloss.NewStyle().Foreground(statusColor).Render(t.Status)

		b.WriteString(fmt.Sprintf(headerFmt,
			common.TruncateString(t.Name, 39),
			statusStr,
			t.NextRun,
			t.LastRun,
			t.Trigger))
	}
}

// ---- Tab 5: Security Events ----

func (m *Model) renderEventsTab(b *strings.Builder) {
	if m.eventsErr != nil {
		b.WriteString(common.Error.Render(fmt.Sprintf("Error loading security events: %v", m.eventsErr)))
		b.WriteString("\n")
	}

	b.WriteString(common.PanelTitle.Render("Security Events"))
	b.WriteString("\n")

	if m.securityEvents == nil {
		b.WriteString(common.Muted.Render("No security event data available.\n"))
		return
	}

	if len(m.securityEvents) == 0 {
		b.WriteString(common.Muted.Render("No security events found.\n"))
		return
	}

	important := 0
	for _, event := range m.securityEvents {
		if event.Important {
			important++
		}
	}
	b.WriteString(fmt.Sprintf("  %s Recent: %d | Important: %d\n\n",
		common.Muted.Render("Summary:"), len(m.securityEvents), important))

	headerFmt := "  %-7s %-10s %-19s %-24s %s\n"
	b.WriteString(fmt.Sprintf(headerFmt, "Event", "Level", "Time", "Provider", "Message"))
	b.WriteString("  " + strings.Repeat("─", 105) + "\n")

	for _, event := range m.securityEvents {
		level := event.Level
		if level == "" {
			level = "Info"
		}
		levelColor := common.HealthyColor
		if event.Important {
			levelColor = common.WarningColor
		}
		b.WriteString(fmt.Sprintf(headerFmt,
			fmt.Sprintf("%d", event.ID),
			lipgloss.NewStyle().Foreground(levelColor).Render(common.TruncateString(level, 9)),
			common.TruncateString(event.Time, 18),
			common.TruncateString(event.Provider, 23),
			common.TruncateString(event.Message, 50)))
	}
}

// ---- Helpers ----

// boolStr returns "Yes" or "No" for a boolean value.
func boolStr(v bool) string {
	if v {
		return "Yes"
	}
	return "No"
}

// colorBool returns a colored string based on a boolean value.
func colorBool(v bool, trueStr, falseStr string) string {
	if v {
		return lipgloss.NewStyle().Foreground(lipgloss.Color(common.CurrentPalette().Secondary)).Render(trueStr)
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color(common.CurrentPalette().Danger)).Render(falseStr)
}
