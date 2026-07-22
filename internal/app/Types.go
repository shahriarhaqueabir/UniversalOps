package app

import (
	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
)

// ── AppInfo ──────────────────────────────────────────────────────────────────

// AppInfo holds metadata about the application.
type AppInfo struct {
	Name      string `json:"name"`
	Version   string `json:"version"`
	GoVersion string `json:"go_version"`
	Uptime    string `json:"uptime"`
}

// EnvReport holds information about the host environment.
type EnvReport struct {
	Hostname    string   `json:"hostname"`
	CPU         string   `json:"cpu"`
	Cores       int      `json:"cores"`
	Memory      string   `json:"memory"`
	OS          string   `json:"os"`
	Arch        string   `json:"arch"`
	Interfaces  []string `json:"interfaces"`
	PackageMgrs []string `json:"package_mgrs"`
	Shells      []string `json:"shells"`
}

// ── Dashboard Types ──────────────────────────────────────────────────────────

// DiagnosticResult holds a quick diagnostic check result.
type DiagnosticResult struct {
	Category string  `json:"category"`
	Status   string  `json:"status"` // "pass", "warn", "fail"
	Message  string  `json:"message"`
	Value    float64 `json:"value"`
	Unit     string  `json:"unit"`
}

// BriefingSection holds a single section of a generated briefing.
type BriefingSection struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Level   string `json:"level"` // "info", "warning", "critical"
}

// GPUData is a compatibility alias for GPUInfo.
type GPUData = GPUInfo

// BatteryData is a compatibility alias for BatteryInfo.
type BatteryData = BatteryInfo

// DashboardData is the top-level dashboard snapshot.
type DashboardData struct {
	CPU         GaugeMetric   `json:"cpu"`
	Memory      GaugeMetric   `json:"memory"`
	Disk        GaugeMetric   `json:"disk"`
	GPU         GPUInfo       `json:"gpu"`
	Battery     BatteryInfo   `json:"battery"`
	Network     NetworkMetric `json:"network"`
	Processes   int           `json:"processes"`
	Connections int           `json:"connections"`
	Alerts      int           `json:"alerts"`
	Uptime      string        `json:"uptime"`
}

