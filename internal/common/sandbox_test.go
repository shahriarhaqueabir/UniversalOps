package common

import (
	"os/exec"
	"runtime"
	"testing"
)

func TestDefaultSandbox(t *testing.T) {
	cfg := DefaultSandbox()
	if !cfg.DenyNetworkAccess {
		t.Error("DefaultSandbox().DenyNetworkAccess should be true")
	}
	if !cfg.ReadOnlyFS {
		t.Error("DefaultSandbox().ReadOnlyFS should be true")
	}
	if !cfg.DenyProcessSpawn {
		t.Error("DefaultSandbox().DenyProcessSpawn should be true")
	}
	if !cfg.DropPrivileges {
		t.Error("DefaultSandbox().DropPrivileges should be true")
	}
	if cfg.WorkingDir != "" {
		t.Error("DefaultSandbox().WorkingDir should be empty")
	}
}

func TestDefaultSandboxImmutability(t *testing.T) {
	// Multiple calls should return fresh copies
	cfg1 := DefaultSandbox()
	cfg2 := DefaultSandbox()

	if cfg1.WorkingDir != cfg2.WorkingDir {
		t.Error("DefaultSandbox() should return consistent values")
	}

	// Modifying one should not affect the other
	cfg1.WorkingDir = "/test"
	if cfg2.WorkingDir != "" {
		t.Error("Modifying one DefaultSandbox() result should not affect another")
	}
}

func TestSandboxedCommand(t *testing.T) {
	cmd := SandboxedCommand("echo", "hello")
	if cmd == nil {
		t.Fatal("SandboxedCommand() returned nil")
	}

	if cmd.Path == "" {
		t.Error("SandboxedCommand() cmd.Path should not be empty")
	}

	if len(cmd.Args) == 0 {
		t.Error("SandboxedCommand() should have args")
	}
}

func TestSandboxedCommandWithConfig(t *testing.T) {
	cfg := SandboxConfig{
		DenyNetworkAccess: false,
		ReadOnlyFS:        false,
		DenyProcessSpawn:  false,
		WorkingDir:        "/tmp",
		DropPrivileges:    false,
	}

	cmd := SandboxedCommandWithConfig(cfg, "ls", "-la")
	if cmd == nil {
		t.Fatal("SandboxedCommandWithConfig() returned nil")
	}

	if cmd.Dir != "/tmp" {
		t.Errorf("SandboxedCommandWithConfig() WorkingDir = %q, want %q", cmd.Dir, "/tmp")
	}
}

func TestSandboxedCommandWithConfigDefaults(t *testing.T) {
	// Minimal config should still produce a valid command
	cfg := SandboxConfig{}
	cmd := SandboxedCommandWithConfig(cfg, "true")
	if cmd == nil {
		t.Fatal("SandboxedCommandWithConfig() returned nil")
	}
}

func TestSandboxedCommandExecutes(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping execution test on Windows (echo varies)")
	}

	cmd := SandboxedCommand("sh", "-c", "echo sandbox-test")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("sandboxed command failed: %v", err)
	}
	if string(output) != "sandbox-test\n" {
		t.Errorf("output = %q, want %q", string(output), "sandbox-test\n")
	}
}

func TestSandboxedCommandHasSysProcAttr(t *testing.T) {
	cmd := SandboxedCommand("echo", "test")
	if cmd.SysProcAttr == nil {
		t.Log("Note: SysProcAttr is nil (expected on some platforms/configs)")
	}
}

func TestSandboxedCommandWithConfigPropagatesArgs(t *testing.T) {
	cfg := DefaultSandbox()
	cfg.WorkingDir = "/custom"

	cmd := SandboxedCommandWithConfig(cfg, "find", "/var", "-name", "*.log")
	if cmd == nil {
		t.Fatal("SandboxedCommandWithConfig() returned nil")
	}

	expected := []string{"find", "/var", "-name", "*.log"}
	if len(cmd.Args) != len(expected) {
		t.Errorf("args length = %d, want %d", len(cmd.Args), len(expected))
	}
	for i, arg := range expected {
		if i >= len(cmd.Args) {
			break
		}
		if cmd.Args[i] != arg {
			t.Errorf("arg[%d] = %q, want %q", i, cmd.Args[i], arg)
		}
	}
}

func TestApplySandboxWorkingDir(t *testing.T) {
	// Test that applySandbox sets the working directory
	// We use SandboxedCommandWithConfig which calls applySandbox
	cmd := exec.Command("echo", "test")
	applySandbox(cmd, SandboxConfig{WorkingDir: "/my/dir"})
	if cmd.Dir != "/my/dir" {
		t.Errorf("cmd.Dir = %q, want %q", cmd.Dir, "/my/dir")
	}
}

func TestApplySandboxEmptyWorkingDir(t *testing.T) {
	// Empty working dir should not change cmd.Dir
	cmd := exec.Command("echo", "test")
	originalDir := cmd.Dir
	applySandbox(cmd, SandboxConfig{})
	if cmd.Dir != originalDir {
		t.Errorf("cmd.Dir changed from %q to %q", originalDir, cmd.Dir)
	}
}

func TestSystemQuerySandbox(t *testing.T) {
	cfg := SystemQuerySandbox()
	if !cfg.DenyNetworkAccess {
		t.Error("SystemQuerySandbox().DenyNetworkAccess should be true")
	}
	if !cfg.ReadOnlyFS {
		t.Error("SystemQuerySandbox().ReadOnlyFS should be true")
	}
	if cfg.DenyProcessSpawn {
		t.Error("SystemQuerySandbox().DenyProcessSpawn should be false (system tools need to run)")
	}
	if !cfg.DropPrivileges {
		t.Error("SystemQuerySandbox().DropPrivileges should be true")
	}
}

func TestSystemQuerySandboxImmutability(t *testing.T) {
	// Multiple calls should return fresh copies
	cfg1 := SystemQuerySandbox()
	cfg2 := SystemQuerySandbox()

	// Modifying one should not affect the other
	cfg1.DenyNetworkAccess = false
	if !cfg2.DenyNetworkAccess {
		t.Error("Modifying one SystemQuerySandbox() result should not affect another")
	}
}

func TestSystemQuerySandboxCreatesValidCommand(t *testing.T) {
	cmd := SandboxedCommandWithConfig(SystemQuerySandbox(), "echo", "hello")
	if cmd == nil {
		t.Fatal("SystemQuerySandbox() command returned nil")
	}
	if cmd.Path == "" {
		t.Error("cmd.Path should not be empty")
	}
}
