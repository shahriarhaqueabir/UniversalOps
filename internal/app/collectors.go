package app

import (
	"context"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/sensors"

	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
	"github.com/shahriarhaqueabir/AllOpsFull/internal/netops"
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
	netOps         *NetOps
	speedCache     map[string]int64
	speedCacheTime time.Time
	speedCacheMu   sync.Mutex
}

func (c *networkCollector) Info() common.CollectorInfo {
	return common.CollectorInfo{
		ID:              common.CollectorNet,
		Name:            "Network Traffic",
		Description:     "Interface bandwidth rates (RX/TX)",
		DefaultInterval: 15 * time.Second, // Bumped from 5s to 15s — each collect takes 1.1-1.6s
		DefaultEnabled:  true,
	}
}

func (c *networkCollector) Collect(ctx context.Context) ([]common.MetricSample, error) {
	// Refresh speed cache every 5 minutes to keep link speed changes in view
	c.speedCacheMu.Lock()
	if c.speedCache == nil || time.Since(c.speedCacheTime) > 5*time.Minute {
		fresh := netops.GetLinkSpeeds()
		c.speedCache = fresh
		c.speedCacheTime = time.Now()
	}
	cache := c.speedCache
	c.speedCacheMu.Unlock()

	ifaces, err := c.netOps.collectInterfaces(cache)
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
	return []common.MetricSample{
		{Name: "process.count", Unit: "count", Value: float64(sysops.GetProcessCount())},
	}, nil
}

// ── Uptime Collector ────────────────────────────────────────────────────────

type uptimeCollector struct{}

func (c *uptimeCollector) Info() common.CollectorInfo {
	return common.CollectorInfo{
		ID:              common.CollectorUptime,
		Name:            "System Uptime",
		Description:     "Seconds since system boot",
		DefaultInterval: 60 * time.Second,
		DefaultEnabled:  true,
	}
}

func (c *uptimeCollector) Collect(ctx context.Context) ([]common.MetricSample, error) {
	info, err := sysops.GetSystemInfo()
	if err != nil {
		return nil, err
	}
	return []common.MetricSample{
		{Name: "system.uptime", Unit: "s", Value: float64(info.UptimeSeconds)},
	}, nil
}

// ── Load Average Collector ──────────────────────────────────────────────────

type loadAvgCollector struct{}

func (c *loadAvgCollector) Info() common.CollectorInfo {
	return common.CollectorInfo{
		ID:              common.CollectorLoad,
		Name:            "Load Average",
		Description:     "System load averages (1m, 5m, 15m)",
		DefaultInterval: 10 * time.Second,
		DefaultEnabled:  true,
	}
}

func (c *loadAvgCollector) Collect(ctx context.Context) ([]common.MetricSample, error) {
	avg, err := load.Avg()
	if err != nil {
		return nil, err
	}
	return []common.MetricSample{
		{Name: "load.1m", Unit: "load", Value: avg.Load1},
		{Name: "load.5m", Unit: "load", Value: avg.Load5},
		{Name: "load.15m", Unit: "load", Value: avg.Load15},
	}, nil
}

// ── Swap Collector ──────────────────────────────────────────────────────────

type swapCollector struct{}

func (c *swapCollector) Info() common.CollectorInfo {
	return common.CollectorInfo{
		ID:              common.CollectorSwap,
		Name:            "Swap Usage",
		Description:     "Swap space utilization percentage",
		DefaultInterval: 15 * time.Second,
		DefaultEnabled:  true,
	}
}

func (c *swapCollector) Collect(ctx context.Context) ([]common.MetricSample, error) {
	stats, err := sysops.GetMemoryStats()
	if err != nil {
		return nil, err
	}
	return []common.MetricSample{
		{Name: "swap.percent", Unit: "%", Value: stats.SwapPercent},
	}, nil
}

// ── Disk I/O Collector ──────────────────────────────────────────────────────

type diskIOCollector struct{}

func (c *diskIOCollector) Info() common.CollectorInfo {
	return common.CollectorInfo{
		ID:              common.CollectorDiskIO,
		Name:            "Disk I/O",
		Description:     "Aggregate disk read/write bytes",
		DefaultInterval: 10 * time.Second,
		DefaultEnabled:  true,
	}
}

func (c *diskIOCollector) Collect(ctx context.Context) ([]common.MetricSample, error) {
	stats, err := sysops.GetDiskIO()
	if err != nil {
		return nil, err
	}
	return []common.MetricSample{
		{Name: "disk.io.read", Unit: "bytes", Value: float64(stats.TotalRead)},
		{Name: "disk.io.write", Unit: "bytes", Value: float64(stats.TotalWrite)},
	}, nil
}

// ── Open File Descriptors Collector ─────────────────────────────────────────

type openFDCollector struct{}

func (c *openFDCollector) Info() common.CollectorInfo {
	return common.CollectorInfo{
		ID:              common.CollectorOpenFD,
		Name:            "Open File Descriptors",
		Description:     "Total open file descriptors across all processes",
		DefaultInterval: 15 * time.Second,
		DefaultEnabled:  false, // expensive, now uses cached snapshot
	}
}

func (c *openFDCollector) Collect(ctx context.Context) ([]common.MetricSample, error) {
	return []common.MetricSample{
		{Name: "system.open_fds", Unit: "count", Value: float64(sysops.GetTotalOpenFDs())},
	}, nil
}

// ── GPU Collector ────────────────────────────────────────────────────────────

type gpuCollector struct {
	sysOps *SysOps
}

func (c *gpuCollector) Info() common.CollectorInfo {
	return common.CollectorInfo{
		ID:              common.CollectorID("gpu"),
		Name:            "GPU Status",
		Description:     "GPU hardware details and memory",
		DefaultInterval: 30 * time.Second,
		DefaultEnabled:  true,
	}
}

func (c *gpuCollector) Collect(ctx context.Context) ([]common.MetricSample, error) {
	gpu := c.sysOps.GetGPUInfo()
	if !gpu.Detected {
		return nil, nil
	}
	return []common.MetricSample{
		{Name: "gpu.memory.total", Unit: "GB", Value: gpu.MemoryGB},
	}, nil
}

// ── Factory ───────────────────────────────────────────────────────────────────

func RegisterCollectors(registry *common.CollectorRegistry, app *App) {
	registry.Register(&cpuCollector{})
	registry.Register(&memoryCollector{})
	registry.Register(&diskCollector{})
	registry.Register(&networkCollector{netOps: app.NetOps, speedCache: make(map[string]int64)})
	registry.Register(&gpuCollector{sysOps: app.SysOps})
	registry.Register(&temperatureCollector{})
	registry.Register(&processCountCollector{})
	registry.Register(&uptimeCollector{})
	registry.Register(&loadAvgCollector{})
	registry.Register(&swapCollector{})
	registry.Register(&diskIOCollector{})
	registry.Register(&openFDCollector{})
}
