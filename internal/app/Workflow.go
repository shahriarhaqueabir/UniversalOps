package app

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/shahriarhaqueabir/AllOpsFull/internal/aiops"
	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
	"github.com/shahriarhaqueabir/AllOpsFull/internal/devops"
	"github.com/shahriarhaqueabir/AllOpsFull/internal/netops"
	"github.com/shahriarhaqueabir/AllOpsFull/internal/secops"
	"github.com/shahriarhaqueabir/AllOpsFull/internal/sysops"
)

// WorkflowAPI exposes reusable operational workflows to the frontend.
type WorkflowAPI struct {
	workflowEngine *common.WorkflowEngine
	sysOps         *SysOps
	secOps         *SecOps
	devOps         *DevOps
	alertAPI       *AlertAPI
}

func NewWorkflowAPI(workflowEngine *common.WorkflowEngine, sysOps *SysOps, secOps *SecOps, devOps *DevOps, alertAPI *AlertAPI) *WorkflowAPI {
	api := &WorkflowAPI{
		workflowEngine: workflowEngine,
		sysOps:         sysOps,
		secOps:         secOps,
		devOps:         devOps,
		alertAPI:       alertAPI,
	}
	api.RegisterDefaultWorkflows()
	return api
}

func (api *WorkflowAPI) ListWorkflows() []common.WorkflowDefinition {
	list := api.workflowEngine.List()

	// Add custom workflows from SQLite
	storage := common.GetStorage()
	if storage != nil {
		custom, _ := storage.ListCustomWorkflows()
		for _, cw := range custom {
			var def common.WorkflowDefinition
			if err := json.Unmarshal([]byte(cw["definition"]), &def); err == nil {
				list = append(list, def)
			}
		}
	}
	return list
}

// SaveCustomWorkflow persists a new or existing workflow.
func (api *WorkflowAPI) SaveCustomWorkflow(def common.WorkflowDefinition) error {
	storage := common.GetStorage()
	if storage == nil {
		return fmt.Errorf("storage unavailable")
	}
	data, err := json.Marshal(def)
	if err != nil {
		return err
	}
	return storage.UpsertCustomWorkflow(def.ID, def.Name, def.Description, string(data))
}

// DeleteCustomWorkflow removes a workflow from persistence.
func (api *WorkflowAPI) DeleteCustomWorkflow(id string) error {
	storage := common.GetStorage()
	if storage == nil {
		return fmt.Errorf("storage unavailable")
	}
	return storage.DeleteCustomWorkflow(id)
}

func (api *WorkflowAPI) GetWorkflow(id string) (common.WorkflowDefinition, error) {
	wf, ok := api.workflowEngine.Get(id)
	if !ok {
		return common.WorkflowDefinition{}, fmt.Errorf("workflow %s not found", id)
	}
	return wf, nil
}

func (api *WorkflowAPI) ExecuteWorkflow(id string) (common.WorkflowDefinition, error) {
	wf, ok := api.workflowEngine.Get(id)
	if !ok {
		return common.WorkflowDefinition{}, fmt.Errorf("workflow %s not found", id)
	}

	for i := range wf.Steps {
		step := &wf.Steps[i]
		common.LogInfo("Workflow %q: Executing step %q (%s)", id, step.ID, step.Label)

		var res any
		var err error

		if step.Action != nil {
			res, err = step.Action(context.Background())
		} else if step.Type == common.StepTypePowerShell && step.Command != "" {
			res, err = devops.RunPowerShell(step.Command)
		} else {
			err = fmt.Errorf("step %s has no action or command", step.ID)
		}

		if err != nil {
			common.LogWarn("Workflow %q: Step %q failed: %v", id, step.ID, err)
			step.Error = err.Error()
			// ABORT POLICY: Workflows are transactional by default. Stop on first error.
			return wf, fmt.Errorf("workflow aborted at step %q: %w", step.ID, err)
		}

		step.Result = res
	}

	return wf, nil
}

