package common

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"
	"unicode/utf8"
)

// shellLike replicates the field names of devops.ShellResult so the
// reflection-based summarizer can be tested without an import cycle.
type shellLike struct {
	Command  string
	Output   string
	ExitCode int
	Duration time.Duration
}

type sampleProcess struct {
	Name       string
	CPUPercent float64
	MemBytes   uint64
}

func TestSummarizeResult_Nil(t *testing.T) {
	summary, items := SummarizeResult(nil)
	if summary != "No data returned" || items != 0 {
		t.Fatalf("nil: got (%q, %d), want (\"No data returned\", 0)", summary, items)
	}
}

func TestSummarizeResult_String(t *testing.T) {
	summary, items := SummarizeResult("hello world")
	if summary != "hello world" || items != 0 {
		t.Fatalf("string: got (%q, %d)", summary, items)
	}

	long := strings.Repeat("a", 300)
	summary, _ = SummarizeResult(long)
	if got := utf8.RuneCountInString(summary); got != 141 { // 140 runes + ellipsis
		t.Fatalf("long string truncated to %d runes, want 141", got)
	}
}

func TestSummarizeResult_Scalar(t *testing.T) {
	summary, items := SummarizeResult(42)
	if summary != "42" || items != 0 {
		t.Fatalf("scalar: got (%q, %d), want (\"42\", 0)", summary, items)
	}
}

func TestSummarizeResult_Slice(t *testing.T) {
	summary, items := SummarizeResult([]sampleProcess{{Name: "a"}, {Name: "b"}})
	if summary != "2 sample processes" || items != 2 {
		t.Fatalf("slice: got (%q, %d), want (\"2 sample processes\", 2)", summary, items)
	}

	summary, items = SummarizeResult([]string{})
	if summary != "No records" || items != 0 {
		t.Fatalf("empty slice: got (%q, %d)", summary, items)
	}
}

func TestSummarizeResult_Map(t *testing.T) {
	summary, items := SummarizeResult(map[string]any{"cpu": 1, "mem": 2, "disk": 3})
	if summary != "3 entries" || items != 3 {
		t.Fatalf("map: got (%q, %d), want (\"3 entries\", 3)", summary, items)
	}

	summary, items = SummarizeResult(map[string]any{})
	if summary != "No entries" || items != 0 {
		t.Fatalf("empty map: got (%q, %d)", summary, items)
	}
}

func TestSummarizeResult_ShellLike(t *testing.T) {
	res := shellLike{
		Command:  "Get-Process",
		Output:   "line1\nline2\n",
		ExitCode: 0,
		Duration: time.Second,
	}
	summary, items := SummarizeResult(res)
	if !strings.Contains(summary, "exit 0") || !strings.Contains(summary, "2 output line") {
		t.Fatalf("shell summary: got %q", summary)
	}
	if items != 2 {
		t.Fatalf("shell items: got %d, want 2", items)
	}

	// Pointer variant (RunPowerShell returns *ShellResult).
	summary, items = SummarizeResult(&res)
	if !strings.Contains(summary, "exit 0") || items != 2 {
		t.Fatalf("pointer shell: got (%q, %d)", summary, items)
	}
}

func TestSummarizeResult_StructSnapshot(t *testing.T) {
	summary, items := SummarizeResult(sampleProcess{Name: "a", CPUPercent: 1})
	if !strings.Contains(summary, "snapshot") || items != 0 {
		t.Fatalf("struct: got (%q, %d)", summary, items)
	}
}

func TestNewSuccessStepResult(t *testing.T) {
	started := time.Now().Add(-250 * time.Millisecond)
	res := NewSuccessStepResult([]string{"a", "b", "c"}, started)

	if res.Status != StepStatusSuccess {
		t.Fatalf("status: got %q", res.Status)
	}
	if res.Items != 3 {
		t.Fatalf("items: got %d, want 3", res.Items)
	}
	if res.DurationNS < 0 {
		t.Fatalf("duration negative: %d", res.DurationNS)
	}
	if _, err := time.Parse(time.RFC3339, res.Timestamp); err != nil {
		t.Fatalf("timestamp %q not RFC3339: %v", res.Timestamp, err)
	}
	if res.Data == nil {
		t.Fatal("data not passed through")
	}
	if res.Error != "" {
		t.Fatalf("error should be empty, got %q", res.Error)
	}
}

func TestNewErrorStepResult(t *testing.T) {
	started := time.Now()
	res := NewErrorStepResult(errors.New("boom"), started)

	if res.Status != StepStatusError {
		t.Fatalf("status: got %q", res.Status)
	}
	if res.Error != "boom" {
		t.Fatalf("error: got %q", res.Error)
	}
	if res.Summary != "Step failed" {
		t.Fatalf("summary: got %q", res.Summary)
	}
	if res.Data != nil {
		t.Fatal("error envelope should not carry data")
	}
}

func TestStepResult_JSONShape(t *testing.T) {
	res := NewSuccessStepResult(map[string]any{"cpu": 1}, time.Now().Add(-time.Millisecond))

	data, err := json.Marshal(res)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	// The envelope must serialize with stable, lowercase snake_case keys.
	for _, want := range []string{
		`"status":"success"`,
		`"summary"`,
		`"items"`,
		`"duration_ns"`,
		`"timestamp"`,
		`"data"`,
	} {
		if !strings.Contains(string(data), want) {
			t.Errorf("marshaled JSON missing %s: %s", want, string(data))
		}
	}
	if strings.Contains(string(data), `"DurationNS"`) {
		t.Errorf("envelope leaked Go field name: %s", string(data))
	}
}
