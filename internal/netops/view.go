package netops

import (
	"fmt"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
)

// View renders the NetOps dashboard.
func (m *Model) View(width, height int) string {
	var b strings.Builder

	// Show workflow report if active
	if m.showReport && m.workflowReport != "" {
		b.WriteString(common.PanelTitle.Render("📊 Network Diagnostic Report"))
		b.WriteString("\n\n")
		b.WriteString(common.Output.Render(m.workflowReport))
		b.WriteString("\n")
		b.WriteString(common.Muted.Render("[r] back to dashboard  [esc] back  [1-5] jump"))
		return common.Panel.Width(width - 4).Render(b.String())
	}

	// Title
	b.WriteString(common.PanelTitle.Render("🌐 Network Operations Dashboard"))
	b.WriteString("\n")

	// Tab bar
	tabs := []string{" Ping ", " DNS ", " Port Scan ", " Traceroute ", " Connections ", " Interfaces "}
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

	// Render the active tab
	switch m.tabIndex {
	case 0:
		m.renderPingTab(&b)
	case 1:
		m.renderDNSTab(&b)
	case 2:
		m.renderPortScanTab(&b)
	case 3:
		m.renderTracerouteTab(&b)
	case 4:
		m.renderConnectionsTab(&b)
	case 5:
		m.renderInterfacesTab(&b)
	}

	// Help footer
	b.WriteString("\n")
	b.WriteString(common.Muted.Render("[Tab/←→] switch tabs  [enter] run  [r] refresh  [R] report  [esc] back  [1-6] jump"))

	return common.Panel.Width(width - 4).Render(b.String())
}

func (m *Model) renderPingTab(b *strings.Builder) {
	target := m.pingTarget
	if target == "" {
		target = "8.8.8.8"
	}

	b.WriteString(common.PanelTitle.Render("Ping"))
	b.WriteString("\n\n")

	b.WriteString(fmt.Sprintf("  Target: %s\n",
		common.Value.Render(target)))
	b.WriteString(fmt.Sprintf("  Count:  %d\n\n", m.pingCount))

	if m.err != nil && m.PingResult == nil {
		b.WriteString(fmt.Sprintf("  Error: %v\n", m.err))
		return
	}

	if m.PingResult == nil {
		b.WriteString(common.Muted.Render(
			"  Press [enter] to ping, or type a target and press [enter].\n  Set target via the command line or edit in the model."))
		return
	}

	r := m.PingResult

	// Stats table
	b.WriteString(fmt.Sprintf("  %-12s %s\n",
		common.Label.Render("Host:"),
		common.Value.Render(r.Target)))
	b.WriteString(fmt.Sprintf("  %-12s %s\n",
		common.Label.Render("IP:"),
		common.Value.Render(r.IP)))
	b.WriteString("\n")

	// Packet loss
	lossColor := common.GetHealthColor(float64(r.Lost) / float64(max(r.Sent, 1)) * 100)
	b.WriteString(fmt.Sprintf("  %-12s %s\n",
		common.Label.Render("Sent:"),
		common.Value.Render(fmt.Sprintf("%d", r.Sent))))
	b.WriteString(fmt.Sprintf("  %-12s %s\n",
		common.Label.Render("Received:"),
		common.Value.Render(fmt.Sprintf("%d", r.Received))))
	b.WriteString(fmt.Sprintf("  %-12s %s\n",
		common.Label.Render("Lost:"),
		lipgloss.NewStyle().Foreground(lossColor).Render(fmt.Sprintf("%d (%.0f%%)", r.Lost, percent(r.Lost, r.Sent)))))
	b.WriteString("\n")

	// RTT statistics
	b.WriteString(fmt.Sprintf("  %-12s %s\n",
		common.Label.Render("Min RTT:"),
		common.Value.Render(r.Min.String())))
	b.WriteString(fmt.Sprintf("  %-12s %s\n",
		common.Label.Render("Max RTT:"),
		common.Value.Render(r.Max.String())))
	b.WriteString(fmt.Sprintf("  %-12s %s\n",
		common.Label.Render("Avg RTT:"),
		common.Value.Render(r.Avg.String())))
	if r.TTL > 0 {
		b.WriteString(fmt.Sprintf("  %-12s %s\n",
			common.Label.Render("TTL:"),
			common.Value.Render(fmt.Sprintf("%d", r.TTL))))
	}
}

