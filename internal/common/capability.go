package common

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

// CapabilityID represents a unique identifier for a system capability (tool/binary).
type CapabilityID string

const (
	CapNmap         CapabilityID = "nmap"
	CapDocker       CapabilityID = "docker"
	CapDockerCompose CapabilityID = "docker-compose"
	CapPodman       CapabilityID = "podman"
	CapKubernetes   CapabilityID = "k8s"
	CapKubectl      CapabilityID = "kubectl"
	CapHelm         CapabilityID = "helm"
	CapTerraform    CapabilityID = "terraform"
	CapOpenTofu     CapabilityID = "tofu"
	CapOllama       CapabilityID = "ollama"
	CapGit          CapabilityID = "git"
	CapPowerShell   CapabilityID = "pwsh"
	CapPython        CapabilityID = "python"
	CapGo           CapabilityID = "go"
	CapNode         CapabilityID = "node"
	CapWireshark     CapabilityID = "wireshark"
	CapSSH          CapabilityID = "ssh"
)

// CapabilityInfo provides status and path information for a detected capability.
type CapabilityInfo struct {
	ID          CapabilityID `json:"id"`
	Available   bool         `json:"available"`
	Path        string       `json:"path"`
	Version     string       `json:"version"`
	IsCustom    bool         `json:"is_custom"`
	IsSupported bool         `json:"is_supported"`
}


/**
 * CapabilityRegistry — Probes the local workstation for installed tools and binaries.
 * Implements the "Capability Gateway" logic by determining what features can be unlocked.
 * Supports dynamic user overrides to avoid hardcoded path assumptions.
 */
type CapabilityRegistry struct {
	mu        sync.RWMutex
	tools     map[CapabilityID]CapabilityInfo
	overrides map[CapabilityID]string
}

// NewCapabilityRegistry initializes a registry and performs an initial probe.
func NewCapabilityRegistry() *CapabilityRegistry {
	r := &CapabilityRegistry{
		tools:     make(map[CapabilityID]CapabilityInfo),
		overrides: make(map[CapabilityID]string),
	}
	r.Refresh()
	return r
}

// SetOverride allows the user to manually specify a path for a capability.
func (r *CapabilityRegistry) SetOverride(id CapabilityID, path string) {
	r.mu.Lock()
	r.overrides[id] = path
	r.mu.Unlock()
	r.Refresh()
}

// Refresh re-scans the system PATH for the registered capability IDs, accounting for overrides.
func (r *CapabilityRegistry) Refresh() {
	r.mu.Lock()
	defer r.mu.Unlock()

	ids := []CapabilityID{
		CapNmap, CapDocker, CapDockerCompose, CapPodman,
		CapKubernetes, CapKubectl, CapHelm, CapTerraform, CapOpenTofu,
		CapOllama, CapGit, CapPowerShell, CapPython, CapGo, CapNode,
		CapWireshark, CapSSH,
	}

	for _, id := range ids {
		var path string
		var err error
		isCustom := false

		// 1. Check User Override (Highest Priority)
		if override, ok := r.overrides[id]; ok && override != "" {
			path = override
			_, err = os.Stat(path)
			isCustom = true
		}

		// 2. Check local 'bin' directory (Self-contained preference)
		if path == "" {
			localPath := filepath.Join("bin", string(id))
			if runtime.GOOS == "windows" {
				localPath += ".exe"
			}
			if _, errStat := os.Stat(localPath); errStat == nil {
				path = localPath
				err = nil
			}
		}

		// 3. Standard PATH lookup
		if path == "" {
			binaryName := string(id)
			if id == CapPowerShell && runtime.GOOS == "windows" {
				binaryName = "powershell"
			} else if id == CapKubernetes {
				binaryName = "kubectl"
			}
			path, err = exec.LookPath(binaryName)
		}

		// 4. Smart Windows Discovery
		if err != nil && runtime.GOOS == "windows" {
			path, err = r.discoverWindows(id)
		}

		version := ""
		if err == nil {
			version = r.detectVersion(id, path)
		}

		r.tools[id] = CapabilityInfo{
			ID:          id,
			Available:   err == nil,
			Path:        path,
			Version:     version,
			IsCustom:    isCustom,
			IsSupported: true, // Mark all as supported if detected for now
		}
	}
}

func (r *CapabilityRegistry) detectVersion(id CapabilityID, path string) string {
	arg := "--version"
	switch id {
	case CapPowerShell, CapNmap:
		arg = "-version"
	case CapGo:
		arg = "version"
	}

	cmd := exec.Command(path, arg)
	out, err := cmd.Output()
	if err != nil {
		return "unknown"
	}

	// Simple first-line extraction
	line := strings.Split(string(out), "\n")[0]
	return strings.TrimSpace(line)
}


// discoverWindows performs heuristic probing for tools likely to be in non-standard paths.
func (r *CapabilityRegistry) discoverWindows(id CapabilityID) (string, error) {
	switch id {
	case CapOllama:
		local := os.Getenv("LOCALAPPDATA")
		path := filepath.Join(local, "Programs", "Ollama", "ollama.exe")
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	case CapDocker:
		path := "C:\\Program Files\\Docker\\Docker\\resources\\bin\\docker.exe"
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}
	return "", fmt.Errorf("auto-discovery failed for %s", id)
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
