package common

import (
	"testing"
)

func TestAlertEngine_MemoryGrowth(t *testing.T) {
	dp := NewDataPipeline(CollectionConfig{Capacity: 10})
	ae := NewAlertEngine(dp)

	ae.AddRule(AlertRule{
		Metric:    "test.metric",
		Condition: AlertGT,
		Threshold: 50,
		FlapCount: 1,
		Severity:  AlertWarning,
	})

	// Fire and resolve many alerts
	for i := 0; i < 100; i++ {
		dp.PushMetric("test.metric", "unit", 60)
		ae.Evaluate()

		dp.PushMetric("test.metric", "unit", 40)
		ae.Evaluate()
	}

	all := ae.AllAlerts()
	if len(all) != 100 {
		t.Errorf("Expected 100 historical alerts, got %d", len(all))
	}

	// This shows that alerts grow indefinitely in memory.
	// In a real production app, we'd want a way to prune these.
}

func TestAlertEngine_DuplicateActiveAlerts(t *testing.T) {
	dp := NewDataPipeline(CollectionConfig{Capacity: 10})
	ae := NewAlertEngine(dp)

	ae.AddRule(AlertRule{
		Metric:    "test.metric",
		Condition: AlertGT,
		Threshold: 50,
		FlapCount: 1,
		Severity:  AlertWarning,
	})

	// Push violation
	dp.PushMetric("test.metric", "unit", 60)
	ae.Evaluate()

	// Push another violation - should not create a new alert
	dp.PushMetric("test.metric", "unit", 70)
	ae.Evaluate()

	if ae.AlertCount() != 1 {
		t.Errorf("Expected 1 active alert, got %d", ae.AlertCount())
	}

	alerts := ae.ActiveAlerts()
	if len(alerts) > 0 && alerts[0].Value != 70 {
		t.Errorf("Expected updated value 70, got %f", alerts[0].Value)
	}
}

func TestAlertEngine_MultipleSeverities(t *testing.T) {
	dp := NewDataPipeline(CollectionConfig{Capacity: 10})
	ae := NewAlertEngine(dp)

	// Warning at 70, Critical at 90
	ae.AddRule(AlertRule{
		Metric:    "cpu",
		Condition: AlertGT,
		Threshold: 70,
		FlapCount: 1,
		Severity:  AlertWarning,
	})
	ae.AddRule(AlertRule{
		Metric:    "cpu",
		Condition: AlertGT,
		Threshold: 90,
		FlapCount: 1,
		Severity:  AlertCritical,
	})

	dp.PushMetric("cpu", "%", 95)
	fired := ae.Evaluate()

	if len(fired) != 2 {
		t.Errorf("Expected 2 alerts fired (Warning and Critical), got %d", len(fired))
	}

	if ae.AlertCount() != 2 {
		t.Errorf("Expected 2 active alerts, got %d", ae.AlertCount())
	}
}

func TestAlertEngine_TemplateMessage(t *testing.T) {
	dp := NewDataPipeline(CollectionConfig{Capacity: 10})
	ae := NewAlertEngine(dp)

	ae.AddRule(AlertRule{
		Metric:    "disk",
		Condition: AlertGT,
		Threshold: 90,
		FlapCount: 1,
		Severity:  AlertCritical,
		Message:   "Disk space critical: {value}% > {threshold}%",
	})

	dp.PushMetric("disk", "%", 95)
	fired := ae.Evaluate()

	if len(fired) > 0 {
		msg := fired[0].Message
		expected := "Disk space critical: 95.0% > 90.0%"
		if msg != expected {
			t.Errorf("Expected message %q, got %q", expected, msg)
		}
	}
}
