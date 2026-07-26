package app

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/shahriarhaqueabir/UniversalOps/internal/common"
)

// ReportTypeMeta describes a report type available for manual generation.
type ReportTypeMeta struct {
	Type        string `json:"type"`
	Label       string `json:"label"`
	Description string `json:"description"`
	Available   bool   `json:"available"`
}

// ReportGenerationResult holds the outcome of a report generation request.
type ReportGenerationResult struct {
	ReportID  string `json:"report_id"`
	Type      string `json:"type"`
	Timestamp string `json:"timestamp"`
	Score     int    `json:"score"`
	Summary   string `json:"summary,omitempty"`
}

// PrebuiltReportTemplate defines a canned auto-report rule template.
// Name follows the convention: {Topic}-{Condition}{Threshold} — PascalCase,
// hyphen-separated tokens with a suffixed condition+value indicator.
type PrebuiltReportTemplate struct {
	ID          string  `json:"id"`
	Category    string  `json:"category"` // "health", "security", "performance"
	PresetName  string  `json:"preset_name"`
	Description string  `json:"description"`
	Metric      string  `json:"metric"`
	Condition   string  `json:"condition"`
	Threshold   float64 `json:"threshold"`
	ReportType  string  `json:"report_type"`
	Schedule    string  `json:"schedule"`
}

// prebuiltTemplates returns the static catalog of prebuilt report templates.
func prebuiltTemplates() []PrebuiltReportTemplate {
	return []PrebuiltReportTemplate{
		{
			ID:          "prebuilt-health-high-cpu",
			Category:    "health",
			PresetName:  "CPU-Util-GT-90",
			Description: "Triggers a health report when CPU utilization exceeds 90%.",
			Metric:      "cpu.percent",
			Condition:   "GT",
			Threshold:   90,
			ReportType:  "health",
			Schedule:    scheduleOnAlert,
		},
		{
			ID:          "prebuilt-health-mem-pressure",
			Category:    "health",
			PresetName:  "Memory-Used-GT-95",
			Description: "Triggers a health report when memory usage exceeds 95%.",
			Metric:      "memory.used_percent",
			Condition:   "GT",
			Threshold:   95,
			ReportType:  "health",
			Schedule:    scheduleOnAlert,
		},
		{
			ID:          "prebuilt-health-disk-space",
			Category:    "health",
			PresetName:  "Disk-Used-GT-90",
			Description: "Generates a daily health report when any disk exceeds 90% usage.",
			Metric:      "disk.used_percent",
			Condition:   "GT",
			Threshold:   90,
			ReportType:  "health",
			Schedule:    scheduleDaily,
		},
		{
			ID:          "prebuilt-health-high-process",
			Category:    "health",
			PresetName:  "Process-Count-GT-250",
			Description: "Triggers an hourly health report when total process count exceeds 250.",
			Metric:      "process.count",
			Condition:   "GT",
			Threshold:   250,
			ReportType:  "health",
			Schedule:    scheduleHourly,
		},
		{
			ID:          "prebuilt-health-network-surge",
			Category:    "health",
			PresetName:  "Network-Rx-GT-1Gbps",
			Description: "Triggers an hourly health report when network receive rate exceeds 1 Gbps.",
			Metric:      "network.rx_rate",
			Condition:   "GT",
			Threshold:   1_000_000_000,
			ReportType:  "health",
			Schedule:    scheduleHourly,
		},
		{
			ID:          "prebuilt-security-port-scan",
			Category:    "security",
			PresetName:  "Open-Ports-GT-50",
			Description: "Triggers a security audit when open port count exceeds 50 (possible scan).",
			Metric:      "connections.open_ports",
			Condition:   "GT",
			Threshold:   50,
			ReportType:  "security",
			Schedule:    scheduleOnAlert,
		},
		{
			ID:          "prebuilt-security-suspicious-conns",
			Category:    "security",
			PresetName:  "External-Conns-GT-20",
			Description: "Triggers a security audit when external connections exceed 20.",
			Metric:      "connections.external",
			Condition:   "GT",
			Threshold:   20,
			ReportType:  "security",
			Schedule:    scheduleOnAlert,
		},
		{
			ID:          "prebuilt-security-fw-change",
			Category:    "security",
			PresetName:  "FW-Rules-Changed",
			Description: "Triggers a security audit when firewall rule count changes from baseline (threshold > 0).",
			Metric:      "security.firewall_rules",
			Condition:   "GT",
			Threshold:   0,
			ReportType:  "security",
			Schedule:    scheduleHourly,
		},
		{
			ID:          "prebuilt-perf-baseline-drift",
			Category:    "performance",
			PresetName:  "Baseline-Deviation-GT-2σ",
			Description: "Generates an AI diagnostic when baseline deviation exceeds 2 sigma.",
			Metric:      "baseline.deviation",
			Condition:   "GT",
			Threshold:   2,
			ReportType:  "auto_diag",
			Schedule:    scheduleDaily,
		},
		{
			ID:          "prebuilt-perf-system-health",
			Category:    "performance",
			PresetName:  "System-Health-LT-70",
			Description: "Generates an AI diagnostic when the overall system health score drops below 70.",
			Metric:      "system.health_score",
			Condition:   "LT",
			Threshold:   70,
			ReportType:  "auto_diag",
			Schedule:    scheduleDaily,
		},
	}
}

