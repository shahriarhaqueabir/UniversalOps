package app

// ── AppInfo ──────────────────────────────────────────────────────────────────

// AppInfo holds metadata about the application.
type AppInfo struct {
	Name      string `json:"name"`
	Version   string `json:"version"`
	GoVersion string `json:"go_version"`
	Uptime    string `json:"uptime"`
}

// ── Dashboard Types ──────────────────────────────────────────────────────────

// DashboardData is the top-level dashboard snapshot.
type DashboardData struct {
	CPU         GaugeMetric   `json:"cpu"`
	Memory      GaugeMetric   `json:"memory"`
	Disk        GaugeMetric   `json:"disk"`
	Network     NetworkMetric `json:"network"`
	Processes   int           `json:"processes"`
	Connections int           `json:"connections"`
	Alerts      int           `json:"alerts"`
	Uptime      string        `json:"uptime"`
}

// GaugeMetric holds a single gauge value with optional history and forecast.
type GaugeMetric struct {
	Value    float64   `json:"value"`
	Unit     string    `json:"unit"`
	History  []float64 `json:"history"`
	Forecast []float64 `json:"forecast"`
	Trend    string    `json:"trend"` // "rising", "falling", "stable"
}

// NetworkMetric holds receive and transmit rate info.
type NetworkMetric struct {
	RXRate float64 `json:"rx_rate"`
	TXRate float64 `json:"tx_rate"`
	Unit   string  `json:"unit"`
}

// ── SysOps Types ─────────────────────────────────────────────────────────────

// CPUInfo holds CPU details and usage.
type CPUInfo struct {
	Percent       float64   `json:"percent"`
	PerCPU        []float64 `json:"per_cpu"`
	ModelName     string    `json:"model_name"`
	LogicalCores  int       `json:"logical_cores"`
	PhysicalCores int       `json:"physical_cores"`
	CoreCount     int       `json:"core_count"`
	LoadAvg1      float64   `json:"load_avg_1"`
	LoadAvg5      float64   `json:"load_avg_5"`
	LoadAvg15     float64   `json:"load_avg_15"`
}

// MemoryInfo holds RAM and swap details.
type MemoryInfo struct {
	TotalBytes     uint64  `json:"total_bytes"`
	AvailableBytes uint64  `json:"available_bytes"`
	UsedBytes      uint64  `json:"used_bytes"`
	UsedPercent    float64 `json:"used_percent"`
	TotalGB        float64 `json:"total_gb"`
	UsedGB         float64 `json:"used_gb"`
	SwapTotal      uint64  `json:"swap_total"`
	SwapUsed       uint64  `json:"swap_used"`
	SwapPercent    float64 `json:"swap_percent"`
}

// DiskInfo holds partition and usage details.
type DiskInfo struct {
	Partitions []DiskPartition `json:"partitions"`
}

// DiskPartition holds usage for a single mount.
type DiskPartition struct {
	Mountpoint  string  `json:"mountpoint"`
	TotalBytes  uint64  `json:"total_bytes"`
	FreeBytes   uint64  `json:"free_bytes"`
	UsedBytes   uint64  `json:"used_bytes"`
	UsedPercent float64 `json:"used_percent"`
	FSType      string  `json:"fs_type"`
	Device      string  `json:"device"`
}

// ProcessInfo holds a single process snapshot.
type ProcessInfo struct {
	PID    int32   `json:"pid"`
	Name   string  `json:"name"`
	CPU    float64 `json:"cpu"`
	Memory float32 `json:"memory"`
	MemPct float32 `json:"mem_pct"`
	Status string  `json:"status"`
	NumFDs int32   `json:"num_fds"`
}

// SystemInfo holds general system info.
type SystemInfo struct {
	Hostname        string `json:"hostname"`
	OS              string `json:"os"`
	Platform        string `json:"platform"`
	PlatformVersion string `json:"platform_version"`
	KernelVersion   string `json:"kernel_version"`
	KernelArch      string `json:"kernel_arch"`
	Uptime          string `json:"uptime"`
	ProcessCount    int    `json:"process_count"`
	Virtualization  string `json:"virtualization"`
}

// ── NetOps Types ─────────────────────────────────────────────────────────────

