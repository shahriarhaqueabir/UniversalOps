package aiops

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestNewModel(t *testing.T) {
	m := NewModel()
	if m == nil {
		t.Fatal("NewModel() returned nil")
	}
	if m.tabIndex != 0 {
		t.Errorf("tabIndex = %d, want 0", m.tabIndex)
	}
	if m.ollamaStatus != nil {
		t.Error("ollamaStatus should be nil initially")
	}
	if len(m.messages) != 0 {
		t.Errorf("messages = %d, want 0", len(m.messages))
	}
	if len(m.reports) != 0 {
		t.Errorf("reports = %d, want 0", len(m.reports))
	}
}

func TestModelTabIndex(t *testing.T) {
	m := NewModel()
	if got := m.TabIndex(); got != 0 {
		t.Errorf("TabIndex() = %d, want 0", got)
	}
}

func TestModelError(t *testing.T) {
	m := NewModel()
	if err := m.Error(); err != nil {
		t.Errorf("Error() = %v, want nil", err)
	}
}

func TestModelOllamaAvailable(t *testing.T) {
	m := NewModel()
	if m.ollamaAvailable() {
		t.Error("ollamaAvailable() should be false when status is nil")
	}

	m.ollamaStatus = &OllamaStatus{Available: false}
	if m.ollamaAvailable() {
		t.Error("ollamaAvailable() should be false when Available is false")
	}

	m.ollamaStatus = &OllamaStatus{Available: true}
	if !m.ollamaAvailable() {
		t.Error("ollamaAvailable() should be true when Available is true")
	}
}

