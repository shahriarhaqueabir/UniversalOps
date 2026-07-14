# SysOps Rework Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Rework SysOps into a comprehensive sysadmin workstation — 14 categories covering inspection → diagnosis → action for daily engineering/sysadmin workflows.

**Architecture:** Backend adds new gopsutil-based collection methods + OS-specific shell execution for actions/logs/packages. Frontend restructures from 3 tabs to 14 category sections organized as a scrollable dashboard with a category sidebar navigator. Each category maps to a specific sysadmin command set.

**Tech Stack:** Go 1.26.5, gopsutil/v4, React 19, TypeScript, Tailwind v4, Radix UI, Recharts, Lucide icons, Wails v2 bindings

---

## File Structure

### Backend (Go)
| File | Responsibility |
|------|---------------|
| `internal/sysops/cpu.go` | **Extend**: Add `GetCPUExtended()` — model, frequency, temperature, affinity |
| `internal/sysops/memory.go` | **Extend**: Add memory breakdown (buffers, slab, huge pages via `/proc/meminfo` on Linux) |
| `internal/sysops/disk.go` | **Extend**: Add `GetDiskIO()` — disk I/O throughput via `disk.IOCounters()` |
| `internal/sysops/system.go` | **Extend**: Add `GetLoggedInUsers()`, `GetTimeInfo()` |
| `internal/sysops/performance.go` | **New**: CPU times (iowait, system, user), context switches, interrupts |
| `internal/sysops/actions.go` | **New**: System actions (reboot, shutdown, flush DNS, sleep, clear temp, package cache, update) |
| `internal/sysops/logs.go` | **New**: OS-native log retrieval (Windows Event Log, Linux journald/dmesg) |
| `internal/sysops/packages.go` | **New**: Package manager detection and listing (apt, dnf, winget, choco, pacman) |
| `internal/sysops/scheduler.go` | **New**: Cron/schtasks listing |
| `internal/sysops/diagnostics.go` | **New**: Extended health diagnostics (SMART, failed services, kernel errors, update check) |
| `internal/app/SysOps.go` | **Extend**: Bind all new methods to Wails frontend |
| `internal/app/Types.go` | **Extend**: Add new TypeScript-mirrored types |

### Frontend (React/TypeScript)
| File | Responsibility |
|------|---------------|
| `src/types/index.ts` | **Extend**: Add new SysOps-specific types |
| `src/pages/SysOps.tsx` | **Major rewrite**: 14-category layout with sidebar navigator |
| `src/pages/SysOps/SystemInfoTab.tsx` | **New**: System Information section |
| `src/pages/SysOps/CpuTab.tsx` | **New**: CPU monitoring section |
| `src/pages/SysOps/MemoryTab.tsx` | **New**: Memory monitoring section |
| `src/pages/SysOps/DiskTab.tsx` | **New**: Disk monitoring section |
| `src/pages/SysOps/ProcessesTab.tsx` | **New**: Process management section |
| `src/pages/SysOps/ServicesTab.tsx` | **New**: Service management section (delegates to DevOps backend) |
| `src/pages/SysOps/LogsTab.tsx` | **New**: System log viewer |
| `src/pages/SysOps/StorageTab.tsx` | **New**: Storage health + mount info |
| `src/pages/SysOps/UsersTab.tsx` | **New**: User/permission viewer |
| `src/pages/SysOps/PerformanceTab.tsx` | **New**: I/O wait, context switches, interrupts, load average |
| `src/pages/SysOps/PackageManagerTab.tsx` | **New**: Package manager integration |
| `src/pages/SysOps/SchedulerTab.tsx` | **New**: Cron/schtasks viewer |
| `src/pages/SysOps/DiagnosticsTab.tsx` | **New**: One-click health check |
| `src/pages/SysOps/ActionsTab.tsx` | **New**: System action buttons |
| `src/components/sysops/HealthBadge.tsx` | **New**: Reusable health status badge |
| `src/components/sysops/CommandOutput.tsx` | **New**: Reusable command output display |

---

## Spec Coverage Matrix

| # | Category | Backend | Frontend | Priority |
|---|----------|---------|----------|----------|
| 1 | System Information | `GetSystemInfo()` + `GetLoggedInUsers()` + `GetCPUExtended()` | SystemInfoTab | P0 |
| 2 | CPU | `GetCPUInfo()` + `GetCPUExtended()` | CpuTab | P0 |
| 3 | Memory | `GetMemoryInfo()` + extended breakdown | MemoryTab | P0 |
| 4 | Disk | `GetDiskInfo()` + `GetDiskIO()` | DiskTab | P0 |
| 5 | Processes | `ListAllProcesses()` + `GetProcessTree()` (existing) | ProcessesTab | P0 |
| 6 | Services | `DevOps.GetServices()` + `DevOps.ControlService()` (reuse) | ServicesTab | P1 |
| 7 | Logs | `GetSystemLogs()` (new) | LogsTab | P1 |
| 8 | Storage | `GetDiskInfo()` + `GetDiskIO()` + SMART check | StorageTab | P1 |
| 9 | Users & Permissions | `host.Users()` (gopsutil) + `GetLoggedInUsers()` | UsersTab | P2 |
| 10 | Performance | `GetPerformanceStats()` (new) | PerformanceTab | P1 |
| 11 | Package Management | `GetInstalledPackages()` (new) | PackageManagerTab | P2 |
| 12 | Scheduler | `GetScheduledTasks()` (new) | SchedulerTab | P2 |
| 13 | Diagnostics | `RunDiagnostics()` (new, extends workflows.go) | DiagnosticsTab | P0 |
| 14 | Actions | `RunSystemAction()` (new) | ActionsTab | P1 |

---

## Known Constraints

1. **BIOS/Serial Number**: Not available via gopsutil. Would need WMI (`Win32_BIOS`) on Windows or `dmidecode` on Linux. Plan: Skip for P0, add as P2 with platform-specific shell commands.
2. **SMART Health**: Needs `smartctl` CLI (not in gopsutil). Plan: Shell out to `smartctl` if available, show "not available" if not.
3. **CPU Temperature**: `cpu.Temperatures()` in gopsutil v4 works on Linux (thermal zones) but often fails on Windows. Plan: Try gopsutil first, fallback to "N/A".
4. **Disk I/O Rate**: `disk.IOCounters()` returns cumulative bytes, not rates. Plan: Store previous snapshot, compute delta for rate display.
5. **Package Management**: Highly OS-specific. Plan: Detect package manager, shell out with appropriate command.
6. **System Logs**: Windows Event Log needs PowerShell. Linux journald/dmesg needs shell. Plan: Platform-specific shell commands.
7. **Top Disk IO Bug**: Current frontend references `read_bytes`/`write_bytes` on ProcessInfo but backend doesn't populate them. Plan: Remove broken "Top Disk IO" card, replace with aggregate disk I/O from `disk.IOCounters()`.

---

## Phase 1: Backend Foundation (New Go Methods)

### Task 1.1: Extended CPU Info

**Files:**
- Modify: `internal/sysops/cpu.go`
- Modify: `internal/app/SysOps.go`
- Modify: `internal/app/Types.go`

- [ ] **Step 1: Add `CPUExtendedStats` struct to `cpu.go`**

```go
// CPUExtendedStats holds extended CPU information.
type CPUExtendedStats struct {
	ModelName    string   `json:"model_name"`
	FrequencyMHz float64  `json:"frequency_mhz"`
	CacheSizeKB  int32    `json:"cache_size_kb"`
	Temperature  float64  `json:"temperature"` // Celsius, 0 if unavailable
	PerCPUInfo   []PerCPUInfo `json:"per_cpu_info"`
}

// PerCPUInfo holds per-core detailed info.
type PerCPUInfo struct {
	Core       int     `json:"core"`
	Frequency  float64 `json:"frequency_mhz"`
	Usage      float64 `json:"usage_percent"`
}
```

- [ ] **Step 2: Add `GetCPUExtended()` function to `cpu.go`**

```go
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

	// Try to get temperature (not available on all platforms)
	temps, err := cpu.Temperatures()
	if err == nil && len(temps) > 0 {
		result.Temperature = temps[0].Temp
	}

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
```

- [ ] **Step 3: Add binding in `app/SysOps.go`**

```go
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
```

- [ ] **Step 4: Add types in `app/Types.go`**

```go
type CPUExtendedInfo struct {
	ModelName    string          `json:"model_name"`
	FrequencyMHz float64         `json:"frequency_mhz"`
	CacheSizeKB  int32           `json:"cache_size_kb"`
	Temperature  float64         `json:"temperature"`
	PerCPUInfo   []PerCPUInfoData `json:"per_cpu_info"`
}

type PerCPUInfoData struct {
	Core      int     `json:"core"`
	Frequency float64 `json:"frequency_mhz"`
	Usage     float64 `json:"usage_percent"`
}
```

- [ ] **Step 5: Write test for `GetCPUExtended()`**

Create `internal/sysops/cpu_extended_test.go`:

```go
package sysops

import "testing"

func TestGetCPUExtended(t *testing.T) {
	stats, err := GetCPUExtended()
	if err != nil {
		t.Fatalf("GetCPUExtended returned error: %v", err)
	}
	if stats.ModelName == "" {
		t.Error("ModelName is empty")
	}
	if stats.FrequencyMHz <= 0 {
		t.Errorf("FrequencyMHz should be > 0, got %f", stats.FrequencyMHz)
	}
	if len(stats.PerCPUInfo) == 0 {
		t.Error("PerCPUInfo is empty")
	}
}
```

- [ ] **Step 6: Run test**

Run: `cd E:\Projects\projectx\AllOpsFull && go test ./internal/sysops/ -run TestGetCPUExtended -v`
Expected: PASS

- [ ] **Step 7: Commit**

```bash
git add internal/sysops/cpu.go internal/sysops/cpu_extended_test.go internal/app/SysOps.go internal/app/Types.go
git commit -m "feat(sysops): add GetCPUExtended with model, frequency, temperature, per-core info"
```

---

### Task 1.2: Disk I/O Stats

**Files:**
- Modify: `internal/sysops/disk.go`
- Modify: `internal/app/SysOps.go`
- Modify: `internal/app/Types.go`

- [ ] **Step 1: Add `DiskIOStats` struct and `GetDiskIO()` to `disk.go`**

```go
type DiskIOStat struct {
	Name        string `json:"name"`
	ReadBytes   uint64 `json:"read_bytes"`
	WriteBytes  uint64 `json:"write_bytes"`
	ReadCount   uint64 `json:"read_count"`
	WriteCount  uint64 `json:"write_count"`
	ReadTimeMs  uint64 `json:"read_time_ms"`
	WriteTimeMs uint64 `json:"write_time_ms"`
}

type DiskIOStats struct {
	Disks      []DiskIOStat `json:"disks"`
	TotalRead  uint64       `json:"total_read_bytes"`
	TotalWrite uint64       `json:"total_write_bytes"`
}

func GetDiskIO() (*DiskIOStats, error) {
	ioCounters, err := disk.IOCounters()
	if err != nil {
		return nil, err
	}

	stats := &DiskIOStats{}
	for name, counter := range ioCounters {
		stat := DiskIOStat{
			Name:        name,
			ReadBytes:   counter.ReadBytes,
			WriteBytes:  counter.WriteBytes,
			ReadCount:   counter.ReadCount,
			WriteCount:  counter.WriteCount,
			ReadTimeMs:  counter.ReadTime,
			WriteTimeMs: counter.WriteTime,
		}
		stats.Disks = append(stats.Disks, stat)
		stats.TotalRead += counter.ReadBytes
		stats.TotalWrite += counter.WriteBytes
	}

	return stats, nil
}
```

- [ ] **Step 2: Add binding in `app/SysOps.go`**

```go
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
```

- [ ] **Step 3: Add types in `app/Types.go`**

```go
type DiskIOEntry struct {
	Name       string `json:"name"`
	ReadBytes  uint64 `json:"read_bytes"`
	WriteBytes uint64 `json:"write_bytes"`
	ReadCount  uint64 `json:"read_count"`
	WriteCount uint64 `json:"write_count"`
}

type DiskIOData struct {
	Disks      []DiskIOEntry `json:"disks"`
	TotalRead  uint64        `json:"total_read_bytes"`
	TotalWrite uint64        `json:"total_write_bytes"`
}
```

- [ ] **Step 4: Write test**

```go
func TestGetDiskIO(t *testing.T) {
	stats, err := GetDiskIO()
	if err != nil {
		t.Fatalf("GetDiskIO returned error: %v", err)
	}
	if len(stats.Disks) == 0 {
		t.Error("No disk I/O stats returned")
	}
}
```

- [ ] **Step 5: Run test**

Run: `cd E:\Projects\projectx\AllOpsFull && go test ./internal/sysops/ -run TestGetDiskIO -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add internal/sysops/disk.go internal/app/SysOps.go internal/app/Types.go
git commit -m "feat(sysops): add GetDiskIO with per-disk read/write bytes and IOPS"
```

---

### Task 1.3: Logged-in Users

**Files:**
- Modify: `internal/sysops/system.go`
- Modify: `internal/app/SysOps.go`
- Modify: `internal/app/Types.go`

- [ ] **Step 1: Add `GetLoggedInUsers()` to `system.go`**

```go
type LoggedInUser struct {
	User     string `json:"user"`
	Terminal string `json:"terminal"`
	Host     string `json:"host"`
	Started  string `json:"started"`
}

func GetLoggedInUsers() ([]LoggedInUser, error) {
	users, err := host.Users()
	if err != nil {
		return nil, err
	}

	var result []LoggedInUser
	for _, u := range users {
		result = append(result, LoggedInUser{
			User:     u.User,
			Terminal: u.Terminal,
			Host:     u.Host,
			Started:  fmt.Sprintf("%d", u.Started),
		})
	}
	return result, nil
}
```

- [ ] **Step 2: Add binding in `app/SysOps.go`**

```go
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
```

- [ ] **Step 3: Add types in `app/Types.go`**

