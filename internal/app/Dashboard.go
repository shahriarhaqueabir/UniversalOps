package app

import (
	"fmt"
	"time"

	"github.com/shahriarhaqueabir/UniversalOps/internal/common"
)

// Dashboard exposes dashboard-related bindings to the frontend.
type Dashboard struct {
	pipeline     *common.DataPipeline
	alerts       *common.AlertEngine
	sysOps       *SysOps
	netOps       *NetOps
	secOps       *SecOps
	devOps       *DevOps
	aiOps        *AIOps
	timeline     *Timeline
	sloEngine    *common.SLOEngine
	uptimeGetter func() string
}

// NewDashboard creates a new Dashboard facade.
func NewDashboard(pipeline *common.DataPipeline, alerts *common.AlertEngine, sysOps *SysOps, netOps *NetOps, secOps *SecOps, devOps *DevOps, aiOps *AIOps, timeline *Timeline, sloEngine *common.SLOEngine, uptimeGetter func() string) *Dashboard {
	return &Dashboard{
		pipeline:     pipeline,
		alerts:       alerts,
		sysOps:       sysOps,
		netOps:       netOps,
		secOps:       secOps,
		devOps:       devOps,
		aiOps:        aiOps,
		timeline:     timeline,
		sloEngine:    sloEngine,
		uptimeGetter: uptimeGetter,
	}
}

// SetSLOEngine sets the SLO engine on the Dashboard (called after storage init).
func (d *Dashboard) SetSLOEngine(engine *common.SLOEngine) {
	defer common.RecoverPanic()
	d.sloEngine = engine
}

// RunQuickDiag performs a quick system diagnostic and returns categorized results.
func (d *Dashboard) RunQuickDiag() []DiagnosticResult {
	defer common.RecoverPanic()
	p := d.pipeline

	cpuMF := p.GetMetricWithForecast(common.MetricCPU)
	memMF := p.GetMetricWithForecast(common.MetricMem)
	diskMF := p.GetMetricWithForecast(common.MetricDisk)
	procMF := p.GetMetricWithForecast(common.MetricProcCnt)

	results := []DiagnosticResult{
		{Category: "CPU", Status: diagStatus(cpuMF.LastValue, 80, 90), Message: cpuDiagMsg(cpuMF.LastValue), Value: cpuMF.LastValue, Unit: "%"},
		{Category: "Memory", Status: diagStatus(memMF.LastValue, 85, 92), Message: memDiagMsg(memMF.LastValue), Value: memMF.LastValue, Unit: "%"},
		{Category: "Disk", Status: diagStatus(diskMF.LastValue, 85, 95), Message: diskDiagMsg(diskMF.LastValue), Value: diskMF.LastValue, Unit: "%"},
		{Category: "Processes", Status: "info", Message: procDiagMsg(int(procMF.LastValue)), Value: procMF.LastValue, Unit: "count"},
	}

	// Add alert count
	alertCount := d.alerts.AlertCount()
	alertStatus := "pass"
	if alertCount > 0 {
		alertStatus = "warn"
	}
	results = append(results, DiagnosticResult{
		Category: "Alerts",
		Status:   alertStatus,
		Message:  fmt.Sprintf("%d active alert(s) in the system", alertCount),
		Value:    float64(alertCount), Unit: "count",
	})

	return results
}

// GenerateDashboardBriefing generates a full operations briefing from pipeline data.
func (d *Dashboard) GenerateDashboardBriefing() []BriefingSection {
	defer common.RecoverPanic()
	p := d.pipeline

	cpuMF := p.GetMetricWithForecast(common.MetricCPU)
	memMF := p.GetMetricWithForecast(common.MetricMem)
	diskMF := p.GetMetricWithForecast(common.MetricDisk)
	netRXMF := p.GetMetricWithForecast(common.MetricNetRX)
	netTXMF := p.GetMetricWithForecast(common.MetricNetTX)

	sections := []BriefingSection{
		{Title: "CPU Analysis", Level: "info", Content: fmt.Sprintf(
			"Current CPU usage is %.1f%% (trend: %s). %s",
			cpuMF.LastValue, trendDirectionString(cpuMF.Trend.Direction),
			cpuDiagMsg(cpuMF.LastValue),
		)},
		{Title: "Memory Analysis", Level: "info", Content: fmt.Sprintf(
			"Current memory utilization is %.1f%% (trend: %s). %s",
			memMF.LastValue, trendDirectionString(memMF.Trend.Direction),
			memDiagMsg(memMF.LastValue),
		)},
		{Title: "Storage Analysis", Level: "info", Content: fmt.Sprintf(
			"Disk usage is at %.1f%% (trend: %s). %s",
			diskMF.LastValue, trendDirectionString(diskMF.Trend.Direction),
			diskDiagMsg(diskMF.LastValue),
		)},
		{Title: "Network Activity", Level: "info", Content: fmt.Sprintf(
			"Network throughput: RX %.2f Mbps / TX %.2f Mbps",
			netRXMF.LastValue/1e6, netTXMF.LastValue/1e6,
		)},
	}

	// Add alert section if alerts exist
	alertCount := d.alerts.AlertCount()
	if alertCount > 0 {
		alerts := d.alerts.ActiveAlerts()
		alertText := ""
		for _, a := range alerts {
			if !a.Resolved {
				alertText += fmt.Sprintf("- [%s] %s: %.1f (threshold: %.1f)\n", a.Level.String(), a.Metric, a.Value, a.Threshold)
			}
		}
		sections = append(sections, BriefingSection{
			Title: "Active Alerts", Level: "warning",
			Content: fmt.Sprintf("There are %d active alert(s):\n%s", alertCount, alertText),
		})
	} else {
		sections = append(sections, BriefingSection{
			Title: "Active Alerts", Level: "info",
			Content: "No active alerts. All monitored metrics are within normal parameters.",
		})
	}

	return sections
}

