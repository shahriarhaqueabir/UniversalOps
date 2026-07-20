package sysops

import (
	"strings"

	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
	"github.com/yusufpapurcu/wmi"
)

// Win32_VideoController represents a WMI GPU entry.
// NOTE: The struct name must exactly match the WMI class name (including the underscore)
// because yusufpapurcu/wmi uses the struct name to build the FROM clause.
type Win32_VideoController struct {
	Name                 string
	AdapterRAM           int32 // WMI returns VT_I4 (signed 32-bit); cast to uint32 to get correct bytes
	DriverVersion        string
	VideoProcessor       string
	Status               string
	VideoModeDescription string
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

// GetGPUInfo queries WMI for GPU information.
// Falls back gracefully if no GPU is detected or WMI fails.
func GetGPUInfo() *GPUInfo {
	var dst []Win32_VideoController
	// Query only essential fields to improve speed and compatibility
	q := "SELECT Name, AdapterRAM, DriverVersion, Status FROM Win32_VideoController"
	if err := wmi.Query(q, &dst); err != nil {
		common.LogWarn("GPU WMI query failed: %v", err)
		return &GPUInfo{Detected: false}
	}

	// Pick the first non-empty GPU entry that isn't a virtual display driver
	for _, v := range dst {
		name := strings.TrimSpace(v.Name)
		if name == "" || strings.Contains(strings.ToLower(name), "virtual") || strings.Contains(strings.ToLower(name), "mirror") {
			continue
		}

		// WMI returns AdapterRAM as a signed 32-bit int (VT_I4).
		// Cast to uint32 first to interpret the bit pattern correctly,
		// then widen to uint64 for the GPUInfo field.
		memBytes := uint64(uint32(v.AdapterRAM))

		// Some drivers report negative AdapterRAM or 0 if not accessible
		if memBytes == 0 && strings.Contains(strings.ToLower(name), "nvidia") {
			// Special handling for NVIDIA if RAM is missing from WMI
			// (Common in some driver versions)
		}

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
