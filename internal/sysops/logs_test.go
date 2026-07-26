package sysops

import (
	"encoding/json"
	"strings"
	"testing"
)

// ── Existing integration test ────────────────────────────────────────────────

func TestGetSystemLogs(t *testing.T) {
	result, err := GetSystemLogs(10, "journald")
	if err != nil {
		// Windows PowerShell event log may fail in test environment
		if strings.Contains(err.Error(), "powershell") || strings.Contains(err.Error(), "not found") {
			t.Skip("System logs not available in this environment")
		}
		t.Logf("GetSystemLogs returned error (may be expected): %v", err)
		return
	}
	t.Logf("Got %d log entries from %s", result.Total, result.Source)
}

// ── parseDotNetDate unit tests ───────────────────────────────────────────────

func TestParseDotNetDate(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{input: `""`, expected: ""},
		{input: `null`, expected: ""},
		{input: `"raw string"`, expected: "raw string"},
		{input: `"/Date(1784892277000)/"`, expected: "2026-07-24T11:24:37Z"},
		{input: `"/Date(1704067200000)/"`, expected: "2024-01-01T00:00:00Z"},
		{input: `"/Date(0)/"`, expected: "1970-01-01T00:00:00Z"},
	}
	for _, tt := range tests {
		got := parseDotNetDate(json.RawMessage(tt.input))
		if got != tt.expected {
			t.Errorf("parseDotNetDate(%q) = %q; want %q", tt.input, got, tt.expected)
		}
	}
}

// ── Windows JSON path ────────────────────────────────────────────────────────

func TestParseWindowsJson_Array(t *testing.T) {
	output := `[
  {
    "TimeGenerated": "/Date(1784892277000)/",
    "EntryType": "Error",
    "Source": "Microsoft-Windows-Kernel-Power",
    "Message": "The system has rebooted without cleanly shutting down first."
  },
  {
    "TimeGenerated": "/Date(1784892278000)/",
    "EntryType": "Warning",
    "Source": "Microsoft-Windows-Perflib",
    "Message": "The Open procedure for service \"WmiApRpl\" failed."
  }
]`
	entries := parseLogOutput(output, "System", "windows")
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}

	e0 := entries[0]
	if e0.Timestamp != "2026-07-24T11:24:37Z" {
		t.Errorf("entry[0].Timestamp = %q; want ISO 8601", e0.Timestamp)
	}
	if e0.Level != "error" {
		t.Errorf("entry[0].Level = %q; want error", e0.Level)
	}
	if e0.Source != "Microsoft-Windows-Kernel-Power" {
		t.Errorf("entry[0].Source = %q", e0.Source)
	}
	if e0.Message != "The system has rebooted without cleanly shutting down first." {
		t.Errorf("entry[0].Message = %q", e0.Message)
	}

	e1 := entries[1]
	if e1.Timestamp != "2026-07-24T11:24:38Z" {
		t.Errorf("entry[1].Timestamp = %q; want ISO 8601", e1.Timestamp)
	}
	if e1.Level != "warning" {
		t.Errorf("entry[1].Level = %q; want warning", e1.Level)
	}
	if e1.Source != "Microsoft-Windows-Perflib" {
		t.Errorf("entry[1].Source = %q", e1.Source)
	}
}

func TestParseWindowsJson_Single(t *testing.T) {
	output := `{
    "TimeGenerated": "/Date(1704067200000)/",
    "EntryType": "Information",
    "Source": "Test-Source",
    "Message": "Single event message"
}`
	entries := parseLogOutput(output, "System", "windows")
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	e := entries[0]
	if e.Timestamp != "2024-01-01T00:00:00Z" {
		t.Errorf("Timestamp = %q; want ISO 8601", e.Timestamp)
	}
	if e.Level != "info" {
		t.Errorf("Level = %q; want info", e.Level)
	}
	if e.Source != "Test-Source" {
		t.Errorf("Source = %q", e.Source)
	}
	if e.Message != "Single event message" {
		t.Errorf("Message = %q", e.Message)
	}
}

func TestParseWindowsJson_EntryTypes(t *testing.T) {
	tests := []struct {
		entryType string
		wantLevel string
	}{
		{"Error", "error"},
		{"Warning", "warning"},
		{"Information", "info"},
		{"SuccessAudit", "info"},
		{"FailureAudit", "info"},
		{"UnknownType", "info"},
		{"", "info"},
	}
	for _, tt := range tests {
		output := `{"TimeGenerated":"","EntryType":"` + tt.entryType + `","Source":"S","Message":"M"}`
		entries := parseLogOutput(output, "System", "windows")
		if len(entries) != 1 {
			t.Fatalf("EntryType=%q: expected 1 entry, got %d", tt.entryType, len(entries))
		}
		if entries[0].Level != tt.wantLevel {
			t.Errorf("EntryType=%q → Level=%q; want %q", tt.entryType, entries[0].Level, tt.wantLevel)
		}
	}
}

