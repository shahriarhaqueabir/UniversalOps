package common

import (
	"context"
	"errors"
	"os/exec"
)

// errStdoutSet is returned when Stdout is already set before calling CombinedOutput/Output.
var errStdoutSet = errors.New("exec: Stdout already set")

// errStderrSet is returned when Stderr is already set before calling CombinedOutput.
var errStderrSet = errors.New("exec: Stderr already set")

// SandboxConfig controls the sandbox restrictions on a command execution.
type SandboxConfig struct {
	DenyNetworkAccess bool
	ReadOnlyFS        bool
	DenyProcessSpawn  bool
	WorkingDir        string
	DropPrivileges    bool
}

// SandboxedCmd wraps exec.Cmd with platform sandbox restrictions.
type SandboxedCmd struct {
	*exec.Cmd
	cfg SandboxConfig
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

// SandboxedCommand creates a sandboxed command with default restrictions.
func SandboxedCommand(name string, args ...string) *SandboxedCmd {
	cmd := exec.Command(name, args...)
	return applySandbox(cmd, DefaultSandbox())
}

// SandboxedCommandWithConfig creates a sandboxed command with custom config.
func SandboxedCommandWithConfig(cfg SandboxConfig, name string, args ...string) *SandboxedCmd {
	cmd := exec.Command(name, args...)
	return applySandbox(cmd, cfg)
}

// SandboxedCommandWithConfigContext creates a sandboxed command with custom config
// that respects the given context for cancellation/deadlines.
func SandboxedCommandWithConfigContext(ctx context.Context, cfg SandboxConfig, name string, args ...string) *SandboxedCmd {
	cmd := exec.CommandContext(ctx, name, args...)
	return applySandbox(cmd, cfg)
}

func applySandbox(cmd *exec.Cmd, cfg SandboxConfig) *SandboxedCmd {
	if cfg.WorkingDir != "" {
		cmd.Dir = cfg.WorkingDir
	}
	return applyPlatformSandbox(cmd, cfg)
}
