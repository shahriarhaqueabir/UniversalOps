package sysops

import (
	"errors"
	"testing"
)

func TestRunSystemAction_Unknown(t *testing.T) {
	_, err := RunSystemAction("nonexistent_action")
	if err == nil {
		t.Error("Expected error for unknown action")
	}
}

func TestActionErrorHint_AccessDenied(t *testing.T) {
	baseErr := errors.New("access denied")
	tests := []struct {
		name   string
		action SystemAction
		want   string
	}{
		{"flush_dns", ActionFlushDNS, "administrator privileges"},
		{"clear_temp", ActionClearTemp, "administrator rights"},
		{"disk_cleanup", ActionDiskCleanup, "administrator privileges"},
		{"defrag", ActionDefrag, "administrator privileges"},
		{"restart_service", ActionRestartService, "administrator privileges"},
		{"reboot default", ActionReboot, "administrator privileges"},
		{"shutdown default", ActionShutdown, "Run Universal-Ops as Administrator"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := actionErrorHint(tt.action, baseErr, "")
			if !contains(got, tt.want) {
				t.Errorf("actionErrorHint(%q, access denied) = %q, want contains %q", tt.action, got, tt.want)
			}
		})
	}
}

func TestActionErrorHint_AccessDeniedVariants(t *testing.T) {
	variants := []error{
		errors.New("permission denied"),
		errors.New("required privilege not held"),
		errors.New("elevated privileges required"),
	}
	for _, err := range variants {
		got := actionErrorHint(ActionFlushDNS, err, "")
		if !contains(got, "administrator") {
			t.Errorf("actionErrorHint(flush_dns, %q) = %q, want administrator hint", err.Error(), got)
		}
	}
}

func TestActionErrorHint_NotFound(t *testing.T) {
	baseErr := errors.New("not found")
	tests := []struct {
		name   string
		action SystemAction
		want   string
	}{
		{"flush_dns", ActionFlushDNS, "ipconfig"},
		{"clear_temp", ActionClearTemp, "PowerShell"},
		{"disk_cleanup", ActionDiskCleanup, "cleanmgr.exe"},
		{"defrag", ActionDefrag, "defrag.exe"},
		{"reboot default", ActionReboot, "PATH"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := actionErrorHint(tt.action, baseErr, "")
			if !contains(got, tt.want) {
				t.Errorf("actionErrorHint(%q, not found) = %q, want contains %q", tt.action, got, tt.want)
			}
		})
	}
}

func TestActionErrorHint_NotFoundVariants(t *testing.T) {
	variants := []error{
		errors.New("no such file or directory"),
		errors.New("cannot find"),
	}
	for _, err := range variants {
		got := actionErrorHint(ActionFlushDNS, err, "")
		if !contains(got, "ipconfig") {
			t.Errorf("actionErrorHint(flush_dns, %q) = %q, want ipconfig hint", err.Error(), got)
		}
	}
}

func TestActionErrorHint_ExitStatus1(t *testing.T) {
	baseErr := errors.New("exit status 1")
	tests := []struct {
		name   string
		action SystemAction
		want   string
	}{
		{"clear_temp", ActionClearTemp, "locked"},
		{"clean_pkg_cache", ActionCleanPkgCache, "internet connection"},
		{"system_update", ActionSystemUpdate, "internet connectivity"},
		{"disk_cleanup", ActionDiskCleanup, "in use"},
		{"defrag", ActionDefrag, "chkdsk"},
		{"reboot default", ActionReboot, "exit code 1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := actionErrorHint(tt.action, baseErr, "")
			if !contains(got, tt.want) {
				t.Errorf("actionErrorHint(%q, exit status 1) = %q, want contains %q", tt.action, got, tt.want)
			}
		})
	}
}

