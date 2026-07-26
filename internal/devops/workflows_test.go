package devops

import (
	"strings"
	"testing"
	"time"
)

func TestDevReportString_Empty(t *testing.T) {
	r := &DevReport{}
	output := r.String()
	if !strings.Contains(output, "Development Operations Report") {
		t.Error("String() should include header")
	}
}

func TestDevReportString_WithShellResult(t *testing.T) {
	r := &DevReport{
		ShellResults: []ShellResult{
			{Command: "go version", ExitCode: 0, Duration: time.Second, Output: "go version go1.26.0"},
			{Command: "git --version", ExitCode: 0, Duration: 500 * time.Millisecond, Output: "git version 2.45.0"},
		},
	}
	output := r.String()
	checks := []string{
		"go version",
		"git --version",
		"exit=0",
		"1s",
		"500ms",
		"go1.26.0",
		"2.45.0",
	}
	for _, c := range checks {
		if !strings.Contains(output, c) {
			t.Errorf("String() missing expected text %q", c)
		}
	}
}

func TestDevReportString_WithProcessesAndServices(t *testing.T) {
	r := &DevReport{
		Processes: []ProcessEntry{
			{PID: 1234, Name: "go.exe", CPU: 5.0, Memory: 50.0},
		},
		Services: []ServiceEntry{
			{Name: "Spooler", Status: "Running", StartType: "Automatic"},
		},
	}
	output := r.String()
	if !strings.Contains(output, "PROCESSES: 1 sampled") {
		t.Errorf("String() should include process count, got: %s", output)
	}
	if !strings.Contains(output, "SERVICES: 1 sampled") {
		t.Errorf("String() should include service count, got: %s", output)
	}
}

func TestDevReportMarkdown_Empty(t *testing.T) {
	r := &DevReport{}
	output := r.Markdown()
	if !strings.Contains(output, "Development Operations Report") {
		t.Error("Markdown() should include header")
	}
}

func TestDevReportMarkdown_WithData(t *testing.T) {
	r := &DevReport{
		ShellResults: []ShellResult{
			{Command: "go version", ExitCode: 0, Duration: time.Second, Output: "go version go1.26.0 windows/amd64"},
		},
		Processes: []ProcessEntry{
			{PID: 5678, Name: "node.exe", CPU: 12.3, Memory: 200.5},
		},
		Services: []ServiceEntry{
			{Name: "Docker", Status: "Running", StartType: "Automatic"},
			{Name: "WSearch", Status: "Stopped", StartType: "Manual"},
		},
	}
	output := r.Markdown()
	checks := []string{
		"Development Operations Report",
		"go version",
		"node.exe",
		"5678",
		"12.3",
		"Docker",
		"Running",
		"WSearch",
		"Stopped",
		"Manual",
	}
	for _, c := range checks {
		if !strings.Contains(output, c) {
			t.Errorf("Markdown() missing expected text %q", c)
		}
	}
}

func TestDevReportMarkdown_OutputTruncation(t *testing.T) {
	longOutput := ""
	for i := 0; i < 100; i++ {
		longOutput += "x"
	}
	r := &DevReport{
		ShellResults: []ShellResult{
			{Command: "long command", ExitCode: 0, Duration: time.Second, Output: longOutput},
		},
	}
	output := r.Markdown()
	if !strings.Contains(output, "...") {
		t.Errorf("Markdown() should truncate output longer than 60 chars, got: %s", output)
	}
}

func TestRunDevDiagnostics_ReturnsReport(t *testing.T) {
	report, err := RunDevDiagnostics()
	if err != nil {
		t.Fatalf("RunDevDiagnostics() failed: %v", err)
	}
	if report == nil {
		t.Fatal("RunDevDiagnostics() returned nil report")
	}
	// Should have at least processes and services (shell commands like "go version" are not in allowlist on Windows)
	if len(report.Processes) == 0 && len(report.Services) == 0 {
		t.Error("RunDevDiagnostics() should have at least some data in processes or services")
	}
	t.Logf("Dev report: %d shell results, %d processes, %d services",
		len(report.ShellResults), len(report.Processes), len(report.Services))
}
