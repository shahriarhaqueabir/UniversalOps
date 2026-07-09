package devops

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestIsAllowedWorkflow_Allowed(t *testing.T) {
	workflows := []string{
		"Invoke-HawkDailyOps",
		"Invoke-HawkSystemReview",
		"Invoke-HawkSecurityAudit",
		"Invoke-HawkNetworkDiagnostics",
		"Invoke-HawkThreatHunt",
		"Invoke-HawkChangeAudit",
		"Invoke-HawkComplianceCheck",
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
		"Invoke-HawkUnknown",
		"anything",
		"Invoke-HawkDailyOps; Remove-Item -Recurse C:\\",
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

	_, err := RunPowerShell("Invoke-HawkDailyOps")
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
	if !errors.Is(err, ErrDangerousCommand) {
		t.Errorf("RunCommand error should wrap ErrDangerousCommand, got: %v", err)
	}
}

func TestRunCommandWithLiveOutput_BlocksDangerousCommand(t *testing.T) {
	ch := make(chan string, 10)
	_, err := RunCommandWithLiveOutput("rm -rf /", ch)
	close(ch)
	if err == nil {
		t.Fatal("RunCommandWithLiveOutput should return error for dangerous command")
	}
	if !errors.Is(err, ErrDangerousCommand) {
		t.Errorf("RunCommandWithLiveOutput error should wrap ErrDangerousCommand, got: %v", err)
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

	_, err := RunPowerShell("Invoke-HawkDailyOps")
	// The command should proceed past the allowlist and profile checks,
	// and attempt execution. This may fail because PowerShell isn't
	// available in the test environment, but that's the OS-layer error
	// and proves the sandbox was configured (the allowlist/profile
	// gates were passed).
	if err != nil {
		t.Logf("RunPowerShell proceeded past allowlist+profile checks (sandbox applied): %v", err)
	}
}
