package sysops

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/shahriarhaqueabir/UniversalOps/internal/common"
	"github.com/shirou/gopsutil/v4/process"
)

// processMetadata caches static process information.
type processMetadata struct {
	Name       string
	PPID       int32
	CreateTime int64
	LastSeen   time.Time
}

// ProcessInfo holds information about a single process.
type ProcessInfo struct {
	PID       int32   `json:"pid"`
	PPID      int32   `json:"ppid"`
	Name      string  `json:"name"`
	CPU       float64 `json:"cpu"`
	Memory    float32 `json:"memory"` // RSS in MB
	MemPct    float32 `json:"mem_pct"`
	Status    string  `json:"status"`
	NumFDs    int32   `json:"num_fds"`
	IsSigned  bool    `json:"is_signed"`
	Publisher string  `json:"publisher,omitempty"`
}

type procTrustInfo struct {
	IsSigned  bool
	Publisher string
}

var (
	procCache   = make(map[int32]processMetadata)
	procCacheMu sync.RWMutex

	trustCache   = make(map[string]procTrustInfo)
	trustCacheMu sync.RWMutex

	lastSnapshot   []ProcessInfo
	lastSnapshotMu sync.RWMutex
	lastSnapshotAt time.Time

	// New: Async Trust Worker
	trustQueue   = make(chan string, 500)
	trustWorkerOnce sync.Once
)

func startTrustWorker() {
	trustWorkerOnce.Do(func() {
		go func() {
			for path := range trustQueue {
				// Avoid redundant checks if already in cache
				trustCacheMu.RLock()
				_, found := trustCache[path]
				trustCacheMu.RUnlock()
				if found {
					continue
				}

				// Perform expensive PowerShell check
				isSigned, publisher := fetchTrustInfo(path)

				// Update cache
				trustCacheMu.Lock()
				trustCache[path] = procTrustInfo{IsSigned: isSigned, Publisher: publisher}
				trustCacheMu.Unlock()
			}
		}()
	})
}

// UpdateProcessSnapshot performs a full system process scan and caches the result.
// OPTIMIZATION: Only fetches expensive metrics (CPU/Mem/FDs) for top active processes.
func UpdateProcessSnapshot() error {
	procs, err := process.Processes()
	if err != nil {
		return err
	}

	now := time.Now()

	// 1. Prune old cache entries (every 5 minutes or so)
	if now.Sub(lastSnapshotAt) > 5*time.Minute {
		pruneCaches(now)
	}

	// Pre-pass: Collect basic metadata and filter
	type procCandidate struct {
		p     *process.Process
		name  string
		ppid  int32
		isNew bool
	}
	candidates := make([]procCandidate, 0, len(procs))
	for _, p := range procs {
		createTime, _ := p.CreateTime()

		procCacheMu.RLock()
		meta, found := procCache[p.Pid]
		procCacheMu.RUnlock()

		// If PID recycled (different creation time), treat as new
		if found && meta.CreateTime != createTime {
			found = false
		}

		isNew := false
		if !found {
			name, err := p.Name()
			if err != nil {
				continue
			}
			ppid, _ := p.Ppid()
			meta = processMetadata{
				Name:       name,
				PPID:       ppid,
				CreateTime: createTime,
				LastSeen:   now,
			}
			procCacheMu.Lock()
			procCache[p.Pid] = meta
			procCacheMu.Unlock()
			isNew = true
		} else {
			// Update last seen
			procCacheMu.Lock()
			meta.LastSeen = now
			procCache[p.Pid] = meta
			procCacheMu.Unlock()
		}
		candidates = append(candidates, procCandidate{p: p, name: meta.Name, ppid: meta.PPID, isNew: isNew})
	}

	// PERFORMANCE: We collect CPU for all initially to find the top N,
	// but we skip MemoryInfo and NumFDs for non-top processes.
	result := make([]ProcessInfo, 0, len(candidates))
	for _, c := range candidates {
		// New processes need a small delay before CPU percent is accurate,
		// but we take a fast sample for the sort pass.
		cpu := float64(0)
		if !c.isNew {
			cpu, _ = c.p.CPUPercent()
		}

		pi := ProcessInfo{
			PID:    c.p.Pid,
			PPID:   c.ppid,
			Name:   c.name,
			CPU:    cpu,
			Status: "running",
		}
		result = append(result, pi)
	}

	// Sort to find top contributors
	sort.Slice(result, func(i, j int) bool {
		return result[i].CPU > result[j].CPU
	})

	// SECOND PASS: Deep metrics for top 50 processes only
	limit := 50
	if len(result) < limit {
		limit = len(result)
	}

	for i := 0; i < limit; i++ {
		p, _ := process.NewProcess(result[i].PID)
		if p == nil {
			continue
		}

		memInfo, _ := p.MemoryInfo()
		if memInfo != nil {
			result[i].Memory = float32(memInfo.RSS) / 1024 / 1024
		}
		memPct, _ := p.MemoryPercent()
		result[i].MemPct = memPct

		status, _ := p.Status()
		if len(status) > 0 {
			result[i].Status = status[0]
		}

		fdCount, _ := p.NumFDs()
		result[i].NumFDs = fdCount

		// TRUST METADATA: Signature and Publisher (Windows only)
		if path, err := p.Exe(); err == nil && path != "" {
			result[i].IsSigned, result[i].Publisher = getTrustInfo(path)
		}
	}

	lastSnapshotMu.Lock()
	lastSnapshot = result
	lastSnapshotAt = now
	lastSnapshotMu.Unlock()

	return nil
}

