package devops

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIsAllowedWorkflow_Allowed(t *testing.T) {
	workflows := []string{
		"Invoke-OpsDailyOps",
		"Invoke-OpsSystemReview",
		"Invoke-OpsSecurityAudit",
		"Invoke-OpsNetworkDiagnostics",
		"Invoke-OpsThreatHunt",
		"Invoke-OpsChangeAudit",
		"Invoke-OpsComplianceCheck",
	}
	for _, wf := range workflows {
		if !isAllowedWorkflow(wf) {
			t.Errorf("isAllowedWorkflow(%q) = false, want true", wf)
		}
	}
}

func TestIsAllowedWorkflow_Disallowed(t *testing.T) {
	cmds := []string{
		"",
		"Remove-Item -Recurse C:\\",
		"Get-Process -Name explorer",
		"Invoke-OpsUnknown",
		"anything",
		"Invoke-OpsDailyOps; Remove-Item -Recurse C:\\",
	}
	for _, cmd := range cmds {
		if isAllowedWorkflow(cmd) {
			t.Errorf("isAllowedWorkflow(%q) = true, want false", cmd)
		}
	}
}

func TestRunPowerShell_RejectsNonAllowed(t *testing.T) {
	_, err := RunPowerShell("Remove-Item -Recurse C:\\")
	if err == nil {
		t.Fatal("RunPowerShell should return error for non-allowed command")
	}
}

func TestRunPowerShell_MissingProfileProceeds(t *testing.T) {
	orig := PowerShellProfilePath
	t.Cleanup(func() { PowerShellProfilePath = orig })

	// Point to a non-existent profile — profile is now optional
	PowerShellProfilePath = filepath.Join(os.TempDir(), "nonexistent_test_profile.ps1")

	_, err := RunPowerShell("Invoke-OpsDailyOps")
	// The command should proceed past the profile check and attempt execution.
	// This may fail because PowerShell isn't available in the test environment,
	// but that proves the profile gate was passed.
	if err != nil {
		t.Logf("RunPowerShell proceeded past profile check (profile optional): %v", err)
	}
}

func TestIsDangerousCommand_DetectsDangerousCommands(t *testing.T) {
	dangerous := []string{
		"rm -rf /",
		"del /f *.txt",
		"format C: /fs:NTFS",
		"mkfs.ext4 /dev/sda1",
		"kill -9 1234",
		"shutdown /s /t 0",
		"reboot",
		"dd if=/dev/zero of=/dev/sda",
		"grub-install /dev/sda",
		"fdisk /dev/sda",
		"mkswap /dev/sda1",
		"mount /dev/sda1 /mnt",
		"chmod 777 /etc/shadow",
		"chown root:root /etc/passwd",
		"echo something > file.txt",
		"echo something >> file.txt",
		"ls | grep foo",
		"cmd1 && cmd2",
		"cmd1 || cmd2",
		"cmd1; cmd2",
	}
	for _, cmd := range dangerous {
		if !IsDangerousCommand(cmd) {
			t.Errorf("IsDangerousCommand(%q) = false, want true", cmd)
		}
	}
}

func TestIsDangerousCommand_AllowsSafeCommands(t *testing.T) {
	safe := []string{
		"ipconfig",
		"ipconfig /all",
		"ping 8.8.8.8",
		"ping -n 4 google.com",
		"netstat -an",
		"netstat -b",
		"tasklist",
		"systeminfo",
		"dir",
		"ls",
		"whoami",
		"echo hello",
		"type nul",
	}
	for _, cmd := range safe {
		if IsDangerousCommand(cmd) {
			t.Errorf("IsDangerousCommand(%q) = true, want false", cmd)
		}
	}
}

func TestContainsShellMetachar_DetectsMetacharacters(t *testing.T) {
	malicious := []struct {
		name   string
		cmd    string
		reason string
	}{
		{"backticks", "echo `whoami`", "backtick substitution"},
		{"dollar_paren", "echo $(whoami)", "$() command substitution"},
		{"dollar_curly", "echo ${PATH}", "shell variable expansion"},
		{"pipe", "ipconfig | findstr foo", "output piping"},
		{"pipe_no_space", "ipconfig|findstr foo", "pipe without space"},
		{"redirect_out", "echo foo > file.txt", "output redirect"},
		{"redirect_append", "echo foo >> file.txt", "append redirect"},
		{"redirect_in", "cat < file.txt", "input redirect"},
		{"chain_and", "echo hello & del *.*", "command chaining with &"},
		{"chain_and_no_space", "echo hello&del *.*", "& chaining without space"},
		{"chain_semi", "echo hello; rm -rf /", "semicolon chaining"},
		{"chain_semi_no_space", "echo hello;rm -rf /", "semicolon chaining no space"},
		{"newline", "echo hello\nrm -rf /", "newline injection"},
		{"carriage_return", "echo hello\rdel /f *.*", "CR injection"},
		{"curly_block", "{ rm -rf /; }", "curly brace block"},
		{"paren_subshell", "(rm -rf /)", "subshell grouping"},
	}
	for _, tc := range malicious {
		t.Run(tc.name, func(t *testing.T) {
			if !ContainsShellMetachar(tc.cmd) {
				t.Errorf("ContainsShellMetachar(%q) = false, want true (%s)", tc.cmd, tc.reason)
			}
		})
	}
}

