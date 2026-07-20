package common

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os/exec"
	"regexp"
	"strconv"
)

// errStdoutSet is returned when Stdout is already set before calling CombinedOutput/Output.
var errStdoutSet = errors.New("exec: Stdout already set")

// errStderrSet is returned when Stderr is already set before calling CombinedOutput.
var errStderrSet = errors.New("exec: Stderr already set")

// Interface name validation regex: alphanumeric, dash, underscore, dot
var interfaceNameRe = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

// ValidInterfaceName validates a network interface name.
func ValidInterfaceName(name string) bool {
	if len(name) == 0 || len(name) > 64 {
		return false
	}
	return interfaceNameRe.MatchString(name)
}

// Username validation regex: alphanumeric, dot, underscore, dash
var usernameRe = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

// ValidUsername validates a system username.
func ValidUsername(name string) bool {
	if len(name) == 0 || len(name) > 256 {
		return false
	}
	return usernameRe.MatchString(name)
}

// ValidServiceName validates a Windows/Linux service name.
func ValidServiceName(name string) bool {
	if len(name) == 0 || len(name) > 256 {
		return false
	}
	// Service names: alphanumeric, dash, underscore, dot
	for _, r := range name {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' || r == '.') {
			return false
		}
	}
	return true
}

// ValidPID validates and parses a process ID string.
func ValidPID(pid string) (int, error) {
	if len(pid) == 0 || len(pid) > 10 {
		return 0, fmt.Errorf("invalid PID: %q", pid)
	}
	// Check all digits
	for _, r := range pid {
		if r < '0' || r > '9' {
			return 0, fmt.Errorf("invalid PID: %q", pid)
		}
	}
	val, err := strconv.Atoi(pid)
	if err != nil || val <= 0 {
		return 0, fmt.Errorf("invalid PID: %q", pid)
	}
	return val, nil
}

// ValidIP validates an IP address (IPv4 or IPv6).
func ValidIP(ip string) bool {
	if len(ip) == 0 {
		return false
	}
	return net.ParseIP(ip) != nil
}

// Firewall rule name validation: alphanumeric, dash, underscore, space
var firewallRuleNameRe = regexp.MustCompile(`^[a-zA-Z0-9\-_ ]+$`)

// ValidFirewallRuleName validates a firewall rule name.
func ValidFirewallRuleName(name string) bool {
	if len(name) == 0 || len(name) > 256 {
		return false
	}
	return firewallRuleNameRe.MatchString(name)
}

// Docker container ID validation: hex characters, 1-64 chars
var containerIDRe = regexp.MustCompile(`^[a-fA-F0-9]{1,64}$`)

// ValidContainerID validates a Docker container ID or name.
// Accepts hex container IDs (up to 64 chars) or container names (alphanumeric, dash, underscore, dot, slash).
func ValidContainerID(id string) bool {
	if len(id) == 0 || len(id) > 128 {
		return false
	}
	// Hex container ID
	if containerIDRe.MatchString(id) {
		return true
	}
	// Container name: alphanumeric, dash, underscore, dot, slash (for compose)
	nameRe := regexp.MustCompile(`^[a-zA-Z0-9._/-]+$`)
	return nameRe.MatchString(id)
}

// K8s namespace/resource-type validation: alphanumeric, dash, underscore, dot
var k8sNameRe = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

// ValidK8sName validates a Kubernetes namespace or resource type name.
func ValidK8sName(name string) bool {
	if len(name) == 0 || len(name) > 253 {
		return false
	}
	return k8sNameRe.MatchString(name)
}

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
	cfg := SandboxConfig{
		DenyNetworkAccess: true,
		ReadOnlyFS:        true,
		DenyProcessSpawn:  false,
		DropPrivileges:    true,
	}
	// On Windows, ReadOnlyFS (implemented via Low Integrity Level)
	// breaks common system tools like ping and netsh because they
	// require access to certain registry keys and local pipes that
	// are restricted at LIL.
	// TODO: Explore using AppContainer SIDs for finer-grained access
	// that allows ping/netsh while still restricting general FS writes.
	if IsWindows() {
		cfg.ReadOnlyFS = false
	}
	return cfg
}

// RemediationSandbox returns a SandboxConfig for system remediation actions.
// Allows process spawning, denies network access, and keeps privileges
// (required for IR actions like taskkill or firewall changes).
func RemediationSandbox() SandboxConfig {
	return SandboxConfig{
		DenyNetworkAccess: true,
		ReadOnlyFS:        false,
		DenyProcessSpawn:  false,
		DropPrivileges:    false,
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
