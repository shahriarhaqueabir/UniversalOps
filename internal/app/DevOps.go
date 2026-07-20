package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strings"
	"time"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
	"github.com/shahriarhaqueabir/AllOpsFull/internal/devops"
	sysopsPkg "github.com/shahriarhaqueabir/AllOpsFull/internal/sysops"
)

// DevOps exposes development operations bindings to the frontend.
type DevOps struct {
	ctx      context.Context
	eventBus *common.EventBus
}

// NewDevOps creates a new DevOps facade.
func NewDevOps(ctx context.Context, eventBus *common.EventBus) *DevOps {
	return &DevOps{
		ctx:      ctx,
		eventBus: eventBus,
	}
}

// RunCommand executes a shell command and returns the result.
func (d *DevOps) RunCommand(cmd string) CommandResult {
	start := time.Now()
	result, err := devops.RunCommand(cmd)
	dur := time.Since(start).Milliseconds()
	if err != nil {
		common.LogWarn("RunCommand failed: %v", err)
		errMsg := sanitizeError(err)
		if result != nil {
			return CommandResult{
				Command:  cmd,
				Output:   result.Output,
				ExitCode: result.ExitCode,
				Duration: dur,
				Error:    errMsg,
			}
		}
		return CommandResult{Command: cmd, Duration: dur, Error: errMsg}
	}
	return CommandResult{
		Command:  cmd,
		Output:   result.Output,
		ExitCode: result.ExitCode,
		Duration: dur,
	}
}

// RunCommandLive executes a command and emits each output line as a Wails event.
// The `id` parameter allows the frontend to correlate lines to a specific command.
func (d *DevOps) RunCommandLive(cmd string, id string) CommandResult {
	start := time.Now()
	lineCh := make(chan string, 100)

	go func() {
		defer common.RecoverPanic()
		for line := range lineCh {
			wailsruntime.EventsEmit(d.ctx, EventCmdLine, map[string]string{
				"id":   id,
				"line": line,
			})
		}
		wailsruntime.EventsEmit(d.ctx, EventCmdDone, id)
	}()

	result, err := devops.RunCommandWithLiveOutput(cmd, lineCh)
	dur := time.Since(start).Milliseconds()
	if err != nil {
		common.LogWarn("RunCommandLive failed: %v", err)
		errMsg := sanitizeError(err)
		if result != nil {
			return CommandResult{
				Command:  cmd,
				Output:   result.Output,
				ExitCode: result.ExitCode,
				Duration: dur,
				Error:    errMsg,
			}
		}
		return CommandResult{Command: cmd, Duration: dur, Error: errMsg}
	}
	return CommandResult{
		Command:  cmd,
		Output:   result.Output,
		ExitCode: result.ExitCode,
		Duration: dur,
	}
}

// GetDevProcesses returns all processes by CPU usage (from system ops).
func (d *DevOps) GetDevProcesses() []ProcessInfo {
	procs, err := sysopsPkg.GetTopProcesses(100)
	if err != nil {
		common.LogWarn("GetDevProcesses failed: %v", err)
		return []ProcessInfo{}
	}
	out := make([]ProcessInfo, 0, len(procs))
	for _, p := range procs {
		out = append(out, ProcessInfo{
			PID:    p.PID,
			Name:   p.Name,
			CPU:    p.CPU,
			Memory: p.Memory,
			MemPct: p.MemPct,
			Status: p.Status,
			NumFDs: p.NumFDs,
		})
	}
	return out
}

// KillProcess terminates a process by PID.
func (d *DevOps) KillProcess(pid int) CommandResult {
	err := devops.KillProcess(int32(pid))
	if err != nil {
		return CommandResult{Error: err.Error()}
	}
	return CommandResult{Output: "Process terminated successfully"}
}

// GetServices returns a list of system services.
func (d *DevOps) GetServices() []ServiceEntry {
	services, err := devops.ListServices(0)
	if err != nil {
		common.LogWarn("GetServices failed: %v", err)
		return []ServiceEntry{}
	}
	out := make([]ServiceEntry, 0, len(services))
	for _, s := range services {
		out = append(out, ServiceEntry{
			Name:        s.Name,
			DisplayName: s.DisplayName,
			Status:      s.Status,
			StartType:   s.StartType,
		})
	}
	return out
}