// GetDashboardData returns a snapshot of all key metrics for the dashboard view.
func (d *Dashboard) GetDashboardData() (out DashboardData) {
	defer func() {
		if r := recover(); r != nil {
			common.LogWarn("GetDashboardData recovered panic: %v", r)
			out = DashboardData{} // Return zero-value instead of crashing
		}
	}()

	p := d.pipeline

	cpuMF := p.GetMetricWithForecast(common.MetricCPU)
	memMF := p.GetMetricWithForecast(common.MetricMem)
	diskMF := p.GetMetricWithForecast(common.MetricDisk)
	netRXMF := p.GetMetricWithForecast(common.MetricNetRX)
	netTXMF := p.GetMetricWithForecast(common.MetricNetTX)
	procMF := p.GetMetricWithForecast(common.MetricProcCnt)

	gpuInfo := d.sysOps.GetGPUInfo()
	battInfo := d.sysOps.GetBatteryInfo()

	uptime := ""
	if d.uptimeGetter != nil {
		uptime = d.uptimeGetter()
	}

	// Fetch health score trend from storage
	healthScore := 100
	var healthTrend []HealthScorePoint
	if s := common.GetStorage(); s != nil {
		trend, err := s.GetHealthScoreTrend(7)
		if err == nil && len(trend) > 0 {
			// Most recent day's score is the current health score
			var latestDay string
			var latestScore int
			for day, score := range trend {
				if day > latestDay {
					latestDay = day
					latestScore = score
				}
			}
			healthScore = latestScore

			// Build ordered trend slice (ascending by day)
			for i := 6; i >= 0; i-- {
				day := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
				if score, ok := trend[day]; ok {
					healthTrend = append(healthTrend, HealthScorePoint{Day: day, Score: score})
				}
			}
		}
	}

	return DashboardData{
		CPU: GaugeMetric{
			Value:    cpuMF.LastValue,
			Unit:     cpuMF.Unit,
			History:  safeLastN(cpuMF.Values, 120),
			Forecast: cpuMF.Forecast,
			Trend:    trendDirectionString(cpuMF.Trend.Direction),
		},
		Memory: GaugeMetric{
			Value:    memMF.LastValue,
			Unit:     memMF.Unit,
			History:  safeLastN(memMF.Values, 120),
			Forecast: memMF.Forecast,
			Trend:    trendDirectionString(memMF.Trend.Direction),
		},
		Disk: GaugeMetric{
			Value:    diskMF.LastValue,
			Unit:     diskMF.Unit,
			History:  safeLastN(diskMF.Values, 120),
			Forecast: diskMF.Forecast,
			Trend:    trendDirectionString(diskMF.Trend.Direction),
		},
		GPU: GPUInfo{
			Name:        gpuInfo.Name,
			Vendor:      gpuInfo.Vendor,
			MemoryGB:    gpuInfo.MemoryGB,
			Driver:      gpuInfo.Driver,
			Detected:    gpuInfo.Detected,
			Temperature: gpuInfo.Temperature,
			Utilization: gpuInfo.Utilization,
			FanSpeed:    gpuInfo.FanSpeed,
		},
		Battery: BatteryInfo{
			Percent:     battInfo.Percent,
			Charging:    battInfo.Charging,
			TimeLeftSec: battInfo.TimeLeftSec,
			Status:      battInfo.Status,
			Detected:    battInfo.Detected,
		},
		Network: NetworkMetric{
			RXRate: netRXMF.LastValue,
			TXRate: netTXMF.LastValue,
			Unit:   "bps",
		},
		Processes:   int(procMF.LastValue),
		Connections: len(d.netOps.GetConnections()),
		Alerts:      d.alerts.AlertCount(),
		Uptime:      uptime,
		HealthScore: healthScore,
		HealthTrend: healthTrend,
	}
}

