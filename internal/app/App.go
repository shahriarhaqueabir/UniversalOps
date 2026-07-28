package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"sync"
	"time"

	goruntime "runtime"

	"github.com/shirou/gopsutil/v4/mem"
	"github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/shahriarhaqueabir/UniversalOps/internal/common"
	"github.com/shahriarhaqueabir/UniversalOps/internal/secops"
	sysopsPkg "github.com/shahriarhaqueabir/UniversalOps/internal/sysops"
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
	Knowledge   *KnowledgeAPI
	Workflows   *WorkflowAPI
	Reports     *ReportsAPI

	// System health facade
	Health *Health

	// SLO/SLI evaluation engine
	sloEngine *common.SLOEngine

	// Workflow engine
	workflowEngine *common.WorkflowEngine

	// Collector architecture
	collectorRegistry *common.CollectorRegistry
	scheduler         *common.CollectorScheduler

	// Capability discovery
	capabilities *common.CapabilityRegistry

	// Unified engine evaluation loop
	engineLoop *common.EngineLoop

	// Previous metric values for significant-change detection.
	lastMu sync.Mutex

	// storageLock protects storage relocation and prevents races with evaluation
	storageLock sync.RWMutex

	// lastSecurityScore is the cached security score to prevent ticker bloat
	lastSecurityScore string

	currentDataDir string
	currentLogsDir string

	// processWorkerQuit signals the background process worker goroutine to stop
	processWorkerQuit chan struct{}
}

// NewApp creates a new App with initialized subsystems.
func NewApp() *App {
	cfg := common.DefaultCollectionConfig()
	pipeline := common.NewDataPipeline(cfg)
	alertEngine := common.NewAlertEngine(pipeline)
	alertEngine.AddDefaultRules()

	eventBus := common.NewEventBus(1000)

	a := &App{
		pipeline:       pipeline,
		alerts:         alertEngine,
		eventBus:       eventBus,
		startedAt:      time.Now(),
		capabilities:   common.NewCapabilityRegistry(),
		workflowEngine: common.NewWorkflowEngine(),
		currentDataDir: "data",
		currentLogsDir: "logs",
	}

	// Initialize the decoupled engine loop with App as WorkflowInvoker
	a.engineLoop = common.NewEngineLoop(pipeline, alertEngine, eventBus, &a.storageLock, a)

	// Initialize subsystems that don't need context yet
	a.SysOps = NewSysOps()
	a.NetOps = NewNetOps(a.eventBus)
	a.SecOps = NewSecOps(a.eventBus)
	a.PipelineAPI = NewPipelineAPI(a.pipeline)
	a.AlertAPI = NewAlertAPI(a.alerts)
	a.Logs = NewLogs()
	a.Knowledge = NewKnowledgeAPI()
	a.Workflows = NewWorkflowAPI(a.workflowEngine, a.SysOps, a.SecOps, a.DevOps, a.AlertAPI)

	// Initialize subsystems that might need context later
	ctx := context.Background()
	a.AIOps = NewAIOps(ctx, a.pipeline, a.Knowledge, a.capabilities, a.PipelineAPI, a.SysOps, a.currentDataDir)
	a.DevOps = NewDevOps(ctx, a.eventBus)
	a.Timeline = NewTimeline(a.eventBus, a.AIOps)
	a.Reports = NewReportsAPI(a.SysOps, a.SecOps, a.AIOps)
	a.Dash = NewDashboard(a.pipeline, a.alerts, a.SysOps, a.NetOps, a.SecOps, a.DevOps, a.AIOps, a.Timeline, nil, func() string {
		return a.GetAppInfo().Uptime
	})

	// Configure engine loop callbacks for UI integration
	a.engineLoop.OnMetricsEmit = func(s common.MetricSnapshot) {
		if a.ctx == nil {
			return
		}
		// Sync Truth to Knowledge Layer (AI context)
		grade := ""
		a.lastMu.Lock()
		grade = a.lastSecurityScore
		a.lastMu.Unlock()

		common.GetKnowledge().UpdateSecurityState(
			grade,
			a.alerts.AlertCount(),
			s.Connections,
			a.GetAppInfo().Uptime,
		)
	}

	a.engineLoop.OnAlertsEmit = func(alerts []common.Alert) {
		if a.ctx == nil {
			return
		}
		a.emitAlertEvents(alerts)
		// Async persistence
		go a.persistAlertsAsync(alerts)
	}

	a.alerts.OnAlertResolved = func(resolved common.Alert) {
		// Persist resolution to DB asynchronously
		go func() {
			defer common.RecoverPanic()
			if s := common.GetStorage(); s != nil {
				if err := s.UpdateAlertResolved(resolved.ID, nil); err != nil {
					common.LogWarn("App: failed to persist alert resolution %s: %v", resolved.ID, err)
				}
			}
		}()
	}

	// Subscribe the event bus to persist events and emit to frontend
	a.eventBus.Subscribe(func(evt common.TimelineEvent) {
		// Persist to database
		if s := common.GetStorage(); s != nil {
			s.InsertEvent(evt, nil)
		}

		// Future: emit to frontend via dedicated subscription channel if needed
		// Currently frontend uses TanStack Query polling — EventsEmit removed
		// to avoid unnecessary IPC overhead.
	})

	// Wire the ReportsAPI as the engine loop's report trigger so alert
	// evaluation and scheduled intervals can auto-generate reports.
	a.engineLoop.SetReportTrigger(a.Reports)

	return a
}

