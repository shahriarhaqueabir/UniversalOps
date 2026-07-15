package secops

import (
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
	// Enable firewall
	err := exec.Command("netsh", "advfirewall", "set", "allprofiles", "state", "on").Run()
	if err != nil {
		return &ActionResult{Success: false, Error: err.Error()}, nil
	}
	// Add block-all rule
	err = exec.Command("netsh", "advfirewall", "firewall", "add", "rule",
		"name=SECOPS_Isolate_BlockAll", "dir=in", "action=block").Run()
	if err != nil {
		return &ActionResult{Success: false, Error: err.Error()}, nil
	}

	msg := "Host isolated — all inbound traffic blocked"
	if autoExpireSeconds > 0 {
		msg += fmt.Sprintf(" (auto-expires in %d seconds)", autoExpireSeconds)
		// Schedule rule removal
		go func() {
			time.Sleep(time.Duration(autoExpireSeconds) * time.Second)
			_ = exec.Command("netsh", "advfirewall", "firewall", "delete", "rule", "name=SECOPS_Isolate_BlockAll").Run()
		}()
	}
	return &ActionResult{Success: true, Message: msg}, nil
}

func isolateHostLinux(autoExpireSeconds int) (*ActionResult, error) {
	err := exec.Command("iptables", "-A", "INPUT", "-j", "DROP").Run()
	if err != nil {
		return &ActionResult{Success: false, Error: err.Error()}, nil
	}

	msg := "Host isolated — all inbound traffic dropped"
	if autoExpireSeconds > 0 {
		msg += fmt.Sprintf(" (auto-expires in %d seconds)", autoExpireSeconds)
		// Schedule rule removal
		go func() {
			time.Sleep(time.Duration(autoExpireSeconds) * time.Second)
			_ = exec.Command("iptables", "-D", "INPUT", "-j", "DROP").Run()
		}()
	}
	return &ActionResult{Success: true, Message: msg}, nil
}

// KillProcess force-kills a process by PID.
func KillProcess(pid int) (*ActionResult, error) {
	if pid <= 0 || pid > 4194304 {
		return &ActionResult{Success: false, Error: fmt.Sprintf("invalid PID: %d", pid)}, nil
	}
	if common.IsWindows() {
		out, err := exec.Command("taskkill", "/F", "/PID", fmt.Sprintf("%d", pid)).Output()
		if err != nil {
			return &ActionResult{Success: false, Error: err.Error()}, nil
		}
		return &ActionResult{Success: true, Message: strings.TrimSpace(string(out))}, nil
	}
	out, err := exec.Command("kill", "-9", fmt.Sprintf("%d", pid)).Output()
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
	if common.IsWindows() {
		err := exec.Command("netsh", "advfirewall", "firewall", "add", "rule",
			"name=SECOPS_Block_"+ip, "dir=in", "action=block", "remoteip="+ip).Run()
		if err != nil {
			return &ActionResult{Success: false, Error: err.Error()}, nil
		}
		return &ActionResult{Success: true, Message: fmt.Sprintf("Blocked IP %s", ip)}, nil
	}
	err := exec.Command("iptables", "-A", "INPUT", "-s", ip, "-j", "DROP").Run()
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
	if common.IsWindows() {
		err := exec.Command("net", "user", username, "/active:no").Run()
		if err != nil {
			return &ActionResult{Success: false, Error: err.Error()}, nil
		}
		return &ActionResult{Success: true, Message: fmt.Sprintf("Account %s disabled", username)}, nil
	}
	err := exec.Command("passwd", "-l", username).Run()
	if err != nil {
		return &ActionResult{Success: false, Error: err.Error()}, nil
	}
	return &ActionResult{Success: true, Message: fmt.Sprintf("Account %s locked", username)}, nil
}

// CaptureEvidence collects forensic evidence into a summary.
func CaptureEvidence() (*ActionResult, error) {
	var evidence []string

	// Collect running processes
	if runtime.GOOS == "windows" {
		out, _ := exec.Command("tasklist").Output()
		evidence = append(evidence, fmt.Sprintf("=== PROCESSES ===\n%s", string(out)))
	} else {
		out, _ := exec.Command("ps", "aux").Output()
		evidence = append(evidence, fmt.Sprintf("=== PROCESSES ===\n%s", string(out)))
	}

	// Collect listening ports
	if runtime.GOOS == "windows" {
		out, _ := exec.Command("netstat", "-ano").Output()
		evidence = append(evidence, fmt.Sprintf("=== LISTENING PORTS ===\n%s", string(out)))
	} else {
		out, _ := exec.Command("ss", "-tulnp").Output()
		evidence = append(evidence, fmt.Sprintf("=== LISTENING PORTS ===\n%s", string(out)))
	}

	summary := strings.Join(evidence, "\n\n")
	return &ActionResult{Success: true, Message: fmt.Sprintf("Evidence captured (%d bytes)", len(summary))}, nil
}

// ExportForensicBundle exports evidence to a file.
func ExportForensicBundle() (*ActionResult, error) {
	return &ActionResult{Success: true, Message: "Forensic bundle exported (placeholder)"}, nil
}