// PingResult holds the results of a ping operation.
type PingResult struct {
	Target   string `json:"target"`
	IP       string `json:"ip"`
	Sent     int    `json:"sent"`
	Received int    `json:"received"`
	Lost     int    `json:"lost"`
	MinMs    int64  `json:"min_ms"`
	MaxMs    int64  `json:"max_ms"`
	AvgMs    int64  `json:"avg_ms"`
	JitterMs float64 `json:"jitter_ms"`
	TTL      int    `json:"ttl"`
	Error    string `json:"error,omitempty"`
}

// DNSResult holds DNS lookup results.
type DNSResult struct {
	Hostname string   `json:"hostname"`
	A        []string `json:"a"`
	AAAA     []string `json:"aaaa"`
	MX       []string `json:"mx"`
	NS       []string `json:"ns"`
	CNAME    string   `json:"cname"`
	TXT      []string `json:"txt"`
	Error    string   `json:"error,omitempty"`
}

// PortResult holds a single port scan result.
type PortResult struct {
	Port    int    `json:"port"`
	Open    bool   `json:"open"`
	Service string `json:"service"`
}

// TraceHop holds a single traceroute hop.
type TraceHop struct {
	Number int     `json:"number"`
	Host   string  `json:"host"`
	IP     string  `json:"ip"`
	RTTsMs []int64 `json:"rtts_ms"`
	Timed  bool    `json:"timed"`
}

// TraceResult holds a complete traceroute result.
type TraceResult struct {
	Target string     `json:"target"`
	Hops   []TraceHop `json:"hops"`
	Error  string     `json:"error,omitempty"`
}

// ConnectionInfo holds a network connection entry.
type ConnectionInfo struct {
	LocalAddr   string `json:"local_addr"`
	RemoteAddr  string `json:"remote_addr"`
	LocalPort   int    `json:"local_port"`
	RemotePort  int    `json:"remote_port"`
	State       string `json:"state"`
	ProcessName string `json:"process_name"`
	PID         int    `json:"pid"`
}

// InterfaceInfo holds a network interface entry.
type InterfaceInfo struct {
	Name      string    `json:"name"`
	MAC       string    `json:"mac"`
	IPs       []string  `json:"ips"`
	IsUp      bool      `json:"is_up"`
	Speed     string    `json:"speed"`
	MTU       int       `json:"mtu"`
	Flags     string    `json:"flags"`
	RXBytes   uint64    `json:"rx_bytes"`
	TXBytes   uint64    `json:"tx_bytes"`
	RXRateBps float64   `json:"rx_rate_bps"`
	TXRateBps float64   `json:"tx_rate_bps"`
	RXHistory []float64 `json:"rx_history"`
	TXHistory []float64 `json:"tx_history"`
}

// ── SecOps Types ─────────────────────────────────────────────────────────────

// FirewallRule holds a firewall rule entry.
type FirewallRule struct {
	Name       string `json:"name"`
	Direction  string `json:"direction"`
	Action     string `json:"action"`
	Protocol   string `json:"protocol"`
	LocalPort  string `json:"local_port"`
	RemotePort string `json:"remote_port"`
	RemoteIP   string `json:"remote_ip"`
	Profile    string `json:"profile"`
	Enabled    bool   `json:"enabled"`
	IsHighRisk bool   `json:"is_high_risk"`
}

// UserInfo holds a local user account entry.
type UserInfo struct {
	Username  string `json:"username"`
	FullName  string `json:"full_name"`
	SID       string `json:"sid"`
	Group     string `json:"group"`
	IsAdmin   bool   `json:"is_admin"`
	IsEnabled bool   `json:"is_enabled"`
}

// ListeningPort holds a listening port entry.
type ListeningPort struct {
	Port        int    `json:"port"`
	Protocol    string `json:"protocol"`
	ProcessName string `json:"process_name"`
	PID         int    `json:"pid"`
	State       string `json:"state"`
	IsExternal  bool   `json:"is_external"`
}

// DefenderStatus holds Windows Defender status.
type DefenderStatus struct {
	Enabled            bool   `json:"enabled"`
	UpToDate           bool   `json:"up_to_date"`
	SignatureAge       string `json:"signature_age"`
	LastScan           string `json:"last_scan"`
	RealTimeProtection bool   `json:"real_time_protection"`
	CloudProtection    bool   `json:"cloud_protection"`
	AMServiceEnabled   bool   `json:"am_service_enabled"`
	AntispywareEnabled bool   `json:"antispyware_enabled"`
	NISEnabled         bool   `json:"nis_enabled"`
	QuickScanAge       int    `json:"quick_scan_age"`
	FullScanAge        int    `json:"full_scan_age"`
}

