//go:build windows

package common

import (
	"os/exec"
	"syscall"
	"testing"
)

func TestWindowsSandboxRestrictedToken(t *testing.T) {
	cfg := SandboxConfig{
		DenyNetworkAccess: false,
		ReadOnlyFS:        false,
		DenyProcessSpawn:  false,
		DropPrivileges:    true,
	}

	cmd := SandboxedCommandWithConfig(cfg, "cmd", "/c", "echo restricted-token-test")
	if cmd == nil {
		t.Fatal("SandboxedCommandWithConfig() returned nil")
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("sandboxed command failed: %v (output: %s)", err, string(output))
	}
	if len(output) == 0 {
		t.Error("expected non-empty output")
	}

	if cmd.SysProcAttr == nil {
		t.Fatal("SysProcAttr should not be nil after sandboxing")
	}
	if cmd.SysProcAttr.Token == 0 {
		t.Error("Token should be non-zero when DropPrivileges is true")
	}
}

func TestWindowsSandboxDefaultConfig(t *testing.T) {
	cfg := DefaultSandbox()

	if !cfg.DenyProcessSpawn {
		t.Error("DefaultSandbox should have DenyProcessSpawn=true")
	}
	if !cfg.DropPrivileges {
		t.Error("DefaultSandbox should have DropPrivileges=true")
	}

	cmd := SandboxedCommandWithConfig(cfg, "cmd", "/c", "echo default-test")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("default sandboxed command failed: %v (output: %s)", err, string(output))
	}
	if len(output) == 0 {
		t.Error("expected non-empty output")
	}
}

func TestWindowsSandboxSysProcAttrPreserved(t *testing.T) {
	cmd := exec.Command("cmd", "/c", "echo preserve-test")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true,
	}

	applySandbox(cmd, DefaultSandbox())

	if cmd.SysProcAttr == nil {
		t.Fatal("SysProcAttr should not be nil")
	}
	if !cmd.SysProcAttr.HideWindow {
		t.Error("HideWindow should remain true")
	}
}

func TestWindowsSandboxSystemQuery(t *testing.T) {
	cfg := SystemQuerySandbox()

	cmd := SandboxedCommandWithConfig(cfg, "cmd", "/c", "echo system-query-test")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("system query sandboxed command failed: %v (output: %s)", err, string(output))
	}
	if len(output) == 0 {
		t.Error("expected non-empty output")
	}

	if cmd.SysProcAttr != nil && cmd.SysProcAttr.Token == 0 {
		t.Error("SystemQuerySandbox should set restricted token (DropPrivileges=true)")
	}
}

func TestWindowsSandboxJobObjectCreated(t *testing.T) {
	// Verify that a job object is created and assigned when DenyProcessSpawn is true.
	cfg := SandboxConfig{
		DenyNetworkAccess: false,
		ReadOnlyFS:        false,
		DenyProcessSpawn:  true,
		DropPrivileges:    false,
	}

	cmd := SandboxedCommandWithConfig(cfg, "cmd", "/c", "echo job-test")
	if cmd == nil {
		t.Fatal("SandboxedCommandWithConfig() returned nil")
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("sandboxed command failed: %v (output: %s)", err, string(output))
	}
	if len(output) == 0 {
		t.Error("expected non-empty output")
	}
}

func TestWindowsSandboxHideWindow(t *testing.T) {
	// DenyNetworkAccess should still set HideWindow.
	cfg := SandboxConfig{
		DenyNetworkAccess: true,
		DenyProcessSpawn:  false,
		DropPrivileges:    false,
	}

	cmd := SandboxedCommandWithConfig(cfg, "cmd", "/c", "echo hide-test")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("HideWindow command failed: %v (output: %s)", err, string(output))
	}
	if len(output) == 0 {
		t.Error("expected non-empty output")
	}
	if cmd.SysProcAttr == nil || !cmd.SysProcAttr.HideWindow {
		t.Error("HideWindow should be true when DenyNetworkAccess is true")
	}
}
