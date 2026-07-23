package sysops

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
)

// Win32_VideoController represents a WMI GPU entry.
type Win32_VideoController struct {
	Name          string
	AdapterRAM    uint32 // WMI reports AdapterRAM as uint32 (capped at ~4 GB)
	DriverVersion string
	Status        string
}

// GPUInfo holds GPU information.
type GPUInfo struct {
	Name     string
	Vendor   string
	Memory   uint64 // AdapterRAM in bytes
	Driver   string
	Status   string
	Detected bool
}

var (
	gpuInfoCache    *GPUInfo
	gpuInfoCacheMu  sync.RWMutex
	lastGpuInfoSync time.Time
)

// GetGPUInfo queries WMI or PowerShell for GPU information.
// Implementation includes a 5-minute cache for hardware metadata to reduce
// unnecessary OS interrogation.
func GetGPUInfo() *GPUInfo {
	gpuInfoCacheMu.RLock()
	if gpuInfoCache != nil && time.Since(lastGpuInfoSync) < 5*time.Minute {
		defer gpuInfoCacheMu.RUnlock()
		return gpuInfoCache
	}
	gpuInfoCacheMu.RUnlock()

	gpuInfoCacheMu.Lock()
	defer gpuInfoCacheMu.Unlock()

	// Double-check after lock
	if gpuInfoCache != nil && time.Since(lastGpuInfoSync) < 5*time.Minute {
		return gpuInfoCache
	}

	if common.IsWindows() {
		gpu := getGPUInfoWMI()
		if gpu.Detected {
			// nvidia-smi reports true VRAM; WMI AdapterRAM is uint32 (~4GB cap)
			if trueVRAM := getNvidiaSmiVRAM(); trueVRAM > 0 {
				gpu.Memory = trueVRAM
			}
			gpuInfoCache = gpu
			lastGpuInfoSync = time.Now()
			return gpu
		}
		// Fallback to PowerShell/CIM
		gpu = getGPUInfoPowerShell()
		if gpu.Detected {
			if trueVRAM := getNvidiaSmiVRAM(); trueVRAM > 0 {
				gpu.Memory = trueVRAM
			}
			gpuInfoCache = gpu
			lastGpuInfoSync = time.Now()
		}
		return gpu
	}
	return &GPUInfo{Detected: false}
}

func getGPUInfoWMI() *GPUInfo {
	var dst []Win32_VideoController
	q := "SELECT Name, AdapterRAM, DriverVersion, Status FROM Win32_VideoController"
	if err := common.WMIQueryWithTimeout(q, &dst, 3*time.Second); err != nil {
		return &GPUInfo{Detected: false}
	}

	for _, v := range dst {
		name := strings.TrimSpace(v.Name)
		if name == "" || isVirtualDriver(name) {
			continue
		}

		memBytes := parseWMIRAM(v.AdapterRAM)

		return &GPUInfo{
			Name:     name,
			Vendor:   extractVendor(name),
			Memory:   memBytes,
			Driver:   v.DriverVersion,
			Status:   v.Status,
			Detected: true,
		}
	}
	return &GPUInfo{Detected: false}
}

func getGPUInfoPowerShell() *GPUInfo {
	cmd := common.SandboxedCommandWithConfig(common.SystemQuerySandbox(), "powershell", "-Command",
		"Get-CimInstance Win32_VideoController | Select-Object Name, AdapterRAM, DriverVersion, Status | ConvertTo-Json")
	output, err := cmd.Output()
	if err != nil {
		return &GPUInfo{Detected: false}
	}

	var results []Win32_VideoController
	// PowerShell might return a single object or an array.
	// We handle this by checking for the bracket prefix.
	jsonStr := strings.TrimSpace(string(output))
	if strings.HasPrefix(jsonStr, "[") {
		if err := json.Unmarshal(output, &results); err != nil {
			return &GPUInfo{Detected: false}
		}
	} else {
		var single Win32_VideoController
		if err := json.Unmarshal(output, &single); err == nil {
			results = append(results, single)
		}
	}

	for _, raw := range results {
		name := strings.TrimSpace(raw.Name)
		if name == "" || isVirtualDriver(name) {
			continue
		}

		return &GPUInfo{
			Name:     name,
			Vendor:   extractVendor(name),
			Memory:   parseWMIRAM(raw.AdapterRAM),
			Driver:   raw.DriverVersion,
			Status:   raw.Status,
			Detected: true,
		}
	}

	return &GPUInfo{Detected: false}
}

func isVirtualDriver(name string) bool {
	l := strings.ToLower(name)
	return strings.Contains(l, "virtual") || strings.Contains(l, "mirror") || strings.Contains(l, "microsoft remote")
}

