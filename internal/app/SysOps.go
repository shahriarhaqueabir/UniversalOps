package app

import (
	"sort"

	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
	"github.com/shahriarhaqueabir/AllOpsFull/internal/sysops"
)

// SysOps exposes system operations bindings to the frontend.
type SysOps struct {
	app       *App
	collector *sysopsFacade
}

type sysopsFacade struct{}

// NewSysOps creates a new SysOps facade.
func NewSysOps(app *App) *SysOps {
	return &SysOps{
		app:       app,
		collector: &sysopsFacade{},
	}
}

// GetCPUInfo returns current CPU usage and information.
func (s *SysOps) GetCPUInfo() CPUInfo {
	stats, err := sysops.GetCPUStats()
	if err != nil {
		common.LogWarn("GetCPUStats failed: %v", err)
		return CPUInfo{}
	}
	return CPUInfo{
		Percent:       stats.Percent,
		PerCPU:        stats.PerCPU,
		ModelName:     stats.ModelName,
		LogicalCores:  stats.LogicalCores,
		PhysicalCores: stats.PhysicalCores,
		CoreCount:     stats.CoreCount,
		LoadAvg1:      stats.LoadAvg1,
		LoadAvg5:      stats.LoadAvg5,
		LoadAvg15:     stats.LoadAvg15,
	}
}

// GetMemoryInfo returns current memory and swap usage.
func (s *SysOps) GetMemoryInfo() MemoryInfo {
	stats, err := sysops.GetMemoryStats()
	if err != nil {
		common.LogWarn("GetMemoryStats failed: %v", err)
		return MemoryInfo{}
	}
	return MemoryInfo{
		TotalBytes:     stats.TotalBytes,
		AvailableBytes: stats.AvailableBytes,
		UsedBytes:      stats.UsedBytes,
		UsedPercent:    stats.UsedPercent,
		TotalGB:        float64(stats.TotalBytes) / (1024 * 1024 * 1024),
		UsedGB:         float64(stats.UsedBytes) / (1024 * 1024 * 1024),
		SwapTotal:      stats.SwapTotal,
		SwapUsed:       stats.SwapUsed,
		SwapPercent:    stats.SwapPercent,
	}
}

// GetDiskInfo returns disk partition and usage information.
func (s *SysOps) GetDiskInfo() DiskInfo {
	stats, err := sysops.GetDiskStats()
	if err != nil {
		common.LogWarn("GetDiskStats failed: %v", err)
		return DiskInfo{}
	}
	parts := make([]DiskPartition, 0, len(stats.Usage))
	for _, u := range stats.Usage {
		parts = append(parts, DiskPartition{
			Mountpoint:  u.Mountpoint,
			TotalBytes:  u.TotalBytes,
			FreeBytes:   u.FreeBytes,
			UsedBytes:   u.UsedBytes,
			UsedPercent: u.UsedPercent,
			FSType:      u.FSType,
			Device:      u.Device,
		})
	}
	return DiskInfo{Partitions: parts}
}

// GetTopProcesses returns the top N processes by CPU usage.
func (s *SysOps) GetTopProcesses(n int) []ProcessInfo {
	procs, err := sysops.GetTopProcesses(n)
	if err != nil {
		common.LogWarn("GetTopProcesses failed: %v", err)
		return nil
	}
	result := make([]ProcessInfo, 0, len(procs))
	for _, p := range procs {
		result = append(result, ProcessInfo{
			PID:    p.PID,
			Name:   p.Name,
			CPU:    p.CPU,
			Memory: p.Memory,
			MemPct: p.MemPct,
			Status: p.Status,
			NumFDs: p.NumFDs,
		})
	}
	return result
}

// GetSystemInfo returns general system information.
func (s *SysOps) GetSystemInfo() SystemInfo {
	info, err := sysops.GetSystemInfo()
	if err != nil {
		common.LogWarn("GetSystemInfo failed: %v", err)
		return SystemInfo{}
	}
	return SystemInfo{
		Hostname:        info.Hostname,
		OS:              info.OS,
		Platform:        info.Platform,
		PlatformVersion: info.PlatformVersion,
		KernelVersion:   info.KernelVersion,
		KernelArch:      info.KernelArch,
		Uptime:          common.FormatUptime(info.UptimeSeconds),
		ProcessCount:    info.ProcessCount,
		Virtualization:  info.Virtualization,
	}
}

// ── sysopsFacade ─────────────────────────────────────────────────────────────

// CollectAllStats wraps sysops.CollectAllStats for App's tick loop.
func (f *sysopsFacade) CollectAllStats() (*common.SystemStats, error) {
	return sysops.CollectAllStats()
}

// ── DevOps process passthrough ────────────────────────────────────────────────

// ListAllProcesses returns all processes by CPU usage.
func (s *SysOps) ListAllProcesses(limit int) []ProcessInfo {
	return s.GetTopProcesses(limit)
}

// GetProcessTree returns processes grouped/sorted by CPU.
func (s *SysOps) GetProcessTree() []ProcessInfo {
	procs := s.GetTopProcesses(200)
	sort.Slice(procs, func(i, j int) bool {
		return procs[i].CPU > procs[j].CPU
	})
	return procs
}
