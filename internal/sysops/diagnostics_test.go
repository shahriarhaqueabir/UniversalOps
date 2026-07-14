package sysops

import "testing"

func TestRunExtendedDiagnostics(t *testing.T) {
	result, err := RunExtendedDiagnostics()
	if err != nil {
		t.Fatalf("RunExtendedDiagnostics returned error: %v", err)
	}
	if len(result.Checks) == 0 {
		t.Error("Expected at least one diagnostic check")
	}
	if result.Score < 0 || result.Score > 100 {
		t.Errorf("Score should be 0-100, got %d", result.Score)
	}
	t.Logf("Diagnostic score: %d, checks: %d, timestamp: %s", result.Score, len(result.Checks), result.Timestamp)
	for _, c := range result.Checks {
		t.Logf("  [%s] %s = %s — %s", c.Status, c.Name, c.Value, c.Message)
	}
}