func getTrustInfo(path string) (bool, string) {
	if path == "" {
		return false, ""
	}

	startTrustWorker()

	// 1. Check Cache
	trustCacheMu.RLock()
	info, found := trustCache[path]
	trustCacheMu.RUnlock()
	if found {
		return info.IsSigned, info.Publisher
	}

	// 2. Enqueue for background check if not in cache
	select {
	case trustQueue <- path:
		// Enqueued
	default:
		// Queue full, skip for now to avoid blocking
	}

	return false, "Verifying..."
}

func fetchTrustInfo(path string) (bool, string) {
	// 2. Fetch via PowerShell (Expensive)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	cmd := common.HiddenCommandContext(ctx, "powershell", "-Command",
		fmt.Sprintf("(Get-AuthenticodeSignature '%s').Status; (Get-AuthenticodeSignature '%s').SignerCertificate.Subject", path, path))

	// HIDE WINDOW on Windows to prevent terminal spam
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.HideWindow = true

	out, err := cmd.Output()
	if err != nil {
		return false, ""
	}

	lines := strings.Split(string(out), "\n")
	if len(lines) < 1 {
		return false, ""
	}

	status := strings.TrimSpace(lines[0])
	isSigned := status == "Valid"
	publisher := ""
	if len(lines) >= 2 {
		publisher = strings.TrimSpace(lines[1])
		// Extract CN if possible
		if idx := strings.Index(publisher, "CN="); idx >= 0 {
			publisher = strings.Split(publisher[idx+3:], ",")[0]
		}
	}

	return isSigned, publisher
}

// GetTopProcesses returns the top N processes from the latest snapshot.
func GetTopProcesses(n int) ([]ProcessInfo, error) {
	lastSnapshotMu.RLock()
	snap := lastSnapshot
	lastSnapshotMu.RUnlock()

	if snap == nil {
		if err := UpdateProcessSnapshot(); err != nil {
			return nil, err
		}
		lastSnapshotMu.RLock()
		snap = lastSnapshot
		lastSnapshotMu.RUnlock()
	}

	result := snap
	if n > 0 && len(result) > n {
		result = result[:n]
	}

	out := make([]ProcessInfo, len(result))
	copy(out, result)
	return out, nil
}

// GetProcessCount returns the number of processes in the latest snapshot.
func GetProcessCount() int {
	lastSnapshotMu.RLock()
	defer lastSnapshotMu.RUnlock()
	return len(lastSnapshot)
}

// GetTotalOpenFDs returns the aggregate number of open FDs in the latest snapshot.
func GetTotalOpenFDs() int32 {
	lastSnapshotMu.RLock()
	defer lastSnapshotMu.RUnlock()
	var total int32
	for _, p := range lastSnapshot {
		total += p.NumFDs
	}
	return total
}

// KillProcessTree terminates a process and all its children.
func KillProcessTree(pid int32) error {
	procs, err := process.Processes()
	if err != nil {
		return err
	}

	// Map children
	children := make(map[int32][]int32)
	for _, p := range procs {
		ppid, _ := p.Ppid()
		children[ppid] = append(children[ppid], p.Pid)
	}

	// Recursive kill
	var killRecursive func(p int32)
	killRecursive = func(p int32) {
		for _, child := range children[p] {
			killRecursive(child)
		}
		proc, _ := process.NewProcess(p)
		if proc != nil {
			_ = proc.Kill()
		}
	}

	killRecursive(pid)
	return nil
}

func pruneCaches(now time.Time) {
	procCacheMu.Lock()
	for pid, meta := range procCache {
		if now.Sub(meta.LastSeen) > 10*time.Minute {
			delete(procCache, pid)
		}
	}
	procCacheMu.Unlock()

	// trustCache is based on executable path, so it can stay longer,
	// but we limit its size to 1000 entries.
	trustCacheMu.Lock()
	if len(trustCache) > 1000 {
		// Simple map reset (fastest way to reclaim memory)
		trustCache = make(map[string]procTrustInfo)
	}
	trustCacheMu.Unlock()
}
