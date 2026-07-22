package app

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
	"github.com/shahriarhaqueabir/AllOpsFull/internal/devops"
	"github.com/shahriarhaqueabir/AllOpsFull/internal/netops"
	"github.com/shahriarhaqueabir/AllOpsFull/internal/secops"
	"github.com/shahriarhaqueabir/AllOpsFull/internal/sysops"
	"github.com/shahriarhaqueabir/AllOpsFull/internal/aiops"
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
		Why:         "Network issues are often multi-faceted. This correlates connectivity, resolution, and exposure in one pass.",
		Risks:       []string{"Minimal", "Port scan restricted to common local ports"},
		TypicalVals: "DNS resolves <100ms, Ping loss 0%",
		Steps: []common.WorkflowStep{
			{
				ID:              "run-diag",
				Label:           "Run Core Diagnostic",
				Description:     "Execute parallel ping, dns, trace, and connection scans.",
				Command:         "netops.RunNetworkDiagnostics()",
				ExpectedOutcome: "Full diagnostic report with markdown summary.",
				Action: func(ctx context.Context) (any, error) {
					res, err := netops.RunNetworkDiagnostics()
					if err != nil {
						return nil, err
					}
					return res.Markdown(), nil
				},
			},
		},
	})

	// ── Workflow: Comprehensive Security Audit ──
	engine.Register(common.WorkflowDefinition{
		ID:          "sec-audit",
		Name:        "Comprehensive Security Audit",
		Description: "Deep inspection of firewall, users, listeners, defender, and events.",
		Why:         "Compliance and safety require a broad view of system posture.",
		Risks:       []string{"None", "Read-only enumeration"},
		TypicalVals: "Zero unauthorized listeners, Defender ACTIVE",
		Steps: []common.WorkflowStep{
			{
				ID:              "run-audit",
				Label:           "Run Full Security Audit",
				Description:     "Correlate OS security parameters and identify risks.",
				Command:         "secops.RunSecurityAudit()",
				ExpectedOutcome: "Detailed risk assessment and posture summary.",
				Action: func(ctx context.Context) (any, error) {
					res, err := secops.RunSecurityAudit()
					if err != nil {
						return nil, err
					}
					return res.Markdown(), nil
				},
			},
		},
	})

	// ── Workflow: Deep System Health Check ──
	engine.Register(common.WorkflowDefinition{
		ID:          "deep-health",
		Name:        "Deep System Health Check",
		Description: "Detailed hardware and OS health inspection.",
		Why:         "Correlating CPU, Memory, Disk, and Processes in a single report identifies cross-resource contention.",
		Risks:       []string{"None"},
		TypicalVals: "CPU < 70%, RAM < 85%, Disk < 90%",
		Steps: []common.WorkflowStep{
			{
				ID:              "run-health",
				Label:           "Run Full Health Scan",
				Description:     "Enumerate per-core stats, memory pressure, and top processes.",
				Command:         "sysops.RunHealthCheck()",
				ExpectedOutcome: "Hardware status and process resource consumption report.",
				Action: func(ctx context.Context) (any, error) {
					res, err := sysops.RunHealthCheck()
					if err != nil {
						return nil, err
					}
					return res.Markdown(), nil
				},
			},
		},
	})

	// ── Workflow: AI System Auditor ──
	engine.Register(common.WorkflowDefinition{
		ID:          "ai-audit",
		Name:        "AI System Auditor",
		Description: "Multi-layered audit with AI-powered risk assessment.",
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
					h, _ := sysops.RunHealthCheck()
					s, _ := secops.RunSecurityAudit()
					n, _ := netops.RunNetworkDiagnostics()
					return map[string]string{
						"health":   h.String(),
						"security": s.String(),
						"network":  n.String(),
					}, nil
				},
			},
			{
				ID:              "ai-analyze",
				Label:           "AI Vulnerability Analysis",
				Description:     "Send aggregated data to Hawk for cross-layer risk correlation.",
				ExpectedOutcome: "Hawk identification of security risks or performance bottlenecks.",
				Action: func(ctx context.Context) (any, error) {
					h, _ := sysops.RunHealthCheck()
					s, _ := secops.RunSecurityAudit()
					n, _ := netops.RunNetworkDiagnostics()
					data := map[string]string{
						"System Health":    h.String(),
						"Security Posture": s.String(),
						"Network Status":   n.String(),
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
		Why:         "Development environments can drift. This ensures Go, Git, and critical services are available and healthy.",
		Risks:       []string{"None"},
		TypicalVals: "Go/Git installed, Core services RUNNING",
		Steps: []common.WorkflowStep{
			{
				ID:              "run-dev-diag",
				Label:           "Run Tooling Audit",
				Description:     "Check compiler versions, git status, and process lists.",
				Command:         "devops.RunDevDiagnostics()",
				ExpectedOutcome: "Status report on dev tools and background services.",
				Action: func(ctx context.Context) (any, error) {
					res, err := devops.RunDevDiagnostics()
					if err != nil {
						return nil, err
					}
					return res.Markdown(), nil
				},
			},
		},
	})
}
