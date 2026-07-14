package netops

import "testing"

func TestRunNetworkHealthCheck(t *testing.T) {
	report := RunNetworkHealthCheck()
	if report.Score < 0 || report.Score > 100 {
		t.Errorf("Score should be 0-100, got %d", report.Score)
	}
	if len(report.Checks) == 0 {
		t.Error("Expected at least one health check")
	}
	t.Logf("Health Score: %d, Summary: %s, Duration: %s", report.Score, report.Summary, report.Duration)
	for _, c := range report.Checks {
		t.Logf("  [%s] %s: %s", c.Status, c.Name, c.Detail)
	}
}
