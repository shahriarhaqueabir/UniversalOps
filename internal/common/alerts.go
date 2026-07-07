package common

import (
	"fmt"
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
		return r.Message
	}
	return fmt.Sprintf("%s %s %.1f (current %.1f)", r.Metric, r.Condition, r.Threshold, value)
}

// ── Alert ───────────────────────────────────────────────────────────────────

// Alert represents a single fired alert instance.
type Alert struct {
	ID        string
	Level     AlertLevel
	Metric    string
	Message   string
	Value     float64
	Threshold float64
	Timestamp time.Time
	Resolved  bool
}

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

			alertKey := fmt.Sprintf("%s:%.4f:%d", rule.Metric, rule.Threshold, rule.Severity)
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
					Message:   rule.evalMessage(val),
					Value:     val,
					Threshold: rule.Threshold,
					Timestamp: time.Now(),
				}
				ae.nextID++
				ae.alerts = append(ae.alerts, a)
				ae.alertKeys[alertKey] = len(ae.alerts) - 1
				fired = append(fired, a)
			}
		} else {
			// Violation cleared
			delete(ae.flapCount, flapKey)

			alertKey := fmt.Sprintf("%s:%.4f:%d", rule.Metric, rule.Threshold, rule.Severity)
			if idx, exists := ae.alertKeys[alertKey]; exists {
				ae.alerts[idx].Resolved = true
				ae.alerts[idx].Value = val
				delete(ae.alertKeys, alertKey)
			}
		}
	}

	return fired
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