// ControlService manages a service state.
func (d *DevOps) ControlService(name, action string) bool {
	common.LogInfo("Service control: %s %s", action, name)
	err := devops.ControlService(name, action)
	if err != nil {
		common.LogWarn("ControlService failed: %v", err)
		return false
	}

	if d.eventBus != nil {
		d.eventBus.Emit(common.NewEvent(
			common.CatDevOps,
			common.EventInfo,
			"devops",
			"Service state changed",
			fmt.Sprintf("Service '%s' %s", name, action),
		))
	}
	return true
}

// RunPowerShell executes a PowerShell command using the OpsForAll hybrid profile.
func (d *DevOps) RunPowerShell(cmd string) CommandResult {
	start := time.Now()
	result, err := devops.RunPowerShell(cmd)
	dur := time.Since(start).Milliseconds()
	if err != nil {
		common.LogWarn("RunPowerShell failed: %v", err)
		errMsg := sanitizeError(err)
		if result != nil {
			return CommandResult{
				Command:  cmd,
				Output:   result.Output,
				ExitCode: result.ExitCode,
				Duration: dur,
				Error:    errMsg,
			}
		}
		return CommandResult{Command: cmd, Duration: dur, Error: errMsg}
	}
	return CommandResult{
		Command:  cmd,
		Output:   result.Output,
		ExitCode: result.ExitCode,
		Duration: dur,
	}
}

// GetDefaultPath returns the application root directory for file browsing.
func (d *DevOps) GetDefaultPath() string {
	root, err := os.Getwd()
	if err != nil {
		return "."
	}
	return root
}

// GetPowerShellWorkflows returns a list of available PowerShell diagnostic workflows.
func (d *DevOps) GetPowerShellWorkflows() []string {
	return []string{
		"Invoke-OpsDailyOps",
		"Invoke-OpsSystemReview",
		"Invoke-OpsSecurityAudit",
		"Invoke-OpsNetworkDiagnostics",
		"Invoke-OpsThreatHunt",
		"Invoke-OpsChangeAudit",
		"Invoke-OpsComplianceCheck",
	}
}

// sanitizeError strips internal details from errors sent to the frontend.
func sanitizeError(err error) string {
	if err == nil {
		return ""
	}
	if errors.Is(err, devops.ErrDangerousCommand) {
		return "command rejected by security policy"
	}
	if errors.Is(err, devops.ErrShellMetachar) {
		return "command rejected: shell metacharacters not allowed"
	}
	// For all other errors, return the message but log full details server-side.
	return err.Error()
}

// toolSpec defines a tool to detect.
type toolSpec struct {
	Name    string
	Command string
	Args    []string
	Parse   func(output string, stderr string) string
}

// GetInstalledTools detects installed development tools and their versions.
func (d *DevOps) GetInstalledTools() []ToolInfo {
	tools := []toolSpec{
		{
			Name:    "Git",
			Command: "git",
			Args:    []string{"--version"},
			Parse: func(out, _ string) string {
				// "git version 2.44.0.windows.1"
				return firstLine(out)
			},
		},
		{
			Name:    "Docker",
			Command: "docker",
			Args:    []string{"--version"},
			Parse: func(out, _ string) string {
				return firstLine(out)
			},
		},
		{
			Name:    "Node.js",
			Command: "node",
			Args:    []string{"--version"},
			Parse: func(out, _ string) string {
				return firstLine(out) // "v20.12.0"
			},
		},
		{
			Name:    "Go",
			Command: "go",
			Args:    []string{"version"},
			Parse: func(out, _ string) string {
				// "go version go1.22.2 windows/amd64"
				fields := strings.Fields(firstLine(out))
				if len(fields) >= 3 {
					return fields[2] // "go1.22.2"
				}
				return firstLine(out)
			},
		},
		{
			Name:    "Python",
			Command: "python",
			Args:    []string{"--version"},
			Parse: func(out, _ string) string {
				// "Python 3.12.3"
				fields := strings.Fields(firstLine(out))
				if len(fields) >= 2 {
					return fields[1]
				}
				return firstLine(out)
			},
		},
		{
			Name:    "Java",
			Command: "java",
			Args:    []string{"-version"},
			Parse: func(out, stderr string) string {
				// java -version writes to stderr: '"17.0.2"' or 'openjdk version "17.0.2"'
				line := firstLine(stderr)
				if line == "" {
					line = firstLine(out)
				}
				return line
			},
		},
		{
			Name:    "Rust",
			Command: "rustc",
			Args:    []string{"--version"},
			Parse: func(out, _ string) string {
				// "rustc 1.77.1 (aedd173a2 2024-03-17)"
				fields := strings.Fields(firstLine(out))
				if len(fields) >= 2 {
					return fields[1]
				}
				return firstLine(out)
			},
		},
		{
			Name:    ".NET",
			Command: "dotnet",
			Args:    []string{"--version"},
			Parse: func(out, _ string) string {
				return firstLine(out)
			},
		},
	}

	results := make([]ToolInfo, 0, len(tools))
	for _, t := range tools {
		results = append(results, detectTool(t))
	}
	return results
}