```go
type LoggedInUserData struct {
	User     string `json:"user"`
	Terminal string `json:"terminal"`
	Host     string `json:"host"`
	Started  string `json:"started"`
}
```

- [ ] **Step 4: Write test**

```go
func TestGetLoggedInUsers(t *testing.T) {
	users, err := GetLoggedInUsers()
	if err != nil {
		t.Fatalf("GetLoggedInUsers returned error: %v", err)
	}
	// At least the current user should be logged in
	t.Logf("Found %d logged-in users", len(users))
}
```

- [ ] **Step 5: Run test**

Run: `cd E:\Projects\projectx\AllOpsFull && go test ./internal/sysops/ -run TestGetLoggedInUsers -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add internal/sysops/system.go internal/app/SysOps.go internal/app/Types.go
git commit -m "feat(sysops): add GetLoggedInUsers via gopsutil host.Users()"
```

---

### Task 1.4: Performance Stats

**Files:**
- Create: `internal/sysops/performance.go`
- Modify: `internal/app/SysOps.go`
- Modify: `internal/app/Types.go`

- [ ] **Step 1: Create `performance.go` with CPU times, context switches, interrupts**

```go
package sysops

import (
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/load"
)

type PerformanceStats struct {
	CPUTimes      CPUTimes     `json:"cpu_times"`
	LoadAverage   LoadAverage  `json:"load_average"`
	ContextSwitch uint64       `json:"context_switches"`
	Interrupts    uint64       `json:"interrupts"`
	IOWait        float64      `json:"io_wait"`
	ProcsRunning  int          `json:"procs_running"`
	ProcsBlocked  int          `json:"procs_blocked"`
}

type CPUTimes struct {
	User   float64 `json:"user"`
	System float64 `json:"system"`
	Idle   float64 `json:"idle"`
	IOWait float64 `json:"iowait"`
	Steal  float64 `json:"steal"`
	Total  float64 `json:"total"`
}

type LoadAverage struct {
	Load1  float64 `json:"load_1"`
	Load5  float64 `json:"load_5"`
	Load15 float64 `json:"load_15"`
}

func GetPerformanceStats() (*PerformanceStats, error) {
	stats := &PerformanceStats{}

	// CPU times
	times, err := cpu.Times(false)
	if err == nil && len(times) > 0 {
		t := times[0]
		total := t.User + t.System + t.Idle + t.Iowait + t.Steal
		stats.CPUTimes = CPUTimes{
			User:   t.User,
			System: t.System,
			Idle:   t.Idle,
			IOWait: t.Iowait,
			Steal:  t.Steal,
			Total:  total,
		}
		if total > 0 {
			stats.IOWait = (t.Iowait / total) * 100
		}
	}

	// Load averages
	loadAvg, err := load.Avg()
	if err == nil {
		stats.LoadAverage = LoadAverage{
			Load1:  loadAvg.Load1,
			Load5:  loadAvg.Load5,
			Load15: loadAvg.Load15,
		}
	}

	// Note: Context switches and interrupts require /proc/stat parsing on Linux
	// or WMI on Windows. For now, return what we have.
	// TODO: Add platform-specific /proc/stat parsing

	return stats, nil
}
```

- [ ] **Step 2: Add binding in `app/SysOps.go`**

```go
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
```

- [ ] **Step 3: Add types in `app/Types.go`**

```go
type CPUTimesData struct {
	User   float64 `json:"user"`
	System float64 `json:"system"`
	Idle   float64 `json:"idle"`
	IOWait float64 `json:"iowait"`
	Steal  float64 `json:"steal"`
	Total  float64 `json:"total"`
}

type LoadAverageData struct {
	Load1  float64 `json:"load_1"`
	Load5  float64 `json:"load_5"`
	Load15 float64 `json:"load_15"`
}

type PerformanceData struct {
	CPUTimes    CPUTimesData    `json:"cpu_times"`
	LoadAverage LoadAverageData `json:"load_average"`
	IOWait      float64         `json:"io_wait"`
}
```

- [ ] **Step 4: Write test**

Create `internal/sysops/performance_test.go`:

```go
package sysops

import "testing"

func TestGetPerformanceStats(t *testing.T) {
	stats, err := GetPerformanceStats()
	if err != nil {
		t.Fatalf("GetPerformanceStats returned error: %v", err)
	}
	if stats.CPUTimes.Total <= 0 {
		t.Error("CPUTimes.Total should be > 0")
	}
	if stats.LoadAverage.Load1 < 0 {
		t.Error("LoadAverage.Load1 should be >= 0")
	}
}
```

- [ ] **Step 5: Run test**

Run: `cd E:\Projects\projectx\AllOpsFull && go test ./internal/sysops/ -run TestGetPerformanceStats -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add internal/sysops/performance.go internal/sysops/performance_test.go internal/app/SysOps.go internal/app/Types.go
git commit -m "feat(sysops): add GetPerformanceStats with CPU times, load average, IO wait"
```

---

### Task 1.5: System Actions

**Files:**
- Create: `internal/sysops/actions.go`
- Modify: `internal/app/SysOps.go`
- Modify: `internal/app/Types.go`

- [ ] **Step 1: Create `actions.go`**

```go
package sysops

import (
	"fmt"
	"os/exec"
	"runtime"
)

type SystemAction string

const (
	ActionReboot         SystemAction = "reboot"
	ActionShutdown       SystemAction = "shutdown"
	ActionSleep          SystemAction = "sleep"
	ActionHibernate      SystemAction = "hibernate"
	ActionFlushDNS       SystemAction = "flush_dns"
	ActionClearTemp      SystemAction = "clear_temp"
	ActionCleanPkgCache  SystemAction = "clean_pkg_cache"
	ActionSystemUpdate   SystemAction = "system_update"
)

type ActionResult struct {
	Action  string `json:"action"`
	Success bool   `json:"success"`
	Message string `json:"message"`
	Output  string `json:"output"`
}

func RunSystemAction(action SystemAction) (*ActionResult, error) {
	result := &ActionResult{Action: string(action)}

	var cmd *exec.Cmd

	switch action {
	case ActionReboot:
		if runtime.GOOS == "windows" {
			cmd = exec.Command("shutdown", "/r", "/t", "0")
		} else {
			cmd = exec.Command("sudo", "shutdown", "-r", "now")
		}
	case ActionShutdown:
		if runtime.GOOS == "windows" {
			cmd = exec.Command("shutdown", "/s", "/t", "0")
		} else {
			cmd = exec.Command("sudo", "shutdown", "-h", "now")
		}
	case ActionSleep:
		if runtime.GOOS == "windows" {
			cmd = exec.Command("rundll32.exe", "powrprof.dll,SetSuspendState", "0,1,0")
		} else {
			cmd = exec.Command("systemctl", "suspend")
		}
	case ActionHibernate:
		if runtime.GOOS == "windows" {
			cmd = exec.Command("rundll32.exe", "powrprof.dll,SetSuspendState", "1,1,0")
		} else {
			cmd = exec.Command("systemctl", "hibernate")
		}
	case ActionFlushDNS:
		if runtime.GOOS == "windows" {
			cmd = exec.Command("ipconfig", "/flushdns")
		} else {
			cmd = exec.Command("sudo", "resolvectl", "flush-caches")
		}
	case ActionClearTemp:
		if runtime.GOOS == "windows" {
			cmd = exec.Command("powershell", "-Command", "Remove-Item -Recurse -Force $env:TEMP\\* -ErrorAction SilentlyContinue")
		} else {
			cmd = exec.Command("sudo", "rm", "-rf", "/tmp/*")
		}
	case ActionCleanPkgCache:
		if runtime.GOOS == "windows" {
			cmd = exec.Command("powershell", "-Command", "winget source update")
		} else {
			// Try apt first, then dnf
			cmd = exec.Command("sudo", "apt-get", "clean")
		}
	case ActionSystemUpdate:
		if runtime.GOOS == "windows" {
			cmd = exec.Command("powershell", "-Command", "winget upgrade --all --accept-package-agreements --accept-source-agreements")
		} else {
			cmd = exec.Command("sudo", "apt-get", "update", "&&", "sudo", "apt-get", "upgrade", "-y")
		}
	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}

	output, err := cmd.CombinedOutput()
	result.Output = string(output)
	if err != nil {
		result.Success = false
		result.Message = fmt.Sprintf("Action failed: %v", err)
		return result, err
	}

	result.Success = true
	result.Message = fmt.Sprintf("Action '%s' completed successfully", action)
	return result, nil
}
```

- [ ] **Step 2: Add binding in `app/SysOps.go`**

```go
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
```

- [ ] **Step 3: Add types in `app/Types.go`**

```go
type ActionResult struct {
	Action  string `json:"action"`
	Success bool   `json:"success"`
	Message string `json:"message"`
	Output  string `json:"output"`
}
```

- [ ] **Step 4: Write test**

Create `internal/sysops/actions_test.go`:

```go
package sysops

import "testing"

func TestRunSystemAction_FlushDNS(t *testing.T) {
	// Only test flush DNS as it's the safest action
	if testing.Short() {
		t.Skip("Skipping system action test in short mode")
	}
	result, err := RunSystemAction(ActionFlushDNS)
	if err != nil {
		t.Logf("FlushDNS returned error (may be expected without sudo): %v", err)
	}
	t.Logf("Result: %+v", result)
}
```

- [ ] **Step 5: Run test**

Run: `cd E:\Projects\projectx\AllOpsFull && go test ./internal/sysops/ -run TestRunSystemAction -v`
Expected: PASS (or skip in short mode)

- [ ] **Step 6: Commit**

```bash
git add internal/sysops/actions.go internal/sysops/actions_test.go internal/app/SysOps.go internal/app/Types.go
git commit -m "feat(sysops): add RunSystemAction for reboot, shutdown, flush DNS, clear temp, etc."
```

---

### Task 1.6: System Logs

**Files:**
- Create: `internal/sysops/logs.go`
- Modify: `internal/app/SysOps.go`
- Modify: `internal/app/Types.go`

- [ ] **Step 1: Create `logs.go`**

```go
package sysops

import (
	"os/exec"
	"runtime"
	"strings"
)

type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Source    string `json:"source"`
	Message   string `json:"message"`
}

type SystemLogsResult struct {
	Entries []LogEntry `json:"entries"`
	Source  string     `json:"source"`
	Total   int        `json:"total"`
}

func GetSystemLogs(n int, source string) (*SystemLogsResult, error) {
	if n <= 0 {
		n = 50
	}

	var cmd *exec.Cmd
	logSource := source

	switch runtime.GOOS {
	case "windows":
		// Windows Event Log via PowerShell
		logSource = "Windows Event Log"
		cmd = exec.Command("powershell", "-Command",
			"Get-EventLog -LogName System -Newest "+itos(n)+
			" | Select-Object TimeGenerated, EntryType, Source, Message"+
			" | ConvertTo-Json")
	case "linux":
		if source == "dmesg" {
			logSource = "dmesg"
			cmd = exec.Command("dmesg", "--time-format=iso", "-T")
		} else {
			logSource = "journald"
			cmd = exec.Command("journalctl", "-n", itos(n), "--no-pager", "-o", "short-iso")
		}
	default:
		return &SystemLogsResult{Source: "unsupported"}, nil
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	entries := parseLogOutput(string(output), logSource, runtime.GOOS)

	return &SystemLogsResult{
		Entries: entries,
		Source:  logSource,
		Total:   len(entries),
	}, nil
}

func itos(n int) string {
	return strings.TrimSpace(string(rune('0'+n%10))) // Simple int to string for small numbers
}

func parseLogOutput(output, source, goos string) []LogEntry {
	lines := strings.Split(output, "\n")
	var entries []LogEntry

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		entry := LogEntry{
			Source:  source,
			Message: line,
		}

		// Simple heuristic: extract timestamp and level
		if strings.Contains(line, "error") || strings.Contains(line, "Error") || strings.Contains(line, "ERROR") {
			entry.Level = "error"
		} else if strings.Contains(line, "warn") || strings.Contains(line, "Warn") || strings.Contains(line, "WARNING") {
			entry.Level = "warning"
		} else {
			entry.Level = "info"
		}

		entries = append(entries, entry)
	}

	return entries
}
```

- [ ] **Step 2: Add binding in `app/SysOps.go`**

```go
func (s *SysOps) GetSystemLogs(n int, source string) SystemLogsResult {
	result, err := sysops.GetSystemLogs(n, source)
	if err != nil {
		common.LogWarn("GetSystemLogs failed: %v", err)
		return SystemLogsResult{}
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
```

- [ ] **Step 3: Add types in `app/Types.go`**

```go
type SystemLogEntry struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Source    string `json:"source"`
	Message   string `json:"message"`
}

type SystemLogsResultData struct {
	Entries []SystemLogEntry `json:"entries"`
	Source  string           `json:"source"`
	Total   int              `json:"total"`
}
```

- [ ] **Step 4: Write test**

Create `internal/sysops/logs_test.go`:

```go
package sysops

import "testing"

func TestGetSystemLogs(t *testing.T) {
	result, err := GetSystemLogs(10, "journald")
	if err != nil {
		t.Logf("GetSystemLogs returned error (may be expected): %v", err)
		return
	}
	t.Logf("Got %d log entries from %s", result.Total, result.Source)
}
```

- [ ] **Step 5: Run test**

Run: `cd E:\Projects\projectx\AllOpsFull && go test ./internal/sysops/ -run TestGetSystemLogs -v`
Expected: PASS (or graceful error on Windows)

- [ ] **Step 6: Commit**

```bash
git add internal/sysops/logs.go internal/sysops/logs_test.go internal/app/SysOps.go internal/app/Types.go
git commit -m "feat(sysops): add GetSystemLogs for journald, dmesg, Windows Event Log"
```

---

### Task 1.7: Package Manager Detection

**Files:**
- Create: `internal/sysops/packages.go`
- Modify: `internal/app/SysOps.go`
- Modify: `internal/app/Types.go`

- [ ] **Step 1: Create `packages.go`**

