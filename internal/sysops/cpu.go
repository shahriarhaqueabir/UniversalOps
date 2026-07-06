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
func GetCPUStats() (*CPUStats, error) {
	// Get CPU usage percentage (with a small interval for accuracy)
	percent, err := cpu.Percent(500*time.Millisecond, false)
	if err != nil {
		return nil, err
	}

	// Get per-CPU percentages
	perCPU, err := cpu.Percent(500*time.Millisecond, true)
	if err != nil {
		perCPU = nil // non-fatal
	}

	// Get CPU info
	cpuInfo, err := cpu.Info()
	if err != nil {
		return nil, err
	}

	// Get load averages
	loadAvg, err := load.Avg()
	var avg1, avg5, avg15 float64
	if err == nil {
		avg1 = loadAvg.Load1
		avg5 = loadAvg.Load5
		avg15 = loadAvg.Load15
	}

	modelName := ""
	coreCount := runtime.NumCPU()
	if len(cpuInfo) > 0 {
		modelName = cpuInfo[0].ModelName
	}

	cpuPercent := 0.0
	if len(percent) > 0 {
		cpuPercent = percent[0]
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
