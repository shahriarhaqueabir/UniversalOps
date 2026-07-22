package common

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

type PendingAction struct {
	Action  string
	Command string
	Params  map[string]interface{}
	Expiry  time.Time
}

type HandshakeRegistry struct {
	mu      sync.RWMutex
	pending map[string]PendingAction
}

var globalHandshakes = func() *HandshakeRegistry {
	r := &HandshakeRegistry{
		pending: make(map[string]PendingAction),
	}
	go r.cleanupLoop()
	return r
}()

func GetHandshakeRegistry() *HandshakeRegistry {
	return globalHandshakes
}

func (r *HandshakeRegistry) cleanupLoop() {
	defer RecoverPanic()
	ticker := time.NewTicker(1 * time.Minute)
	for {
		<-ticker.C
		r.Cleanup()
	}
}

func (r *HandshakeRegistry) Cleanup() {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	for id, p := range r.pending {
		if now.After(p.Expiry) {
			delete(r.pending, id)
		}
	}
}

func (r *HandshakeRegistry) Register(action string, command string, params map[string]interface{}) string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	id := hex.EncodeToString(b)

	r.mu.Lock()
	r.pending[id] = PendingAction{
		Action:  action,
		Command: command,
		Params:  params,
		Expiry:  time.Now().Add(5 * time.Minute),
	}
	r.mu.Unlock()
	return id
}

func (r *HandshakeRegistry) Consume(id string) (PendingAction, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	p, ok := r.pending[id]
	if !ok {
		return PendingAction{}, fmt.Errorf("invalid or expired handshake ID")
	}
	if time.Now().After(p.Expiry) {
		delete(r.pending, id)
		return PendingAction{}, fmt.Errorf("handshake expired")
	}
	delete(r.pending, id)
	return p, nil
}

// CreatePreview generates an ActionPreview for an action and registers it.
func (r *HandshakeRegistry) CreatePreview(action string, params map[string]interface{}) ActionPreview {
	command := ""
	switch action {
	case "kill_process":
		command = fmt.Sprintf("taskkill /F /PID %v", params["pid"])
	case "block_ip":
		command = fmt.Sprintf("netsh advfirewall firewall add rule name=\"Block %v\" dir=in action=block remoteip=%v", params["ip"], params["ip"])
	case "isolate_host":
		command = "netsh advfirewall set allprofiles firewallpolicy blockinbound,blockoutbound"
	case "restart_service":
		// Use PowerShell for consistent service management on Windows
		command = fmt.Sprintf("powershell -Command \"Restart-Service -Name '%v' -Force\"", params["name"])
	case "kill_process_tree":
		command = fmt.Sprintf("Get-Process -Id %v | Stop-Process -Force -ErrorAction SilentlyContinue", params["pid"])
	}

	id := r.Register(action, command, params)

	preview := ActionPreview{
		HandshakeID: id,
		Action:      action,
		Command:     command,
	}

	switch action {
	case "kill_process":
		pid := params["pid"]
		preview.Description = fmt.Sprintf("Terminate process with PID %v", pid)
		preview.Risks = []string{"Data loss if process has unsaved state", "System instability if it is a critical service"}
		preview.Rollback = "Restarting the process manually (if persistent service)"
	case "block_ip":
		ip := params["ip"]
		preview.Description = fmt.Sprintf("Add firewall rule to block traffic from %v", ip)
		preview.Risks = []string{"Legitimate users sharing the IP may be blocked", "Misconfiguration could block internal traffic"}
		preview.Rollback = "Removing the rule via NetOps tab"
	case "isolate_host":
		preview.Description = "Isolate host from all network communication"
		preview.Risks = []string{"Remote access will be lost immediately", "Application services will stop functioning"}
		preview.Rollback = "Manual console access to flush iptables/nftables"
	case "restart_service":
		name := params["name"]
		preview.Description = fmt.Sprintf("Restart service %q", name)
		preview.Risks = []string{"Temporary service downtime during restart"}
		preview.Rollback = "Systemd will attempt to bring it back up automatically"
	case "workflow":
		if wid, ok := params["workflow_id"].(string); ok {
			preview.WorkflowID = wid
			preview.Description = fmt.Sprintf("Execute operational workflow: %s", wid)
			preview.Risks = []string{"Step-specific risks apply (see workflow definition)", "Potential for multi-resource impact during sequential execution"}
		}
	case "kill_process_tree":
		pid := params["pid"]
		preview.Description = fmt.Sprintf("Terminate process tree starting at PID %v", pid)
		preview.Risks = []string{"Unsaved data in the application and all its sub-processes may be lost", "System instability if critical services are terminated"}
		rollback := "Applications must be restarted manually"
		preview.Rollback = rollback
	default:
		preview.Description = "Perform system action"
		preview.Risks = []string{"Unknown risks for custom action"}
		preview.Rollback = "Check system logs for details"
	}

	return preview
}
