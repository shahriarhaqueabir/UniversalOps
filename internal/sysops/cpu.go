package sysops

import (
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/load"
)

// CPUStats holds CPU information.
type CPUStats struct {
	Percent       float64
	PerCPU        []float64
	ModelName     string
	LogicalCores  int
	PhysicalCores int
	CoreCount     int // Deprecated: use LogicalCores
	LoadAvg1      float64
	LoadAvg5      float64
	LoadAvg15     float64
}

// CPUExtendedStats holds extended CPU information.
type CPUExtendedStats struct {
	ModelName    string        `json:"model_name"`
	FrequencyMHz float64       `json:"frequency_mhz"`
	CacheSizeKB  int32         `json:"cache_size_kb"`
	Temperature  float64       `json:"temperature"`
	PerCPUInfo   []PerCPUInfo  `json:"per_cpu_info"`
}

// PerCPUInfo holds per-core detailed info.
type PerCPUInfo struct {
	Core      int     `json:"core"`
	Frequency float64 `json:"frequency_mhz"`
	Usage     float64 `json:"usage_percent"`
}

// GetCPUExtended returns extended CPU information including model, frequency, temperature, and per-core stats.
func GetCPUExtended() (*CPUExtendedStats, error) {
	info, err := cpu.Info()
	if err != nil {
		return nil, err
	}

	result := &CPUExtendedStats{}
	if len(info) > 0 {
		result.ModelName = info[0].ModelName
		result.FrequencyMHz = info[0].Mhz
		result.CacheSizeKB = info[0].CacheSize
	}

	// Temperature not available in gopsutil v4 — leave as 0

	// Per-CPU info
	perCPU, _ := cpu.Percent(0, true)
	for i, ci := range info {
		pInfo := PerCPUInfo{Core: i, Frequency: ci.Mhz}
		if i < len(perCPU) {
			pInfo.Usage = perCPU[i]
		}
		result.PerCPUInfo = append(result.PerCPUInfo, pInfo)
	}

	return result, nil
}

// GetCPUStats returns current CPU usage and info.
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

	logicalCores, _ := cpu.Counts(true)
	physicalCores, _ := cpu.Counts(false)

	// If logicalCores is 0 (failure), fallback to runtime.NumCPU()
	if logicalCores == 0 {
		logicalCores = runtime.NumCPU()
	}

	var perCPU []float64
	if logicalCores > 0 {
		perCPUEstimate := cpuPercent / float64(logicalCores)
		perCPU = make([]float64, logicalCores)
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
		Percent:       cpuPercent,
		PerCPU:        perCPU,
		ModelName:     modelName,
		LogicalCores:  logicalCores,
		PhysicalCores: physicalCores,
		CoreCount:     logicalCores,
		LoadAvg1:      avg1,
		LoadAvg5:      avg5,
		LoadAvg15:     avg15,
	}, nil
}
