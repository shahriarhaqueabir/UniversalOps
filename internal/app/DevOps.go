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

	"github.com/shahriarhaqueabir/UniversalOps/internal/common"
	"github.com/shahriarhaqueabir/UniversalOps/internal/devops"
	sysopsPkg "github.com/shahriarhaqueabir/UniversalOps/internal/sysops"
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
	defer common.RecoverPanic()
	start := time.Now()
	result, err := devops.RunCommand(cmd)
	dur := time.Since(start).Milliseconds()
	if err != nil {
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
		if d.ctx == nil {
			for range lineCh {
			}
			return
		}
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
	defer common.RecoverPanic()
	procs, err := sysopsPkg.GetTopProcesses(100)
	if err != nil {
		return []ProcessInfo{}
	}
	out := make([]ProcessInfo, 0, len(procs))
	for _, p := range procs {
		out = append(out, ProcessInfo{
			PID:       p.PID,
			Name:      p.Name,
			CPU:       p.CPU,
			Memory:    p.Memory,
			MemPct:    p.MemPct,
			Status:    p.Status,
			NumFDs:    p.NumFDs,
			IsSigned:  p.IsSigned,
			Publisher: p.Publisher,
		})
	}
	return out
}

// KillProcess terminates a process by PID.
func (d *DevOps) KillProcess(pid int) CommandResult {
	defer common.RecoverPanic()
	err := devops.KillProcess(int32(pid))
	if err != nil {
		return CommandResult{Error: err.Error()}
	}
	return CommandResult{Output: "Process terminated successfully"}
}

