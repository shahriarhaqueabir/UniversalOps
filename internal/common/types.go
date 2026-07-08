package common

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
