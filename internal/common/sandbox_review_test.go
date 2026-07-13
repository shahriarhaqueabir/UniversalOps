//go:build windows

package common

import (
	"testing"
	"golang.org/x/sys/windows"
)

func TestReadOnlyFSWithoutDropPrivileges(t *testing.T) {
	cfg := SandboxConfig{
		ReadOnlyFS: true,
		DropPrivileges: false,
	}

	cmd := SandboxedCommandWithConfig(cfg, "cmd", "/c", "echo test")

	if cmd.SysProcAttr == nil {
		t.Fatal("SysProcAttr is nil")
	}

	if cmd.SysProcAttr.Token == 0 {
		t.Error("Token should NOT be 0 when ReadOnlyFS is true, even if DropPrivileges is false")
	} else {
		t.Logf("Token is %v", cmd.SysProcAttr.Token)
	}
}

func TestIntegrityLevelNotLeakedToParent(t *testing.T) {
	// Simple test: ensure we can still perform a Medium Integrity operation
	// (like opening a file for writing in a standard location) after
	// creating a Low Integrity sandbox.

	cfg := SandboxConfig{
		ReadOnlyFS: true,
		DropPrivileges: false,
	}

	// Create sandbox (previously this would jail the parent)
	_ = SandboxedCommandWithConfig(cfg, "cmd", "/c", "echo test")

	// Try a Medium IL operation in the parent
	// For example, calling OpenProcessToken with TOKEN_ADJUST_DEFAULT
	var token windows.Token
	err := windows.OpenProcessToken(windows.CurrentProcess(), windows.TOKEN_ADJUST_DEFAULT, &token)
	if err != nil {
		t.Errorf("Parent process seems jailed! Cannot OpenProcessToken: %v", err)
	} else {
		token.Close()
	}
}
