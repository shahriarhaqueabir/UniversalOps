package sysops

import (
	"github.com/shirou/gopsutil/v4/mem"
)

// MemoryStats holds RAM information.
type MemoryStats struct {
	TotalBytes     uint64
	AvailableBytes uint64
	UsedBytes      uint64
	UsedPercent    float64
	SwapTotal      uint64
	SwapUsed       uint64
	SwapPercent    float64
}

// GetMemoryStats returns current memory usage.
func GetMemoryStats() (*MemoryStats, error) {
	v, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}

	swap, err := mem.SwapMemory()
	swapTotal := uint64(0)
	swapUsed := uint64(0)
	swapPct := 0.0
	if err == nil {
		swapTotal = swap.Total
		swapUsed = swap.Used
		swapPct = swap.UsedPercent
	}

	return &MemoryStats{
		TotalBytes:     v.Total,
		AvailableBytes: v.Available,
		UsedBytes:      v.Used,
		UsedPercent:    v.UsedPercent,
		SwapTotal:      swapTotal,
		SwapUsed:       swapUsed,
		SwapPercent:    swapPct,
	}, nil
}