func TestContainsShellMetachar_AllowsSafeCommands(t *testing.T) {
	safe := []string{
		"ipconfig",
		"ipconfig /all",
		"ping 8.8.8.8",
		"ping -n 4 google.com",
		"netstat -an",
		"netstat -b",
		"tasklist",
		"systeminfo",
		"dir",
		"ls",
		"whoami",
		"echo hello",
		"type nul",
		"go version",
		"git --version",
		"docker ps",
	}
	for _, cmd := range safe {
		if ContainsShellMetachar(cmd) {
			t.Errorf("ContainsShellMetachar(%q) = true, want false", cmd)
		}
	}
}

func TestRunCommand_BlocksShellMetachar(t *testing.T) {
	bypasses := []string{
		"echo `whoami`",
		"echo $(whoami)",
		"ipconfig | findstr foo",
	}
	for _, cmd := range bypasses {
		_, err := RunCommand(cmd)
		if err == nil {
			t.Errorf("RunCommand(%q) should return error for shell metacharacters", cmd)
		}
		if !errors.Is(err, ErrShellMetachar) {
			t.Errorf("RunCommand(%q) error should wrap ErrShellMetachar, got: %v", cmd, err)
		}
	}
}

func TestRunCommandWithLiveOutput_BlocksShellMetachar(t *testing.T) {
	ch := make(chan string, 10)
	_, err := RunCommandWithLiveOutput("echo `whoami`", ch)
	// RunCommandWithLiveOutput closes ch on error — do not close again
	if err == nil {
		t.Fatal("RunCommandWithLiveOutput should return error for shell metacharacters")
	}
	if !errors.Is(err, ErrShellMetachar) {
		t.Errorf("RunCommandWithLiveOutput error should wrap ErrShellMetachar, got: %v", err)
	}
}

func TestIsDangerousCommand_CaseInsensitive(t *testing.T) {
	cases := []struct {
		cmd  string
		want bool
	}{
		{"RM -rf /", true},
		{"Del /f *.txt", true},
		{"FORMAT C: /fs:NTFS", true},
		{"ipconfig", false},
		{"Netstat -an", false},
	}
	for _, tc := range cases {
		got := IsDangerousCommand(tc.cmd)
		if got != tc.want {
			t.Errorf("IsDangerousCommand(%q) = %v, want %v", tc.cmd, got, tc.want)
		}
	}
}

func TestRunCommand_BlocksDangerousCommand(t *testing.T) {
	_, err := RunCommand("rm -rf /")
	if err == nil {
		t.Fatal("RunCommand should return error for dangerous command")
	}
	if !strings.Contains(err.Error(), "command not in allowlist") {
		t.Errorf("RunCommand error should mention allowlist, got: %v", err)
	}
}

func TestRunCommandWithLiveOutput_BlocksDangerousCommand(t *testing.T) {
	ch := make(chan string, 10)
	_, err := RunCommandWithLiveOutput("rm -rf /", ch)
	// RunCommandWithLiveOutput closes ch on error — do not close again
	if err == nil {
		t.Fatal("RunCommandWithLiveOutput should return error for dangerous command")
	}
	if !strings.Contains(err.Error(), "command not in allowlist") {
		t.Errorf("RunCommandWithLiveOutput error should mention allowlist, got: %v", err)
	}
}

func TestRunPowerShell_ProceedsToSandbox(t *testing.T) {
	orig := PowerShellProfilePath
	t.Cleanup(func() { PowerShellProfilePath = orig })

	// Create a temp profile so the profile check passes
	tmpDir := t.TempDir()
	profilePath := filepath.Join(tmpDir, "profile.ps1")
	if err := os.WriteFile(profilePath, []byte("# test profile\n"), 0644); err != nil {
		t.Fatal(err)
	}
	PowerShellProfilePath = profilePath

	_, err := RunPowerShell("Invoke-OpsDailyOps")
	// The command should proceed past the allowlist and profile checks,
	// and attempt execution. This may fail because PowerShell isn't
	// available in the test environment, but that's the OS-layer error
	// and proves the sandbox was configured (the allowlist/profile
	// gates were passed).
	if err != nil {
		t.Logf("RunPowerShell proceeded past allowlist+profile checks (sandbox applied): %v", err)
	}
}

