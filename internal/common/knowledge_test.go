package common

import (
	"testing"
)

func TestKnowledgeManager(t *testing.T) {
	dp := NewDataPipeline(CollectionConfig{})
	InitKnowledge(dp)
	km := GetKnowledge()

	// Test Initial State
	s := km.GetSnapshot()
	if s.SystemCPUUtilization != 0 {
		t.Errorf("expected 0 CPU, got %f", s.SystemCPUUtilization)
	}

	// Test Update
	km.UpdateSystemState("A", 2, 10, "1h30m", []string{"eth0"})

	s2 := km.GetSnapshot()
	if s2.SecurityGrade != "A" {
		t.Errorf("expected Grade A, got %s", s2.SecurityGrade)
	}
}