// detectTool runs a single tool's version command and returns a ToolInfo.
func detectTool(t toolSpec) ToolInfo {
	toolPath, err := exec.LookPath(t.Command)
	if err != nil {
		return ToolInfo{Name: t.Name, Status: "not-found"}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, t.Command, t.Args...)
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		// Context deadline exceeded or non-zero exit — still might have output.
		if ctx.Err() == context.DeadlineExceeded {
			return ToolInfo{Name: t.Name, Status: "error", Version: "timeout", Path: toolPath}
		}
		// For tools like java -version, output is on stderr and exit code may be 0.
		// Non-zero exit with some output still counts as an error.
		output := strings.TrimSpace(stdout.String())
		errOutput := strings.TrimSpace(stderr.String())
		if output != "" || errOutput != "" {
			version := t.Parse(output, errOutput)
			if version != "" {
				return ToolInfo{Name: t.Name, Status: "installed", Version: version, Path: toolPath}
			}
		}
		return ToolInfo{Name: t.Name, Status: "error", Version: err.Error(), Path: toolPath}
	}

	output := strings.TrimSpace(stdout.String())
	errOutput := strings.TrimSpace(stderr.String())
	version := t.Parse(output, errOutput)
	if version == "" {
		version = output
	}
	return ToolInfo{Name: t.Name, Status: "installed", Version: version, Path: toolPath}
}

// firstLine returns the first non-empty line from text.
func firstLine(text string) string {
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			return line
		}
	}
	return strings.TrimSpace(text)
}

// ══════════════════════════════════════════════
//  Container Status
// ══════════════════════════════════════════════

// GetContainers returns Docker container status and summary.
func (d *DevOps) GetContainers() ContainerSummary {
	if _, err := exec.LookPath("docker"); err != nil {
		return ContainerSummary{Containers: []ContainerInfo{}}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "docker", "ps", "-a", "--format", `{{.ID}}\t{{.Names}}\t{{.Image}}\t{{.State}}\t{{.Status}}\t{{.Ports}}`)
	var stdout strings.Builder
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return ContainerSummary{Containers: []ContainerInfo{}}
	}

	var containers []ContainerInfo
	running, stopped, failed := 0, 0, 0

	for _, line := range strings.Split(stdout.String(), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 6)
		if len(parts) < 5 {
			continue
		}
		c := ContainerInfo{
			ID:     parts[0],
			Name:   parts[1],
			Image:  parts[2],
			State:  parts[3],
			Status: parts[4],
		}
		if len(parts) == 6 {
			c.Ports = parts[5]
		}
		containers = append(containers, c)

		state := strings.ToLower(c.State)
		switch state {
		case "running":
			running++
		case "exited":
			stopped++
			// Check for non-zero exit code in status string
			if strings.Contains(c.Status, "(0)") == false && strings.Contains(c.Status, "Exited") {
				failed++
			}
		default:
			stopped++
		}
	}

	return ContainerSummary{
		Running:    running,
		Stopped:    stopped,
		Failed:     failed,
		Total:      running + stopped,
		Containers: containers,
	}
}

// parseIntOr returns n or fallback if parsing fails.
func parseIntOr(s string, fallback int) int {
	n := 0
	if _, err := fmt.Sscanf(s, "%d", &n); err != nil {
		return fallback
	}
	return n
}

// RunDevOpsDiagnostics runs a health check on all dev tools.
func (d *DevOps) RunDevOpsDiagnostics() DevOpsDiagResult {
	result := devops.RunDevOpsDiagnostics()
	checks := make([]DevOpsDiagCheck, 0, len(result.Checks))
	for _, c := range result.Checks {
		checks = append(checks, DevOpsDiagCheck{Name: c.Name, Status: c.Status, Message: c.Message, Value: c.Value})
	}
	return DevOpsDiagResult{Checks: checks, Score: result.Score, Timestamp: result.Timestamp}
}

