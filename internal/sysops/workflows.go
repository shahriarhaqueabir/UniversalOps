package sysops

import (
	"fmt"
	"strings"

	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
)

// HealthReport is a combined system health report.
type HealthReport struct {
	CPU    *CPUStats
	Memory *MemoryStats
	Disk   *DiskUsage
	System *SystemInfo
	Procs  []ProcessInfo
}

// RunHealthCheck collects all system stats and returns a combined report.
func RunHealthCheck() (*HealthReport, error) {
	report := &HealthReport{}
	var errs []string

	cpu, err := GetCPUStats()
	if err != nil {
		errs = append(errs, fmt.Sprintf("CPU: %v", err))
	} else {
		report.CPU = cpu
	}

	mem, err := GetMemoryStats()
	if err != nil {
		errs = append(errs, fmt.Sprintf("Memory: %v", err))
	} else {
		report.Memory = mem
	}

	// Disk
	diskStats, err := GetDiskStats()
	if err != nil {
		errs = append(errs, fmt.Sprintf("Disk: %v", err))
	} else if len(diskStats.Usage) > 0 {
		report.Disk = &diskStats.Usage[0]
	}

	sys, err := GetSystemInfo()
	if err != nil {
		errs = append(errs, fmt.Sprintf("System: %v", err))
	} else {
		report.System = sys
	}

	procs, err := GetTopProcesses(10)
	if err != nil {
		errs = append(errs, fmt.Sprintf("Processes: %v", err))
	} else {
		report.Procs = procs
	}

	if len(errs) > 0 && report.CPU == nil && report.Memory == nil && report.Disk == nil && report.System == nil {
		return nil, fmt.Errorf("all collectors failed: %s", strings.Join(errs, "; "))
	}

	return report, nil
}

// String returns a plain-text summary of the health report.
func (r *HealthReport) String() string {
	var b strings.Builder

	b.WriteString("=== System Health Report ===\n\n")

	if r.CPU != nil {
		b.WriteString(fmt.Sprintf("CPU: %.1f%% across %d cores\n", r.CPU.Percent, r.CPU.CoreCount))
	}
	if r.Memory != nil {
		usedGB := float64(r.Memory.TotalBytes-r.Memory.AvailableBytes) / (1024 * 1024 * 1024)
		totalGB := float64(r.Memory.TotalBytes) / (1024 * 1024 * 1024)
		b.WriteString(fmt.Sprintf("MEM: %.1f%% used (%.1f/%.1f GB)\n", r.Memory.UsedPercent, usedGB, totalGB))
	}
	if r.Disk != nil {
		b.WriteString(fmt.Sprintf("DISK: %.1f%% used (%s/%s)\n",
			r.Disk.UsedPercent,
			common.FormatBytes(r.Disk.UsedBytes),
			common.FormatBytes(r.Disk.TotalBytes)))
	}
	if r.System != nil {
		b.WriteString(fmt.Sprintf("UPTIME: %s\n", common.FormatUptime(r.System.UptimeSeconds)))
		b.WriteString(fmt.Sprintf("PROCS: %d running\n", r.System.ProcessCount))
	}
	if len(r.Procs) > 0 {
		b.WriteString(fmt.Sprintf("\nTop %d Processes (by CPU):\n", len(r.Procs)))
		for _, p := range r.Procs {
			b.WriteString(fmt.Sprintf("  PID %-6d %-25s CPU %5.1f%% MEM %5.1f MB\n",
				p.PID, p.Name, p.CPU, p.Memory))
		}
	}

	return b.String()
}

// Markdown returns a markdown-formatted health report.
func (r *HealthReport) Markdown() string {
	var b strings.Builder

	b.WriteString("# 🖥 System Health Report\n\n")

	if r.System != nil {
		b.WriteString("## System Overview\n\n")
		b.WriteString("| Metric | Value |\n|--------|-------|\n")
		b.WriteString(fmt.Sprintf("| Hostname | %s |\n", r.System.Hostname))
		b.WriteString(fmt.Sprintf("| OS | %s %s |\n", r.System.Platform, r.System.PlatformVersion))
		b.WriteString(fmt.Sprintf("| Kernel | %s |\n", r.System.KernelVersion))
		b.WriteString(fmt.Sprintf("| Uptime | %s |\n", common.FormatUptime(r.System.UptimeSeconds)))
		b.WriteString(fmt.Sprintf("| Processes | %d |\n", r.System.ProcessCount))
		b.WriteString("\n")
	}

	b.WriteString("## Resource Usage\n\n")
	b.WriteString("| Resource | Usage |\n|----------|-------|\n")

	if r.CPU != nil {
		b.WriteString(fmt.Sprintf("| CPU | %.1f%% (%d cores) |\n", r.CPU.Percent, r.CPU.CoreCount))
	}
	if r.Memory != nil {
		usedGB := float64(r.Memory.TotalBytes-r.Memory.AvailableBytes) / (1024 * 1024 * 1024)
		totalGB := float64(r.Memory.TotalBytes) / (1024 * 1024 * 1024)
		b.WriteString(fmt.Sprintf("| Memory | %.1f%% (%.1f/%.1f GB) |\n", r.Memory.UsedPercent, usedGB, totalGB))
	}
	if r.Disk != nil {
		b.WriteString(fmt.Sprintf("| Disk | %.1f%% (%s/%s) |\n",
			r.Disk.UsedPercent,
			common.FormatBytes(r.Disk.UsedBytes),
			common.FormatBytes(r.Disk.TotalBytes)))
	}

	if len(r.Procs) > 0 {
		b.WriteString("\n## Top Processes\n\n")
		b.WriteString("| PID | Name | CPU % | Memory (MB) |\n|-----|------|-------|-------------|\n")
		for _, p := range r.Procs {
			b.WriteString(fmt.Sprintf("| %d | %s | %.1f | %.1f |\n", p.PID, p.Name, p.CPU, p.Memory))
		}
	}

	return b.String()
}
