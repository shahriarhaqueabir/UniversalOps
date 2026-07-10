package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
	"github.com/shahriarhaqueabir/AllOpsFull/internal/devops"
	sysopsPkg "github.com/shahriarhaqueabir/AllOpsFull/internal/sysops"
)

// DevOps exposes development operations bindings to the frontend.
type DevOps struct {
	app *App
}

// NewDevOps creates a new DevOps facade.
func NewDevOps(app *App) *DevOps {
	return &DevOps{app: app}
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
		for line := range lineCh {
			wailsruntime.EventsEmit(d.app.ctx, EventCmdLine, map[string]string{
				"id":   id,
				"line": line,
			})
		}
		wailsruntime.EventsEmit(d.app.ctx, EventCmdDone, id)
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

// ListDirectory lists the contents of a directory.
func (d *DevOps) ListDirectory(path string) []FileEntry {
	entries, err := devops.ListDir(path)
	if err != nil {
		common.LogWarn("ListDirectory failed: %v", err)
		return nil
	}
	out := make([]FileEntry, 0, len(entries))
	for _, e := range entries {
		out = append(out, FileEntry{
			Name:     e.Name,
			Path:     e.Path,
			Size:     e.Size,
			RawSize:  e.RawSize,
			IsDir:    e.IsDir,
			IsBinary: e.IsBinary,
			Mode:     e.Mode.String(),
			ModTime:  e.ModTime.Format(time.RFC3339),
		})
	}
	return out
}

// ReadFile reads the contents of a file.
func (d *DevOps) ReadFile(path string) string {
	content, err := devops.ReadFile(path)
	if err != nil {
		common.LogWarn("ReadFile failed: %v", err)
		return ""
	}
	return content
}

// WriteFile writes the contents to a file.
func (d *DevOps) WriteFile(path string, data string) bool {
	err := devops.WriteFile(path, data)
	if err != nil {
		common.LogWarn("WriteFile failed: %v", err)
		return false
	}
	return true
}

// GetDevProcesses returns all processes by CPU usage (from system ops).
func (d *DevOps) GetDevProcesses() []ProcessInfo {
	procs, err := sysopsPkg.GetTopProcesses(100)
	if err != nil {
		common.LogWarn("GetDevProcesses failed: %v", err)
		return nil
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

// TailLog reads the last n lines from a file.
func (d *DevOps) TailLog(path string, n int) []string {
	lines, err := devops.TailLog(path, n)
	if err != nil {
		common.LogWarn("TailLog failed: %v", err)
		return nil
	}
	return lines
}

// SearchLog searches a file for lines containing the pattern.
func (d *DevOps) SearchLog(path string, pattern string) []string {
	lines, err := devops.SearchLog(path, pattern)
	if err != nil {
		common.LogWarn("SearchLog failed: %v", err)
		return nil
	}
	return lines
}

// GetServices returns a list of system services.
func (d *DevOps) GetServices() []ServiceEntry {
	services, err := devops.ListServices(0)
	if err != nil {
		common.LogWarn("GetServices failed: %v", err)
		return nil
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

	d.app.eventBus.Emit(common.NewEvent(
		common.CatDevOps,
		common.EventInfo,
		"devops",
		"Service state changed",
		fmt.Sprintf("Service '%s' %s", name, action),
	))
	return true
}

// RunPowerShell executes a PowerShell command using the HawkwardHybrid profile.
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

// GetDefaultPath returns a safe default directory for file browsing.
func (d *DevOps) GetDefaultPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return os.TempDir()
	}
	return home
}

// GetPowerShellWorkflows returns a list of available PowerShell diagnostic workflows.
func (d *DevOps) GetPowerShellWorkflows() []string {
	return []string{
		"Invoke-HawkDailyOps",
		"Invoke-HawkSystemReview",
		"Invoke-HawkSecurityAudit",
		"Invoke-HawkNetworkDiagnostics",
		"Invoke-HawkThreatHunt",
		"Invoke-HawkChangeAudit",
		"Invoke-HawkComplianceCheck",
	}
}

// sanitizeError strips internal details from errors sent to the frontend.
func sanitizeError(err error) string {
	if errors.Is(err, devops.ErrDangerousCommand) {
		return "command rejected by security policy"
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
		return ContainerSummary{}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "docker", "ps", "-a", "--format", `{{.ID}}\t{{.Names}}\t{{.Image}}\t{{.State}}\t{{.Status}}\t{{.Ports}}`)
	var stdout strings.Builder
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return ContainerSummary{}
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

// ══════════════════════════════════════════════
//  Git Summary
// ══════════════════════════════════════════════

// findGitRepos discovers git repositories in common locations.
func findGitRepos(maxRepos int) []string {
	var candidates []string

	// Current working directory
	if cwd, err := os.Getwd(); err == nil {
		candidates = append(candidates, cwd)
	}

	// Home directory and common subdirectories
	home, err := os.UserHomeDir()
	if err == nil {
		candidates = append(candidates, home)
		for _, sub := range []string{"Documents", "Projects", "src", "dev", "code", "repos", "GitHub", "git"} {
			cand := filepath.Join(home, sub)
			if info, err := os.Stat(cand); err == nil && info.IsDir() {
				candidates = append(candidates, cand)
			}
		}
	}

	var found []string
	seen := make(map[string]bool)

	for _, dir := range candidates {
		// Check if dir itself is a git repo
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			if !seen[dir] {
				seen[dir] = true
				found = append(found, dir)
			}
			if len(found) >= maxRepos {
				return found
			}
		}

		// Scan one level deep for git repos
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			path := filepath.Join(dir, entry.Name())
			if _, err := os.Stat(filepath.Join(path, ".git")); err == nil {
				if !seen[path] {
					seen[path] = true
					found = append(found, path)
				}
				if len(found) >= maxRepos {
					return found
				}
			}
		}
	}
	return found
}

// gitRun runs a git command in the given directory with a timeout.
func gitRun(dir string, args ...string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	var stdout strings.Builder
	cmd.Stdout = &stdout
	_ = cmd.Run()
	return strings.TrimSpace(stdout.String())
}

// parseIntOr returns n or fallback if parsing fails.
func parseIntOr(s string, fallback int) int {
	n := 0
	if _, err := fmt.Sscanf(s, "%d", &n); err != nil {
		return fallback
	}
	return n
}

// GetGitSummary returns aggregated git repository status.
func (d *DevOps) GetGitSummary() GitSummary {
	if _, err := exec.LookPath("git"); err != nil {
		return GitSummary{}
	}

	paths := findGitRepos(10)
	var repos []GitRepoInfo

	for _, dir := range paths {
		branch := gitRun(dir, "rev-parse", "--abbrev-ref", "HEAD")
		if branch == "" {
			continue
		}

		statusPorcelain := gitRun(dir, "status", "--porcelain")
		statusUntracked := gitRun(dir, "status", "--porcelain", "-u")

		modified := 0
		for _, line := range strings.Split(statusPorcelain, "\n") {
			line = strings.TrimSpace(line)
			if line != "" && !strings.HasPrefix(line, "?") {
				modified++
			}
		}

		untracked := 0
		for _, line := range strings.Split(statusUntracked, "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "?") {
				untracked++
			}
		}

		ahead, behind := 0, 0
		upstream := gitRun(dir, "rev-parse", "--abbrev-ref", "@{upstream}")
		if upstream != "" {
			counts := gitRun(dir, "rev-list", "--left-right", "--count", "HEAD..."+upstream)
			if counts != "" {
				fmt.Sscanf(counts, "%d %d", &ahead, &behind)
			}
		}

		repos = append(repos, GitRepoInfo{
			Path:           dir,
			Branch:         branch,
			ModifiedFiles:  modified,
			UntrackedFiles: untracked,
			Ahead:          ahead,
			Behind:         behind,
			Clean:          modified == 0 && untracked == 0,
		})
	}

	return GitSummary{
		Repositories: repos,
		TotalRepos:   len(repos),
	}
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
		return nil
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
		return nil
	}

	servers := parseSSOutput(stdout.String())
	for i := range servers {
		servers[i].Health = healthCheckProbe(servers[i].Port)
	}
	return servers
}

// parseNetstatOutput parses Windows netstat -ano output for listening ports on localhost.
func parseNetstatOutput(output string) []LocalServer {
	var servers []LocalServer
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
	var servers []LocalServer

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
