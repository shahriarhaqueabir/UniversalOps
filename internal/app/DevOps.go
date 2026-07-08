package app

import (
	"errors"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"

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
			runtime.EventsEmit(d.app.ctx, EventCmdLine, map[string]string{
				"id":   id,
				"line": line,
			})
		}
		runtime.EventsEmit(d.app.ctx, EventCmdDone, id)
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
			Name:    e.Name,
			Path:    e.Path,
			Size:    e.Size,
			IsDir:   e.IsDir,
			Mode:    e.Mode.String(),
			ModTime: e.ModTime.Format(time.RFC3339),
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