// SystemSnapshot provides a single-call state for Batch IPC.
type SystemSnapshot struct {
	Metrics   DashboardData   `json:"metrics"`
	Timeline  []TimelineEvent `json:"timeline"`
	Alerts    []AlertInfo     `json:"alerts"`
	Timestamp string          `json:"timestamp"`
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

// DataPoint represents a timestamped value for the frontend.
type DataPoint struct {
	Time  string  `json:"time"`
	Value float64 `json:"value"`
}

// MetricDef describes a tracked metric.
type MetricDef struct {
	Name  string `json:"name"`
	Unit  string `json:"unit"`
	Label string `json:"label"`
}

// ── SysOps Types ─────────────────────────────────────────────────────────────

// HardwareInfo holds full workstation telemetry.
type HardwareInfo struct {
	CPU       CPUExtendedInfo `json:"cpu"`
	GPU       GPUInfo         `json:"gpu"`
	Battery   BatteryInfo     `json:"battery"`
	Sensors   []SensorData    `json:"sensors"`
	Baseboard BaseboardInfo   `json:"baseboard"`
}

// SensorData holds a single hardware sensor reading.
type SensorData struct {
	Name     string  `json:"name"`
	Type     string  `json:"type"` // Temperature, Fan, Voltage
	Value    float64 `json:"value"`
	Unit     string  `json:"unit"`
	Category string  `json:"category"` // CPU, GPU, Board
}

// BaseboardInfo holds motherboard information.
type BaseboardInfo struct {
	Manufacturer string `json:"manufacturer"`
	Product      string `json:"product"`
	Version      string `json:"version"`
	SerialNumber string `json:"serial_number"`
}

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

// CPUExtendedInfo holds extended CPU details for SysOps.
type CPUExtendedInfo struct {
	ModelName    string           `json:"model_name"`
	FrequencyMHz float64          `json:"frequency_mhz"`
	CacheSizeKB  int32            `json:"cache_size_kb"`
	Temperature  float64          `json:"temperature"`
	PerCPUInfo   []PerCPUInfoData `json:"per_cpu_info"`
}

// PerCPUInfoData holds per-core info for frontend.
type PerCPUInfoData struct {
	Core      int     `json:"core"`
	Frequency float64 `json:"frequency_mhz"`
	Usage     float64 `json:"usage_percent"`
}

// MemoryInfo holds RAM and swap details.
type MemoryInfo struct {
	TotalBytes     uint64  `json:"total_bytes"`
	AvailableBytes uint64  `json:"available_bytes"`
	UsedBytes      uint64  `json:"used_bytes"`
	UsedPercent    float64 `json:"used_percent"`
	CachedBytes    uint64  `json:"cached_bytes"`
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

// DiskIOEntry holds I/O stats for a single disk.
type DiskIOEntry struct {
	Name       string `json:"name"`
	ReadBytes  uint64 `json:"read_bytes"`
	WriteBytes uint64 `json:"write_bytes"`
	ReadCount  uint64 `json:"read_count"`
	WriteCount uint64 `json:"write_count"`
}

// DiskIOData holds aggregate disk I/O information.
type DiskIOData struct {
	Disks      []DiskIOEntry `json:"disks"`
	TotalRead  uint64        `json:"total_read_bytes"`
	TotalWrite uint64        `json:"total_write_bytes"`
}

// ProcessInfo holds a single process snapshot.
type ProcessInfo struct {
	PID       int32   `json:"pid"`
	PPID      int32   `json:"ppid"`
	Name      string  `json:"name"`
	CPU       float64 `json:"cpu"`
	Memory    float32 `json:"memory"`
	MemPct    float32 `json:"mem_pct"`
	Status    string  `json:"status"`
	NumFDs    int32   `json:"num_fds"`
	IsSigned  bool    `json:"is_signed"`
	Publisher string  `json:"publisher"`
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

// GPUInfo holds detailed GPU information.
type GPUInfo struct {
	Name        string  `json:"name"`
	Vendor      string  `json:"vendor"`
	MemoryGB    float64 `json:"memory_gb"`
	Driver      string  `json:"driver"`
	Detected    bool    `json:"detected"`
	Temperature float64 `json:"temperature"`
	Utilization float64 `json:"utilization"`
	FanSpeed    float64 `json:"fan_speed"`
}

// BatteryInfo holds detailed battery information.
type BatteryInfo struct {
	Percent     float64 `json:"percent"`
	Charging    bool    `json:"charging"`
	TimeLeftSec int64   `json:"time_left_sec"`
	Status      string  `json:"status"`
	Detected    bool    `json:"detected"`
}

// LoggedInUserData holds logged-in user info for frontend.
type LoggedInUserData struct {
	User     string `json:"user"`
	Terminal string `json:"terminal"`
	Host     string `json:"host"`
	Started  string `json:"started"`
}

// CPUTimesData holds CPU time breakdown for frontend.
type CPUTimesData struct {
	User   float64 `json:"user"`
	System float64 `json:"system"`
	Idle   float64 `json:"idle"`
	IOWait float64 `json:"iowait"`
	Steal  float64 `json:"steal"`
	Total  float64 `json:"total"`
}

// LoadAverageData holds load average for frontend.
type LoadAverageData struct {
	Load1  float64 `json:"load_1"`
	Load5  float64 `json:"load_5"`
	Load15 float64 `json:"load_15"`
}

// PerformanceData holds performance stats for frontend.
type PerformanceData struct {
	CPUTimes    CPUTimesData    `json:"cpu_times"`
	LoadAverage LoadAverageData `json:"load_average"`
	IOWait      float64         `json:"io_wait"`
}

// ActionResult holds the result of a system action for frontend.
type ActionResult struct {
	Action  string `json:"action"`
	Success bool   `json:"success"`
	Message string `json:"message"`
	Output  string `json:"output"`
}

// SystemLogEntry holds a single log entry for frontend.
type SystemLogEntry struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Source    string `json:"source"`
	Message   string `json:"message"`
}

// SystemLogsResultData holds log results for frontend.
type SystemLogsResultData struct {
	Entries []SystemLogEntry `json:"entries"`
	Source  string           `json:"source"`
	Total   int              `json:"total"`
}

// ScheduledTaskData holds a scheduled task for frontend (SysOps).
type ScheduledTaskData struct {
	Name     string `json:"name"`
	Schedule string `json:"schedule"`
	Command  string `json:"command"`
	Enabled  bool   `json:"enabled"`
	NextRun  string `json:"next_run"`
}

// DiagnosticCheckData holds a single diagnostic check for frontend.
type DiagnosticCheckData struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Message string `json:"message"`
	Value   string `json:"value"`
}

// ExtendedDiagnosticResult holds the full diagnostic result for frontend.
type ExtendedDiagnosticResult struct {
	Checks    []DiagnosticCheckData `json:"checks"`
	Score     int                   `json:"score"`
	Timestamp string                `json:"timestamp"`
}

// PackageManagerData holds info about a package manager and its installed packages.
type PackageManagerData struct {
	Name     string        `json:"name"`
	Found    bool          `json:"found"`
	Packages []PackageInfo `json:"packages"`
}

// PackageInfo holds info about a single installed package.
type PackageInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// ── NetOps Types ─────────────────────────────────────────────────────────────

// NetworkHealthCheck holds a single network health check result.
type NetworkHealthCheck struct {
	Name   string `json:"name"`
	Status string `json:"status"` // "pass", "warn", "fail"
	Detail string `json:"detail"`
	Score  int    `json:"score"`
}

// NetworkHealthReport holds the full network health check report for the frontend.
type NetworkHealthReport struct {
	Score    int                  `json:"score"`
	Checks   []NetworkHealthCheck `json:"checks"`
	Summary  string               `json:"summary"`
	Duration string               `json:"duration"`
}

// PingResult holds the results of a ping operation.
type PingResult struct {
	Target   string  `json:"target"`
	IP       string  `json:"ip"`
	Sent     int     `json:"sent"`
	Received int     `json:"received"`
	Lost     int     `json:"lost"`
	MinMs    int64   `json:"min_ms"`
	MaxMs    int64   `json:"max_ms"`
	AvgMs    int64   `json:"avg_ms"`
	JitterMs float64 `json:"jitter_ms"`
	TTL      int     `json:"ttl"`
	Error    string  `json:"error,omitempty"`
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
	Protocol    string `json:"protocol"`
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

// NetworkChangeType enumerates the kinds of interface state changes.
type NetworkChangeType string

const (
	ChangeUp          NetworkChangeType = "up"
	ChangeDown        NetworkChangeType = "down"
	ChangeIPAdded     NetworkChangeType = "ip_added"
	ChangeIPRemoved   NetworkChangeType = "ip_removed"
	ChangeAppeared    NetworkChangeType = "appeared"
	ChangeDisappeared NetworkChangeType = "disappeared"
)

// NetworkChange records a single detected interface state change.
type NetworkChange struct {
	Type      NetworkChangeType `json:"type"` // up, down, ip_added, ip_removed, appeared, disappeared
	Interface string            `json:"interface"`
	Detail    string            `json:"detail"`    // e.g. new IP address
	Timestamp string            `json:"timestamp"` // RFC3339
}

// ── SecOps Types ─────────────────────────────────────────────────────────────

// SecurityScore holds a computed security posture score.
type SecurityScore struct {
	Score           int            `json:"score"`
	Grade           string         `json:"grade"`
	Breakdown       map[string]int `json:"breakdown"`
	Recommendations []string       `json:"recommendations"`
}

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
	Username             string `json:"username"`
	FullName             string `json:"full_name"`
	SID                  string `json:"sid"`
	Group                string `json:"group"`
	IsAdmin              bool   `json:"is_admin"`
	IsEnabled            bool   `json:"is_enabled"`
	PasswordNeverExpires bool   `json:"password_never_expires"`
	LastLogon            string `json:"last_logon"`
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

// ScheduledTask holds a scheduled task entry for SecOps.
type ScheduledTask struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	NextRun string `json:"next_run"`
	LastRun string `json:"last_run"`
	Author  string `json:"author"`
	Trigger string `json:"trigger"`
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
	ThreatsDetected    int    `json:"threats_detected"`
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

// FirewallProfile holds the status of a single firewall profile.
type FirewallProfile struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}

// FirewallStatus holds the global firewall status.
type FirewallStatus struct {
	Enabled  bool              `json:"enabled"`
	Profiles []FirewallProfile `json:"profiles"`
}

// RiskInfo holds a detected security risk.
type RiskInfo struct {
	Category       string `json:"category"`
	Severity       string `json:"severity"`
	Title          string `json:"title"`
	Description    string `json:"description"`
	Recommendation string `json:"recommendation"`
}

// SecuritySummary holds a unified security posture overview for the AI summary panel.
type SecuritySummary struct {
	Score           int      `json:"score"`
	Summary         string   `json:"summary"`
	Risks           []string `json:"risks"`
	Recommendations []string `json:"recommendations"`
	AnalyzedAt      string   `json:"analyzedAt"`
}

// PasswordPolicy holds password policy configuration.
type PasswordPolicy struct {
	MaxAge           int  `json:"max_age"`
	MinLength        int  `json:"min_length"`
	Complexity       bool `json:"complexity"`
	LockoutThreshold int  `json:"lockout_threshold"`
	LockoutDuration  int  `json:"lockout_duration"`
}

// FailedLogin holds a failed login attempt record.
type FailedLogin struct {
	Time     string `json:"time"`
	Username string `json:"username"`
	SourceIP string `json:"source_ip"`
	Count    int    `json:"count"`
}

// LockedAccount holds a locked account record.
type LockedAccount struct {
	Username    string `json:"username"`
	LockedSince string `json:"locked_since"`
}

// DiskEncryption holds disk encryption status.
type DiskEncryption struct {
	Volume    string `json:"volume"`
	Encrypted bool   `json:"encrypted"`
	Method    string `json:"method"`
	Status    string `json:"status"`
}

// SecureBoot holds secure boot status.
type SecureBoot struct {
	Enabled bool   `json:"enabled"`
	State   string `json:"state"`
}

// SystemService holds a system service entry.
type SystemService struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Status      string `json:"status"`
	StartupType string `json:"startup_type"`
}