func (m *Model) renderDNSTab(b *strings.Builder) {
	hostname := m.pingTarget
	if hostname == "" {
		hostname = "google.com"
	}

	b.WriteString(common.PanelTitle.Render("DNS Lookup"))
	b.WriteString("\n\n")

	b.WriteString(fmt.Sprintf("  Hostname: %s\n\n",
		common.Value.Render(hostname)))

	if m.DNSResult == nil {
		b.WriteString(common.Muted.Render(
			"  Press [enter] to perform DNS lookup.\n  Results will be displayed here."))
		return
	}

	r := m.DNSResult

	if r.Error != "" {
		b.WriteString(fmt.Sprintf("  %-12s %s\n",
			common.Label.Render("Error:"),
			common.Error.Render(r.Error)))
		return
	}

	// A Records
	if len(r.A) > 0 {
		b.WriteString(fmt.Sprintf("%s\n", common.Section.Render("A Records (IPv4)")))
		for _, ip := range r.A {
			b.WriteString(fmt.Sprintf("  • %s\n", common.Value.Render(ip)))
		}
		b.WriteString("\n")
	}

	// AAAA Records
	if len(r.AAAA) > 0 {
		b.WriteString(fmt.Sprintf("%s\n", common.Section.Render("AAAA Records (IPv6)")))
		for _, ip := range r.AAAA {
			b.WriteString(fmt.Sprintf("  • %s\n", common.Value.Render(ip)))
		}
		b.WriteString("\n")
	}

	// MX Records
	if len(r.MX) > 0 {
		b.WriteString(fmt.Sprintf("%s\n", common.Section.Render("MX Records (Mail)")))
		for _, mx := range r.MX {
			b.WriteString(fmt.Sprintf("  • %s\n", common.Value.Render(mx)))
		}
		b.WriteString("\n")
	}

	// NS Records
	if len(r.NS) > 0 {
		b.WriteString(fmt.Sprintf("%s\n", common.Section.Render("NS Records (Nameservers)")))
		for _, ns := range r.NS {
			b.WriteString(fmt.Sprintf("  • %s\n", common.Value.Render(ns)))
		}
		b.WriteString("\n")
	}

	// CNAME
	if r.CNAME != "" {
		b.WriteString(fmt.Sprintf("  %-12s %s\n",
			common.Label.Render("CNAME:"),
			common.Value.Render(r.CNAME)))
		b.WriteString("\n")
	}

	// TXT Records
	if len(r.TXT) > 0 {
		b.WriteString(fmt.Sprintf("%s\n", common.Section.Render("TXT Records")))
		maxLen := 60
		for _, txt := range r.TXT {
			display := txt
			if len(display) > maxLen {
				display = display[:maxLen-3] + "..."
			}
			b.WriteString(fmt.Sprintf("  • %s\n", common.Value.Render(display)))
		}
		b.WriteString("\n")
	}
}

func (m *Model) renderPortScanTab(b *strings.Builder) {
	host := m.pingTarget
	if host == "" {
		host = "localhost"
	}

	b.WriteString(common.PanelTitle.Render("Port Scan"))
	b.WriteString("\n\n")

	b.WriteString(fmt.Sprintf("  Target: %s\n\n",
		common.Value.Render(host)))

	if len(m.PortResults) == 0 {
		b.WriteString(common.Muted.Render(
			"  Press [enter] to scan common ports.\n  Scans ~25 well-known ports with a 500ms timeout per port."))
		return
	}

	// Table header
	b.WriteString(fmt.Sprintf("  %-6s %-5s %-16s\n",
		common.Label.Render("Port"),
		common.Label.Render("State"),
		common.Label.Render("Service")))
	b.WriteString("  " + strings.Repeat("─", 30) + "\n")

	// Count open vs closed
	openCount := 0
	for _, r := range m.PortResults {
		state := "closed"
		stateColor := lipgloss.Color(common.CurrentPalette().Muted)
		if r.Open {
			state = "open"
			stateColor = lipgloss.Color(common.CurrentPalette().Secondary)
			openCount++
		}
		b.WriteString(fmt.Sprintf("  %-6d %s %-16s\n",
			r.Port,
			lipgloss.NewStyle().Foreground(stateColor).Render(fmt.Sprintf("%-5s", state)),
			common.Value.Render(r.Service)))
	}

	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("  %s %d/%d ports open\n",
		common.Value.Render("Summary:"),
		openCount, len(m.PortResults)))
}