// ══════════════════════════════════════════════
//  Local Servers
// ══════════════════════════════════════════════

// detectFramework maps a process name to a known framework.
func detectFramework(processName string) string {
	lower := strings.ToLower(processName)
	switch {
	case strings.Contains(lower, "node"):
		return "Node.js"
	case strings.Contains(lower, "python"):
		return "Python"
	case strings.Contains(lower, "go") && !strings.Contains(lower, "google"):
		return "Go"
	case strings.Contains(lower, "dotnet"):
		return ".NET"
	case strings.Contains(lower, "java"):
		return "Java"
	case strings.Contains(lower, "docker"):
		return "Docker"
	case strings.Contains(lower, "redis"):
		return "Redis"
	case strings.Contains(lower, "nginx"):
		return "Nginx"
	case strings.Contains(lower, "apache") || strings.Contains(lower, "httpd"):
		return "Apache"
	default:
		return ""
	}
}

// healthCheckProbe attempts a quick HTTP GET to determine if a server is healthy.
func healthCheckProbe(port int) string {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("http://127.0.0.1:%d/", port), nil)
	if err != nil {
		return "unknown"
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "error"
		}
		return "unknown"
	}
	resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		return "healthy"
	}
	return "error"
}

// GetLocalServers returns locally listening servers with framework detection and health checks.
func (d *DevOps) GetLocalServers() []LocalServer {
	if runtime.GOOS == "windows" {
		return d.getLocalServersWindows()
	}
	return d.getLocalServersUnix()
}

// getLocalServersWindows uses netstat on Windows.
func (d *DevOps) getLocalServersWindows() []LocalServer {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "netstat", "-ano")
	var stdout strings.Builder
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return []LocalServer{}
	}

	servers := parseNetstatOutput(stdout.String())
	// Run health checks
	for i := range servers {
		servers[i].Health = healthCheckProbe(servers[i].Port)
	}
	return servers
}

// getLocalServersUnix uses ss on Linux/macOS.
func (d *DevOps) getLocalServersUnix() []LocalServer {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "ss", "-tlnp")
	var stdout strings.Builder
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return []LocalServer{}
	}

	servers := parseSSOutput(stdout.String())
	for i := range servers {
		servers[i].Health = healthCheckProbe(servers[i].Port)
	}
	return servers
}

// parseNetstatOutput parses Windows netstat -ano output for listening ports on localhost.
func parseNetstatOutput(output string) []LocalServer {
	servers := []LocalServer{}
	pidProcess := make(map[string]string)

	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Proto") || strings.HasPrefix(line, "Active") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 5 {
			continue
		}

		proto := strings.ToLower(parts[0])
		if proto != "tcp" && proto != "tcp4" && proto != "tcp6" {
			continue
		}

		// Check if state is LISTENING (Windows netstat has state in parts[3])
		if parts[1] != "LISTENING" && parts[2] != "LISTENING" {
			continue
		}

		// Find the local address field
		var localAddr string
		var pid string
		for i, p := range parts {
			if strings.Contains(p, "LISTENING") {
				// The address is typically the field before LISTENING
				if i > 0 {
					localAddr = parts[i-1]
				}
				if i+1 < len(parts) {
					pid = parts[i+1]
				}
				break
			}
		}
		if localAddr == "" {
			continue
		}

		// Parse port from address
		lastColon := strings.LastIndex(localAddr, ":")
		if lastColon < 0 {
			continue
		}
		port := parseIntOr(localAddr[lastColon+1:], 0)
		if port == 0 {
			continue
		}

		// Filter for localhost
		ip := localAddr[:lastColon]
		if ip != "127.0.0.1" && ip != "0.0.0.0" && ip != "[::1]" && ip != "::" && ip != "0:0:0:0" {
			continue
		}

		procName := ""
		if pid != "" && pid != "0" {
			if name, ok := pidProcess[pid]; ok {
				procName = name
			} else {
				procName = lookupProcessName(pid)
				pidProcess[pid] = procName
			}
		}

		servers = append(servers, LocalServer{
			Port:      port,
			Protocol:  proto,
			Process:   procName,
			PID:       parseIntOr(pid, 0),
			Framework: detectFramework(procName),
		})
	}
	return servers
}