// TLSCertificate holds TLS certificate info.
type TLSCertificate struct {
	Subject    string `json:"subject"`
	Issuer     string `json:"issuer"`
	NotAfter   string `json:"not_after"`
	KeySize    int    `json:"key_size"`
	IsExpiring bool   `json:"is_expiring"`
	DaysLeft   int    `json:"days_left"`
}

// SSHConfig holds SSH configuration.
type SSHConfig struct {
	PermitRootLogin        string `json:"permit_root_login"`
	PasswordAuthentication string `json:"password_authentication"`
	PubkeyAuthentication   string `json:"pubkey_authentication"`
	X11Forwarding          string `json:"x11_forwarding"`
	MaxAuthTries           string `json:"max_auth_tries"`
}

// HardeningCheck holds a single hardening check result.
type HardeningCheck struct {
	Category    string `json:"category"`
	Check       string `json:"check"`
	Passed      bool   `json:"passed"`
	Severity    string `json:"severity"`
	Remediation string `json:"remediation"`
}

// AuditCheckItem holds a single audit check result.
type AuditCheckItem struct {
	Category    string `json:"category"`
	Check       string `json:"check"`
	Passed      bool   `json:"passed"`
	Description string `json:"description"`
	Remediation string `json:"remediation"`
}

// SecurityAuditResult holds the full audit result.
type SecurityAuditResult struct {
	Score     int              `json:"score"`
	Total     int              `json:"total"`
	Passed    int              `json:"passed"`
	Failed    int              `json:"failed"`
	Items     []AuditCheckItem `json:"items"`
	Timestamp string           `json:"timestamp"`
}

