package common

import (
	"testing"
)

func TestBaselinesEngine_Calculate(t *testing.T) {
	p := NewDataPipeline(DefaultCollectionConfig())
	be := NewBaselinesEngine(p)

	// We manually set a baseline since we can't easily wait for the background ticker
	// in a unit test without mocking the storage.
	// For now, we test the drift detection logic.

	entry := BaselineEntry{
		Metric: "test.metric",
		Avg:    10.5,
		StdDev: 1.0,
	}

	// 1. Test "No Drift"
	if _, ok := be.checkDrift(entry, 11.5); ok {
		t.Error("Expected no drift for value 11.5 (within 2.5σ)")
	}

	// 2. Test "High Severity Drift" (> 5σ)
	// Threshold = 10.5 + (5.0 * 1.0) = 15.5
	drift, ok := be.checkDrift(entry, 16.0)
	if !ok {
		t.Fatal("Expected drift detection for value 16.0")
	}
	if drift.Severity != "high" {
		t.Errorf("Expected severity 'high', got %s", drift.Severity)
	}

	// 3. Test "Medium Severity Drift" (> 3.5σ)
	// Threshold = 10.5 + (3.5 * 1.0) = 14.0
	drift, ok = be.checkDrift(entry, 14.5)
	if !ok {
		t.Fatal("Expected drift detection for value 14.5")
	}
	if drift.Severity != "med" {
		t.Errorf("Expected severity 'med', got %s", drift.Severity)
	}
}
