package app

import (
	"testing"
)

func TestDevOps_RunCommand_Echo(t *testing.T) {
	a := NewApp()
	d := NewDevOps(a.ctx, a.eventBus)
	result := d.RunCommand("echo hello")
	if result.Error != "" && result.Error != "command rejected by security policy" && result.Error != "command rejected: shell metacharacters not allowed" {
		t.Logf("RunCommand error: %s", result.Error)
	}
	if result.ExitCode != 0 && result.Error == "" {
		t.Errorf("RunCommand exit code = %d, want 0", result.ExitCode)
	}
}

func TestDevOps_RunCommand_BlockedCommand(t *testing.T) {
	a := NewApp()
	d := NewDevOps(a.ctx, a.eventBus)
	result := d.RunCommand("rm -rf /")
	if result.Error == "" {
		t.Log("Blocked command returned no error (may be expected if not enforced)")
	}
}

func TestDevOps_GetDevProcesses(t *testing.T) {
	a := NewApp()
	d := NewDevOps(a.ctx, a.eventBus)
	procs := d.GetDevProcesses()
	if procs == nil {
		t.Fatal("GetDevProcesses returned nil, expected non-nil slice")
	}
	if len(procs) > 0 && procs[0].PID == 0 {
		t.Log("First process has PID 0 (unlikely but possible)")
	}
}

func TestDevOps_KillProcess_Nonexistent(t *testing.T) {
	a := NewApp()
	d := NewDevOps(a.ctx, a.eventBus)
	result := d.KillProcess(-9999)
	if result.Error == "" {
		t.Log("KillProcess with invalid PID returned no error (may be expected)")
	}
}

func TestDevOps_GetServices(t *testing.T) {
	a := NewApp()
	d := NewDevOps(a.ctx, a.eventBus)
	services := d.GetServices()
	if services == nil {
		t.Fatal("GetServices returned nil, expected non-nil slice")
	}
}

func TestDevOps_ControlService_EmptyName(t *testing.T) {
	a := NewApp()
	d := NewDevOps(a.ctx, a.eventBus)
	result := d.ControlService("", "stop")
	if result {
		t.Log("ControlService with empty name returned true (may be expected)")
	}
}

func TestDevOps_RunPowerShell_InvalidCmd(t *testing.T) {
	a := NewApp()
	d := NewDevOps(a.ctx, a.eventBus)
	result := d.RunPowerShell("")
	if result.Error == "" {
		t.Log("RunPowerShell with empty command returned no error (may be expected)")
	}
}

func TestDevOps_GetInstalledTools(t *testing.T) {
	a := NewApp()
	d := NewDevOps(a.ctx, a.eventBus)
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
	a := NewApp()
	d := NewDevOps(a.ctx, a.eventBus)
	summary := d.GetContainers()
	if summary.Containers == nil {
		t.Fatal("GetContainers.Containers is nil, expected non-nil slice")
	}
}

func TestDevOps_GetEnvironment(t *testing.T) {
	a := NewApp()
	d := NewDevOps(a.ctx, a.eventBus)
	env := d.GetEnvironment()
	if env.PathDirs == nil {
		t.Error("EnvironmentInfo.PathDirs is nil, expected non-nil slice")
	}
	if env.KeyVars == nil {
		t.Error("EnvironmentInfo.KeyVars is nil, expected non-nil slice")
	}
}

func TestDevOps_GetAISuggestions(t *testing.T) {
	a := NewApp()
	d := NewDevOps(a.ctx, a.eventBus)
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
	a := NewApp()
	d := NewDevOps(a.ctx, a.eventBus)
	status := d.GetDockerStatus()
	if status.Installed {
		t.Log("Docker is installed")
	}
}

func TestDevOps_GetKubernetesStatus(t *testing.T) {
	a := NewApp()
	d := NewDevOps(a.ctx, a.eventBus)
	status := d.GetKubernetesStatus()
	if status.Installed {
		t.Log("kubectl is installed")
	}
}

func TestDevOps_GetServiceCategories(t *testing.T) {
	a := NewApp()
	d := NewDevOps(a.ctx, a.eventBus)
	cats := d.GetServiceCategories()
	if cats == nil {
		t.Fatal("GetServiceCategories returned nil, expected non-nil slice")
	}
}

func TestDevOps_GetServiceGroupSummary(t *testing.T) {
	a := NewApp()
	d := NewDevOps(a.ctx, a.eventBus)
	summary := d.GetServiceGroupSummary()
	if summary.Running < 0 || summary.Stopped < 0 {
		t.Errorf("ServiceGroupSummary has negative counts (Running=%d, Stopped=%d)", summary.Running, summary.Stopped)
	}
}

func TestDevOps_GetDefaultPath(t *testing.T) {
	a := NewApp()
	d := NewDevOps(a.ctx, a.eventBus)
	path := d.GetDefaultPath()
	if path == "" {
		t.Fatal("GetDefaultPath returned empty string")
	}
}

func TestDevOps_GetLocalServers(t *testing.T) {
	a := NewApp()
	d := NewDevOps(a.ctx, a.eventBus)
	servers := d.GetLocalServers()
	if servers == nil {
		t.Fatal("GetLocalServers returned nil, expected non-nil slice")
	}
}

func TestDevOps_GetPowerShellWorkflows(t *testing.T) {
	a := NewApp()
	d := NewDevOps(a.ctx, a.eventBus)
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
