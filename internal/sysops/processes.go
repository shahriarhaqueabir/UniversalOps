package sysops

import (
	"sort"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/process"
)

// processMetadata caches static process information.
type processMetadata struct {
	Name      string
	PPID      int32
	LastSeen  time.Time
}

var (
	procCache   = make(map[int32]processMetadata)
	procCacheMu sync.RWMutex
)

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

// GetTopProcesses returns the top N processes by CPU usage.
// Uses a metadata cache to reduce overhead of redundant syscalls for static info.
func GetTopProcesses(n int) ([]ProcessInfo, error) {
	procs, err := process.Processes()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	var result []ProcessInfo

	// Pre-size result slice to avoid reallocations
	result = make([]ProcessInfo, 0, len(procs))

	for _, p := range procs {
		pid := p.Pid

		// 1. Resolve Metadata (Cache-First)
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
		} else {
			// Update last seen to prevent cache eviction
			procCacheMu.Lock()
			meta.LastSeen = now
			procCache[pid] = meta
			procCacheMu.Unlock()
		}

		// 2. Fetch Dynamic Metrics
		// Note: CPUPercent(0) returns usage since last call for this *process object.
		// If we create a new object every time, we might get 0 or weird values.
		// However, gopsutil v4 handles some internal caching if configured.
		cpu, _ := p.CPUPercent()

		memInfo, err := p.MemoryInfo()
		rss := float32(0)
		if err == nil && memInfo != nil {
			rss = float32(memInfo.RSS) / 1024 / 1024 // MB
		}

		memPct, _ := p.MemoryPercent()

		status, err := p.Status()
		statusStr := ""
		if err == nil && len(status) > 0 {
			statusStr = status[0]
		}

		// 3. Optional but expensive: NumFDs
		// Only fetch for top processes? No, we need it for all if we want to sort accurately by something else.
		// For now, keep it but monitor performance.
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

	// 4. Periodic Cache Eviction (cleanup dead PIDs)
	go func() {
		procCacheMu.Lock()
		defer procCacheMu.Unlock()
		if len(procCache) > 2000 { // Only cleanup if cache grows large
			for pid, meta := range procCache {
				if now.Sub(meta.LastSeen) > 5*time.Minute {
					delete(procCache, pid)
				}
			}
		}
	}()

	// Sort by CPU descending
	sort.Slice(result, func(i, j int) bool {
		return result[i].CPU > result[j].CPU
	})

	if n > 0 && len(result) > n {
		result = result[:n]
	}

	return result, nil
}
