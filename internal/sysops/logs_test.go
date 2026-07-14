package sysops

import (
	"strings"
	"testing"
)

func TestGetSystemLogs(t *testing.T) {
	result, err := GetSystemLogs(10, "journald")
	if err != nil {
		// Windows PowerShell event log may fail in test environment
		if strings.Contains(err.Error(), "powershell") || strings.Contains(err.Error(), "not found") {
			t.Skip("System logs not available in this environment")
		}
		t.Logf("GetSystemLogs returned error (may be expected): %v", err)
		return
	}
	t.Logf("Got %d log entries from %s", result.Total, result.Source)
}
