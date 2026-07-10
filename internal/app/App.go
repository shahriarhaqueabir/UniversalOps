package app

import (
	"context"
	"strings"

	"sync"
	"time"

	goruntime "runtime"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
)

// App is the main application struct bound to the Wails frontend.
// All exported methods are callable from JavaScript.
type App struct {
	ctx       context.Context
	startedAt time.Time

	pipeline *common.DataPipeline
	alerts   *common.AlertEngine

	// Ops layer facades
	SysOps      *SysOps
	NetOps      *NetOps
	SecOps      *SecOps
	DevOps      *DevOps
	AIOps       *AIOps
	Dash        *Dashboard
	PipelineAPI *PipelineAPI
	AlertAPI    *AlertAPI
	Logs        *Logs

	// Tick loop control
	tickQuit       chan struct{}
	tickIntervalCh chan time.Duration
	tickWg         sync.WaitGroup
}

// NewApp creates a new App with initialized subsystems.
func NewApp() *App {
	cfg := common.DefaultCollectionConfig()
	pipeline := common.NewDataPipeline(cfg)
	alertEngine := common.NewAlertEngine(pipeline)
	alertEngine.AddDefaultRules()

	a := &App{
		pipeline:       pipeline,
		alerts:         alertEngine,
		startedAt:      time.Now(),
		tickQuit:       make(chan struct{}),
		tickIntervalCh: make(chan time.Duration, 1),
	}

	// Initialize subsystem facades
	a.SysOps = NewSysOps(a)
	a.NetOps = NewNetOps(a)
	a.SecOps = NewSecOps(a)
	a.DevOps = NewDevOps(a)
	a.AIOps = NewAIOps(a)
	a.Dash = NewDashboard(a)
	a.PipelineAPI = NewPipelineAPI(a)
	a.AlertAPI = NewAlertAPI(a)
	a.Logs = NewLogs(a)

	return a
}

// Startup is called by Wails when the application starts.
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx

	// Initialize persistent storage
	if err := common.InitStorage("hawkward.db"); err != nil {
		common.LogWarn("Failed to init storage: %v", err)
	}

	// Initialize the common logger
	if err := common.InitLogger("hawkward-gui.log"); err != nil {
		common.LogWarn("Failed to init logger: %v", err)
	}

	common.LogInfo("Hawkward GUI starting up")

	// Initialize Prometheus metrics exporter
	common.InitMetricsExporter(0)

	// Start the metrics collection tick loop
	a.startTickLoop()
}

// Shutdown is called by Wails when the application shuts down.
func (a *App) Shutdown(ctx context.Context) {
	common.LogInfo("Hawkward GUI shutting down")

	// Stop the tick loop
	close(a.tickQuit)
	a.tickWg.Wait()

	// Close persistent storage
	if s := common.GetStorage(); s != nil {
		s.Close()
	}

	common.CloseLogger()
}

// GetAppInfo returns metadata about the application.
func (a *App) GetAppInfo() AppInfo {
	return AppInfo{
		Name:      "Hawkward Operations Platform",
		Version:   "1.3.0",
		GoVersion: goruntime.Version(),
		Uptime:    common.FormatUptime(uint64(time.Since(a.startedAt).Seconds())),
	}
}

// ── Dialogs ─────────────────────────────────────────────────────────────────

// OpenFileDialog shows a file open dialog and returns the selected path.
func (a *App) OpenFileDialog(title string, filters []string) (string, error) {
	var wailsFilters []runtime.FileFilter
	for _, f := range filters {
		// Expecting "Name|*.ext"
		parts := strings.Split(f, "|")
		if len(parts) == 2 {
			wailsFilters = append(wailsFilters, runtime.FileFilter{
				DisplayName: parts[0],
				Pattern:     parts[1],
			})
		}
	}

	return runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title:   title,
		Filters: wailsFilters,
	})
}

// SaveFileDialog shows a file save dialog and returns the selected path.
func (a *App) SaveFileDialog(title string, filename string, filters []string) (string, error) {
	var wailsFilters []runtime.FileFilter
	for _, f := range filters {
		parts := strings.Split(f, "|")
		if len(parts) == 2 {
			wailsFilters = append(wailsFilters, runtime.FileFilter{
				DisplayName: parts[0],
				Pattern:     parts[1],
			})
		}
	}

	return runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           title,
		DefaultFilename: filename,
		Filters:         wailsFilters,
	})
}

// ── Internal helpers ─────────────────────────────────────────────────────────

// startTickLoop begins the periodic metrics collection goroutine.
func (a *App) startTickLoop() {
	a.tickWg.Add(1)
	go func() {
		defer a.tickWg.Done()
		interval := 3 * time.Second
		if a.pipeline != nil {
			interval = a.pipeline.Config().TickInterval
		}
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-a.tickQuit:
				return
			case newInterval := <-a.tickIntervalCh:
				ticker.Stop()
				ticker = time.NewTicker(newInterval)
			case <-ticker.C:
				a.collectAndEmit()
			}
		}
	}()
}