```go
package sysops

import (
	"os/exec"
	"runtime"
	"strings"
)

type PackageInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type PackageManagerInfo struct {
	Name     string        `json:"name"`
	Found    bool          `json:"found"`
	Packages []PackageInfo `json:"packages"`
}

func GetInstalledPackages() []PackageManagerInfo {
	var managers []PackageManagerInfo

	switch runtime.GOOS {
	case "linux":
		// Try apt
		aptPkgs := getAptPackages()
		managers = append(managers, aptPkgs)

		// Try dnf
		dnfPkgs := getDnfPackages()
		managers = append(managers, dnfPkgs)

		// Try pacman
		pacmanPkgs := getPacmanPackages()
		managers = append(managers, pacmanPkgs)

	case "windows":
		// Try winget
		wingetPkgs := getWingetPackages()
		managers = append(managers, wingetPkgs)

		// Try choco
		chocoPkgs := getChocoPackages()
		managers = append(managers, chocoPkgs)
	}

	return managers
}

func getAptPackages() PackageManagerInfo {
	result := PackageManagerInfo{Name: "apt", Found: false}
	cmd := exec.Command("dpkg", "--get-selections")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return result
	}
	result.Found = true

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) >= 2 && parts[1] == "install" {
			result.Packages = append(result.Packages, PackageInfo{
				Name: parts[0],
			})
		}
	}
	return result
}

func getDnfPackages() PackageManagerInfo {
	result := PackageManagerInfo{Name: "dnf", Found: false}
	cmd := exec.Command("dnf", "list", "installed", "-q")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return result
	}
	result.Found = true

	lines := strings.Split(string(output), "\n")
	for _, line := range lines[1:] { // skip header
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			result.Packages = append(result.Packages, PackageInfo{
				Name:    parts[0],
				Version: parts[1],
			})
		}
	}
	return result
}

func getPacmanPackages() PackageManagerInfo {
	result := PackageManagerInfo{Name: "pacman", Found: false}
	cmd := exec.Command("pacman", "-Q")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return result
	}
	result.Found = true

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			result.Packages = append(result.Packages, PackageInfo{
				Name:    parts[0],
				Version: parts[1],
			})
		}
	}
	return result
}

func getWingetPackages() PackageManagerInfo {
	result := PackageManagerInfo{Name: "winget", Found: false}
	cmd := exec.Command("winget", "list", "--accept-source-agreements")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return result
	}
	result.Found = true

	lines := strings.Split(string(output), "\n")
	for _, line := range lines[3:] { // skip header lines
		parts := strings.Fields(line)
		if len(parts) >= 3 {
			result.Packages = append(result.Packages, PackageInfo{
				Name:    parts[0],
				Version: parts[1],
			})
		}
	}
	return result
}

func getChocoPackages() PackageManagerInfo {
	result := PackageManagerInfo{Name: "choco", Found: false}
	cmd := exec.Command("choco", "list", "--local-only")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return result
	}
	result.Found = true

	lines := strings.Split(string(output), "\n")
	for _, line := range lines[2:] { // skip header
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			result.Packages = append(result.Packages, PackageInfo{
				Name:    parts[0],
				Version: parts[1],
			})
		}
	}
	return result
}
```

- [ ] **Step 2: Add binding in `app/SysOps.go`**

```go
func (s *SysOps) GetInstalledPackages() []PackageManagerData {
	managers := sysops.GetInstalledPackages()
	var result []PackageManagerData
	for _, m := range managers {
		pkgs := make([]PackageData, 0, len(m.Packages))
		for _, p := range m.Packages {
			pkgs = append(pkgs, PackageData{Name: p.Name, Version: p.Version})
		}
		result = append(result, PackageManagerData{
			Name:     m.Name,
			Found:    m.Found,
			Packages: pkgs,
		})
	}
	return result
}
```

- [ ] **Step 3: Add types in `app/Types.go`**

```go
type PackageData struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type PackageManagerData struct {
	Name     string        `json:"name"`
	Found    bool          `json:"found"`
	Packages []PackageData `json:"packages"`
}
```

- [ ] **Step 4: Write test**

Create `internal/sysops/packages_test.go`:

```go
package sysops

import "testing"

func TestGetInstalledPackages(t *testing.T) {
	managers := GetInstalledPackages()
	if len(managers) == 0 {
		t.Error("Expected at least one package manager")
	}
	for _, m := range managers {
		t.Logf("Package manager %s: found=%v, packages=%d", m.Name, m.Found, len(m.Packages))
	}
}
```

- [ ] **Step 5: Run test**

Run: `cd E:\Projects\projectx\AllOpsFull && go test ./internal/sysops/ -run TestGetInstalledPackages -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add internal/sysops/packages.go internal/sysops/packages_test.go internal/app/SysOps.go internal/app/Types.go
git commit -m "feat(sysops): add GetInstalledPackages for apt, dnf, pacman, winget, choco"
```

---

### Task 1.8: Scheduler (Cron/Tasks)

**Files:**
- Create: `internal/sysops/scheduler.go`
- Modify: `internal/app/SysOps.go`
- Modify: `internal/app/Types.go`

- [ ] **Step 1: Create `scheduler.go`**

```go
package sysops

import (
	"os/exec"
	"runtime"
	"strings"
)

type ScheduledTaskInfo struct {
	Name     string `json:"name"`
	Schedule string `json:"schedule"`
	Command  string `json:"command"`
	Enabled  bool   `json:"enabled"`
	NextRun  string `json:"next_run"`
}

func GetScheduledTasks() ([]ScheduledTaskInfo, error) {
	switch runtime.GOOS {
	case "linux":
		return getCronTasks()
	case "windows":
		return getSchtasks()
	default:
		return nil, nil
	}
}

func getCronTasks() ([]ScheduledTaskInfo, error) {
	cmd := exec.Command("crontab", "-l")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err // likely no crontab
	}

	var tasks []ScheduledTaskInfo
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 6 {
			schedule := strings.Join(parts[:5], " ")
			tasks = append(tasks, ScheduledTaskInfo{
				Name:     parts[5],
				Schedule: schedule,
				Command:  strings.Join(parts[5:], " "),
				Enabled:  true,
			})
		}
	}
	return tasks, nil
}

func getSchtasks() ([]ScheduledTaskInfo, error) {
	cmd := exec.Command("schtasks", "/query", "/fo", "LIST", "/v")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	var tasks []ScheduledTaskInfo
	// Parse schtasks output (simplified)
	blocks := strings.Split(string(output), "\n\n")
	for _, block := range blocks {
		if block == "" {
			continue
		}
		task := ScheduledTaskInfo{}
		for _, line := range strings.Split(block, "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "TaskName:") {
				task.Name = strings.TrimPrefix(line, "TaskName: ")
			} else if strings.HasPrefix(line, "Next Run Time:") {
				task.NextRun = strings.TrimPrefix(line, "Next Run Time: ")
			} else if strings.HasPrefix(line, "Status:") {
				status := strings.TrimPrefix(line, "Status: ")
				task.Enabled = status == "Ready" || status == "Running"
			}
		}
		if task.Name != "" {
			tasks = append(tasks, task)
		}
	}
	return tasks, nil
}
```

- [ ] **Step 2: Add binding in `app/SysOps.go`**

```go
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
```

- [ ] **Step 3: Add types in `app/Types.go`**

```go
type ScheduledTaskData struct {
	Name     string `json:"name"`
	Schedule string `json:"schedule"`
	Command  string `json:"command"`
	Enabled  bool   `json:"enabled"`
	NextRun  string `json:"next_run"`
}
```

- [ ] **Step 4: Write test**

Create `internal/sysops/scheduler_test.go`:

```go
package sysops

import "testing"

func TestGetScheduledTasks(t *testing.T) {
	tasks, err := GetScheduledTasks()
	if err != nil {
		t.Logf("GetScheduledTasks returned error (may be expected): %v", err)
		return
	}
	t.Logf("Found %d scheduled tasks", len(tasks))
}
```

- [ ] **Step 5: Run test**

Run: `cd E:\Projects\projectx\AllOpsFull && go test ./internal/sysops/ -run TestGetScheduledTasks -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add internal/sysops/scheduler.go internal/sysops/scheduler_test.go internal/app/SysOps.go internal/app/Types.go
git commit -m "feat(sysops): add GetScheduledTasks for crontab and Windows schtasks"
```

---

### Task 1.9: Extended Diagnostics

**Files:**
- Modify: `internal/sysops/workflows.go`
- Modify: `internal/app/SysOps.go`
- Modify: `internal/app/Types.go`

- [ ] **Step 1: Add `ExtendedDiagnostics` to `workflows.go`**

```go
type DiagnosticCheck struct {
	Name    string `json:"name"`
	Status  string `json:"status"` // "pass", "warn", "fail"
	Message string `json:"message"`
	Value   string `json:"value"`
}

type DiagnosticResult struct {
	Checks    []DiagnosticCheck `json:"checks"`
	Score     int               `json:"score"`
	Timestamp string            `json:"timestamp"`
}

func RunExtendedDiagnostics() (*DiagnosticResult, error) {
	result := &DiagnosticResult{}
	score := 100

	// CPU check
	cpu, err := GetCPUStats()
	if err == nil {
		check := DiagnosticCheck{Name: "CPU Usage", Value: fmt.Sprintf("%.1f%%", cpu.Percent)}
		if cpu.Percent > 90 {
			check.Status = "fail"
			check.Message = "CPU usage critical"
			score -= 30
		} else if cpu.Percent > 80 {
			check.Status = "warn"
			check.Message = "CPU usage elevated"
			score -= 15
		} else {
			check.Status = "pass"
			check.Message = "CPU usage normal"
		}
		result.Checks = append(result.Checks, check)
	}

	// Memory check
	mem, err := GetMemoryStats()
	if err == nil {
		check := DiagnosticCheck{Name: "Memory Usage", Value: fmt.Sprintf("%.1f%%", mem.UsedPercent)}
		if mem.UsedPercent > 90 {
			check.Status = "fail"
			check.Message = "Memory usage critical"
			score -= 25
		} else if mem.UsedPercent > 80 {
			check.Status = "warn"
			check.Message = "Memory usage elevated"
			score -= 10
		} else {
			check.Status = "pass"
			check.Message = "Memory usage normal"
		}
		result.Checks = append(result.Checks, check)
	}

	// Disk check
	disk, err := GetDiskStats()
	if err == nil {
		for _, p := range disk.Usage {
			check := DiagnosticCheck{
				Name:  fmt.Sprintf("Disk %s", p.Mountpoint),
				Value: fmt.Sprintf("%.1f%%", p.UsedPercent),
			}
			if p.UsedPercent > 95 {
				check.Status = "fail"
				check.Message = "Disk space critical"
				score -= 20
			} else if p.UsedPercent > 85 {
				check.Status = "warn"
				check.Message = "Disk space low"
				score -= 10
			} else {
				check.Status = "pass"
				check.Message = "Disk space OK"
			}
			result.Checks = append(result.Checks, check)
		}
	}

	// Swap check
	if mem != nil && mem.SwapTotal > 0 {
		check := DiagnosticCheck{
			Name:  "Swap Usage",
			Value: fmt.Sprintf("%.1f%%", mem.SwapPercent),
		}
		if mem.SwapPercent > 80 {
			check.Status = "warn"
			check.Message = "Swap usage high"
			score -= 10
		} else {
			check.Status = "pass"
			check.Message = "Swap usage normal"
		}
		result.Checks = append(result.Checks, check)
	}

	// Temperature check
	ext, err := GetCPUExtended()
	if err == nil && ext.Temperature > 0 {
		check := DiagnosticCheck{
			Name:  "CPU Temperature",
			Value: fmt.Sprintf("%.1f°C", ext.Temperature),
		}
		if ext.Temperature > 85 {
			check.Status = "fail"
			check.Message = "Temperature critical"
			score -= 20
		} else if ext.Temperature > 70 {
			check.Status = "warn"
			check.Message = "Temperature elevated"
			score -= 5
		} else {
			check.Status = "pass"
			check.Message = "Temperature normal"
		}
		result.Checks = append(result.Checks, check)
	}

	if score < 0 {
		score = 0
	}
	result.Score = score

	return result, nil
}
```

- [ ] **Step 2: Add binding in `app/SysOps.go`**

```go
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
		Checks: checks,
		Score:  result.Score,
	}
}
```

- [ ] **Step 3: Add types in `app/Types.go`**

```go
type DiagnosticCheckData struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Message string `json:"message"`
	Value   string `json:"value"`
}

type ExtendedDiagnosticResult struct {
	Checks []DiagnosticCheckData `json:"checks"`
	Score  int                   `json:"score"`
}
```

- [ ] **Step 4: Write test**

Create `internal/sysops/diagnostics_test.go`:

```go
package sysops

import "testing"

func TestRunExtendedDiagnostics(t *testing.T) {
	result, err := RunExtendedDiagnostics()
	if err != nil {
		t.Fatalf("RunExtendedDiagnostics returned error: %v", err)
	}
	if len(result.Checks) == 0 {
		t.Error("Expected at least one diagnostic check")
	}
	if result.Score < 0 || result.Score > 100 {
		t.Errorf("Score should be 0-100, got %d", result.Score)
	}
	t.Logf("Diagnostic score: %d, checks: %d", result.Score, len(result.Checks))
}
```

- [ ] **Step 5: Run test**

Run: `cd E:\Projects\projectx\AllOpsFull && go test ./internal/sysops/ -run TestRunExtendedDiagnostics -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add internal/sysops/workflows.go internal/sysops/diagnostics_test.go internal/app/SysOps.go internal/app/Types.go
git commit -m "feat(sysops): add RunExtendedDiagnostics with CPU, memory, disk, swap, temperature checks"
```

---

### Task 1.10: Bind All New Methods in main.go

**Files:**
- Verify: `main.go` (Wails binding)

- [ ] **Step 1: Verify all new methods are bound**

Check that `main.go` binds the `SysOps` struct which already includes all new methods. Since methods are added to the `SysOps` struct in `app/SysOps.go`, they should automatically be available via Wails binding.

