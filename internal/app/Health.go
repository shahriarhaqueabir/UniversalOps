package app

import (
	"fmt"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"

	"github.com/shahriarhaqueabir/UniversalOps/internal/common"
)

// ── Health types exposed to the frontend ──

// CollectorHealth describes the operational health of a single collector.
type CollectorHealth struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Enabled   bool   `json:"enabled"`
	Status    string `json:"status"` // "healthy" | "degraded" | "critical" | "stale"
	LastRun   string `json:"last_run"`
	IntervalMs int   `json:"interval_ms"`
	Error     string `json:"error,omitempty"`
}

// HealthSummary is the top-level health response returned to the frontend.
type HealthSummary struct {
	Overall  string            `json:"overall"`  // "healthy" | "degraded" | "critical"
	Uptime   string            `json:"uptime"`
	CPU      SystemHealthMetric `json:"cpu"`
	Memory   SystemHealthMetric `json:"memory"`
	Disk     SystemHealthMetric `json:"disk"`
	Load     SystemHealthMetric `json:"load"`
	Alerts   int               `json:"alerts"`
	Collectors []CollectorHealth `json:"collectors"`
}

// SystemHealthMetric holds a named measurement with its status.
type SystemHealthMetric struct {
	Name    string  `json:"name"`
	Value   float64 `json:"value"`
	Unit    string  `json:"unit"`
	Status  string  `json:"status"` // "healthy" | "degraded" | "critical"
	Message string  `json:"message"`
}

// Health exposes system health bindings to the frontend.
type Health struct {
	pipeline *common.DataPipeline
	alerts   *common.AlertEngine
	registry *common.CollectorRegistry
	appStart time.Time
}

// NewHealth creates a new Health facade.
func NewHealth(pipeline *common.DataPipeline, alerts *common.AlertEngine, registry *common.CollectorRegistry) *Health {
	return &Health{
		pipeline: pipeline,
		alerts:   alerts,
		registry: registry,
		appStart: time.Now(),
	}
}

// GetHealthStatus assembles and returns a comprehensive HealthSummary.
// Called periodically by the frontend's useWidgetRefresh hook.
func (h *Health) GetHealthStatus() HealthSummary {
	defer common.RecoverPanic()
	summary := HealthSummary{
		Uptime:    common.FormatUptime(uint64(time.Since(h.appStart).Seconds())),
		Alerts:    h.alerts.AlertCount(),
	}

	// ── CPU ──
	cpuPct, _ := cpu.Percent(0, false)
	cpuVal := 0.0
	if len(cpuPct) > 0 {
		cpuVal = cpuPct[0]
	}
	cpuMetric := h.buildMetric("CPU", cpuVal, "%", 80, 92)
	summary.CPU = cpuMetric

	// ── Memory ──
	memInfo, _ := mem.VirtualMemory()
	memVal := 0.0
	if memInfo != nil {
		memVal = memInfo.UsedPercent
	}
	memMetric := h.buildMetric("Memory", memVal, "%", 85, 94)
	summary.Memory = memMetric

	// ── Disk ──
	diskPct, _ := disk.Usage("/")
	diskVal := 0.0
	if diskPct != nil {
		diskVal = diskPct.UsedPercent
	}
	diskMetric := h.buildMetric("Disk", diskVal, "%", 85, 95)
	summary.Disk = diskMetric

	// ── Load ──
	loadAvg, _ := load.Avg()
	loadVal := 0.0
	if loadAvg != nil {
		loadVal = loadAvg.Load1
	}
	loadMetric := h.buildMetric("Load (1m)", loadVal, "", 0, 0)
	loadMetric.Message = fmt.Sprintf("%.2f / %.2f / %.2f", loadAvg.Load1, loadAvg.Load5, loadAvg.Load15)
	summary.Load = loadMetric

	// ── Collector health ──
	if h.registry != nil {
		statuses := h.registry.Snapshot()
		summary.Collectors = make([]CollectorHealth, 0, len(statuses))
		for _, cs := range statuses {
			ch := CollectorHealth{
				ID:         string(cs.ID),
				Name:       cs.Name,
				Enabled:    cs.Enabled,
				LastRun:    cs.LastRun,
				IntervalMs: cs.IntervalMs,
			}
			if !cs.Enabled {
				ch.Status = "stale"
			} else if cs.LastRun == "" {
				ch.Status = "critical"
				ch.Error = "collector has never run"
			} else {
				// Healthy if last run is within 2x the interval
				lastRunAt, err := time.Parse(time.RFC3339, cs.LastRun)
				if err != nil || time.Since(lastRunAt) > time.Duration(cs.IntervalMs*2)*time.Millisecond {
					ch.Status = "degraded"
					ch.Error = "collector data may be stale"
				} else {
					ch.Status = "healthy"
				}
			}
			summary.Collectors = append(summary.Collectors, ch)
		}
	}

	// ── Overall verdict ──
	summary.Overall = h.computeOverall(summary)

	return summary
}

// buildMetric constructs a SystemHealthMetric with a status derived from thresholds.
func (h *Health) buildMetric(name string, value float64, unit string, warnThresh, critThresh float64) SystemHealthMetric {
	m := SystemHealthMetric{
		Name:  name,
		Value: value,
		Unit:  unit,
	}
	if critThresh > 0 && value >= critThresh {
		m.Status = "critical"
		m.Message = fmt.Sprintf("%.1f%s — exceeds critical threshold (%.0f%s)", value, unit, critThresh, unit)
	} else if warnThresh > 0 && value >= warnThresh {
		m.Status = "degraded"
		m.Message = fmt.Sprintf("%.1f%s — exceeds warning threshold (%.0f%s)", value, unit, warnThresh, unit)
	} else {
		m.Status = "healthy"
		m.Message = fmt.Sprintf("%.1f%s — within normal range", value, unit)
	}
	return m
}

// computeOverall derives the top-level health status from sub-metrics.
func (h *Health) computeOverall(s HealthSummary) string {
	// Check criticals first
	for _, m := range []SystemHealthMetric{s.CPU, s.Memory, s.Disk} {
		if m.Status == "critical" {
			return "critical"
		}
	}

	// Check degraded
	for _, m := range []SystemHealthMetric{s.CPU, s.Memory, s.Disk} {
		if m.Status == "degraded" {
			return "degraded"
		}
	}

	// Check collector health
	for _, c := range s.Collectors {
		if c.Status == "critical" {
			return "critical"
		}
		if c.Status == "degraded" {
			return "degraded"
		}
	}

	// Check alerts
	if s.Alerts > 0 {
		return "degraded"
	}

	return "healthy"
}
