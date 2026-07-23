package common

import (
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

// WorkflowInvoker defines an interface for triggering operational workflows.
type WorkflowInvoker interface {
	TriggerWorkflow(id string) (string, error)
}

// EngineLoop handles the periodic evaluation of system health, alerts, and metrics.
type EngineLoop struct {
	pipeline *DataPipeline
	alerts   *AlertEngine
	eventBus *EventBus

	// New: Baseline drift detection
	baselines *BaselinesEngine

	// New: Autonomous workflow invocation
	invoker WorkflowInvoker

	// Internal state for diagnostic isolation
	auditRunning   bool
	auditMu        sync.Mutex
	lastAuditTime  time.Time
	auditCooldown  time.Duration

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

func NewEngineLoop(p *DataPipeline, a *AlertEngine, eb *EventBus, sl *sync.RWMutex, invoker WorkflowInvoker) *EngineLoop {
	return &EngineLoop{
		pipeline:    p,
		alerts:      a,
		eventBus:    eb,
		storageLock: sl,
		baselines:   NewBaselinesEngine(p),
		invoker:     invoker,
		quit:        make(chan struct{}),
		auditCooldown: 5 * time.Minute, // Default 5 min cooldown for autonomous audits
	}
}

func (e *EngineLoop) Start(interval time.Duration) {
	e.wg.Add(1)
	go func() {
		defer RecoverPanic()
		defer e.wg.Done()

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		// Recalculate baselines every hour
		baseTicker := time.NewTicker(1 * time.Hour)
		defer baseTicker.Stop()

		for {
			select {
			case <-e.quit:
				return
			case <-ticker.C:
				e.Step()
			case <-baseTicker.C:
				e.baselines.RecalculateBaselines()
				e.DailyAnalysis()
			}
		}
	}()
}

// DailyAnalysis calculates the persistent health scorecard.
func (e *EngineLoop) DailyAnalysis() {
	s := GetStorage()
	if s == nil {
		return
	}

	score := 100
	metrics := []string{MetricCPU, MetricMem, MetricDisk}

	for _, m := range metrics {
		if drift, ok := e.baselines.DetectDrift(m); ok {
			if drift.Severity == "high" {
				score -= 15
			} else if drift.Severity == "med" {
				score -= 5
			}
		}
	}

	if score < 0 { score = 0 }
	_ = s.UpsertHealthScore(score)
}

func (e *EngineLoop) Stop() {
	close(e.quit)
	e.wg.Wait()
}

// Step performs a single evaluation and emission cycle.
func (e *EngineLoop) Step() {
	defer RecoverPanic()

	var newAlerts []Alert
	var snapshot MetricSnapshot

	// 1. Capture Data (Lane 1 & 2)
	{
		// We only hold the storage lock while reading metrics and evaluating alerts
		// to prevent blocking the rest of the 3s tick logic.
		if e.storageLock != nil {
			e.storageLock.RLock()
			defer e.storageLock.RUnlock()
		}

		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer RecoverPanic()
			defer wg.Done()
			newAlerts = e.alerts.Evaluate()
		}()

		go func() {
			defer RecoverPanic()
			defer wg.Done()
			snapshot = e.CaptureSnapshot()
		}()

		wg.Wait()
	}

	// 2. Update External State (Prometheus, etc)
	e.UpdatePrometheus(snapshot)

	// 3. Detect Spikes & Drift
	e.DetectSpikes(snapshot)
	e.DetectDrift()

	// 4. Notify Callbacks
	if e.OnMetricsEmit != nil {
		e.OnMetricsEmit(snapshot)
	}
	if len(newAlerts) > 0 {
		// CORRELATION: Check for multi-subsystem incidents
		if inc := e.alerts.CorrelateAlerts(newAlerts); inc != nil {
			if e.eventBus != nil {
				e.eventBus.Emit(NewEventWithMeta(
					CatAlert, EventCritical, "engine",
					inc.Title,
					fmt.Sprintf("Correlated breach detected across %d metrics: %v", len(inc.Metrics), inc.Metrics),
					map[string]string{"incident_id": inc.ID},
				))
			}
		}

		if e.OnAlertsEmit != nil {
			e.OnAlertsEmit(newAlerts)
		}
	}
}

func (e *EngineLoop) DetectDrift() {
	metrics := []string{MetricCPU, MetricMem, MetricDisk}
	for _, m := range metrics {
		if drift, ok := e.baselines.DetectDrift(m); ok {
			if e.eventBus != nil {
				e.eventBus.Emit(NewEventWithMeta(
					CatAlert, EventWarning, "engine",
					fmt.Sprintf("%s baseline drift detected", m),
					fmt.Sprintf("Current average (%.1f) is %.1fσ away from persistent baseline (%.1f).",
						drift.Current, drift.Deviation, drift.Baseline),
					map[string]string{
						"metric":    drift.Metric,
						"baseline":  fmt.Sprintf("%.1f", drift.Baseline),
						"deviation": fmt.Sprintf("%.1f", drift.Deviation),
						"severity":  drift.Severity,
					},
				))
			}
		}
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
		e.autonomousAudit("diag-slow-pc", "CPU Spike")
	}
	if s.Memory > prevMem+10 && prevMem > 0 {
		e.emitSpike("Memory pressure increasing", fmt.Sprintf("Memory usage increased from %.0f%% to %.0f%%", prevMem, s.Memory), "mem")
		e.autonomousAudit("diag-slow-pc", "Memory Pressure")
	}
	if s.Disk > prevDisk+10 && prevDisk > 0 {
		e.emitSpike("Disk usage increasing", fmt.Sprintf("Disk usage increased from %.0f%% to %.0f%%", prevDisk, s.Disk), "disk")
	}
}

func (e *EngineLoop) autonomousAudit(workflowID, reason string) {
	if e.invoker == nil {
		return
	}

	e.auditMu.Lock()
	if e.auditRunning || time.Since(e.lastAuditTime) < e.auditCooldown {
		e.auditMu.Unlock()
		return
	}
	e.auditRunning = true
	e.lastAuditTime = time.Now()
	e.auditMu.Unlock()

	go func() {
		defer RecoverPanic()
		defer func() {
			e.auditMu.Lock()
			e.auditRunning = false
			e.auditMu.Unlock()
		}()

		reportID, err := e.invoker.TriggerWorkflow(workflowID)
		if err == nil && e.eventBus != nil {
			e.eventBus.Emit(NewEventWithMeta(
				CatAlert, EventInfo, "engine",
				"Autonomous Diagnostic Complete",
				fmt.Sprintf("Hawk automatically executed '%s' due to %s.", workflowID, reason),
				map[string]string{"report_id": reportID, "workflow_id": workflowID},
			))
		}
	}()
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
