package sysops

import "testing"

func TestGetPerformanceStats(t *testing.T) {
	stats, err := GetPerformanceStats()
	if err != nil {
		t.Fatalf("GetPerformanceStats returned error: %v", err)
	}
	if stats.CPUTimes.Total <= 0 {
		t.Error("CPUTimes.Total should be > 0")
	}
}