// ReportsAPI exposes consolidated report retrieval and generation to the frontend.
type ReportsAPI struct {
	sysOps *SysOps
	secOps *SecOps
	aiOps  *AIOps
}

// NewReportsAPI creates a new ReportsAPI facade.
func NewReportsAPI(sysOps *SysOps, secOps *SecOps, aiOps *AIOps) *ReportsAPI {
	return &ReportsAPI{
		sysOps: sysOps,
		secOps: secOps,
		aiOps:  aiOps,
	}
}

// ListAllReports returns all persisted reports (health, security, auto_diag) aggregated by recency.
func (r *ReportsAPI) ListAllReports() []common.ReportRecord {
	storage := common.GetStorage()
	if storage == nil {
		return []common.ReportRecord{}
	}
	reports, _ := storage.ListAllReports()
	return reports
}

// DeleteReport removes a single report by ID.
func (r *ReportsAPI) DeleteReport(id string) bool {
	storage := common.GetStorage()
	if storage == nil {
		return false
	}
	err := storage.DeleteReport(id)
	return err == nil
}

// GetReportTypes returns available report types with metadata.
func (r *ReportsAPI) GetReportTypes() []ReportTypeMeta {
	storage := common.GetStorage()
	_ = storage // used for future availability checks
	return []ReportTypeMeta{
		{
			Type:        "health",
			Label:       "Health Diagnostic",
			Description: "Full system health check — CPU, memory, disk, processes, and uptime.",
			Available:   r.sysOps != nil,
		},
		{
			Type:        "security",
			Label:       "Security Audit",
			Description: "Comprehensive security audit — open ports, firewall rules, listeners, UAC status.",
			Available:   r.secOps != nil,
		},
		{
			Type:        "auto_diag",
			Label:       "Auto-Diagnostic",
			Description: "AI-powered diagnostic report with system analysis and recommendations.",
			Available:   r.aiOps != nil,
		},
	}
}

// GenerateReport triggers generation of a report of the given type.
// Health and security reports are self-persisted by their subsystem methods.
// Auto-diagnostic reports are persisted here.
// Supported types: "health", "security", "auto_diag".
func (r *ReportsAPI) GenerateReport(reportType string) (*ReportGenerationResult, error) {
	switch reportType {
	case "health":
		if r.sysOps == nil {
			return nil, fmt.Errorf("SysOps subsystem unavailable")
		}
		result := r.sysOps.RunDiagnostic()
		return &ReportGenerationResult{
			ReportID:  result.ID,
			Type:      "health",
			Timestamp: result.Timestamp,
			Score:     result.Score,
			Summary:   fmt.Sprintf("Health score: %d/100 — checked %d categories", result.Score, len(result.Checks)),
		}, nil

	case "security":
		if r.secOps == nil {
			return nil, fmt.Errorf("SecOps subsystem unavailable")
		}
		result := r.secOps.RunAudit()
		return &ReportGenerationResult{
			ReportID:  result.ID,
			Type:      "security",
			Timestamp: result.Timestamp,
			Score:     result.Score,
			Summary:   fmt.Sprintf("Security score: %d/100 — %d passed, %d failed", result.Score, result.Passed, result.Failed),
		}, nil

	case "auto_diag":
		if r.aiOps == nil {
			return nil, fmt.Errorf("AIOps subsystem unavailable")
		}
		sections := []string{
			"System Overview",
			"Resource Utilization",
			"Network Health",
			"Security Posture",
			"Recommendations",
		}
		content := r.aiOps.GenerateReport(sections)
		data, _ := json.Marshal(map[string]interface{}{
			"sections":     sections,
			"content":      content,
			"generated_by": "AIOps",
		})
		storage := common.GetStorage()
		if storage == nil {
			return nil, fmt.Errorf("storage unavailable")
		}
		id := fmt.Sprintf("auto-diag-%d", time.Now().Unix())
		score := int(r.aiOps.GetConfidenceScore().Overall)
		err := storage.InsertReport(common.ReportRecord{
			ID:        id,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Type:      "auto_diag",
			Score:     score,
			DataJSON:  string(data),
		})
		if err != nil {
			return nil, fmt.Errorf("persist auto_diag report: %w", err)
		}
		return &ReportGenerationResult{
			ReportID:  id,
			Type:      "auto_diag",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Score:     score,
			Summary:   "AI-powered diagnostic generated successfully",
		}, nil

	default:
		return nil, fmt.Errorf("unknown report type: %q (supported: health, security, auto_diag)", reportType)
	}
}

