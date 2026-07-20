package sysops

import (
	"sort"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/process"
)

// processMetadata caches static process information.
type processMetadata struct {
	Name     string
	PPID     int32
	LastSeen time.Time
}

// ProcessInfo holds information about a single process.
type ProcessInfo struct {
	PID    int32
	PPID   int32
	Name   string
	CPU    float64
	Memory float32 // RSS in MB
	MemPct float32
	Status string
	NumFDs int32
}

var (
	procCache   = make(map[int32]processMetadata)
	procCacheMu sync.RWMutex

	lastSnapshot   []ProcessInfo
	lastSnapshotMu sync.RWMutex
	lastSnapshotAt time.Time
)

// UpdateProcessSnapshot performs a full system process scan and caches the result.
// This is intended to be called by a background worker to avoid blocking collectors.
func UpdateProcessSnapshot() error {
	procs, err := process.Processes()
	if err != nil {
		return err
	}

	now := time.Now()
	result := make([]ProcessInfo, 0, len(procs))

	for _, p := range procs {
		pid := p.Pid

		procCacheMu.RLock()
		meta, found := procCache[pid]
		procCacheMu.RUnlock()

		if !found {
			name, err := p.Name()
			if err != nil {
				continue
			}
			ppid, _ := p.Ppid()
			meta = processMetadata{Name: name, PPID: ppid, LastSeen: now}

			procCacheMu.Lock()
			procCache[pid] = meta
			procCacheMu.Unlock()
		}

		cpu, _ := p.CPUPercent()
		memInfo, _ := p.MemoryInfo()
		rss := float32(0)
		if memInfo != nil {
			rss = float32(memInfo.RSS) / 1024 / 1024
		}
		memPct, _ := p.MemoryPercent()
		status, _ := p.Status()
		statusStr := ""
		if len(status) > 0 {
			statusStr = status[0]
		}
		fdCount, _ := p.NumFDs()

		result = append(result, ProcessInfo{
			PID:    pid,
			PPID:   meta.PPID,
			Name:   meta.Name,
			CPU:    cpu,
			Memory: rss,
			MemPct: memPct,
			Status: statusStr,
			NumFDs: fdCount,
		})
	}

	// Sort by CPU descending
	sort.Slice(result, func(i, j int) bool {
		return result[i].CPU > result[j].CPU
	})

	lastSnapshotMu.Lock()
	lastSnapshot = result
	lastSnapshotAt = now
	lastSnapshotMu.Unlock()

	return nil
}

// GetTopProcesses returns the top N processes from the latest snapshot.
func GetTopProcesses(n int) ([]ProcessInfo, error) {
	lastSnapshotMu.RLock()
	snap := lastSnapshot
	lastSnapshotMu.RUnlock()

	if snap == nil {
		// Update outside the lock to avoid self-deadlock (UpdateProcessSnapshot
		// also acquires lastSnapshotMu.Lock).
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

	// Return a copy to prevent race conditions on slice header
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
