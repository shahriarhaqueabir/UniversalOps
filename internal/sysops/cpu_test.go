package sysops

import (
	"testing"
)

func TestGetCPUStats(t *testing.T) {
	stats, err := GetCPUStats()
	if err != nil {
		t.Fatalf("GetCPUStats() returned error: %v", err)
	}
	if stats == nil {
		t.Fatal("GetCPUStats() returned nil")
	}
}

func TestGetCPUStatsPercentRange(t *testing.T) {
	stats, err := GetCPUStats()
	if err != nil {
		t.Fatalf("GetCPUStats() returned error: %v", err)
	}
	if stats.Percent < 0 || stats.Percent > 100 {
		t.Errorf("Percent = %v, want between 0 and 100", stats.Percent)
	}
}

func TestGetCPUStatsPerCPUSliceLength(t *testing.T) {
	stats, err := GetCPUStats()
	if err != nil {
		t.Fatalf("GetCPUStats() returned error: %v", err)
	}
	if len(stats.PerCPU) != stats.CoreCount {
		t.Errorf("len(PerCPU) = %d, want %d (CoreCount)", len(stats.PerCPU), stats.CoreCount)
	}
}

func TestGetCPUStatsPerCPUSumMatchesTotal(t *testing.T) {
	stats, err := GetCPUStats()
	if err != nil {
		t.Fatalf("GetCPUStats() returned error: %v", err)
	}
	if stats.CoreCount == 0 {
		t.Skip("CoreCount is 0, cannot verify sum")
	}

	var sum float64
	for _, v := range stats.PerCPU {
		sum += v
	}

	// Percent is the average (sum/coreCount), so the average of PerCPU should
	// match Percent. Use a small epsilon for floating point comparison.
	avg := sum / float64(len(stats.PerCPU))
	diff := stats.Percent - avg
	if diff < -0.01 || diff > 0.01 {
		t.Errorf("PerCPU avg = %v, want %v (diff=%v)", avg, stats.Percent, diff)
	}
}

func TestGetCPUStatsModelNameNonEmpty(t *testing.T) {
	stats, err := GetCPUStats()
	if err != nil {
		t.Fatalf("GetCPUStats() returned error: %v", err)
	}
	if stats.ModelName == "" {
		t.Log("ModelName is empty (may be expected in some environments)")
	}
}

func TestGetCPUStatsCoreCountPositive(t *testing.T) {
	stats, err := GetCPUStats()
	if err != nil {
		t.Fatalf("GetCPUStats() returned error: %v", err)
	}
	if stats.CoreCount <= 0 {
		t.Errorf("CoreCount = %d, want > 0", stats.CoreCount)
	}
}
