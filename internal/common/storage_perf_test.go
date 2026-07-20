package common

import (
	"testing"
	"time"
)

func TestGenerateSQLiteTelemetry(t *testing.T) {
	// Initialize in-memory but with instrumentation active
	err := InitStorage(":memory:")
	if err != nil {
		t.Fatalf("InitStorage failed: %v", err)
	}
	s := GetStorage()
	defer s.Close()

	// Simulate collection for 3 seconds (enough to exercise the path)
	start := time.Now()
	for time.Since(start) < 3*time.Second {
		s.InsertMetric("cpu.percent", "%", 42.0)
		s.InsertMetric("memory.percent", "%", 55.0)
		time.Sleep(100 * time.Millisecond)
	}

	// Flush to ensure everything is counted
	s.flushMetrics()
}