func parseWMIRAM(val interface{}) uint64 {
	if val == nil {
		return 0
	}
	switch v := val.(type) {
	case int32:
		return uint64(uint32(v))
	case uint32:
		return uint64(v)
	case int64:
		return uint64(v)
	case uint64:
		return v
	case float64:
		return uint64(v)
	}
	return 0
}

// extractVendor infers the vendor from the GPU name string.
// GPUStats holds real-time GPU performance metrics.
type GPUStats struct {
	Temperature float64 `json:"temperature"`
	Utilization float64 `json:"utilization"`
	FanSpeed    float64 `json:"fan_speed"`
}

// GetGPUStats attempts to get real-time metrics for the first detected discrete GPU.
func GetGPUStats() *GPUStats {
	if !common.IsWindows() {
		return nil
	}

	stats := &GPUStats{}

	// 1. Try NVIDIA-SMI (Fastest and most reliable for NVIDIA cards)
	if nvidiaStats := getStatsNvidiaSmi(); nvidiaStats != nil {
		return nvidiaStats
	}

	// 2. Try LibreHardwareMonitor (Fallback for AMD/Intel or when SMI is missing)
	type Sensor struct {
		Value float64
	}
	var dst []Sensor

	_ = common.WMIQueryNamespaceWithTimeout("SELECT Value FROM Sensor WHERE SensorType='Temperature' AND Name LIKE '%GPU%'", &dst, "root\\LibreHardwareMonitor", 2*time.Second)
	if len(dst) > 0 && !math.IsNaN(dst[0].Value) && !math.IsInf(dst[0].Value, 0) {
		stats.Temperature = dst[0].Value
	}

	_ = common.WMIQueryNamespaceWithTimeout("SELECT Value FROM Sensor WHERE SensorType='Load' AND Name LIKE '%GPU Core%'", &dst, "root\\LibreHardwareMonitor", 2*time.Second)
	if len(dst) > 0 && !math.IsNaN(dst[0].Value) && !math.IsInf(dst[0].Value, 0) {
		stats.Utilization = dst[0].Value
	}

	_ = common.WMIQueryNamespaceWithTimeout("SELECT Value FROM Sensor WHERE SensorType='Fan' AND Name LIKE '%GPU%'", &dst, "root\\LibreHardwareMonitor", 2*time.Second)
	if len(dst) > 0 && !math.IsNaN(dst[0].Value) && !math.IsInf(dst[0].Value, 0) {
		stats.FanSpeed = dst[0].Value
	}

	return stats
}

// getNvidiaSmiVRAM queries nvidia-smi for the true physical VRAM in bytes.
// This overrides the WMI AdapterRAM value which is capped at ~4GB (uint32).
func getNvidiaSmiVRAM() uint64 {
	cmd := common.SandboxedCommandWithConfig(common.SystemQuerySandbox(), "nvidia-smi",
		"--query-gpu=memory.total",
		"--format=csv,noheader,nounits")
	out, err := cmd.Output()
	if err != nil {
		return 0
	}
	var vramMiB float64
	if _, err := fmt.Sscanf(strings.TrimSpace(string(out)), "%f", &vramMiB); err != nil {
		return 0
	}
	if vramMiB <= 0 {
		return 0
	}
	return uint64(vramMiB) * 1024 * 1024
}

func getStatsNvidiaSmi() *GPUStats {
	// Query: temp, utilization, fan speed
	cmd := common.SandboxedCommandWithConfig(common.SystemQuerySandbox(), "nvidia-smi",
		"--query-gpu=temperature.gpu,utilization.gpu,fan.speed",
		"--format=csv,noheader,nounits")

	out, err := cmd.Output()
	if err != nil {
		return nil
	}

	// Output format: 45, 12, 30
	parts := strings.Split(strings.TrimSpace(string(out)), ",")
	if len(parts) < 2 {
		return nil
	}

	s := &GPUStats{}
	n1, _ := fmt.Sscanf(parts[0], "%f", &s.Temperature)
	n2, _ := fmt.Sscanf(parts[1], "%f", &s.Utilization)

	if n1 < 1 || n2 < 1 {
		// Parsing failed, likely returned an error string instead of numbers
		return nil
	}

	if len(parts) > 2 {
		_, _ = fmt.Sscanf(parts[2], "%f", &s.FanSpeed)
	}

	return s
}

// extractVendor infers the vendor from the GPU name string.
func extractVendor(name string) string {
	lower := strings.ToLower(name)
	switch {
	case strings.Contains(lower, "nvidia"), strings.Contains(lower, "geforce"), strings.Contains(lower, "quadro"), strings.Contains(lower, "rtx"), strings.Contains(lower, "gtx"):
		return "NVIDIA"
	case strings.Contains(lower, "amd"), strings.Contains(lower, "radeon"):
		return "AMD"
	case strings.Contains(lower, "intel"):
		return "Intel"
	default:
		return "Unknown"
	}
}
