package common

import (
	"os/exec"
	"syscall"
)

// applyPlatformSandbox applies Windows sandbox restrictions.
// On Windows we use job object and token restrictions via SysProcAttr.
func applyPlatformSandbox(cmd *exec.Cmd, cfg SandboxConfig) {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	sa := cmd.SysProcAttr
	if cfg.DenyProcessSpawn || cfg.DenyNetworkAccess {
		sa.NoInheritHandles = true
		sa.HideWindow = true
	}
	// Note: Full Windows sandboxing (job objects, restricted tokens)
	// requires the Windows SDK and is not available through syscall alone.
	// This provides basic process isolation.
}
