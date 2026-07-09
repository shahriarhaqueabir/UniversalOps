package sysops

import (
	"testing"
)

func TestGetSystemInfo(t *testing.T) {
	info, err := GetSystemInfo()
	if err != nil {
		t.Fatalf("GetSystemInfo failed: %v", err)
	}

	if info.Hostname == "" {
		t.Error("Hostname is empty")
	}

	if info.Platform == "" {
		t.Error("Platform is empty")
	}

	if info.UptimeSeconds == 0 {
		t.Log("Uptime reported as 0 (possible on some systems or during early boot)")
	}
}