// TriggerWorkflow implements common.WorkflowInvoker.
// It executes a workflow and persists the result, returning the report ID.
func (a *App) TriggerWorkflow(id string) (string, error) {
	wf, err := a.Workflows.ExecuteWorkflow(id)
	if err != nil {
		return "", err
	}

	// Persist to reports table
	storage := common.GetStorage()
	if storage != nil {
		data, _ := json.Marshal(wf)
		reportID := fmt.Sprintf("auto-%s-%d", id, time.Now().Unix())
		err = storage.InsertReport(common.ReportRecord{
			ID:        reportID,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Type:      "auto_diag",
			Score:     0, // Score logic could be added to workflows later
			DataJSON:  string(data),
		})
		return reportID, err
	}
	return "", fmt.Errorf("storage unavailable")
}

// Startup is called by Wails when the application starts.
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx

	// Standard OS directory structure (with portable fallback)
	dataDir, _ := common.ConfigDir()
	logsDir, _ := common.LogsDir()

	if err := os.MkdirAll(dataDir, 0755); err != nil {
		common.LogWarn("Failed to create data directory: %v", err)
	}
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		common.LogWarn("Failed to create logs directory: %v", err)
	}

	a.currentDataDir = dataDir
	a.currentLogsDir = logsDir

	// Update subsystems with config-driven paths
	a.AIOps.SetDataDir(dataDir)
	a.AIOps.ctx = ctx
	a.DevOps.ctx = ctx

	// Initialize persistent storage locally
	dbPath := filepath.Join(dataDir, "universalops.db")
	if err := common.InitStorage(dbPath); err != nil {
		common.LogWarn("Failed to init local storage: %v", err)
	}

	// Initialize SLO engine and seed defaults
	if s := common.GetStorage(); s != nil {
		a.sloEngine = common.NewSLOEngine(s)
		a.sloEngine.SeedDefaultSLOs()
		a.Dash.SetSLOEngine(a.sloEngine)

		// Restore alert history from DB so it survives restarts
		if records, err := s.QueryAlertHistory(2000); err == nil && len(records) > 0 {
			a.alerts.RestoreFromDB(records)
		}
	}

	// Initialize the session logger locally
	logPath := filepath.Join(logsDir, "universalops.log")
	if err := common.InitLogger(logPath); err != nil {
		common.LogWarn("Failed to init local logger: %v", err)
	}

	// Load persisted settings into the pipeline before starting loops
	a.PipelineAPI.LoadSettings()
	a.AIOps.LoadModels()

	common.LogInfo("Universal-Ops initialized in self-contained mode (portable)")

	// Initialize System Knowledge Layer (linked to Pipeline for consistency)
	common.InitKnowledge(a.pipeline)

	// Validate Ollama environment variables
	validateOllamaEnv()

	// Prometheus metrics exporter removed — 100% local, zero telemetry.
	// If Prometheus-style instrumentation is needed, add it behind
	// //go:build prometheus and restore the InitMetricsExporter call.

	// Start the modular collector system
	a.collectorRegistry = common.NewCollectorRegistry()
	RegisterCollectors(a.collectorRegistry, a)
	a.scheduler = common.NewCollectorScheduler(a.collectorRegistry, a.pipeline)
	a.scheduler.Start()

	// Initialize Health facade now that the registry exists
	a.Health = NewHealth(a.pipeline, a.alerts, a.collectorRegistry)

	// Start the modular engine evaluation loop
	a.engineLoop.Start(2 * time.Second)

	// ── LibreHardwareMonitor (bundled, user-activated) ──────────────────
	lhm := common.GetLHMManager()
	if lhm.IsAvailable() {
		common.LogInfo("LHM binary found at %s (v%s) — waiting for user to enable sensors",
			lhm.BinaryPath(), lhm.Status().Version)
	} else {
		common.LogInfo("LHM not downloaded — user can enable from Hardware tab")
	}

	a.startProcessWorker()
}