// PrivilegeEvent holds a privilege escalation event.
type PrivilegeEvent struct {
	Time      string `json:"time"`
	Username  string `json:"username"`
	Privilege string `json:"privilege"`
	Process   string `json:"process"`
}

// SecTimelineEvent holds a chronological security event.
type SecTimelineEvent struct {
	Time     string `json:"time"`
	Type     string `json:"type"`
	Detail   string `json:"detail"`
	Severity string `json:"severity"`
}

// PublicExposure holds public-facing port info.
type PublicExposure struct {
	Port        int    `json:"port"`
	Protocol    string `json:"protocol"`
	ProcessName string `json:"process_name"`
	Severity    string `json:"severity"`
}

// SecActionResult aliases common.SecActionResult for Wails bindings.
type SecActionResult = common.SecActionResult

// ── DevOps Types ─────────────────────────────────────────────────────────────

// DevOpsDiagCheck holds a single DevOps diagnostic check result for the frontend.
type DevOpsDiagCheck struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Message string `json:"message"`
	Value   string `json:"value"`
}

// DevOpsDiagResult holds the complete DevOps diagnostic result for the frontend.
type DevOpsDiagResult struct {
	Checks    []DevOpsDiagCheck `json:"checks"`
	Score     int               `json:"score"`
	Timestamp string            `json:"timestamp"`
}

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

