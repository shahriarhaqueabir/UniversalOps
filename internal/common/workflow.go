package common

import (
	"context"
	"sync"
)

type WorkflowStepType string

const (
	StepTypeQuery      WorkflowStepType = "query"
	StepTypePowerShell WorkflowStepType = "powershell"
	StepTypeAction     WorkflowStepType = "action"
)

// WorkflowStep represents a single unit of work in a multi-step operation.
type WorkflowStep struct {
	ID          string           `json:"id"`
	Type        WorkflowStepType `json:"type"`
	Label       string           `json:"label"`
	Description string           `json:"description"`

	// Command to be executed (for Dry Run preview)
	Command string `json:"command"`

	// ExpectedOutcome describes what the step intends to achieve
	ExpectedOutcome string `json:"expected_outcome"`

	// Action is the actual Go logic to execute
	Action func(ctx context.Context) (any, error) `json:"-"`

	// Results from the execution
	Result any    `json:"result,omitempty"`
	Error  string `json:"error,omitempty"`
}


// WorkflowDefinition defines a reusable operational sequence.
type WorkflowDefinition struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Why         string         `json:"why"`
	Risks       []string       `json:"risks"`
	TypicalVals string         `json:"typical_values"`
	Steps       []WorkflowStep `json:"steps"`
}

// WorkflowEngine manages the library of reusable operational workflows.
type WorkflowEngine struct {
	mu        sync.RWMutex
	library   map[string]WorkflowDefinition
}

func NewWorkflowEngine() *WorkflowEngine {
	return &WorkflowEngine{
		library: make(map[string]WorkflowDefinition),
	}
}

func (e *WorkflowEngine) Register(wf WorkflowDefinition) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.library[wf.ID] = wf
}

func (e *WorkflowEngine) Get(id string) (WorkflowDefinition, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	wf, ok := e.library[id]
	return wf, ok
}

func (e *WorkflowEngine) List() []WorkflowDefinition {
	e.mu.RLock()
	defer e.mu.RUnlock()
	var list []WorkflowDefinition
	for _, wf := range e.library {
		list = append(list, wf)
	}
	return list
}
