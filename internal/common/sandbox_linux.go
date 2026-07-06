package common

import (
	"os/exec"
	"syscall"
)

// applyPlatformSandbox applies Linux sandbox restrictions using namespaces.
func applyPlatformSandbox(cmd *exec.Cmd, cfg SandboxConfig) {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	sa := cmd.SysProcAttr

	if cfg.DenyNetworkAccess {
		sa.Cloneflags |= syscall.CLONE_NEWNET
	}
	if cfg.ReadOnlyFS {
		sa.Cloneflags |= syscall.CLONE_NEWNS
	}
	if cfg.DenyProcessSpawn {
		sa.Cloneflags |= syscall.CLONE_NEWPID
	}
	if cfg.DropPrivileges {
		sa.Cloneflags |= syscall.CLONE_NEWUSER
	}
}
