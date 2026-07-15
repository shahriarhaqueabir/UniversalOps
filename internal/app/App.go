package app

import (
	"context"
	"fmt"
	"net/url"
	"os"
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
	eventBus *common.EventBus

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
	Timeline    *Timeline
	NetDesign   *NetworkDesignAPI

	// Collector architecture
	collectorRegistry *common.CollectorRegistry
	scheduler         *common.CollectorScheduler

	// Alert and dashboard evaluation loop
	alertQuit chan struct{}
	alertWg   sync.WaitGroup

	// Previous metric values for significant-change detection.
	// Guarded by lastMu since evaluateAndEmit() can run concurrently from
	// both the periodic alert-loop ticker and TriggerCollector() (a Wails
	// binding invoked directly from the frontend goroutine). See IPC-3.
	lastMu                     sync.Mutex
	lastCPU, lastMem, lastDisk float64
}

// NewApp creates a new App with initialized subsystems.
func NewApp() *App {
	cfg := common.DefaultCollectionConfig()
	pipeline := common.NewDataPipeline(cfg)
	alertEngine := common.NewAlertEngine(pipeline)
	alertEngine.AddDefaultRules()

	a := &App{
		pipeline:  pipeline,
		alerts:    alertEngine,
		eventBus:  common.NewEventBus(1000),
		startedAt: time.Now(),
		alertQuit: make(chan struct{}),
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
	a.Timeline = NewTimeline(a)
	a.NetDesign = NewNetworkDesignAPI(a)

	// Subscribe the event bus to persist events and emit to frontend
	a.eventBus.Subscribe(func(evt common.TimelineEvent) {
		// Persist to database
		if s := common.GetStorage(); s != nil {
			s.InsertEvent(evt)
		}

		// Emit to frontend via Wails runtime
		if a.ctx != nil {
			converted := convertTimelineEvent(evt)
			runtime.EventsEmit(a.ctx, EventTimeline, converted)
		}
	})

	return a
}

// Startup is called by Wails when the application starts.
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx

	// Initialize persistent storage
	if err := common.InitStorage("opsforall.db"); err != nil {
		common.LogWarn("Failed to init storage: %v", err)
	}

	// Initialize the common logger
	if err := common.InitLogger("opsforall.log"); err != nil {
		common.LogWarn("Failed to init logger: %v", err)
	}

	common.LogInfo("OpsForAll starting up")

	// Validate Ollama environment variables
	validateOllamaEnv()

	// Initialize Prometheus metrics exporter
	common.InitMetricsExporter(0)

	// Start the modular collector system
	a.collectorRegistry = common.NewCollectorRegistry()
	RegisterCollectors(a.collectorRegistry, a)
	a.scheduler = common.NewCollectorScheduler(a.collectorRegistry, a.pipeline)
	a.scheduler.Start()

	// Start the alert and dashboard evaluation loop
	a.startAlertLoop()
}

// Shutdown is called by Wails when the application shuts down.
func (a *App) Shutdown(ctx context.Context) {
	common.LogInfo("OpsForAll shutting down")

	// Stop the collector scheduler
	if a.scheduler != nil {
		a.scheduler.Stop(5 * time.Second)
	}

	// Stop the alert evaluation loop
	close(a.alertQuit)
	alertDone := make(chan struct{})
	go func() {
		a.alertWg.Wait()
		close(alertDone)
	}()
	select {
	case <-alertDone:
	case <-time.After(5 * time.Second):
		common.LogWarn("Alert loop did not shut down within 5s")
	}

	// Close persistent storage
	if s := common.GetStorage(); s != nil {
		s.Close()
	}

	common.CloseLogger()
}

// GetAppInfo returns metadata about the application.
func (a *App) GetAppInfo() AppInfo {
	return AppInfo{
		Name:      "OpsForAll Universal Platform",
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

// startAlertLoop begins the periodic alert evaluation and dashboard event
// emission goroutine. It does NOT collect data — collectors run independently
// via the CollectorScheduler.
func (a *App) startAlertLoop() {
	a.alertWg.Add(1)
	go func() {
		defer common.RecoverPanic()
		defer a.alertWg.Done()
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-a.alertQuit:
				return
			case <-ticker.C:
				a.evaluateAndEmit()
			}
		}
	}()
}

