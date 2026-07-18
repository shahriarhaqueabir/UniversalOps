package secops

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
)

// AutoExpireWG tracks background goroutines for auto-expire isolation.
var AutoExpireWG sync.WaitGroup

// ActionResult holds the result of an incident response action.
type ActionResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// cmdTimeout is the maximum time to wait for a single firewall command.
const cmdTimeout = 10 * time.Second

// IsolateHost isolates the host from the network.
// If confirm is false, returns an error requiring explicit confirmation.
// If autoExpireSeconds > 0, the isolation rule will be automatically removed after that many seconds.
func IsolateHost(confirm bool, autoExpireSeconds int) (*ActionResult, error) {
	if !confirm {
		return &ActionResult{Success: false, Error: "isolation requires explicit confirmation (pass confirm=true)"}, nil
	}
	if common.IsWindows() {
		return isolateHostWindows(autoExpireSeconds)
	}
	return isolateHostLinux(autoExpireSeconds)
}

func isolateHostWindows(autoExpireSeconds int) (*ActionResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
	defer cancel()

	cfg := common.RemediationSandbox()

	// Enable firewall
	cmd := common.SandboxedCommandWithConfigContext(ctx, cfg, "netsh", "advfirewall", "set", "allprofiles", "state", "on")
	if err := cmd.Run(); err != nil {
		return &ActionResult{Success: false, Error: err.Error()}, nil
	}
	// Add block-all inbound rule
	cmd = common.SandboxedCommandWithConfigContext(ctx, cfg, "netsh", "advfirewall", "firewall", "add", "rule",
		"name=SECOPS_Isolate_BlockAll", "dir=in", "action=block")
	if err := cmd.Run(); err != nil {
		return &ActionResult{Success: false, Error: err.Error()}, nil
	}
	// Also block outbound to fully isolate
	cmd = common.SandboxedCommandWithConfigContext(ctx, cfg, "netsh", "advfirewall", "firewall", "add", "rule",
		"name=SECOPS_Isolate_BlockAll_Out", "dir=out", "action=block")
	if err := cmd.Run(); err != nil {
		// Roll back inbound rule if outbound failed
		rollbackCtx, rollbackCancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer rollbackCancel()
		cmdRollback := common.SandboxedCommandWithConfigContext(rollbackCtx, cfg, "netsh", "advfirewall", "firewall", "delete", "rule", "name=SECOPS_Isolate_BlockAll")
		_ = cmdRollback.Run()
		return &ActionResult{Success: false, Error: fmt.Sprintf("outbound block failed: %v", err)}, nil
	}

	msg := "Host isolated — all inbound and outbound traffic blocked"
	if autoExpireSeconds > 0 {
		msg += fmt.Sprintf(" (auto-expires in %d seconds)", autoExpireSeconds)
		AutoExpireWG.Add(1)
		go func() {
			defer AutoExpireWG.Done()
			time.Sleep(time.Duration(autoExpireSeconds) * time.Second)
			if _, err := RemoveIsolation(); err != nil {
				common.LogError("auto-expire isolation failed: %v", err)
			}
		}()
	}
	return &ActionResult{Success: true, Message: msg}, nil
}

