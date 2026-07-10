package sysops

import (
	"strings"

	"github.com/yusufpapurcu/wmi"
)

// win32VideoController represents a WMI GPU entry.
type win32VideoController struct {
	Name                 string
	AdapterRAM           uint64
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
	var dst []win32VideoController
	q := wmi.CreateQuery(&dst, "")
	if err := wmi.Query(q, &dst); err != nil {
		return &GPUInfo{Detected: false}
	}

	// Pick the first non-empty GPU entry.
	for _, v := range dst {
		name := strings.TrimSpace(v.Name)
		if name == "" {
			continue
		}
		return &GPUInfo{
			Name:     name,
			Vendor:   extractVendor(name),
			Memory:   v.AdapterRAM,
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