func (m *Model) renderConnectionsTab(b *strings.Builder) {
	b.WriteString(common.PanelTitle.Render("Network Connections"))
	b.WriteString("\n\n")

	if len(m.Connections) == 0 {
		b.WriteString(common.Muted.Render(
			"  Press [enter] to fetch network connections.\n  Uses netstat -ano on Windows."))
		return
	}

	// Summary stats
	listeningCount := 0
	establishedCount := 0
	timeWaitCount := 0
	for _, c := range m.Connections {
		switch c.State {
		case "LISTENING", "LISTEN":
			listeningCount++
		case "ESTABLISHED":
			establishedCount++
		case "TIME_WAIT":
			timeWaitCount++
		}
	}

	b.WriteString(fmt.Sprintf("  Total: %d | Listening: %s | Established: %s | TimeWait: %s\n\n",
		len(m.Connections),
		lipgloss.NewStyle().Foreground(lipgloss.Color(common.CurrentPalette().Secondary)).Render(fmt.Sprintf("%d", listeningCount)),
		lipgloss.NewStyle().Foreground(lipgloss.Color(common.CurrentPalette().Info)).Render(fmt.Sprintf("%d", establishedCount)),
		lipgloss.NewStyle().Foreground(lipgloss.Color(common.CurrentPalette().Muted)).Render(fmt.Sprintf("%d", timeWaitCount))))

	// Table header
	b.WriteString(fmt.Sprintf("  %-22s %-22s %-11s %-7s\n",
		common.Label.Render("Local"),
		common.Label.Render("Remote"),
		common.Label.Render("State"),
		common.Label.Render("PID")))
	b.WriteString("  " + strings.Repeat("─", 65) + "\n")

	// Show up to first 20 connections
	maxDisplay := 20
	displayed := 0
	for _, c := range m.Connections {
		if displayed >= maxDisplay {
			b.WriteString(fmt.Sprintf("  ... and %d more\n", len(m.Connections)-maxDisplay))
			break
		}

		stateColor := lipgloss.Color(common.CurrentPalette().Muted)
		switch c.State {
		case "LISTENING", "LISTEN":
			stateColor = lipgloss.Color(common.CurrentPalette().Secondary)
		case "ESTABLISHED":
			stateColor = lipgloss.Color(common.CurrentPalette().Info)
		case "TIME_WAIT":
			stateColor = lipgloss.Color(common.CurrentPalette().Warning)
		case "CLOSE_WAIT":
			stateColor = lipgloss.Color(common.CurrentPalette().Danger)
		}

		localStr := c.LocalAddr
		remoteStr := c.RemoteAddr
		if len(localStr) > 21 {
			localStr = localStr[:18] + "..."
		}
		if len(remoteStr) > 21 {
			remoteStr = remoteStr[:18] + "..."
		}

		b.WriteString(fmt.Sprintf("  %-22s %-22s %s %d\n",
			localStr,
			remoteStr,
			lipgloss.NewStyle().Foreground(stateColor).Render(fmt.Sprintf("%-11s", c.State)),
			c.PID))
		displayed++
	}
}

func (m *Model) renderTracerouteTab(b *strings.Builder) {
	target := m.pingTarget
	if target == "" {
		target = "8.8.8.8"
	}

	b.WriteString(common.PanelTitle.Render("Traceroute"))
	b.WriteString("\n\n")

	b.WriteString(fmt.Sprintf("  Target: %s\n\n",
		common.Value.Render(target)))

	if m.TraceResult == nil {
		b.WriteString(common.Muted.Render(
			"  Press [enter] to trace the network path.\n  Uses tracert on Windows and traceroute on Unix-like systems."))
		return
	}

	if len(m.TraceResult.Hops) == 0 {
		b.WriteString(common.Muted.Render(
			"  No hops parsed from traceroute output."))
		return
	}

	b.WriteString(fmt.Sprintf("  %-5s %-18s %-28s %s\n",
		common.Label.Render("Hop"),
		common.Label.Render("Address"),
		common.Label.Render("Host"),
		common.Label.Render("RTT")))
	b.WriteString("  " + strings.Repeat("─", 70) + "\n")

	for _, hop := range m.TraceResult.Hops {
		host := hop.Host
		if host == "" {
			host = "-"
		}
		ip := hop.IP
		if ip == "" && hop.Timed {
			ip = "*"
		}
		if ip == "" {
			ip = "-"
		}
		b.WriteString(fmt.Sprintf("  %-5d %-18s %-28s %s\n",
			hop.Number,
			ip,
			truncate(host, 27),
			formatRTTs(hop.RTTs, hop.Timed)))
	}
}