func (api *WorkflowAPI) RegisterDefaultWorkflows() {
	engine := api.workflowEngine

	// ── Workflow: Diagnose Slow PC ──
	engine.Register(common.WorkflowDefinition{
		ID:          "diag-slow-pc",
		Name:        "Diagnose Slow PC",
		Description: "Identify resource bottlenecks and rogue processes.",
		Category:    common.WorkflowCategorySystem,
		Why:         "Context switching between Task Manager, Resource Monitor, and terminal is inefficient. This correlates CPU, RAM, and Disk I/O instantly.",
		Risks:       []string{"Minimal", "Collection has negligible overhead"},
		TypicalVals: "CPU < 70%, RAM < 85%, Disk < 90%",
		Steps: []common.WorkflowStep{
			{
				ID:              "check-cpu",
				Label:           "CPU Pressure Audit",
				Description:     "Collect per-core utilization and load averages.",
				Command:         "Get-WmiObject Win32_Processor",
				ExpectedOutcome: "Identify core saturation or frequency throttling.",
				Action: func(ctx context.Context) (any, error) {
					return api.sysOps.GetCPUInfo(), nil
				},
			},
			{
				ID:              "check-mem",
				Label:           "Memory Saturation",
				Description:     "Measure RAM and swap usage.",
				Command:         "Get-WmiObject Win32_OperatingSystem",
				ExpectedOutcome: "Determine if paging/swapping is causing lag.",
				Action: func(ctx context.Context) (any, error) {
					return api.sysOps.GetMemoryInfo(), nil
				},
			},
			{
				ID:              "top-procs",
				Label:           "Impactful Processes",
				Description:     "Filter top 10 processes by resource consumption.",
				Command:         "Get-Process | Sort-Object CPU -Descending | Select-Object -First 10",
				ExpectedOutcome: "Point to the exact PID responsible for the slowdown.",
				Action: func(ctx context.Context) (any, error) {
					return api.sysOps.GetTopProcesses(10), nil
				},
			},
		},
	})

	// ── Workflow: Quick Security Sweep ──
	engine.Register(common.WorkflowDefinition{
		ID:          "sec-sweep",
		Name:        "Quick Security Sweep",
		Description: "Verify basic system hardening and perimeter safety.",
		Category:    common.WorkflowCategorySecurity,
		Why:         "Manually checking firewall, listeners, and defender status is error-prone.",
		Risks:       []string{"None", "Read-only operation"},
		TypicalVals: "Firewall ON, No 0.0.0.0 listeners without reason",
		Steps: []common.WorkflowStep{
			{
				ID:              "firewall-status",
				Type:            common.StepTypePowerShell,
				Label:           "Firewall State",
				Description:     "Check Domain, Private, and Public profile states.",
				Command:         "Get-NetFirewallProfile | Select Name, Enabled",
				ExpectedOutcome: "Ensure all profiles are ACTIVE.",
				Action: func(ctx context.Context) (any, error) {
					return api.secOps.GetFirewallStatus(), nil
				},
			},
			{
				ID:              "check-listeners",
				Type:            common.StepTypePowerShell,
				Label:           "External Listeners",
				Description:     "Identify processes listening on all interfaces (0.0.0.0).",
				Command:         "Get-NetTCPConnection -State Listen | Where-Object { $_.LocalAddress -eq '0.0.0.0' }",
				ExpectedOutcome: "Flag any suspicious exposed services.",
				Action: func(ctx context.Context) (any, error) {
					return api.secOps.GetListeningPorts(), nil
				},
			},
		},
	})

	// ── Workflow: Core Maintenance ──
	engine.Register(common.WorkflowDefinition{
		ID:          "core-maintenance",
		Name:        "Core Maintenance",
		Description: "Comprehensive system health sweep: resources, alerts, storage, processes, and services.",
		Category:    common.WorkflowCategorySystem,
		Why:         "Routine maintenance prevents drift. This workflow captures a full-system snapshot in one pass, replacing fragmented manual checks across multiple tools.",
		Risks:       []string{"None", "Read-only, zero side effects"},
		TypicalVals: "CPU <70%, RAM <85%, Disk <90%, 0 critical alerts",
		Steps: []common.WorkflowStep{
			{
				ID:              "health-snapshot",
				Label:           "System Health Snapshot",
				Description:     "Capture CPU, RAM, and Disk utilization at this moment.",
				Command:         "Get-WmiObject Win32_Processor,Win32_OperatingSystem,Win32_LogicalDisk",
				ExpectedOutcome: "Baseline resource utilization within normal thresholds.",
				Action: func(ctx context.Context) (any, error) {
					return map[string]any{
						"cpu":    api.sysOps.GetCPUInfo(),
						"memory": api.sysOps.GetMemoryInfo(),
						"disk":   api.sysOps.GetDiskInfo(),
					}, nil
				},
			},
			{
				ID:              "alert-review",
				Label:           "Active Alert Review",
				Description:     "Enumerate all active security and system alerts.",
				Command:         "Get-WinEvent -LogName Security -MaxEvents 20",
				ExpectedOutcome: "Zero critical or high-severity unacknowledged alerts.",
				Action: func(ctx context.Context) (any, error) {
					return api.alertAPI.GetActiveAlerts(), nil
				},
			},
			{
				ID:              "storage-integrity",
				Label:           "Storage Integrity Check",
				Description:     "Verify disk partition health and free-space margins.",
				Command:         "Get-PSDrive -PSProvider FileSystem | Select-Object Name,Used,Free",
				ExpectedOutcome: "All volumes have >15% free space; no orphaned mount points.",
				Action: func(ctx context.Context) (any, error) {
					return api.sysOps.GetDiskInfo(), nil
				},
			},
			{
				ID:              "process-audit",
				Label:           "Process Audit",
				Description:     "Identify the top 15 processes by CPU and memory consumption.",
				Command:         "Get-Process | Sort-Object CPU -Descending | Select-Object -First 15",
				ExpectedOutcome: "No unexpected high-CPU processes; confirm known services.",
				Action: func(ctx context.Context) (any, error) {
					return api.sysOps.GetTopProcesses(15), nil
				},
			},
			{
				ID:              "service-health",
				Label:           "Service Health",
				Description:     "Enumerate all running and stopped Windows services.",
				Command:         "Get-Service | Select-Object Name,Status,StartType | Sort-Object Status",
				ExpectedOutcome: "All critical services are running; no unexpected stopped services.",
				Action: func(ctx context.Context) (any, error) {
					return api.devOps.GetServices(), nil
				},
			},
		},
	})

	// ── Workflow: Full Network Diagnostic ──
	engine.Register(common.WorkflowDefinition{
		ID:          "net-diag",
		Name:        "Full Network Diagnostic",
		Description: "Multi-layered network sanity check (Ping, DNS, Ports, Trace, Conns).",
		Category:    common.WorkflowCategoryNetwork,
		Why:         "Network issues are often multi-faceted. This correlates connectivity, resolution, and exposure in one pass.",
		Risks:       []string{"Minimal", "Port scan restricted to common local ports"},
		TypicalVals: "DNS resolves <100ms, Ping loss 0%",
		Steps: []common.WorkflowStep{
			{
				ID:              "ping-check",
				Label:           "Ping Connectivity",
				Description:     "Ping default gateway and public endpoint (8.8.8.8).",
				Command:         "ping -n 3 8.8.8.8",
				ExpectedOutcome: "Packet loss < 1%, latency within expected range.",
				Action: func(ctx context.Context) (any, error) {
					return netops.Ping("8.8.8.8", 3)
				},
			},
			{
				ID:              "dns-check",
				Label:           "DNS Resolution",
				Description:     "Resolve google.com via system DNS.",
				Command:         "nslookup google.com",
				ExpectedOutcome: "Resolution succeeds within 100ms.",
				Action: func(ctx context.Context) (any, error) {
					return netops.LookupDNS("google.com")
				},
			},
			{
				ID:              "port-scan",
				Label:           "Local Port Scan",
				Description:     "Scan common ports on localhost for exposed services.",
				Command:         "netstat -an | findstr LISTENING",
				ExpectedOutcome: "No unexpected open ports on all interfaces.",
				Action: func(ctx context.Context) (any, error) {
					return netops.ScanPorts("localhost", netops.DefaultScanPorts())
				},
			},
			{
				ID:              "trace-route",
				Label:           "Traceroute",
				Description:     "Trace the route to 8.8.8.8.",
				Command:         "tracert 8.8.8.8",
				ExpectedOutcome: "Full path identified; no unexpected routing loops.",
				Action: func(ctx context.Context) (any, error) {
					return netops.TraceRoute("8.8.8.8")
				},
			},
			{
				ID:              "conn-audit",
				Label:           "Connection Audit",
				Description:     "Enumerate active TCP connections.",
				Command:         "netstat -ano",
				ExpectedOutcome: "All connections accounted for; no suspicious outbound links.",
				Action: func(ctx context.Context) (any, error) {
					return netops.GetConnections()
				},
			},
		},
	})

	// ── Workflow: Comprehensive Security Audit ──
	engine.Register(common.WorkflowDefinition{
		ID:          "sec-audit",
		Name:        "Comprehensive Security Audit",
		Description: "Deep inspection of firewall, users, listeners, defender, and events.",
		Category:    common.WorkflowCategorySecurity,
		Why:         "Compliance and safety require a broad view of system posture.",
		Risks:       []string{"None", "Read-only enumeration"},
		TypicalVals: "Zero unauthorized listeners, Defender ACTIVE",
		Steps: []common.WorkflowStep{
			{
				ID:              "firewall-rules",
				Label:           "Firewall Rules",
				Description:     "Enumerate all firewall rules with direction and action.",
				Command:         "Get-NetFirewallRule | Select Name, Direction, Action",
				ExpectedOutcome: "No disabled or overly permissive rules.",
				Action: func(ctx context.Context) (any, error) {
					return secops.GetFirewallRules()
				},
			},
			{
				ID:              "user-audit",
				Label:           "User Account Audit",
				Description:     "List all local user accounts and admin group membership.",
				Command:         "Get-LocalUser; net localgroup administrators",
				ExpectedOutcome: "Limited admin accounts, no unexpected enabled users.",
				Action: func(ctx context.Context) (any, error) {
					return secops.GetUsers()
				},
			},
			{
				ID:              "listeners",
				Label:           "Listening Ports",
				Description:     "Identify processes with open network listeners.",
				Command:         "Get-NetTCPConnection -State Listen",
				ExpectedOutcome: "Only expected services are listening.",
				Action: func(ctx context.Context) (any, error) {
					return secops.GetListeningPorts()
				},
			},
			{
				ID:              "defender",
				Label:           "Defender Status",
				Description:     "Verify Windows Defender real-time protection and signature age.",
				Command:         "Get-MpComputerStatus",
				ExpectedOutcome: "Real-time protection ON, signatures up to date.",
				Action: func(ctx context.Context) (any, error) {
					return secops.GetDefenderStatus()
				},
			},
			{
				ID:              "sec-events",
				Label:           "Security Events",
				Description:     "Review recent security log events.",
				Command:         "Get-WinEvent -LogName Security -MaxEvents 50",
				ExpectedOutcome: "No critical security events in the recent window.",
				Action: func(ctx context.Context) (any, error) {
					return secops.GetSecurityEvents()
				},
			},
		},
	})

	// ── Workflow: Deep System Health Check ──
	engine.Register(common.WorkflowDefinition{
		ID:          "deep-health",
		Name:        "Deep System Health Check",
		Description: "Detailed hardware and OS health inspection.",
		Category:    common.WorkflowCategorySystem,
		Why:         "Correlating CPU, Memory, Disk, and Processes in a single report identifies cross-resource contention.",
		Risks:       []string{"None"},
		TypicalVals: "CPU < 70%, RAM < 85%, Disk < 90%",
		Steps: []common.WorkflowStep{
			{
				ID:              "cpu-stats",
				Label:           "CPU Statistics",
				Description:     "Per-core utilization and load averages.",
				Command:         "Get-WmiObject Win32_Processor",
				ExpectedOutcome: "Identify core saturation or frequency throttling.",
				Action: func(ctx context.Context) (any, error) {
					return sysops.GetCPUStats()
				},
			},
			{
				ID:              "mem-stats",
				Label:           "Memory Statistics",
				Description:     "RAM usage, available bytes, and swap analysis.",
				Command:         "Get-WmiObject Win32_OperatingSystem",
				ExpectedOutcome: "Determine if paging/swapping is causing lag.",
				Action: func(ctx context.Context) (any, error) {
					return sysops.GetMemoryStats()
				},
			},
			{
				ID:              "disk-usage",
				Label:           "Disk Usage",
				Description:     "Partition utilization and free space.",
				Command:         "Get-PSDrive -PSProvider FileSystem",
				ExpectedOutcome: "All volumes have >15% free space.",
				Action: func(ctx context.Context) (any, error) {
					return sysops.GetDiskStats()
				},
			},
			{
				ID:              "system-info",
				Label:           "System Information",
				Description:     "OS details, uptime, and process count.",
				Command:         "Get-WmiObject Win32_OperatingSystem",
				ExpectedOutcome: "OS version and uptime within expected bounds.",
				Action: func(ctx context.Context) (any, error) {
					return sysops.GetSystemInfo()
				},
			},
			{
				ID:              "top-procs",
				Label:           "Top Processes",
				Description:     "Identify top 10 processes by resource consumption.",
				Command:         "Get-Process | Sort-Object CPU -Descending | Select-Object -First 10",
				ExpectedOutcome: "No unexpected high-CPU processes.",
				Action: func(ctx context.Context) (any, error) {
					return sysops.GetTopProcesses(10)
				},
			},
		},
	})

	// ── Workflow: AI System Auditor ──
	engine.Register(common.WorkflowDefinition{
		ID:          "ai-audit",
		Name:        "AI System Auditor",
		Description: "Multi-layered audit with AI-powered risk assessment.",
		Category:    common.WorkflowCategoryIntelligence,
		Why:         "Traditional reports require human interpretation. This uses local AI to synthesize system, security, and network data into actionable findings.",
		Risks:       []string{"None", "Read-only, high CPU during AI processing"},
		TypicalVals: "AI Confidence > 80%",
		Steps: []common.WorkflowStep{
			{
				ID:              "collect-data",
				Label:           "Collect Audit Data",
				Description:     "Gather system health, security posture, and network diagnostics.",
				ExpectedOutcome: "Aggregated raw data for AI analysis.",
				Action: func(ctx context.Context) (any, error) {
					var errs []string
					h, err := sysops.RunHealthCheck()
					if err != nil {
						errs = append(errs, fmt.Sprintf("Health: %v", err))
					}
					s, err := secops.RunSecurityAudit()
					if err != nil {
						errs = append(errs, fmt.Sprintf("Security: %v", err))
					}
					n, err := netops.RunNetworkDiagnostics()
					if err != nil {
						errs = append(errs, fmt.Sprintf("Network: %v", err))
					}
					result := map[string]any{}
					if h != nil {
						result["health"] = h.String()
					} else {
						result["health"] = "Health check unavailable"
					}
					if s != nil {
						result["security"] = s.String()
					} else {
						result["security"] = "Security audit unavailable"
					}
					if n != nil {
						result["network"] = n.String()
					} else {
						result["network"] = "Network diagnostics unavailable"
					}
					if len(errs) > 0 {
						result["warnings"] = strings.Join(errs, "; ")
					}
					return result, nil
				},
			},
			{
				ID:              "ai-analyze",
				Label:           "AI Vulnerability Analysis",
				Description:     "Send aggregated data to Hawk for cross-layer risk correlation.",
				ExpectedOutcome: "Hawk identification of security risks or performance bottlenecks.",
				Action: func(ctx context.Context) (any, error) {
					var errs []string
					h, err := sysops.RunHealthCheck()
					if err != nil {
						errs = append(errs, fmt.Sprintf("Health: %v", err))
					}
					s, err := secops.RunSecurityAudit()
					if err != nil {
						errs = append(errs, fmt.Sprintf("Security: %v", err))
					}
					n, err := netops.RunNetworkDiagnostics()
					if err != nil {
						errs = append(errs, fmt.Sprintf("Network: %v", err))
					}
					data := map[string]string{}
					if h != nil {
						data["System Health"] = h.String()
					} else {
						data["System Health"] = "Unavailable"
					}
					if s != nil {
						data["Security Posture"] = s.String()
					} else {
						data["Security Posture"] = "Unavailable"
					}
					if n != nil {
						data["Network Status"] = n.String()
					} else {
						data["Network Status"] = "Unavailable"
					}
					if len(errs) > 0 {
						data["Warnings"] = strings.Join(errs, "; ")
					}
					report := aiops.GenerateEnhancedReport("Full System Audit", data)
					return report.OllamaReport()
				},
			},
		},
	})

	// ── Workflow: DevOps Environment Audit ──
	engine.Register(common.WorkflowDefinition{
		ID:          "devops-audit",
		Name:        "DevOps Environment Audit",
		Description: "Verify development tools, services, and process environment.",
		Category:    common.WorkflowCategoryDevOps,
		Why:         "Development environments can drift. This ensures Go, Git, and critical services are available and healthy.",
		Risks:       []string{"None"},
		TypicalVals: "Go/Git installed, Core services RUNNING",
		Steps: []common.WorkflowStep{
			{
				ID:              "check-git",
				Label:           "Version Control (Git)",
				Description:     "Verify Git CLI availability.",
				Command:         "git version",
				ExpectedOutcome: "Git is installed and functional.",
				Action: func(ctx context.Context) (any, error) {
					diag := devops.RunDevOpsDiagnostics()
					for _, c := range diag.Checks {
						if c.Name == "Git" {
							return c, nil
						}
					}
					return devops.DiagnosticCheck{Name: "Git", Status: "fail", Message: "Git check unavailable"}, nil
				},
			},
			{
				ID:              "check-docker-k8s",
				Label:           "Container Platforms (Docker, K8s)",
				Description:     "Verify Docker daemon and kubectl cluster connectivity.",
				Command:         "docker info; kubectl cluster-info",
				ExpectedOutcome: "Docker daemon running, K8s cluster reachable or tool available.",
				Action: func(ctx context.Context) (any, error) {
					diag := devops.RunDevOpsDiagnostics()
					var results []devops.DiagnosticCheck
					for _, c := range diag.Checks {
						if strings.Contains(c.Name, "Docker") || strings.Contains(c.Name, "Kubernetes") || strings.Contains(c.Name, "kubectl") {
							results = append(results, c)
						}
					}
					return results, nil
				},
			},
			{
				ID:              "check-sdk",
				Label:           "SDK Availability (Node.js, Go)",
				Description:     "Verify Node.js and Go runtimes.",
				Command:         "node --version; go version",
				ExpectedOutcome: "SDKs are installed and at expected versions.",
				Action: func(ctx context.Context) (any, error) {
					diag := devops.RunDevOpsDiagnostics()
					var results []devops.DiagnosticCheck
					for _, c := range diag.Checks {
						if c.Name == "Node.js" || c.Name == "Go" {
							results = append(results, c)
						}
					}
					return results, nil
				},
			},
			{
				ID:              "check-services",
				Label:           "Running Services",
				Description:     "Enumerate running background services.",
				Command:         "Get-Service | Where-Object Status -eq Running",
				ExpectedOutcome: "All critical services are running.",
				Action: func(ctx context.Context) (any, error) {
					return api.devOps.GetServices(), nil
				},
			},
		},
	})

	// ── PowerShell Diagnostic Workflows ──
	// These delegate to the PowerShell profile functions sourced from
	// profiles/powershell_profile.ps1. The WorkflowEngine executes them
	// via devops.RunPowerShell when Action is nil and Type is PowerShell.

	// ── Workflow: Daily Operations Snapshot ──
	engine.Register(common.WorkflowDefinition{
		ID:          "ps-daily-ops",
		Name:        "Daily Operations Snapshot",
		Description: "Quick overview of OS, disk volumes, performance counters, and top processes.",
		Category:    common.WorkflowCategorySystem,
		Why:         "A single command that replaces four separate checks — ideal for a morning health check.",
		Risks:       []string{"None", "Read-only, lightweight counters"},
		TypicalVals: "CPU < 70%, Disk > 15% free",
		Steps: []common.WorkflowStep{
			{
				ID:              "run-daily-ops",
				Type:            common.StepTypePowerShell,
				Label:           "Run Daily Ops Snapshot",
				Description:     "Collect OS info, volume usage, performance counters, and top 5 processes.",
				Command:         "Invoke-OpsDailyOps",
				ExpectedOutcome: "Formatted report of system vitals for the day.",
			},
		},
	})

	// ── Workflow: System Review ──
	engine.Register(common.WorkflowDefinition{
		ID:          "ps-system-review",
		Name:        "System Review",
		Description: "Comprehensive system information audit — OS, CPU, Memory, Disks, Volumes, BIOS, GPU, and uptime.",
		Category:    common.WorkflowCategorySystem,
		Why:         "One-shot inventory of every hardware and OS detail. Use before capacity planning or handoff.",
		Risks:       []string{"None", "Read-only CIM queries"},
		TypicalVals: "All components reported, no errors",
		Steps: []common.WorkflowStep{
			{
				ID:              "run-system-review",
				Type:            common.StepTypePowerShell,
				Label:           "Run System Review",
				Description:     "Execute full system info audit across all hardware and OS layers.",
				Command:         "Invoke-OpsSystemReview",
				ExpectedOutcome: "Complete hardware and OS inventory report.",
			},
		},
	})

	// ── Workflow: Security Surface Audit ──
	engine.Register(common.WorkflowDefinition{
		ID:          "ps-security-audit",
		Name:        "Security Surface Audit",
		Description: "Enumerate admin groups and verify Windows Defender protection status.",
		Category:    common.WorkflowCategorySecurity,
		Why:         "Quickly confirm who has admin access and whether real-time protection is active.",
		Risks:       []string{"None", "Read-only enumeration"},
		TypicalVals: "Defender ON, limited admin group membership",
		Steps: []common.WorkflowStep{
			{
				ID:              "run-sec-audit",
				Type:            common.StepTypePowerShell,
				Label:           "Run Security Audit",
				Description:     "Check administrator group membership and antivirus status.",
				Command:         "Invoke-OpsSecurityAudit",
				ExpectedOutcome: "List of admin users and Defender protection status.",
			},
		},
	})

	// ── Workflow: Network Diagnostics ──
	engine.Register(common.WorkflowDefinition{
		ID:          "ps-network-diagnostics",
		Name:        "Network Diagnostics",
		Description: "Detailed adapter configuration and active TCP connections snapshot.",
		Category:    common.WorkflowCategoryNetwork,
		Why:         "Correlate IP, DNS, gateway, and live connections to pinpoint network misconfigurations.",
		Risks:       []string{"Minimal", "Read-only netstat snapshot"},
		TypicalVals: "All adapters have valid IP, no unexpected listeners",
		Steps: []common.WorkflowStep{
			{
				ID:              "run-net-diag",
				Type:            common.StepTypePowerShell,
				Label:           "Run Network Diagnostics",
				Description:     "Enumerate adapter details and capture top 20 TCP connections.",
				Command:         "Invoke-OpsNetworkDiagnostics",
				ExpectedOutcome: "Adapter configuration table and active connection list.",
			},
		},
	})

	// ── Workflow: Threat Hunting Primer ──
	engine.Register(common.WorkflowDefinition{
		ID:          "ps-threat-hunt",
		Name:        "Threat Hunting Primer",
		Description: "Surface suspicious listening sockets and recently modified system executables.",
		Category:    common.WorkflowCategorySecurity,
		Why:         "First-pass triage for compromised systems — unusual sockets and new binaries are strong indicators.",
		Risks:       []string{"Minimal", "Read-only enumeration"},
		TypicalVals: "No unexpected LISTENING sockets, System32 binaries unchanged",
		Steps: []common.WorkflowStep{
			{
				ID:              "run-threat-hunt",
				Type:            common.StepTypePowerShell,
				Label:           "Run Threat Hunt",
				Description:     "Scan for suspicious sockets and recently modified executables in System32.",
				Command:         "Invoke-OpsThreatHunt",
				ExpectedOutcome: "List of listening sockets and recently modified system binaries.",
			},
		},
	})

	// ── Workflow: System Change Audit ──
	engine.Register(common.WorkflowDefinition{
		ID:          "ps-change-audit",
		Name:        "System Change Audit",
		Description: "Review recently installed programs from the Windows registry.",
		Category:    common.WorkflowCategorySystem,
		Why:         "Detect unauthorized or unexpected software installations.",
		Risks:       []string{"None", "Read-only registry query"},
		TypicalVals: "Only known software in install history",
		Steps: []common.WorkflowStep{
			{
				ID:              "run-change-audit",
				Type:            common.StepTypePowerShell,
				Label:           "Run Change Audit",
				Description:     "Query the registry for the 10 most recently installed programs.",
				Command:         "Invoke-OpsChangeAudit",
				ExpectedOutcome: "Sorted list of recent installations with dates.",
			},
		},
	})

	// ── Workflow: Compliance Check ──
	engine.Register(common.WorkflowDefinition{
		ID:          "ps-compliance-check",
		Name:        "Compliance Check",
		Description: "Audit password policy and account lockout settings.",
		Category:    common.WorkflowCategorySecurity,
		Why:         "Verify that domain or local password policies meet security baselines.",
		Risks:       []string{"None", "Read-only policy query"},
		TypicalVals: "Password complexity enforced, lockout after 5 attempts",
		Steps: []common.WorkflowStep{
			{
				ID:              "run-compliance",
				Type:            common.StepTypePowerShell,
				Label:           "Run Compliance Check",
				Description:     "Retrieve current password and lockout policy settings.",
				Command:         "Invoke-OpsComplianceCheck",
				ExpectedOutcome: "Password policy details and lockout thresholds.",
			},
		},
	})
}
