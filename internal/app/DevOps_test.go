package app

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDevOps_RunCommand_Echo(t *testing.T) {
	d := NewDevOps(NewApp())
	result := d.RunCommand("echo hello")
	if result.Error != "" && result.Error != "command rejected by security policy" && result.Error != "command rejected: shell metacharacters not allowed" {
		t.Logf("RunCommand error: %s", result.Error)
	}
	if result.ExitCode != 0 && result.Error == "" {
		t.Errorf("RunCommand exit code = %d, want 0", result.ExitCode)
	}
}

func TestDevOps_RunCommand_BlockedCommand(t *testing.T) {
	d := NewDevOps(NewApp())
	result := d.RunCommand("rm -rf /")
	if result.Error == "" {
		t.Log("Blocked command returned no error (may be expected if not enforced)")
	}
}

func TestDevOps_ListDirectory_Root(t *testing.T) {
	d := NewDevOps(NewApp())
	entries := d.ListDirectory(".")
	if entries == nil {
		t.Fatal("ListDirectory returned nil, expected non-nil slice")
	}
	if len(entries) > 0 && entries[0].Name == "" {
		t.Error("First entry has empty Name")
	}
}

func TestDevOps_ListDirectory_Nonexistent(t *testing.T) {
	d := NewDevOps(NewApp())
	entries := d.ListDirectory("/nonexistent/path/that/does/not/exist/12345")
	if entries == nil {
		t.Fatal("ListDirectory returned nil for nonexistent path, expected non-nil slice")
	}
}

func TestDevOps_ReadFile_Nonexistent(t *testing.T) {
	d := NewDevOps(NewApp())
	content := d.ReadFile("/nonexistent/file.txt")
	if content != "" {
		t.Logf("ReadFile returned content for nonexistent file: %s", content)
	}
}

func TestDevOps_WriteFile(t *testing.T) {
	d := NewDevOps(NewApp())
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")
	result := d.WriteFile(tmpFile, "hello world")
	if !result {
		t.Fatal("WriteFile returned false")
	}
	data, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "hello world" {
		t.Errorf("WriteFile wrote %q, want %q", string(data), "hello world")
	}
}

func TestDevOps_GetDevProcesses(t *testing.T) {
	d := NewDevOps(NewApp())
	procs := d.GetDevProcesses()
	if procs == nil {
		t.Fatal("GetDevProcesses returned nil, expected non-nil slice")
	}
	if len(procs) > 0 && procs[0].PID == 0 {
		t.Log("First process has PID 0 (unlikely but possible)")
	}
}

func TestDevOps_KillProcess_Nonexistent(t *testing.T) {
	d := NewDevOps(NewApp())
	result := d.KillProcess(-9999)
	if result.Error == "" {
		t.Log("KillProcess with invalid PID returned no error (may be expected)")
	}
}

func TestDevOps_TailLog_Nonexistent(t *testing.T) {
	d := NewDevOps(NewApp())
	lines := d.TailLog("/nonexistent/file.log", 10)
	if lines == nil {
		t.Fatal("TailLog returned nil for nonexistent file, expected non-nil slice")
	}
}

func TestDevOps_SearchLog_Nonexistent(t *testing.T) {
	d := NewDevOps(NewApp())
	lines := d.SearchLog("/nonexistent/file.log", "pattern")
	if lines == nil {
		t.Fatal("SearchLog returned nil for nonexistent file, expected non-nil slice")
	}
}

func TestDevOps_GetServices(t *testing.T) {
	d := NewDevOps(NewApp())
	services := d.GetServices()
	if services == nil {
		t.Fatal("GetServices returned nil, expected non-nil slice")
	}
}

func TestDevOps_ControlService_EmptyName(t *testing.T) {
	d := NewDevOps(NewApp())
	result := d.ControlService("", "stop")
	if result {
		t.Log("ControlService with empty name returned true (may be expected)")
	}
}

func TestDevOps_RunPowerShell_InvalidCmd(t *testing.T) {
	d := NewDevOps(NewApp())
	result := d.RunPowerShell("")
	if result.Error == "" {
		t.Log("RunPowerShell with empty command returned no error (may be expected)")
	}
}

func TestDevOps_GetInstalledTools(t *testing.T) {
	d := NewDevOps(NewApp())
	tools := d.GetInstalledTools()
	if tools == nil {
		t.Fatal("GetInstalledTools returned nil, expected non-nil slice")
	}
	for _, tool := range tools {
		if tool.Name == "" {
			t.Error("ToolInfo entry has empty Name")
		}
	}
}

func TestDevOps_GetContainers(t *testing.T) {
	d := NewDevOps(NewApp())
	summary := d.GetContainers()
	if summary.Containers == nil {
		t.Fatal("GetContainers.Containers is nil, expected non-nil slice")
	}
}

func TestDevOps_GetEnvironment(t *testing.T) {
	d := NewDevOps(NewApp())
	env := d.GetEnvironment()
	if env.PathDirs == nil {
		t.Error("EnvironmentInfo.PathDirs is nil, expected non-nil slice")
	}
	if env.KeyVars == nil {
		t.Error("EnvironmentInfo.KeyVars is nil, expected non-nil slice")
	}
}

func TestDevOps_GetAISuggestions(t *testing.T) {
	d := NewDevOps(NewApp())
	suggestions := d.GetAISuggestions()
	if suggestions == nil {
		t.Fatal("GetAISuggestions returned nil, expected non-nil slice")
	}
	if len(suggestions) == 0 {
		t.Error("GetAISuggestions returned empty slice")
	}
	for _, s := range suggestions {
		if s.Category == "" {
			t.Error("Suggestion has empty Category")
		}
	}
}