func TestGenerateReport(t *testing.T) {
	tests := []struct {
		name     string
		sections []ReportSection
		want     string
	}{
		{
			name:     "empty sections",
			sections: nil,
			want:     "",
		},
		{
			name: "single section",
			sections: []ReportSection{
				{Title: "CPU", Content: "Usage: 45%"},
			},
			want: "=== CPU ===\nUsage: 45%\n",
		},
		{
			name: "multiple sections",
			sections: []ReportSection{
				{Title: "CPU", Content: "Usage: 45%"},
				{Title: "Memory", Content: "Usage: 60%"},
			},
			want: "=== CPU ===\nUsage: 45%\n\n=== Memory ===\nUsage: 60%\n",
		},
		{
			name: "content with trailing newline",
			sections: []ReportSection{
				{Title: "Disk", Content: "Usage: 80%\n"},
			},
			want: "=== Disk ===\nUsage: 80%\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GenerateReport(tt.sections); got != tt.want {
				t.Errorf("GenerateReport() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestReportToMarkdown(t *testing.T) {
	tests := []struct {
		name     string
		sections []ReportSection
		want     string
	}{
		{
			name:     "empty sections",
			sections: nil,
			want:     "",
		},
		{
			name: "single section",
			sections: []ReportSection{
				{Title: "CPU", Content: "Usage: 45%"},
			},
			want: "## CPU\n\nUsage: 45%\n",
		},
		{
			name: "multiple sections",
			sections: []ReportSection{
				{Title: "CPU", Content: "Usage: 45%"},
				{Title: "Memory", Content: "Usage: 60%"},
			},
			want: "## CPU\n\nUsage: 45%\n\n## Memory\n\nUsage: 60%\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ReportToMarkdown(tt.sections); got != tt.want {
				t.Errorf("ReportToMarkdown() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestReportToJSON(t *testing.T) {
	sections := []ReportSection{{Title: "CPU", Content: "Usage: 45%"}}
	got, err := ReportToJSON(sections)
	if err != nil {
		t.Fatalf("ReportToJSON() error = %v", err)
	}

	var decoded []ReportSection
	if err := json.Unmarshal([]byte(got), &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if len(decoded) != 1 || decoded[0].Title != "CPU" {
		t.Errorf("decoded = %+v, want CPU section", decoded)
	}
}

func TestReportToCSV(t *testing.T) {
	got, err := ReportToCSV([]ReportSection{
		{Title: "CPU", Content: "Usage: 45%"},
		{Title: "Notes", Content: "line one\nline two"},
	})
	if err != nil {
		t.Fatalf("ReportToCSV() error = %v", err)
	}
	if !strings.HasPrefix(got, "title,content\n") {
		t.Errorf("CSV = %q, want header", got)
	}
	if !strings.Contains(got, "CPU,Usage: 45%") {
		t.Errorf("CSV = %q, want CPU row", got)
	}
	if !strings.Contains(got, "\"line one\nline two\"") {
		t.Errorf("CSV = %q, want quoted multiline content", got)
	}
}

func TestSaveReport(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "report.txt")

	err := SaveReport("hello world", path)
	if err != nil {
		t.Fatalf("SaveReport() error = %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile() error = %v", err)
	}
	if string(data) != "hello world" {
		t.Errorf("content = %q, want %q", string(data), "hello world")
	}
}

func TestSaveReport_InvalidPath(t *testing.T) {
	err := SaveReport("content", "/nonexistent/dir/report.txt")
	if err == nil {
		t.Error("SaveReport() should return error for invalid path")
	}
}

func TestAddSection(t *testing.T) {
	tests := []struct {
		name     string
		sections []ReportSection
		input    string
		wantLen  int
	}{
		{
			name:     "add title and content",
			sections: nil,
			input:    "CPU|Usage: 45%",
			wantLen:  1,
		},
		{
			name:     "add title only",
			sections: nil,
			input:    "Memory",
			wantLen:  1,
		},
		{
			name:     "empty input adds nothing",
			sections: nil,
			input:    "",
			wantLen:  0,
		},
		{
			name:     "appends to existing sections",
			sections: []ReportSection{{Title: "CPU", Content: "45%"}},
			input:    "Memory|60%",
			wantLen:  2,
		},
		{
			name:     "trims whitespace from title",
			sections: nil,
			input:    "  Disk  |  80%",
			wantLen:  1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := addSection(tt.sections, tt.input)
			if len(got) != tt.wantLen {
				t.Errorf("len(sections) = %d, want %d", len(got), tt.wantLen)
			}
		})
	}
}

func TestModelInit(t *testing.T) {
	m := NewModel()
	cmd := m.Init()
	if cmd == nil {
		t.Error("Init() returned nil command")
	}

	// The command should produce an OllamaStatusMsg
	msg := cmd()
	if _, ok := msg.(OllamaStatusMsg); !ok {
		t.Errorf("Init() command produced %T, want OllamaStatusMsg", msg)
	}
}

func TestModelUpdate_OllamaStatusMsg(t *testing.T) {
	m := NewModel()
	status := &OllamaStatus{Available: true, Model: "test", Version: "0.1"}

	cmd := m.Update(OllamaStatusMsg{Status: status})
	if cmd != nil {
		t.Error("Update(OllamaStatusMsg) should return nil command")
	}
	if m.ollamaStatus != status {
		t.Error("ollamaStatus not updated")
	}
}

func TestModelUpdate_ChatResultMsg(t *testing.T) {
	m := NewModel()

	// Success case
	cmd := m.Update(ChatResultMsg{Response: "hello"})
	if cmd != nil {
		t.Error("Update(ChatResultMsg) should return nil command")
	}
	if len(m.messages) != 1 {
		t.Errorf("messages = %d, want 1", len(m.messages))
	}
	if m.messages[0].Content != "hello" {
		t.Errorf("message content = %q, want %q", m.messages[0].Content, "hello")
	}
	if m.messages[0].Role != "assistant" {
		t.Errorf("message role = %q, want %q", m.messages[0].Role, "assistant")
	}

	// Error case
	m.err = nil
	cmd = m.Update(ChatResultMsg{Err: assertError("test error")})
	if cmd != nil {
		t.Error("Update(ChatResultMsg{Err}) should return nil command")
	}
	if m.err == nil {
		t.Error("err should be set")
	}
	if len(m.messages) != 1 {
		t.Errorf("messages should not change on error, got %d", len(m.messages))
	}
}

func TestModelUpdate_ReportSavedMsg(t *testing.T) {
	m := NewModel()

	// Success case
	cmd := m.Update(ReportSavedMsg{Path: "/tmp/report.txt"})
	if cmd != nil {
		t.Error("Update(ReportSavedMsg) should return nil command")
	}
	if m.output != "Report saved to: /tmp/report.txt" {
		t.Errorf("output = %q, want %q", m.output, "Report saved to: /tmp/report.txt")
	}

	// Error case
	cmd = m.Update(ReportSavedMsg{Err: assertError("write failed")})
	if cmd != nil {
		t.Error("Update(ReportSavedMsg{Err}) should return nil command")
	}
	if m.err == nil {
		t.Error("err should be set")
	}
}

func TestModelHandleReportSubmit(t *testing.T) {
	t.Run("generate command", func(t *testing.T) {
		m := NewModel()
		m.reports = []ReportSection{{Title: "Test", Content: "content"}}
		cmd := m.handleReportSubmit("generate")
		if cmd != nil {
			t.Error("handleReportSubmit(generate) should return nil command")
		}
		if m.output == "" {
			t.Error("output should be set")
		}
	})

	t.Run("markdown command", func(t *testing.T) {
		m := NewModel()
		m.reports = []ReportSection{{Title: "Test", Content: "content"}}
		cmd := m.handleReportSubmit("markdown")
		if cmd != nil {
			t.Error("handleReportSubmit(markdown) should return nil command")
		}
		if m.output == "" {
			t.Error("output should be set")
		}
	})

	t.Run("json command", func(t *testing.T) {
		m := NewModel()
		m.reports = []ReportSection{{Title: "Test", Content: "content"}}
		cmd := m.handleReportSubmit("json")
		if cmd != nil {
			t.Error("handleReportSubmit(json) should return nil command")
		}
		if !strings.Contains(m.output, `"Title": "Test"`) {
			t.Errorf("output = %q, want JSON report", m.output)
		}
	})

	t.Run("csv command", func(t *testing.T) {
		m := NewModel()
		m.reports = []ReportSection{{Title: "Test", Content: "content"}}
		cmd := m.handleReportSubmit("csv")
		if cmd != nil {
			t.Error("handleReportSubmit(csv) should return nil command")
		}
		if !strings.Contains(m.output, "Test,content") {
			t.Errorf("output = %q, want CSV report", m.output)
		}
	})

	t.Run("clear command", func(t *testing.T) {
		m := NewModel()
		m.reports = []ReportSection{{Title: "Test", Content: "content"}}
		m.output = "some output"
		m.err = assertError("old error")
		cmd := m.handleReportSubmit("clear")
		if cmd != nil {
			t.Error("handleReportSubmit(clear) should return nil command")
		}
		if len(m.reports) != 0 {
			t.Errorf("reports = %d, want 0", len(m.reports))
		}
		if m.output != "" {
			t.Error("output should be cleared")
		}
		if m.err != nil {
			t.Error("err should be cleared")
		}
	})

	t.Run("save without content returns error", func(t *testing.T) {
		m := NewModel()
		// No reports added, so content is empty
		cmd := m.handleReportSubmit("save")
		if cmd != nil {
			t.Error("handleReportSubmit(save) with no content should return nil command")
		}
		if m.err == nil {
			t.Error("err should be set for empty save")
		}
	})

	t.Run("save-json custom path", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "report.json")
		m := NewModel()
		m.reports = []ReportSection{{Title: "Test", Content: "content"}}
		cmd := m.handleReportSubmit("save-json:" + path)
		if cmd == nil {
			t.Fatal("handleReportSubmit(save-json) should return command")
		}
		msg := cmd()
		saved, ok := msg.(ReportSavedMsg)
		if !ok {
			t.Fatalf("save command returned %T, want ReportSavedMsg", msg)
		}
		if saved.Err != nil {
			t.Fatalf("save-json error = %v", saved.Err)
		}
		if saved.Path != path {
			t.Errorf("saved path = %q, want %q", saved.Path, path)
		}
	})

	t.Run("save-csv custom path", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "report.csv")
		m := NewModel()
		m.reports = []ReportSection{{Title: "Test", Content: "content"}}
		cmd := m.handleReportSubmit("save-csv:" + path)
		if cmd == nil {
			t.Fatal("handleReportSubmit(save-csv) should return command")
		}
		msg := cmd()
		saved, ok := msg.(ReportSavedMsg)
		if !ok {
			t.Fatalf("save command returned %T, want ReportSavedMsg", msg)
		}
		if saved.Err != nil {
			t.Fatalf("save-csv error = %v", saved.Err)
		}
	})

	t.Run("add section from input", func(t *testing.T) {
		m := NewModel()
		_ = m.handleReportSubmit("CPU|Usage: 45%")
		if len(m.reports) != 1 {
			t.Errorf("reports = %d, want 1", len(m.reports))
		}
		if m.reports[0].Title != "CPU" {
			t.Errorf("title = %q, want %q", m.reports[0].Title, "CPU")
		}
	})
}

func TestModelHandleChatSubmit_OllamaNotAvailable(t *testing.T) {
	m := NewModel()
	m.ollamaStatus = &OllamaStatus{Available: false}
	cmd := m.handleChatSubmit("hello")
	if cmd != nil {
		t.Error("handleChatSubmit() should return nil when Ollama unavailable")
	}
	if m.err == nil {
		t.Error("err should be set when Ollama unavailable")
	}
}

func TestModelKeyPressTabNavigation(t *testing.T) {
	m := NewModel()

	// Tab key → tabIndex should go to 1
	m.handleKeyPress(tea.KeyPressMsg{Code: tea.KeyTab})
	if m.tabIndex != 1 {
		t.Errorf("tabIndex = %d, want 1", m.tabIndex)
	}

	// Shift+Tab → tabIndex should go back to 0
	m.handleKeyPress(tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift})
	if m.tabIndex != 0 {
		t.Errorf("tabIndex = %d, want 0", m.tabIndex)
	}
}

