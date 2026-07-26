package devops

import (
	"testing"
)

func TestRunDevOpsDiagnostics_ReturnsResult(t *testing.T) {
	result := RunDevOpsDiagnostics()
	if len(result.Checks) == 0 {
		t.Error("RunDevOpsDiagnostics returned no checks")
	}
	if result.Timestamp == "" {
		t.Error("RunDevOpsDiagnostics should include a timestamp")
	}
	t.Logf("DevOps diagnostics: %d checks, score=%d", len(result.Checks), result.Score)
	for _, c := range result.Checks {
		t.Logf("  %s: %s (%s)", c.Name, c.Status, c.Value)
	}
}

func TestRunDevOpsDiagnostics_AllChecksPresent(t *testing.T) {
	result := RunDevOpsDiagnostics()
	// The exact names vary by environment (docker could be "Docker CLI" or "Docker Daemon")
	// We check that we got the expected number of checks and that certain key names are present
	if len(result.Checks) < 4 {
		t.Errorf("RunDevOpsDiagnostics returned %d checks, want at least 4", len(result.Checks))
	}
	// Key checks that should always be present regardless of environment
	foundGit := false
	foundGo := false
	foundNode := false
	for _, c := range result.Checks {
		if c.Name == "Git" || c.Name == "Git Status" {
			foundGit = true
		}
		if c.Name == "Go" {
			foundGo = true
		}
		if c.Name == "Node.js" {
			foundNode = true
		}
	}
	if !foundGit {
		t.Error("RunDevOpsDiagnostics missing Git check")
	}
	if !foundGo {
		t.Error("RunDevOpsDiagnostics missing Go check")
	}
	if !foundNode {
		t.Error("RunDevOpsDiagnostics missing Node.js check")
	}
}

func TestRunDevOpsDiagnostics_ValidStatuses(t *testing.T) {
	result := RunDevOpsDiagnostics()
	validStatuses := map[string]bool{"pass": true, "fail": true, "warn": true}
	for _, c := range result.Checks {
		if !validStatuses[c.Status] {
			t.Errorf("Check %q has invalid status: %q", c.Name, c.Status)
		}
		if c.Name == "" {
			t.Error("Check has empty name")
		}
	}
}
