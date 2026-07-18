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
	HandshakeID string   `json:"handshake_id"`
	Action      string   `json:"action"`
	Description string   `json:"description"`
	Risks       []string `json:"risks"`
	Rollback    string   `json:"rollback"`
}

// SystemKnowledge represents the unified "Current Truth" of the system.
type SystemKnowledge struct {
	CPUUsage      float64 `json:"cpu_usage"`
	MemoryUsage   float64 `json:"memory_usage"`
	DiskUsage     float64 `json:"disk_usage"`
	ActiveConns   int     `json:"active_conns"`
	Anomalies     int     `json:"anomalies"`
	Uptime        string  `json:"uptime"`
	SecurityGrade string  `json:"security_grade"`
}
// These prevent excessive output in terminal views where space is constrained.
const (
	MaxFirewallRules        = 100 // cap for firewall rule lists in Markdown reports
	MaxConnections          = 20  // cap for connection table in Markdown reports
	MaxScheduledTasks       = 20  // cap for scheduled task table in Markdown reports
	MaxFirewallRulesDisplay = 30  // cap for firewall rule rows in Markdown table
)