func TestParseWindowsJson_DotNetDate(t *testing.T) {
	output := `{
    "TimeGenerated": "/Date(1784892277000)/",
    "EntryType": "Information",
    "Source": "S",
    "Message": "M"
}`
	entries := parseLogOutput(output, "System", "windows")
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Timestamp != "2026-07-24T11:24:37Z" {
		t.Errorf("got %q; want 2026-07-24T11:24:37Z", entries[0].Timestamp)
	}
}

func TestParseWindowsJson_Utf16LeNullBytes(t *testing.T) {
	// Simulate UTF-16LE output: null bytes between every ASCII char
	input := `{"TimeGenerated":"","EntryType":"Information","Source":"Win","Message":"Msg"}`
	var buf []byte
	for _, b := range []byte(input) {
		buf = append(buf, b, 0x00)
	}
	output := string(buf)

	entries := parseLogOutput(output, "System", "windows")
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry after null-byte stripping, got %d", len(entries))
	}
	if entries[0].Source != "Win" {
		t.Errorf("Source = %q; want Win", entries[0].Source)
	}
	if entries[0].Message != "Msg" {
		t.Errorf("Message = %q; want Msg", entries[0].Message)
	}
}

func TestParseWindowsJson_Empty(t *testing.T) {
	if got := parseLogOutput("", "System", "windows"); got != nil {
		t.Error("expected nil for empty input")
	}
	if got := parseLogOutput("[]", "System", "windows"); got != nil {
		t.Error("expected nil for []")
	}
	if got := parseLogOutput("null", "System", "windows"); got != nil {
		t.Error("expected nil for null")
	}
	if got := parseLogOutput("  ", "System", "windows"); got != nil {
		t.Error("expected nil for whitespace")
	}
}

func TestParseWindowsJson_Invalid(t *testing.T) {
	if got := parseLogOutput("{bad json}", "System", "windows"); got != nil {
		t.Error("expected nil for invalid JSON")
	}
}

func TestParseWindowsJson_SkipBadObject(t *testing.T) {
	output := `[
  {"TimeGenerated":"","EntryType":"Information","Source":"S1","Message":"good"},
  "this is not an object",
  {"TimeGenerated":"","EntryType":"Error","Source":"S2","Message":"also good"}
]`
	entries := parseLogOutput(output, "System", "windows")
	// The invalid object is skipped; two valid entries remain.
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries (skipping bad object), got %d", len(entries))
	}
	if entries[0].Source != "S1" {
		t.Errorf("entry[0].Source = %q; want S1", entries[0].Source)
	}
	if entries[1].Source != "S2" {
		t.Errorf("entry[1].Source = %q; want S2", entries[1].Source)
	}
}

// ── Non-Windows path (line-by-line) ──────────────────────────────────────────

func TestParseLinux_LineByLine(t *testing.T) {
	output := "line one\nerror: something failed\nline three\n"
	entries := parseLogOutput(output, "journald", "linux")
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}
	if entries[0].Message != "line one" {
		t.Errorf("entry[0].Message = %q", entries[0].Message)
	}
	if entries[0].Source != "journald" {
		t.Errorf("entry[0].Source = %q", entries[0].Source)
	}
	if entries[1].Level != "error" {
		t.Errorf("entry[1].Level = %q; want error", entries[1].Level)
	}
}

func TestParseLinux_LevelDetection(t *testing.T) {
	tests := []struct {
		msg       string
		wantLevel string
	}{
		{"this is an Error", "error"},
		{"Fatal: something broke", "error"},
		{"Critical failure", "error"},
		{"warning: low disk", "warning"},
		{"WARN: high memory", "warning"},
		{"debug output", "debug"},
		{"trace information", "debug"},
		{"regular info line", "info"},
	}
	for _, tt := range tests {
		entries := parseLogOutput(tt.msg, "journald", "linux")
		if len(entries) != 1 {
			t.Fatalf("msg=%q: expected 1 entry, got %d", tt.msg, len(entries))
		}
		if entries[0].Level != tt.wantLevel {
			t.Errorf("msg=%q → Level=%q; want %q", tt.msg, entries[0].Level, tt.wantLevel)
		}
	}
}

func TestParseLinux_SkipEmptyLines(t *testing.T) {
	output := "first\n\n\nsecond\n  \nthird"
	entries := parseLogOutput(output, "journald", "linux")
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}
}

func TestParseLinux_Empty(t *testing.T) {
	entries := parseLogOutput("", "journald", "linux")
	if len(entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(entries))
	}
}