// ToolInfo holds information about an installed development tool.
type ToolInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Path    string `json:"path"`
	Status  string `json:"status"` // "installed", "not-found", "error"
}

// ContainerInfo holds information about a single container.
type ContainerInfo struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Image  string `json:"image"`
	State  string `json:"state"`  // running, exited, created, paused
	Status string `json:"status"` // "Up 2 hours", "Exited (0) 3 days ago"
	Ports  string `json:"ports"`
}

// ContainerSummary holds Docker container status overview.
type ContainerSummary struct {
	Running    int             `json:"running"`
	Stopped    int             `json:"stopped"`
	Failed     int             `json:"failed"`
	Total      int             `json:"total"`
	Containers []ContainerInfo `json:"containers"`
}

// GitRepoInfo holds status for a single git repository.
type GitRepoInfo struct {
	Path           string `json:"path"`
	Branch         string `json:"branch"`
	ModifiedFiles  int    `json:"modified_files"`
	UntrackedFiles int    `json:"untracked_files"`
	Ahead          int    `json:"ahead"`
	Behind         int    `json:"behind"`
	Clean          bool   `json:"clean"`
}

// GitSummary holds aggregated git repository status.
type GitSummary struct {
	Repositories []GitRepoInfo `json:"repositories"`
	TotalRepos   int           `json:"total_repos"`
}

// LocalServer holds information about a locally listening server.
type LocalServer struct {
	Port      int    `json:"port"`
	Protocol  string `json:"protocol"`
	Process   string `json:"process"`
	PID       int    `json:"pid"`
	Framework string `json:"framework"`
	Health    string `json:"health"` // "healthy", "unknown", "error"
}

// EnvVarInfo holds a single environment variable.
type EnvVarInfo struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// ToolVersion holds a tool name and its version string.
type ToolVersion struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// EnvironmentInfo holds environment and SDK details.
type EnvironmentInfo struct {
	PathDirs        []string      `json:"path_dirs"`
	KeyVars         []EnvVarInfo  `json:"key_vars"`
	SDKs            []ToolVersion `json:"sdks"`
	PackageManagers []ToolVersion `json:"package_managers"`
}

