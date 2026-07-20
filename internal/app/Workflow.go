package app

import (
	"context"
	"fmt"

	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
	"github.com/shahriarhaqueabir/AllOpsFull/internal/devops"
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
	return api.workflowEngine.List()
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
		if step.Action != nil {
			res, err := step.Action(context.Background())
			if err != nil {
				step.Error = err.Error()
			} else {
				step.Result = res
			}
		} else if step.Type == common.StepTypePowerShell && step.Command != "" {
			// Transparently execute the command string for PowerShell steps
			res, err := devops.RunPowerShell(step.Command)
			if err != nil {
				step.Error = err.Error()
			} else {
				step.Result = res
			}
		}
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
}
