package common

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// CapabilityID represents a unique identifier for a system capability (tool/binary).
type CapabilityID string

const (
	CapNmap          CapabilityID = "nmap"
	CapDocker        CapabilityID = "docker"
	CapDockerCompose CapabilityID = "docker-compose"
	CapPodman        CapabilityID = "podman"
	CapKubernetes    CapabilityID = "k8s"
	CapKubectl       CapabilityID = "kubectl"
	CapHelm          CapabilityID = "helm"
	CapTerraform     CapabilityID = "terraform"
	CapOpenTofu      CapabilityID = "tofu"
	CapOllama        CapabilityID = "ollama"
	CapGit           CapabilityID = "git"
	CapPowerShell    CapabilityID = "powershell"
	CapPwsh          CapabilityID = "pwsh"
	CapPython        CapabilityID = "python"
	CapGo            CapabilityID = "go"
	CapNode          CapabilityID = "node"
	CapWireshark     CapabilityID = "wireshark"
	CapSSH           CapabilityID = "ssh"
	CapNvidiaSmi     CapabilityID = "nvidia-smi"
	CapLHM           CapabilityID = "lhm"
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

var (
	capabilityRegistry   *CapabilityRegistry
	capabilityRegistryMu sync.Once
)

// GetCapabilityRegistry returns the global singleton registry.
func GetCapabilityRegistry() *CapabilityRegistry {
	capabilityRegistryMu.Do(func() {
		capabilityRegistry = NewCapabilityRegistry()
	})
	return capabilityRegistry
}

/*
CapabilityRegistry — Probes the local workstation for installed tools and binaries.
  - Implements the "Capability Gateway" logic by determining what features can be unlocked.
  - Supports dynamic user overrides to avoid hardcoded path assumptions.
*/
type CapabilityRegistry struct {
	mu         sync.RWMutex
	tools      map[CapabilityID]CapabilityInfo
	overrides  map[CapabilityID]string
	dynamic    []CapabilityID
	refreshing atomic.Bool
	stopCh     chan struct{}
}

// NewCapabilityRegistry initializes a registry and performs an initial probe.
func NewCapabilityRegistry() *CapabilityRegistry {
	r := &CapabilityRegistry{
		tools:     make(map[CapabilityID]CapabilityInfo),
		overrides: make(map[CapabilityID]string),
		dynamic:   []CapabilityID{},
		stopCh:    make(chan struct{}),
	}
	r.Refresh()
	return r
}

// SetOverride allows the user to manually specify a path for a capability.
func (r *CapabilityRegistry) SetOverride(id CapabilityID, path string) {
	// Validate: path must exist and be a regular file (allow empty to clear)
	if path != "" {
		if info, err := os.Stat(path); err != nil {
			LogWarn("SetOverride(%s, %s): path does not exist or cannot be accessed", id, path)
			return
		} else if info.IsDir() {
			LogWarn("SetOverride(%s, %s): path is a directory, not a binary", id, path)
			return
		}
	}
	r.mu.Lock()
	r.overrides[id] = path
	r.mu.Unlock()
	r.Refresh()
}

// AddCapability manually registers a new ID to be scanned.
func (r *CapabilityRegistry) AddCapability(id CapabilityID) {
	r.mu.Lock()
	r.dynamic = append(r.dynamic, id)
	r.mu.Unlock()
	r.Refresh()
}

// Refresh re-scans the system PATH for the registered capability IDs, accounting for overrides.
func (r *CapabilityRegistry) Refresh() {
	// 1. Priority Sync Scan (limited subset, one-at-a-time with per-item locks)
	//    Each RefreshBatch call acquires/releases the lock independently,
	//    preventing long write-lock holds during subprocess execution.
	priorityIds := []CapabilityID{
		CapOllama, CapPowerShell, CapPwsh, // Bare essentials for initialization
	}
	for _, id := range priorityIds {
		r.RefreshBatch([]CapabilityID{id})
	}

	// 2. Background Refresh for everything else
	go func() {
		defer RecoverPanic()

		// Use atomic flag to prevent multiple concurrent background refreshes
		if !r.refreshing.CompareAndSwap(false, true) {
			return
		}
		defer r.refreshing.Store(false)

		allIds := []CapabilityID{
			CapNmap, CapDocker, CapDockerCompose, CapPodman,
			CapKubernetes, CapKubectl, CapHelm, CapTerraform, CapOpenTofu,
			CapPython, CapNode, CapWireshark, CapSSH, CapNvidiaSmi, CapLHM, CapGit,
		}

		// Append dynamically registered capabilities
		r.mu.RLock()
		allIds = append(allIds, r.dynamic...)
		r.mu.RUnlock()

		// Perform background scan one-by-one with local locks to avoid
		// holding the global lock during external binary executions.
		for _, id := range allIds {
			select {
			case <-r.stopCh:
				return
			default:
			}
			r.RefreshBatch([]CapabilityID{id})
			// Small sleep between tools to reduce burst I/O impact
			time.Sleep(50 * time.Millisecond)
		}
	}()
}

// Shutdown cancels any ongoing background refresh. Safe to call multiple times.
func (r *CapabilityRegistry) Shutdown() {
	select {
	case <-r.stopCh:
		// Already closed
	default:
		close(r.stopCh)
	}
}

// RefreshBatch scans a specific subset of tools.
func (r *CapabilityRegistry) RefreshBatch(ids []CapabilityID) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.refreshBatch(ids)
}

