package sysops

import (
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/load"
)

// CPUStats holds CPU information.
type CPUStats struct {
	Percent   float64
	PerCPU    []float64
	ModelName string
	CoreCount int
	LoadAvg1  float64
	LoadAvg5  float64
	LoadAvg15 float64
}

// GetCPUStats returns current CPU usage and info.
//
// Uses a single blocking cpu.Percent call (500ms delta) for total CPU percentage,
// then estimates per-CPU values from total / core count. This avoids the ~1s wall
// time that two sequential cpu.Percent calls would incur, keeping collection within
// the 1s tick interval.
func GetCPUStats() (*CPUStats, error) {
	// Single blocking call for total CPU percentage (500ms delta)
	percent, err := cpu.Percent(500*time.Millisecond, false)
	if err != nil {
		return nil, err
	}

	cpuPercent := 0.0
	if len(percent) > 0 {
		cpuPercent = percent[0]
	}

	// Estimate per-CPU from total / core count instead of a second blocking call
	coreCount := runtime.NumCPU()
	var perCPU []float64
	if coreCount > 0 {
		perCPUEstimate := cpuPercent / float64(coreCount)
		perCPU = make([]float64, coreCount)
		for i := range perCPU {
			perCPU[i] = perCPUEstimate
		}
	}

	// Get CPU info (instant, cached by gopsutil)
	cpuInfo, err := cpu.Info()
	if err != nil {
		return nil, err
	}

	// Get load averages (instant, reads /proc or equivalent)
	loadAvg, err := load.Avg()
	var avg1, avg5, avg15 float64
	if err == nil {
		avg1 = loadAvg.Load1
		avg5 = loadAvg.Load5
		avg15 = loadAvg.Load15
	}

	modelName := ""
	if len(cpuInfo) > 0 {
		modelName = cpuInfo[0].ModelName
	}

	return &CPUStats{
		Percent:   cpuPercent,
		PerCPU:    perCPU,
		ModelName: modelName,
		CoreCount: coreCount,
		LoadAvg1:  avg1,
		LoadAvg5:  avg5,
		LoadAvg15: avg15,
	}, nil
}
