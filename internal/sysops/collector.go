package sysops

import (
	"fmt"
	"strings"

	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
)

// CollectAllStats gathers all system metrics into a single snapshot.
func CollectAllStats() (*common.SystemStats, error) {
	var errs []string
	successes := 0

	cpuStats, err := GetCPUStats()
	if err != nil {
		errs = append(errs, fmt.Sprintf("cpu: %v", err))
	} else {
		successes++
	}

	memStats, err := GetMemoryStats()
	if err != nil {
		errs = append(errs, fmt.Sprintf("memory: %v", err))
	} else {
		successes++
	}

	diskStats, err := GetDiskStats()
	if err != nil {
		errs = append(errs, fmt.Sprintf("disk: %v", err))
	} else {
		successes++
	}

	info, err := GetSystemInfo()
	if err != nil {
		errs = append(errs, fmt.Sprintf("system: %v", err))
	} else {
		successes++
	}

	stats := buildSystemStats(cpuStats, memStats, diskStats, info)
	if successes > 0 {
		return stats, nil
	}

	if len(errs) > 0 {
		return nil, fmt.Errorf("collect system stats: %s", strings.Join(errs, "; "))
	}
	return stats, nil
}

func buildSystemStats(cpuStats *CPUStats, memStats *MemoryStats, diskStats *DiskStats, info *SystemInfo) *common.SystemStats {
	stats := &common.SystemStats{}

	if cpuStats != nil {
		stats.CPUPercent = cpuStats.Percent
	}

	if memStats != nil {
		stats.MemoryUsed = memStats.UsedPercent
		stats.MemoryTotal = memStats.TotalBytes
		stats.MemoryUsedGB = float64(memStats.UsedBytes) / (1024 * 1024 * 1024)
		stats.MemoryTotalGB = float64(memStats.TotalBytes) / (1024 * 1024 * 1024)
	}

	if diskStats != nil {
		stats.DiskUsed, stats.DiskFree = primaryDiskUsage(diskStats)
	}

	if info != nil {
		stats.Uptime = common.FormatUptime(info.UptimeSeconds)
		stats.ProcessCount = info.ProcessCount
	}

	return stats
}

func primaryDiskUsage(diskStats *DiskStats) (float64, uint64) {
	if diskStats == nil {
		return 0, 0
	}

	for _, usage := range diskStats.Usage {
		if usage.Mountpoint == "/" || usage.Mountpoint == "C:\\" {
			return usage.UsedPercent, usage.FreeBytes
		}
	}
	if len(diskStats.Usage) > 0 {
		return diskStats.Usage[0].UsedPercent, diskStats.Usage[0].FreeBytes
	}
	return 0, 0
}