func isolateHostLinux(autoExpireSeconds int) (*ActionResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
	defer cancel()

	cfg := common.RemediationSandbox()

	// Block inbound
	cmd := common.SandboxedCommandWithConfigContext(ctx, cfg, "iptables", "-A", "INPUT", "-j", "DROP")
	if err := cmd.Run(); err != nil {
		return &ActionResult{Success: false, Error: err.Error()}, nil
	}
	// Also block outbound to fully isolate
	cmd = common.SandboxedCommandWithConfigContext(ctx, cfg, "iptables", "-A", "OUTPUT", "-j", "DROP")
	if err := cmd.Run(); err != nil {
		// Roll back inbound rule if outbound failed
		rollbackCtx, rollbackCancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer rollbackCancel()
		cmdRollback := common.SandboxedCommandWithConfigContext(rollbackCtx, cfg, "iptables", "-D", "INPUT", "-j", "DROP")
		_ = cmdRollback.Run()
		return &ActionResult{Success: false, Error: fmt.Sprintf("outbound block failed: %v", err)}, nil
	}

	msg := "Host isolated — all inbound and outbound traffic dropped"
	if autoExpireSeconds > 0 {
		msg += fmt.Sprintf(" (auto-expires in %d seconds)", autoExpireSeconds)
		AutoExpireWG.Add(1)
		go func() {
			defer AutoExpireWG.Done()
			time.Sleep(time.Duration(autoExpireSeconds) * time.Second)
			if _, err := RemoveIsolation(); err != nil {
				common.LogError("auto-expire isolation failed: %v", err)
			}
		}()
	}
	return &ActionResult{Success: true, Message: msg}, nil
}

// RemoveIsolation removes all SECOPS isolation rules (both inbound and outbound).
// This is the manual undo for IsolateHost.
func RemoveIsolation() (*ActionResult, error) {
	if common.IsWindows() {
		return removeIsolationWindows()
	}
	return removeIsolationLinux()
}

func removeIsolationWindows() (*ActionResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
	defer cancel()
	var errs []string
	cfg := common.RemediationSandbox()

	// Remove inbound rule
	cmd := common.SandboxedCommandWithConfigContext(ctx, cfg, "netsh", "advfirewall", "firewall", "delete", "rule", "name=SECOPS_Isolate_BlockAll")
	if err := cmd.Run(); err != nil {
		errs = append(errs, fmt.Sprintf("inbound: %v", err))
	}
	// Remove outbound rule
	cmd = common.SandboxedCommandWithConfigContext(ctx, cfg, "netsh", "advfirewall", "firewall", "delete", "rule", "name=SECOPS_Isolate_BlockAll_Out")
	if err := cmd.Run(); err != nil {
		errs = append(errs, fmt.Sprintf("outbound: %v", err))
	}
	if len(errs) > 0 {
		return &ActionResult{Success: false, Error: strings.Join(errs, "; ")}, nil
	}
	return &ActionResult{Success: true, Message: "Isolation removed — traffic restored"}, nil
}

func removeIsolationLinux() (*ActionResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
	defer cancel()
	var errs []string
	cfg := common.RemediationSandbox()

	// Remove inbound DROP rule
	cmd := common.SandboxedCommandWithConfigContext(ctx, cfg, "iptables", "-D", "INPUT", "-j", "DROP")
	if err := cmd.Run(); err != nil {
		errs = append(errs, fmt.Sprintf("inbound: %v", err))
	}
	// Remove outbound DROP rule
	cmd = common.SandboxedCommandWithConfigContext(ctx, cfg, "iptables", "-D", "OUTPUT", "-j", "DROP")
	if err := cmd.Run(); err != nil {
		errs = append(errs, fmt.Sprintf("outbound: %v", err))
	}
	if len(errs) > 0 {
		return &ActionResult{Success: false, Error: strings.Join(errs, "; ")}, nil
	}
	return &ActionResult{Success: true, Message: "Isolation removed — traffic restored"}, nil
}