func (m *Model) renderInterfacesTab(b *strings.Builder) {
	b.WriteString(common.PanelTitle.Render("Network Interfaces"))
	b.WriteString("\n\n")

	if len(m.InterfaceData) == 0 {
		b.WriteString(common.Muted.Render(
			"  Press [enter] to fetch network interface information."))
		return
	}

	// List of interfaces (Compact)
	for i, iface := range m.InterfaceData {
		statusStr := "up"
		if !iface.IsUp {
			statusStr = "down"
		}

		prefix := "  "
		style := lipgloss.NewStyle().PaddingLeft(2)
		if i == m.selectedIndex {
			prefix = "> "
			style = style.
				Background(lipgloss.Color(common.CurrentPalette().Primary)).
				Foreground(lipgloss.Color(common.CurrentPalette().Background)).
				Bold(true)
		}

		namePart := fmt.Sprintf("%-15s", iface.Name)
		ipPart := "no IP"
		if len(iface.IPs) > 0 {
			ipPart = iface.IPs[0]
		}
		b.WriteString(prefix + style.Render(fmt.Sprintf("%s [%-4s] %-20s",
			namePart, statusStr, ipPart)))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(m.renderInterfacesDetail(60))
}

func (m *Model) renderInterfacesDetail(width int) string {
	if m.selectedIndex >= len(m.InterfaceData) {
		return common.Muted.Render("  Select an interface to see details")
	}
	iface := m.InterfaceData[m.selectedIndex]

	// Left: Stats
	var stats strings.Builder
	stats.WriteString(fmt.Sprintf("%s %s\n", common.Label.Render("MAC:"), common.Value.Render(iface.MAC)))
	stats.WriteString(fmt.Sprintf("%s %d\n", common.Label.Render("MTU:"), iface.MTU))
	stats.WriteString(fmt.Sprintf("%s %s\n", common.Label.Render("Flags:"), common.Muted.Render(iface.Flags)))
	stats.WriteString(fmt.Sprintf("%s %s\n", common.Label.Render("Speed:"), common.Value.Render(iface.Speed)))

	if len(iface.IPs) > 1 {
		stats.WriteString(common.Label.Render("Other IPs:") + "\n")
		for i := 1; i < len(iface.IPs); i++ {
			stats.WriteString("  • " + iface.IPs[i] + "\n")
		}
	}

	// Right: Sparklines and Rates
	var trends strings.Builder
	trends.WriteString(fmt.Sprintf("%s %s/s\n", common.Label.Render("RX:"), common.Value.Render(formatRate(iface.RXRateBps))))
	trends.WriteString(lipgloss.NewStyle().Foreground(common.InfoColor).Render(common.RenderSparkline(iface.RXHistory, width/2)) + "\n\n")
	trends.WriteString(fmt.Sprintf("%s %s/s\n", common.Label.Render("TX:"), common.Value.Render(formatRate(iface.TXRateBps))))
	trends.WriteString(lipgloss.NewStyle().Foreground(common.HealthyColor).Render(common.RenderSparkline(iface.TXHistory, width/2)) + "\n")

	return common.Panel.Width(width).Render(
		lipgloss.JoinHorizontal(lipgloss.Top,
			lipgloss.NewStyle().Width(width/2).Render(stats.String()),
			lipgloss.NewStyle().Width(width/2).Render(trends.String()),
		),
	)
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

func formatRate(bytesPerSecond float64) string {
	if bytesPerSecond <= 0 {
		return "0 B"
	}
	return common.FormatBytes(uint64(bytesPerSecond))
}

// percent calculates a percentage.
func percent(value, total int) float64 {
	if total == 0 {
		return 0
	}
	return float64(value) / float64(total) * 100
}

// max returns the larger of two integers.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func formatRTTs(rtts []time.Duration, timed bool) string {
	if len(rtts) == 0 {
		if timed {
			return "*"
		}
		return "-"
	}
	parts := make([]string, 0, len(rtts))
	for _, rtt := range rtts {
		parts = append(parts, fmt.Sprintf("%dms", rtt.Milliseconds()))
	}
	return strings.Join(parts, " / ")
}

func truncate(value string, maxLen int) string {
	if len(value) <= maxLen {
		return value
	}
	if maxLen <= 3 {
		return value[:maxLen]
	}
	return value[:maxLen-3] + "..."
}
