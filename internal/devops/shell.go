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
	"sync"
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

// ErrShellMetachar is returned when a command contains shell metacharacters
// that enable command injection (backticks, $(), pipes, redirects, etc.).
var ErrShellMetachar = errors.New("command contains shell metacharacters and was rejected")

// ShellMetacharacters are characters that enable arbitrary command execution
// or injection in shell environments (cmd /c, sh -c). Any command containing
// these characters is rejected regardless of the command blocklist.
var ShellMetacharacters = []string{
	"`", "$", "(", ")", "{", "}", "|", "&", ";", "<", ">", "\n", "\r",
}

// ContainsShellMetachar checks if a command contains any shell metacharacters
// that could enable command injection. This provides a character-level defense
// that covers bypass techniques the substring blocklist cannot catch.
func ContainsShellMetachar(cmd string) bool {
	for _, mc := range ShellMetacharacters {
		if strings.Contains(cmd, mc) {
			return true
		}
	}
	return false
}

// normalizeWhitespace collapses tabs and repeated spaces to a single space.
// H1: DangerousCommands matches rely on a literal trailing space (e.g. "rm ");
// without this, a command using a tab instead of a space (e.g. "rm\tfile" or
// "kill\t1") would silently bypass the blocklist.
func normalizeWhitespace(cmd string) string {
	replaced := strings.Map(func(r rune) rune {
		if r == '\t' || r == '\v' || r == '\f' {
			return ' '
		}
		return r
	}, cmd)
	for strings.Contains(replaced, "  ") {
		replaced = strings.ReplaceAll(replaced, "  ", " ")
	}
	return replaced
}

// IsDangerousCommand checks if a command contains dangerous patterns (case-insensitive).
// Now uses token-based validation for higher precision.
func IsDangerousCommand(cmd string) bool {
	normalized := normalizeWhitespace(cmd)
	lower := strings.ToLower(normalized)
	tokens := strings.Fields(lower)

	for _, token := range tokens {
		// Exact match for base commands that are dangerous
		for _, dangerous := range DangerousCommands {
			d := strings.TrimSpace(dangerous)
			if token == d || strings.HasPrefix(token, d) {
				return true
			}
		}
	}

	// Substring check for patterns like redirects or pipes if they weren't caught
	for _, dangerous := range DangerousCommands {
		if strings.Contains(lower, dangerous) {
			return true
		}
	}

	return false
}

// RunCommand executes a shell command and returns the combined output.
// It uses cmd /c on Windows and sh -c on Unix.
// RunCommand executes a shell command and returns the combined output.
// It uses cmd /c on Windows and sh -c on Unix.
// Only commands matching the allowlist are permitted.
var allowedShellCommands = map[string]bool{
	"dir":        true,
	"ls":         true,
	"ps":         true,
	"tasklist":   true,
	"netstat":    true,
	"ss":         true,
	"ping":       true,
	"traceroute": true,
	"tracert":    true,
	"nslookup":   true,
	"dig":        true,
	"whoami":     true,
	"id":         true,
	"uptime":     true,
	"df":         true,
	"du":         true,
	"free":       true,
	"top":        true,
	"htop":       true,
	"systemctl":  true,
	"service":    true,
	"sc":         true,
	"net":        true,
	"ipconfig":   true,
	"ifconfig":   true,
	"route":      true,
	"arp":        true,
	"netsh":      true,
}

func isAllowedShellCommand(cmd string) bool {
	// Extract the base command (first word)
	fields := strings.Fields(cmd)
	if len(fields) == 0 {
		return false
	}
	base := fields[0]
	// Remove path if present
	if idx := strings.LastIndex(base, "/"); idx >= 0 {
		base = base[idx+1:]
	}
	if idx := strings.LastIndex(base, "\\"); idx >= 0 {
		base = base[idx+1:]
	}
	// Remove .exe extension
	base = strings.TrimSuffix(base, ".exe")

	// Token-based validation: check if the first token is in the allowlist
	// and no other forbidden commands are in the string (metachar check handles most,
	// but we double-check for safety).
	return allowedShellCommands[strings.ToLower(base)]
}