// DockerStatus holds Docker daemon and container overview.
type DockerStatus struct {
	Installed  bool             `json:"installed"`
	Running    bool             `json:"running"` // daemon running
	Version    string           `json:"version"`
	Containers ContainerSummary `json:"containers"`
}

// KubernetesStatus holds kubectl and cluster connectivity info.
type KubernetesStatus struct {
	Installed bool   `json:"installed"`
	Connected bool   `json:"connected"`
	Cluster   string `json:"cluster"`
	Nodes     int    `json:"nodes"`
	Pods      int    `json:"pods"`
}

// ServiceCategory groups services by their function type.
type ServiceCategory struct {
	Category string        `json:"category"` // databases, message-queues, web-servers, containers, other
	Services []ServiceInfo `json:"services"`
}

// ServiceInfo is a simplified service entry for categorization.
type ServiceInfo struct {
	Name   string `json:"name"`
	Status string `json:"status"` // running, stopped, unknown
}

// ServiceGroupSummary aggregates service counts by category.
type ServiceGroupSummary struct {
	Databases     int `json:"databases"`
	MessageQueues int `json:"messageQueues"`
	WebServers    int `json:"webServers"`
	Containers    int `json:"containers"`
	Other         int `json:"other"`
	Running       int `json:"running"`
	Stopped       int `json:"stopped"`
}

// ── Pipeline Types ───────────────────────────────────────────────────────────

// ── AIOps Types ──────────────────────────────────────────────────────────────

// ChatMessage holds a chat message.
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OllamaStatus holds Ollama service status.
type OllamaStatus struct {
	Available       bool     `json:"available"`
	BinaryExists    bool     `json:"binary_exists"`
	Model           string   `json:"model"`
	Version         string   `json:"version"`
	AvailableModels []string `json:"available_models"`
	Error           string   `json:"error,omitempty"`
}

