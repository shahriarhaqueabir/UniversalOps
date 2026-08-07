package app

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/shahriarhaqueabir/UniversalOps/internal/common"
)

// fakeResult is a deliberately arbitrary struct to prove the envelope wraps
// native (non-shell) step payloads.
type fakeResult struct {
	Name  string
	Count int
}

func TestExecuteWorkflow_WrapsNativeResult(t *testing.T) {
	engine := common.NewWorkflowEngine()
	engine.Register(common.WorkflowDefinition{
		ID:          "test-native",
		Name:        "Native Test",
		Description: "step with a Go action",
		Category:    common.WorkflowCategorySystem,
		Steps: []common.WorkflowStep{
			{
				ID:    "step-1",
				Type:  common.StepTypeAction,
				Label: "Probe",
				Action: func(ctx context.Context) (any, error) {
					time.Sleep(5 * time.Millisecond)
					return fakeResult{Name: "probe", Count: 7}, nil
				},
			},
		},
	})

	api := &WorkflowAPI{workflowEngine: engine}
	wf, err := api.ExecuteWorkflow("test-native")
	if err != nil {
		t.Fatalf("ExecuteWorkflow: %v", err)
	}

	step := wf.Steps[0]
	env, ok := step.Result.(*common.StepResult)
	if !ok {
		t.Fatalf("result type = %T, want *common.StepResult", step.Result)
	}
	if env.Status != common.StepStatusSuccess {
		t.Fatalf("status: got %q", env.Status)
	}
	if env.DurationNS <= 0 {
		t.Fatalf("duration_ns: got %d, want > 0", env.DurationNS)
	}
	if !strings.Contains(env.Summary, "snapshot") {
		t.Fatalf("summary should describe native struct, got %q", env.Summary)
	}
	if _, err := time.Parse(time.RFC3339, env.Timestamp); err != nil {
		t.Fatalf("timestamp %q invalid: %v", env.Timestamp, err)
	}
	raw, ok := env.Data.(fakeResult)
	if !ok || raw.Count != 7 {
		t.Fatalf("data passthrough: got %#v", env.Data)
	}
	if step.Error != "" {
		t.Fatalf("step.Error should be empty, got %q", step.Error)
	}
}

func TestExecuteWorkflow_WrapsSliceWithItemCount(t *testing.T) {
	engine := common.NewWorkflowEngine()
	engine.Register(common.WorkflowDefinition{
		ID:   "test-slice",
		Name: "Slice Test",
		Steps: []common.WorkflowStep{
			{
				ID:    "step-1",
				Type:  common.StepTypeAction,
				Label: "List",
				Action: func(ctx context.Context) (any, error) {
					return []string{"a", "b", "c"}, nil
				},
			},
		},
	})

	api := &WorkflowAPI{workflowEngine: engine}
	wf, err := api.ExecuteWorkflow("test-slice")
	if err != nil {
		t.Fatalf("ExecuteWorkflow: %v", err)
	}

	env := wf.Steps[0].Result.(*common.StepResult)
	if env.Items != 3 {
		t.Fatalf("items: got %d, want 3", env.Items)
	}
	if env.Summary != "3 strings" {
		t.Fatalf("summary: got %q, want \"3 strings\"", env.Summary)
	}
}

func TestExecuteWorkflow_ErrorStepAbortsAndWraps(t *testing.T) {
	engine := common.NewWorkflowEngine()
	engine.Register(common.WorkflowDefinition{
		ID:   "test-err",
		Name: "Error Test",
		Steps: []common.WorkflowStep{
			{
				ID:    "step-1",
				Type:  common.StepTypeAction,
				Label: "Failing",
				Action: func(ctx context.Context) (any, error) {
					return nil, errors.New("kaboom")
				},
			},
			{
				ID:    "step-2",
				Type:  common.StepTypeAction,
				Label: "Never Runs",
				Action: func(ctx context.Context) (any, error) {
					return "ran", nil
				},
			},
		},
	})

	api := &WorkflowAPI{workflowEngine: engine}
	wf, err := api.ExecuteWorkflow("test-err")
	if err == nil {
		t.Fatal("expected abort error")
	}
	if !strings.Contains(err.Error(), "workflow aborted at step \"step-1\"") {
		t.Fatalf("abort error: %v", err)
	}

	step := wf.Steps[0]
	env, ok := step.Result.(*common.StepResult)
	if !ok {
		t.Fatalf("result type = %T, want *common.StepResult", step.Result)
	}
	if env.Status != common.StepStatusError {
		t.Fatalf("status: got %q", env.Status)
	}
	if env.Error != "kaboom" {
		t.Fatalf("envelope error: got %q", env.Error)
	}
	if step.Error != "kaboom" {
		t.Fatalf("step.Error: got %q", step.Error)
	}

	// The step after the failure must never have executed.
	if wf.Steps[1].Result != nil {
		t.Fatalf("step-2 should not have run, got result %#v", wf.Steps[1].Result)
	}
}
