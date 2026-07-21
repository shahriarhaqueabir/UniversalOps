package common

import (
	"sync"
	"testing"
	"time"
)

type MockWorkflowInvoker struct {
	mu           sync.Mutex
	TriggeredIDs []string
}

func (m *MockWorkflowInvoker) TriggerWorkflow(id string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.TriggeredIDs = append(m.TriggeredIDs, id)
	return "mock-report-id", nil
}

func TestEngineLoop_AutonomousTrigger(t *testing.T) {
	p := NewDataPipeline(DefaultCollectionConfig())
	eb := NewEventBus(10)
	invoker := &MockWorkflowInvoker{}

	e := NewEngineLoop(p, NewAlertEngine(p), eb, nil, invoker)

	// Simulate a CPU spike: Jump from 10% to 50%
	// The threshold in engine.go is prevCPU + 15

	// Initial state
	s1 := MetricSnapshot{CPU: 10.0, Timestamp: time.Now()}
	e.DetectSpikes(s1)

	// Spike state
	s2 := MetricSnapshot{CPU: 50.0, Timestamp: time.Now()}
	e.DetectSpikes(s2)

	// Autonomous trigger is a goroutine, wait a moment
	time.Sleep(50 * time.Millisecond)

	invoker.mu.Lock()
	defer invoker.mu.Unlock()

	if len(invoker.TriggeredIDs) == 0 {
		t.Fatal("Expected 'diag-slow-pc' workflow to be triggered automatically")
	}

	if invoker.TriggeredIDs[0] != "diag-slow-pc" {
		t.Errorf("Expected triggered workflow 'diag-slow-pc', got %s", invoker.TriggeredIDs[0])
	}
}