func TestModelKeyPressArrowNavigation(t *testing.T) {
	m := NewModel()

	// Right key → tabIndex = 1
	m.handleKeyPress(tea.KeyPressMsg{Code: tea.KeyRight})
	if m.tabIndex != 1 {
		t.Errorf("tabIndex = %d, want 1", m.tabIndex)
	}

	// Left key → tabIndex = 0
	m.handleKeyPress(tea.KeyPressMsg{Code: tea.KeyLeft})
	if m.tabIndex != 0 {
		t.Errorf("tabIndex = %d, want 0", m.tabIndex)
	}
}

func TestModelKeyPressInputOverridesNav(t *testing.T) {
	m := NewModel()

	// 'l' and 'h' are printable characters, so they get added to input
	// rather than triggering navigation. Navigation works via Tab/arrows.
	m.handleKeyPress(tea.KeyPressMsg{Text: "l"})
	if m.input != "l" {
		t.Errorf("input = %q, want %q", m.input, "l")
	}
	// tabIndex should be unchanged
	if m.tabIndex != 0 {
		t.Errorf("tabIndex = %d, want 0 (unchanged)", m.tabIndex)
	}

	m.handleKeyPress(tea.KeyPressMsg{Text: "h"})
	if m.input != "lh" {
		t.Errorf("input = %q, want %q", m.input, "lh")
	}
	if m.tabIndex != 0 {
		t.Errorf("tabIndex = %d, want 0 (unchanged)", m.tabIndex)
	}
}

