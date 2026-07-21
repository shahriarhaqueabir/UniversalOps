package common

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// MetricSnapshot represents a unified delta of all core system metrics.
type MetricSnapshot struct {
	CPU         float64
	Memory      float64
	Disk        float64
	Processes   int
	Connections int
	RXRate      float64
	TXRate      float64
	Timestamp   time.Time
}

// EngineLoop handles the periodic evaluation of system health, alerts, and metrics.
// It is decoupled from the UI framework (Wails) and can run in headless mode.
type EngineLoop struct {
	pipeline *DataPipeline
	alerts   *AlertEngine
	eventBus *EventBus

	// Callbacks for UI/External notification
	OnMetricsEmit func(snapshot MetricSnapshot)
	OnAlertsEmit  func(newAlerts []Alert)

	// Internal state for spike detection
	lastMu                     sync.Mutex
	lastCPU, lastMem, lastDisk float64

	quit chan struct{}
	wg   sync.WaitGroup

	storageLock *sync.RWMutex
}

func NewEngineLoop(p *DataPipeline, a *AlertEngine, eb *EventBus, sl *sync.RWMutex) *EngineLoop {
	return &EngineLoop{
		pipeline:    p,
		alerts:      a,
		eventBus:    eb,
		storageLock: sl,
		quit:        make(chan struct{}),
	}
}

func (e *EngineLoop) Start(interval time.Duration) {
	e.wg.Add(1)
	go func() {
		defer RecoverPanic()
		defer e.wg.Done()

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-e.quit:
				return
			case <-ticker.C:
				e.Step()
			}
		}
	}()
}

func (e *EngineLoop) Stop() {
	close(e.quit)
	e.wg.Wait()
}

// Step performs a single evaluation and emission cycle.
func (e *EngineLoop) Step() {
	defer RecoverPanic()

	// 1. Lock for storage safety
	if e.storageLock != nil {
		e.storageLock.RLock()
		defer e.storageLock.RUnlock()
	}

	// 2. Evaluate Alerts
	newAlerts := e.alerts.Evaluate()

	// 3. Capture Snapshot
	snapshot := e.CaptureSnapshot()

	// 4. Update External State (Prometheus, etc)
	e.UpdatePrometheus(snapshot)

	// 5. Detect Spikes
	e.DetectSpikes(snapshot)

	// 6. Notify Callbacks
	if e.OnMetricsEmit != nil {
		e.OnMetricsEmit(snapshot)
	}
	if len(newAlerts) > 0 && e.OnAlertsEmit != nil {
		e.OnAlertsEmit(newAlerts)
	}
}

func (e *EngineLoop) CaptureSnapshot() MetricSnapshot {
	cpu := e.pipeline.GetMetricWithForecast(MetricCPU).LastValue
	mem := e.pipeline.GetMetricWithForecast(MetricMem).LastValue
	disk := e.pipeline.GetMetricWithForecast(MetricDisk).LastValue
	proc := e.pipeline.GetMetricWithForecast(MetricProcCnt).LastValue
	rx := e.pipeline.GetMetricWithForecast(MetricNetRX).LastValue
	tx := e.pipeline.GetMetricWithForecast(MetricNetTX).LastValue

	return MetricSnapshot{
		CPU:       cpu,
		Memory:    mem,
		Disk:      disk,
		Processes: int(proc),
		RXRate:    rx,
		TXRate:    tx,
		Timestamp: time.Now(),
	}
}

func (e *EngineLoop) UpdatePrometheus(s MetricSnapshot) {
	SetCPUMetric(s.CPU)
	SetMemoryMetric(s.Memory)
	SetDiskMetric(s.Disk)
	SetProcessCountMetric(float64(s.Processes))
	SetAlertCountMetric(float64(e.alerts.AlertCount()))
	IncPipelineTick()
}

func (e *EngineLoop) DetectSpikes(s MetricSnapshot) {
	e.lastMu.Lock()
	prevCPU, prevMem, prevDisk := e.lastCPU, e.lastMem, e.lastDisk
	e.lastCPU = s.CPU
	e.lastMem = s.Memory
	e.lastDisk = s.Disk
	e.lastMu.Unlock()

	if s.CPU > prevCPU+15 && prevCPU > 0 {
		e.emitSpike("CPU spiked", fmt.Sprintf("CPU usage jumped from %.0f%% to %.0f%%", prevCPU, s.CPU), "cpu")
	}
	if s.Memory > prevMem+10 && prevMem > 0 {
		e.emitSpike("Memory pressure increasing", fmt.Sprintf("Memory usage increased from %.0f%% to %.0f%%", prevMem, s.Memory), "mem")
	}
	if s.Disk > prevDisk+10 && prevDisk > 0 {
		e.emitSpike("Disk usage increasing", fmt.Sprintf("Disk usage increased from %.0f%% to %.0f%%", prevDisk, s.Disk), "disk")
	}
}

func (e *EngineLoop) emitSpike(title, detail, metric string) {
	if e.eventBus == nil {
		return
	}
	e.eventBus.Emit(NewEventWithMeta(
		CatSystem, EventWarning, "sysops",
		title, detail,
		map[string]string{"metric": metric},
	))
}
