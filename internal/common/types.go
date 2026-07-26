package common

// SystemStats holds a snapshot of system metrics following OTel semantic conventions.
type SystemStats struct {
	SystemCPUUtilization float64
	SystemMemoryUsage    float64 // percentage
	SystemMemoryTotal    uint64  // bytes
	SystemMemoryUsedGB   float64
	SystemMemoryTotalGB  float64
	SystemDiskUsage      float64 // percentage
	SystemDiskFree       uint64  // bytes
	SystemUptime         string
	ProcessCount         int

	// History for sparklines
	CPUHistory []float64
	MemHistory []float64

	// Anomaly count for status bar
	AnomalyCount int
}

// TrendInfo describes the direction and magnitude of a trend.
type TrendInfo struct {
	Direction   TrendDirection `json:"direction"`
	ChangePct   float64        `json:"change_pct"`   // percent change over the window
	Slope       float64        `json:"slope"`        // linear regression slope
	Intercept   float64        `json:"intercept"`    // linear regression intercept
	Correlation float64        `json:"correlation"`  // Pearson R (how well the line fits)
}

// TrendDirection indicates the direction of movement.
type TrendDirection int

const (
	TrendStable  TrendDirection = 0
	TrendRising  TrendDirection = 1
	TrendFalling TrendDirection = -1
)

// ActionPreview represents a proposed system change.
type ActionPreview struct {
	HandshakeID string         `json:"handshake_id"`
	Action      string         `json:"action"`
	Command     string         `json:"command,omitempty"`
	Description string         `json:"description"`
	Risks       []string       `json:"risks"`
	Rollback    string         `json:"rollback"`
	TypicalVals string         `json:"typical_values"`
	WorkflowID  string         `json:"workflow_id,omitempty"`
	Steps       []WorkflowStep `json:"steps,omitempty"`
}


// SystemKnowledge represents the unified "Current Truth" of the system.
type SystemKnowledge struct {
	SystemCPUUtilization float64 `json:"system.cpu.utilization"`
	CPUTrend             string  `json:"cpu_trend"` // "rising", "falling", "stable"
	SystemMemoryUsage    float64 `json:"system.memory.usage"`
	MemoryTrend          string  `json:"memory_trend"`
	SystemDiskUsage      float64 `json:"system.disk.usage"`
	DiskTrend            string  `json:"disk_trend"`
	ActiveConns          int     `json:"active_conns"`
	Anomalies            int     `json:"anomalies"`
	SystemUptime         string  `json:"system.uptime"`
	SecurityGrade        string  `json:"security_grade"`
}

// SecActionResult holds the result of a security action.
type SecActionResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Detail  string `json:"detail,omitempty"`
	Error   string `json:"error,omitempty"`
}

// These prevent excessive output in terminal views where space is constrained.
const (
	MaxFirewallRules        = 100 // cap for firewall rule lists in Markdown reports
	MaxConnections          = 20  // cap for connection table in Markdown reports
	MaxScheduledTasks       = 20  // cap for scheduled task table in Markdown reports
	MaxFirewallRulesDisplay = 30  // cap for firewall rule rows in Markdown table
)
