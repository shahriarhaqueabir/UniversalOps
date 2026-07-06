package secops

import (
	"fmt"

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
			m.showReport = true
			m.workflowReport = "Error running security audit: " + msg.Err.Error()
		} else {
			m.workflowReport = msg.Report
			m.showReport = true
		}
		m.loading = false

	case FirewallResult:
		if msg.Err != nil {
			m.fwErr = msg.Err
		} else {
			m.firewallRules = msg.Rules
			m.fwErr = nil
		}
		m.checkReady()

	case UsersResult:
		if msg.Err != nil {
			m.usersErr = msg.Err
		} else {
			m.users = msg.Users
			m.groups = msg.Groups
			m.usersErr = nil
		}
		m.checkReady()

	case ListeningResult:
		if msg.Err != nil {
			m.lpErr = msg.Err
		} else {
			m.listeningPorts = msg.Ports
			m.lpErr = nil
		}
		m.checkReady()

	case DefenderResult:
		if msg.Err != nil {
			m.defErr = msg.Err
		} else {
			m.defenderStatus = msg.Status
			m.defErr = nil
		}
		m.checkReady()

	case TasksResult:
		if msg.Err != nil {
			m.tasksErr = msg.Err
		} else {
			m.scheduledTasks = msg.Tasks
			m.tasksErr = nil
		}
		m.checkReady()

	case EventsResult:
		if msg.Err != nil {
			m.eventsErr = msg.Err
		} else {
			m.securityEvents = msg.Events
			m.eventsErr = nil
		}
		m.checkReady()

	case tea.KeyPressMsg:
		return m.handleKeyPress(msg)
	}

	return nil
}

// checkReady sets the ready flag once we have at least some data or all fetches have errored.
func (m *Model) checkReady() {
	// We're ready if we have data or all sources have been tried
	if m.firewallRules != nil || m.fwErr != nil ||
		m.users != nil || m.usersErr != nil ||
		m.listeningPorts != nil || m.lpErr != nil ||
		m.defenderStatus != nil || m.defErr != nil ||
		m.scheduledTasks != nil || m.tasksErr != nil ||
		m.securityEvents != nil || m.eventsErr != nil {
		m.ready = true
		m.loading = false
	}
}

// handleKeyPress processes keyboard events for the SecOps view.
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
	case "r":
		return m.Init()
	case "R":
		if m.showReport {
			m.showReport = false
			return nil
		}
		m.loading = true
		return func() tea.Msg {
			report, err := RunSecurityAudit()
			if err != nil {
				return WorkflowResultMsg{Err: err}
			}
			return WorkflowResultMsg{Report: report.String()}
		}
	case "esc":
		return nil
	default:
		return nil
	}
	return nil
}

// String representation of the model (for debugging).
func (m *Model) String() string {
	if !m.ready {
		return "SecOps: loading..."
	}
	counts := fmt.Sprintf("FW=%d USR=%d PORTS=%d TASKS=%d EVENTS=%d",
		len(m.firewallRules), len(m.users), len(m.listeningPorts), len(m.scheduledTasks), len(m.securityEvents))
	if m.defenderStatus != nil {
		counts += " DEF=ok"
	} else {
		counts += " DEF=?"
	}
	return "SecOps: " + counts
}