// parseSSOutput parses Linux ss -tlnp output for listening ports.
func parseSSOutput(output string) []LocalServer {
	servers := []LocalServer{}

	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "State") || strings.HasPrefix(line, "LISTEN") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 4 {
			continue
		}

		// ss output: State  Recv-Q  Send-Q  Local Address:Port  Peer Address:Port  Process
		localAddr := ""
		for _, p := range parts {
			if strings.Contains(p, ":") && (strings.HasPrefix(p, "127.") || strings.HasPrefix(p, "0.0.0.0") || strings.HasPrefix(p, "[::1]") || p == "*:*") {
				localAddr = p
				break
			}
		}
		if localAddr == "" {
			continue
		}

		lastColon := strings.LastIndex(localAddr, ":")
		if lastColon < 0 {
			continue
		}
		port := parseIntOr(localAddr[lastColon+1:], 0)
		if port == 0 {
			continue
		}

		// Extract process name from the rest of the line
		procName := ""
		for _, p := range parts {
			if strings.HasPrefix(p, "users:") {
				// Format: users:(("process",pid=1234,fd=5))
				start := strings.Index(p, "(")
				end := strings.Index(p, ",")
				if start >= 0 && end > start {
					procName = p[start+1 : end]
				}
			}
		}

		servers = append(servers, LocalServer{
			Port:      port,
			Protocol:  "tcp",
			Process:   procName,
			Framework: detectFramework(procName),
		})
	}
	return servers
}

// lookupProcessName returns the process name for a given PID string (Windows).
func lookupProcessName(pid string) string {
	if runtime.GOOS != "windows" {
		return ""
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "tasklist", "/fi", "PID eq "+pid, "/fo", "csv", "/nh")
	var stdout strings.Builder
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return ""
	}
	// CSV output: "taskname.exe","PID","Session","Session#","Mem"
	line := strings.TrimSpace(stdout.String())
	if line == "" {
		return ""
	}
	fields := strings.Split(line, ",")
	if len(fields) >= 1 {
		name := strings.Trim(fields[0], "\" ")
		return name
	}
	return ""
}

// ══════════════════════════════════════════════
//  Environment Panel
// ══════════════════════════════════════════════

// GetEnvironment returns environment variables, SDKs, and package managers.
func (d *DevOps) GetEnvironment() EnvironmentInfo {
	// PATH directories (first 20)
	pathRaw := os.Getenv("PATH")
	pathDirs := strings.Split(pathRaw, string(os.PathListSeparator))
	if len(pathDirs) > 20 {
		pathDirs = pathDirs[:20]
	}

	// Key environment variables
	keyNames := []string{"GOPATH", "GOROOT", "JAVA_HOME", "PYTHONPATH", "NODE_PATH", "HOME", "USER", "SHELL", "OLLAMA_MODEL"}
	var keyVars []EnvVarInfo
	for _, name := range keyNames {
		val := os.Getenv(name)
		if val != "" {
			keyVars = append(keyVars, EnvVarInfo{Name: name, Value: val})
		}
	}

	// SDKs
	sdkSpecs := []toolSpec{
		{Name: "Go", Command: "go", Args: []string{"version"}, Parse: func(out, _ string) string {
			fields := strings.Fields(firstLine(out))
			if len(fields) >= 3 {
				return fields[2]
			}
			return firstLine(out)
		}},
		{Name: "Java", Command: "java", Args: []string{"-version"}, Parse: func(out, stderr string) string {
			line := firstLine(stderr)
			if line == "" {
				line = firstLine(out)
			}
			return line
		}},
		{Name: ".NET", Command: "dotnet", Args: []string{"--version"}, Parse: func(out, _ string) string {
			return firstLine(out)
		}},
		{Name: "Rust", Command: "rustc", Args: []string{"--version"}, Parse: func(out, _ string) string {
			fields := strings.Fields(firstLine(out))
			if len(fields) >= 2 {
				return fields[1]
			}
			return firstLine(out)
		}},
	}
	var sdks []ToolVersion
	for _, spec := range sdkSpecs {
		td := detectTool(spec)
		if td.Status == "installed" {
			sdks = append(sdks, ToolVersion{Name: td.Name, Version: td.Version})
		}
	}

	// Package managers
	pmSpecs := []toolSpec{
		{Name: "npm", Command: "npm", Args: []string{"--version"}, Parse: func(out, _ string) string {
			return firstLine(out)
		}},
		{Name: "pip", Command: "pip", Args: []string{"--version"}, Parse: func(out, _ string) string {
			fields := strings.Fields(firstLine(out))
			if len(fields) >= 2 {
				return fields[1]
			}
			return firstLine(out)
		}},
		{Name: "cargo", Command: "cargo", Args: []string{"--version"}, Parse: func(out, _ string) string {
			fields := strings.Fields(firstLine(out))
			if len(fields) >= 2 {
				return fields[1]
			}
			return firstLine(out)
		}},
		{Name: "go", Command: "go", Args: []string{"version"}, Parse: func(out, _ string) string {
			fields := strings.Fields(firstLine(out))
			if len(fields) >= 3 {
				return fields[2]
			}
			return firstLine(out)
		}},
	}
	var pkgMgrs []ToolVersion
	for _, spec := range pmSpecs {
		td := detectTool(spec)
		if td.Status == "installed" {
			pkgMgrs = append(pkgMgrs, ToolVersion{Name: td.Name, Version: td.Version})
		}
	}

	return EnvironmentInfo{
		PathDirs:        pathDirs,
		KeyVars:         keyVars,
		SDKs:            sdks,
		PackageManagers: pkgMgrs,
	}
}

