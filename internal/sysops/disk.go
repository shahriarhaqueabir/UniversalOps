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

// DiskIOStat holds I/O stats for a single disk.
type DiskIOStat struct {
	Name        string `json:"name"`
	ReadBytes   uint64 `json:"read_bytes"`
	WriteBytes  uint64 `json:"write_bytes"`
	ReadCount   uint64 `json:"read_count"`
	WriteCount  uint64 `json:"write_count"`
	ReadTimeMs  uint64 `json:"read_time_ms"`
	WriteTimeMs uint64 `json:"write_time_ms"`
}

// DiskIOStats holds aggregate disk I/O information.
type DiskIOStats struct {
	Disks      []DiskIOStat `json:"disks"`
	TotalRead  uint64       `json:"total_read_bytes"`
	TotalWrite uint64       `json:"total_write_bytes"`
}

// GetDiskIO returns disk I/O throughput statistics.
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
