package app

import (
	"fmt"
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
		CachedBytes:    stats.CachedBytes,
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

// GetGPUInfo returns GPU hardware information.
func (s *SysOps) GetGPUInfo() GPUInfo {
	gpu := sysops.GetGPUInfo()
	if gpu == nil || !gpu.Detected {
		return GPUInfo{Detected: false}
	}
	return GPUInfo{
		Name:     gpu.Name,
		Vendor:   gpu.Vendor,
		MemoryGB: float64(gpu.Memory) / (1024 * 1024 * 1024),
		Driver:   gpu.Driver,
		Detected: true,
	}
}

// GetBatteryInfo returns battery status information.
func (s *SysOps) GetBatteryInfo() BatteryInfo {
	bat := sysops.GetBatteryInfo()
	if bat == nil || !bat.Detected {
		return BatteryInfo{Detected: false}
	}
	return BatteryInfo{
		Percent:     bat.Percent,
		Charging:    bat.Charging,
		TimeLeftSec: bat.TimeLeftSec,
		Status:      bat.Status,
		Detected:    true,
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

// GetRecommendations returns heuristic system health recommendations.
func (s *SysOps) GetRecommendations() []SystemRecommendation {
	var recs []SystemRecommendation

	// CPU check
	cpu := s.GetCPUInfo()
	if cpu.Percent > 90 {
		recs = append(recs, SystemRecommendation{
			Category: "cpu",
			Severity: "critical",
			Message:  fmt.Sprintf("CPU usage at %.0f%% — investigate high-CPU processes", cpu.Percent),
		})
	} else if cpu.Percent > 80 {
		recs = append(recs, SystemRecommendation{
			Category: "cpu",
			Severity: "warning",
			Message:  fmt.Sprintf("CPU usage elevated at %.0f%%", cpu.Percent),
		})
	}

	// Memory check
	mem := s.GetMemoryInfo()
	if mem.UsedPercent > 90 {
		recs = append(recs, SystemRecommendation{
			Category: "memory",
			Severity: "critical",
			Message:  fmt.Sprintf("Memory usage at %.0f%% — consider closing applications or adding RAM", mem.UsedPercent),
		})
	} else if mem.UsedPercent > 85 {
		recs = append(recs, SystemRecommendation{
			Category: "memory",
			Severity: "warning",
			Message:  fmt.Sprintf("Memory usage elevated at %.0f%%", mem.UsedPercent),
		})
	}

	// Disk check (per partition)
	disk := s.GetDiskInfo()
	for _, p := range disk.Partitions {
		if p.UsedPercent > 95 {
			recs = append(recs, SystemRecommendation{
				Category: "disk",
				Severity: "critical",
				Message:  fmt.Sprintf("Disk %s at %.0f%% — critical space shortage", p.Mountpoint, p.UsedPercent),
			})
		} else if p.UsedPercent > 85 {
			recs = append(recs, SystemRecommendation{
				Category: "disk",
				Severity: "warning",
				Message:  fmt.Sprintf("Disk %s at %.0f%% — running low on space", p.Mountpoint, p.UsedPercent),
			})
		}
	}

	// Uptime check
	sysInfo := s.GetSystemInfo()
	if sysInfo.Uptime != "" {
		// Parse uptime string — format is "3d 5h 12m"
		days := 0
		fmt.Sscanf(sysInfo.Uptime, "%dd", &days)
		if days > 7 {
			recs = append(recs, SystemRecommendation{
				Category: "uptime",
				Severity: "info",
				Message:  fmt.Sprintf("System has been up for %s — consider rebooting for updates", sysInfo.Uptime),
			})
		}
	}

	if len(recs) == 0 {
		recs = append(recs, SystemRecommendation{
			Category: "general",
			Severity: "info",
			Message:  "System is healthy — no issues detected",
		})
	}

	return recs
}