// ══════════════════════════════════════════════
//  AI Suggestions
// ══════════════════════════════════════════════

// severityRank returns a numeric rank for sorting (higher = more severe).
func severityRank(s string) int {
	switch s {
	case "critical":
		return 3
	case "warning":
		return 2
	case "info":
		return 1
	default:
		return 0
	}
}

// GetAISuggestions synthesizes Docker, Git, and tool data into actionable suggestions.
func (d *DevOps) GetAISuggestions() []DevOpsSuggestion {
	var suggestions []DevOpsSuggestion

	// ── Docker checks ──
	containers := d.GetContainers()
	if containers.Total > 0 {
		if containers.Failed > 0 {
			suggestions = append(suggestions, DevOpsSuggestion{
				Category: "docker",
				Severity: "critical",
				Message:  fmt.Sprintf("%d container(s) in failed state", containers.Failed),
				Action:   "Check container logs with 'docker logs <name>' and restart failed containers.",
			})
		}
		if containers.Stopped > 0 && containers.Failed == 0 {
			suggestions = append(suggestions, DevOpsSuggestion{
				Category: "docker",
				Severity: "warning",
				Message:  fmt.Sprintf("%d container(s) stopped (not running)", containers.Stopped),
				Action:   "Review if stopped containers should be running. Start with 'docker start <name>'.",
			})
		}
	} else {
		// Docker binary not found or no containers
		tools := d.GetInstalledTools()
		for _, t := range tools {
			if t.Name == "Docker" && t.Status == "not-found" {
				suggestions = append(suggestions, DevOpsSuggestion{
					Category: "docker",
					Severity: "info",
					Message:  "Docker is not installed on this system",
					Action:   "Install Docker Desktop or Docker Engine for container management.",
				})
			}
		}
	}

	// ── Git checks ──
	// Git summary logic removed in pruning.

	// ── Node.js checks ──
	tools := d.GetInstalledTools()
	for _, t := range tools {
		if t.Name == "Node.js" {
			if t.Status == "not-found" {
				suggestions = append(suggestions, DevOpsSuggestion{
					Category: "node",
					Severity: "warning",
					Message:  "Node.js is not installed",
					Action:   "Install Node.js from https://nodejs.org/ for frontend builds and tooling.",
				})
			} else if t.Status == "installed" {
				// Parse major version from e.g. "v20.12.0"
				ver := strings.TrimPrefix(t.Version, "v")
				parts := strings.Split(ver, ".")
				if len(parts) >= 1 {
					major := 0
					fmt.Sscanf(parts[0], "%d", &major)
					if major > 0 && major < 18 {
						suggestions = append(suggestions, DevOpsSuggestion{
							Category: "node",
							Severity: "warning",
							Message:  fmt.Sprintf("Node.js v%d is outdated (EOL). Current LTS is v22+.", major),
							Action:   "Upgrade to the latest LTS release from https://nodejs.org/.",
						})
					}
				}
			}
		}
	}

	// ── General tool checks ──
	for _, t := range tools {
		if t.Status == "error" {
			suggestions = append(suggestions, DevOpsSuggestion{
				Category: "general",
				Severity: "info",
				Message:  fmt.Sprintf("Tool '%s' returned an error: %s", t.Name, t.Version),
				Action:   fmt.Sprintf("Reinstall or update '%s' and verify it works from the command line.", t.Name),
			})
		}
	}

	// ── Sort by severity (critical first) ──
	sort.Slice(suggestions, func(i, j int) bool {
		return severityRank(suggestions[i].Severity) > severityRank(suggestions[j].Severity)
	})

	// If nothing found, add a healthy status
	if len(suggestions) == 0 {
		suggestions = append(suggestions, DevOpsSuggestion{
			Category: "general",
			Severity: "info",
			Message:  "All checks passed — no issues detected",
			Action:   "Everything looks good!",
		})
	}

	return suggestions
}