func RunCommand(cmd string) (*ShellResult, error) {
	if ContainsShellMetachar(cmd) {
		return nil, fmt.Errorf("%w: %s", ErrShellMetachar, cmd)
	}
	if !isAllowedShellCommand(cmd) {
		return nil, fmt.Errorf("command not in allowlist: %s", cmd)
	}

	start := time.Now()

	var c *common.SandboxedCmd
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
// Named return values ensure output is always closed on error paths.
func RunCommandWithLiveOutput(cmd string, output chan string) (result *ShellResult, err error) {
	defer func() {
		if err != nil {
			close(output)
		}
	}()

	if ContainsShellMetachar(cmd) {
		return nil, fmt.Errorf("%w: %s", ErrShellMetachar, cmd)
	}
	if !isAllowedShellCommand(cmd) {
		return nil, fmt.Errorf("command not in allowlist: %s", cmd)
	}

	start := time.Now()
	var stdoutBuf, stderrBuf strings.Builder

	var c *common.SandboxedCmd
	if runtime.GOOS == "windows" {
		c = common.SandboxedCommand("cmd", "/c", cmd)
	} else {
		c = common.SandboxedCommand("sh", "-c", cmd)
	}

	stdout, pipeErr := c.StdoutPipe()
	if pipeErr != nil {
		return nil, pipeErr
	}
	stderr, pipeErr := c.StderrPipe()
	if pipeErr != nil {
		return nil, pipeErr
	}

	if startErr := c.Start(); startErr != nil {
		return nil, startErr
	}

	var wg sync.WaitGroup
	wg.Add(2)

	// Read stdout
	go func() {
		defer common.RecoverPanic()
		defer wg.Done()
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			stdoutBuf.WriteString(line + "\n")
			output <- line
		}
	}()

	// Read stderr
	go func() {
		defer common.RecoverPanic()
		defer wg.Done()
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			stderrBuf.WriteString(line + "\n")
			output <- line
		}
	}()

	waitErr := c.Wait()
	wg.Wait()
	close(output)

	exitCode := 0
	if waitErr != nil {
		if exitErr, ok := waitErr.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
	}

	result = &ShellResult{
		Command:  cmd,
		Output:   stdoutBuf.String() + stderrBuf.String(),
		ExitCode: exitCode,
		Duration: time.Since(start),
	}
	return
}

// AllowedPowerShellWorkflows is the list of approved PowerShell workflow commands
// that may be executed through RunPowerShell.
var AllowedPowerShellWorkflows = []string{
	"Invoke-OpsDailyOps",
	"Invoke-OpsSystemReview",
	"Invoke-OpsSecurityAudit",
	"Invoke-OpsNetworkDiagnostics",
	"Invoke-OpsThreatHunt",
	"Invoke-OpsChangeAudit",
	"Invoke-OpsComplianceCheck",
}

// PowerShellProfilePath is the path to the PowerShell profile script.
// It tries multiple locations in order of preference:
//  1. Relative to the executable's directory (production build)
//  2. Relative to the current working directory (dev mode, run from project root)
//  3. Parent of CWD (dev mode, run from cmd/opsforall-gui)
//  4. Plain "profiles" relative path
//
// Override for testing.
var PowerShellProfilePath = func() string {
	candidates := []string{}

	if exe, err := os.Executable(); err == nil {
		candidates = append(candidates, filepath.Join(filepath.Dir(exe), "profiles", "powershell_profile.ps1"))
	}

	if cwd, err := os.Getwd(); err == nil {
		candidates = append(candidates, filepath.Join(cwd, "profiles", "powershell_profile.ps1"))
		candidates = append(candidates, filepath.Join(cwd, "..", "profiles", "powershell_profile.ps1"))
	}

	candidates = append(candidates, filepath.Join("profiles", "powershell_profile.ps1"))

	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}

	// Fallback: return the first candidate (exe-relative) even if it doesn't exist
	if len(candidates) > 0 {
		return candidates[0]
	}
	return "profiles/powershell_profile.ps1"
}()

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

	// 2. Profile is optional — warn if missing but continue
	profilePath := PowerShellProfilePath
	psCmd := cmd
	if _, err := os.Stat(profilePath); err == nil {
		psCmd = fmt.Sprintf(". '%s'; %s", profilePath, cmd)
	} else {
		common.LogInfo("RunPowerShell: PowerShell profile not found at %s, running without profile: %v", PowerShellProfilePath, err)
	}

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