// Shutdown is called by Wails when the application shuts down.
func (a *App) Shutdown(ctx context.Context) {
	common.LogInfo("Universal-Ops shutting down")

	// 0. Stop bundled LHM before general process cleanup
	common.GetLHMManager().Stop()

	// 1. Terminate any active child processes (zombie prevention)
	common.CleanupActiveProcesses()

	// Stop the collector scheduler
	if a.scheduler != nil {
		a.scheduler.Stop(5 * time.Second)
	}

	// Stop the engine loop
	if a.engineLoop != nil {
		a.engineLoop.Stop()
	}

	// Signal the process worker to stop
	if a.processWorkerQuit != nil {
		close(a.processWorkerQuit)
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
		Name:      "Universal-Ops Operations Platform",
		Version:   "1.3.1",
		GoVersion: goruntime.Version(),
		Uptime:    common.FormatUptime(uint64(time.Since(a.startedAt).Seconds())),
	}
}

type PerformanceProfile struct {
	Category      string  `json:"category"` // "laptop", "workstation", "high-end"
	Parallelism   int     `json:"parallelism"`
	ScanIntensity string  `json:"scan_intensity"` // "low", "medium", "high"
	CPUThreads    int     `json:"cpu_threads"`
	MemoryGB      float64 `json:"memory_gb"`
}

func (a *App) GetPerformanceProfile() PerformanceProfile {
	v, err := mem.VirtualMemory()
	cores := goruntime.NumCPU()

	totalMem := uint64(0)
	if err == nil && v != nil {
		totalMem = v.Total
	}

	memGB := float64(totalMem) / 1024 / 1024 / 1024

	profile := PerformanceProfile{
		CPUThreads: cores,
		MemoryGB:   memGB,
	}

	if cores <= 4 || memGB <= 8 {
		profile.Category = "laptop"
		profile.Parallelism = 2
		profile.ScanIntensity = "low"
	} else if cores <= 12 || memGB <= 32 {
		profile.Category = "workstation"
		profile.Parallelism = 8
		profile.ScanIntensity = "medium"
	} else {
		profile.Category = "high-end"
		profile.Parallelism = 32
		profile.ScanIntensity = "high"
	}

	return profile
}

type BaselineSnapshot struct {
	Timestamp time.Time `json:"timestamp"`
	Hardware  any       `json:"hardware"`
	Software  any       `json:"software"`
	Network   any       `json:"network"`
	Security  any       `json:"security"`
	Health    any       `json:"health"`
}

func (a *App) GenerateBaselineSnapshot() (*BaselineSnapshot, error) {
	// Snapshot using current state of facades
	cpuInfo := a.SysOps.GetCPUInfo()
	memInfo := a.SysOps.GetMemoryInfo()
	diskInfo := a.SysOps.GetDiskInfo()

	caps := a.GetSystemCapabilities()

	netInfo := a.NetOps.GetInterfaces()
	secStatus := a.SecOps.GetFirewallStatus()

	snapshot := &BaselineSnapshot{
		Timestamp: time.Now(),
		Hardware: map[string]any{
			"cpu":  cpuInfo,
			"mem":  memInfo,
			"disk": diskInfo,
		},
		Software: caps,
		Network:  netInfo,
		Security: secStatus,
		Health: map[string]any{
			"uptime": a.GetAppInfo().Uptime,
		},
	}

	// Save to data/baseline.json for future comparison
	path := filepath.Join(a.GetDataDir(), "baseline.json")
	if b, err := json.MarshalIndent(snapshot, "", "  "); err == nil {
		os.WriteFile(path, b, 0644)
	}

	return snapshot, nil
}