// GetServices returns a list of system services.
func (d *DevOps) GetServices() []ServiceEntry {
	defer common.RecoverPanic()
	services, err := devops.ListServices(0)
	if err != nil {
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
	defer common.RecoverPanic()
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

// RunPowerShellLive executes a PowerShell command and streams output.
// Enforces safety policies before execution.
func (d *DevOps) RunPowerShellLive(cmd string, id string) CommandResult {
	if devops.ContainsShellMetachar(cmd) {
		return CommandResult{Command: cmd, Error: sanitizeError(devops.ErrShellMetachar)}
	}
	if devops.IsDangerousCommand(cmd) {
		return CommandResult{Command: cmd, Error: sanitizeError(devops.ErrDangerousCommand)}
	}

	start := time.Now()
	lineCh := make(chan string, 100)

	go func() {
		defer common.RecoverPanic()
		if d.ctx == nil {
			for range lineCh {
			}
			return
		}
		for line := range lineCh {
			wailsruntime.EventsEmit(d.ctx, EventCmdLine, map[string]string{
				"id":   id,
				"line": line,
			})
		}
		wailsruntime.EventsEmit(d.ctx, EventCmdDone, id)
	}()

	result, err := devops.RunPowerShellWithLiveOutput(cmd, lineCh)
	dur := time.Since(start).Milliseconds()
	if err != nil {
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

// RunGitBashLive executes a command via Git Bash and streams output.
// Enforces safety policies before execution.
func (d *DevOps) RunGitBashLive(cmd string, id string) CommandResult {
	if devops.ContainsShellMetachar(cmd) {
		return CommandResult{Command: cmd, Error: sanitizeError(devops.ErrShellMetachar)}
	}
	if devops.IsDangerousCommand(cmd) {
		return CommandResult{Command: cmd, Error: sanitizeError(devops.ErrDangerousCommand)}
	}

	start := time.Now()
	lineCh := make(chan string, 100)

	go func() {
		defer common.RecoverPanic()
		if d.ctx == nil {
			for range lineCh {
			}
			return
		}
		for line := range lineCh {
			wailsruntime.EventsEmit(d.ctx, EventCmdLine, map[string]string{
				"id":   id,
				"line": line,
			})
		}
		wailsruntime.EventsEmit(d.ctx, EventCmdDone, id)
	}()

	result, err := devops.RunGitBashWithLiveOutput(cmd, lineCh)
	dur := time.Since(start).Milliseconds()
	if err != nil {
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
	defer common.RecoverPanic()
	root, err := os.Getwd()
	if err != nil {
		return "."
	}
	return root
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
	return err.Error()
}

// toolSpec defines a tool to detect.
type toolSpec struct {
	Name    string
	Command string
	Args    []string
	Parse   func(output string, stderr string) string
}

// detectTool runs a single tool's version command and returns a ToolInfo.
func detectTool(t toolSpec) ToolInfo {
	toolPath, err := exec.LookPath(t.Command)
	if err != nil {
		return ToolInfo{Name: t.Name, Status: "not-found"}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	cmd := common.HiddenCommandContext(ctx, t.Command, t.Args...)
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
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

// GetInstalledTools detects installed development tools and their versions.
func (d *DevOps) GetInstalledTools() []ToolInfo {
	defer common.RecoverPanic()
	tools := []toolSpec{
		{
			Name:    "Git",
			Command: "git",
			Args:    []string{"--version"},
			Parse: func(out, _ string) string {
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
				return firstLine(out)
			},
		},
		{
			Name:    "Go",
			Command: "go",
			Args:    []string{"version"},
			Parse: func(out, _ string) string {
				fields := strings.Fields(firstLine(out))
				if len(fields) >= 3 {
					return fields[2]
				}
				return firstLine(out)
			},
		},
		{
			Name:    "Python",
			Command: "python",
			Args:    []string{"--version"},
			Parse: func(out, _ string) string {
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

// parseIntOr returns n or fallback if parsing fails.
func parseIntOr(s string, fallback int) int {
	n := 0
	if _, err := fmt.Sscanf(s, "%d", &n); err != nil {
		return fallback
	}
	return n
}

// GetContainers returns Docker container status and summary.
func (d *DevOps) GetContainers() ContainerSummary {
	defer common.RecoverPanic()
	res, err := devops.GetContainers()
	if err != nil {
		return ContainerSummary{Containers: []ContainerInfo{}}
	}

	containers := make([]ContainerInfo, len(res.Containers))
	for i, c := range res.Containers {
		containers[i] = ContainerInfo{
			ID:     c.ID,
			Name:   c.Name,
			Image:  c.Image,
			State:  c.State,
			Status: c.Status,
			Ports:  c.Ports,
		}
	}

	return ContainerSummary{
		Running:    res.Running,
		Stopped:    res.Stopped,
		Failed:     res.Failed,
		Total:      res.Total,
		Containers: containers,
	}
}

// GetContainersDomain returns domain-level container summary for MCP.
func (d *DevOps) GetContainersDomain() devops.ContainerSummary {
	res, _ := devops.GetContainers()
	return res
}

// RunDevOpsDiagnostics runs a health check on all dev tools.
func (d *DevOps) RunDevOpsDiagnostics() DevOpsDiagResult {
	defer common.RecoverPanic()
	result := devops.RunDevOpsDiagnostics()
	checks := make([]DevOpsDiagCheck, 0, len(result.Checks))
	for _, c := range result.Checks {
		checks = append(checks, DevOpsDiagCheck{Name: c.Name, Status: c.Status, Message: c.Message, Value: c.Value})
	}
	return DevOpsDiagResult{Checks: checks, Score: result.Score, Timestamp: result.Timestamp}
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

		if parts[1] != "LISTENING" && parts[2] != "LISTENING" {
			continue
		}

		var localAddr string
		var pid string
		for i, p := range parts {
			if strings.Contains(p, "LISTENING") {
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

		lastColon := strings.LastIndex(localAddr, ":")
		if lastColon < 0 {
			continue
		}
		port := parseIntOr(localAddr[lastColon+1:], 0)
		if port == 0 {
			continue
		}

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

		procName := ""
		for _, p := range parts {
			if strings.HasPrefix(p, "users:") {
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

// categorizeService maps a service name to its category.
func categorizeService(name string) string {
	lower := strings.ToLower(name)

	databases := []string{"mysql", "postgres", "postgresql", "sqlite", "mongodb", "mongo",
		"redis", "mariadb", "sqlserver", "mssql", "oracle", "couchdb", "cassandra", "memcached"}
	messageQueues := []string{"rabbitmq", "nats", "kafka", "zeromq", "activemq", "mosquito"}
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

// lookupProcessName returns the process name for a given PID string (Windows).
func lookupProcessName(pid string) string {
	if runtime.GOOS != "windows" {
		return ""
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cmd := common.HiddenCommandContext(ctx, "tasklist", "/fi", "PID eq "+pid, "/fo", "csv", "/nh")
	var stdout strings.Builder
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return ""
	}
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

// GetLocalServers returns locally listening servers with framework detection and health checks.
func (d *DevOps) GetLocalServers() []LocalServer {
	defer common.RecoverPanic()
	if runtime.GOOS == "windows" {
		return d.getLocalServersWindows()
	}
	return d.getLocalServersUnix()
}

func (d *DevOps) getLocalServersWindows() []LocalServer {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := common.HiddenCommandContext(ctx, "netstat", "-ano")
	var stdout strings.Builder
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return []LocalServer{}
	}

	servers := parseNetstatOutput(stdout.String())
	for i := range servers {
		servers[i].Health = healthCheckProbe(servers[i].Port)
	}
	return servers
}

func (d *DevOps) getLocalServersUnix() []LocalServer {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := common.HiddenCommandContext(ctx, "ss", "-tlnp")
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

// GetEnvironment returns environment variables, SDKs, and package managers.
func (d *DevOps) GetEnvironment() EnvironmentInfo {
	defer common.RecoverPanic()
	pathRaw := os.Getenv("PATH")
	pathDirs := strings.Split(pathRaw, string(os.PathListSeparator))
	if len(pathDirs) > 20 {
		pathDirs = pathDirs[:20]
	}

	keyNames := []string{"GOPATH", "GOROOT", "JAVA_HOME", "PYTHONPATH", "NODE_PATH", "HOME", "USER", "SHELL", "OLLAMA_MODEL"}
	var keyVars []EnvVarInfo
	for _, name := range keyNames {
		val := os.Getenv(name)
		if val != "" {
			keyVars = append(keyVars, EnvVarInfo{Name: name, Value: val})
		}
	}

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

// GetAISuggestions synthesizes Docker, Git, and tool data into actionable suggestions.
func (d *DevOps) GetAISuggestions() []DevOpsSuggestion {
	defer common.RecoverPanic()
	var suggestions []DevOpsSuggestion

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
		if t.Status == "error" {
			suggestions = append(suggestions, DevOpsSuggestion{
				Category: "general",
				Severity: "info",
				Message:  fmt.Sprintf("Tool '%s' returned an error: %s", t.Name, t.Version),
				Action:   fmt.Sprintf("Reinstall or update '%s' and verify it works from the command line.", t.Name),
			})
		}
	}

	sort.Slice(suggestions, func(i, j int) bool {
		return severityRank(suggestions[i].Severity) > severityRank(suggestions[j].Severity)
	})

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

// GetDockerStatus checks Docker installation, daemon status, and container counts.
func (d *DevOps) GetDockerStatus() DockerStatus {
	defer common.RecoverPanic()
	res, _ := devops.GetDockerStatus()

	summary := ContainerSummary{
		Running: res.Summary.Running,
		Stopped: res.Summary.Stopped,
		Failed:  res.Summary.Failed,
		Total:   res.Summary.Total,
	}

	containers := make([]ContainerInfo, len(res.Summary.Containers))
	for i, c := range res.Summary.Containers {
		containers[i] = ContainerInfo{
			ID:     c.ID,
			Name:   c.Name,
			Image:  c.Image,
			State:  c.State,
			Status: c.Status,
			Ports:  c.Ports,
		}
	}
	summary.Containers = containers

	return DockerStatus{
		Installed: res.Installed,
		Running:   res.Running,
		Version:   res.Version,
		Containers: summary,
	}
}

// GetKubernetesStatus checks kubectl availability and cluster connectivity.
func (d *DevOps) GetKubernetesStatus() KubernetesStatus {
	defer common.RecoverPanic()
	res, _ := devops.GetKubernetesStatus()

	return KubernetesStatus{
		Installed: res.Installed,
		Connected: res.Connected,
		Cluster:   res.Cluster,
		Nodes:     res.Nodes,
		Pods:      res.Pods,
	}
}

// GetKubernetesStatusDomain returns domain-level K8s status for MCP.
func (d *DevOps) GetKubernetesStatusDomain() devops.KubernetesStatus {
	res, _ := devops.GetKubernetesStatus()
	return res
}

// GetServiceCategories returns services grouped by function type.
func (d *DevOps) GetServiceCategories() []ServiceCategory {
	defer common.RecoverPanic()
	services := d.GetServices()
	if len(services) == 0 {
		return []ServiceCategory{}
	}

	buckets := make(map[string][]ServiceInfo)
	for _, s := range services {
		cat := categorizeService(s.Name)
		si := ServiceInfo{
			Name:   s.Name,
			Status: normalizeServiceStatus(s.Status),
		}
		buckets[cat] = append(buckets[cat], si)
	}

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
	defer common.RecoverPanic()
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

// GetDockerStats returns real-time stats for all running containers.
func (d *DevOps) GetDockerStats() []DockerStatsEntry {
	defer common.RecoverPanic()
	entries, err := devops.GetDockerStats()
	if err != nil {
		return []DockerStatsEntry{}
	}
	out := make([]DockerStatsEntry, 0, len(entries))
	for _, e := range entries {
		out = append(out, DockerStatsEntry{
			ContainerID:   e.ContainerID,
			Name:          e.Name,
			CPUPercent:    e.CPUPercent,
			MemoryUsage:   e.MemoryUsage,
			MemoryLimit:   e.MemoryLimit,
			MemoryPercent: e.MemoryPercent,
			NetIO:         e.NetIO,
			BlockIO:       e.BlockIO,
			PIDCount:      e.PIDCount,
		})
	}
	return out
}

// DockerComposeList returns all Docker Compose projects with their services.
func (d *DevOps) DockerComposeList() []DockerComposeProject {
	defer common.RecoverPanic()
	projects, err := devops.DockerComposeList()
	if err != nil {
		return []DockerComposeProject{}
	}
	out := make([]DockerComposeProject, 0, len(projects))
	for _, p := range projects {
		proj := DockerComposeProject{
			Project: p.Project,
			Status:  p.Status,
			WorkDir: p.WorkDir,
		}
		services, svcErr := devops.DockerComposePS(p.WorkDir)
		if svcErr == nil {
			svcOut := make([]DockerComposeService, 0, len(services))
			for _, s := range services {
				svcOut = append(svcOut, DockerComposeService{
					Name:  s.Name,
					State: s.State,
					Ports: s.Ports,
				})
			}
			proj.Services = svcOut
		}
		out = append(out, proj)
	}
	return out
}

// GetDockerNetworks returns all Docker networks.
func (d *DevOps) GetDockerNetworks() []DockerNetworkInfo {
	defer common.RecoverPanic()
	networks, err := devops.GetDockerNetworks()
	if err != nil {
		return []DockerNetworkInfo{}
	}
	out := make([]DockerNetworkInfo, 0, len(networks))
	for _, n := range networks {
		out = append(out, DockerNetworkInfo{
			ID:         n.ID,
			Name:       n.Name,
			Driver:     n.Driver,
			Scope:      n.Scope,
			Subnet:     n.Subnet,
			Gateway:    n.Gateway,
			Containers: n.Containers,
		})
	}
	return out
}

// GetDockerVolumes returns all Docker volumes.
func (d *DevOps) GetDockerVolumes() []DockerVolumeInfo {
	defer common.RecoverPanic()
	volumes, err := devops.GetDockerVolumes()
	if err != nil {
		return []DockerVolumeInfo{}
	}
	out := make([]DockerVolumeInfo, 0, len(volumes))
	for _, v := range volumes {
		out = append(out, DockerVolumeInfo{
			Driver:     v.Driver,
			Name:       v.Name,
			Mountpoint: v.Mountpoint,
			Size:       v.Size,
		})
	}
	return out
}

// DockerPrune prunes unused Docker resources.
func (d *DevOps) DockerPrune() DockerActionResult {
	defer common.RecoverPanic()
	msg, err := devops.DockerPrune()
	if err != nil {
		return DockerActionResult{
			Action:  "prune",
			Message: err.Error(),
			Success: false,
		}
	}
	return DockerActionResult{
		Action:  "prune",
		Message: msg,
		Success: true,
	}
}

// DockerKill forcefully stops a container.
func (d *DevOps) DockerKill(id string) DockerActionResult {
	defer common.RecoverPanic()
	msg, err := devops.DockerKill(id)
	if err != nil {
		return DockerActionResult{
			Action:  "kill",
			Message: err.Error(),
			Success: false,
		}
	}
	return DockerActionResult{
		Action:  "kill",
		Message: msg,
		Success: true,
	}
}

// DockerStart starts a stopped container.
func (d *DevOps) DockerStart(id string) DockerActionResult {
	defer common.RecoverPanic()
	err := devops.ControlContainer(id, "start")
	if err != nil {
		return DockerActionResult{
			Action:  "start",
			Message: err.Error(),
			Success: false,
		}
	}
	return DockerActionResult{
		Action:  "start",
		Message: "Container started successfully",
		Success: true,
	}
}

// DockerPause pauses a container.
func (d *DevOps) DockerPause(id string) DockerActionResult {
	defer common.RecoverPanic()
	msg, err := devops.DockerPause(id)
	if err != nil {
		return DockerActionResult{
			Action:  "pause",
			Message: err.Error(),
			Success: false,
		}
	}
	return DockerActionResult{
		Action:  "pause",
		Message: msg,
		Success: true,
	}
}

// DockerUnpause unpauses a paused container.
func (d *DevOps) DockerUnpause(id string) DockerActionResult {
	defer common.RecoverPanic()
	msg, err := devops.DockerUnpause(id)
	if err != nil {
		return DockerActionResult{
			Action:  "unpause",
			Message: err.Error(),
			Success: false,
		}
	}
	return DockerActionResult{
		Action:  "unpause",
		Message: msg,
		Success: true,
	}
}

// DockerRename renames a container.
func (d *DevOps) DockerRename(id string, newName string) DockerActionResult {
	defer common.RecoverPanic()
	msg, err := devops.DockerRename(id, newName)
	if err != nil {
		return DockerActionResult{
			Action:  "rename",
			Message: err.Error(),
			Success: false,
		}
	}
	return DockerActionResult{
		Action:  "rename",
		Message: msg,
		Success: true,
	}
}

// GetK8sNamespaces returns all Kubernetes namespaces.
func (d *DevOps) GetK8sNamespaces() []K8sNamespaceInfo {
	defer common.RecoverPanic()
	items, err := devops.GetK8sNamespaces()
	if err != nil {
		return []K8sNamespaceInfo{}
	}
	out := make([]K8sNamespaceInfo, 0, len(items))
	for _, item := range items {
		out = append(out, K8sNamespaceInfo{
			Name:   item.Name,
			Status: item.Status,
			Age:    item.Age,
		})
	}
	return out
}

// GetK8sDeployments returns Kubernetes deployments in the given namespace.
func (d *DevOps) GetK8sDeployments(namespace string) []K8sResourceItem {
	defer common.RecoverPanic()
	items, err := devops.GetK8sDeployments(namespace)
	if err != nil {
		return []K8sResourceItem{}
	}
	out := make([]K8sResourceItem, 0, len(items))
	for _, item := range items {
		out = append(out, K8sResourceItem{
			Name:      item.Name,
			Namespace: item.Namespace,
			Status:    item.Status,
			Age:       item.Age,
			Details:   item.Details,
		})
	}
	return out
}

// GetK8sServices returns Kubernetes services in the given namespace.
func (d *DevOps) GetK8sServices(namespace string) []K8sResourceItem {
	defer common.RecoverPanic()
	items, err := devops.GetK8sServices(namespace)
	if err != nil {
		return []K8sResourceItem{}
	}
	out := make([]K8sResourceItem, 0, len(items))
	for _, item := range items {
		out = append(out, K8sResourceItem{
			Name:      item.Name,
			Namespace: item.Namespace,
			Status:    item.Status,
			Age:       item.Age,
			Details:   item.Details,
		})
	}
	return out
}

// GetK8sPods returns Kubernetes pods in the given namespace.
func (d *DevOps) GetK8sPods(namespace string) []K8sResourceItem {
	defer common.RecoverPanic()
	items, err := devops.GetK8sPods(namespace)
	if err != nil {
		return []K8sResourceItem{}
	}
	out := make([]K8sResourceItem, 0, len(items))
	for _, item := range items {
		out = append(out, K8sResourceItem{
			Name:      item.Name,
			Namespace: item.Namespace,
			Status:    item.Status,
			Age:       item.Age,
			Details:   item.Details,
		})
	}
	return out
}

// GetK8sEvents returns Kubernetes events in the given namespace with a limit.
func (d *DevOps) GetK8sEvents(namespace string, limit int) []K8sEvent {
	defer common.RecoverPanic()
	items, err := devops.GetK8sEvents(namespace, limit)
	if err != nil {
		return []K8sEvent{}
	}
	out := make([]K8sEvent, 0, len(items))
	for _, item := range items {
		out = append(out, K8sEvent{
			LastSeen: item.LastSeen,
			Type:     item.Type,
			Reason:   item.Reason,
			Object:   item.Object,
			Message:  item.Message,
		})
	}
	return out
}

// GetK8sRollouts returns Kubernetes rollout status in the given namespace.
func (d *DevOps) GetK8sRollouts(namespace string) []K8sRolloutStatus {
	defer common.RecoverPanic()
	items, err := devops.GetK8sRollouts(namespace)
	if err != nil {
		return []K8sRolloutStatus{}
	}
	out := make([]K8sRolloutStatus, 0, len(items))
	for _, item := range items {
		out = append(out, K8sRolloutStatus{
			Name:      item.Name,
			Kind:      item.Kind,
			Ready:     item.Ready,
			Replicas:  item.Replicas,
			Updated:   item.Updated,
			Available: item.Available,
		})
	}
	return out
}

// K8sRestartDeployment restarts a Kubernetes deployment.
func (d *DevOps) K8sRestartDeployment(name, namespace string) K8sActionResult {
	defer common.RecoverPanic()
	msg, err := devops.K8sRestartDeployment(name, namespace)
	if err != nil {
		return K8sActionResult{
			Action:  "restart",
			Message: err.Error(),
			Success: false,
		}
	}
	return K8sActionResult{
		Action:  "restart",
		Message: msg,
		Success: true,
	}
}

// K8sRollbackDeployment rolls back a Kubernetes deployment to a previous revision.
func (d *DevOps) K8sRollbackDeployment(name, namespace string, revision int) K8sActionResult {
	defer common.RecoverPanic()
	msg, err := devops.K8sRollbackDeployment(name, namespace, revision)
	if err != nil {
		return K8sActionResult{
			Action:  "rollback",
			Message: err.Error(),
			Success: false,
		}
	}
	return K8sActionResult{
		Action:  "rollback",
		Message: msg,
		Success: true,
	}
}
