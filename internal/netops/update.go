package netops

import (
	"fmt"
	"time"

	tea "charm.land/bubbletea/v2"
)

// WorkflowResultMsg is sent when a workflow completes.
type WorkflowResultMsg struct {
	Report string
	Err    error
}

// Update handles all messages and delegates to the appropriate handler.
func (m *Model) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case WorkflowResultMsg:
		if msg.Err != nil {
			m.err = msg.Err
			m.showReport = true
			m.workflowReport = "Error running network diagnostics: " + msg.Err.Error()
			m.ready = true
		} else {
			m.workflowReport = msg.Report
			m.showReport = true
			m.ready = true
		}

	case PingResultMsg:
		if msg.Err != nil {
			m.err = msg.Err
		} else {
			m.PingResult = msg.Result
			m.err = nil
			m.ready = true
		}

	case DNSResultMsg:
		if msg.Err != nil {
			m.err = msg.Err
		} else {
			m.DNSResult = msg.Result
			m.err = nil
			m.ready = true
		}

	case PortScanResultMsg:
		if msg.Err != nil {
			m.err = msg.Err
		} else {
			m.PortResults = msg.Results
			m.err = nil
			m.ready = true
		}

	case TraceRouteResultMsg:
		if msg.Err != nil {
			m.err = msg.Err
		} else {
			m.TraceResult = msg.Result
			m.err = nil
			m.ready = true
		}

	case ConnectionsResultMsg:
		if msg.Err != nil {
			m.err = msg.Err
		} else {
			m.Connections = msg.Connections
			m.err = nil
			m.ready = true
		}

	case InterfacesResultMsg:
		if msg.Err != nil {
			m.err = msg.Err
		} else {
			m.InterfaceData = mergeInterfaceBandwidthHistory(m.InterfaceData, msg.Interfaces)
			m.lastCounters = msg.Counters
			m.lastCapture = time.Now()
			m.err = nil
			m.ready = true
		}

	case tea.KeyPressMsg:
		return m.handleKeyPress(msg)
	}

	return nil
}

// handleKeyPress processes keyboard events for the NetOps view.
func (m *Model) handleKeyPress(msg tea.KeyPressMsg) tea.Cmd {
	switch msg.String() {
	case "tab", "l", "right":
		m.tabIndex = (m.tabIndex + 1) % 6
	case "shift+tab", "h", "left":
		m.tabIndex = (m.tabIndex - 1 + 6) % 6
	case "1":
		m.tabIndex = 0
	case "2":
		m.tabIndex = 1
	case "3":
		m.tabIndex = 2
	case "4":
		m.tabIndex = 3
	case "5":
		m.tabIndex = 4
	case "6":
		m.tabIndex = 5
	case "up", "k":
		if m.tabIndex == 5 && m.selectedIndex > 0 {
			m.selectedIndex--
		}
	case "down", "j":
		if m.tabIndex == 5 && m.selectedIndex < len(m.InterfaceData)-1 {
			m.selectedIndex++
		}
	case "r":
		return m.refreshCurrentTab()
	case "R":
		if m.showReport {
			m.showReport = false
			return nil
		}
		m.ready = false
		return func() tea.Msg {
			report, err := RunNetworkDiagnostics()
			if err != nil {
				return WorkflowResultMsg{Err: err}
			}
			return WorkflowResultMsg{Report: report.String()}
		}
	case "enter":
		return m.actOnCurrentTab()
	case "esc":
		return nil
	default:
		return nil
	}
	return nil
}

// refreshCurrentTab refreshes data for the current tab.
func (m *Model) refreshCurrentTab() tea.Cmd {
	switch m.tabIndex {
	case 0: // Ping
		if m.pingTarget != "" {
			return func() tea.Msg {
				result, err := Ping(m.pingTarget, m.pingCount)
				return PingResultMsg{Result: result, Err: err}
			}
		}
	case 1: // DNS
		if m.pingTarget != "" {
			return func() tea.Msg {
				result, err := LookupDNS(m.pingTarget)
				return DNSResultMsg{Result: result, Err: err}
			}
		}
	case 2: // Port Scan
		if m.pingTarget != "" {
			return func() tea.Msg {
				results, err := ScanCommonPorts(m.pingTarget)
				return PortScanResultMsg{Results: results, Err: err}
			}
		}
	case 3: // Traceroute
		if m.pingTarget != "" {
			return func() tea.Msg {
				result, err := TraceRoute(m.pingTarget)
				return TraceRouteResultMsg{Result: result, Err: err}
			}
		}
	case 4: // Connections
		return func() tea.Msg {
			conns, err := GetConnections()
			return ConnectionsResultMsg{Connections: conns, Err: err}
		}
	case 5: // Interfaces
		last := m.lastCounters
		elapsed := time.Since(m.lastCapture)
		return func() tea.Msg {
			res, err := GetInterfaces(last, elapsed)
			return InterfacesResultMsg{Interfaces: res.Interfaces, Counters: res.Counters, Err: err}
		}
	}
	return nil
}

// actOnCurrentTab performs the primary action for the current tab.
func (m *Model) actOnCurrentTab() tea.Cmd {
	switch m.tabIndex {
	case 0: // Ping
		if m.pingTarget == "" {
			m.pingTarget = "8.8.8.8"
		}
		return func() tea.Msg {
			result, err := Ping(m.pingTarget, m.pingCount)
			return PingResultMsg{Result: result, Err: err}
		}
	case 1: // DNS
		if m.pingTarget == "" {
			m.pingTarget = "google.com"
		}
		return func() tea.Msg {
			result, err := LookupDNS(m.pingTarget)
			return DNSResultMsg{Result: result, Err: err}
		}
	case 2: // Port Scan
		if m.pingTarget == "" {
			m.pingTarget = "localhost"
		}
		return func() tea.Msg {
			results, err := ScanCommonPorts(m.pingTarget)
			return PortScanResultMsg{Results: results, Err: err}
		}
	case 3: // Traceroute
		if m.pingTarget == "" {
			m.pingTarget = "8.8.8.8"
		}
		return func() tea.Msg {
			result, err := TraceRoute(m.pingTarget)
			return TraceRouteResultMsg{Result: result, Err: err}
		}
	case 4: // Connections
		return func() tea.Msg {
			conns, err := GetConnections()
			return ConnectionsResultMsg{Connections: conns, Err: err}
		}
	case 5: // Interfaces
		last := m.lastCounters
		elapsed := time.Since(m.lastCapture)
		return func() tea.Msg {
			res, err := GetInterfaces(last, elapsed)
			return InterfacesResultMsg{Interfaces: res.Interfaces, Counters: res.Counters, Err: err}
		}
	}
	return nil
}

// String representation of the model (for debugging).
func (m *Model) String() string {
	if m.PingResult == nil && m.DNSResult == nil && len(m.PortResults) == 0 && m.TraceResult == nil {
		return "NetOps: no data"
	}
	return fmt.Sprintf("NetOps: tab=%d", m.tabIndex)
}
