package app

import (
	"fmt"
	"sort"

	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
	"github.com/shahriarhaqueabir/AllOpsFull/internal/sysops"
)

// SysOps exposes system operations bindings to the frontend.
type SysOps struct {
	collector *sysopsFacade
}

type sysopsFacade struct{}

// NewSysOps creates a new SysOps facade.
func NewSysOps() *SysOps {
	return &SysOps{
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
		return []ProcessInfo{}
	}
	result := make([]ProcessInfo, 0, len(procs))
	for _, p := range procs {
		result = append(result, ProcessInfo{
			PID:    p.PID,
			PPID:   p.PPID,
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

// GetCPUExtended returns extended CPU information.
func (s *SysOps) GetCPUExtended() CPUExtendedInfo {
	stats, err := sysops.GetCPUExtended()
	if err != nil {
		common.LogWarn("GetCPUExtended failed: %v", err)
		return CPUExtendedInfo{}
	}
	perCPU := make([]PerCPUInfoData, 0, len(stats.PerCPUInfo))
	for _, p := range stats.PerCPUInfo {
		perCPU = append(perCPU, PerCPUInfoData{
			Core:      p.Core,
			Frequency: p.Frequency,
			Usage:     p.Usage,
		})
	}
	return CPUExtendedInfo{
		ModelName:    stats.ModelName,
		FrequencyMHz: stats.FrequencyMHz,
		CacheSizeKB:  stats.CacheSizeKB,
		Temperature:  stats.Temperature,
		PerCPUInfo:   perCPU,
	}
}

// GetDiskIO returns disk I/O throughput statistics.
func (s *SysOps) GetDiskIO() DiskIOData {
	stats, err := sysops.GetDiskIO()
	if err != nil {
		common.LogWarn("GetDiskIO failed: %v", err)
		return DiskIOData{}
	}
	disks := make([]DiskIOEntry, 0, len(stats.Disks))
	for _, d := range stats.Disks {
		disks = append(disks, DiskIOEntry{
			Name:       d.Name,
			ReadBytes:  d.ReadBytes,
			WriteBytes: d.WriteBytes,
			ReadCount:  d.ReadCount,
			WriteCount: d.WriteCount,
		})
	}
	return DiskIOData{
		Disks:      disks,
		TotalRead:  stats.TotalRead,
		TotalWrite: stats.TotalWrite,
	}
}

// GetLoggedInUsers returns all currently logged-in users.
func (s *SysOps) GetLoggedInUsers() []LoggedInUserData {
	users, err := sysops.GetLoggedInUsers()
	if err != nil {
		common.LogWarn("GetLoggedInUsers failed: %v", err)
		return []LoggedInUserData{}
	}
	var result []LoggedInUserData
	for _, u := range users {
		result = append(result, LoggedInUserData{
			User:     u.User,
			Terminal: u.Terminal,
			Host:     u.Host,
			Started:  u.Started,
		})
	}
	return result
}

// GetPerformanceStats returns system performance metrics.
func (s *SysOps) GetPerformanceStats() PerformanceData {
	stats, err := sysops.GetPerformanceStats()
	if err != nil {
		common.LogWarn("GetPerformanceStats failed: %v", err)
		return PerformanceData{}
	}
	return PerformanceData{
		CPUTimes: CPUTimesData{
			User:   stats.CPUTimes.User,
			System: stats.CPUTimes.System,
			Idle:   stats.CPUTimes.Idle,
			IOWait: stats.CPUTimes.IOWait,
			Steal:  stats.CPUTimes.Steal,
			Total:  stats.CPUTimes.Total,
		},
		LoadAverage: LoadAverageData{
			Load1:  stats.LoadAverage.Load1,
			Load5:  stats.LoadAverage.Load5,
			Load15: stats.LoadAverage.Load15,
		},
		IOWait: stats.IOWait,
	}
}

// RunSystemAction executes a system action.
func (s *SysOps) RunSystemAction(action string) ActionResult {
	result, err := sysops.RunSystemAction(sysops.SystemAction(action))
	if err != nil {
		return ActionResult{
			Action:  action,
			Success: false,
			Message: err.Error(),
		}
	}
	return ActionResult{
		Action:  result.Action,
		Success: result.Success,
		Message: result.Message,
		Output:  result.Output,
	}
}

// GetSystemLogs retrieves OS system logs.
func (s *SysOps) GetSystemLogs(n int, source string) SystemLogsResultData {
	result, err := sysops.GetSystemLogs(n, source)
	if err != nil {
		common.LogWarn("GetSystemLogs failed: %v", err)
		return SystemLogsResultData{}
	}
	entries := make([]SystemLogEntry, 0, len(result.Entries))
	for _, e := range result.Entries {
		entries = append(entries, SystemLogEntry{
			Timestamp: e.Timestamp,
			Level:     e.Level,
			Source:    e.Source,
			Message:   e.Message,
		})
	}
	return SystemLogsResultData{
		Entries: entries,
		Source:  result.Source,
		Total:   result.Total,
	}
}

// GetScheduledTasks returns all scheduled tasks.
func (s *SysOps) GetScheduledTasks() []ScheduledTaskData {
	tasks, err := sysops.GetScheduledTasks()
	if err != nil {
		common.LogWarn("GetScheduledTasks failed: %v", err)
		return []ScheduledTaskData{}
	}
	var result []ScheduledTaskData
	for _, t := range tasks {
		result = append(result, ScheduledTaskData{
			Name:     t.Name,
			Schedule: t.Schedule,
			Command:  t.Command,
			Enabled:  t.Enabled,
			NextRun:  t.NextRun,
		})
	}
	return result
}

// RunExtendedDiagnostics runs a set of system health checks and returns a score from 0-100.
func (s *SysOps) RunExtendedDiagnostics() ExtendedDiagnosticResult {
	result, err := sysops.RunExtendedDiagnostics()
	if err != nil {
		common.LogWarn("RunExtendedDiagnostics failed: %v", err)
		return ExtendedDiagnosticResult{}
	}
	checks := make([]DiagnosticCheckData, 0, len(result.Checks))
	for _, c := range result.Checks {
		checks = append(checks, DiagnosticCheckData{
			Name:    c.Name,
			Status:  c.Status,
			Message: c.Message,
			Value:   c.Value,
		})
	}
	return ExtendedDiagnosticResult{
		Checks:    checks,
		Score:     result.Score,
		Timestamp: result.Timestamp,
	}
}

// executeRestartService restarts a system service (internal use for handshake).
func (s *SysOps) executeRestartService(name string) SecActionResult {
	result, err := sysops.RestartService(name)
	if err != nil {
		return SecActionResult{Success: false, Message: err.Error()}
	}
	return SecActionResult{Success: result.Success, Message: result.Message}
}
