package app

import (
	"context"
	"time"

	"github.com/shirou/gopsutil/v4/sensors"

	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
	"github.com/shahriarhaqueabir/AllOpsFull/internal/sysops"
)

// ── CPU Collector ────────────────────────────────────────────────────────────

type cpuCollector struct{}

func (c *cpuCollector) Info() common.CollectorInfo {
	return common.CollectorInfo{
		ID:              common.CollectorCPU,
		Name:            "CPU Usage",
		Description:     "CPU utilization percentage",
		DefaultInterval: 3 * time.Second,
		DefaultEnabled:  true,
	}
}

func (c *cpuCollector) Collect(ctx context.Context) ([]common.MetricSample, error) {
	stats, err := sysops.GetCPUStats()
	if err != nil {
		return nil, err
	}
	return []common.MetricSample{
		{Name: "cpu.percent", Unit: "%", Value: stats.Percent},
	}, nil
}

// ── Memory Collector ─────────────────────────────────────────────────────────

type memoryCollector struct{}

func (c *memoryCollector) Info() common.CollectorInfo {
	return common.CollectorInfo{
		ID:              common.CollectorMem,
		Name:            "Memory Usage",
		Description:     "RAM and swap utilization percentage",
		DefaultInterval: 5 * time.Second,
		DefaultEnabled:  true,
	}
}

func (c *memoryCollector) Collect(ctx context.Context) ([]common.MetricSample, error) {
	stats, err := sysops.GetMemoryStats()
	if err != nil {
		return nil, err
	}
	return []common.MetricSample{
		{Name: "memory.percent", Unit: "%", Value: stats.UsedPercent},
	}, nil
}

// ── Disk Collector ──────────────────────────────────────────────────────────

type diskCollector struct{}

func (c *diskCollector) Info() common.CollectorInfo {
	return common.CollectorInfo{
		ID:              common.CollectorDisk,
		Name:            "Disk Usage",
		Description:     "Primary disk partition utilization percentage",
		DefaultInterval: 10 * time.Second,
		DefaultEnabled:  true,
	}
}

func (c *diskCollector) Collect(ctx context.Context) ([]common.MetricSample, error) {
	stats, err := sysops.GetDiskStats()
	if err != nil {
		return nil, err
	}
	pct, _ := primaryDiskUsage(stats)
	return []common.MetricSample{
		{Name: "disk.percent", Unit: "%", Value: pct},
	}, nil
}

func primaryDiskUsage(stats *sysops.DiskStats) (float64, uint64) {
	if stats == nil {
		return 0, 0
	}
	for _, usage := range stats.Usage {
		if usage.Mountpoint == "/" || usage.Mountpoint == "C:\\" {
			return usage.UsedPercent, usage.FreeBytes
		}
	}
	if len(stats.Usage) > 0 {
		return stats.Usage[0].UsedPercent, stats.Usage[0].FreeBytes
	}
	return 0, 0
}

// ── Network Collector ─────────────────────────────────────────────────────────

type networkCollector struct {
	netOps *NetOps
}

func (c *networkCollector) Info() common.CollectorInfo {
	return common.CollectorInfo{
		ID:              common.CollectorNet,
		Name:            "Network Traffic",
		Description:     "Interface bandwidth rates (RX/TX)",
		DefaultInterval: 5 * time.Second,
		DefaultEnabled:  true,
	}
}

func (c *networkCollector) Collect(ctx context.Context) ([]common.MetricSample, error) {
	ifaces, err := c.netOps.collectInterfaces()
	if err != nil {
		return nil, err
	}
	var rxTotal, txTotal float64
	for _, iface := range ifaces {
		rxTotal += iface.RXRateBps
		txTotal += iface.TXRateBps
	}
	return []common.MetricSample{
		{Name: "network.rx.rate", Unit: "bps", Value: rxTotal},
		{Name: "network.tx.rate", Unit: "bps", Value: txTotal},
	}, nil
}

// ── Temperature Collector ────────────────────────────────────────────────────

type temperatureCollector struct{}

func (c *temperatureCollector) Info() common.CollectorInfo {
	return common.CollectorInfo{
		ID:              common.CollectorTemp,
		Name:            "CPU Temperature",
		Description:     "CPU core temperature readings",
		DefaultInterval: 15 * time.Second,
		DefaultEnabled:  false,
	}
}

func (c *temperatureCollector) Collect(ctx context.Context) ([]common.MetricSample, error) {
	temps, err := sensors.SensorsTemperatures()
	if err != nil {
		return nil, err
	}
	if len(temps) == 0 {
		return nil, nil
	}
	return []common.MetricSample{
		{Name: "cpu.temperature", Unit: "°C", Value: temps[0].Temperature},
	}, nil
}

// ── Process Count Collector ─────────────────────────────────────────────────

type processCountCollector struct{}

func (c *processCountCollector) Info() common.CollectorInfo {
	return common.CollectorInfo{
		ID:              common.CollectorProc,
		Name:            "Process Count",
		Description:     "Total running processes on the system",
		DefaultInterval: 15 * time.Second,
		DefaultEnabled:  true,
	}
}

func (c *processCountCollector) Collect(ctx context.Context) ([]common.MetricSample, error) {
	info, err := sysops.GetSystemInfo()
	if err != nil {
		return nil, err
	}
	return []common.MetricSample{
		{Name: "process.count", Unit: "count", Value: float64(info.ProcessCount)},
	}, nil
}

// ── Factory ───────────────────────────────────────────────────────────────────

func RegisterCollectors(registry *common.CollectorRegistry, app *App) {
	registry.Register(&cpuCollector{})
	registry.Register(&memoryCollector{})
	registry.Register(&diskCollector{})
	registry.Register(&networkCollector{netOps: app.NetOps})
	registry.Register(&temperatureCollector{})
	registry.Register(&processCountCollector{})
}
