package common

import (
	"os/exec"
	"sync"
)

// CapabilityID represents a unique identifier for a system capability (tool/binary).
type CapabilityID string

const (
	CapNmap       CapabilityID = "nmap"
	CapDocker     CapabilityID = "docker"
	CapOllama     CapabilityID = "ollama"
	CapGit        CapabilityID = "git"
	CapPowerShell CapabilityID = "pwsh"
)

// CapabilityInfo provides status and path information for a detected capability.
type CapabilityInfo struct {
	ID        CapabilityID `json:"id"`
	Available bool         `json:"available"`
	Path      string       `json:"path"`
}

/**
 * CapabilityRegistry — Probes the local workstation for installed tools and binaries.
 * Implements the "Capability Gateway" logic by determining what features can be unlocked.
 */
type CapabilityRegistry struct {
	mu    sync.RWMutex
	tools map[CapabilityID]CapabilityInfo
}

// NewCapabilityRegistry initializes a registry and performs an initial probe.
func NewCapabilityRegistry() *CapabilityRegistry {
	r := &CapabilityRegistry{tools: make(map[CapabilityID]CapabilityInfo)}
	r.Refresh()
	return r
}

// Refresh re-scans the system PATH for the registered capability IDs.
func (r *CapabilityRegistry) Refresh() {
	r.mu.Lock()
	defer r.mu.Unlock()

	ids := []CapabilityID{CapNmap, CapDocker, CapOllama, CapGit, CapPowerShell}
	for _, id := range ids {
		path, err := exec.LookPath(string(id))
		r.tools[id] = CapabilityInfo{
			ID:        id,
			Available: err == nil,
			Path:      path,
		}
	}
}

// List returns a slice of all detected capabilities.
func (r *CapabilityRegistry) List() []CapabilityInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]CapabilityInfo, 0, len(r.tools))
	for _, info := range r.tools {
		out = append(out, info)
	}
	return out
}

// IsAvailable returns true if the specific capability ID was found on the system.
func (r *CapabilityRegistry) IsAvailable(id CapabilityID) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	info, ok := r.tools[id]
	return ok && info.Available
}
