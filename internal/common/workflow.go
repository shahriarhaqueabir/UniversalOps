package common

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"
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

// Step status values used by the standardized StepResult envelope.
const (
	StepStatusSuccess = "success"
	StepStatusError   = "error"
)

// StepResult is the standardized execution envelope attached to every workflow
// step after it runs. All step payloads are wrapped in this shape so consumers
// can render status, duration, summary, item count, and raw data uniformly,
// regardless of the underlying Go type that produced the payload.
type StepResult struct {
	Status     string `json:"status"`          // "success" | "error"
	Summary    string `json:"summary"`         // one-line human-readable interpretation
	Items      int    `json:"items,omitempty"` // top-level element count (rows / entries / lines)
	DurationNS int64  `json:"duration_ns"`     // wall-clock execution time in nanoseconds
	Timestamp  string `json:"timestamp"`       // RFC3339 completion timestamp (UTC)
	Data       any    `json:"data,omitempty"`  // raw step payload
	Error      string `json:"error,omitempty"` // error message when Status == "error"
}

// NewSuccessStepResult wraps a successful step payload in the standard envelope.
func NewSuccessStepResult(data any, started time.Time) *StepResult {
	summary, items := SummarizeResult(data)
	return &StepResult{
		Status:     StepStatusSuccess,
		Summary:    summary,
		Items:      items,
		DurationNS: time.Since(started).Nanoseconds(),
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
		Data:       data,
	}
}

// NewErrorStepResult wraps a step failure in the standard envelope. The raw
// payload is intentionally not attached; the error message carries the detail.
func NewErrorStepResult(err error, started time.Time) *StepResult {
	msg := ""
	if err != nil {
		msg = err.Error()
	}
	return &StepResult{
		Status:     StepStatusError,
		Summary:    "Step failed",
		DurationNS: time.Since(started).Nanoseconds(),
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
		Error:      msg,
	}
}

// SummarizeResult produces a one-line human-readable summary and a top-level
// element count for an arbitrary step result. It understands shell outputs,
// slices, arrays, maps, structs, and scalars; unknown shapes fall back to a
// type-name snapshot label. Reflection-based so it works with any payload the
// workflow engine produces, including results from custom workflows.
func SummarizeResult(res any) (summary string, items int) {
	if res == nil {
		return "No data returned", 0
	}

	v := reflect.ValueOf(res)
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		if v.IsNil() {
			return "No data returned", 0
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.String:
		s := v.String()
		if s == "" {
			return "Empty string", 0
		}
		if len(s) > 140 {
			s = s[:140] + "…"
		}
		return s, 0

	case reflect.Slice, reflect.Array:
		n := v.Len()
		if n == 0 {
			return "No records", 0
		}
		return fmt.Sprintf("%d %s", n, pluralizeNoun(humanizeName(v.Type().Elem().Name()))), n

	case reflect.Map:
		n := v.Len()
		if n == 0 {
			return "No entries", 0
		}
		return fmt.Sprintf("%d entries", n), n

	case reflect.Struct:
		if shellSummary, lines, ok := summarizeShellLike(v); ok {
			return shellSummary, lines
		}
		name := humanizeName(v.Type().Name())
		if name == "" {
			name = "structured"
		}
		return fmt.Sprintf("%s snapshot (%d fields)", name, v.NumField()), 0

	default:
		return fmt.Sprintf("%v", v.Interface()), 0
	}
}

// summarizeShellLike recognizes the devops.ShellResult shape (fields named
// Command/Output/ExitCode) via reflection, avoiding an import cycle between
// common and devops. Returns the summary line and the number of output lines.
func summarizeShellLike(v reflect.Value) (summary string, lines int, ok bool) {
	for _, name := range []string{"Command", "Output", "ExitCode"} {
		f := v.FieldByName(name)
		if !f.IsValid() {
			return "", 0, false
		}
	}
	cmd := v.FieldByName("Command")
	out := v.FieldByName("Output")
	code := v.FieldByName("ExitCode")
	if cmd.Kind() != reflect.String || out.Kind() != reflect.String || code.Kind() != reflect.Int {
		return "", 0, false
	}

	output := out.String()
	lines = strings.Count(output, "\n")
	if output != "" && !strings.HasSuffix(output, "\n") {
		lines++
	}
	return fmt.Sprintf("Shell · exit %d · %d output line(s)", code.Int(), lines), lines, true
}

// humanizeName converts a Go type name to a lowercase, space-separated label,
// e.g. "FirewallRule" -> "firewall rule", "CPUInfo" -> "cpuinfo".
func humanizeName(name string) string {
	if name == "" {
		return ""
	}
	var b strings.Builder
	runes := []rune(name)
	for i, r := range runes {
		if i > 0 && r >= 'A' && r <= 'Z' && !(runes[i-1] >= 'A' && runes[i-1] <= 'Z') {
			b.WriteByte(' ')
		}
		b.WriteRune(r)
	}
	return strings.ToLower(b.String())
}

// pluralizeNoun applies a naive pluralization suitable for summary labels.
func pluralizeNoun(word string) string {
	if word == "" {
		return "records"
	}
	switch {
	case strings.HasSuffix(word, "s"), strings.HasSuffix(word, "x"),
		strings.HasSuffix(word, "ch"), strings.HasSuffix(word, "sh"):
		return word + "es"
	case strings.HasSuffix(word, "y") && len(word) > 1 &&
		!strings.ContainsRune("aeiou", rune(word[len(word)-2])):
		return word[:len(word)-1] + "ies"
	default:
		return word + "s"
	}
}

// WorkflowCategory groups workflows into functional domains for library filtering.
type WorkflowCategory string

const (
	WorkflowCategorySystem       WorkflowCategory = "system"
	WorkflowCategorySecurity     WorkflowCategory = "security"
	WorkflowCategoryNetwork      WorkflowCategory = "network"
	WorkflowCategoryIntelligence WorkflowCategory = "intelligence"
	WorkflowCategoryDevOps       WorkflowCategory = "devops"
)

// WorkflowDefinition defines a reusable operational sequence.
type WorkflowDefinition struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Category    WorkflowCategory `json:"category"`
	Why         string           `json:"why"`
	Risks       []string         `json:"risks"`
	TypicalVals string           `json:"typical_values"`
	Steps       []WorkflowStep   `json:"steps"`
}

// WorkflowEngine manages the library of reusable operational workflows.
type WorkflowEngine struct {
	mu      sync.RWMutex
	library map[string]WorkflowDefinition
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
