package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/shahriarhaqueabir/UniversalOps/internal/common"
	"github.com/shahriarhaqueabir/UniversalOps/internal/devops"
	"github.com/shahriarhaqueabir/UniversalOps/internal/netops"
	"github.com/shahriarhaqueabir/UniversalOps/internal/secops"
	"github.com/shahriarhaqueabir/UniversalOps/internal/sysops"
)

// Validation constants for MCP tool arguments.
const (
	maxTargetLength  = 253 // max DNS hostname length
	maxCountValue    = 20  // max ping/traceroute count
	maxTTLValue      = 64  // max TTL
	maxPortsPerScan  = 100 // max ports per scan request
	maxProcessListN  = 500 // max processes to list
	maxLogEntries    = 500 // max log entries to retrieve
	maxQueryLimit    = 500 // max query result limit
	maxSearchLength  = 200 // max search string length
	maxMetricNameLen = 100 // max metric name length
)

// hostnameRe matches valid hostnames or IP addresses (no shell metacharacters).
var hostnameRe = regexp.MustCompile(`^[a-zA-Z0-9._:/\-\[\]]+$`)

// validateTarget checks that a network target (hostname/IP) is safe and within limits.
func validateTarget(target string) error {
	if target == "" {
		return fmt.Errorf("target is required")
	}
	if len(target) > maxTargetLength {
		return fmt.Errorf("target too long (max %d chars)", maxTargetLength)
	}
	if !hostnameRe.MatchString(target) {
		return fmt.Errorf("target contains invalid characters")
	}
	return nil
}

// validatePorts checks that a port list is within safe bounds.
func validatePorts(ports []int) error {
	if len(ports) == 0 {
		return nil
	}
	if len(ports) > maxPortsPerScan {
		return fmt.Errorf("too many ports (max %d)", maxPortsPerScan)
	}
	for _, p := range ports {
		if p < 1 || p > 65535 {
			return fmt.Errorf("invalid port: %d", p)
		}
	}
	return nil
}

// CountMaxMCPToolCalls is the maximum number of MCP tool calls allowed per
// AI interaction. This prevents runaway LLM tool-calling loops from exhausting
// system resources (e.g., continuous port scans).
const CountMaxMCPToolCalls = 10