// IsOnboarded returns true if the user has completed the onboarding flow.
func (a *App) IsOnboarded() bool {
	return common.IsOnboarded()
}

// MarkOnboarded marks the onboarding flow as completed.
func (a *App) MarkOnboarded() {
	if err := common.MarkOnboarded(); err != nil {
		common.LogWarn("MarkOnboarded: %v", err)
	}
}

// ClearOnboarded removes the onboarding marker, forcing setup on next run.
func (a *App) ClearOnboarded() {
	if err := common.ClearOnboarded(); err != nil {
		common.LogWarn("ClearOnboarded: %v", err)
	}
}

// ApplyOperationalProfile adjusts engine parameters based on a selected profile (eco, standard, burst).
func (a *App) ApplyOperationalProfile(profile string) {
	interval := 3000
	level := "debug" // AUDIT: Force debug level for Phase 0 telemetry

	switch profile {
	case "eco":
		interval = 10000
	case "burst":
		interval = 1000
	}

	common.LogInfo("App: Applying operational profile %q (interval: %vms, log: %s)", profile, interval, level)
	a.PipelineAPI.UpdateSettings(interval, 0, 4, 2000)
	a.SetLogLevel(level)
}

// SetLogLevel updates the backend system log verbosity.
func (a *App) SetLogLevel(level string) {
	common.SetLogLevel(level)
}

// UpdateStorageConfig moves the internal database and log files to a new location.
func (a *App) UpdateStorageConfig(dbDir string) error {
	a.storageLock.Lock()
	defer a.storageLock.Unlock()

	// Safety Check: Ensure directory is writable
	testFile := filepath.Join(dbDir, ".write_test")
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return fmt.Errorf("cannot create directory: %w", err)
	}
	if err := os.WriteFile(testFile, []byte("ok"), 0644); err != nil {
		return fmt.Errorf("directory is not writable: %w", err)
	}
	os.Remove(testFile)

	dbPath := filepath.Join(dbDir, "universalops.db")

	common.LogInfo("App: Suspending operations for storage relocation to %s", dbDir)

	// 1. Stop the scheduler to prevent background writes during relocation
	if a.scheduler != nil {
		a.scheduler.Stop(2 * time.Second)
	}

	// 2. Shutdown current storage
	s := common.GetStorage()
	if s != nil {
		s.Close()
	}

	// 3. Initialize new storage
	if err := common.InitStorage(dbPath); err != nil {
		return fmt.Errorf("failed to re-init storage: %w", err)
	}

	a.currentDataDir = dbDir

	// 4. Restart the scheduler
	if a.scheduler != nil {
		a.scheduler = common.NewCollectorScheduler(a.collectorRegistry, a.pipeline)
		a.scheduler.Start()
	}

	common.LogInfo("App: Storage successfully relocated and operations resumed.")
	return nil
}

func (a *App) GetDataDir() string {
	if a.currentDataDir == "" {
		return "data"
	}
	return a.currentDataDir
}

func (a *App) GetLogsDir() string {
	if a.currentLogsDir == "" {
		return "logs"
	}
	return a.currentLogsDir
}

// UpdateLogsConfig relocates the active log file.
func (a *App) UpdateLogsConfig(logDir string) error {
	logPath := filepath.Join(logDir, "universalops.log")
	common.LogInfo("App: Relocating log file to %s", logPath)

	common.CloseLogger()
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log dir: %w", err)
	}

	if err := common.InitLogger(logPath); err != nil {
		return fmt.Errorf("failed to re-init logger: %w", err)
	}

	a.currentLogsDir = logDir
	common.LogInfo("App: Logs successfully relocated.")
	return nil
}

// GetSystemCapabilities returns a list of detected tools and binaries on the host.
func (a *App) GetSystemCapabilities() []common.CapabilityInfo {
	if a.capabilities == nil {
		return nil
	}
	return a.capabilities.List()
}

// VerifyCapability triggers a focused re-scan of a specific tool and returns its new state.
func (a *App) VerifyCapability(id string) (common.CapabilityInfo, error) {
	if a.capabilities == nil {
		return common.CapabilityInfo{}, fmt.Errorf("capability registry unavailable")
	}

	capID := common.CapabilityID(id)
	a.capabilities.RefreshBatch([]common.CapabilityID{capID})

	list := a.capabilities.List()
	for _, info := range list {
		if info.ID == capID {
			return info, nil
		}
	}

	return common.CapabilityInfo{}, fmt.Errorf("capability %s not found after verification", id)
}