Run: `cd E:\Projects\projectx\AllOpsFull && go build ./...`
Expected: Build succeeds with no errors

- [ ] **Step 2: Commit if any changes needed**

```bash
git add main.go
git commit -m "chore: verify Wails bindings for new SysOps methods"
```

---

## Phase 2: Frontend Types

### Task 2.1: Add New TypeScript Types

**Files:**
- Modify: `cmd/opsforall-gui/frontend/src/types/index.ts`

- [ ] **Step 1: Add new SysOps types**

```typescript
// ── Extended SysOps Types ──

export interface PerCPUInfo {
  core: number
  frequency_mhz: number
  usage_percent: number
}

export interface CPUExtendedInfo {
  model_name: string
  frequency_mhz: number
  cache_size_kb: number
  temperature: number
  per_cpu_info: PerCPUInfo[]
}

export interface DiskIOEntry {
  name: string
  read_bytes: number
  write_bytes: number
  read_count: number
  write_count: number
}

export interface DiskIOData {
  disks: DiskIOEntry[]
  total_read_bytes: number
  total_write_bytes: number
}

export interface LoggedInUserData {
  user: string
  terminal: string
  host: string
  started: string
}

export interface CPUTimesData {
  user: number
  system: number
  idle: number
  iowait: number
  steal: number
  total: number
}

export interface LoadAverageData {
  load_1: number
  load_5: number
  load_15: number
}

export interface PerformanceData {
  cpu_times: CPUTimesData
  load_average: LoadAverageData
  io_wait: number
}

export interface ActionResult {
  action: string
  success: boolean
  message: string
  output: string
}

export interface SystemLogEntry {
  timestamp: string
  level: string
  source: string
  message: string
}

export interface SystemLogsResult {
  entries: SystemLogEntry[]
  source: string
  total: number
}

export interface PackageData {
  name: string
  version: string
}

export interface PackageManagerData {
  name: string
  found: boolean
  packages: PackageData[]
}

export interface ScheduledTaskData {
  name: string
  schedule: string
  command: string
  enabled: boolean
  next_run: string
}

export interface DiagnosticCheckData {
  name: string
  status: 'pass' | 'warn' | 'fail'
  message: string
  value: string
}

export interface ExtendedDiagnosticResult {
  checks: DiagnosticCheckData[]
  score: number
}
```

- [ ] **Step 2: Verify TypeScript compiles**

Run: `cd E:\Projects\projectx\AllOpsFull\cmd\opsforall-gui\frontend && npx tsc --noEmit`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
git add cmd/opsforall-gui/frontend/src/types/index.ts
git commit -m "feat(types): add SysOps extended types for CPU, disk IO, performance, logs, packages, diagnostics"
```

---

## Phase 3: Frontend Restructure

### Task 3.1: Create SysOps Category Components Structure

**Files:**
- Create: `cmd/opsforall-gui/frontend/src/pages/SysOps/SystemInfoTab.tsx`
- Create: `cmd/opsforall-gui/frontend/src/pages/SysOps/CpuTab.tsx`
- Create: `cmd/opsforall-gui/frontend/src/pages/SysOps/MemoryTab.tsx`
- Create: `cmd/opsforall-gui/frontend/src/pages/SysOps/DiskTab.tsx`
- Create: `cmd/opsforall-gui/frontend/src/pages/SysOps/ProcessesTab.tsx`
- Create: `cmd/opsforall-gui/frontend/src/pages/SysOps/ServicesTab.tsx`
- Create: `cmd/opsforall-gui/frontend/src/pages/SysOps/LogsTab.tsx`
- Create: `cmd/opsforall-gui/frontend/src/pages/SysOps/StorageTab.tsx`
- Create: `cmd/opsforall-gui/frontend/src/pages/SysOps/UsersTab.tsx`
- Create: `cmd/opsforall-gui/frontend/src/pages/SysOps/PerformanceTab.tsx`
- Create: `cmd/opsforall-gui/frontend/src/pages/SysOps/PackageManagerTab.tsx`
- Create: `cmd/opsforall-gui/frontend/src/pages/SysOps/SchedulerTab.tsx`
- Create: `cmd/opsforall-gui/frontend/src/pages/SysOps/DiagnosticsTab.tsx`
- Create: `cmd/opsforall-gui/frontend/src/pages/SysOps/ActionsTab.tsx`
- Create: `cmd/opsforall-gui/frontend/src/components/sysops/HealthBadge.tsx`
- Create: `cmd/opsforall-gui/frontend/src/components/sysops/CommandOutput.tsx`

- [ ] **Step 1: Create directory structure**

Run: `mkdir -p cmd/opsforall-gui/frontend/src/pages/SysOps`

- [ ] **Step 2: Create `HealthBadge.tsx`**

```tsx
import { cn } from '@/lib/utils'
import { CheckCircle, AlertTriangle, XCircle } from 'lucide-react'

interface HealthBadgeProps {
  status: 'pass' | 'warn' | 'fail'
  value: string
  label: string
}

const statusConfig = {
  pass: { icon: CheckCircle, color: 'text-success', bg: 'bg-success/20', border: 'border-success/30' },
  warn: { icon: AlertTriangle, color: 'text-warning', bg: 'bg-warning/20', border: 'border-warning/30' },
  fail: { icon: XCircle, color: 'text-danger', bg: 'bg-danger/20', border: 'border-danger/30' },
}

export function HealthBadge({ status, value, label }: HealthBadgeProps) {
  const config = statusConfig[status]
  const Icon = config.icon

  return (
    <div className={cn('flex items-center gap-3 px-4 py-2 rounded-xl border', config.bg, config.border)}>
      <Icon size={16} className={config.color} />
      <div className="flex flex-col">
        <span className={cn('text-sm font-bold tabular-nums', config.color)}>{value}</span>
        <span className="text-[10px] font-bold text-text-faint uppercase tracking-widest">{label}</span>
      </div>
    </div>
  )
}
```

- [ ] **Step 3: Create `CommandOutput.tsx`**

```tsx
import { Terminal } from 'lucide-react'

interface CommandOutputProps {
  output: string
  title?: string
}

export function CommandOutput({ output, title }: CommandOutputProps) {
  return (
    <div className="bg-panel-3 border border-border rounded-xl overflow-hidden">
      {title && (
        <div className="flex items-center gap-2 px-4 py-2 border-b border-border bg-panel-2">
          <Terminal size={14} className="text-accent" />
          <span className="text-xs font-bold text-text-faint uppercase">{title}</span>
        </div>
      )}
      <pre className="p-4 text-sm font-mono text-text-dim overflow-x-auto whitespace-pre-wrap">
        {output || 'No output'}
      </pre>
    </div>
  )
}
```

- [ ] **Step 4: Commit**

```bash
git add cmd/opsforall-gui/frontend/src/pages/SysOps/ cmd/opsforall-gui/frontend/src/components/sysops/
git commit -m "feat(sysops): create category component structure and shared components"
```

---

### Task 3.2: Implement SystemInfoTab

**Files:**
- Create: `cmd/opsforall-gui/frontend/src/pages/SysOps/SystemInfoTab.tsx`

- [ ] **Step 1: Create `SystemInfoTab.tsx`**

```tsx
import { Server, Cpu, Clock, Users, Copy, Globe } from 'lucide-react'
import { useQuery } from '@tanstack/react-query'
import { useBackend } from '@/hooks/useBackend'
import { useSettingsStore } from '@/stores/useSettingsStore'
import type { SystemInfo, CPUInfo, CPUExtendedInfo, LoggedInUserData } from '@/types'

interface SystemInfoTabProps {
  sysInfo: SystemInfo
  cpuInfo: CPUInfo
}

