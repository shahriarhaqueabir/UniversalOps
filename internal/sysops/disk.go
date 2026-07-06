package sysops

import (
	"github.com/shirou/gopsutil/v4/disk"
)

// DiskStats holds disk usage information.
type DiskStats struct {
	Partitions []disk.PartitionStat
	Usage      []DiskUsage
}

// DiskUsage holds usage stats for a single partition.
type DiskUsage struct {
	Mountpoint  string
	TotalBytes  uint64
	FreeBytes   uint64
	UsedBytes   uint64
	UsedPercent float64
	FSType      string
	Device      string
}

// GetDiskStats returns disk partition and usage information.
func GetDiskStats() (*DiskStats, error) {
	parts, err := disk.Partitions(false)
	if err != nil {
		return nil, err
	}

	stats := &DiskStats{}
	for _, part := range parts {
		usage, err := disk.Usage(part.Mountpoint)
		if err != nil {
			continue // skip partitions we can't read
		}
		stats.Partitions = append(stats.Partitions, part)
		stats.Usage = append(stats.Usage, DiskUsage{
			Mountpoint:  part.Mountpoint,
			TotalBytes:  usage.Total,
			FreeBytes:   usage.Free,
			UsedBytes:   usage.Used,
			UsedPercent: usage.UsedPercent,
			FSType:      part.Fstype,
			Device:      part.Device,
		})
	}

	return stats, nil
}
