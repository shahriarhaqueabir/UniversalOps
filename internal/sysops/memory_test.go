package sysops

import (
	"testing"
)

func TestGetMemoryStats(t *testing.T) {
	stats, err := GetMemoryStats()
	if err != nil {
		t.Fatalf("GetMemoryStats failed: %v", err)
	}

	if stats.TotalBytes == 0 {
		t.Error("Total memory reported as 0")
	}

	if stats.UsedPercent < 0 || stats.UsedPercent > 100 {
		t.Errorf("Invalid used percent: %f", stats.UsedPercent)
	}

	if stats.UsedBytes > stats.TotalBytes {
		t.Errorf("Used bytes (%d) exceeds total bytes (%d)", stats.UsedBytes, stats.TotalBytes)
	}
}