// GetSystemSnapshot returns a full state snapshot for efficient Batch IPC.
func (d *Dashboard) GetSystemSnapshot() (out SystemSnapshot) {
	defer func() {
		if r := recover(); r != nil {
			common.LogWarn("GetSystemSnapshot recovered panic: %v", r)
			out = SystemSnapshot{} // Return zero-value instead of crashing
		}
	}()

	metrics := d.GetDashboardData()

	// Fetch last 10 alerts
	alerts := d.alerts.ActiveAlerts()
	var alertInfos []AlertInfo
	for _, a := range alerts {
		alertInfos = append(alertInfos, convertAlert(a))
	}
	if len(alertInfos) > 10 {
		alertInfos = alertInfos[:10]
	}

	// Fetch last 20 events
	timeline := d.timeline.GetRecentEvents(20)

	return SystemSnapshot{
		Metrics:   metrics,
		Alerts:    alertInfos,
		Timeline:  timeline,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
}

// ── Cross-Pillar Summary Methods ──────────────────────────────────────────────

// GetSecuritySummary returns a lightweight security posture summary for the dashboard.
func (d *Dashboard) GetSecuritySummary() SecuritySummary {
	defer common.RecoverPanic()
	if d.secOps == nil {
		return SecuritySummary{}
	}
	return d.secOps.GetSecuritySummary()
}

// GetSecurityScore returns the current security score for the dashboard.
func (d *Dashboard) GetSecurityScore() SecurityScore {
	defer common.RecoverPanic()
	if d.secOps == nil {
		return SecurityScore{}
	}
	return d.secOps.GetSecurityScore()
}

// GetDevOpsSummary returns a lightweight DevOps health summary for the dashboard.
func (d *Dashboard) GetDevOpsSummary() DevOpsSummary {
	defer common.RecoverPanic()
	if d.devOps == nil {
		return DevOpsSummary{}
	}
	svcSummary := d.devOps.GetServiceGroupSummary()
	dockerStatus := d.devOps.GetDockerStatus()
	k8sStatus := d.devOps.GetKubernetesStatus()

	totalSvcs := svcSummary.Databases + svcSummary.MessageQueues + svcSummary.WebServers + svcSummary.Containers + svcSummary.Other
	summary := "All services nominal"
	if svcSummary.Stopped > 0 {
		summary = fmt.Sprintf("%d service(s) stopped", svcSummary.Stopped)
	}
	if !dockerStatus.Running {
		summary = "Docker daemon not running"
	}

	return DevOpsSummary{
		ServiceCount:    totalSvcs,
		RunningCount:    svcSummary.Running,
		DockerInstalled: dockerStatus.Installed,
		DockerRunning:   dockerStatus.Running,
		ContainerCount:  dockerStatus.Containers.Running,
		K8sInstalled:    k8sStatus.Installed,
		K8sConnected:    k8sStatus.Connected,
		K8sPods:         k8sStatus.Pods,
		Summary:         summary,
	}
}

// GetDockerStatus returns Docker daemon status for the dashboard.
func (d *Dashboard) GetDockerStatus() DockerStatus {
	defer common.RecoverPanic()
	if d.devOps == nil {
		return DockerStatus{}
	}
	return d.devOps.GetDockerStatus()
}

// GetKubernetesStatus returns Kubernetes cluster status for the dashboard.
func (d *Dashboard) GetKubernetesStatus() KubernetesStatus {
	defer common.RecoverPanic()
	if d.devOps == nil {
		return KubernetesStatus{}
	}
	return d.devOps.GetKubernetesStatus()
}

// GetAIOpsSummary returns a lightweight AIOps status summary for the dashboard.
func (d *Dashboard) GetAIOpsSummary() AIOpsSummary {
	defer common.RecoverPanic()
	if d.aiOps == nil {
		return AIOpsSummary{}
	}
	ollamaStatus := d.aiOps.GetOllamaStatus()
	anomalies := d.aiOps.DetectAnomalies()
	insights := d.aiOps.GetAIInsights()

	anomalyCount := len(anomalies)
	criticalAnomalies := 0
	for _, a := range anomalies {
		if a.Severity == "critical" {
			criticalAnomalies++
		}
	}

	return AIOpsSummary{
		OllamaAvailable:   ollamaStatus.Available,
		OllamaModel:       ollamaStatus.Model,
		AnomalyCount:      anomalyCount,
		CriticalAnomalies: criticalAnomalies,
		RecentInsights:    insights,
	}
}

// GetSLOSummary evaluates all SLO definitions and returns the summary.
func (d *Dashboard) GetSLOSummary() common.SLOSummary {
	defer common.RecoverPanic()
	if d.sloEngine == nil {
		return common.SLOSummary{}
	}
	summary, err := d.sloEngine.EvaluateAll()
	if err != nil {
		common.LogWarn("Dashboard: GetSLOSummary failed: %v", err)
		return common.SLOSummary{}
	}
	return summary
}

// GetSLODefinitions returns all SLO definitions.
func (d *Dashboard) GetSLODefinitions() []common.SLODefinition {
	defer common.RecoverPanic()
	if s := common.GetStorage(); s != nil {
		defs, err := s.ListSLODefinitions()
		if err != nil {
			common.LogWarn("Dashboard: GetSLODefinitions failed: %v", err)
			return nil
		}
		return defs
	}
	return nil
}

// ── Dashboard Layout Persistence (DASH-01) ───────────────────────────────────

const dashboardLayoutKey = "dashboardLayout"

// SaveDashboardLayout persists the drag-and-drop widget layout as JSON.
// The layout is a JSON array of widget descriptors (id + order). Stored in the
// settings table so it survives restarts without a schema change.
func (d *Dashboard) SaveDashboardLayout(layoutJSON string) error {
	defer common.RecoverPanic()
	if layoutJSON == "" {
		return fmt.Errorf("layout is required")
	}
	s := common.GetStorage()
	if s == nil {
		return fmt.Errorf("storage not initialized")
	}
	if err := s.UpsertSetting(dashboardLayoutKey, layoutJSON); err != nil {
		common.LogWarn("Dashboard: SaveDashboardLayout failed: %v", err)
		return err
	}
	return nil
}

// GetDashboardLayout returns the persisted widget layout JSON, or "" if none.
func (d *Dashboard) GetDashboardLayout() string {
	defer common.RecoverPanic()
	s := common.GetStorage()
	if s == nil {
		return ""
	}
	val, err := s.GetSetting(dashboardLayoutKey)
	if err != nil {
		common.LogWarn("Dashboard: GetDashboardLayout failed: %v", err)
		return ""
	}
	return val
}

// ResetDashboardLayout clears the persisted layout so the default order is used.
func (d *Dashboard) ResetDashboardLayout() error {
	defer common.RecoverPanic()
	s := common.GetStorage()
	if s == nil {
		return fmt.Errorf("storage not initialized")
	}
	if err := s.UpsertSetting(dashboardLayoutKey, ""); err != nil {
		common.LogWarn("Dashboard: ResetDashboardLayout failed: %v", err)
		return err
	}
	return nil
}

// ── Diagnostic helpers ─────────────────────────────────────────────────────────

func diagStatus(value float64, warnAt, failAt float64) string {
	switch {
	case value >= failAt:
		return "fail"
	case value >= warnAt:
		return "warn"
	default:
		return "pass"
	}
}

func cpuDiagMsg(value float64) string {
	switch {
	case value >= 90:
		return "Critical CPU utilization — possible process contention or runaway threads."
	case value >= 80:
		return "Elevated CPU utilization — check for background jobs or bottlenecks."
	case value >= 50:
		return "Moderate CPU utilization — normal under load."
	default:
		return "CPU utilization is normal — system is responsive."
	}
}

func memDiagMsg(value float64) string {
	switch {
	case value >= 92:
		return "Critical memory pressure — system may be swapping heavily."
	case value >= 85:
		return "High memory utilization — consider closing unused applications."
	case value >= 60:
		return "Moderate memory utilization — within expected range."
	default:
		return "Memory utilization is healthy — adequate headroom."
	}
}

func diskDiagMsg(value float64) string {
	switch {
	case value >= 95:
		return "Critical disk usage — immediately free space to prevent failures."
	case value >= 85:
		return "High disk usage — consider cleanup or storage expansion."
	case value >= 60:
		return "Moderate disk usage — monitor for growth trends."
	default:
		return "Disk usage is healthy — sufficient free space available."
	}
}

func procDiagMsg(count int) string {
	switch {
	case count > 500:
		return "High process count — investigate for fork bombs or excessive services."
	case count > 200:
		return "Moderate process count — typical for an active system."
	default:
		return "Low process count — system is running lean."
	}
}
