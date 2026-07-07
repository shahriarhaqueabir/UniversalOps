package common

import (
	"time"

	tea "charm.land/bubbletea/v2"
)

// Screen represents the active screen/layer in the TUI.
type Screen int

const (
	ScreenDashboard Screen = iota
	ScreenMainMenu
	ScreenSysOps
	ScreenNetOps
	ScreenSecOps
	ScreenDevOps
	ScreenAIOps
	ScreenOnboarding
	ScreenHelp
)

// Screen names for display.
var ScreenNames = map[Screen]string{
	ScreenDashboard:  "Dashboard",
	ScreenMainMenu:   "Main Menu",
	ScreenSysOps:     "System Operations",
	ScreenNetOps:     "Network Operations",
	ScreenSecOps:     "Security Operations",
	ScreenDevOps:     "Development Operations",
	ScreenAIOps:      "AI Operations",
	ScreenOnboarding: "Welcome",
	ScreenHelp:       "Help",
}

// MenuItem represents a selectable menu item.
type MenuItem struct {
	Title       string
	Description string
	Screen      Screen
	Key         string
}

// SystemStats holds a snapshot of system metrics.
type SystemStats struct {
	CPUPercent    float64
	MemoryUsed    float64 // percentage
	MemoryTotal   uint64  // bytes
	MemoryUsedGB  float64
	MemoryTotalGB float64
	DiskUsed      float64 // percentage
	DiskFree      uint64  // bytes
	Uptime        string
	ProcessCount  int

	// History for sparklines
	CPUHistory []float64
	MemHistory []float64

	// Anomaly count for status bar
	AnomalyCount int
}

// Display limits for report output.
// These prevent excessive output in terminal views where space is constrained.
const (
	MaxFirewallRules        = 100 // cap for firewall rule lists in Markdown reports
	MaxConnections          = 20  // cap for connection table in Markdown reports
	MaxScheduledTasks       = 20  // cap for scheduled task table in Markdown reports
	MaxFirewallRulesDisplay = 30  // cap for firewall rule rows in Markdown table
)

// TickMsg is sent periodically to refresh dashboards.
type TickMsg time.Time

// StartTickCmd creates a ticking command for periodic refresh.
func StartTickCmd(interval time.Duration) tea.Cmd {
	return tea.Tick(interval, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}
