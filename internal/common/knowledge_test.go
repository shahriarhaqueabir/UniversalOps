package common

import (
	"testing"
)

func TestKnowledgeManager(t *testing.T) {
	InitKnowledge()
	km := GetKnowledge()

	// Test Initial State
	s := km.GetSnapshot()
	if s.CPUUsage != 0 {
		t.Errorf("expected 0 CPU, got %f", s.CPUUsage)
	}

	// Test Update
	km.Update(func(sk *SystemKnowledge) {
		sk.CPUUsage = 45.5
		sk.MemoryUsage = 70.2
		sk.SecurityGrade = "A"
	})

	s2 := km.GetSnapshot()
	if s2.CPUUsage != 45.5 {
		t.Errorf("expected 45.5 CPU, got %f", s2.CPUUsage)
	}
	if s2.SecurityGrade != "A" {
		t.Errorf("expected Grade A, got %s", s2.SecurityGrade)
	}
}