// OllamaProgress holds progress information for Ollama operations.
type OllamaProgress struct {
	Status    string  `json:"status"`
	Percent   float64 `json:"percent"`
	Total     int64   `json:"total"`
	Completed int64   `json:"completed"`
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

// LogStats holds aggregated log statistics for the Overview tab.
type LogStats struct {
	TotalToday     int              `json:"totalToday"`
	TotalThisHour  int              `json:"totalThisHour"`
	TotalLastMin   int              `json:"totalLastMin"`
	ErrorCount     int              `json:"errorCount"`
	WarningCount   int              `json:"warningCount"`
	InfoCount      int              `json:"infoCount"`
	DebugCount     int              `json:"debugCount"`
	TopSources     []LogSourceCount `json:"topSources"`
	TrendingErrors []TrendingError  `json:"trendingErrors"`
}

// LogSourceCount pairs a log module with its occurrence count.
type LogSourceCount struct {
	Source string `json:"source"`
	Count  int    `json:"count"`
}

// TrendingError holds a frequently occurring error message.
type TrendingError struct {
	Message  string `json:"message"`
	Count    int    `json:"count"`
	LastSeen string `json:"lastSeen"`
}

// LogTimelinePoint represents a time-bucketed log count for charting.
type LogTimelinePoint struct {
	Timestamp string `json:"timestamp"`
	Total     int    `json:"total"`
	Errors    int    `json:"errors"`
	Warnings  int    `json:"warnings"`
	Info      int    `json:"info"`
}

// LogSummary holds a deterministic summary of recent log activity.
type LogSummary struct {
	TopSource   string `json:"topSource"`
	TopMessage  string `json:"topMessage"`
	ErrorTrend  string `json:"trend"`       // "increasing", "stable", "decreasing"
	SummaryText string `json:"summaryText"` // Human-readable one-liner
}

// SystemRecommendation is a heuristic recommendation for system health.
type SystemRecommendation struct {
	Category string `json:"category"` // cpu, memory, disk, uptime
	Severity string `json:"severity"` // info, warning, critical
	Message  string `json:"message"`
}

// DevOpsSuggestion is an actionable suggestion derived from DevOps data.
type DevOpsSuggestion struct {
	Category string `json:"category"` // "docker", "git", "node", "general"
	Severity string `json:"severity"` // "info", "warning", "critical"
	Message  string `json:"message"`
	Action   string `json:"action"` // suggested action
}

// AIInsight is a synthesized observation from the AIOps engine.
type AIInsight struct {
	Category  string `json:"category"` // "performance", "security", "network", "storage"
	Severity  string `json:"severity"` // "info", "warning", "critical"
	Title     string `json:"title"`
	Message   string `json:"message"`
	Action    string `json:"action"` // suggested action
	Timestamp string `json:"timestamp"`
}

// AIConfidence holds the overall confidence score and per-factor breakdown.
type AIConfidence struct {
	Overall   float64            `json:"overall"` // 0-100
	Factors   map[string]float64 `json:"factors"` // per-dimension scores
	UpdatedAt string             `json:"updatedAt"`
}

// ConversationMessage represents a single chat message.
type ConversationMessage struct {
	ID        int64  `json:"id"`
	SessionID string `json:"session_id"`
	Role      string `json:"role"`
	Content   string `json:"content"`
	Timestamp string `json:"timestamp"`
}

// NetworkSummary provides a deterministic overview of network state.
type NetworkSummary struct {
	SummaryText  string   `json:"summaryText"`
	TopInterface string   `json:"topInterface"`
	Issues       []string `json:"issues"`
}

// GatewayInfo holds information about the default gateway.
type GatewayInfo struct {
	IP        string `json:"ip"`
	Interface string `json:"interface"`
	Reachable bool   `json:"reachable"`
}

// ── NetOps Extended Types ─────────────────────────────────────────────────────

// DoHResultData holds the result of a DNS-over-HTTPS test.
type DoHResultData struct {
	Server     string  `json:"server"`
	LatencyMs  float64 `json:"latency_ms"`
	Success    bool    `json:"success"`
	ResolvedIP string  `json:"resolved_ip"`
}

// PingResultMultiData holds multi-target ping results for a single target.
type PingResultMultiData struct {
	Target         string    `json:"target"`
	MinMs          float64   `json:"min_ms"`
	AvgMs          float64   `json:"avg_ms"`
	MaxMs          float64   `json:"max_ms"`
	StdDevMs       float64   `json:"stddev_ms"`
	PacketLoss     float64   `json:"packet_loss"`
	JitterMs       float64   `json:"jitter_ms"`
	IndividualRTTs []float64 `json:"individual_rtts"`
	Success        bool      `json:"success"`
	Error          string    `json:"error,omitempty"`
}

// PingStatsData holds aggregate statistics across multiple ping results.
type PingStatsData struct {
	AvgLatency  float64 `json:"avg_latency"`
	MaxLatency  float64 `json:"max_latency"`
	TotalLoss   float64 `json:"total_loss"`
	WorstTarget string  `json:"worst_target"`
}

// VPNStatusData holds the current VPN connection status.
type VPNStatusData struct {
	Active    bool   `json:"active"`
	Type      string `json:"type"`
	Interface string `json:"interface"`
	RemoteIP  string `json:"remote_ip"`
	LocalIP   string `json:"local_ip"`
	Protocol  string `json:"protocol"`
}

// BandwidthSampleData holds a single bandwidth measurement.
type BandwidthSampleData struct {
	Timestamp     string  `json:"timestamp"`
	RxBytesPerSec float64 `json:"rx_bytes_per_sec"`
	TxBytesPerSec float64 `json:"tx_bytes_per_sec"`
	Interface     string  `json:"interface"`
}

// NetworkActionResult holds the result of a network action.
type NetworkActionResult struct {
	Action  string `json:"action"`
	Message string `json:"message"`
	Success bool   `json:"success"`
}