func (r *CapabilityRegistry) refreshBatch(ids []CapabilityID) {
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

		// 2. Standard PATH lookup
		if path == "" {
			binaryName := string(id)
			if id == CapKubernetes {
				binaryName = "kubectl"
			}
			path, err = exec.LookPath(binaryName)
		}

		// 3. Check local 'bin' directory (Fallback for bundled tools)
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

		// 4. Smart Windows Discovery
		if err != nil && runtime.GOOS == "windows" {
			path, err = r.discoverWindows(id)
		}

		// 5. Specialized WMI check for LHM (with 5s timeout)
		if id == CapLHM && runtime.GOOS == "windows" {
			// LHM is detected via WMI namespace existence
			// We use a simple powershell check to avoid adding a heavy dependency to common if not already present.
			// However, since we want a fast check, we check for a known WMI namespace.
			lhmCtx, lhmCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer lhmCancel()
			cmd := HiddenCommandContext(lhmCtx, "powershell", "-Command", "Get-CimInstance -Namespace root\\LibreHardwareMonitor -ClassName Sensor -ErrorAction SilentlyContinue | Select-Object -First 1")
			if errCheck := cmd.Run(); errCheck == nil {
				path = "WMI:root\\LibreHardwareMonitor"
				err = nil
			} else {
				err = fmt.Errorf("LibreHardwareMonitor WMI namespace not found")
			}
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
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var cmd *exec.Cmd
	var out []byte
	var err error

	switch id {
	case CapPowerShell:
		// powershell -Version expects a version argument; query PSVersionTable instead
		cmd = HiddenCommandContext(ctx, path, "-NoProfile", "-Command",
			"Write-Host (Get-Variable PSVersionTable -ValueOnly).PSVersion.ToString()")
		out, err = cmd.Output()

	case CapSSH:
		// ssh -V outputs to stderr; use CombinedOutput to capture it
		cmd = HiddenCommandContext(ctx, path, "-V")
		out, err = cmd.CombinedOutput()

	case CapKubernetes, CapKubectl:
		// kubectl version contacts the API server; --client avoids needing a cluster
		cmd = HiddenCommandContext(ctx, path, "version", "--client")
		out, err = cmd.Output()

	case CapLHM:
		return "not applicable"

	default:
		arg := "--version"
		switch id {
		case CapNmap:
			arg = "-version"
		case CapGo:
			arg = "version"
		}
		cmd = HiddenCommandContext(ctx, path, arg)
		out, err = cmd.Output()
	}

	if err != nil {
		LogWarn("version detection failed for %s: %v", id, err)
		return "unknown"
	}

	// First-line extraction with fallback for blank/garbled output
	lines := strings.Split(string(out), "\n")
	for _, l := range lines {
		if trimmed := strings.TrimSpace(l); trimmed != "" {
			return trimmed
		}
	}
	return "unknown"
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
	case CapNvidiaSmi:
		path := "C:\\Windows\\System32\\nvidia-smi.exe"
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

// GetPath returns the verified system path for a capability if available.
func (r *CapabilityRegistry) GetPath(id CapabilityID) (string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	info, ok := r.tools[id]
	if !ok || !info.Available {
		return "", false
	}
	return info.Path, true
}
