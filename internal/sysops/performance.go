package sysops

import (
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/load"
)

// PerformanceStats holds system performance metrics.
type PerformanceStats struct {
	CPUTimes     CPUTimes    `json:"cpu_times"`
	LoadAverage  LoadAverage `json:"load_average"`
	ContextSwitch uint64     `json:"context_switches"`
	Interrupts   uint64      `json:"interrupts"`
	IOWait       float64     `json:"io_wait"`
	ProcsRunning int         `json:"procs_running"`
	ProcsBlocked int         `json:"procs_blocked"`
}

// CPUTimes holds CPU time breakdown.
type CPUTimes struct {
	User   float64 `json:"user"`
	System float64 `json:"system"`
	Idle   float64 `json:"idle"`
	IOWait float64 `json:"iowait"`
	Steal  float64 `json:"steal"`
	Total  float64 `json:"total"`
}

// LoadAverage holds system load averages.
type LoadAverage struct {
	Load1  float64 `json:"load_1"`
	Load5  float64 `json:"load_5"`
	Load15 float64 `json:"load_15"`
}

// GetPerformanceStats returns system performance metrics.
func GetPerformanceStats() (*PerformanceStats, error) {
	stats := &PerformanceStats{}

	// CPU times
	times, err := cpu.Times(false)
	if err == nil && len(times) > 0 {
		t := times[0]
		total := t.User + t.System + t.Idle + t.Iowait + t.Steal
		stats.CPUTimes = CPUTimes{
			User:   t.User,
			System: t.System,
			Idle:   t.Idle,
			IOWait: t.Iowait,
			Steal:  t.Steal,
			Total:  total,
		}
		if total > 0 {
			stats.IOWait = (t.Iowait / total) * 100
		}
	}

	// Load averages
	loadAvg, err := load.Avg()
	if err == nil {
		stats.LoadAverage = LoadAverage{
			Load1:  loadAvg.Load1,
			Load5:  loadAvg.Load5,
			Load15: loadAvg.Load15,
		}
	}

	// Context switches and interrupts require /proc/stat parsing (Linux)
	// or WMI on Windows. For now, return what we have.

	return stats, nil
}
