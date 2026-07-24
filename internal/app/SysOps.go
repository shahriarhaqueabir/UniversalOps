package app

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
	"github.com/shahriarhaqueabir/AllOpsFull/internal/sysops"
)

// safeFloat guards against NaN and Inf values that would break JSON serialization
// in Wails IPC and crash the frontend renderer.
func safeFloat(v float64) float64 {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return 0
	}
	return v
}

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
			PID:       p.PID,
			PPID:      p.PPID,
			Name:      p.Name,
			CPU:       p.CPU,
			Memory:    p.Memory,
			MemPct:    p.MemPct,
			Status:    p.Status,
			NumFDs:    p.NumFDs,
			IsSigned:  p.IsSigned,
			Publisher: p.Publisher,
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
func (s *SysOps) GetGPUInfo() (out GPUInfo) {
	defer func() {
		if r := recover(); r != nil {
			common.LogWarn("GetGPUInfo recovered panic: %v", r)
			out = GPUInfo{Detected: false}
		}
	}()

	gpu := sysops.GetGPUInfo()
	if gpu == nil || !gpu.Detected {
		return GPUInfo{Detected: false}
	}

	stats := sysops.GetGPUStats()

	out = GPUInfo{
		Name:     gpu.Name,
		Vendor:   gpu.Vendor,
		MemoryGB: safeFloat(float64(gpu.Memory) / (1024 * 1024 * 1024)),
		Driver:   gpu.Driver,
		Detected: true,
	}

	if stats != nil {
		out.Temperature = safeFloat(stats.Temperature)
		out.Utilization = safeFloat(stats.Utilization)
		out.FanSpeed = safeFloat(stats.FanSpeed)
	}

	return out
}

