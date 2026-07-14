package sysops

import "testing"

func TestGetCPUExtended(t *testing.T) {
	stats, err := GetCPUExtended()
	if err != nil {
		t.Fatalf("GetCPUExtended returned error: %v", err)
	}
	if stats.ModelName == "" {
		t.Error("ModelName is empty")
	}
	if stats.FrequencyMHz <= 0 {
		t.Errorf("FrequencyMHz should be > 0, got %f", stats.FrequencyMHz)
	}
	if len(stats.PerCPUInfo) == 0 {
		t.Error("PerCPUInfo is empty")
	}
}