// ══════════════════════════════════════════════
//  Docker Status
// ══════════════════════════════════════════════

// GetDockerStatus checks Docker installation, daemon status, and container counts.
func (d *DevOps) GetDockerStatus() DockerStatus {
	var status DockerStatus

	// 1. Check if docker CLI exists
	dockerPath, err := exec.LookPath("docker")
	if err != nil {
		return status // Installed=false by default
	}
	status.Installed = true

	// 2. Get Docker version
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	verCmd := exec.CommandContext(ctx, dockerPath, "--version")
	var verOut strings.Builder
	verCmd.Stdout = &verOut
	if err := verCmd.Run(); err == nil {
		// Output: "Docker version 27.5.1, build a1678b9"
		verLine := firstLine(verOut.String())
		fields := strings.Fields(verLine)
		if len(fields) >= 3 {
			status.Version = strings.TrimRight(fields[2], ",")
		} else {
			status.Version = verLine
		}
	}

	// 3. Check if Docker daemon is running via docker info
	ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()

	infoCmd := exec.CommandContext(ctx2, dockerPath, "info", "--format", `{{.ServerVersion}}`)
	var infoOut strings.Builder
	infoCmd.Stdout = &infoOut
	if err := infoCmd.Run(); err == nil {
		status.Running = true
		ver := strings.TrimSpace(infoOut.String())
		if ver != "" && status.Version == "" {
			status.Version = ver
		}
	}

	// 4. Count containers by status
	ctx3, cancel3 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel3()

	psCmd := exec.CommandContext(ctx3, dockerPath, "ps", "-a", "--format", `{{.Status}}`)
	var psOut strings.Builder
	psCmd.Stdout = &psOut
	if err := psCmd.Run(); err == nil {
		for _, line := range strings.Split(psOut.String(), "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			status.Containers.Total++
			lower := strings.ToLower(line)
			if strings.HasPrefix(lower, "up") {
				status.Containers.Running++
			} else if strings.HasPrefix(lower, "restarting") {
				status.Containers.Failed++
			} else {
				status.Containers.Stopped++
			}
		}
	}

	return status
}

// ══════════════════════════════════════════════
//  Kubernetes Status
// ══════════════════════════════════════════════

// GetKubernetesStatus checks kubectl availability and cluster connectivity.
func (d *DevOps) GetKubernetesStatus() KubernetesStatus {
	var status KubernetesStatus

	// 1. Check if kubectl exists
	kubectlPath, err := exec.LookPath("kubectl")
	if err != nil {
		return status // Installed=false, Connected=false by default
	}
	status.Installed = true

	// 2. Check cluster connectivity and parse cluster name
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	clusterCmd := exec.CommandContext(ctx, kubectlPath, "cluster-info")
	var clusterOut strings.Builder
	clusterCmd.Stdout = &clusterOut
	var clusterErr strings.Builder
	clusterCmd.Stderr = &clusterErr
	if err := clusterCmd.Run(); err == nil {
		status.Connected = true
		// Output line: "Kubernetes control plane is running at https://..."
		for _, line := range strings.Split(clusterOut.String(), "\n") {
			line = strings.TrimSpace(line)
			if strings.Contains(line, "control plane") || strings.Contains(line, "Kubernetes master") {
				// Extract host from URL
				if idx := strings.Index(line, "http"); idx >= 0 {
					url := line[idx:]
					// Parse //host:port
					if start := strings.Index(url, "//"); start >= 0 {
						host := url[start+2:]
						if end := strings.Index(host, ":"); end >= 0 {
							host = host[:end]
						}
						if end := strings.Index(host, "/"); end >= 0 {
							host = host[:end]
						}
						status.Cluster = host
					}
				}
				break
			}
		}
	}

	if !status.Connected {
		return status
	}

	// 3. Count nodes
	ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()

	nodeCmd := exec.CommandContext(ctx2, kubectlPath, "get", "nodes", "--no-headers")
	var nodeOut strings.Builder
	nodeCmd.Stdout = &nodeOut
	if err := nodeCmd.Run(); err == nil {
		for _, line := range strings.Split(nodeOut.String(), "\n") {
			if strings.TrimSpace(line) != "" {
				status.Nodes++
			}
		}
	}

	// 4. Count pods across all namespaces
	ctx3, cancel3 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel3()

	podCmd := exec.CommandContext(ctx3, kubectlPath, "get", "pods", "--all-namespaces", "--no-headers")
	var podOut strings.Builder
	podCmd.Stdout = &podOut
	if err := podCmd.Run(); err == nil {
		for _, line := range strings.Split(podOut.String(), "\n") {
			if strings.TrimSpace(line) != "" {
				status.Pods++
			}
		}
	}

	return status
}

