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

// ReportTrigger defines an interface for triggering report generation
// from alert rule conditions or scheduled cycles.
type ReportTrigger interface {
	// TriggerReport generates a report of the given type and returns its ID.
	TriggerReport(reportType string) (string, error)
	// GetEnabledReportRules returns all enabled auto-report rules.
	GetEnabledReportRules() ([]AutoReportRule, error)
	// TouchRule updates the last-triggered timestamp for a rule.
	TouchRule(ruleID string) error
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

	// New: Report rule auto-generation
	reportTrigger ReportTrigger

	// Internal state for diagnostic isolation
	auditRunning  bool
	auditMu       sync.Mutex
	lastAuditTime time.Time
	auditCooldown time.Duration

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
		pipeline:      p,
		alerts:        a,
		eventBus:      eb,
		storageLock:   sl,
		baselines:     NewBaselinesEngine(p),
		invoker:       invoker,
		quit:          make(chan struct{}),
		auditCooldown: 5 * time.Minute, // Default 5 min cooldown for autonomous audits
	}
}

// SetReportTrigger attaches the report rule evaluator to the engine loop.
func (e *EngineLoop) SetReportTrigger(rt ReportTrigger) {
	e.reportTrigger = rt
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

		// Write initial health score at startup (not just on 1h ticker)
		e.DailyAnalysis()
		e.checkScheduledReports()

		for {
			select {
			case <-e.quit:
				return
			case <-ticker.C:
				e.Step()
			case <-baseTicker.C:
				e.baselines.RecalculateBaselines()
				e.DailyAnalysis()
				e.checkScheduledReports()
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

	if score < 0 {
		score = 0
	}
	if err := s.UpsertHealthScore(score); err != nil {
		LogError("EngineLoop: failed to upsert health score %d: %v", score, err)
	}
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
	// Storage lock is held ONLY during the parallel capture phase, then released
	// immediately. The anonymous function ensures defer fires at block end.
	func() {
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
	}()

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

		// 5. Check report rules for on_alert triggers
		e.checkReportRulesForAlerts(newAlerts)
	}
}

// checkScheduledReports triggers reports for rules with hourly/daily schedules.
func (e *EngineLoop) checkScheduledReports() {
	if e.reportTrigger == nil {
		return
	}

	rules, err := e.reportTrigger.GetEnabledReportRules()
	if err != nil || len(rules) == 0 {
		return
	}

	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}
		// on_alert is handled in checkReportRulesForAlerts
		if rule.Schedule != "hourly" && rule.Schedule != "daily" {
			continue
		}

		if !e.canTriggerReport(rule) {
			continue
		}

		ruleID := rule.ID
		go func(rt ReportTrigger, rid, rtype string) {
			defer RecoverPanic()
			reportID, err := rt.TriggerReport(rtype)
			if err != nil {
				LogError("EngineLoop: scheduled report %s failed for rule %s: %v", rtype, rid, err)
				return
			}
			_ = rt.TouchRule(rid)
			LogInfo("EngineLoop: scheduled report %s generated (id=%s) from rule %s (schedule=%s)", rtype, reportID, rid, rule.Schedule)
		}(e.reportTrigger, ruleID, rule.ReportType)
	}
}

// checkReportRulesForAlerts evaluates enabled report rules against fired alerts
// and auto-generates reports for matching rules.
func (e *EngineLoop) checkReportRulesForAlerts(newAlerts []Alert) {
	if e.reportTrigger == nil {
		return
	}

	rules, err := e.reportTrigger.GetEnabledReportRules()
	if err != nil || len(rules) == 0 {
		return
	}

	for _, rule := range rules {
		if rule.Schedule != "on_alert" || !rule.Enabled {
			continue
		}

		// Check if any fired alert matches this rule's metric + condition
		for _, alert := range newAlerts {
			if alert.Metric != rule.Metric {
				continue
			}
			matches := false
			switch rule.Condition {
			case "GT":
				matches = alert.Value > rule.Threshold
			case "LT":
				matches = alert.Value < rule.Threshold
			}
			if !matches {
				continue
			}

			// Rule matched — check cooldown
			if !e.canTriggerReport(rule) {
				continue
			}

			ruleID := rule.ID
			go func(rt ReportTrigger, rid, rtype string) {
				defer RecoverPanic()
				reportID, err := rt.TriggerReport(rtype)
				if err != nil {
					LogError("EngineLoop: auto-report %s failed for rule %s: %v", rtype, rid, err)
					return
				}
				_ = rt.TouchRule(rid)
				LogInfo("EngineLoop: auto-report %s generated (id=%s) from rule %s", rtype, reportID, rid)
			}(e.reportTrigger, ruleID, rule.ReportType)
			break // one report per rule per cycle
		}
	}
}

// canTriggerReport checks if enough time has elapsed since the rule was last triggered.
func (e *EngineLoop) canTriggerReport(rule AutoReportRule) bool {
	if rule.LastTriggeredAt == nil || *rule.LastTriggeredAt == "" {
		return true
	}

	last, err := time.Parse(time.RFC3339, *rule.LastTriggeredAt)
	if err != nil {
		// Fallback for non-RFC formats if any
		last, err = time.Parse("2006-01-02 15:04:05", *rule.LastTriggeredAt)
		if err != nil {
			return true // cannot parse? let it trigger
		}
	}

	elapsed := time.Since(last)

	switch rule.Schedule {
	case "on_alert":
		return elapsed > 30*time.Minute
	case "hourly":
		return elapsed > 55*time.Minute
	case "daily":
		return elapsed > 23*time.Hour
	default:
		return true
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

// UpdatePrometheus is deliberately a no-op. Prometheus metrics were removed
// to align with the 100% local / zero telemetry north star. If you need
// Prometheus-style instrumentation in the future, wire it here behind a
// //go:build prometheus tag and add prometheus/client_golang back to go.mod.
func (e *EngineLoop) UpdatePrometheus(_ MetricSnapshot) {}

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
