package sysops

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
)

// actionErrorHints maps common exit code patterns to user-friendly messages per action.
func actionErrorHint(action SystemAction, err error, output string) string {
	errMsg := err.Error()
	lower := strings.ToLower(errMsg + " " + output)

	// Access denied — most common for privileged operations
	if strings.Contains(lower, "access denied") || strings.Contains(lower, "permission denied") ||
		strings.Contains(lower, "required privilege") || strings.Contains(lower, "elevated") {
		switch action {
		case ActionFlushDNS:
			return "DNS cache flush requires administrator privileges. Run Universal-Ops as Administrator."
		case ActionClearTemp:
			return "Clearing temp files needs administrator rights for some system folders. Run as Administrator."
		case ActionDiskCleanup:
			return "Disk Cleanup requires administrator privileges. Run Universal-Ops as Administrator."
		case ActionDefrag:
			return "Defragmentation requires administrator privileges. Run Universal-Ops as Administrator."
		case ActionRestartService:
			return "Restarting services requires administrator privileges. Run Universal-Ops as Administrator."
		default:
			return "This action requires administrator privileges. Run Universal-Ops as Administrator."
		}
	}

	// File not found
	if strings.Contains(lower, "not found") || strings.Contains(lower, "no such file") ||
		strings.Contains(lower, "cannot find") {
		switch action {
		case ActionFlushDNS:
			return "ipconfig command not found — ensure System32 is in your PATH"
		case ActionClearTemp:
			return "PowerShell not found or TEMP path inaccessible"
		case ActionDiskCleanup:
			return "cleanmgr.exe not found — Windows System32 directory may be missing"
		case ActionDefrag:
			return "defrag.exe not found — ensure System32 is in your PATH"
		default:
			return "Required system utility not found. Check your system PATH."
		}
	}

	// Exit code 1 — generic failure
	if strings.Contains(errMsg, "exit status 1") {
		switch action {
		case ActionFlushDNS:
			if strings.Contains(lower, "unrecognized") || strings.Contains(lower, "bad") {
				return "ipconfig /flushdns failed — network stack may be corrupted. Try 'netsh winsock reset'"
			}
			return "DNS cache flush failed. Try running as Administrator."
		case ActionClearTemp:
			return "Temp file cleanup encountered locked files — some files may be in use. Try again after a reboot."
		case ActionCleanPkgCache:
			return "Package cache update failed — check your internet connection and try again"
		case ActionSystemUpdate:
			return "System update check failed — ensure internet connectivity and try again"
		case ActionDiskCleanup:
			return "Disk cleanup encountered an error — some files may be in use"
		case ActionDefrag:
			return "Defragmentation failed — check drive health with 'chkdsk C:'"
		default:
			return "Action failed with exit code 1 — check the Output tab for details"
		}
	}

	// Timeout / did not respond
	if strings.Contains(lower, "timeout") || strings.Contains(lower, "timed out") {
		return "Action timed out — the system may be unresponsive or the command hung. Try again."
	}

	// Service-specific
	if action == ActionRestartService {
		if strings.Contains(lower, "not exist") || strings.Contains(lower, "not found") {
			return "Service name not recognized — verify the service name in Services.msc"
		}
		if strings.Contains(lower, "not start") || strings.Contains(lower, "failed to start") {
			return "Service exists but failed to restart — check the service configuration or Event Viewer for details"
		}
	}

	// Fallback
	return fmt.Sprintf("Action failed: %s. Check the Output tab for full details.", errMsg)
}

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
	ActionDiskCleanup    SystemAction = "disk_cleanup"
	ActionDefrag         SystemAction = "defrag"
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
			cmd = common.HiddenCommand("shutdown", "/r", "/t", "0")
		} else {
			cmd = exec.Command("sudo", "shutdown", "-r", "now")
		}
	case ActionShutdown:
		if runtime.GOOS == "windows" {
			cmd = common.HiddenCommand("shutdown", "/s", "/t", "0")
		} else {
			cmd = exec.Command("sudo", "shutdown", "-h", "now")
		}
	case ActionSleep:
		if runtime.GOOS == "windows" {
			cmd = common.HiddenCommand("rundll32.exe", "powrprof.dll,SetSuspendState", "0,1,0")
		} else {
			cmd = exec.Command("systemctl", "suspend")
		}
	case ActionHibernate:
		if runtime.GOOS == "windows" {
			cmd = common.HiddenCommand("rundll32.exe", "powrprof.dll,SetSuspendState", "1,1,0")
		} else {
			cmd = exec.Command("systemctl", "hibernate")
		}
	case ActionFlushDNS:
		if runtime.GOOS == "windows" {
			cmd = common.HiddenCommand("ipconfig", "/flushdns")
		} else {
			cmd = exec.Command("sudo", "resolvectl", "flush-caches")
		}
	case ActionClearTemp:
		if runtime.GOOS == "windows" {
			cmd = common.HiddenCommand("powershell", "-NoProfile", "-Command", "Remove-Item -Recurse -Force $env:TEMP\\* -ErrorAction SilentlyContinue")
		} else {
			cmd = exec.Command("sudo", "rm", "-rf", "/tmp/*")
		}
	case ActionCleanPkgCache:
		if runtime.GOOS == "windows" {
			cmd = common.HiddenCommand("powershell", "-Command", "winget source update")
		} else {
			cmd = exec.Command("sudo", "apt-get", "clean")
		}
	case ActionSystemUpdate:
		if runtime.GOOS == "windows" {
			cmd = common.HiddenCommand("powershell", "-Command", "winget upgrade --all --accept-package-agreements --accept-source-agreements")
		} else {
			cmd = exec.Command("sudo", "apt-get", "update", "&&", "sudo", "apt-get", "upgrade", "-y")
		}
	case ActionDiskCleanup:
		if runtime.GOOS == "windows" {
			cmd = common.HiddenCommand("cleanmgr", "/sagerun:1")
		} else {
			cmd = exec.Command("sudo", "apt-get", "autoremove", "-y")
		}
	case ActionDefrag:
		if runtime.GOOS == "windows" {
			cmd = common.HiddenCommand("defrag", "C:", "/O")
		} else {
			cmd = exec.Command("sudo", "fstrim", "-av")
		}
	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}

	output, err := cmd.CombinedOutput()
	result.Output = string(output)
	if err != nil {
		result.Success = false
		result.Message = actionErrorHint(action, err, string(output))
		return result, fmt.Errorf("%s: %w", result.Message, err)
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
		cmd = common.HiddenCommand("powershell", "-Command", fmt.Sprintf("Restart-Service -Name '%s' -Force", name))
	} else {
		cmd = exec.Command("sudo", "systemctl", "restart", name)
	}

	output, err := cmd.CombinedOutput()
	result.Output = string(output)
	if err != nil {
		result.Success = false
		result.Message = actionErrorHint(ActionRestartService, err, string(output))
		return result, fmt.Errorf("%s: %w", result.Message, err)
	}

	result.Success = true
	result.Message = fmt.Sprintf("Service %q restarted successfully", name)
	return result, nil
}
