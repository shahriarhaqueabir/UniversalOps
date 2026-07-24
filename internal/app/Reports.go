package app

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
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
			ReportID:  fmt.Sprintf("health-%d", time.Now().Unix()),
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
			ReportID:  fmt.Sprintf("sec-audit-%d", time.Now().Unix()),
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
		err := storage.InsertReport(common.ReportRecord{
			ID:        id,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Type:      "auto_diag",
			Score:     0,
			DataJSON:  string(data),
		})
		if err != nil {
			return nil, fmt.Errorf("persist auto_diag report: %w", err)
		}
		return &ReportGenerationResult{
			ReportID:  id,
			Type:      "auto_diag",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Score:     0,
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
