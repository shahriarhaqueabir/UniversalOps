package common

import (
	"testing"
)

func TestNewAlertEngine(t *testing.T) {
	dp := NewDataPipeline(DefaultCollectionConfig())
	ae := NewAlertEngine(dp)
	if ae == nil {
		t.Fatal("NewAlertEngine returned nil")
	}
	if ae.AlertCount() != 0 {
		t.Errorf("expected initial alert count 0, got %d", ae.AlertCount())
	}
}

func TestAlertRuleFiresCorrectly(t *testing.T) {
	dp := NewDataPipeline(CollectionConfig{Capacity: 10})
	ae := NewAlertEngine(dp)

	// Rule: CPU > 80, flap count 1 (immediate)
	ae.AddRule(AlertRule{
		Metric:    MetricCPU,
		Condition: AlertGT,
		Threshold: 80,
		FlapCount: 1,
		Severity:  AlertWarning,
	})

	// Push value that violates
	dp.PushMetric(MetricCPU, "%", 90)

	fired := ae.Evaluate()
	if len(fired) != 1 {
		t.Fatalf("expected 1 alert fired, got %d", len(fired))
	}
	if fired[0].Metric != MetricCPU {
		t.Errorf("expected metric %s, got %s", MetricCPU, fired[0].Metric)
	}
	if fired[0].Value != 90 {
		t.Errorf("expected value 90, got %f", fired[0].Value)
	}

	// Should be in active alerts
	active := ae.ActiveAlerts()
	if len(active) != 1 {
		t.Errorf("expected 1 active alert, got %d", len(active))
	}
}

func TestAlertRuleNoFireBelowThreshold(t *testing.T) {
	dp := NewDataPipeline(CollectionConfig{Capacity: 10})
	ae := NewAlertEngine(dp)

	ae.AddRule(AlertRule{
		Metric:    MetricMem,
		Condition: AlertGT,
		Threshold: 80,
		FlapCount: 1,
		Severity:  AlertWarning,
	})

	dp.PushMetric(MetricMem, "%", 50)

	fired := ae.Evaluate()
	if len(fired) != 0 {
		t.Errorf("expected 0 alerts for safe value, got %d", len(fired))
	}
}

func TestFlapDetectionPreventsFalsePositives(t *testing.T) {
	dp := NewDataPipeline(CollectionConfig{Capacity: 10})
	ae := NewAlertEngine(dp)

	// Rule requires 3 consecutive violations before firing
	ae.AddRule(AlertRule{
		Metric:    MetricCPU,
		Condition: AlertGT,
		Threshold: 80,
		FlapCount: 3,
		Severity:  AlertCritical,
	})

	// Two violations in a row – not enough for flap=3
	dp.PushMetric(MetricCPU, "%", 90)
	ae.Evaluate()
	dp.PushMetric(MetricCPU, "%", 91)
	fired := ae.Evaluate()

	if len(fired) != 0 {
		t.Errorf("expected 0 alerts before flap threshold met, got %d", len(fired))
	}

	// Third consecutive violation should fire
	dp.PushMetric(MetricCPU, "%", 92)
	fired = ae.Evaluate()

	if len(fired) != 1 {
		t.Fatalf("expected 1 alert after 3 violations, got %d", len(fired))
	}
}

func TestFlapResetOnRecovery(t *testing.T) {
	dp := NewDataPipeline(CollectionConfig{Capacity: 10})
	ae := NewAlertEngine(dp)

	ae.AddRule(AlertRule{
		Metric:    MetricCPU,
		Condition: AlertGT,
		Threshold: 80,
		FlapCount: 2,
		Severity:  AlertWarning,
	})

	// One violation, then recovery
	dp.PushMetric(MetricCPU, "%", 90)
	ae.Evaluate()

	dp.PushMetric(MetricCPU, "%", 50) // back to normal
	ae.Evaluate()

	// Two more violations should fire now (flap counter was reset)
	dp.PushMetric(MetricCPU, "%", 90)
	ae.Evaluate()
	dp.PushMetric(MetricCPU, "%", 91)
	fired := ae.Evaluate()

	if len(fired) != 1 {
		t.Errorf("expected 1 alert after reset+2 violations, got %d", len(fired))
	}
}

func TestAlertAutoResolveOnRecovery(t *testing.T) {
	dp := NewDataPipeline(CollectionConfig{Capacity: 10})
	ae := NewAlertEngine(dp)

	ae.AddRule(AlertRule{
		Metric:    MetricCPU,
		Condition: AlertGT,
		Threshold: 80,
		FlapCount: 1,
		Severity:  AlertWarning,
	})

	// Fire the alert
	dp.PushMetric(MetricCPU, "%", 90)
	ae.Evaluate()

	if ae.AlertCount() != 1 {
		t.Fatalf("expected 1 active alert, got %d", ae.AlertCount())
	}

	// Recover
	dp.PushMetric(MetricCPU, "%", 50)
	ae.Evaluate()

	active := ae.ActiveAlerts()
	if len(active) != 0 {
		t.Errorf("expected 0 active alerts after recovery, got %d", len(active))
	}
}

