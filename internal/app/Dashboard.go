package app

import (
	"fmt"

	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
)

// Dashboard exposes dashboard-related bindings to the frontend.
type Dashboard struct {
	app *App
}

// NewDashboard creates a new Dashboard facade.
func NewDashboard(app *App) *Dashboard {
	return &Dashboard{app: app}
}

// RunQuickDiag performs a quick system diagnostic and returns categorized results.
func (d *Dashboard) RunQuickDiag() []DiagnosticResult {
	p := d.app.pipeline

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
	alertCount := d.app.alerts.AlertCount()
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
	p := d.app.pipeline

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
	alertCount := d.app.alerts.AlertCount()
	if alertCount > 0 {
		alerts := d.app.alerts.GetAlerts()
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
func (d *Dashboard) GetDashboardData() DashboardData {
	p := d.app.pipeline

	cpuMF := p.GetMetricWithForecast(common.MetricCPU)
	memMF := p.GetMetricWithForecast(common.MetricMem)
	diskMF := p.GetMetricWithForecast(common.MetricDisk)
	netRXMF := p.GetMetricWithForecast(common.MetricNetRX)
	netTXMF := p.GetMetricWithForecast(common.MetricNetTX)
	procMF := p.GetMetricWithForecast(common.MetricProcCnt)

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
		Network: NetworkMetric{
			RXRate: netRXMF.LastValue,
			TXRate: netTXMF.LastValue,
			Unit:   "bps",
		},
		Processes:   int(procMF.LastValue),
		Connections: len(d.app.NetOps.GetConnections()),
		Alerts:      d.app.alerts.AlertCount(),
		Uptime:      d.app.GetAppInfo().Uptime,
	}
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