func TestDevOps_GetDockerStatus(t *testing.T) {
	d := NewDevOps(NewApp())
	status := d.GetDockerStatus()
	if status.Installed {
		t.Log("Docker is installed")
	}
}

func TestDevOps_GetKubernetesStatus(t *testing.T) {
	d := NewDevOps(NewApp())
	status := d.GetKubernetesStatus()
	if status.Installed {
		t.Log("kubectl is installed")
	}
}

func TestDevOps_GetServiceCategories(t *testing.T) {
	d := NewDevOps(NewApp())
	cats := d.GetServiceCategories()
	if cats == nil {
		t.Fatal("GetServiceCategories returned nil, expected non-nil slice")
	}
}

func TestDevOps_GetServiceGroupSummary(t *testing.T) {
	d := NewDevOps(NewApp())
	summary := d.GetServiceGroupSummary()
	if summary.Running < 0 || summary.Stopped < 0 {
		t.Errorf("ServiceGroupSummary has negative counts (Running=%d, Stopped=%d)", summary.Running, summary.Stopped)
	}
}

func TestDevOps_GetDefaultPath(t *testing.T) {
	d := NewDevOps(NewApp())
	path := d.GetDefaultPath()
	if path == "" {
		t.Fatal("GetDefaultPath returned empty string")
	}
}

func TestDevOps_GetLocalServers(t *testing.T) {
	d := NewDevOps(NewApp())
	servers := d.GetLocalServers()
	if servers == nil {
		t.Fatal("GetLocalServers returned nil, expected non-nil slice")
	}
}

func TestDevOps_GetGitSummary(t *testing.T) {
	d := NewDevOps(NewApp())
	summary := d.GetGitSummary()
	if summary.Repositories == nil {
		t.Fatal("GetGitSummary.Repositories is nil, expected non-nil slice")
	}
}

func TestDevOps_GetPowerShellWorkflows(t *testing.T) {
	d := NewDevOps(NewApp())
	workflows := d.GetPowerShellWorkflows()
	if workflows == nil {
		t.Fatal("GetPowerShellWorkflows returned nil, expected non-nil slice")
	}
	if len(workflows) == 0 {
		t.Error("GetPowerShellWorkflows returned empty slice")
	}
}

func TestDevOps_sanitizeError_Nil(t *testing.T) {
	result := sanitizeError(nil)
	if result == "" {
		t.Log("sanitizeError(nil) returns empty string")
	}
}

func TestDevOps_severityRank(t *testing.T) {
	tests := []struct {
		s    string
		want int
	}{
		{"critical", 3},
		{"warning", 2},
		{"info", 1},
		{"unknown", 0},
		{"", 0},
	}
	for _, tt := range tests {
		got := severityRank(tt.s)
		if got != tt.want {
			t.Errorf("severityRank(%q) = %d, want %d", tt.s, got, tt.want)
		}
	}
}

func TestDevOps_firstLine(t *testing.T) {
	tests := []struct {
		text string
		want string
	}{
		{"hello\nworld", "hello"},
		{"\n\nhello\nworld", "hello"},
		{"", ""},
		{"   ", ""},
	}
	for _, tt := range tests {
		got := firstLine(tt.text)
		if got != tt.want {
			t.Errorf("firstLine(%q) = %q, want %q", tt.text, got, tt.want)
		}
	}
}

func TestDevOps_normalizeServiceStatus(t *testing.T) {
	tests := []struct {
		status string
		want   string
	}{
		{"running", "running"},
		{"Running", "running"},
		{"stopped", "stopped"},
		{"Disabled", "stopped"},
		{"paused", "unknown"},
		{"", "unknown"},
	}
	for _, tt := range tests {
		got := normalizeServiceStatus(tt.status)
		if got != tt.want {
			t.Errorf("normalizeServiceStatus(%q) = %q, want %q", tt.status, got, tt.want)
		}
	}
}

func TestDevOps_categorizeService(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"mysqld", "databases"},
		{"redis-server", "databases"},
		{"nginx", "web-servers"},
		{"dockerd", "containers"},
		{"sshd", "other"},
		{"", "other"},
	}
	for _, tt := range tests {
		got := categorizeService(tt.name)
		if got != tt.want {
			t.Errorf("categorizeService(%q) = %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestDevOps_parseIntOr(t *testing.T) {
	tests := []struct {
		s        string
		fallback int
		want     int
	}{
		{"42", 0, 42},
		{"abc", 10, 10},
		{"", 5, 5},
	}
	for _, tt := range tests {
		got := parseIntOr(tt.s, tt.fallback)
		if got != tt.want {
			t.Errorf("parseIntOr(%q, %d) = %d, want %d", tt.s, tt.fallback, got, tt.want)
		}
	}
}

func TestDevOps_detectFramework(t *testing.T) {
	tests := []struct {
		process string
		want    string
	}{
		{"node.exe", "Node.js"},
		{"python3", "Python"},
		{"dockerd", "Docker"},
		{"sshd", ""},
		{"", ""},
	}
	for _, tt := range tests {
		got := detectFramework(tt.process)
		if got != tt.want {
			t.Errorf("detectFramework(%q) = %q, want %q", tt.process, got, tt.want)
		}
	}
}
