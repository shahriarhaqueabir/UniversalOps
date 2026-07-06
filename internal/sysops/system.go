package sysops

import (
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/process"
)

// SystemInfo holds general system information.
type SystemInfo struct {
	Hostname        string
	OS              string
	Platform        string
	PlatformVersion string
	KernelVersion   string
	KernelArch      string
	UptimeSeconds   uint64
	ProcessCount    int
	Virtualization  string
}

// GetSystemInfo returns general system information.
func GetSystemInfo() (*SystemInfo, error) {
	info, err := host.Info()
	if err != nil {
		return nil, err
	}

	// Count processes
	procs, err := process.Processes()
	procCount := 0
	if err == nil {
		procCount = len(procs)
	}

	virt := ""
	if info.VirtualizationSystem != "" {
		virt = info.VirtualizationSystem + " (" + info.VirtualizationRole + ")"
	}

	return &SystemInfo{
		Hostname:        info.Hostname,
		OS:              info.OS,
		Platform:        info.Platform,
		PlatformVersion: info.PlatformVersion,
		KernelVersion:   info.KernelVersion,
		KernelArch:      info.KernelArch,
		UptimeSeconds:   info.Uptime,
		ProcessCount:    procCount,
		Virtualization:  virt,
	}, nil
}