// Tool represents an MCP Tool definition.
type Tool struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"inputSchema"`
}

// Server provides internal MCP-compatible tool definitions for the local AI.
type Server struct {
	pipeline *common.DataPipeline

	// Providers injected from the app layer to avoid circular dependencies
	NetProvider    NetProvider
	DevOpsProvider DevOpsProvider
}

type NetProvider interface {
	GetInterfacesDomain() []netops.InterfaceInfo
}

type DevOpsProvider interface {
	GetContainersDomain() devops.ContainerSummary
	GetKubernetesStatusDomain() devops.KubernetesStatus
}

func NewServer(pipeline *common.DataPipeline, net NetProvider, dev DevOpsProvider) *Server {
	return &Server{
		pipeline:       pipeline,
		NetProvider:    net,
		DevOpsProvider: dev,
	}
}

// ListTools returns all available MCP tools for the AI.
func (s *Server) ListTools() ([]Tool, error) {
	return []Tool{
		// ── System Tools ──────────────────────────────────────────────────────
		{
			Name:        "get_system_telemetry",
			Description: "Return a full SystemKnowledge snapshot with CPU, RAM, Disk, Network, Load, Swap, Disk I/O, Process count, and Security grade.",
			InputSchema: json.RawMessage(`{"type":"object","properties":{}}`),
		},
		{
			Name:        "get_process_list",
			Description: "Return the top-N processes sorted by CPU usage.",
			InputSchema: json.RawMessage(`{"type":"object","properties":{"n":{"type":"integer","description":"Number of processes to return (default 20)","default":20}}}`),
		},
		{
			Name:        "get_system_logs",
			Description: "Retrieve recent system logs from a given source (e.g. 'Application', 'System', 'Security').",
			InputSchema: json.RawMessage(`{"type":"object","properties":{"n":{"type":"integer","description":"Number of log entries","default":50},"source":{"type":"string","description":"Log source (Application, System, Security, etc.)","default":"Application"}}}`),
		},
		{
			Name:        "get_scheduled_tasks",
			Description: "List all scheduled tasks / cron jobs on the system.",
			InputSchema: json.RawMessage(`{"type":"object","properties":{}}`),
		},
		{
			Name:        "get_hardware_info",
			Description: "Return system hardware & platform details (OS, kernel, hostname, arch, uptime).",
			InputSchema: json.RawMessage(`{"type":"object","properties":{}}`),
		},
		{
			Name:        "get_disk_usage",
			Description: "Return per-partition disk usage statistics.",
			InputSchema: json.RawMessage(`{"type":"object","properties":{}}`),
		},
		{
			Name:        "get_baseboard_info",
			Description: "Return motherboard / baseboard hardware info (manufacturer, product, serial).",
			InputSchema: json.RawMessage(`{"type":"object","properties":{}}`),
		},

		// ── DevOps Tools ─────────────────────────────────────────────────────
		{
			Name:        "get_docker_summary",
			Description: "Return summary of Docker containers (running, stopped, failed) and a list of all containers with their status.",
			InputSchema: json.RawMessage(`{"type":"object","properties":{}}`),
		},
		{
			Name:        "get_k8s_status",
			Description: "Return Kubernetes cluster connectivity status, cluster name, node count, and pod count.",
			InputSchema: json.RawMessage(`{"type":"object","properties":{}}`),
		},
		{
			Name:        "get_k8s_pods",
			Description: "List Kubernetes pods in a given namespace (default all).",
			InputSchema: json.RawMessage(`{"type":"object","properties":{"namespace":{"type":"string","description":"Namespace to filter pods (default all)"}}}`),
		},
		{
			Name:        "get_k8s_events",
			Description: "List recent Kubernetes events for troubleshooting cluster issues.",
			InputSchema: json.RawMessage(`{"type":"object","properties":{"namespace":{"type":"string","description":"Namespace to filter events (default all)"},"limit":{"type":"integer","description":"Max events to return","default":50}}}`),
		},

		// ── Network Tools ─────────────────────────────────────────────────────
		{
			Name:        "get_network_interfaces",
			Description: "List all network interfaces with their status, IPs, MAC, speed, and real-time RX/TX rates.",
			InputSchema: json.RawMessage(`{"type":"object","properties":{}}`),
		},
		{
			Name:        "ping",
			Description: "Ping a target host or IP address.",
			InputSchema: json.RawMessage(`{"type":"object","properties":{"target":{"type":"string","description":"Hostname or IP to ping"},"count":{"type":"integer","description":"Number of pings (default 4)","default":4}},"required":["target"]}`),
		},
		{
			Name:        "dns_lookup",
			Description: "Perform a DNS lookup for a hostname.",
			InputSchema: json.RawMessage(`{"type":"object","properties":{"hostname":{"type":"string","description":"Hostname to resolve"},"server":{"type":"string","description":"Optional DNS server to query"}},"required":["hostname"]}`),
		},
		{
			Name:        "port_scan",
			Description: "Scan specific ports on a remote host.",
			InputSchema: json.RawMessage(`{"type":"object","properties":{"host":{"type":"string","description":"Host to scan"},"ports":{"type":"array","items":{"type":"integer"},"description":"List of ports to scan (e.g. [22,80,443])"}},"required":["host","ports"]}`),
		},
		{
			Name:        "traceroute",
			Description: "Trace the network path to a target host.",
			InputSchema: json.RawMessage(`{"type":"object","properties":{"target":{"type":"string","description":"Hostname or IP to trace"},"max_ttl":{"type":"integer","description":"Maximum TTL (default 30)","default":30}},"required":["target"]}`),
		},
		{
			Name:        "get_network_connections",
			Description: "List all active network connections with process association.",
			InputSchema: json.RawMessage(`{"type":"object","properties":{}}`),
		},
		{
			Name:        "get_network_health",
			Description: "Run a comprehensive network health check (gateway, DNS, internet).",
			InputSchema: json.RawMessage(`{"type":"object","properties":{}}`),
		},

		// ── Security Tools ────────────────────────────────────────────────────
		{
			Name:        "get_firewall_rules",
			Description: "List all active firewall rules.",
			InputSchema: json.RawMessage(`{"type":"object","properties":{}}`),
		},
		{
			Name:        "get_listening_ports",
			Description: "List all ports currently in LISTEN state with process info.",
			InputSchema: json.RawMessage(`{"type":"object","properties":{}}`),
		},
		{
			Name:        "get_defender_status",
			Description: "Return Windows Defender / antivirus status and signature age. (Windows only)",
			InputSchema: json.RawMessage(`{"type":"object","properties":{}}`),
		},
		{
			Name:        "run_security_audit",
			Description: "Execute a comprehensive security audit checklist.",
			InputSchema: json.RawMessage(`{"type":"object","properties":{}}`),
		},
		{
			Name:        "get_failed_logins",
			Description: "List recent failed login / authentication attempts.",
			InputSchema: json.RawMessage(`{"type":"object","properties":{}}`),
		},
		{
			Name:        "get_security_summary",
			Description: "Return the computed security summary and posture score.",
			InputSchema: json.RawMessage(`{"type":"object","properties":{}}`),
		},

		// ── Database / Storage Tools ──────────────────────────────────────────
		{
			Name:        "query_metric_history",
			Description: "Retrieve historical metric values for a named metric from the SQLite store.",
			InputSchema: json.RawMessage(`{"type":"object","properties":{"metric":{"type":"string","description":"Metric name (e.g. cpu.percent, memory.percent, disk.percent, network.rx.rate, load.1m, swap.percent)"},"limit":{"type":"integer","description":"Number of data points (default 50)","default":50}},"required":["metric"]}`),
		},
		{
			Name:        "query_events",
			Description: "Query recent timeline events by category and severity level.",
			InputSchema: json.RawMessage(`{"type":"object","properties":{"category":{"type":"string","description":"Event category filter (system, network, security, pipeline, etc.)"},"level":{"type":"string","description":"Minimum severity level (info, warning, error, critical)"},"limit":{"type":"integer","description":"Max events to return (default 50)","default":50}},"required":[]}`),
		},
		{
			Name:        "get_app_logs",
			Description: "Search internal application logs (UniversalOps system logs) by level and text pattern. Useful for self-diagnosis of app errors.",
			InputSchema: json.RawMessage(`{"type":"object","properties":{"level":{"type":"string","description":"Log level filter (info, warn, error, debug)"},"search":{"type":"string","description":"Text to search for in log messages"},"limit":{"type":"integer","description":"Max entries (default 50)","default":50}},"required":[]}`),
		},
		{
			Name:        "query_alerts",
			Description: "Return recent alert history from the alert engine.",
			InputSchema: json.RawMessage(`{"type":"object","properties":{"limit":{"type":"integer","description":"Number of alerts (default 50)","default":50}},"required":[]}`),
		},
	}, nil
}

// CallTool executes a tool call from the AI.
func (s *Server) CallTool(ctx context.Context, name string, arguments json.RawMessage) (interface{}, error) {
	switch name {
	// ── System ──────────────────────────────────────────────────────────
	case "get_system_telemetry":
		return s.handleGetTelemetry()
	case "get_process_list":
		var args struct {
			N int `json:"n"`
		}
		if err := json.Unmarshal(arguments, &args); err != nil {
			return nil, err
		}
		if args.N <= 0 {
			args.N = 20
		}
		if args.N > maxProcessListN {
			args.N = maxProcessListN
		}
		return s.handleGetProcessList(args.N)
	case "get_system_logs":
		var args struct {
			N      int    `json:"n"`
			Source string `json:"source"`
		}
		if err := json.Unmarshal(arguments, &args); err != nil {
			return nil, err
		}
		if args.N <= 0 {
			args.N = 50
		}
		if args.N > maxLogEntries {
			args.N = maxLogEntries
		}
		if args.Source == "" {
			args.Source = "Application"
		}
		return s.handleGetSystemLogs(args.N, args.Source)
	case "get_scheduled_tasks":
		return s.handleGetScheduledTasks()
	case "get_hardware_info":
		return s.handleGetHardwareInfo()
	case "get_disk_usage":
		return s.handleGetDiskUsage()
	case "get_baseboard_info":
		return s.handleGetBaseboardInfo()

	// ── DevOps ──────────────────────────────────────────────────────────
	case "get_docker_summary":
		return s.handleGetDockerSummary()
	case "get_k8s_status":
		return s.handleGetK8sStatus()
	case "get_k8s_pods":
		var args struct {
			Namespace string `json:"namespace"`
		}
		if err := json.Unmarshal(arguments, &args); err != nil {
			return nil, err
		}
		return s.handleGetK8sPods(args.Namespace)
	case "get_k8s_events":
		var args struct {
			Namespace string `json:"namespace"`
			Limit     int    `json:"limit"`
		}
		if err := json.Unmarshal(arguments, &args); err != nil {
			return nil, err
		}
		return s.handleGetK8sEvents(args.Namespace, args.Limit)

	// ── Network ─────────────────────────────────────────────────────────
	case "get_network_interfaces":
		return s.handleGetInterfaces()
	case "ping":
		var args struct {
			Target string `json:"target"`
			Count  int    `json:"count"`
		}
		if err := json.Unmarshal(arguments, &args); err != nil {
			return nil, err
		}
		if err := validateTarget(args.Target); err != nil {
			return nil, fmt.Errorf("invalid target: %w", err)
		}
		if args.Count <= 0 {
			args.Count = 4
		}
		if args.Count > maxCountValue {
			args.Count = maxCountValue
		}
		return s.handlePing(args.Target, args.Count)
	case "dns_lookup":
		var args struct {
			Hostname string `json:"hostname"`
			Server   string `json:"server"`
		}
		if err := json.Unmarshal(arguments, &args); err != nil {
			return nil, err
		}
		if err := validateTarget(args.Hostname); err != nil {
			return nil, fmt.Errorf("invalid hostname: %w", err)
		}
		if args.Server != "" {
			if err := validateTarget(args.Server); err != nil {
				return nil, fmt.Errorf("invalid server: %w", err)
			}
		}
		return s.handleDNSSLookup(args.Hostname, args.Server)
	case "port_scan":
		var args struct {
			Host  string `json:"host"`
			Ports []int  `json:"ports"`
		}
		if err := json.Unmarshal(arguments, &args); err != nil {
			return nil, err
		}
		if err := validateTarget(args.Host); err != nil {
			return nil, fmt.Errorf("invalid host: %w", err)
		}
		if err := validatePorts(args.Ports); err != nil {
			return nil, err
		}
		common.LogInfo("MCP port_scan: host=%s ports=%v", args.Host, args.Ports)
		return s.handlePortScan(args.Host, args.Ports)
	case "traceroute":
		var args struct {
			Target string `json:"target"`
			MaxTTL int    `json:"max_ttl"`
		}
		if err := json.Unmarshal(arguments, &args); err != nil {
			return nil, err
		}
		if err := validateTarget(args.Target); err != nil {
			return nil, fmt.Errorf("invalid target: %w", err)
		}
		if args.MaxTTL <= 0 {
			args.MaxTTL = 30
		}
		if args.MaxTTL > maxTTLValue {
			args.MaxTTL = maxTTLValue
		}
		return s.handleTraceRoute(args.Target, args.MaxTTL)
	case "get_network_connections":
		return s.handleGetConnections()
	case "get_network_health":
		return s.handleGetNetworkHealth()

	// ── Security ────────────────────────────────────────────────────────
	case "get_firewall_rules":
		return s.handleGetFirewallRules()
	case "get_listening_ports":
		return s.handleGetListeningPorts()
	case "get_defender_status":
		return s.handleGetDefenderStatus()
	case "run_security_audit":
		return s.handleRunSecurityAudit()
	case "get_failed_logins":
		return s.handleGetFailedLogins()
	case "get_security_summary":
		return s.handleGetSecuritySummary()

	// ── Database / Storage ──────────────────────────────────────────────
	case "query_metric_history":
		var args struct {
			Metric string `json:"metric"`
			Limit  int    `json:"limit"`
		}
		if err := json.Unmarshal(arguments, &args); err != nil {
			return nil, err
		}
		if args.Limit <= 0 {
			args.Limit = 50
		}
		if args.Limit > maxQueryLimit {
			args.Limit = maxQueryLimit
		}
		if len(args.Metric) > maxMetricNameLen {
			return nil, fmt.Errorf("metric name too long (max %d chars)", maxMetricNameLen)
		}
		return s.handleQueryMetricHistory(args.Metric, args.Limit)
	case "query_events":
		var args struct {
			Category string `json:"category"`
			Level    string `json:"level"`
			Limit    int    `json:"limit"`
		}
		if err := json.Unmarshal(arguments, &args); err != nil {
			return nil, err
		}
		if args.Limit <= 0 {
			args.Limit = 50
		}
		if args.Limit > maxQueryLimit {
			args.Limit = maxQueryLimit
		}
		return s.handleQueryEvents(args.Category, args.Level, args.Limit)
	case "get_app_logs":
		var args struct {
			Level  string `json:"level"`
			Search string `json:"search"`
			Limit  int    `json:"limit"`
		}
		if err := json.Unmarshal(arguments, &args); err != nil {
			return nil, err
		}
		if args.Limit <= 0 {
			args.Limit = 50
		}
		if args.Limit > maxQueryLimit {
			args.Limit = maxQueryLimit
		}
		if len(args.Search) > maxSearchLength {
			return nil, fmt.Errorf("search string too long (max %d chars)", maxSearchLength)
		}
		return s.handleQueryLogs(args.Level, args.Search, args.Limit)
	case "query_alerts":
		var args struct {
			Limit int `json:"limit"`
		}
		if err := json.Unmarshal(arguments, &args); err != nil {
			return nil, err
		}
		if args.Limit <= 0 {
			args.Limit = 50
		}
		if args.Limit > maxQueryLimit {
			args.Limit = maxQueryLimit
		}
		return s.handleQueryAlerts(args.Limit)

	default:
		return nil, fmt.Errorf("unknown tool: %s", name)
	}
}

// ── System Tool Handlers ──────────────────────────────────────────────────

func (s *Server) handleGetTelemetry() (interface{}, error) {
	km := common.GetKnowledge()
	if km == nil {
		return nil, fmt.Errorf("knowledge manager not initialized")
	}
	return km.GetSnapshot(), nil
}

func (s *Server) handleGetProcessList(n int) (interface{}, error) {
	procs, err := sysops.GetTopProcesses(n)
	if err != nil {
		return nil, fmt.Errorf("get processes: %w", err)
	}
	return procs, nil
}

func (s *Server) handleGetSystemLogs(n int, source string) (interface{}, error) {
	result, err := sysops.GetSystemLogs(n, source)
	if err != nil {
		return nil, fmt.Errorf("get logs: %w", err)
	}
	return result, nil
}

func (s *Server) handleGetScheduledTasks() (interface{}, error) {
	tasks, err := sysops.GetScheduledTasks()
	if err != nil {
		return nil, fmt.Errorf("get scheduled tasks: %w", err)
	}
	return tasks, nil
}

func (s *Server) handleGetHardwareInfo() (interface{}, error) {
	info, err := sysops.GetSystemInfo()
	if err != nil {
		return nil, fmt.Errorf("get system info: %w", err)
	}
	return info, nil
}

func (s *Server) handleGetDiskUsage() (interface{}, error) {
	stats, err := sysops.GetDiskStats()
	if err != nil {
		return nil, fmt.Errorf("get disk stats: %w", err)
	}
	return stats, nil
}

func (s *Server) handleGetBaseboardInfo() (interface{}, error) {
	return sysops.GetBaseboardInfo(), nil
}

// ── DevOps Tool Handlers ───────────────────────────────────────────────────

func (s *Server) handleGetDockerSummary() (interface{}, error) {
	if s.DevOpsProvider == nil {
		return nil, fmt.Errorf("DevOps provider not available")
	}
	return s.DevOpsProvider.GetContainersDomain(), nil
}

func (s *Server) handleGetK8sStatus() (interface{}, error) {
	if s.DevOpsProvider == nil {
		return nil, fmt.Errorf("DevOps provider not available")
	}
	return s.DevOpsProvider.GetKubernetesStatusDomain(), nil
}

func (s *Server) handleGetK8sPods(namespace string) (interface{}, error) {
	pods, err := devops.GetK8sPods(namespace)
	if err != nil {
		return nil, fmt.Errorf("get k8s pods: %w", err)
	}
	return pods, nil
}

func (s *Server) handleGetK8sEvents(namespace string, limit int) (interface{}, error) {
	events, err := devops.GetK8sEvents(namespace, limit)
	if err != nil {
		return nil, fmt.Errorf("get k8s events: %w", err)
	}
	return events, nil
}

// ── Network Tool Handlers ──────────────────────────────────────────────────

func (s *Server) handleGetInterfaces() (interface{}, error) {
	if s.NetProvider == nil {
		return nil, fmt.Errorf("network provider not available")
	}
	return s.NetProvider.GetInterfacesDomain(), nil
}

func (s *Server) handlePing(target string, count int) (interface{}, error) {
	result, err := netops.Ping(target, count)
	if err != nil {
		return nil, fmt.Errorf("ping %s: %w", target, err)
	}
	return result, nil
}

func (s *Server) handleDNSSLookup(hostname, server string) (interface{}, error) {
	var result *netops.DNSResult
	var err error
	if server != "" {
		result, err = netops.LookupDNS(hostname, server)
	} else {
		result, err = netops.LookupDNS(hostname)
	}
	if err != nil {
		return nil, fmt.Errorf("dns lookup %s: %w", hostname, err)
	}
	return result, nil
}

func (s *Server) handlePortScan(host string, ports []int) (interface{}, error) {
	result, err := netops.ScanPorts(host, ports)
	if err != nil {
		return nil, fmt.Errorf("port scan %s: %w", host, err)
	}
	return result, nil
}

func (s *Server) handleTraceRoute(target string, maxTTL int) (interface{}, error) {
	result, err := netops.TraceRouteWithMaxHops(target, maxTTL)
	if err != nil {
		return nil, fmt.Errorf("traceroute %s: %w", target, err)
	}
	return result, nil
}

func (s *Server) handleGetConnections() (interface{}, error) {
	conns, err := netops.GetConnections()
	if err != nil {
		return nil, fmt.Errorf("get connections: %w", err)
	}
	return conns, nil
}

func (s *Server) handleGetNetworkHealth() (interface{}, error) {
	// RunNetworkHealthCheck returns a value type (no error).
	return netops.RunNetworkHealthCheck(), nil
}

// ── Security Tool Handlers ──────────────────────────────────────────────────

func (s *Server) handleGetFirewallRules() (interface{}, error) {
	rules, err := secops.GetFirewallRules()
	if err != nil {
		return nil, fmt.Errorf("get firewall rules: %w", err)
	}
	return rules, nil
}

func (s *Server) handleGetListeningPorts() (interface{}, error) {
	ports, err := secops.GetListeningPorts()
	if err != nil {
		return nil, fmt.Errorf("get listening ports: %w", err)
	}
	return ports, nil
}

func (s *Server) handleGetDefenderStatus() (interface{}, error) {
	status, err := secops.GetDefenderStatus()
	if err != nil {
		return nil, fmt.Errorf("get defender status: %w", err)
	}
	return status, nil
}

func (s *Server) handleRunSecurityAudit() (interface{}, error) {
	result, err := secops.RunSecurityAuditChecklist()
	if err != nil {
		return nil, fmt.Errorf("security audit: %w", err)
	}
	return result, nil
}

func (s *Server) handleGetFailedLogins() (interface{}, error) {
	logins, err := secops.GetFailedLogins()
	if err != nil {
		return nil, fmt.Errorf("get failed logins: %w", err)
	}
	return logins, nil
}

func (s *Server) handleGetSecuritySummary() (interface{}, error) {
	// GetSecuritySummary returns a value type (no error).
	return secops.GetSecuritySummary(), nil
}

// ── Storage / Database Handlers ──────────────────────────────────────────────

func (s *Server) getStorageOrError() (*common.Storage, error) {
	storage := common.GetStorage()
	if storage == nil {
		return nil, fmt.Errorf("storage not initialized")
	}
	return storage, nil
}

func (s *Server) handleQueryMetricHistory(metric string, limit int) (interface{}, error) {
	storage, err := s.getStorageOrError()
	if err != nil {
		return nil, err
	}

	// Also fetch timestamps if available for richer output
	values, err := storage.GetMetricHistory(metric, limit)
	if err != nil {
		return nil, fmt.Errorf("query metric %s: %w", metric, err)
	}
	return map[string]interface{}{
		"metric": metric,
		"count":  len(values),
		"values": values,
	}, nil
}

func (s *Server) handleQueryEvents(category, level string, limit int) (interface{}, error) {
	storage, err := s.getStorageOrError()
	if err != nil {
		return nil, err
	}
	events, err := storage.QueryEvents(category, level, limit, 0)
	if err != nil {
		return nil, fmt.Errorf("query events: %w", err)
	}
	return events, nil
}

func (s *Server) handleQueryLogs(level, search string, limit int) (interface{}, error) {
	storage, err := s.getStorageOrError()
	if err != nil {
		return nil, err
	}
	entries, err := storage.QueryLogs(level, search, limit)
	if err != nil {
		return nil, fmt.Errorf("query logs: %w", err)
	}
	return entries, nil
}

func (s *Server) handleQueryAlerts(limit int) (interface{}, error) {
	storage, err := s.getStorageOrError()
	if err != nil {
		return nil, err
	}
	alerts, err := storage.QueryAlertHistory(limit)
	if err != nil {
		return nil, fmt.Errorf("query alerts: %w", err)
	}
	return alerts, nil
}

// ── Helpers ────────────────────────────────────────────────────────────────
// (reserved for future response-wrapping helpers)