// SetCapabilityOverride allows the frontend to manually set a path for a tool.
func (a *App) SetCapabilityOverride(id string, path string) {
	if a.capabilities != nil {
		a.capabilities.SetOverride(common.CapabilityID(id), path)
		common.LogInfo("App: Capability override set for %s -> %s", id, path)
	}
}

// ── Database Maintenance ───────────────────────────────────────────────────

// VacuumDatabase rebuilds the database to reclaim space.
func (a *App) VacuumDatabase() error {
	s := common.GetStorage()
	if s == nil {
		return fmt.Errorf("storage unavailable")
	}
	return s.Vacuum()
}

// AnalyzeDatabase verifies integrity and updates statistics.
func (a *App) AnalyzeDatabase() error {
	s := common.GetStorage()
	if s == nil {
		return fmt.Errorf("storage unavailable")
	}
	// Integrity check is a separate pragma, Analyze is for query planning
	return s.Analyze()
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

// ReadTextFile reads the content of a text file and returns it as a string.
func (a *App) ReadTextFile(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// SelectFolderDialog shows a directory selection dialog and returns the selected path.
func (a *App) SelectFolderDialog(title string) (string, error) {
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: title,
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

// startProcessWorker performs periodic heavy process scans in the background.
// This decouples the per-tick collectors from the multi-millisecond cost
// of a full system process enumeration.
func (a *App) startProcessWorker() {
	a.processWorkerQuit = make(chan struct{})
	go func() {
		defer common.RecoverPanic()

		// Full process scan every 5 seconds
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		// Memory reporting ticker (60s)
		memTicker := time.NewTicker(60 * time.Second)
		defer memTicker.Stop()

		// Run immediate first scan
		_ = sysopsPkg.UpdateProcessSnapshot()

		tickCount := 0

		for {
			select {
			case <-ticker.C:
				if err := sysopsPkg.UpdateProcessSnapshot(); err != nil {
					common.LogWarn("Process snapshot failed: %v", err)
				}
				tickCount++
				if tickCount%12 == 0 { // Every 60s
					goruntime.GC()
				}
			case <-memTicker.C:
				var m goruntime.MemStats
				goruntime.ReadMemStats(&m)
				common.LogInfo("Engine Memory: Alloc=%vMB, TotalAlloc=%vMB, Sys=%vMB, NumGC=%v",
					m.Alloc/1024/1024, m.TotalAlloc/1024/1024, m.Sys/1024/1024, m.NumGC)
			case <-a.processWorkerQuit:
				common.LogInfo("Process worker goroutine stopped")
				return
			}
		}
	}()
}

// refreshSecurityScore executes heavy OS scans for security posture.
// Decoupled into its own worker to prevent ticker-blocking.

// persistAlertsAsync handles DB writes without blocking the metrics stream.
// Runs as a goroutine — MUST have its own RecoverPanic guard.
func (a *App) persistAlertsAsync(newAlerts []common.Alert) {
	defer common.RecoverPanic()

	storage := common.GetStorage()
	if storage == nil {
		return
	}

	tx, err := storage.Begin()
	if err != nil {
		return
	}

	for _, alert := range newAlerts {
		storage.InsertAlert(common.AlertRecord{
			ID:        alert.ID,
			Timestamp: alert.Timestamp,
			Level:     alert.Level.String(),
			Metric:    alert.Metric,
			Message:   alert.Message,
			Value:     alert.Value,
			Threshold: alert.Threshold,
		}, tx)
	}
	if err := tx.Commit(); err != nil {
		common.LogError("App: failed to commit alert transaction: %v", err)
	}
}

// emitAlertEvents sends fired alerts to the frontend.
func (a *App) emitAlertEvents(alerts []common.Alert) {
	for _, alert := range alerts {
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

	// Trigger immediate alert evaluation in a separate goroutine to avoid
	// blocking the Wails binding return.
	if a.engineLoop != nil {
		go a.engineLoop.Step()
	}

	return nil
}

// ConfirmAction executes a previously registered pending action via safety handshake.
func (a *App) ConfirmAction(handshakeID string) common.SecActionResult {
	pending, err := common.GetHandshakeRegistry().Consume(handshakeID)
	if err != nil {
		return common.SecActionResult{Success: false, Error: err.Error()}
	}

	// Compliance: Log the user's approval of the AI-suggested action
	if storage := common.GetStorage(); storage != nil {
		argsJSON, _ := json.Marshal(pending.Params)
		ctxSnap := ""
		if knowledge := common.GetKnowledge(); knowledge != nil {
			snap, _ := json.Marshal(knowledge.GetSnapshot())
			ctxSnap = string(snap)
		}
		_ = storage.LogDecision(pending.SessionID, pending.Action, string(argsJSON), ctxSnap, true)
	}

	// Helper to safely extract string params
	getStringParam := func(m map[string]interface{}, key string) string {
		if val, ok := m[key]; ok {
			if s, ok := val.(string); ok {
				return s
			}
		}
		return ""
	}

	switch pending.Action {
	case "kill_process", "KillProcess":
		pidRaw := pending.Params["pid"]
		var pid int
		switch v := pidRaw.(type) {
		case int:
			pid = v
		case string:
			fmt.Sscanf(v, "%d", &pid)
		case float64:
			pid = int(v)
		}
		if pid == 0 {
			return common.SecActionResult{Success: false, Error: "Invalid PID (0 or missing)"}
		}
		return a.SecOps.executeKillProcess(pid)

	case "kill_process_tree", "KillProcessTree":
		pidRaw := pending.Params["pid"]
		var pid int
		switch v := pidRaw.(type) {
		case int:
			pid = v
		case string:
			fmt.Sscanf(v, "%d", &pid)
		case float64:
			pid = int(v)
		}
		if pid == 0 {
			return common.SecActionResult{Success: false, Error: "Invalid PID (0 or missing)"}
		}
		return a.SecOps.executeKillProcessTree(pid)

	case "ApplyHardening":
		check := getStringParam(pending.Params, "check")
		if check == "" {
			return common.SecActionResult{Success: false, Error: "Missing 'check' parameter"}
		}
		return a.SecOps.executeApplyHardening(check)

	case "reboot", "shutdown", "flush_dns", "disk_cleanup", "clear_temp", "defrag", "system_update", "clean_pkg_cache", "sleep", "hibernate", "clear_arp":
		// Generic system actions
		return a.SysOps.executeSystemAction(pending.Action)

	case "block_ip", "BlockIP":
		ip := getStringParam(pending.Params, "ip")
		if ip == "" {
			return common.SecActionResult{Success: false, Error: "Missing 'ip' parameter"}
		}
		return a.SecOps.executeBlockIP(ip)

	case "isolate_host", "IsolateHost":
		// Handle both bool and string from AI
		confirm := true
		if v, ok := pending.Params["confirm"].(bool); ok {
			confirm = v
		}
		return a.SecOps.executeIsolateHost(confirm, 3600) // Default 1hr isolation

	case "restart_service", "RestartService":
		// Try both 'name' and 'service' (AI sometimes confuses these)
		name := getStringParam(pending.Params, "name")
		if name == "" {
			name = getStringParam(pending.Params, "service")
		}

		if name == "" {
			return common.SecActionResult{Success: false, Error: "Missing service name parameter"}
		}

		common.LogInfo("ConfirmAction: Restarting service %q", name)
		res := a.SysOps.executeRestartService(name)
		return common.SecActionResult{Success: res.Success, Message: res.Message}

	case "CaptureEvidence":
		return a.SecOps.executeCaptureEvidence()

	case "ExportForensicBundle":
		snapshotID := getStringParam(pending.Params, "id")
		return a.SecOps.executeExportForensicBundle(snapshotID)

	case "workflow":
		workflowID := getStringParam(pending.Params, "workflow_id")
		if workflowID == "" {
			return common.SecActionResult{Success: false, Error: "Missing 'workflow_id' parameter"}
		}
		reportID, err := a.TriggerWorkflow(workflowID)
		if err != nil {
			return common.SecActionResult{Success: false, Error: "Workflow execution failed: " + err.Error()}
		}
		return common.SecActionResult{Success: true, Message: "Workflow executed successfully", Detail: "Report ID: " + reportID}

	default:
		return common.SecActionResult{Success: false, Error: "Unknown or unimplemented action type: " + pending.Action}
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
		common.LogInfo("OLLAMA_MODEL not set, defaulting to universalops")
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