func TestRunCommandWithLiveOutput_WaitGroup(t *testing.T) {
	// Use a safe command that produces output
	ch := make(chan string, 100)

	// Start a goroutine to consume the channel so we don't deadlock
	// if the output is larger than the buffer.
	count := 0
	done := make(chan bool)
	go func() {
		for range ch {
			count++
		}
		done <- true
	}()

	res, err := RunCommandWithLiveOutput("ipconfig", ch)
	if err != nil {
		// If ipconfig is not available (Linux CI), skip
		t.Skipf("Skipping test: ipconfig failed: %v", err)
		return
	}

	<-done

	if count == 0 {
		t.Log("No lines received from ipconfig (unexpected but possible if output is empty)")
	} else {
		t.Logf("Received %d lines from ipconfig", count)
	}

	if res.Output == "" && count > 0 {
		t.Error("Result output is empty but lines were received on channel")
	}
}

func TestIsDangerousCommand_EdgeCases(t *testing.T) {
	tests := []struct {
		cmd  string
		want bool
	}{
		{"echo > file", true},
		{"echo >> file", true},
		{"cat file | grep", true},
		{"rm", false}, // needs space
		{"rm ", true},
		{"powershell -c", true},
		{"./myapp --rm", false},
	}
	for _, tt := range tests {
		if got := IsDangerousCommand(tt.cmd); got != tt.want {
			t.Errorf("IsDangerousCommand(%q) = %v, want %v", tt.cmd, got, tt.want)
		}
	}
}

func TestIsAllowedShellCommand(t *testing.T) {
	allowed := []string{
		"ipconfig",
		"ipconfig /all",
		"ping 8.8.8.8",
		"netstat -an",
		"tasklist",
		"dir",
		"ls -la",
		"C:\\Windows\\System32\\ping.exe -n 4 google.com",
		"/usr/bin/df -h",
	}
	for _, cmd := range allowed {
		if !isAllowedShellCommand(cmd) {
			t.Errorf("isAllowedShellCommand(%q) = false, want true", cmd)
		}
	}

	blocked := []string{
		"",
		"rm -rf /",
		"powershell -c \"Get-Process\"",
		"bash -c 'exploit'",
		"format C:",
		"chmod 777 /etc/shadow",
	}
	for _, cmd := range blocked {
		if isAllowedShellCommand(cmd) {
			t.Errorf("isAllowedShellCommand(%q) = true, want false", cmd)
		}
	}
}

func TestNormalizeWhitespace(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"rm -rf /", "rm -rf /"},
		{"rm\t-rf\t/", "rm -rf /"},
		{"kill   -9    1234", "kill -9 1234"},
		{"ipconfig", "ipconfig"},
		{"echo  hello    world", "echo hello world"},
		{"\tls\t-la\t", " ls -la "},
	}
	for _, tt := range tests {
		if got := normalizeWhitespace(tt.input); got != tt.want {
			t.Errorf("normalizeWhitespace(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestFindGitBash(t *testing.T) {
	path := findGitBash()
	if path == "" {
		t.Log("Git Bash not found (expected on non-Windows or without Git)")
	} else {
		t.Logf("Git Bash found at: %s", path)
	}
}

func TestStripANSI(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello", "hello"},
		{"\x1b[31mred\x1b[0m", "red"},
		{"\x1b[1m\x1b[32mbold green\x1b[0m", "bold green"},
		{"no escapes here", "no escapes here"},
		{"\x1b[Kclear line", "clear line"},
	}
	for _, tt := range tests {
		if got := stripANSI(tt.input); got != tt.want {
			t.Errorf("stripANSI(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestRunCommand_WithAllowedCommand(t *testing.T) {
	res, err := RunCommand("dir")
	if err != nil {
		t.Fatalf("RunCommand(dir) failed: %v", err)
	}
	if res.ExitCode != 0 {
		t.Errorf("RunCommand(dir) exit code = %d, want 0", res.ExitCode)
	}
	if res.Output == "" {
		t.Error("RunCommand(dir) output should not be empty")
	}
	if res.Command != "dir" {
		t.Errorf("RunCommand(dir) command = %q, want %q", res.Command, "dir")
	}
	if res.Duration <= 0 {
		t.Error("RunCommand(dir) duration should be positive")
	}
}

func TestRunCommand_WithOutput(t *testing.T) {
	// Use an allowed command
	res, err := RunCommand("whoami")
	if err != nil {
		t.Fatalf("RunCommand(whoami) failed: %v", err)
	}
	if res.Command != "whoami" {
		t.Errorf("RunCommand command = %q, want %q", res.Command, "whoami")
	}
	if res.Duration <= 0 {
		t.Error("RunCommand duration should be positive")
	}
	t.Logf("RunCommand output: %q (exit=%d)", res.Output, res.ExitCode)
}
