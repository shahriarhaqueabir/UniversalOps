package common

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// ── AlertLevel ──────────────────────────────────────────────────────────────

// AlertLevel represents the severity of an alert.
type AlertLevel int

const (
	AlertInfo AlertLevel = iota
	AlertWarning
	AlertCritical
)

func (l AlertLevel) String() string {
	switch l {
	case AlertInfo:
		return "INFO"
	case AlertWarning:
		return "WARNING"
	case AlertCritical:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}

// ── AlertCondition ──────────────────────────────────────────────────────────

// AlertCondition specifies how a metric value is compared to a threshold.
type AlertCondition int

const (
	AlertGT AlertCondition = iota // value > threshold
	AlertLT                       // value < threshold
)

// maxStoredAlerts caps the in-memory alert history so a flapping rule can't
// grow ae.alerts without bound for the lifetime of the process (M2). Oldest
// *resolved* alerts are dropped first; active alerts are always kept.
const maxStoredAlerts = 2000

func (c AlertCondition) String() string {
	switch c {
	case AlertGT:
		return ">"
	case AlertLT:
		return "<"
	default:
		return "?"
	}
}

// ── AlertRule ───────────────────────────────────────────────────────────────

// AlertRule defines a single alerting condition.
type AlertRule struct {
	Metric    string         // canonical metric name (e.g. "cpu.percent")
	Condition AlertCondition // GT or LT
	Threshold float64        // threshold value
	FlapCount int            // consecutive violations before the alert fires
	Severity  AlertLevel     // severity once activated
	Message   string         // optional template; may contain {value} and {threshold}
}

func (r AlertRule) evalMessage(value float64) string {
	if r.Message != "" {
		msg := strings.ReplaceAll(r.Message, "{value}", fmt.Sprintf("%.1f", value))
		msg = strings.ReplaceAll(msg, "{threshold}", fmt.Sprintf("%.1f", r.Threshold))
		msg = strings.ReplaceAll(msg, "{metric}", r.Metric)
		return msg
	}
	return fmt.Sprintf("%s %s %.1f (current %.1f)", r.Metric, r.Condition, r.Threshold, value)
}

// ── Alert ───────────────────────────────────────────────────────────────────

// Alert represents a single fired alert instance.
type Alert struct {
	ID        string
	Level     AlertLevel
	Metric    string
	Condition AlertCondition
	Message   string
	Value     float64
	Threshold float64
	Timestamp time.Time
	Resolved  bool
}

// AlertResolvedCallback is called when an alert transitions from active to resolved.
type AlertResolvedCallback func(resolved Alert)

// ── AlertEngine ─────────────────────────────────────────────────────────────

// AlertEngine evaluates alert rules against the data pipeline with flap
// detection and deduplication.
type AlertEngine struct {
	mu        sync.RWMutex
	pipeline  *DataPipeline
	rules     []AlertRule
	alerts    []Alert
	flapCount map[string]int // "metric:cond:threshold" -> consecutive hits
	alertKeys map[string]int // active alert key -> index in alerts slice
	nextID    int

	// OnAlertResolved is called when an active alert transitions to resolved.
	// The AlertEngine does NOT persist to DB — the caller is responsible for
	// persisting the resolution (via UpdateAlertResolved or similar).
	OnAlertResolved AlertResolvedCallback
}

// NewAlertEngine creates an alert engine attached to the given pipeline.
func NewAlertEngine(pipeline *DataPipeline) *AlertEngine {
	return &AlertEngine{
		pipeline:  pipeline,
		flapCount: make(map[string]int),
		alertKeys: make(map[string]int),
	}
}

// AddRule registers a new alert rule.
func (ae *AlertEngine) AddRule(rule AlertRule) {
	ae.mu.Lock()
	defer ae.mu.Unlock()
	ae.rules = append(ae.rules, rule)
}

// RemoveRule removes the first matching rule by metric and threshold.
func (ae *AlertEngine) RemoveRule(metric string, threshold float64) {
	ae.mu.Lock()
	defer ae.mu.Unlock()
	for i, r := range ae.rules {
		if r.Metric == metric && r.Threshold == threshold {
			ae.rules = append(ae.rules[:i], ae.rules[i+1:]...)
			return
		}
	}
}

// AddDefaultRules registers sensible default thresholds for CPU, memory, disk,
// and CPU temperature.
func (ae *AlertEngine) AddDefaultRules() {
	defaults := []AlertRule{
		{
			Metric:    MetricCPU,
			Condition: AlertGT,
			Threshold: 90,
			FlapCount: 2,
			Severity:  AlertCritical,
		},
		{
			Metric:    MetricCPU,
			Condition: AlertGT,
			Threshold: 70,
			FlapCount: 3,
			Severity:  AlertWarning,
		},
		{
			Metric:    MetricMem,
			Condition: AlertGT,
			Threshold: 90,
			FlapCount: 2,
			Severity:  AlertCritical,
		},
		{
			Metric:    MetricMem,
			Condition: AlertGT,
			Threshold: 80,
			FlapCount: 3,
			Severity:  AlertWarning,
		},
		{
			Metric:    MetricDisk,
			Condition: AlertGT,
			Threshold: 95,
			FlapCount: 2,
			Severity:  AlertCritical,
		},
		{
			Metric:    MetricDisk,
			Condition: AlertGT,
			Threshold: 85,
			FlapCount: 3,
			Severity:  AlertWarning,
		},
	}
	for _, r := range defaults {
		ae.AddRule(r)
	}
}

// ── Alert Correlation ──────────────────────────────────────────────────────

// Incident represents a group of correlated alerts occurring within a close time window.
type Incident struct {
	ID        string     `json:"id"`
	Title     string     `json:"title"`
	Metrics   []string   `json:"metrics"`
	Timestamp time.Time  `json:"timestamp"`
	Severity  AlertLevel `json:"severity"`
}

// CorrelateAlerts groups recent alerts into a single incident if they affect related subsystems.
func (ae *AlertEngine) CorrelateAlerts(newAlerts []Alert) *Incident {
	if len(newAlerts) < 2 {
		return nil
	}

	// Simple correlation: If we have multiple alerts on CPU/Memory/Disk at once, it's a "Resource Contention" incident
	resourceMetrics := 0
	metrics := []string{}
	maxSeverity := AlertInfo

	for _, a := range newAlerts {
		if a.Metric == MetricCPU || a.Metric == MetricMem || a.Metric == MetricDisk {
			resourceMetrics++
		}
		metrics = append(metrics, a.Metric)
		if a.Level > maxSeverity {
			maxSeverity = a.Level
		}
	}

	if resourceMetrics >= 2 {
		incID := fmt.Sprintf("inc-%d", time.Now().Unix())
		inc := &Incident{
			ID:        incID,
			Title:     "Multi-Subsystem Resource Pressure",
			Metrics:   metrics,
			Timestamp: time.Now(),
			Severity:  maxSeverity,
		}

		// PERSIST: Save to incidents table
		if s := GetStorage(); s != nil {
			_ = s.InsertIncident(IncidentRecord{
				ID:        incID,
				Timestamp: inc.Timestamp.UTC().Format(time.RFC3339),
				Title:     inc.Title,
				Details:   fmt.Sprintf("Correlated breach detected across: %v", metrics),
				Severity:  inc.Severity.String(),
			})
		}

		return inc
	}

	return nil
}

// Evaluate checks all rules against the latest pipeline data and returns any
// newly fired or resolved alerts.
func (ae *AlertEngine) Evaluate() []Alert {
	ae.mu.Lock()
	defer ae.mu.Unlock()

	var fired []Alert

	for _, rule := range ae.rules {
		// Get the last value from the pipeline
		mf := ae.pipeline.GetMetricWithForecast(rule.Metric)
		if len(mf.Values) == 0 {
			continue // no data yet
		}
		val := mf.LastValue

		// Check condition
		violates := false
		switch rule.Condition {
		case AlertGT:
			violates = val > rule.Threshold
		case AlertLT:
			violates = val < rule.Threshold
		}

		// Flap key
		flapKey := fmt.Sprintf("%s:%s:%.4f", rule.Metric, rule.Condition, rule.Threshold)

		if violates {
			ae.flapCount[flapKey]++

			alertKey := fmt.Sprintf("%s:%s:%.4f:%d", rule.Metric, rule.Condition, rule.Threshold, rule.Severity)
			if _, exists := ae.alertKeys[alertKey]; exists {
				// Already fired – update value, keep firing
				idx := ae.alertKeys[alertKey]
				ae.alerts[idx].Value = val
				continue
			}

			// Only fire once flap threshold is reached
			if ae.flapCount[flapKey] >= rule.FlapCount {
				a := Alert{
					ID:        fmt.Sprintf("alert-%d", ae.nextID),
					Level:     rule.Severity,
					Metric:    rule.Metric,
					Condition: rule.Condition,
					Message:   rule.evalMessage(val),
					Value:     val,
					Threshold: rule.Threshold,
					Timestamp: time.Now(),
				}
				ae.nextID++
				ae.alerts = append(ae.alerts, a)
				ae.alertKeys[alertKey] = len(ae.alerts) - 1
				fired = append(fired, a)
				ae.trimLocked()
			}
		} else {
			// Violation cleared
			delete(ae.flapCount, flapKey)

			alertKey := fmt.Sprintf("%s:%s:%.4f:%d", rule.Metric, rule.Condition, rule.Threshold, rule.Severity)
			if idx, exists := ae.alertKeys[alertKey]; exists {
				ae.alerts[idx].Resolved = true
				ae.alerts[idx].Value = val
				delete(ae.alertKeys, alertKey)

				// Notify the caller so the resolution can be persisted to DB
				if ae.OnAlertResolved != nil {
					ae.OnAlertResolved(ae.alerts[idx])
				}
			}
		}
	}

	return fired
}

// trimLocked drops the oldest resolved alerts once the history exceeds
// maxStoredAlerts, then rebuilds alertKeys to match the new indices. Callers
// must already hold ae.mu for writing.
func (ae *AlertEngine) trimLocked() {
	if len(ae.alerts) <= maxStoredAlerts {
		return
	}
	overflow := len(ae.alerts) - maxStoredAlerts
	kept := ae.alerts[:0:0]
	dropped := 0
	for _, a := range ae.alerts {
		if dropped < overflow && a.Resolved {
			dropped++
			continue
		}
		kept = append(kept, a)
	}
	ae.alerts = kept
	for k := range ae.alertKeys {
		delete(ae.alertKeys, k)
	}
	for i, a := range ae.alerts {
		if !a.Resolved {
			alertKey := fmt.Sprintf("%s:%s:%.4f:%d", a.Metric, a.Condition, a.Threshold, a.Level)
			ae.alertKeys[alertKey] = i
		}
	}
}

// ActiveAlerts returns all currently unresolved alerts.
func (ae *AlertEngine) ActiveAlerts() []Alert {
	ae.mu.RLock()
	defer ae.mu.RUnlock()

	var active []Alert
	for _, a := range ae.alerts {
		if !a.Resolved {
			active = append(active, a)
		}
	}
	return active
}

// AllAlerts returns every alert that has ever been fired.
func (ae *AlertEngine) AllAlerts() []Alert {
	ae.mu.RLock()
	defer ae.mu.RUnlock()

	out := make([]Alert, len(ae.alerts))
	copy(out, ae.alerts)
	return out
}

// ResolveAlert marks a specific alert as resolved by its ID.
func (ae *AlertEngine) ResolveAlert(id string) {
	ae.mu.Lock()
	defer ae.mu.Unlock()

	for i := range ae.alerts {
		if ae.alerts[i].ID == id {
			ae.alerts[i].Resolved = true
			// Remove from alertKeys map
			for k, idx := range ae.alertKeys {
				if idx == i {
					delete(ae.alertKeys, k)
					break
				}
			}
			return
		}
	}
}

// GetRules returns all currently active alert rules.
func (ae *AlertEngine) GetRules() []AlertRule {
	ae.mu.RLock()
	defer ae.mu.RUnlock()

	out := make([]AlertRule, len(ae.rules))
	copy(out, ae.rules)
	return out
}

// AlertCount returns the number of unresolved alerts.
func (ae *AlertEngine) AlertCount() int {
	ae.mu.RLock()
	defer ae.mu.RUnlock()

	count := 0
	for _, a := range ae.alerts {
		if !a.Resolved {
			count++
		}
	}
	return count
}

// ── DB Restore ──────────────────────────────────────────────────────────────

// parseAlertLevel converts a level string back to AlertLevel.
func parseAlertLevel(s string) AlertLevel {
	switch s {
	case "CRITICAL":
		return AlertCritical
	case "WARNING":
		return AlertWarning
	default:
		return AlertInfo
	}
}

// RestoreFromDB loads persisted alerts from storage into the engine's in-memory
// state. This is called once at startup so that alert history survives restarts.
// Active (unresolved) alerts are re-indexed in alertKeys so the engine can
// update or resolve them on subsequent Evaluate() cycles.
func (ae *AlertEngine) RestoreFromDB(records []AlertRecord) {
	ae.mu.Lock()
	defer ae.mu.Unlock()

	if len(records) == 0 {
		return
	}

	// Pre-allocate the alerts slice
	ae.alerts = make([]Alert, 0, len(records))
	ae.alertKeys = make(map[string]int)

	for _, r := range records {
		a := Alert{
			ID:        r.ID,
			Level:     parseAlertLevel(r.Level),
			Metric:    r.Metric,
			Condition: AlertGT, // condition not persisted; default is fine for display
			Message:   r.Message,
			Value:     r.Value,
			Threshold: r.Threshold,
			Timestamp: r.Timestamp,
			Resolved:  r.Resolved,
		}
		ae.alerts = append(ae.alerts, a)

		// Index active alerts so Evaluate() can update/resolve them
		if !a.Resolved {
			alertKey := fmt.Sprintf("%s:%s:%.4f:%d", a.Metric, a.Condition, a.Threshold, a.Level)
			ae.alertKeys[alertKey] = len(ae.alerts) - 1
		}
	}

	// Advance nextID past any restored IDs
	for _, r := range records {
		var idNum int
		if _, err := fmt.Sscanf(r.ID, "alert-%d", &idNum); err == nil && idNum >= ae.nextID {
			ae.nextID = idNum + 1
		}
	}

	LogInfo("AlertEngine: restored %d alerts from DB (%d active)", len(ae.alerts), len(ae.alertKeys))
}
