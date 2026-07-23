package secops

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
)

// AutoExpireWG tracks background goroutines for auto-expire isolation.
var AutoExpireWG sync.WaitGroup

// cmdTimeout is the maximum time to wait for a single firewall command.
const cmdTimeout = 10 * time.Second

// KillProcess force-kills a process by PID.
func KillProcess(pid int) (*common.SecActionResult, error) {
	if pid <= 0 || pid > 4194304 {
		return &common.SecActionResult{Success: false, Error: fmt.Sprintf("invalid PID: %d", pid)}, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
	defer cancel()
	cfg := common.RemediationSandbox()

	if runtime.GOOS == "windows" {
		cmd := common.SandboxedCommandWithConfigContext(ctx, cfg, "taskkill", "/F", "/PID", fmt.Sprintf("%d", pid))
		out, err := cmd.Output()
		if err != nil {
			return &common.SecActionResult{Success: false, Error: err.Error()}, nil
		}
		return &common.SecActionResult{Success: true, Message: strings.TrimSpace(string(out))}, nil
	}
	cmd := common.SandboxedCommandWithConfigContext(ctx, cfg, "kill", "-9", fmt.Sprintf("%d", pid))
	out, err := cmd.Output()
	if err != nil {
		return &common.SecActionResult{Success: false, Error: err.Error()}, nil
	}
	return &common.SecActionResult{Success: true, Message: strings.TrimSpace(string(out))}, nil
}

// BlockIP blocks an IP address via firewall.
func BlockIP(ip string) (*common.SecActionResult, error) {
	if !common.ValidIP(ip) {
		return &common.SecActionResult{Success: false, Error: fmt.Sprintf("invalid IP address: %q", ip)}, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
	defer cancel()
	cfg := common.RemediationSandbox()

	if runtime.GOOS == "windows" {
		cmd := common.SandboxedCommandWithConfigContext(ctx, cfg, "netsh", "advfirewall", "firewall", "add", "rule",
			"name=SECOPS_Block_"+ip, "dir=in", "action=block", "remoteip="+ip)
		if err := cmd.Run(); err != nil {
			return &common.SecActionResult{Success: false, Error: err.Error()}, nil
		}
		return &common.SecActionResult{Success: true, Message: fmt.Sprintf("Blocked IP %s", ip)}, nil
	}
	cmd := common.SandboxedCommandWithConfigContext(ctx, cfg, "iptables", "-A", "INPUT", "-s", ip, "-j", "DROP")
	if err := cmd.Run(); err != nil {
		return &common.SecActionResult{Success: false, Error: err.Error()}, nil
	}
	return &common.SecActionResult{Success: true, Message: fmt.Sprintf("Blocked IP %s", ip)}, nil
}

// DisableAccount disables a local user account.
func DisableAccount(username string) (*common.SecActionResult, error) {
	if !common.ValidUsername(username) {
		return &common.SecActionResult{Success: false, Error: fmt.Sprintf("invalid username: %q", username)}, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
	defer cancel()
	cfg := common.RemediationSandbox()

	if runtime.GOOS == "windows" {
		cmd := common.SandboxedCommandWithConfigContext(ctx, cfg, "net", "user", username, "/active:no")
		if err := cmd.Run(); err != nil {
			return &common.SecActionResult{Success: false, Error: err.Error()}, nil
		}
		return &common.SecActionResult{Success: true, Message: fmt.Sprintf("Account %s disabled", username)}, nil
	}
	cmd := common.SandboxedCommandWithConfigContext(ctx, cfg, "passwd", "-l", username)
	if err := cmd.Run(); err != nil {
		return &common.SecActionResult{Success: false, Error: err.Error()}, nil
	}
	return &common.SecActionResult{Success: true, Message: fmt.Sprintf("Account %s locked", username)}, nil
}

// CaptureEvidence collects a structured forensic snapshot of the system.
func CaptureEvidence() (*common.SecActionResult, error) {
	s := common.GetStorage()
	if s == nil {
		return &common.SecActionResult{Success: false, Error: "storage not initialized"}, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// 1. Collect Data
	// Process List
	procs, _ := common.HiddenCommandContext(ctx, "powershell", "-Command", "Get-Process | Select-Object Id,ProcessName,CPU,WorkingSet,Path | ConvertTo-Json").Output()
	if runtime.GOOS != "windows" {
		procs, _ = exec.CommandContext(ctx, "ps", "aux", "--json").Output()
	}

	// Connections
	conns, _ := common.HiddenCommandContext(ctx, "powershell", "-Command", "Get-NetTCPConnection | Select-Object LocalAddress,LocalPort,RemoteAddress,RemotePort,State,OwningProcess | ConvertTo-Json").Output()

	// Env
	env := strings.Join(os.Environ(), "\n")

	data := map[string]interface{}{
		"processes":   string(procs),
		"connections": string(conns),
		"environment": env,
		"platform":    runtime.GOOS,
		"arch":        runtime.GOARCH,
	}

	dataJSON, _ := json.Marshal(data)

	id := fmt.Sprintf("forensic-%d", time.Now().Unix())
	record := common.ForensicRecord{
		ID:        id,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Type:      "full_snapshot",
		Title:     "System Forensic Snapshot",
		DataJSON:  string(dataJSON),
		Metadata:  fmt.Sprintf("Captured via automated response. Platform: %s", runtime.GOOS),
	}

	if err := s.InsertForensic(record); err != nil {
		return &common.SecActionResult{Success: false, Error: err.Error()}, nil
	}

	return &common.SecActionResult{Success: true, Message: fmt.Sprintf("Structured evidence captured: %s", id)}, nil
}

// ExportForensicBundle exports evidence to a file.
func ExportForensicBundle(id string) (*common.SecActionResult, error) {
	s := common.GetStorage()
	if s == nil {
		return &common.SecActionResult{Success: false, Error: "storage not initialized"}, nil
	}

	record, err := s.GetForensic(id)
	if err != nil || record == nil {
		return &common.SecActionResult{Success: false, Error: "snapshot not found"}, nil
	}

	// Export the forensic snapshot to a JSON file in data/forensics/
	dataDir, _ := common.ConfigDir()
	exportDir := fmt.Sprintf("%s/forensics", dataDir)
	_ = common.HiddenCommand("cmd", "/c", "mkdir", exportDir).Run()

	filename := fmt.Sprintf("%s/%s.json", exportDir, id)
	err = os.WriteFile(filename, []byte(record.DataJSON), 0644)
	if err != nil {
		return &common.SecActionResult{Success: false, Error: err.Error()}, nil
	}

	return &common.SecActionResult{Success: true, Message: fmt.Sprintf("Forensic snapshot exported to %s", filename)}, nil
}

// RemoveIsolation removes all SECOPS isolation rules (both inbound and outbound).
func RemoveIsolation() (*common.SecActionResult, error) {
	return IsolateHost(false, 0)
}