// evaluateAndEmit reads metrics from the pipeline, evaluates alert rules,
// and emits dashboard + alert events to the frontend.
// Data collection is handled separately by the CollectorScheduler.
func (a *App) evaluateAndEmit() {
	if a.ctx == nil || a.ctx.Err() != nil {
		return
	}

	// Snapshot active alert IDs before evaluation (for resolve detection)
	prevActive := make(map[string]bool)
	for _, al := range a.alerts.ActiveAlerts() {
		prevActive[al.ID] = true
	}

	// Evaluate alerts
	newAlerts := a.alerts.Evaluate()

	// Persist fired alerts to SQLite
	if store := common.GetStorage(); store != nil {
		for _, alert := range newAlerts {
			store.InsertAlert(common.AlertRecord{
				ID:        alert.ID,
				Timestamp: alert.Timestamp,
				Level:     alert.Level.String(),
				Metric:    alert.Metric,
				Message:   alert.Message,
				Value:     alert.Value,
				Threshold: alert.Threshold,
			})
		}

		for id := range prevActive {
			found := false
			for _, al := range a.alerts.ActiveAlerts() {
				if al.ID == id {
					found = true
					break
				}
			}
			if !found {
				_ = store.UpdateAlertResolved(id)
			}
		}
	}

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

	// Detect significant changes and emit timeline events.
	// lastCPU/lastMem/lastDisk are shared with concurrent callers (the
	// alert-loop ticker and TriggerCollector), so the read-compare-write
	// sequence is done under lastMu to avoid a data race (IPC-3).
	currCPU := cpuMF.LastValue
	currMem := memMF.LastValue
	currDisk := diskMF.LastValue

	a.lastMu.Lock()
	prevCPU, prevMem, prevDisk := a.lastCPU, a.lastMem, a.lastDisk
	a.lastCPU = currCPU
	a.lastMem = currMem
	a.lastDisk = currDisk
	a.lastMu.Unlock()

	if currCPU > prevCPU+15 && prevCPU > 0 {
		a.eventBus.Emit(common.NewEventWithMeta(
			common.CatSystem, common.EventWarning, "sysops",
			"CPU spiked",
			fmt.Sprintf("CPU usage jumped from %.0f%% to %.0f%%", prevCPU, currCPU),
			map[string]string{"from": fmt.Sprintf("%.1f", prevCPU), "to": fmt.Sprintf("%.1f", currCPU)},
		))
	}
	if currMem > prevMem+10 && prevMem > 0 {
		a.eventBus.Emit(common.NewEventWithMeta(
			common.CatSystem, common.EventWarning, "sysops",
			"Memory pressure increasing",
			fmt.Sprintf("Memory usage increased from %.0f%% to %.0f%%", prevMem, currMem),
			map[string]string{"from": fmt.Sprintf("%.1f", prevMem), "to": fmt.Sprintf("%.1f", currMem)},
		))
	}
	if currDisk > prevDisk+10 && prevDisk > 0 {
		a.eventBus.Emit(common.NewEventWithMeta(
			common.CatSystem, common.EventWarning, "sysops",
			"Disk usage increasing",
			fmt.Sprintf("Disk usage increased from %.0f%% to %.0f%%", prevDisk, currDisk),
			map[string]string{"from": fmt.Sprintf("%.1f", prevDisk), "to": fmt.Sprintf("%.1f", currDisk)},
		))
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

		a.eventBus.Emit(common.NewEventWithMeta(
			common.CatAlert,
			common.EventCritical,
			"alerts",
			alert.Metric+" alert fired",
			alert.Message,
			map[string]string{"threshold": fmt.Sprintf("%.1f", alert.Threshold), "value": fmt.Sprintf("%.1f", alert.Value)},
		))
	}
}

// ── Collector Bindings ────────────────────────────────────────────────────────

// ListCollectors returns the status of all registered collectors.
func (a *App) ListCollectors() []common.CollectorStatus {
	if a.collectorRegistry == nil {
		return nil
	}
	return a.collectorRegistry.Snapshot()
}

// SetCollectorEnabled enables or disables a collector by ID.
func (a *App) SetCollectorEnabled(id string, enabled bool) error {
	reg := a.collectorRegistry
	if reg == nil {
		return fmt.Errorf("collector system not initialized")
	}
	cid := common.CollectorID(id)
	if enabled {
		return reg.Enable(cid)
	}
	return reg.Disable(cid)
}

// SetCollectorInterval sets the interval for a collector in milliseconds.
func (a *App) SetCollectorInterval(id string, intervalMs int) error {
	reg := a.collectorRegistry
	if reg == nil {
		return fmt.Errorf("collector system not initialized")
	}
	if intervalMs <= 0 {
		return fmt.Errorf("interval must be positive")
	}
	return reg.SetInterval(common.CollectorID(id), time.Duration(intervalMs)*time.Millisecond)
}

// TriggerCollector forces an immediate one-shot collection for the given collector.
func (a *App) TriggerCollector(id string) error {
	reg := a.collectorRegistry
	if reg == nil {
		return fmt.Errorf("collector system not initialized")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	samples, err := reg.CollectNow(ctx, common.CollectorID(id))
	if err != nil {
		return err
	}

	// Push samples to the pipeline
	for _, s := range samples {
		a.pipeline.PushMetric(s.Name, s.Unit, s.Value)
	}

	// Trigger immediate alert evaluation
	a.evaluateAndEmit()

	return nil
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

// validateOllamaEnv checks OLLAMA_HOST and OLLAMA_MODEL at startup.
func validateOllamaEnv() {
	host := os.Getenv("OLLAMA_HOST")
	if host == "" {
		common.LogInfo("OLLAMA_HOST not set, defaulting to http://localhost:11434")
	} else {
		u, err := url.Parse(host)
		if err != nil || u.Scheme == "" || u.Host == "" {
			common.LogWarn("OLLAMA_HOST=%q is not a valid URL, will fall back to default", host)
		} else {
			common.LogInfo("OLLAMA_HOST=%s", host)
		}
	}
	model := os.Getenv("OLLAMA_MODEL")
	if model == "" {
		common.LogInfo("OLLAMA_MODEL not set, defaulting to agentic-coder")
	} else {
		common.LogInfo("OLLAMA_MODEL=%s", model)
	}
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
