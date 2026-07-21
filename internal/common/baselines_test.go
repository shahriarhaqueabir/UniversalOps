package common

import (
	"testing"
)

func TestBaselinesEngine_Calculate(t *testing.T) {
	p := NewDataPipeline(DefaultCollectionConfig())
	be := NewBaselinesEngine(p)

	// Simulate stable history: [10, 11, 10, 12, 10, 11...]
	history := []float64{10, 11, 10, 12, 10, 11, 10, 11, 10, 12}

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

	// 2. Test "High Severity Drift" (> 2.5σ)
	// Threshold = 10.5 + (2.5 * 1.0) = 13.0
	drift, ok := be.checkDrift(entry, 15.0)
	if !ok {
		t.Fatal("Expected drift detection for value 15.0")
	}
	if drift.Severity != "high" {
		t.Errorf("Expected severity 'high', got %s", drift.Severity)
	}

	// 3. Test "Medium Severity Drift" (> 1.5σ)
	// Threshold = 10.5 + (1.5 * 1.0) = 12.0
	drift, ok = be.checkDrift(entry, 12.5)
	if !ok {
		t.Fatal("Expected drift detection for value 12.5")
	}
	if drift.Severity != "med" {
		t.Errorf("Expected severity 'med', got %s", drift.Severity)
	}
}
