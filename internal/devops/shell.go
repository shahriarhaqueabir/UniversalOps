package devops

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
)

// ShellResult holds the result of a shell command execution.
type ShellResult struct {
	Command  string
	Output   string
	ExitCode int
	Duration time.Duration
}

// DangerousCommands is the list of command patterns that are blocked server-side.
// These are checked as substrings (case-insensitive).
var DangerousCommands = []string{
	"rm ", "del ", "format ", "mkfs", "kill ", "shutdown", "reboot",
	"dd ", "grub", "fdisk", "mkswap", "mount ", "chmod", "chown",
	"> ", ">> ", "| ", "&&", "||", ";",
	"powershell", "pwsh ", "cmd ", "taskkill ", "wmic ",
	"net ", "sc ", "reg ", "curl ", "wget ", "certutil ",
	"bitsadmin", "Remove-Item", "New-Item", "Start-Process",
	"rd ", "rmdir ", "takeown ", "icacls ", "cacls ", "schtasks ",
	"attrib ", "& ",
}

// ErrDangerousCommand is returned when a dangerous command is rejected.
var ErrDangerousCommand = errors.New("command contains dangerous patterns and was blocked server-side")

// IsDangerousCommand checks if a command contains dangerous patterns (case-insensitive).
func IsDangerousCommand(cmd string) bool {
	lower := strings.ToLower(cmd)
	for _, dangerous := range DangerousCommands {
		if strings.Contains(lower, dangerous) {
			return true
		}
	}
	return false
}

// RunCommand executes a shell command and returns the combined output.
// It uses cmd /c on Windows and sh -c on Unix.
func RunCommand(cmd string) (*ShellResult, error) {
	if IsDangerousCommand(cmd) {
		return nil, fmt.Errorf("%w: %s", ErrDangerousCommand, cmd)
	}

	start := time.Now()

	var c *exec.Cmd
	if runtime.GOOS == "windows" {
		c = common.SandboxedCommand("cmd", "/c", cmd)
	} else {
		c = common.SandboxedCommand("sh", "-c", cmd)
	}

	output, err := c.CombinedOutput()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
	}

	return &ShellResult{
		Command:  cmd,
		Output:   string(output),
		ExitCode: exitCode,
		Duration: time.Since(start),
	}, nil
}

// RunCommandWithLiveOutput executes a command and streams each line of output
// through the provided channel. The channel is closed when the command finishes.
func RunCommandWithLiveOutput(cmd string, output chan string) (*ShellResult, error) {
	if IsDangerousCommand(cmd) {
		return nil, fmt.Errorf("%w: %s", ErrDangerousCommand, cmd)
	}

	start := time.Now()
	var stdoutBuf, stderrBuf strings.Builder

	var c *exec.Cmd
	if runtime.GOOS == "windows" {
		c = common.SandboxedCommand("cmd", "/c", cmd)
	} else {
		c = common.SandboxedCommand("sh", "-c", cmd)
	}

	stdout, err := c.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderr, err := c.StderrPipe()
	if err != nil {
		return nil, err
	}

	if err := c.Start(); err != nil {
		return nil, err
	}

	// Read stdout
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			stdoutBuf.WriteString(line + "\n")
			output <- line
		}
	}()

	// Read stderr
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			stderrBuf.WriteString(line + "\n")
			output <- line
		}
	}()

	err = c.Wait()
	close(output)

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
	}

	return &ShellResult{
		Command:  cmd,
		Output:   stdoutBuf.String() + stderrBuf.String(),
		ExitCode: exitCode,
		Duration: time.Since(start),
	}, nil
}

// AllowedPowerShellWorkflows is the list of approved PowerShell workflow commands
// that may be executed through RunPowerShell.
var AllowedPowerShellWorkflows = []string{
	"Invoke-HawkDailyOps",
	"Invoke-HawkSystemReview",
	"Invoke-HawkSecurityAudit",
	"Invoke-HawkNetworkDiagnostics",
	"Invoke-HawkThreatHunt",
	"Invoke-HawkChangeAudit",
	"Invoke-HawkComplianceCheck",
}

// PowerShellProfilePath is the path to the PowerShell profile script.
// It is resolved relative to the executable's working directory.
// Override for testing.
var PowerShellProfilePath = filepath.Join("profiles", "powershell_profile.ps1")

// isAllowedWorkflow checks whether cmd is one of the approved PowerShell workflow names.
func isAllowedWorkflow(cmd string) bool {
	for _, wf := range AllowedPowerShellWorkflows {
		if cmd == wf {
			return true
		}
	}
	return false
}

// RunPowerShell executes an allowed PowerShell workflow command.
// It enforces a server-side allowlist, requires the profile to exist,
// and runs inside the system-query sandbox.
func RunPowerShell(cmd string) (*ShellResult, error) {
	start := time.Now()

	// 1. Server-side allowlist enforcement — reject unknown commands
	if !isAllowedWorkflow(cmd) {
		return nil, fmt.Errorf("RunPowerShell: command %q is not in the allowed PowerShell workflow allowlist", cmd)
	}

	// 2. Profile must exist — no silent fallback
	profilePath := PowerShellProfilePath
	if _, err := os.Stat(profilePath); err != nil {
		return nil, fmt.Errorf("RunPowerShell: PowerShell profile not found at %s: %w", PowerShellProfilePath, err)
	}
	psCmd := fmt.Sprintf(". '%s'; %s", profilePath, cmd)

	// 3. Use 'pwsh' (PowerShell 7) if available, otherwise 'powershell'
	shell := "pwsh"
	if _, err := exec.LookPath("pwsh"); err != nil {
		shell = "powershell"
	}

	// 4. Run inside the system-query sandbox (replaces raw exec.Command)
	c := common.SandboxedCommandWithConfig(common.SystemQuerySandbox(), shell, "-NoProfile", "-Command", psCmd)

	output, err := c.CombinedOutput()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
	}

	return &ShellResult{
		Command:  cmd,
		Output:   string(output),
		ExitCode: exitCode,
		Duration: time.Since(start),
	}, nil
}