export function SystemInfoTab({ sysInfo, cpuInfo }: SystemInfoTabProps) {
  const { call } = useBackend()
  const { refreshInterval } = useSettingsStore()

  const { data: cpuExtended } = useQuery<CPUExtendedInfo>({
    queryKey: ['sysops-cpu-extended'],
    queryFn: async () => { const r = await call('SysOps.GetCPUExtended'); return r as CPUExtendedInfo },
    refetchInterval: refreshInterval,
  })

  const { data: users = [] } = useQuery<LoggedInUserData[]>({
    queryKey: ['sysops-users'],
    queryFn: async () => { const r = await call('SysOps.GetLoggedInUsers'); return (r as LoggedInUserData[]) || [] },
    refetchInterval: refreshInterval,
  })

  const copyHostname = () => navigator.clipboard.writeText(sysInfo.hostname)

  return (
    <div className="space-y-8">
      {/* OS Information */}
      <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-8 shadow-2xl">
        <div className="flex items-center gap-3 mb-6">
          <Globe size={20} className="text-accent" />
          <h3 className="text-lg font-bold text-text uppercase tracking-widest">OS Information</h3>
        </div>
        <div className="grid grid-cols-2 gap-6">
          <InfoRow label="Hostname" value={sysInfo.hostname} action={<button onClick={copyHostname} className="p-1.5 hover:bg-panel-3 rounded-lg"><Copy size={14} className="text-text-faint" /></button>} />
          <InfoRow label="Platform" value={sysInfo.platform} />
          <InfoRow label="Kernel Version" value={sysInfo.kernel_version} />
          <InfoRow label="Architecture" value={sysInfo.kernel_arch} />
          <InfoRow label="Uptime" value={sysInfo.uptime} />
          <InfoRow label="Virtualization" value={sysInfo.virtualization || 'None'} />
        </div>
      </div>

      {/* Hardware Summary */}
      <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-8 shadow-2xl">
        <div className="flex items-center gap-3 mb-6">
          <Cpu size={20} className="text-accent" />
          <h3 className="text-lg font-bold text-text uppercase tracking-widest">Hardware Summary</h3>
        </div>
        <div className="grid grid-cols-2 gap-6">
          <InfoRow label="CPU Model" value={cpuExtended?.model_name || 'N/A'} />
          <InfoRow label="CPU Frequency" value={cpuExtended ? `${cpuExtended.frequency_mhz.toFixed(0)} MHz` : 'N/A'} />
          <InfoRow label="Physical Cores" value={cpuInfo.physical_cores} />
          <InfoRow label="Logical Cores" value={cpuInfo.logical_cores} />
          <InfoRow label="Cache Size" value={cpuExtended ? `${cpuExtended.cache_size_kb} KB` : 'N/A'} />
          <InfoRow label="Processes" value={sysInfo.process_count} />
        </div>
      </div>

      {/* Logged-in Users */}
      <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-8 shadow-2xl">
        <div className="flex items-center gap-3 mb-6">
          <Users size={20} className="text-success" />
          <h3 className="text-lg font-bold text-text uppercase tracking-widest">Logged-in Users</h3>
        </div>
        {users.length === 0 ? (
          <p className="text-text-dim text-sm">No users detected</p>
        ) : (
          <div className="space-y-3">
            {users.map((u, i) => (
              <div key={i} className="flex items-center justify-between py-2 border-b border-border last:border-0">
                <span className="text-sm font-bold text-text">{u.user}</span>
                <span className="text-xs text-text-faint">{u.terminal} from {u.host || 'local'}</span>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}

function InfoRow({ label, value, action }: { label: string; value: string | number; action?: React.ReactNode }) {
  return (
    <div className="flex items-center justify-between py-2 border-b border-border/50 last:border-0">
      <span className="text-xs font-bold text-text-faint uppercase tracking-wider">{label}</span>
      <div className="flex items-center gap-2">
        <span className="text-sm font-medium text-text">{String(value)}</span>
        {action}
      </div>
    </div>
  )
}
```

- [ ] **Step 2: Verify TypeScript compiles**

Run: `cd E:\Projects\projectx\AllOpsFull\cmd\opsforall-gui\frontend && npx tsc --noEmit`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
git add cmd/opsforall-gui/frontend/src/pages/SysOps/SystemInfoTab.tsx
git commit -m "feat(sysops): implement SystemInfoTab with OS, hardware, logged-in users"
```

---

### Task 3.3: Implement CpuTab

**Files:**
- Create: `cmd/opsforall-gui/frontend/src/pages/SysOps/CpuTab.tsx`

- [ ] **Step 1: Create `CpuTab.tsx`**

```tsx
import { Cpu, Thermometer, Activity } from 'lucide-react'
import { useQuery } from '@tanstack/react-query'
import { useBackend } from '@/hooks/useBackend'
import { useSettingsStore } from '@/stores/useSettingsStore'
import type { CPUInfo, CPUExtendedInfo, PerformanceData } from '@/types'
import { Bar } from '../SysOps' // Will need to export Bar from main SysOps

interface CpuTabProps {
  cpuInfo: CPUInfo
}

export function CpuTab({ cpuInfo }: CpuTabProps) {
  const { call } = useBackend()
  const { refreshInterval } = useSettingsStore()

  const { data: cpuExtended } = useQuery<CPUExtendedInfo>({
    queryKey: ['sysops-cpu-extended'],
    queryFn: async () => { const r = await call('SysOps.GetCPUExtended'); return r as CPUExtendedInfo },
    refetchInterval: refreshInterval,
  })

  const { data: perfData } = useQuery<PerformanceData>({
    queryKey: ['sysops-performance'],
    queryFn: async () => { const r = await call('SysOps.GetPerformanceStats'); return r as PerformanceData },
    refetchInterval: refreshInterval,
  })

  const saturation = (cpuInfo.load_avg_1 / cpuInfo.logical_cores) * 100

  return (
    <div className="space-y-8">
      {/* CPU Overview */}
      <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-8 shadow-2xl">
        <div className="flex items-center justify-between mb-6">
          <div className="flex items-center gap-3">
            <Cpu size={20} className="text-accent" />
            <h3 className="text-lg font-bold text-text uppercase tracking-widest">Processor Health</h3>
          </div>
          <div className="flex items-center gap-4">
            <span className="text-3xl font-bold text-text tabular-nums">{cpuInfo.percent.toFixed(1)}%</span>
            <span className={`text-xs font-bold px-2 py-0.5 rounded border ${saturation > 80 ? 'bg-danger/20 text-danger border-danger/30' : 'bg-success/20 text-success border-success/30'}`}>
              {saturation.toFixed(0)}% Saturation
            </span>
          </div>
        </div>
        <div className="grid grid-cols-3 gap-6 mb-6">
          <div className="text-center">
            <p className="text-2xl font-bold text-text tabular-nums">{cpuInfo.physical_cores}</p>
            <p className="text-xs font-bold text-text-faint uppercase">Physical Cores</p>
          </div>
          <div className="text-center">
            <p className="text-2xl font-bold text-text tabular-nums">{cpuInfo.logical_cores}</p>
            <p className="text-xs font-bold text-text-faint uppercase">Logical Cores</p>
          </div>
          <div className="text-center">
            <p className="text-2xl font-bold text-text tabular-nums">{cpuExtended?.temperature ? `${cpuExtended.temperature.toFixed(1)}°C` : 'N/A'}</p>
            <p className="text-xs font-bold text-text-faint uppercase">Temperature</p>
          </div>
        </div>
        <div className="grid grid-cols-2 gap-x-8 gap-y-3">
          {cpuInfo.per_cpu.map((p, i) => (
            <Bar key={i} label={`Core ${i}`} value={p} />
          ))}
        </div>
      </div>

      {/* Load Average */}
      {perfData && (
        <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-8 shadow-2xl">
          <div className="flex items-center gap-3 mb-6">
            <Activity size={20} className="text-accent" />
            <h3 className="text-lg font-bold text-text uppercase tracking-widest">Load Average</h3>
          </div>
          <div className="grid grid-cols-3 gap-6">
            <div className="text-center">
              <p className="text-3xl font-bold text-text tabular-nums">{perfData.load_average.load_1.toFixed(2)}</p>
              <p className="text-xs font-bold text-text-faint uppercase">1 Minute</p>
            </div>
            <div className="text-center">
              <p className="text-3xl font-bold text-text tabular-nums">{perfData.load_average.load_5.toFixed(2)}</p>
              <p className="text-xs font-bold text-text-faint uppercase">5 Minutes</p>
            </div>
            <div className="text-center">
              <p className="text-3xl font-bold text-text tabular-nums">{perfData.load_average.load_15.toFixed(2)}</p>
              <p className="text-xs font-bold text-text-faint uppercase">15 Minutes</p>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
```

- [ ] **Step 2: Verify TypeScript compiles**

Run: `cd E:\Projects\projectx\AllOpsFull\cmd\opsforall-gui\frontend && npx tsc --noEmit`
Expected: No errors (may need to export Bar from SysOps.tsx)

- [ ] **Step 3: Commit**

```bash
git add cmd/opsforall-gui/frontend/src/pages/SysOps/CpuTab.tsx
git commit -m "feat(sysops): implement CpuTab with per-core bars, temperature, load average"
```

---

### Task 3.4: Implement MemoryTab, DiskTab, ProcessesTab, DiagnosticsTab

**Files:**
- Create: `cmd/opsforall-gui/frontend/src/pages/SysOps/MemoryTab.tsx`
- Create: `cmd/opsforall-gui/frontend/src/pages/SysOps/DiskTab.tsx`
- Create: `cmd/opsforall-gui/frontend/src/pages/SysOps/ProcessesTab.tsx`
- Create: `cmd/opsforall-gui/frontend/src/pages/SysOps/DiagnosticsTab.tsx`

- [ ] **Step 1: Create `MemoryTab.tsx`**

```tsx
import { MemoryStick } from 'lucide-react'
import type { MemoryInfo } from '@/types'

interface MemoryTabProps {
  memInfo: MemoryInfo
}

export function MemoryTab({ memInfo }: MemoryTabProps) {
  const availableGB = memInfo.total_gb - memInfo.used_gb
  const cachedGB = memInfo.cached_bytes / (1024 * 1024 * 1024)
  const swapUsedGB = memInfo.swap_used / (1024 * 1024 * 1024)
  const swapTotalGB = memInfo.swap_total / (1024 * 1024 * 1024)

  return (
    <div className="space-y-8">
      <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-8 shadow-2xl">
        <div className="flex items-center justify-between mb-6">
          <div className="flex items-center gap-3">
            <MemoryStick size={20} className="text-success" />
            <h3 className="text-lg font-bold text-text uppercase tracking-widest">RAM Usage</h3>
          </div>
          <span className="text-3xl font-bold text-success tabular-nums">{memInfo.used_percent.toFixed(1)}%</span>
        </div>
        <div className="h-6 bg-panel-3 rounded-full overflow-hidden border border-border mb-6">
          <div className="h-full rounded-full bg-gradient-to-r from-success/60 to-success transition-all duration-700" style={{ width: `${memInfo.used_percent}%` }} />
        </div>
        <div className="grid grid-cols-4 gap-6">
          <StatBox label="Total" value={`${memInfo.total_gb.toFixed(1)} GB`} />
          <StatBox label="Used" value={`${memInfo.used_gb.toFixed(1)} GB`} color="text-success" />
          <StatBox label="Available" value={`${availableGB.toFixed(1)} GB`} />
          <StatBox label="Cached" value={`${cachedGB.toFixed(1)} GB`} />
        </div>
      </div>

      {memInfo.swap_total > 0 && (
        <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-8 shadow-2xl">
          <div className="flex items-center justify-between mb-6">
            <h3 className="text-lg font-bold text-text uppercase tracking-widest">Swap Usage</h3>
            <span className="text-2xl font-bold text-warning tabular-nums">{memInfo.swap_percent.toFixed(1)}%</span>
          </div>
          <div className="h-4 bg-panel-3 rounded-full overflow-hidden border border-border">
            <div className="h-full rounded-full bg-gradient-to-r from-warning/60 to-warning transition-all duration-700" style={{ width: `${memInfo.swap_percent}%` }} />
          </div>
          <div className="flex justify-between mt-3 text-sm">
            <span className="text-text-dim">{swapUsedGB.toFixed(1)} GB used</span>
            <span className="text-text-faint">{swapTotalGB.toFixed(1)} GB total</span>
          </div>
        </div>
      )}
    </div>
  )
}

function StatBox({ label, value, color = 'text-text' }: { label: string; value: string; color?: string }) {
  return (
    <div className="text-center">
      <p className={`text-xl font-bold tabular-nums ${color}`}>{value}</p>
      <p className="text-xs font-bold text-text-faint uppercase">{label}</p>
    </div>
  )
}
```

- [ ] **Step 2: Create `DiskTab.tsx`**

```tsx
import { Disc, HardDrive } from 'lucide-react'
import { useQuery } from '@tanstack/react-query'
import { useBackend } from '@/hooks/useBackend'
import { useSettingsStore } from '@/stores/useSettingsStore'
import type { DiskInfo, DiskIOData } from '@/types'

export function DiskTab({ diskInfo }: { diskInfo: DiskInfo }) {
  const { call } = useBackend()
  const { refreshInterval } = useSettingsStore()

  const { data: diskIO } = useQuery<DiskIOData>({
    queryKey: ['sysops-disk-io'],
    queryFn: async () => { const r = await call('SysOps.GetDiskIO'); return r as DiskIOData },
    refetchInterval: refreshInterval,
  })

  return (
    <div className="space-y-8">
      {/* Partition Usage */}
      <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-8 shadow-2xl">
        <div className="flex items-center gap-3 mb-6">
          <Disc size={20} className="text-accent" />
          <h3 className="text-lg font-bold text-text uppercase tracking-widest">Disk Usage</h3>
        </div>
        <div className="space-y-6">
          {diskInfo.partitions.map((p, i) => (
            <div key={i}>
              <div className="flex items-center justify-between mb-2">
                <span className="text-sm font-bold text-text">{p.mountpoint}</span>
                <span className="text-sm font-bold text-text tabular-nums">{p.used_percent.toFixed(1)}%</span>
              </div>
              <div className="h-4 bg-panel-3 rounded-full overflow-hidden border border-border">
                <div className="h-full rounded-full bg-gradient-to-r from-accent/60 to-accent transition-all duration-700" style={{ width: `${p.used_percent}%` }} />
              </div>
              <div className="flex justify-between mt-1 text-xs text-text-faint">
                <span>{(p.total_bytes / 1e9).toFixed(1)} GB total · {(p.free_bytes / 1e9).toFixed(1)} GB free</span>
                <span>{p.fs_type} · {p.device}</span>
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* Disk I/O */}
      {diskIO && (
        <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-8 shadow-2xl">
          <div className="flex items-center gap-3 mb-6">
            <HardDrive size={20} className="text-warning" />
            <h3 className="text-lg font-bold text-text uppercase tracking-widest">Disk I/O</h3>
          </div>
          <div className="grid grid-cols-2 gap-6 mb-6">
            <div className="text-center">
              <p className="text-2xl font-bold text-text tabular-nums">{(diskIO.total_read_bytes / 1e9).toFixed(2)} GB</p>
              <p className="text-xs font-bold text-text-faint uppercase">Total Read</p>
            </div>
            <div className="text-center">
              <p className="text-2xl font-bold text-text tabular-nums">{(diskIO.total_write_bytes / 1e9).toFixed(2)} GB</p>
              <p className="text-xs font-bold text-text-faint uppercase">Total Write</p>
            </div>
          </div>
          <div className="space-y-3">
            {diskIO.disks.map((d, i) => (
              <div key={i} className="flex items-center justify-between py-2 border-b border-border/50 last:border-0">
                <span className="text-sm font-bold text-text">{d.name}</span>
                <div className="flex gap-6 text-xs">
                  <span className="text-accent tabular-nums">R: {(d.read_bytes / 1e6).toFixed(1)} MB</span>
                  <span className="text-warning tabular-nums">W: {(d.write_bytes / 1e6).toFixed(1)} MB</span>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}
```

- [ ] **Step 3: Create `ProcessesTab.tsx` (migrate existing process management)**

```tsx
import { useState } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { Search, Trash2, TreePine, List } from 'lucide-react'
import { useBackend } from '@/hooks/useBackend'
import { useSettingsStore } from '@/stores/useSettingsStore'
import { ConfirmDialog } from '@/components/dialogs/ConfirmDialog'
import type { ProcessInfo } from '@/types'
import { cn } from '@/lib/utils'

export function ProcessesTab() {
  const { call } = useBackend()
  const { refreshInterval } = useSettingsStore()
  const queryClient = useQueryClient()
  const [search, setSearch] = useState('')
  const [view, setView] = useState<'list' | 'tree'>('list')
  const [killTarget, setKillTarget] = useState<{ pid: number; name: string } | null>(null)

  const { data: processes = [] } = useQuery<ProcessInfo[]>({
    queryKey: ['sysops-processes'],
    queryFn: async () => { const r = await call('SysOps.ListAllProcesses', 100); return (r as ProcessInfo[]) || [] },
    refetchInterval: refreshInterval,
  })

  const { data: processTree = [] } = useQuery<ProcessInfo[]>({
    queryKey: ['sysops-process-tree'],
    queryFn: async () => { const r = await call('SysOps.GetProcessTree'); return (r as ProcessInfo[]) || [] },
    refetchInterval: refreshInterval,
  })

  const killProcess = async (pid: number) => {
    await call('DevOps.KillProcess', pid)
    queryClient.invalidateQueries({ queryKey: ['sysops-processes'] })
    setKillTarget(null)
  }

  const filtered = processes.filter(p => p.name.toLowerCase().includes(search.toLowerCase()))

  return (
    <div className="space-y-6">
      <ConfirmDialog
        open={killTarget !== null}
        title="Kill Process"
        description={`Terminate "${killTarget?.name}" (PID: ${killTarget?.pid})?`}
        type="danger"
        confirmText="Kill"
        onConfirm={() => killProcess(killTarget!.pid)}
        onClose={() => setKillTarget(null)}
      />

      <div className="flex items-center gap-4">
        <div className="relative flex-1">
          <Search size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-text-faint" />
          <input
            type="text"
            placeholder="Filter processes..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full bg-panel border border-border rounded-xl pl-10 pr-4 py-2 text-sm text-text placeholder-text-faint focus:outline-none focus:border-accent"
          />
        </div>
        <div className="flex gap-1 bg-panel border border-border rounded-lg p-1">
          <button onClick={() => setView('list')} className={cn("px-3 py-1.5 rounded text-xs font-bold", view === 'list' ? "bg-accent text-white" : "text-text-faint")}>
            <List size={14} />
          </button>
          <button onClick={() => setView('tree')} className={cn("px-3 py-1.5 rounded text-xs font-bold", view === 'tree' ? "bg-accent text-white" : "text-text-faint")}>
            <TreePine size={14} />
          </button>
        </div>
        <span className="text-sm text-text-faint">{filtered.length} active</span>
      </div>

      <div className="bg-panel border border-border rounded-xl overflow-hidden">
        <div className="max-h-[600px] overflow-y-auto">
          <table className="w-full text-left">
            <thead className="sticky top-0 bg-panel-2 border-b border-border">
              <tr>
                <th className="px-4 py-3 text-xs font-bold text-text-faint uppercase">Process</th>
                <th className="px-4 py-3 text-xs font-bold text-text-faint uppercase text-right">CPU %</th>
                <th className="px-4 py-3 text-xs font-bold text-text-faint uppercase text-right">RAM (MB)</th>
                <th className="px-4 py-3 text-xs font-bold text-text-faint uppercase text-right">FDs</th>
                <th className="px-4 py-3 w-12" />
              </tr>
            </thead>
            <tbody>
              {filtered.map(p => (
                <tr key={p.pid} className="border-b border-border/20 hover:bg-sidebar-hover group">
                  <td className="px-4 py-3">
                    <span className="text-sm font-medium text-text">{p.name}</span>
                    <span className="text-xs text-text-faint ml-2">PID {p.pid}</span>
                  </td>
                  <td className="px-4 py-3 text-right text-sm font-bold text-accent tabular-nums">{p.cpu.toFixed(1)}%</td>
                  <td className="px-4 py-3 text-right text-sm text-text-dim tabular-nums">{p.memory.toFixed(0)}</td>
                  <td className="px-4 py-3 text-right text-sm text-text-faint tabular-nums">{p.num_fds}</td>
                  <td className="px-4 py-3">
                    <button onClick={() => setKillTarget({ pid: p.pid, name: p.name })} className="opacity-0 group-hover:opacity-100 p-1.5 text-text-faint hover:text-danger hover:bg-danger/10 rounded-lg transition-all">
                      <Trash2 size={14} />
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}
```

- [ ] **Step 4: Create `DiagnosticsTab.tsx`**

```tsx
import { useQuery } from '@tanstack/react-query'
import { Activity, CheckCircle, AlertTriangle, XCircle } from 'lucide-react'
import { useBackend } from '@/hooks/useBackend'
import { useSettingsStore } from '@/stores/useSettingsStore'
import type { ExtendedDiagnosticResult } from '@/types'
import { cn } from '@/lib/utils'

export function DiagnosticsTab() {
  const { call } = useBackend()
  const { refreshInterval } = useSettingsStore()

  const { data: diagnostics, isLoading } = useQuery<ExtendedDiagnosticResult>({
    queryKey: ['sysops-diagnostics'],
    queryFn: async () => { const r = await call('SysOps.RunExtendedDiagnostics'); return r as ExtendedDiagnosticResult },
    refetchInterval: refreshInterval,
  })

  if (isLoading) {
    return <div className="animate-pulse space-y-4">{[1, 2, 3].map(i => <div key={i} className="h-16 bg-panel-2 rounded-xl" />)}</div>
  }

  const scoreColor = (diagnostics?.score || 0) >= 80 ? 'text-success' : (diagnostics?.score || 0) >= 60 ? 'text-warning' : 'text-danger'

  return (
    <div className="space-y-8">
      {/* Score Card */}
      <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-8 shadow-2xl text-center">
        <Activity size={32} className="text-accent mx-auto mb-4" />
        <p className={cn('text-6xl font-bold tabular-nums', scoreColor)}>{diagnostics?.score || 0}</p>
        <p className="text-sm font-bold text-text-faint uppercase tracking-widest mt-2">Health Score</p>
      </div>

      {/* Check Results */}
      <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-8 shadow-2xl">
        <h3 className="text-lg font-bold text-text uppercase tracking-widest mb-6">Diagnostic Checks</h3>
        <div className="space-y-3">
          {diagnostics?.checks.map((check, i) => {
            const Icon = check.status === 'pass' ? CheckCircle : check.status === 'warn' ? AlertTriangle : XCircle
            const color = check.status === 'pass' ? 'text-success' : check.status === 'warn' ? 'text-warning' : 'text-danger'
            const bg = check.status === 'pass' ? 'bg-success/10' : check.status === 'warn' ? 'bg-warning/10' : 'bg-danger/10'

            return (
              <div key={i} className={cn('flex items-center justify-between p-4 rounded-xl border', bg, 'border-border/50')}>
                <div className="flex items-center gap-3">
                  <Icon size={18} className={color} />
                  <span className="text-sm font-bold text-text">{check.name}</span>
                </div>
                <div className="flex items-center gap-4">
                  <span className="text-sm font-bold text-text tabular-nums">{check.value}</span>
                  <span className={cn('text-xs font-bold', color)}>{check.message}</span>
                </div>
              </div>
            )
          })}
        </div>
      </div>

      {/* Quick Health Summary */}
      <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-8 shadow-2xl">
        <h3 className="text-lg font-bold text-text uppercase tracking-widest mb-4">Quick Health Check</h3>
        <div className="grid grid-cols-2 gap-3">
          {[
            { label: 'CPU < 80%', pass: (diagnostics?.checks.find(c => c.name === 'CPU Usage')?.status || 'fail') === 'pass' },
            { label: 'Memory < 75%', pass: (diagnostics?.checks.find(c => c.name === 'Memory Usage')?.status || 'fail') === 'pass' },
            { label: 'Disk > 20% free', pass: !diagnostics?.checks.some(c => c.name.startsWith('Disk') && c.status === 'fail') },
            { label: 'Swap OK', pass: (diagnostics?.checks.find(c => c.name === 'Swap Usage')?.status || 'pass') !== 'fail' },
            { label: 'Temperature OK', pass: (diagnostics?.checks.find(c => c.name === 'CPU Temperature')?.status || 'pass') !== 'fail' },
          ].map((item, i) => (
            <div key={i} className={cn('flex items-center gap-2 px-3 py-2 rounded-lg border', item.pass ? 'bg-success/10 border-success/30' : 'bg-danger/10 border-danger/30')}>
              {item.pass ? <CheckCircle size={14} className="text-success" /> : <XCircle size={14} className="text-danger" />}
              <span className={cn('text-sm font-bold', item.pass ? 'text-success' : 'text-danger')}>{item.label}</span>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
```

- [ ] **Step 5: Verify TypeScript compiles**

Run: `cd E:\Projects\projectx\AllOpsFull\cmd\opsforall-gui\frontend && npx tsc --noEmit`
Expected: No errors

- [ ] **Step 6: Commit**

```bash
git add cmd/opsforall-gui/frontend/src/pages/SysOps/
git commit -m "feat(sysops): implement MemoryTab, DiskTab, ProcessesTab, DiagnosticsTab"
```

---

### Task 3.5: Implement Remaining Tabs (Services, Logs, Users, Performance, Packages, Scheduler, Actions)

**Files:**
- Create: `cmd/opsforall-gui/frontend/src/pages/SysOps/ServicesTab.tsx`
- Create: `cmd/opsforall-gui/frontend/src/pages/SysOps/LogsTab.tsx`
- Create: `cmd/opsforall-gui/frontend/src/pages/SysOps/UsersTab.tsx`
- Create: `cmd/opsforall-gui/frontend/src/pages/SysOps/PerformanceTab.tsx`
- Create: `cmd/opsforall-gui/frontend/src/pages/SysOps/PackageManagerTab.tsx`
- Create: `cmd/opsforall-gui/frontend/src/pages/SysOps/SchedulerTab.tsx`
- Create: `cmd/opsforall-gui/frontend/src/pages/SysOps/ActionsTab.tsx`

- [ ] **Step 1: Create `ServicesTab.tsx` (delegates to DevOps backend)**

```tsx
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { Settings, Play, Square, RotateCcw } from 'lucide-react'
import { useBackend } from '@/hooks/useBackend'
import { useSettingsStore } from '@/stores/useSettingsStore'
import type { ServiceEntry } from '@/types'
import { cn } from '@/lib/utils'

export function ServicesTab() {
  const { call } = useBackend()
  const { refreshInterval } = useSettingsStore()
  const queryClient = useQueryClient()

  const { data: services = [] } = useQuery<ServiceEntry[]>({
    queryKey: ['devops-services'],
    queryFn: async () => { const r = await call('DevOps.GetServices'); return (r as ServiceEntry[]) || [] },
    refetchInterval: refreshInterval,
  })

  const controlService = async (name: string, action: string) => {
    await call('DevOps.ControlService', name, action)
    queryClient.invalidateQueries({ queryKey: ['devops-services'] })
  }

  const running = services.filter(s => s.status === 'Running').length
  const stopped = services.filter(s => s.status !== 'Running').length

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-6">
        <div className="flex items-center gap-2 px-4 py-2 bg-success/10 border border-success/30 rounded-xl">
          <Play size={14} className="text-success" />
          <span className="text-sm font-bold text-success tabular-nums">{running}</span>
          <span className="text-xs text-text-faint">Running</span>
        </div>
        <div className="flex items-center gap-2 px-4 py-2 bg-warning/10 border border-warning/30 rounded-xl">
          <Square size={14} className="text-warning" />
          <span className="text-sm font-bold text-warning tabular-nums">{stopped}</span>
          <span className="text-xs text-text-faint">Stopped</span>
        </div>
      </div>

      <div className="bg-panel border border-border rounded-xl overflow-hidden">
        <div className="max-h-[600px] overflow-y-auto">
          <table className="w-full text-left">
            <thead className="sticky top-0 bg-panel-2 border-b border-border">
              <tr>
                <th className="px-4 py-3 text-xs font-bold text-text-faint uppercase">Service</th>
                <th className="px-4 py-3 text-xs font-bold text-text-faint uppercase">Status</th>
                <th className="px-4 py-3 text-xs font-bold text-text-faint uppercase">Start Type</th>
                <th className="px-4 py-3 text-xs font-bold text-text-faint uppercase text-right">Actions</th>
              </tr>
            </thead>
            <tbody>
              {services.map((s, i) => (
                <tr key={i} className="border-b border-border/20 hover:bg-sidebar-hover">
                  <td className="px-4 py-3">
                    <span className="text-sm font-medium text-text">{s.display_name || s.name}</span>
                  </td>
                  <td className="px-4 py-3">
                    <span className={cn('text-xs font-bold px-2 py-0.5 rounded-full', s.status === 'Running' ? 'bg-success/20 text-success' : 'bg-warning/20 text-warning')}>
                      {s.status}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-sm text-text-dim">{s.start_type}</td>
                  <td className="px-4 py-3 text-right">
                    <div className="flex gap-1 justify-end">
                      {s.status !== 'Running' && (
                        <button onClick={() => controlService(s.name, 'start')} className="p-1.5 text-success hover:bg-success/10 rounded-lg"><Play size={14} /></button>
                      )}
                      {s.status === 'Running' && (
                        <>
                          <button onClick={() => controlService(s.name, 'stop')} className="p-1.5 text-danger hover:bg-danger/10 rounded-lg"><Square size={14} /></button>
                          <button onClick={() => controlService(s.name, 'restart')} className="p-1.5 text-warning hover:bg-warning/10 rounded-lg"><RotateCcw size={14} /></button>
                        </>
                      )}
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}
```

- [ ] **Step 2: Create `LogsTab.tsx`**

```tsx
import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Terminal, Filter } from 'lucide-react'
import { useBackend } from '@/hooks/useBackend'
import { useSettingsStore } from '@/stores/useSettingsStore'
import type { SystemLogsResult } from '@/types'
import { cn } from '@/lib/utils'

export function LogsTab() {
  const { call } = useBackend()
  const { refreshInterval } = useSettingsStore()
  const [source, setSource] = useState('journald')
  const [count, setCount] = useState(50)

  const { data: logs } = useQuery<SystemLogsResult>({
    queryKey: ['sysops-logs', source, count],
    queryFn: async () => { const r = await call('SysOps.GetSystemLogs', count, source); return r as SystemLogsResult },
    refetchInterval: refreshInterval,
  })

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <div className="flex gap-1 bg-panel border border-border rounded-lg p-1">
          {['journald', 'dmesg'].map(s => (
            <button key={s} onClick={() => setSource(s)} className={cn("px-3 py-1.5 rounded text-xs font-bold", source === s ? "bg-accent text-white" : "text-text-faint")}>
              {s}
            </button>
          ))}
        </div>
        <select value={count} onChange={(e) => setCount(Number(e.target.value))} className="bg-panel border border-border rounded-lg px-3 py-1.5 text-sm text-text">
          <option value={25}>25 lines</option>
          <option value={50}>50 lines</option>
          <option value={100}>100 lines</option>
          <option value={200}>200 lines</option>
        </select>
        <span className="text-sm text-text-faint">{logs?.total || 0} entries from {logs?.source || source}</span>
      </div>

      <div className="bg-panel-3 border border-border rounded-xl overflow-hidden">
        <div className="max-h-[600px] overflow-y-auto p-4">
          {logs?.entries.map((entry, i) => (
            <div key={i} className="flex items-start gap-3 py-1.5 border-b border-border/20 last:border-0">
              <span className={cn('text-xs font-bold px-1.5 py-0.5 rounded', entry.level === 'error' ? 'bg-danger/20 text-danger' : entry.level === 'warning' ? 'bg-warning/20 text-warning' : 'bg-panel-2 text-text-faint')}>
                {entry.level.toUpperCase()}
              </span>
              <span className="text-sm font-mono text-text-dim flex-1 break-all">{entry.message}</span>
            </div>
          ))}
          {(!logs?.entries || logs.entries.length === 0) && (
            <p className="text-text-faint text-sm text-center py-8">No log entries</p>
          )}
        </div>
      </div>
    </div>
  )
}
```

- [ ] **Step 3: Create `UsersTab.tsx`**

```tsx
import { useQuery } from '@tanstack/react-query'
import { Users } from 'lucide-react'
import { useBackend } from '@/hooks/useBackend'
import { useSettingsStore } from '@/stores/useSettingsStore'
import type { LoggedInUserData } from '@/types'

export function UsersTab() {
  const { call } = useBackend()
  const { refreshInterval } = useSettingsStore()

  const { data: users = [] } = useQuery<LoggedInUserData[]>({
    queryKey: ['sysops-users'],
    queryFn: async () => { const r = await call('SysOps.GetLoggedInUsers'); return (r as LoggedInUserData[]) || [] },
    refetchInterval: refreshInterval,
  })

  return (
    <div className="space-y-6">
      <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-8 shadow-2xl">
        <div className="flex items-center gap-3 mb-6">
          <Users size={20} className="text-success" />
          <h3 className="text-lg font-bold text-text uppercase tracking-widest">Active Users</h3>
          <span className="text-sm text-text-faint ml-auto">{users.length} logged in</span>
        </div>
        {users.length === 0 ? (
          <p className="text-text-dim text-center py-8">No active users detected</p>
        ) : (
          <div className="space-y-3">
            {users.map((u, i) => (
              <div key={i} className="flex items-center justify-between p-4 bg-panel-2 rounded-xl border border-border">
                <div>
                  <p className="text-sm font-bold text-text">{u.user}</p>
                  <p className="text-xs text-text-faint">{u.terminal} from {u.host || 'local'}</p>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
```

- [ ] **Step 4: Create `PerformanceTab.tsx`**

```tsx
import { useQuery } from '@tanstack/react-query'
import { Activity, Clock, Zap } from 'lucide-react'
import { useBackend } from '@/hooks/useBackend'
import { useSettingsStore } from '@/stores/useSettingsStore'
import type { PerformanceData } from '@/types'

export function PerformanceTab() {
  const { call } = useBackend()
  const { refreshInterval } = useSettingsStore()

  const { data: perf } = useQuery<PerformanceData>({
    queryKey: ['sysops-performance'],
    queryFn: async () => { const r = await call('SysOps.GetPerformanceStats'); return r as PerformanceData },
    refetchInterval: refreshInterval,
  })

  if (!perf) return <div className="animate-pulse h-32 bg-panel-2 rounded-xl" />

  const cpuUserPct = perf.cpu_times.total > 0 ? (perf.cpu_times.user / perf.cpu_times.total) * 100 : 0
  const cpuSystemPct = perf.cpu_times.total > 0 ? (perf.cpu_times.system / perf.cpu_times.total) * 100 : 0
  const cpuIdlePct = perf.cpu_times.total > 0 ? (perf.cpu_times.idle / perf.cpu_times.total) * 100 : 0

  return (
    <div className="space-y-8">
      {/* CPU Time Breakdown */}
      <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-8 shadow-2xl">
        <div className="flex items-center gap-3 mb-6">
          <Activity size={20} className="text-accent" />
          <h3 className="text-lg font-bold text-text uppercase tracking-widest">CPU Time Breakdown</h3>
        </div>
        <div className="grid grid-cols-4 gap-6">
          <div className="text-center">
            <p className="text-2xl font-bold text-accent tabular-nums">{cpuUserPct.toFixed(1)}%</p>
            <p className="text-xs font-bold text-text-faint uppercase">User</p>
          </div>
          <div className="text-center">
            <p className="text-2xl font-bold text-warning tabular-nums">{cpuSystemPct.toFixed(1)}%</p>
            <p className="text-xs font-bold text-text-faint uppercase">System</p>
          </div>
          <div className="text-center">
            <p className="text-2xl font-bold text-success tabular-nums">{cpuIdlePct.toFixed(1)}%</p>
            <p className="text-xs font-bold text-text-faint uppercase">Idle</p>
          </div>
          <div className="text-center">
            <p className="text-2xl font-bold text-danger tabular-nums">{perf.io_wait.toFixed(1)}%</p>
            <p className="text-xs font-bold text-text-faint uppercase">I/O Wait</p>
          </div>
        </div>
      </div>

      {/* Load Average */}
      <div className="bg-panel border border-border rounded-[var(--radius-lg)] p-8 shadow-2xl">
        <div className="flex items-center gap-3 mb-6">
          <Clock size={20} className="text-accent" />
          <h3 className="text-lg font-bold text-text uppercase tracking-widest">Load Average</h3>
        </div>
        <div className="grid grid-cols-3 gap-6">
          <div className="text-center">
            <p className="text-3xl font-bold text-text tabular-nums">{perf.load_average.load_1.toFixed(2)}</p>
            <p className="text-xs font-bold text-text-faint uppercase">1 min</p>
          </div>
          <div className="text-center">
            <p className="text-3xl font-bold text-text tabular-nums">{perf.load_average.load_5.toFixed(2)}</p>
            <p className="text-xs font-bold text-text-faint uppercase">5 min</p>
          </div>
          <div className="text-center">
            <p className="text-3xl font-bold text-text tabular-nums">{perf.load_average.load_15.toFixed(2)}</p>
            <p className="text-xs font-bold text-text-faint uppercase">15 min</p>
          </div>
        </div>
      </div>
    </div>
  )
}
```

- [ ] **Step 5: Create `PackageManagerTab.tsx`**

```tsx
import { useQuery } from '@tanstack/react-query'
import { Package, Search } from 'lucide-react'
import { useBackend } from '@/hooks/useBackend'
import { useSettingsStore } from '@/stores/useSettingsStore'
import type { PackageManagerData } from '@/types'
import { useState } from 'react'

export function PackageManagerTab() {
  const { call } = useBackend()
  const { refreshInterval } = useSettingsStore()
  const [search, setSearch] = useState('')

  const { data: managers = [] } = useQuery<PackageManagerData[]>({
    queryKey: ['sysops-packages'],
    queryFn: async () => { const r = await call('SysOps.GetInstalledPackages'); return (r as PackageManagerData[]) || [] },
    refetchInterval: refreshInterval,
  })

  const activeManager = managers.find(m => m.found)

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <div className="flex items-center gap-2 px-4 py-2 bg-panel-2 border border-border rounded-xl">
          <Package size={14} className="text-accent" />
          <span className="text-sm font-bold text-text">{activeManager?.name || 'None detected'}</span>
        </div>
        <span className="text-sm text-text-faint">{activeManager?.packages.length || 0} packages</span>
      </div>

      {activeManager?.found && (
        <>
          <div className="relative">
            <Search size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-text-faint" />
            <input
              type="text"
              placeholder="Search packages..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="w-full bg-panel border border-border rounded-xl pl-10 pr-4 py-2 text-sm text-text placeholder-text-faint focus:outline-none focus:border-accent"
            />
          </div>
          <div className="bg-panel border border-border rounded-xl overflow-hidden">
            <div className="max-h-[500px] overflow-y-auto">
              <table className="w-full text-left">
                <thead className="sticky top-0 bg-panel-2 border-b border-border">
                  <tr>
                    <th className="px-4 py-3 text-xs font-bold text-text-faint uppercase">Package</th>
                    <th className="px-4 py-3 text-xs font-bold text-text-faint uppercase">Version</th>
                  </tr>
                </thead>
                <tbody>
                  {activeManager.packages
                    .filter(p => p.name.toLowerCase().includes(search.toLowerCase()))
                    .map((p, i) => (
                      <tr key={i} className="border-b border-border/20 hover:bg-sidebar-hover">
                        <td className="px-4 py-2 text-sm font-medium text-text">{p.name}</td>
                        <td className="px-4 py-2 text-sm text-text-dim">{p.version}</td>
                      </tr>
                    ))}
                </tbody>
              </table>
            </div>
          </div>
        </>
      )}

      {!activeManager?.found && (
        <div className="text-center py-12">
          <Package size={48} className="text-text-faint mx-auto mb-4" />
          <p className="text-text-dim">No package manager detected</p>
          <p className="text-xs text-text-faint mt-2">Supported: apt, dnf, pacman, winget, choco</p>
        </div>
      )}
    </div>
  )
}
```

- [ ] **Step 6: Create `SchedulerTab.tsx`**

```tsx
import { useQuery } from '@tanstack/react-query'
import { Calendar, CheckCircle, XCircle } from 'lucide-react'
import { useBackend } from '@/hooks/useBackend'
import { useSettingsStore } from '@/stores/useSettingsStore'
import type { ScheduledTaskData } from '@/types'

export function SchedulerTab() {
  const { call } = useBackend()
  const { refreshInterval } = useSettingsStore()

  const { data: tasks = [] } = useQuery<ScheduledTaskData[]>({
    queryKey: ['sysops-scheduler'],
    queryFn: async () => { const r = await call('SysOps.GetScheduledTasks'); return (r as ScheduledTaskData[]) || [] },
    refetchInterval: refreshInterval,
  })

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-3 mb-2">
        <Calendar size={20} className="text-accent" />
        <h3 className="text-lg font-bold text-text uppercase tracking-widest">Scheduled Tasks</h3>
        <span className="text-sm text-text-faint ml-auto">{tasks.length} tasks</span>
      </div>

      {tasks.length === 0 ? (
        <div className="text-center py-12 bg-panel border border-border rounded-xl">
          <Calendar size={48} className="text-text-faint mx-auto mb-4" />
          <p className="text-text-dim">No scheduled tasks found</p>
        </div>
      ) : (
        <div className="bg-panel border border-border rounded-xl overflow-hidden">
          <div className="max-h-[500px] overflow-y-auto">
            <table className="w-full text-left">
              <thead className="sticky top-0 bg-panel-2 border-b border-border">
                <tr>
                  <th className="px-4 py-3 text-xs font-bold text-text-faint uppercase">Task</th>
                  <th className="px-4 py-3 text-xs font-bold text-text-faint uppercase">Schedule</th>
                  <th className="px-4 py-3 text-xs font-bold text-text-faint uppercase">Status</th>
                  <th className="px-4 py-3 text-xs font-bold text-text-faint uppercase">Next Run</th>
                </tr>
              </thead>
              <tbody>
                {tasks.map((t, i) => (
                  <tr key={i} className="border-b border-border/20 hover:bg-sidebar-hover">
                    <td className="px-4 py-3 text-sm font-medium text-text">{t.name}</td>
                    <td className="px-4 py-3 text-sm text-text-dim font-mono">{t.schedule}</td>
                    <td className="px-4 py-3">
                      {t.enabled ? <CheckCircle size={14} className="text-success" /> : <XCircle size={14} className="text-text-faint" />}
                    </td>
                    <td className="px-4 py-3 text-sm text-text-faint">{t.next_run || 'N/A'}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </div>
  )
}
```

- [ ] **Step 7: Create `ActionsTab.tsx`**

```tsx
import { useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { useBackend } from '@/hooks/useBackend'
import { ConfirmDialog } from '@/components/dialogs/ConfirmDialog'
import { CommandOutput } from '@/components/sysops/CommandOutput'
import type { ActionResult } from '@/types'
import { cn } from '@/lib/utils'
import { Power, RotateCcw, Moon, Snowflake, Wifi, Trash2, Download, RefreshCw } from 'lucide-react'

interface ActionDef {
  id: string
  label: string
  icon: React.ReactNode
  color: string
  bg: string
  border: string
  confirmMessage: string
  danger?: boolean
}

const actions: ActionDef[] = [
  { id: 'reboot', label: 'Reboot', icon: <RotateCcw size={20} />, color: 'text-warning', bg: 'bg-warning/10', border: 'border-warning/30', confirmMessage: 'Are you sure you want to reboot the system?' },
  { id: 'shutdown', label: 'Shutdown', icon: <Power size={20} />, color: 'text-danger', bg: 'bg-danger/10', border: 'border-danger/30', confirmMessage: 'Are you sure you want to shutdown the system?', danger: true },
  { id: 'sleep', label: 'Sleep', icon: <Moon size={20} />, color: 'text-accent', bg: 'bg-accent/10', border: 'border-accent/30', confirmMessage: 'Put system to sleep?' },
  { id: 'hibernate', label: 'Hibernate', icon: <Snowflake size={20} />, color: 'text-accent', bg: 'bg-accent/10', border: 'border-accent/30', confirmMessage: 'Put system to hibernate?' },
  { id: 'flush_dns', label: 'Flush DNS', icon: <Wifi size={20} />, color: 'text-success', bg: 'bg-success/10', border: 'border-success/30', confirmMessage: 'Flush DNS cache?' },
  { id: 'clear_temp', label: 'Clear Temp Files', icon: <Trash2 size={20} />, color: 'text-warning', bg: 'bg-warning/10', border: 'border-warning/30', confirmMessage: 'Clear temporary files?' },
  { id: 'clean_pkg_cache', label: 'Clean Package Cache', icon: <Download size={20} />, color: 'text-warning', bg: 'bg-warning/10', border: 'border-warning/30', confirmMessage: 'Clean package manager cache?' },
  { id: 'system_update', label: 'System Update', icon: <RefreshCw size={20} />, color: 'text-accent', bg: 'bg-accent/10', border: 'border-accent/30', confirmMessage: 'Run system update? This may take a while.' },
]

export function ActionsTab() {
  const { call } = useBackend()
  const [confirmTarget, setConfirmTarget] = useState<ActionDef | null>(null)
  const [result, setResult] = useState<ActionResult | null>(null)

  const executeMutation = useMutation({
    mutationFn: async (action: string) => {
      const r = await call('SysOps.RunSystemAction', action)
      return r as ActionResult
    },
    onSuccess: (data) => setResult(data),
  })

  const handleConfirm = () => {
    if (confirmTarget) {
      executeMutation.mutate(confirmTarget.id)
      setConfirmTarget(null)
    }
  }

  return (
    <div className="space-y-6">
      <ConfirmDialog
        open={confirmTarget !== null}
        title={`Execute ${confirmTarget?.label}`}
        description={confirmTarget?.confirmMessage || ''}
        type={confirmTarget?.danger ? 'danger' : 'warning'}
        confirmText="Execute"
        onConfirm={handleConfirm}
        onClose={() => setConfirmTarget(null)}
      />

      <div className="grid grid-cols-4 gap-4">
        {actions.map(action => (
          <button
            key={action.id}
            onClick={() => { setResult(null); setConfirmTarget(action) }}
            className={cn('flex flex-col items-center gap-3 p-6 rounded-xl border transition-all hover:scale-[1.02]', action.bg, action.border)}
          >
            <div className={action.color}>{action.icon}</div>
            <span className="text-sm font-bold text-text">{action.label}</span>
          </button>
        ))}
      </div>

      {result && (
        <div className={cn('p-4 rounded-xl border', result.success ? 'bg-success/10 border-success/30' : 'bg-danger/10 border-danger/30')}>
          <p className={cn('text-sm font-bold', result.success ? 'text-success' : 'text-danger')}>{result.message}</p>
          {result.output && <CommandOutput output={result.output} title="Command Output" />}
        </div>
      )}

      {executeMutation.isPending && (
        <div className="text-center py-8">
          <div className="animate-spin w-8 h-8 border-2 border-accent border-t-transparent rounded-full mx-auto" />
          <p className="text-sm text-text-faint mt-3">Executing action...</p>
        </div>
      )}
    </div>
  )
}
```

- [ ] **Step 8: Verify TypeScript compiles**

Run: `cd E:\Projects\projectx\AllOpsFull\cmd\opsforall-gui\frontend && npx tsc --noEmit`
Expected: No errors

- [ ] **Step 9: Commit**

```bash
git add cmd/opsforall-gui/frontend/src/pages/SysOps/
git commit -m "feat(sysops): implement all remaining category tabs (Services, Logs, Users, Performance, Packages, Scheduler, Actions)"
```

---

### Task 3.6: Rewrite Main SysOps.tsx with 14-Category Navigation

**Files:**
- Modify: `cmd/opsforall-gui/frontend/src/pages/SysOps.tsx`

- [ ] **Step 1: Rewrite `SysOps.tsx` with sidebar navigation and 14 categories**

This is the main restructure. The page becomes a sidebar navigator + content area pattern.

```tsx
import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import {
  Cpu, Server, MemoryStick, Disc, Activity, Settings, FileText,
  HardDrive, Users, BarChart3, Package, Calendar, Stethoscope,
  Zap, ChevronRight, Monitor,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { useBackend } from '@/hooks/useBackend'
import { useSettingsStore } from '@/stores/useSettingsStore'
import { DataFreshnessIndicator } from '@/components/ui/DataFreshnessIndicator'
import type { CPUInfo, MemoryInfo, SystemInfo, DiskInfo, MetricDataPoint } from '@/types'

// Tab imports
import { SystemInfoTab } from './SysOps/SystemInfoTab'
import { CpuTab } from './SysOps/CpuTab'
import { MemoryTab } from './SysOps/MemoryTab'
import { DiskTab } from './SysOps/DiskTab'
import { ProcessesTab } from './SysOps/ProcessesTab'
import { ServicesTab } from './SysOps/ServicesTab'
import { LogsTab } from './SysOps/LogsTab'
import { StorageTab } from './SysOps/StorageTab'
import { UsersTab } from './SysOps/UsersTab'
import { PerformanceTab } from './SysOps/PerformanceTab'
import { PackageManagerTab } from './SysOps/PackageManagerTab'
import { SchedulerTab } from './SysOps/SchedulerTab'
import { DiagnosticsTab } from './SysOps/DiagnosticsTab'
import { ActionsTab } from './SysOps/ActionsTab'

type SysOpsCategory = 'system-info' | 'cpu' | 'memory' | 'disk' | 'processes' | 'services' | 'logs' | 'storage' | 'users' | 'performance' | 'packages' | 'scheduler' | 'diagnostics' | 'actions'

interface CategoryDef {
  id: SysOpsCategory
  label: string
  icon: React.ReactNode
  group: 'inspection' | 'diagnosis' | 'action'
}

const categories: CategoryDef[] = [
  { id: 'system-info', label: 'System Info', icon: <Server size={18} />, group: 'inspection' },
  { id: 'cpu', label: 'CPU', icon: <Cpu size={18} />, group: 'inspection' },
  { id: 'memory', label: 'Memory', icon: <MemoryStick size={18} />, group: 'inspection' },
  { id: 'disk', label: 'Disk', icon: <Disc size={18} />, group: 'inspection' },
  { id: 'processes', label: 'Processes', icon: <Activity size={18} />, group: 'inspection' },
  { id: 'services', label: 'Services', icon: <Settings size={18} />, group: 'inspection' },
  { id: 'logs', label: 'Logs', icon: <FileText size={18} />, group: 'inspection' },
  { id: 'storage', label: 'Storage', icon: <HardDrive size={18} />, group: 'inspection' },
  { id: 'users', label: 'Users', icon: <Users size={18} />, group: 'inspection' },
  { id: 'performance', label: 'Performance', icon: <BarChart3 size={18} />, group: 'diagnosis' },
  { id: 'packages', label: 'Packages', icon: <Package size={18} />, group: 'diagnosis' },
  { id: 'scheduler', label: 'Scheduler', icon: <Calendar size={18} />, group: 'diagnosis' },
  { id: 'diagnostics', label: 'Diagnostics', icon: <Stethoscope size={18} />, group: 'diagnosis' },
  { id: 'actions', label: 'Actions', icon: <Zap size={18} />, group: 'action' },
]

// Re-export Bar for use in child tabs
export function Bar({ label, value, max = 100, color, unit = '%', showLabel = true }: { label: string; value: number; max?: number; color?: string; unit?: string; showLabel?: boolean }) {
  const pct = Math.min((value / max) * 100, 100)
  const barColor = color ?? (pct >= 70 ? 'var(--color-danger)' : pct >= 25 ? 'var(--color-warning)' : 'var(--color-success)')
  return (
    <div className="space-y-1">
      {showLabel && (
        <div className="flex items-center justify-between">
          <span className="text-text-dim text-sm font-medium">{label}</span>
          <span className="text-text font-bold text-sm tabular-nums">{value.toFixed(1)}{unit}</span>
        </div>
      )}
      <div className="h-3 bg-panel-3 rounded-full overflow-hidden border border-border">
        <div className="h-full rounded-full transition-all duration-700" style={{ width: `${pct}%`, background: `linear-gradient(90deg, ${barColor}88, ${barColor})` }} />
      </div>
    </div>
  )
}

export function SysOps() {
  const { call } = useBackend()
  const { refreshInterval } = useSettingsStore()
  const [activeCategory, setActiveCategory] = useState<SysOpsCategory>('diagnostics')

  // Shared queries
  const { data: cpuInfo, dataUpdatedAt: cpuUpdatedAt } = useQuery<CPUInfo>({
    queryKey: ['sysops-cpu'],
    queryFn: async () => { const r = await call('SysOps.GetCPUInfo'); return r as CPUInfo },
    refetchInterval: refreshInterval,
  })

  const { data: memInfo } = useQuery<MemoryInfo>({
    queryKey: ['sysops-mem'],
    queryFn: async () => { const r = await call('SysOps.GetMemoryInfo'); return r as MemoryInfo },
    refetchInterval: refreshInterval,
  })

  const { data: sysInfo } = useQuery<SystemInfo>({
    queryKey: ['sysops-sys'],
    queryFn: async () => { const r = await call('SysOps.GetSystemInfo'); return r as SystemInfo },
    refetchInterval: refreshInterval,
  })

  const { data: diskInfo } = useQuery<DiskInfo>({
    queryKey: ['sysops-disk'],
    queryFn: async () => { const r = await call('SysOps.GetDiskInfo'); return r as DiskInfo },
    refetchInterval: refreshInterval,
  })

  if (!cpuInfo || !memInfo || !sysInfo || !diskInfo) {
    return (
      <div className="space-y-4 animate-pulse">
        <div className="h-8 w-48 bg-panel-2 rounded" />
        <div className="grid grid-cols-2 gap-4">
          <div className="h-32 bg-panel-2 rounded" />
          <div className="h-32 bg-panel-2 rounded" />
        </div>
      </div>
    )
  }

  const inspectionCategories = categories.filter(c => c.group === 'inspection')
  const diagnosisCategories = categories.filter(c => c.group === 'diagnosis')
  const actionCategories = categories.filter(c => c.group === 'action')

  const renderContent = () => {
    switch (activeCategory) {
      case 'system-info': return <SystemInfoTab sysInfo={sysInfo} cpuInfo={cpuInfo} />
      case 'cpu': return <CpuTab cpuInfo={cpuInfo} />
      case 'memory': return <MemoryTab memInfo={memInfo} />
      case 'disk': return <DiskTab diskInfo={diskInfo} />
      case 'processes': return <ProcessesTab />
      case 'services': return <ServicesTab />
      case 'logs': return <LogsTab />
      case 'storage': return <DiskTab diskInfo={diskInfo} /> {/* Storage tab reuses Disk with I/O */}
      case 'users': return <UsersTab />
      case 'performance': return <PerformanceTab />
      case 'packages': return <PackageManagerTab />
      case 'scheduler': return <SchedulerTab />
      case 'diagnostics': return <DiagnosticsTab />
      case 'actions': return <ActionsTab />
      default: return <DiagnosticsTab />
    }
  }

  return (
    <div className="flex flex-col h-full bg-[var(--color-bg)]">
      {/* Header */}
      <div className="border-b border-border bg-panel-2 py-4 px-6">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold text-text flex items-center gap-4">
              <Monitor size={32} className="text-accent" /> SYSTEM OPERATIONS
            </h1>
            <p className="text-text-dim text-sm mt-1">Inspection, diagnosis, and action for system administrators.</p>
            <DataFreshnessIndicator lastUpdated={cpuUpdatedAt ? new Date(cpuUpdatedAt) : null} className="mt-1" />
          </div>
        </div>
      </div>

      {/* Content: Sidebar + Main */}
      <div className="flex-1 flex overflow-hidden">
        {/* Sidebar */}
        <div className="w-56 border-r border-border bg-panel-2 overflow-y-auto p-3">
          <CategoryGroup label="INSPECTION" categories={inspectionCategories} active={activeCategory} onSelect={setActiveCategory} />
          <CategoryGroup label="DIAGNOSIS" categories={diagnosisCategories} active={activeCategory} onSelect={setActiveCategory} />
          <CategoryGroup label="ACTION" categories={actionCategories} active={activeCategory} onSelect={setActiveCategory} />
        </div>

        {/* Main Content */}
        <div className="flex-1 overflow-y-auto p-8">
          {renderContent()}
        </div>
      </div>
    </div>
  )
}

function CategoryGroup({ label, categories, active, onSelect }: { label: string; categories: CategoryDef[]; active: SysOpsCategory; onSelect: (id: SysOpsCategory) => void }) {
  return (
    <div className="mb-4">
      <p className="text-[10px] font-bold text-text-faint uppercase tracking-widest px-3 mb-2">{label}</p>
      {categories.map(cat => (
        <button
          key={cat.id}
          onClick={() => onSelect(cat.id)}
          className={cn(
            'w-full flex items-center gap-3 px-3 py-2 rounded-lg text-sm font-bold transition-all mb-0.5',
            active === cat.id ? 'bg-accent text-white' : 'text-text-dim hover:text-text hover:bg-sidebar-hover'
          )}
        >
          {cat.icon}
          {cat.label}
        </button>
      ))}
    </div>
  )
}
```

- [ ] **Step 2: Verify the full build**

Run: `cd E:\Projects\projectx\AllOpsFull\cmd\opsforall-gui\frontend && npm run build`
Expected: Build succeeds

- [ ] **Step 3: Run TypeScript check**

Run: `cd E:\Projects\projectx\AllOpsFull\cmd\opsforall-gui\frontend && npx tsc --noEmit`
Expected: No errors

- [ ] **Step 4: Run lint**

Run: `cd E:\Projects\projectx\AllOpsFull\cmd\opsforall-gui\frontend && npm run lint`
Expected: Clean

- [ ] **Step 5: Run tests**

Run: `cd E:\Projects\projectx\AllOpsFull\cmd\opsforall-gui\frontend && npm test -- --run`
Expected: All tests pass

- [ ] **Step 6: Commit**

```bash
git add cmd/opsforall-gui/frontend/src/pages/SysOps.tsx
git commit -m "feat(sysops): restructure SysOps into 14-category sidebar layout with inspection/diagnosis/action groups"
```

---

## Phase 4: Verification

### Task 4.1: Full Build and Test Verification

- [ ] **Step 1: Run full Go build**

Run: `cd E:\Projects\projectx\AllOpsFull && go build ./...`
Expected: Build succeeds

- [ ] **Step 2: Run Go tests**

Run: `cd E:\Projects\projectx\AllOpsFull && go test ./internal/sysops/ -v`
Expected: All tests pass

- [ ] **Step 3: Run frontend build**

Run: `cd E:\Projects\projectx\AllOpsFull\cmd\opsforall-gui\frontend && npm run build`
Expected: Build succeeds

- [ ] **Step 4: Run frontend tests**

Run: `cd E:\Projects\projectx\AllOpsFull\cmd\opsforall-gui\frontend && npm test -- --run`
Expected: All tests pass

- [ ] **Step 5: Run TypeScript check**

Run: `cd E:\Projects\projectx\AllOpsFull\cmd\opsforall-gui\frontend && npx tsc --noEmit`
Expected: No errors

- [ ] **Step 6: Run lint**

Run: `cd E:\Projects\projectx\AllOpsFull\cmd\opsforall-gui\frontend && npm run lint`
Expected: Clean

---

## Known Issues to Address

1. **Top Disk IO Bug**: The current "Top Disk IO" card in SysOps references `read_bytes`/`write_bytes` on ProcessInfo which don't exist in the backend. This will be removed in the restructure.

2. **StorageTab Duplication**: The Storage tab currently reuses DiskTab. If SMART health is needed later, it will require `smartctl` CLI integration.

3. **Process Tree**: The `GetProcessTree` method doesn't build a real hierarchy client-side. The `ProcessTreeItem` component manually reconstructs it. This works but is inefficient for large process lists.

4. **Service Entry Type**: The ServicesTab references `ServiceEntry` from DevOps types. Ensure this type exists in `types/index.ts`.

5. **Package Manager Detection**: The `getWingetPackages` and `getChocoPackages` functions use simple string parsing which may break with output format changes.

6. **Platform-Specific Commands**: System actions assume specific command formats. Testing on both Windows and Linux is needed.

---

## Execution Summary

| Phase | Tasks | Estimated Time |
|-------|-------|---------------|
| Phase 1: Backend Foundation | 10 tasks | 2-3 hours |
| Phase 2: Frontend Types | 1 task | 15 minutes |
| Phase 3: Frontend Restructure | 6 tasks | 3-4 hours |
| Phase 4: Verification | 1 task | 30 minutes |
| **Total** | **18 tasks** | **6-8 hours** |

---

## Alternative: Simplified Approach

If the full 14-category rework is too large, a simplified approach would be:

1. **Keep existing 3-tab structure** (Analysis, Runtime, Inventory)
2. **Add new sections to existing tabs** (e.g., Performance section in Analysis, Actions section in Inventory)
3. **Add new backend methods** for the missing features
4. **Skip categories that overlap with other modules** (Services → DevOps, Users → SecOps)

This would reduce the scope to ~8 tasks and ~3 hours of work.

---

**Plan complete and saved to `docs/superpowers/plans/2026-07-14-sysops-rework.md`. Two execution options:**

**1. Subagent-Driven (recommended)** - I dispatch a fresh subagent per task, review between tasks, fast iteration

**2. Inline Execution** - Execute tasks in this session using executing-plans, batch execution with checkpoints

**Which approach?**