// ScheduledTask holds a scheduled task entry.
type ScheduledTask struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	NextRun string `json:"next_run"`
	LastRun string `json:"last_run"`
	Author  string `json:"author"`
	Trigger string `json:"trigger"`
}

// SecurityEvent holds a security event log entry.
type SecurityEvent struct {
	ID        int    `json:"id"`
	Level     string `json:"level"`
	Provider  string `json:"provider"`
	Time      string `json:"time"`
	Message   string `json:"message"`
	Important bool   `json:"important"`
}

// ── DevOps Types ─────────────────────────────────────────────────────────────

// CommandResult holds the result of a shell command.
type CommandResult struct {
	Command  string `json:"command"`
	Output   string `json:"output"`
	ExitCode int    `json:"exit_code"`
	Duration int64  `json:"duration_ms"`
	Error    string `json:"error,omitempty"`
}

// LogEntry holds a log line with metadata.
type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Module    string `json:"module"`
	Message   string `json:"message"`
	Line      string `json:"line"`
	Source    string `json:"source"`
}

// FileEntry holds a file or directory entry.
type FileEntry struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	Size     string `json:"size"`
	RawSize  int64  `json:"raw_size"`
	IsDir    bool   `json:"is_dir"`
	IsBinary bool   `json:"is_binary"`
	Mode     string `json:"mode"`
	ModTime  string `json:"mod_time"`
}

// ServiceEntry holds a system service entry.
type ServiceEntry struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Status      string `json:"status"`
	StartType   string `json:"start_type"`
}

// ── AIOps Types ──────────────────────────────────────────────────────────────

// ChatMessage holds a chat message.
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OllamaStatus holds Ollama service status.
type OllamaStatus struct {
	Available       bool     `json:"available"`
	Model           string   `json:"model"`
	Version         string   `json:"version"`
	AvailableModels []string `json:"available_models"`
	Error           string   `json:"error,omitempty"`
}

// AnomalyInfo holds a detected anomaly.
type AnomalyInfo struct {
	Metric    string  `json:"metric"`
	Value     float64 `json:"value"`
	Expected  float64 `json:"expected"`
	Deviation float64 `json:"deviation"`
	Severity  string  `json:"severity"`
	Timestamp string  `json:"timestamp"`
}

// ── Pipeline Types ───────────────────────────────────────────────────────────

// MetricHistory holds metric time-series data.
type MetricHistory struct {
	Name      string    `json:"name"`
	Unit      string    `json:"unit"`
	Values    []float64 `json:"values"`
	Forecast  []float64 `json:"forecast"`
	Trend     TrendInfo `json:"trend"`
	Stats     StatsInfo `json:"stats"`
	LastValue float64   `json:"last_value"`
}

// TrendInfo describes the direction and magnitude of a trend.
type TrendInfo struct {
	Direction   string  `json:"direction"` // "rising", "falling", "stable"
	ChangePct   float64 `json:"change_pct"`
	Slope       float64 `json:"slope"`
	Correlation float64 `json:"correlation"`
}

// StatsInfo holds rolling window statistics.
type StatsInfo struct {
	Min   float64 `json:"min"`
	Max   float64 `json:"max"`
	Avg   float64 `json:"avg"`
	P50   float64 `json:"p50"`
	P95   float64 `json:"p95"`
	P99   float64 `json:"p99"`
	Count int     `json:"count"`
}

// ── Alert Types ──────────────────────────────────────────────────────────────

// AlertInfo holds a firing alert.
type AlertInfo struct {
	ID        string  `json:"id"`
	Level     string  `json:"level"`
	Metric    string  `json:"metric"`
	Message   string  `json:"message"`
	Value     float64 `json:"value"`
	Threshold float64 `json:"threshold"`
	Timestamp string  `json:"timestamp"`
	Resolved  bool    `json:"resolved"`
}

// AlertRuleInfo holds a user-defined alert rule.
type AlertRuleInfo struct {
	Metric    string  `json:"metric"`
	Condition string  `json:"condition"`
	Threshold float64 `json:"threshold"`
	FlapCount int     `json:"flap_count"`
	Severity  string  `json:"severity"`
	Message   string  `json:"message"`
}

// ── Log Types ────────────────────────────────────────────────────────────────

// LogFilter holds filter parameters for log queries.
type LogFilter struct {
	Level string `json:"level"` // "info", "warn", "error"
	Since string `json:"since"` // ISO 8601 timestamp
	N     int    `json:"n"`     // max entries
}