// ══════════════════════════════════════════════
//  Service Categorization
// ══════════════════════════════════════════════

// categorizeService maps a service name to its category.
func categorizeService(name string) string {
	lower := strings.ToLower(name)

	databases := []string{"mysql", "postgres", "postgresql", "sqlite", "mongodb", "mongo",
		"redis", "mariadb", "sqlserver", "mssql", "oracle", "couchdb", "cassandra", "memcached"}
	messageQueues := []string{"rabbitmq", "nats", "kafka", "zeromq", "activemq", "mosquitto"}
	webServers := []string{"nginx", "apache", "httpd", "iis", "caddy", "traefik", "haproxy"}
	containers := []string{"docker", "podman", "containerd", "containerd-shim", "docker-proxy"}

	for _, db := range databases {
		if strings.Contains(lower, db) {
			return "databases"
		}
	}
	for _, mq := range messageQueues {
		if strings.Contains(lower, mq) {
			return "message-queues"
		}
	}
	for _, ws := range webServers {
		if strings.Contains(lower, ws) {
			return "web-servers"
		}
	}
	for _, ct := range containers {
		if strings.Contains(lower, ct) {
			return "containers"
		}
	}

	return "other"
}

// normalizeServiceStatus converts various status strings to a standard form.
func normalizeServiceStatus(status string) string {
	lower := strings.ToLower(status)
	if strings.Contains(lower, "running") {
		return "running"
	}
	if strings.Contains(lower, "stopped") || strings.Contains(lower, "disabled") {
		return "stopped"
	}
	return "unknown"
}

// GetServiceCategories returns services grouped by function type.
func (d *DevOps) GetServiceCategories() []ServiceCategory {
	services := d.GetServices()
	if len(services) == 0 {
		return []ServiceCategory{}
	}

	// Bucket services by category
	buckets := make(map[string][]ServiceInfo)
	for _, s := range services {
		cat := categorizeService(s.Name)
		si := ServiceInfo{
			Name:   s.Name,
			Status: normalizeServiceStatus(s.Status),
		}
		buckets[cat] = append(buckets[cat], si)
	}

	// Build output in consistent order
	order := []string{"databases", "message-queues", "web-servers", "containers", "other"}
	var categories []ServiceCategory
	for _, cat := range order {
		svcs, ok := buckets[cat]
		if !ok || len(svcs) == 0 {
			continue
		}
		categories = append(categories, ServiceCategory{
			Category: cat,
			Services: svcs,
		})
	}

	return categories
}

// GetServiceGroupSummary returns aggregated service counts by category.
func (d *DevOps) GetServiceGroupSummary() ServiceGroupSummary {
	cats := d.GetServiceCategories()
	var summary ServiceGroupSummary

	for _, cat := range cats {
		count := len(cat.Services)
		for _, svc := range cat.Services {
			if svc.Status == "running" {
				summary.Running++
			} else {
				summary.Stopped++
			}
		}
		switch cat.Category {
		case "databases":
			summary.Databases = count
		case "message-queues":
			summary.MessageQueues = count
		case "web-servers":
			summary.WebServers = count
		case "containers":
			summary.Containers = count
		default:
			summary.Other = count
		}
	}

	return summary
}

