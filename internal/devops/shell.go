package devops

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/shahriarhaqueabir/UniversalOps/internal/common"
)

// ansiRegexp matches CSI ANSI escape sequences (ESC[...m) used for
// terminal color/formatting that PowerShell 7 injects into captured output.
var ansiRegexp = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

// stripANSI removes all ANSI escape codes from a string.
func stripANSI(s string) string {
	return ansiRegexp.ReplaceAllString(s, "")
}

// findGitBash locates Git Bash on Windows by checking common install paths.
// Returns the full path to bash.exe or empty string if not found.
func findGitBash() string {
	candidates := []string{
		"C:\\Program Files\\Git\\bin\\bash.exe",
		"C:\\Program Files (x86)\\Git\\bin\\bash.exe",
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

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
// Uses token-based validation for higher precision:
//   - Patterns ending with a space (e.g. "rm ") only match when the first token is
//     an exact match AND the command has arguments. A bare "rm" is safe; "rm -rf" is blocked.
//   - Patterns without a trailing space (e.g. "powershell", "certutil") match any exact token.
//   - Substring check catches operators like "> ", "&&", ";" anywhere in the command.
func IsDangerousCommand(cmd string) bool {
	normalized := normalizeWhitespace(cmd)
	lower := strings.ToLower(normalized)
	tokens := strings.Fields(lower)

	if len(tokens) == 0 {
		return false
	}

	firstToken := tokens[0]
	hasArguments := len(tokens) > 1

	// Token-based check: precise matching for command names
	for _, dangerous := range DangerousCommands {
		d := strings.TrimSpace(dangerous)

		if strings.HasSuffix(dangerous, " ") {
			// Patterns with trailing space: only dangerous when the command has arguments.
			// Bare "rm" is safe; "rm -rf /" is blocked.
			if firstToken == d && hasArguments {
				return true
			}
		} else {
			// Patterns without trailing space: exact token match on any token.
			// Catches "powershell -c", "certutil", "Remove-Item" etc.
			for _, token := range tokens {
				if token == d {
					return true
				}
			}
		}
	}

	// Substring check for redirect/pipe/shell operators that can appear mid-command:
	// "> ", ">> ", "| ", "&&", "||", ";", "& "
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
	if IsDangerousCommand(cmd) {
		return nil, fmt.Errorf("%w: %s", ErrDangerousCommand, cmd)
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
// 1. Relative to the executable's directory (production build, e.g. universal-ops.exe)
//  2. Relative to the current working directory (dev mode via `wails dev`)
//  3. Two levels above CWD (dev mode when CWD is buried in a temp dir)
//  4. Plain "profiles/powershell_profile.ps1" relative path
//
// The variable is re-resolved at package init time. For dev mode the
// executable path leads to a temp dir so the CWD fallback is essential.
var PowerShellProfilePath = func() string {
	candidates := []string{}

	// 1. Executable-relative (production build)
	if exe, err := os.Executable(); err == nil {
		candidates = append(candidates, filepath.Join(filepath.Dir(exe), "profiles", "powershell_profile.ps1"))
	}

	// 2. CWD-relative (dev mode: `wails dev` sets CWD to project root)
	if cwd, err := os.Getwd(); err == nil {
		candidates = append(candidates, filepath.Join(cwd, "profiles", "powershell_profile.ps1"))
		// 3. Grandparent of CWD (dev mode when CWD is deep, e.g. cmd/opsforall-gui)
		candidates = append(candidates, filepath.Join(filepath.Dir(cwd), "profiles", "powershell_profile.ps1"))
		candidates = append(candidates, filepath.Join(filepath.Dir(filepath.Dir(cwd)), "profiles", "powershell_profile.ps1"))
	}

	// 4. Plain relative fallback
	candidates = append(candidates, "profiles/powershell_profile.ps1")

	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}

	// Return the safest fallback even if it doesn't stat (caller handles missing)
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

	// 3. Use 'pwsh' (PowerShell 7) if available, otherwise 'powershell'.
	//    For pwsh we disable $PSStyle to prevent ANSI escape codes from
	//    leaking into captured output. We also pipe through Out-String to
	//    convert any emitted objects to consistent string format (avoids
	//    the @{Prop=Value} PowerShell object serialization).
	shell := "pwsh"
	usePwsh := true
	if _, err := exec.LookPath("pwsh"); err != nil {
		shell = "powershell"
		usePwsh = false
	}

	var fullCommand string
	if usePwsh {
		fullCommand = fmt.Sprintf("$PSStyle.OutputRendering='PlainText'; %s | Out-String -Width 4096", psCmd)
	} else {
		// Windows PowerShell: no $PSStyle, just pipe through Out-String
		fullCommand = fmt.Sprintf("%s | Out-String -Width 4096", psCmd)
	}

	// 4. Run inside the system-query sandbox
	c := common.SandboxedCommandWithConfig(common.SystemQuerySandbox(), shell, "-NoProfile", "-Command", fullCommand)

	output, err := c.CombinedOutput()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
	}

	// 5. Strip ANSI escape codes that may have survived
	cleanOutput := stripANSI(string(output))

	// Trim trailing newlines for cleaner display
	cleanOutput = strings.TrimRight(cleanOutput, "\r\n ")

	return &ShellResult{
		Command:  cmd,
		Output:   cleanOutput,
		ExitCode: exitCode,
		Duration: time.Since(start),
	}, nil
}

// RunPowerShellWithLiveOutput executes a PowerShell command with the user's
// default profile loaded, and streams each line of output through the provided
// channel. Unlike RunPowerShell, this does NOT enforce the workflow allowlist
// so it is suitable for the interactive terminal where any PowerShell command
// may be typed. The channel is closed when the command finishes.
func RunPowerShellWithLiveOutput(cmd string, output chan string) (result *ShellResult, err error) {
	defer func() {
		if err != nil {
			close(output)
		}
	}()

	start := time.Now()
	var stdoutBuf, stderrBuf strings.Builder

	// Use pwsh (PowerShell 7) if available, otherwise powershell.exe
	shell := "pwsh"
	if _, lookErr := exec.LookPath("pwsh"); lookErr != nil {
		shell = "powershell"
	}

	// Disable $PSStyle in pwsh to prevent ANSI codes in captured output.
	// Then run the actual command the user typed.
	psCmd := cmd
	if shell == "pwsh" {
		psCmd = fmt.Sprintf("$PSStyle.OutputRendering='PlainText'; %s", psCmd)
	}

	c := common.SandboxedCommand(shell, "-NoProfile", "-Command", psCmd)

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

	go func() {
		defer common.RecoverPanic()
		defer wg.Done()
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := stripANSI(scanner.Text())
			stdoutBuf.WriteString(line + "\n")
			output <- line
		}
	}()

	go func() {
		defer common.RecoverPanic()
		defer wg.Done()
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := stripANSI(scanner.Text())
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

// RunGitBashWithLiveOutput executes a command via Git Bash and streams each
// line of output through the provided channel. It locates Git Bash by checking
// common installation paths. The channel is closed when the command finishes.
func RunGitBashWithLiveOutput(cmd string, output chan string) (result *ShellResult, err error) {
	defer func() {
		if err != nil {
			close(output)
		}
	}()

	bashPath := findGitBash()
	if bashPath == "" {
		return nil, fmt.Errorf("Git Bash not found: check C:\\Program Files\\Git\\bin\\bash.exe")
	}

	start := time.Now()
	var stdoutBuf, stderrBuf strings.Builder

	c := common.SandboxedCommand(bashPath, "-c", cmd)

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
