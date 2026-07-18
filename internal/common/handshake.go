package common

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

type PendingAction struct {
	Action string
	Params map[string]interface{}
	Expiry time.Time
}

type HandshakeRegistry struct {
	mu      sync.RWMutex
	pending map[string]PendingAction
}

var globalHandshakes = &HandshakeRegistry{
	pending: make(map[string]PendingAction),
}

func GetHandshakeRegistry() *HandshakeRegistry {
	return globalHandshakes
}

func (r *HandshakeRegistry) Register(action string, params map[string]interface{}) string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	id := hex.EncodeToString(b)

	r.mu.Lock()
	r.pending[id] = PendingAction{
		Action: action,
		Params: params,
		Expiry: time.Now().Add(5 * time.Minute),
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
	id := r.Register(action, params)

	preview := ActionPreview{
		HandshakeID: id,
		Action:      action,
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
	default:
		preview.Description = "Perform system action"
		preview.Risks = []string{"Unknown risks for custom action"}
		preview.Rollback = "Check system logs for details"
	}

	return preview
}
