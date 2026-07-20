package sysops

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
)

// SystemAction represents a system action type.
type SystemAction string

const (
	ActionReboot         SystemAction = "reboot"
	ActionShutdown       SystemAction = "shutdown"
	ActionSleep          SystemAction = "sleep"
	ActionHibernate      SystemAction = "hibernate"
	ActionFlushDNS       SystemAction = "flush_dns"
	ActionClearTemp      SystemAction = "clear_temp"
	ActionCleanPkgCache  SystemAction = "clean_pkg_cache"
	ActionSystemUpdate   SystemAction = "system_update"
	ActionRestartService SystemAction = "restart_service"
)

// ActionResult holds the result of a system action.
type ActionResult struct {
	Action  string `json:"action"`
	Success bool   `json:"success"`
	Message string `json:"message"`
	Output  string `json:"output"`
}

// RunSystemAction executes a system action and returns the result.
func RunSystemAction(action SystemAction) (*ActionResult, error) {
	result := &ActionResult{Action: string(action)}

	var cmd *exec.Cmd

	switch action {
	case ActionReboot:
		if runtime.GOOS == "windows" {
			cmd = exec.Command("shutdown", "/r", "/t", "0")
		} else {
			cmd = exec.Command("sudo", "shutdown", "-r", "now")
		}
	case ActionShutdown:
		if runtime.GOOS == "windows" {
			cmd = exec.Command("shutdown", "/s", "/t", "0")
		} else {
			cmd = exec.Command("sudo", "shutdown", "-h", "now")
		}
	case ActionSleep:
		if runtime.GOOS == "windows" {
			cmd = exec.Command("rundll32.exe", "powrprof.dll,SetSuspendState", "0,1,0")
		} else {
			cmd = exec.Command("systemctl", "suspend")
		}
	case ActionHibernate:
		if runtime.GOOS == "windows" {
			cmd = exec.Command("rundll32.exe", "powrprof.dll,SetSuspendState", "1,1,0")
		} else {
			cmd = exec.Command("systemctl", "hibernate")
		}
	case ActionFlushDNS:
		if runtime.GOOS == "windows" {
			cmd = exec.Command("ipconfig", "/flushdns")
		} else {
			cmd = exec.Command("sudo", "resolvectl", "flush-caches")
		}
	case ActionClearTemp:
		if runtime.GOOS == "windows" {
			cmd = exec.Command("powershell", "-Command", "Remove-Item -Recurse -Force $env:TEMP\\* -ErrorAction SilentlyContinue")
		} else {
			cmd = exec.Command("sudo", "rm", "-rf", "/tmp/*")
		}
	case ActionCleanPkgCache:
		if runtime.GOOS == "windows" {
			cmd = exec.Command("powershell", "-Command", "winget source update")
		} else {
			cmd = exec.Command("sudo", "apt-get", "clean")
		}
	case ActionSystemUpdate:
		if runtime.GOOS == "windows" {
			cmd = exec.Command("powershell", "-Command", "winget upgrade --all --accept-package-agreements --accept-source-agreements")
		} else {
			cmd = exec.Command("sudo", "apt-get", "update", "&&", "sudo", "apt-get", "upgrade", "-y")
		}
	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}

	output, err := cmd.CombinedOutput()
	result.Output = string(output)
	if err != nil {
		result.Success = false
		result.Message = fmt.Sprintf("Action failed: %v", err)
		return result, err
	}

	result.Success = true
	result.Message = fmt.Sprintf("Action '%s' completed successfully", action)
	return result, nil
}

// RestartService restarts a system service by name.
func RestartService(name string) (*ActionResult, error) {
	result := &ActionResult{Action: "restart_service"}

	// Validate service name to prevent command injection (SEC-1).
	if !common.ValidServiceName(name) {
		return nil, fmt.Errorf("invalid service name: %q", name)
	}

	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		cmd = exec.Command("powershell", "-Command", fmt.Sprintf("Restart-Service -Name '%s' -Force", name))
	} else {
		cmd = exec.Command("sudo", "systemctl", "restart", name)
	}

	output, err := cmd.CombinedOutput()
	result.Output = string(output)
	if err != nil {
		result.Success = false
		result.Message = fmt.Sprintf("Restart failed: %v", err)
		return result, err
	}

	result.Success = true
	result.Message = fmt.Sprintf("Service %q restarted successfully", name)
	return result, nil
}