// collectAndEmit collects all stats, pushes to pipeline, evaluates alerts,
// and emits events to the frontend.
func (a *App) collectAndEmit() {
	if a.ctx == nil {
		return
	}
	if a.ctx.Err() != nil {
		return
	}

	// 1. Collect ALL raw values FIRST — same timestamp window
	stats, err := a.SysOps.collector.CollectAllStats()
	if err != nil {
		common.LogWarn("CollectAllStats failed: %v", err)
		return
	}

	// Collect network rates at the same time
	var rxTotal, txTotal float64
	ifaces, err := a.NetOps.collectInterfaces()
	if err == nil && len(ifaces) > 0 {
		for _, iface := range ifaces {
			rxTotal += iface.RXRateBps
			txTotal += iface.TXRateBps
		}
	}

	// 2. Push ALL metrics to pipeline (fast, no blocking I/O)
	if stats != nil {
		a.pipeline.PushMetric(common.MetricCPU, "%", stats.CPUPercent)
		a.pipeline.PushMetric(common.MetricMem, "%", stats.MemoryUsed)
		a.pipeline.PushMetric(common.MetricDisk, "%", stats.DiskUsed)
		a.pipeline.PushMetric(common.MetricProcCnt, "count", float64(stats.ProcessCount))
		a.pipeline.PushMetric(common.MetricNetRX, "bps", rxTotal)
		a.pipeline.PushMetric(common.MetricNetTX, "bps", txTotal)
	}

	// Evaluate alerts
	newAlerts := a.alerts.Evaluate()

	// Get current values for the dashboard event
	cpuMF := a.pipeline.GetMetricWithForecast(common.MetricCPU)
	memMF := a.pipeline.GetMetricWithForecast(common.MetricMem)
	diskMF := a.pipeline.GetMetricWithForecast(common.MetricDisk)
	netRXMF := a.pipeline.GetMetricWithForecast(common.MetricNetRX)
	netTXMF := a.pipeline.GetMetricWithForecast(common.MetricNetTX)
	procMF := a.pipeline.GetMetricWithForecast(common.MetricProcCnt)

	metricsEvent := MetricsEvent{
		CPU: GaugeMetric{
			Value:    cpuMF.LastValue,
			Unit:     cpuMF.Unit,
			History:  safeLastN(cpuMF.Values, 60),
			Forecast: cpuMF.Forecast,
			Trend:    trendDirectionString(cpuMF.Trend.Direction),
		},
		Memory: GaugeMetric{
			Value:    memMF.LastValue,
			Unit:     memMF.Unit,
			History:  safeLastN(memMF.Values, 60),
			Forecast: memMF.Forecast,
			Trend:    trendDirectionString(memMF.Trend.Direction),
		},
		Disk: GaugeMetric{
			Value:    diskMF.LastValue,
			Unit:     diskMF.Unit,
			History:  safeLastN(diskMF.Values, 60),
			Forecast: diskMF.Forecast,
			Trend:    trendDirectionString(diskMF.Trend.Direction),
		},
		Network: NetworkMetric{
			RXRate: lastValue(netRXMF.Values),
			TXRate: lastValue(netTXMF.Values),
			Unit:   "bps",
		},
		Processes:   int(procMF.LastValue),
		Connections: 0, // fetched on-demand
	}

	// Update Prometheus metrics
	common.SetCPUMetric(cpuMF.LastValue)
	common.SetMemoryMetric(memMF.LastValue)
	common.SetDiskMetric(diskMF.LastValue)
	common.SetProcessCountMetric(procMF.LastValue)
	common.SetAlertCountMetric(float64(a.alerts.AlertCount()))
	common.IncPipelineTick()

	// Emit metrics event
	runtime.EventsEmit(a.ctx, EventMetrics, metricsEvent)

	// Emit alert events for newly fired alerts
	for _, alert := range newAlerts {
		runtime.EventsEmit(a.ctx, EventAlert, AlertEvent{
			Action:     "fired",
			Alert:      convertAlert(alert),
			AlertCount: a.alerts.AlertCount(),
		})
	}
}

// ── Utility functions ────────────────────────────────────────────────────────

func safeLastN(values []float64, n int) []float64 {
	if len(values) <= n {
		return values
	}
	return values[len(values)-n:]
}

func lastValue(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	return values[len(values)-1]
}

func trendDirectionString(dir common.TrendDirection) string {
	switch dir {
	case common.TrendRising:
		return "rising"
	case common.TrendFalling:
		return "falling"
	default:
		return "stable"
	}
}

func convertAlert(a common.Alert) AlertInfo {
	return AlertInfo{
		ID:        a.ID,
		Level:     a.Level.String(),
		Metric:    a.Metric,
		Message:   a.Message,
		Value:     a.Value,
		Threshold: a.Threshold,
		Timestamp: a.Timestamp.Format(time.RFC3339),
		Resolved:  a.Resolved,
	}
}
