package sysops

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
	case StatsResult:
		if msg.Err != nil {
			m.err = msg.Err
		} else {
			m.stats = msg.Stats
			m.err = nil
			m.ready = true
		}

	case WorkflowResultMsg:
		if msg.Err != nil {
			m.err = msg.Err
			m.showReport = true
			m.workflowReport = "Error running health check: " + msg.Err.Error()
			m.ready = true
		} else {
			m.workflowReport = msg.Report
			m.showReport = true
			m.ready = true
		}

	case tea.KeyPressMsg:
		return m.handleKeyPress(msg)
	}

	return nil
}

// handleKeyPress processes keyboard events for the SysOps view.
func (m *Model) handleKeyPress(msg tea.KeyPressMsg) tea.Cmd {
	switch msg.String() {
	case "tab", "l", "right":
		m.tabIndex = (m.tabIndex + 1) % 3
	case "shift+tab", "h", "left":
		m.tabIndex = (m.tabIndex - 1 + 3) % 3
	case "1":
		m.tabIndex = 0
	case "2":
		m.tabIndex = 1
	case "3":
		m.tabIndex = 2
	case "r":
		m.ready = false
		return m.Init()
	case "R":
		if m.showReport {
			m.showReport = false
			return nil
		}
		m.ready = false
		return func() tea.Msg {
			report, err := RunHealthCheck()
			if err != nil {
				return WorkflowResultMsg{Err: err}
			}
			return WorkflowResultMsg{Report: report.String()}
		}
	case "esc":
		// Return to main menu is handled by the root model
		return nil
	default:
		return nil
	}
	return nil
}

// String representation of the model (for debugging).
func (m *Model) String() string {
	if m.stats == nil {
		return "SysOps: no data"
	}
	return fmt.Sprintf("SysOps: CPU=%.1f%% MEM=%.1f%% DISK=%.1f%%",
		m.stats.CPUPercent, m.stats.MemoryUsed, m.stats.DiskUsed)
}
