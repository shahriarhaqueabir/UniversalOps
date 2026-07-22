package sysops

import (
	"runtime"
	"sync"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/yusufpapurcu/wmi"
)

var (
	lastTimes   []cpu.TimesStat
	lastTimesMu sync.Mutex
)

func calculateDelta(last, current []cpu.TimesStat) []float64 {
	deltas := make([]float64, len(current))
	for i := range current {
		t1 := last[i]
		t2 := current[i]

		t1All := t1.User + t1.System + t1.Idle + t1.Nice + t1.Iowait + t1.Irq + t1.Softirq + t1.Steal
		t2All := t2.User + t2.System + t2.Idle + t2.Nice + t2.Iowait + t2.Irq + t2.Softirq + t2.Steal

		tAll := t2All - t1All
		tIdle := t2.Idle - t1.Idle

		if tAll > 0 {
			deltas[i] = (float64(tAll - tIdle) / float64(tAll)) * 100
		}
	}
	return deltas
}

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

	// PROBE: Query LibreHardwareMonitor for Temperature
	result.Temperature = getTemperatureLibre()

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

func getTemperatureLibre() float64 {
	type Sensor struct {
		Value float64
	}
	var dst []Sensor
	// 1. Try LibreHardwareMonitor
	q := "SELECT Value FROM Sensor WHERE SensorType='Temperature' AND (Name LIKE '%Package%' OR Name LIKE '%Core%')"
	if err := wmi.QueryNamespace(q, &dst, "root\\LibreHardwareMonitor"); err == nil && len(dst) > 0 {
		return dst[0].Value
	}

	// 2. Try MSAcpi_ThermalZoneTemperature (Windows Native, often requires Admin)
	type MSAcpi_ThermalZoneTemperature struct {
		CurrentTemperature uint32
	}
	var acpi []MSAcpi_ThermalZoneTemperature
	if err := wmi.QueryNamespace("SELECT CurrentTemperature FROM MSAcpi_ThermalZoneTemperature", &acpi, "root\\wmi"); err == nil && len(acpi) > 0 {
		// Value is in 10ths of degrees Kelvin
		return (float64(acpi[0].CurrentTemperature) / 10.0) - 273.15
	}

	return 0
}

// GetCPUStats returns current CPU usage and info.
func GetCPUStats() (*CPUStats, error) {
	lastTimesMu.Lock()
	defer lastTimesMu.Unlock()

	currentTimes, err := cpu.Times(true)
	if err != nil {
		return nil, err
	}

	var perCPU []float64
	if lastTimes != nil && len(lastTimes) == len(currentTimes) {
		perCPU = calculateDelta(lastTimes, currentTimes)
	} else {
		perCPU = make([]float64, len(currentTimes))
	}
	lastTimes = currentTimes

	cpuPercent := 0.0
	if len(perCPU) > 0 {
		total := 0.0
		for _, p := range perCPU {
			total += p
		}
		cpuPercent = total / float64(len(perCPU))
	}

	logicalCores, _ := cpu.Counts(true)
	physicalCores, _ := cpu.Counts(false)

	// If logicalCores is 0 (failure), fallback to runtime.NumCPU()
	if logicalCores == 0 {
		logicalCores = runtime.NumCPU()
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