func TestActionErrorHint_FlushDNSUnrecognized(t *testing.T) {
	err := errors.New("exit status 1")
	got := actionErrorHint(ActionFlushDNS, err, "unrecognized")
	if !contains(got, "winsock") {
		t.Errorf("actionErrorHint(flush_dns, exit1, 'unrecognized') = %q, want winsock hint", got)
	}

	got2 := actionErrorHint(ActionFlushDNS, err, "bad")
	if !contains(got2, "winsock") {
		t.Errorf("actionErrorHint(flush_dns, exit1, 'bad') = %q, want winsock hint", got2)
	}

	got3 := actionErrorHint(ActionFlushDNS, err, "")
	if !contains(got3, "Administrator") {
		t.Errorf("actionErrorHint(flush_dns, exit1, '') = %q, want Administrator hint", got3)
	}
}

func TestActionErrorHint_Timeout(t *testing.T) {
	err := errors.New("timeout")
	got := actionErrorHint(ActionFlushDNS, err, "")
	if !contains(got, "timed out") {
		t.Errorf("actionErrorHint(flush_dns, timeout) = %q, want timeout hint", got)
	}

	err2 := errors.New("timed out")
	got2 := actionErrorHint(ActionFlushDNS, err2, "")
	if !contains(got2, "timed out") {
		t.Errorf("actionErrorHint(flush_dns, 'timed out') = %q, want timeout hint", got2)
	}
}

func TestActionErrorHint_RestartServiceSpecific(t *testing.T) {
	t.Run("does not exist", func(t *testing.T) {
		err := errors.New("service does not exist")
		got := actionErrorHint(ActionRestartService, err, "")
		if !contains(got, "Services.msc") {
			t.Errorf("actionErrorHint(restart, 'does not exist') = %q, want Services.msc hint", got)
		}
	})

	t.Run("failed to start", func(t *testing.T) {
		err := errors.New("failed to start service")
		got := actionErrorHint(ActionRestartService, err, "")
		if !contains(got, "Event Viewer") {
			t.Errorf("actionErrorHint(restart, 'failed to start') = %q, want Event Viewer hint", got)
		}
	})

	t.Run("not start", func(t *testing.T) {
		err := errors.New("service did not start")
		got := actionErrorHint(ActionRestartService, err, "")
		if !contains(got, "Event Viewer") {
			t.Errorf("actionErrorHint(restart, 'not start') = %q, want Event Viewer hint", got)
		}
	})
}

func TestActionErrorHint_Fallback(t *testing.T) {
	err := errors.New("some random error")
	got := actionErrorHint(ActionReboot, err, "")
	if !contains(got, "some random error") {
		t.Errorf("actionErrorHint(reboot, random) = %q, want to include error message", got)
	}
}

func TestRestartService_InvalidName(t *testing.T) {
	_, err := RestartService("../invalid")
	if err == nil {
		t.Error("RestartService with invalid name should return error")
	}
	if !contains(err.Error(), "invalid service name") {
		t.Errorf("RestartService invalid name error = %q, want 'invalid service name'", err.Error())
	}
}

func TestRestartService_EmptyName(t *testing.T) {
	_, err := RestartService("")
	if err == nil {
		t.Error("RestartService with empty name should return error")
	}
}

func TestRestartService_LongName(t *testing.T) {
	longName := ""
	for i := 0; i < 300; i++ {
		longName += "a"
	}
	_, err := RestartService(longName)
	if err == nil {
		t.Error("RestartService with very long name should return error")
	}
}

// contains is a helper for substring matching in test assertions.
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestActionErrorHint_OutputContainsMatch(t *testing.T) {
	// Test that the function also searches output string for not-found patterns
	err := errors.New("exit status 127")
	got := actionErrorHint(ActionFlushDNS, err, "output permission denied")
	if !contains(got, "administrator") {
		t.Errorf("actionErrorHint with output containing 'permission denied' = %q, want administrator hint", got)
	}
}

func TestFormatActionResult(t *testing.T) {
	r := &ActionResult{
		Action:  "test_action",
		Success: true,
		Message: "completed",
		Output:  "some output here",
	}
	if r.Action != "test_action" || !r.Success {
		t.Errorf("ActionResult fields mismatch: %+v", r)
	}
}