func TestModelKeyPressTabWraps(t *testing.T) {
	m := NewModel()
	m.tabIndex = 1

	// Right at tab 1 → wraps to 0
	m.handleKeyPress(tea.KeyPressMsg{Code: tea.KeyRight})
	if m.tabIndex != 0 {
		t.Errorf("tabIndex = %d, want 0 (wrapped)", m.tabIndex)
	}

	// Left at tab 0 → wraps to 1
	m.handleKeyPress(tea.KeyPressMsg{Code: tea.KeyLeft})
	if m.tabIndex != 1 {
		t.Errorf("tabIndex = %d, want 1 (wrapped back)", m.tabIndex)
	}
}

func TestModelKeyPressInputAccumulation(t *testing.T) {
	m := NewModel()

	m.handleKeyPress(tea.KeyPressMsg{Text: "h"})
	m.handleKeyPress(tea.KeyPressMsg{Text: "e"})
	m.handleKeyPress(tea.KeyPressMsg{Text: "y"})
	if m.input != "hey" {
		t.Errorf("input = %q, want %q", m.input, "hey")
	}

	// Backspace
	m.handleKeyPress(tea.KeyPressMsg{Code: tea.KeyBackspace})
	if m.input != "he" {
		t.Errorf("input = %q, want %q", m.input, "he")
	}
}

func TestGenerateTextReport(t *testing.T) {
	m := NewModel()
	m.reports = []ReportSection{{Title: "Test", Content: "data"}}
	got := m.generateTextReport()
	if got == "" {
		t.Error("generateTextReport() should not be empty")
	}
}

// Helper: implements error for testing.
type assertError string

func (e assertError) Error() string { return string(e) }

// TestReportSection tests the ReportSection struct.
func TestReportSection(t *testing.T) {
	s := ReportSection{Title: "Test", Content: "content"}
	if s.Title != "Test" {
		t.Errorf("Title = %q, want %q", s.Title, "Test")
	}
	if s.Content != "content" {
		t.Errorf("Content = %q, want %q", s.Content, "content")
	}
}
