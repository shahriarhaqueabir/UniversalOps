package secops

import (
	"testing"
)

func TestAuditCheckItemStruct(t *testing.T) {
	item := AuditCheckItem{
		Category:    "Firewall",
		Check:       "Firewall enabled",
		Passed:      true,
		Description: "Test description",
		Remediation: "Test remediation",
	}
	if !item.Passed {
		t.Error("expected Passed to be true")
	}
}

func TestSecurityAuditResultStruct(t *testing.T) {
	result := SecurityAuditResult{
		Score:     85,
		Total:     10,
		Passed:    8,
		Failed:    2,
		Items:     []AuditCheckItem{},
		Timestamp: "2026-01-01T00:00:00Z",
	}
	if result.Score != 85 {
		t.Errorf("expected score 85, got %d", result.Score)
	}
}
