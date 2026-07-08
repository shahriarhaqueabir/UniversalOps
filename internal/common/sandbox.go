package common

import (
	"context"
	"os/exec"
)

// SandboxConfig controls the sandbox restrictions on a command execution.
type SandboxConfig struct {
	DenyNetworkAccess bool
	ReadOnlyFS        bool
	DenyProcessSpawn  bool
	WorkingDir        string
	DropPrivileges    bool
}

// DefaultSandbox returns a SandboxConfig with standard restrictions.
// Denies network access, read-only filesystem, denies sub-process spawning,
// and drops privileges. Suitable for user-invoked DevOps shell commands.
func DefaultSandbox() SandboxConfig {
	return SandboxConfig{
		DenyNetworkAccess: true,
		ReadOnlyFS:        true,
		DenyProcessSpawn:  true,
		DropPrivileges:    true,
	}
}

// SystemQuerySandbox returns a SandboxConfig for read-only system queries.
// Suitable for secops, sysops, and diagnostic commands that invoke
// OS tools (netsh, netstat, powershell) but don't need network access.
// DenyProcessSpawn is false because these tools are themselves sub-processes.
func SystemQuerySandbox() SandboxConfig {
	return SandboxConfig{
		DenyNetworkAccess: true,
		ReadOnlyFS:        true,
		DenyProcessSpawn:  false,
		DropPrivileges:    true,
	}
}

// SandboxedCommand creates an exec.Cmd with sandbox restrictions.
func SandboxedCommand(name string, args ...string) *exec.Cmd {
	cmd := exec.Command(name, args...)
	applySandbox(cmd, DefaultSandbox())
	return cmd
}

// SandboxedCommandWithConfig creates an exec.Cmd with custom sandbox config.
func SandboxedCommandWithConfig(cfg SandboxConfig, name string, args ...string) *exec.Cmd {
	cmd := exec.Command(name, args...)
	applySandbox(cmd, cfg)
	return cmd
}

// SandboxedCommandWithConfigContext creates an exec.Cmd with sandbox restrictions
// that respects the given context for cancellation/deadlines.
func SandboxedCommandWithConfigContext(ctx context.Context, cfg SandboxConfig, name string, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, name, args...)
	applySandbox(cmd, cfg)
	return cmd
}

func applySandbox(cmd *exec.Cmd, cfg SandboxConfig) {
	if cfg.WorkingDir != "" {
		cmd.Dir = cfg.WorkingDir
	}
	applyPlatformSandbox(cmd, cfg)
}
