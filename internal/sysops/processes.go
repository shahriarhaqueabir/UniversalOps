package sysops

import (
	"sort"

	"github.com/shirou/gopsutil/v4/process"
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
func GetTopProcesses(n int) ([]ProcessInfo, error) {
	procs, err := process.Processes()
	if err != nil {
		return nil, err
	}

	var result []ProcessInfo
	for _, p := range procs {
		name, err := p.Name()
		if err != nil {
			continue
		}

		cpu, err := p.CPUPercent()
		if err != nil {
			cpu = 0
		}

		memInfo, err := p.MemoryInfo()
		rss := float32(0)
		memPct := float32(0)
		if err == nil && memInfo != nil {
			rss = float32(memInfo.RSS) / 1024 / 1024 // MB
		}

		memPercent, err := p.MemoryPercent()
		if err == nil {
			memPct = memPercent
		}

		status, err := p.Status()
		statusStr := ""
		if err == nil && len(status) > 0 {
			statusStr = status[0]
		}

		// Get open file descriptors count
		fds, err := p.NumFDs()
		fdCount := int32(0)
		if err == nil {
			fdCount = fds
		}

		// Get parent PID
		ppid, err := p.Ppid()
		if err != nil {
			ppid = 0
		}

		result = append(result, ProcessInfo{
			PID:    p.Pid,
			PPID:   ppid,
			Name:   name,
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

	if n > 0 && len(result) > n {
		result = result[:n]
	}

	return result, nil
}