func TestAlertLessThanCondition(t *testing.T) {
	dp := NewDataPipeline(CollectionConfig{Capacity: 10})
	ae := NewAlertEngine(dp)

	ae.AddRule(AlertRule{
		Metric:    MetricMem,
		Condition: AlertLT,
		Threshold: 10,
		FlapCount: 1,
		Severity:  AlertWarning,
	})

	dp.PushMetric(MetricMem, "%", 5)
	fired := ae.Evaluate()

	if len(fired) != 1 {
		t.Fatalf("expected 1 alert for AlertLT, got %d", len(fired))
	}
	if fired[0].Value != 5 {
		t.Errorf("expected value 5, got %f", fired[0].Value)
	}
}

func TestResolveAlert(t *testing.T) {
	dp := NewDataPipeline(CollectionConfig{Capacity: 10})
	ae := NewAlertEngine(dp)

	ae.AddRule(AlertRule{
		Metric:    MetricCPU,
		Condition: AlertGT,
		Threshold: 80,
		FlapCount: 1,
		Severity:  AlertWarning,
	})

	dp.PushMetric(MetricCPU, "%", 90)
	fired := ae.Evaluate()

	if len(fired) != 1 {
		t.Fatalf("expected 1 fired alert, got %d", len(fired))
	}

	ae.ResolveAlert(fired[0].ID)
	if ae.AlertCount() != 0 {
		t.Errorf("expected 0 active alerts after resolve, got %d", ae.AlertCount())
	}
}

func TestMultipleRulesDifferentMetrics(t *testing.T) {
	dp := NewDataPipeline(CollectionConfig{Capacity: 10})
	ae := NewAlertEngine(dp)

	ae.AddRule(AlertRule{
		Metric:    MetricCPU,
		Condition: AlertGT,
		Threshold: 80,
		FlapCount: 1,
		Severity:  AlertWarning,
	})
	ae.AddRule(AlertRule{
		Metric:    MetricMem,
		Condition: AlertGT,
		Threshold: 85,
		FlapCount: 1,
		Severity:  AlertCritical,
	})

	dp.PushMetric(MetricCPU, "%", 90)
	dp.PushMetric(MetricMem, "%", 95)

	fired := ae.Evaluate()
	if len(fired) != 2 {
		t.Fatalf("expected 2 alerts for two rules, got %d", len(fired))
	}

	active := ae.ActiveAlerts()
	if len(active) != 2 {
		t.Errorf("expected 2 active alerts, got %d", len(active))
	}
}

func TestDefaultRulesAdded(t *testing.T) {
	dp := NewDataPipeline(CollectionConfig{Capacity: 10})
	ae := NewAlertEngine(dp)
	ae.AddDefaultRules()

	// Push values that trigger all default rules
	dp.PushMetric(MetricCPU, "%", 95)
	dp.PushMetric(MetricMem, "%", 95)
	dp.PushMetric(MetricDisk, "%", 96)

	fired := ae.Evaluate()

	// With flapCount 2 for critical rules, 3 for warning rules,
	// no alerts should fire on the first violation
	if len(fired) != 0 {
		t.Errorf("expected 0 alerts on first violation (flap), got %d", len(fired))
	}

	// Second violation
	dp.PushMetric(MetricCPU, "%", 95)
	dp.PushMetric(MetricMem, "%", 95)
	dp.PushMetric(MetricDisk, "%", 96)
	fired = ae.Evaluate()

	// Critical rules have flapCount=2, so they fire now
	// CPU critical (95 > 90), MEM critical (95 > 90), DISK critical (96 > 95)
	// Warning rules still need one more
	if len(fired) != 3 {
		t.Errorf("expected 3 alerts (critical), got %d", len(fired))
	}
}

func TestAlertLevelStrings(t *testing.T) {
	if AlertInfo.String() != "INFO" {
		t.Errorf("expected INFO, got %s", AlertInfo.String())
	}
	if AlertWarning.String() != "WARNING" {
		t.Errorf("expected WARNING, got %s", AlertWarning.String())
	}
	if AlertCritical.String() != "CRITICAL" {
		t.Errorf("expected CRITICAL, got %s", AlertCritical.String())
	}
}

func TestAllAlerts(t *testing.T) {
	dp := NewDataPipeline(CollectionConfig{Capacity: 10})
	ae := NewAlertEngine(dp)

	ae.AddRule(AlertRule{
		Metric:    MetricCPU,
		Condition: AlertGT,
		Threshold: 80,
		FlapCount: 1,
		Severity:  AlertWarning,
	})

	dp.PushMetric(MetricCPU, "%", 90)
	ae.Evaluate()

	// Recover
	dp.PushMetric(MetricCPU, "%", 50)
	ae.Evaluate()

	all := ae.AllAlerts()
	if len(all) != 1 {
		t.Errorf("expected 1 total alert, got %d", len(all))
	}
	if !all[0].Resolved {
		t.Error("expected alert to be resolved")
	}
}