// GetBatteryInfo returns battery status information.
func (s *SysOps) GetBatteryInfo() (out BatteryInfo) {
	defer func() {
		if r := recover(); r != nil {
			common.LogWarn("GetBatteryInfo recovered panic: %v", r)
			out = BatteryInfo{Detected: false}
		}
	}()

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

// ProcessNode represents a process in a tree structure for visualization.
type ProcessNode struct {
	ProcessInfo
	Children []ProcessNode `json:"children,omitempty"`
}

// GetProcessTreeGraph returns a recursive hierarchical structure of processes.
func (s *SysOps) GetProcessTreeGraph() ProcessNode {
	procs, err := sysops.GetTopProcesses(0) // Get all for full mapping
	if err != nil {
		return ProcessNode{}
	}

	// 1. Create lookup map
	nodes := make(map[int32]*ProcessNode)
	for _, p := range procs {
		nodes[p.PID] = &ProcessNode{
			ProcessInfo: ProcessInfo{
				PID:    p.PID,
				PPID:   p.PPID,
				Name:   p.Name,
				CPU:    p.CPU,
				Memory: p.Memory,
				MemPct: p.MemPct,
				Status: p.Status,
				NumFDs: p.NumFDs,
			},
		}
	}

	// 2. Build tree and identify root (System Idle or first available)
	var root *ProcessNode
	for _, n := range nodes {
		if parent, ok := nodes[n.PPID]; ok && n.PID != n.PPID {
			parent.Children = append(parent.Children, *n)
		} else if root == nil || n.PID == 0 || n.PID == 4 {
			// PID 0 (Idle) or 4 (System) are common roots on Windows
			root = n
		}
	}

	if root == nil && len(nodes) > 0 {
		// Fallback to first node if no clear root found
		for _, n := range nodes {
			root = n
			break
		}
	}

	if root == nil {
		return ProcessNode{}
	}
	return *root
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

// RunSystemAction requests a safety handshake for a system-level action.
func (s *SysOps) RunSystemAction(action string) common.ActionPreview {
	command := ""
	description := ""
	risks := []string{"Potential temporary service disruption"}
	rollback := "N/A"

	switch action {
	case "flush_dns":
		command = "ipconfig /flushdns"
		description = "Flush the system DNS resolver cache"
	case "clear_arp":
		command = "arp -d *"
		description = "Clear the system ARP table"
	case "disk_cleanup":
		command = "cleanmgr /sagerun:1"
		description = "Run Windows Disk Cleanup utility"
		risks = []string{"Removal of temporary system and update files"}
	case "defrag":
		command = "defrag C: /O"
		description = "Optimize and defragment primary drive"
		risks = []string{"Significant disk I/O load during operation"}
	case "reboot":
		command = "shutdown /r /t 0"
		description = "Restart the workstation immediately"
		risks = []string{"Unsaved data will be lost", "All active sessions will terminate"}
		rollback = "Manual physical restart if it fails to boot"
	case "shutdown":
		command = "shutdown /s /t 0"
		description = "Shutdown the workstation immediately"
		risks = []string{"Unsaved data will be lost", "Remote access will be terminated"}
	default:
		command = fmt.Sprintf("Action: %s", action)
		description = fmt.Sprintf("Execute system action: %s", action)
	}

	id := common.GetHandshakeRegistry().Register(action, command, map[string]interface{}{"action": action})
	return common.ActionPreview{
		HandshakeID: id,
		Action:      action,
		Command:     command,
		Description: description,
		Risks:       risks,
		Rollback:    rollback,
	}
}

// executeSystemAction executes the actual system action.
func (s *SysOps) executeSystemAction(action string) common.SecActionResult {
	result, err := sysops.RunSystemAction(sysops.SystemAction(action))
	if err != nil {
		return common.SecActionResult{Success: false, Error: err.Error()}
	}
	return common.SecActionResult{Success: result.Success, Message: result.Message}
}

// GetHardwareInfo returns comprehensive workstation telemetry.
// RecoverPanic prevents a WMI or reflection panic from crashing the Wails process.
func (s *SysOps) GetHardwareInfo() (out HardwareInfo) {
	defer func() {
		if r := recover(); r != nil {
			common.LogWarn("GetHardwareInfo recovered panic: %v", r)
			out = HardwareInfo{} // Return zero-value instead of crashing
		}
	}()
	cpu := s.GetCPUExtended()
	gpu := s.GetGPUInfo()
	battery := s.GetBatteryInfo()

	// Baseboard info
	var board BaseboardInfo
	if b := sysops.GetBaseboardInfo(); b != nil {
		board = BaseboardInfo{
			Manufacturer: b.Manufacturer,
			Product:      b.Product,
			Version:      b.Version,
			SerialNumber: b.SerialNumber,
		}
	}

	// Sensors — always initialize as empty slice to avoid nil JSON serialization
	sensors := make([]SensorData, 0)
	if common.IsWindows() {
		// Attempt to get Fan speeds from Libre
		type Sensor struct {
			Name  string
			Value float64
			Unit  string
		}
		var dst []Sensor
		_ = common.WMIQueryNamespaceWithTimeout("SELECT Name, Value FROM Sensor WHERE SensorType='Fan' OR SensorType='Temperature'", &dst, "root\\LibreHardwareMonitor", 2*time.Second)
		for _, s := range dst {
			sensors = append(sensors, SensorData{
				Name:  s.Name,
				Type:  "Sensor",
				Value: s.Value,
				Unit:  "RPM", // Fans are usually RPM, temp handled separately in UI but we can generalize
			})
		}
	}

	return HardwareInfo{
		CPU:       cpu,
		GPU:       gpu,
		Battery:   battery,
		Baseboard: board,
		Sensors:   sensors,
	}
}

// GetCPUExtended returns extended CPU details.
func (s *SysOps) GetCPUExtended() (out CPUExtendedInfo) {
	defer func() {
		if r := recover(); r != nil {
			common.LogWarn("GetCPUExtended recovered panic: %v", r)
			out = CPUExtendedInfo{}
		}
	}()

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

// ── LHM (LibreHardwareMonitor) Management ───────────────────────────────────

// GetLHMStatus returns the current state of the bundled LHM instance.
func (s *SysOps) GetLHMStatus() LHMStatusResult {
	mgr := common.GetLHMManager()
	st := mgr.Status()
	return LHMStatusResult{
		Available:  st.Available,
		Running:    st.Running,
		NeedsAdmin: st.NeedsAdmin,
		Version:    st.Version,
		Error:      st.Error,
	}
}

// GetLHMAuthorization returns the explanation shown to the user before the UAC
// elevation prompt.  This lets the app display exactly what the admin elevation
// is required for, before triggering the Windows UAC dialog.
func (s *SysOps) GetLHMAuthorization() LHMAuthorization {
	return LHMAuthorization{
		Reason: "Universal-Ops needs admin privileges to start LibreHardwareMonitor, " +
			"which reads low-level hardware sensor data (CPU temperature, GPU " +
			"temperature, fan speeds, voltages). These sensors are restricted to " +
			"admin-level processes by Windows for security reasons.",
		Capabil: []string{
			"Real-time CPU package temperature",
			"GPU temperature, utilization, and fan speed",
			"Motherboard fan speeds and voltages",
			"NVMe SSD temperature",
		},
		Risks: []string{
			"LibreHardwareMonitor will run as a hidden background process",
			"It communicates only via local WMI — no network traffic",
			"The process stops automatically when Universal-Ops closes",
		},
		BinaryName: "LibreHardwareMonitor.exe",
		Publisher:  "LibreHardwareMonitor Contributors (MPL-2.0)",
	}
}

// DownloadLHM downloads the LHM binary (user consent already obtained via frontend).
func (s *SysOps) DownloadLHM() LHMStatusResult {
	mgr := common.GetLHMManager()
	if mgr.IsAvailable() {
		return s.GetLHMStatus()
	}

	ctx := context.Background()
	err := mgr.Download(ctx, func(downloaded int64, total int64) {
		common.LogInfo("LHM: Download progress %d / %d", downloaded, total)
	})
	if err != nil {
		st := s.GetLHMStatus()
		st.Error = fmt.Sprintf("Download failed: %v", err)
		return st
	}
	return s.GetLHMStatus()
}

// StartLHM starts the LHM process with admin elevation (triggers Windows UAC).
func (s *SysOps) StartLHM() LHMStatusResult {
	mgr := common.GetLHMManager()

	// Ensure binary exists first
	if !mgr.IsAvailable() {
		st := s.GetLHMStatus()
		st.Error = "LHM binary not found — download first"
		return st
	}

	err := mgr.Start()
	if err != nil {
		st := s.GetLHMStatus()
		st.Error = err.Error()
		return st
	}
	return s.GetLHMStatus()
}

// StopLHM terminates the LHM background process.
func (s *SysOps) StopLHM() LHMStatusResult {
	mgr := common.GetLHMManager()
	_ = mgr.Stop()
	return s.GetLHMStatus()
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

	out := ExtendedDiagnosticResult{
		Checks:    checks,
		Score:     result.Score,
		Timestamp: result.Timestamp,
	}

	// PERSIST: Save to reports table
	storage := common.GetStorage()
	if storage != nil {
		data, marshalErr := json.Marshal(out)
		if marshalErr != nil {
			common.LogError("SysOps: failed to marshal health report: %v", marshalErr)
		}
		id := fmt.Sprintf("health-%d", time.Now().Unix())
		if err := storage.InsertReport(common.ReportRecord{
			ID:        id,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Type:      "health",
			Score:     out.Score,
			DataJSON:  string(data),
		}); err != nil {
			common.LogError("SysOps: failed to persist health report: %v", err)
		}
	}

	return out
}

// ListHistoricalHealthReports returns summary of all persisted health diagnostics.
func (s *SysOps) ListHistoricalHealthReports() []common.ReportRecord {
	storage := common.GetStorage()
	if storage == nil {
		return []common.ReportRecord{}
	}
	reports, _ := storage.ListReportsByType("health")
	return reports
}

// GetHistoricalHealthReport retrieves a specific health diagnostic by ID.
func (s *SysOps) GetHistoricalHealthReport(id string) (ExtendedDiagnosticResult, error) {
	storage := common.GetStorage()
	if storage == nil {
		return ExtendedDiagnosticResult{}, fmt.Errorf("storage unavailable")
	}
	r, err := storage.GetReport(id)
	if err != nil || r == nil {
		return ExtendedDiagnosticResult{}, fmt.Errorf("report not found")
	}
	var res ExtendedDiagnosticResult
	err = json.Unmarshal([]byte(r.DataJSON), &res)
	return res, err
}

// GetInstalledPackages returns detected package managers and their installed packages.
func (s *SysOps) GetInstalledPackages() []PackageManagerData {
	sysPkgs := sysops.GetInstalledPackages()
	out := make([]PackageManagerData, 0, len(sysPkgs))
	for _, p := range sysPkgs {
		pkgs := make([]PackageInfo, 0, len(p.Packages))
		for _, pi := range p.Packages {
			pkgs = append(pkgs, PackageInfo{Name: pi.Name, Version: pi.Version})
		}
		out = append(out, PackageManagerData{
			Name:     p.Name,
			Found:    p.Found,
			Packages: pkgs,
		})
	}
	return out
}

// executeRestartService restarts a system service (internal use for handshake).
func (s *SysOps) executeRestartService(name string) common.SecActionResult {
	result, err := sysops.RestartService(name)
	if err != nil {
		return common.SecActionResult{Success: false, Message: err.Error()}
	}
	return common.SecActionResult{Success: result.Success, Message: result.Message}
}