// GetReport retrieves a single full report with data_json by ID.
func (r *ReportsAPI) GetReport(id string) (*common.ReportRecord, error) {
	storage := common.GetStorage()
	if storage == nil {
		return nil, fmt.Errorf("storage unavailable")
	}
	return storage.GetReport(id)
}

// =============================================================================
// Prebuilt Report Templates
// =============================================================================

// GetPrebuiltTemplates returns the catalog of prebuilt report templates.
func (r *ReportsAPI) GetPrebuiltTemplates() []PrebuiltReportTemplate {
	return prebuiltTemplates()
}

// AddRuleFromTemplate creates a new auto-report rule from a prebuilt template ID.
func (r *ReportsAPI) AddRuleFromTemplate(templateID string) (*common.AutoReportRule, error) {
	for _, t := range prebuiltTemplates() {
		if t.ID == templateID {
			return r.AddReportRule(t.PresetName, t.Description, t.Metric, t.Condition, t.Threshold, t.ReportType, t.Schedule)
		}
	}
	return nil, fmt.Errorf("unknown prebuilt template: %q", templateID)
}

// =============================================================================
// Auto-Report Rules Management
// =============================================================================

// GetReportRules returns all configured auto-report rules.
func (r *ReportsAPI) GetReportRules() []common.AutoReportRule {
	storage := common.GetStorage()
	if storage == nil {
		return []common.AutoReportRule{}
	}
	rules, _ := storage.ListReportRules()
	return rules
}

const (
	scheduleOnAlert = "on_alert"
	scheduleHourly  = "hourly"
	scheduleDaily   = "daily"
)

// AddReportRule creates a new auto-report rule.
func (r *ReportsAPI) AddReportRule(name, description, metric, condition string, threshold float64, reportType, schedule string) (*common.AutoReportRule, error) {
	storage := common.GetStorage()
	if storage == nil {
		return nil, fmt.Errorf("storage unavailable")
	}

	// Validate schedule
	switch schedule {
	case scheduleOnAlert, scheduleHourly, scheduleDaily:
		// valid
	default:
		schedule = scheduleOnAlert
	}

	// Validate condition
	if condition != "GT" && condition != "LT" {
		condition = "GT"
	}

	// Validate report type
	switch reportType {
	case "health", "security", "auto_diag":
		// valid
	default:
		reportType = "health"
	}

	now := time.Now().UTC().Format(time.RFC3339)
	rule := common.AutoReportRule{
		ID:          fmt.Sprintf("rule-%d", time.Now().UnixNano()),
		Name:        name,
		Description: description,
		Metric:      metric,
		Condition:   condition,
		Threshold:   threshold,
		ReportType:  reportType,
		Schedule:    schedule,
		Enabled:     true,
		CreatedAt:   now,
	}

	if err := storage.InsertReportRule(rule); err != nil {
		return nil, fmt.Errorf("insert rule: %w", err)
	}
	return &rule, nil
}

// UpdateReportRule modifies an existing auto-report rule.
func (r *ReportsAPI) UpdateReportRule(ruleID, name, description string, enabled bool) error {
	storage := common.GetStorage()
	if storage == nil {
		return fmt.Errorf("storage unavailable")
	}
	existing, err := storage.GetReportRule(ruleID)
	if err != nil {
		return err
	}
	if existing == nil {
		return fmt.Errorf("rule %q not found", ruleID)
	}

	existing.Name = name
	existing.Description = description
	existing.Enabled = enabled

	return storage.UpdateReportRule(*existing)
}

// DeleteReportRule removes a report rule by ID.
func (r *ReportsAPI) DeleteReportRule(ruleID string) bool {
	storage := common.GetStorage()
	if storage == nil {
		return false
	}
	return storage.DeleteReportRule(ruleID) == nil
}

// =============================================================================
// Implement common.ReportTrigger interface (used by EngineLoop)
// =============================================================================

// TriggerReport implements common.ReportTrigger.
// It generates a report of the given type and returns the report ID.
func (r *ReportsAPI) TriggerReport(reportType string) (string, error) {
	result, err := r.GenerateReport(reportType)
	if err != nil {
		return "", err
	}
	return result.ReportID, nil
}

// GetEnabledReportRules implements common.ReportTrigger.
func (r *ReportsAPI) GetEnabledReportRules() ([]common.AutoReportRule, error) {
	storage := common.GetStorage()
	if storage == nil {
		return nil, nil
	}
	all, err := storage.ListReportRules()
	if err != nil {
		return nil, err
	}
	var enabled []common.AutoReportRule
	for _, rule := range all {
		if rule.Enabled {
			enabled = append(enabled, rule)
		}
	}
	return enabled, nil
}

// TouchRule implements common.ReportTrigger.
func (r *ReportsAPI) TouchRule(ruleID string) error {
	storage := common.GetStorage()
	if storage == nil {
		return fmt.Errorf("storage unavailable")
	}
	return storage.TouchReportRule(ruleID)
}
