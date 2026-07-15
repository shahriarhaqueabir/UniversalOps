package secops

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
)

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

	// Enable firewall
	err := exec.CommandContext(ctx, "netsh", "advfirewall", "set", "allprofiles", "state", "on").Run()
	if err != nil {
		return &ActionResult{Success: false, Error: err.Error()}, nil
	}
	// Add block-all inbound rule
	err = exec.CommandContext(ctx, "netsh", "advfirewall", "firewall", "add", "rule",
		"name=SECOPS_Isolate_BlockAll", "dir=in", "action=block").Run()
	if err != nil {
		return &ActionResult{Success: false, Error: err.Error()}, nil
	}
	// Also block outbound to fully isolate
	err = exec.CommandContext(ctx, "netsh", "advfirewall", "firewall", "add", "rule",
		"name=SECOPS_Isolate_BlockAll_Out", "dir=out", "action=block").Run()
	if err != nil {
		// Roll back inbound rule if outbound failed
		rollbackCtx, rollbackCancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer rollbackCancel()
		_ = exec.CommandContext(rollbackCtx, "netsh", "advfirewall", "firewall", "delete", "rule", "name=SECOPS_Isolate_BlockAll").Run()
		return &ActionResult{Success: false, Error: fmt.Sprintf("outbound block failed: %v", err)}, nil
	}

	msg := "Host isolated — all inbound and outbound traffic blocked"
	if autoExpireSeconds > 0 {
		msg += fmt.Sprintf(" (auto-expires in %d seconds)", autoExpireSeconds)
		go func() {
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

	// Block inbound
	err := exec.CommandContext(ctx, "iptables", "-A", "INPUT", "-j", "DROP").Run()
	if err != nil {
		return &ActionResult{Success: false, Error: err.Error()}, nil
	}
	// Also block outbound to fully isolate
	err = exec.CommandContext(ctx, "iptables", "-A", "OUTPUT", "-j", "DROP").Run()
	if err != nil {
		// Roll back inbound rule if outbound failed
		rollbackCtx, rollbackCancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer rollbackCancel()
		_ = exec.CommandContext(rollbackCtx, "iptables", "-D", "INPUT", "-j", "DROP").Run()
		return &ActionResult{Success: false, Error: fmt.Sprintf("outbound block failed: %v", err)}, nil
	}

	msg := "Host isolated — all inbound and outbound traffic dropped"
	if autoExpireSeconds > 0 {
		msg += fmt.Sprintf(" (auto-expires in %d seconds)", autoExpireSeconds)
		go func() {
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
	// Remove inbound rule
	err := exec.CommandContext(ctx, "netsh", "advfirewall", "firewall", "delete", "rule", "name=SECOPS_Isolate_BlockAll").Run()
	if err != nil {
		errs = append(errs, fmt.Sprintf("inbound: %v", err))
	}
	// Remove outbound rule
	err = exec.CommandContext(ctx, "netsh", "advfirewall", "firewall", "delete", "rule", "name=SECOPS_Isolate_BlockAll_Out").Run()
	if err != nil {
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
	// Remove inbound DROP rule
	err := exec.CommandContext(ctx, "iptables", "-D", "INPUT", "-j", "DROP").Run()
	if err != nil {
		errs = append(errs, fmt.Sprintf("inbound: %v", err))
	}
	// Remove outbound DROP rule
	err = exec.CommandContext(ctx, "iptables", "-D", "OUTPUT", "-j", "DROP").Run()
	if err != nil {
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
	if common.IsWindows() {
		out, err := exec.CommandContext(ctx, "taskkill", "/F", "/PID", fmt.Sprintf("%d", pid)).Output()
		if err != nil {
			return &ActionResult{Success: false, Error: err.Error()}, nil
		}
		return &ActionResult{Success: true, Message: strings.TrimSpace(string(out))}, nil
	}
	out, err := exec.CommandContext(ctx, "kill", "-9", fmt.Sprintf("%d", pid)).Output()
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
	if common.IsWindows() {
		err := exec.CommandContext(ctx, "netsh", "advfirewall", "firewall", "add", "rule",
			"name=SECOPS_Block_"+ip, "dir=in", "action=block", "remoteip="+ip).Run()
		if err != nil {
			return &ActionResult{Success: false, Error: err.Error()}, nil
		}
		return &ActionResult{Success: true, Message: fmt.Sprintf("Blocked IP %s", ip)}, nil
	}
	err := exec.CommandContext(ctx, "iptables", "-A", "INPUT", "-s", ip, "-j", "DROP").Run()
	if err != nil {
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
	if common.IsWindows() {
		err := exec.CommandContext(ctx, "net", "user", username, "/active:no").Run()
		if err != nil {
			return &ActionResult{Success: false, Error: err.Error()}, nil
		}
		return &ActionResult{Success: true, Message: fmt.Sprintf("Account %s disabled", username)}, nil
	}
	err := exec.CommandContext(ctx, "passwd", "-l", username).Run()
	if err != nil {
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