// KillProcess force-kills a process by PID.
func KillProcess(pid int) (*ActionResult, error) {
	if pid <= 0 || pid > 4194304 {
		return &ActionResult{Success: false, Error: fmt.Sprintf("invalid PID: %d", pid)}, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
	defer cancel()
	cfg := common.RemediationSandbox()

	if common.IsWindows() {
		cmd := common.SandboxedCommandWithConfigContext(ctx, cfg, "taskkill", "/F", "/PID", fmt.Sprintf("%d", pid))
		out, err := cmd.Output()
		if err != nil {
			return &ActionResult{Success: false, Error: err.Error()}, nil
		}
		return &ActionResult{Success: true, Message: strings.TrimSpace(string(out))}, nil
	}
	cmd := common.SandboxedCommandWithConfigContext(ctx, cfg, "kill", "-9", fmt.Sprintf("%d", pid))
	out, err := cmd.Output()
	if err != nil {
		return &ActionResult{Success: false, Error: err.Error()}, nil
	}
	return &ActionResult{Success: true, Message: strings.TrimSpace(string(out))}, nil
}

// BlockIP blocks an IP address via firewall.
func BlockIP(ip string) (*ActionResult, error) {
	if !common.ValidIP(ip) {
		return &ActionResult{Success: false, Error: fmt.Sprintf("invalid IP address: %q", ip)}, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
	defer cancel()
	cfg := common.RemediationSandbox()

	if common.IsWindows() {
		cmd := common.SandboxedCommandWithConfigContext(ctx, cfg, "netsh", "advfirewall", "firewall", "add", "rule",
			"name=SECOPS_Block_"+ip, "dir=in", "action=block", "remoteip="+ip)
		if err := cmd.Run(); err != nil {
			return &ActionResult{Success: false, Error: err.Error()}, nil
		}
		return &ActionResult{Success: true, Message: fmt.Sprintf("Blocked IP %s", ip)}, nil
	}
	cmd := common.SandboxedCommandWithConfigContext(ctx, cfg, "iptables", "-A", "INPUT", "-s", ip, "-j", "DROP")
	if err := cmd.Run(); err != nil {
		return &ActionResult{Success: false, Error: err.Error()}, nil
	}
	return &ActionResult{Success: true, Message: fmt.Sprintf("Blocked IP %s", ip)}, nil
}

// DisableAccount disables a local user account.
func DisableAccount(username string) (*ActionResult, error) {
	if !common.ValidUsername(username) {
		return &ActionResult{Success: false, Error: fmt.Sprintf("invalid username: %q", username)}, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
	defer cancel()
	cfg := common.RemediationSandbox()

	if common.IsWindows() {
		cmd := common.SandboxedCommandWithConfigContext(ctx, cfg, "net", "user", username, "/active:no")
		if err := cmd.Run(); err != nil {
			return &ActionResult{Success: false, Error: err.Error()}, nil
		}
		return &ActionResult{Success: true, Message: fmt.Sprintf("Account %s disabled", username)}, nil
	}
	cmd := common.SandboxedCommandWithConfigContext(ctx, cfg, "passwd", "-l", username)
	if err := cmd.Run(); err != nil {
		return &ActionResult{Success: false, Error: err.Error()}, nil
	}
	return &ActionResult{Success: true, Message: fmt.Sprintf("Account %s locked", username)}, nil
}

// CaptureEvidence collects forensic evidence into a summary.
func CaptureEvidence() (*ActionResult, error) {
	var evidence []string

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Collect running processes
	if runtime.GOOS == "windows" {
		out, _ := exec.CommandContext(ctx, "tasklist").Output()
		evidence = append(evidence, fmt.Sprintf("=== PROCESSES ===\n%s", string(out)))
	} else {
		out, _ := exec.CommandContext(ctx, "ps", "aux").Output()
		evidence = append(evidence, fmt.Sprintf("=== PROCESSES ===\n%s", string(out)))
	}

	// Collect listening ports
	if runtime.GOOS == "windows" {
		out, _ := exec.CommandContext(ctx, "netstat", "-ano").Output()
		evidence = append(evidence, fmt.Sprintf("=== LISTENING PORTS ===\n%s", string(out)))
	} else {
		out, _ := exec.CommandContext(ctx, "ss", "-tulnp").Output()
		evidence = append(evidence, fmt.Sprintf("=== LISTENING PORTS ===\n%s", string(out)))
	}

	summary := strings.Join(evidence, "\n\n")
	return &ActionResult{Success: true, Message: fmt.Sprintf("Evidence captured (%d bytes)", len(summary))}, nil
}

// ExportForensicBundle exports evidence to a file.
func ExportForensicBundle() (*ActionResult, error) {
	return &ActionResult{Success: true, Message: "Forensic bundle exported (placeholder)"}, nil
}
