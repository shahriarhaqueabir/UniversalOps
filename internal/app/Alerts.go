package app

import (
	"github.com/shahriarhaqueabir/UniversalOps/internal/common"
)

// AlertAPI exposes alert management bindings to the frontend.
type AlertAPI struct {
	alerts *common.AlertEngine
}

// NewAlertAPI creates a new AlertAPI facade.
func NewAlertAPI(alerts *common.AlertEngine) *AlertAPI {
	return &AlertAPI{alerts: alerts}
}

// GetActiveAlerts returns all currently unresolved alerts.
func (a *AlertAPI) GetActiveAlerts() []AlertInfo {
	alerts := a.alerts.ActiveAlerts()
	out := make([]AlertInfo, 0, len(alerts))
	for _, alert := range alerts {
		out = append(out, convertAlert(alert))
	}
	return out
}

// GetAlertHistory returns all alerts that have ever fired.
func (a *AlertAPI) GetAlertHistory() []AlertInfo {
	alerts := a.alerts.AllAlerts()
	out := make([]AlertInfo, 0, len(alerts))
	for _, alert := range alerts {
		out = append(out, convertAlert(alert))
	}
	return out
}

// GetRules returns all currently active alert rules.
func (a *AlertAPI) GetRules() []AlertRuleInfo {
	rules := a.alerts.GetRules()
	out := make([]AlertRuleInfo, 0, len(rules))
	for _, r := range rules {
		out = append(out, AlertRuleInfo{
			Metric:    r.Metric,
			Condition: r.Condition.String(),
			Threshold: r.Threshold,
			FlapCount: r.FlapCount,
			Severity:  r.Severity.String(),
			Message:   r.Message,
		})
	}
	return out
}

// ResolveAlert marks a specific alert as resolved by its ID.
func (a *AlertAPI) ResolveAlert(id string) {
	a.alerts.ResolveAlert(id)
}

// AddRule adds a new alert rule.
// severity: "info", "warning", "critical"
// condition: "gt" (greater than) or "lt" (less than)
func (a *AlertAPI) AddRule(metric string, threshold float64, severity string, condition string) {
	level := common.AlertWarning
	switch severity {
	case "info":
		level = common.AlertInfo
	case "critical":
		level = common.AlertCritical
	}

	cond := common.AlertGT
	if condition == "lt" {
		cond = common.AlertLT
	}

	rule := common.AlertRule{
		Metric:    metric,
		Condition: cond,
		Threshold: threshold,
		FlapCount: 2,
		Severity:  level,
	}
	a.alerts.AddRule(rule)
}

// RemoveRule removes an alert rule by metric and threshold.
func (a *AlertAPI) RemoveRule(metric string, threshold float64) {
	a.alerts.RemoveRule(metric, threshold)
}

// GetAlertCount returns the number of unresolved alerts.
func (a *AlertAPI) GetAlertCount() int {
	return a.alerts.AlertCount()
}

// EvaluateNow forces an immediate alert evaluation.
func (a *AlertAPI) EvaluateNow() []AlertInfo {
	fired := a.alerts.Evaluate()
	out := make([]AlertInfo, 0, len(fired))
	for _, alert := range fired {
		out = append(out, convertAlert(alert))
	}
	return out
}
