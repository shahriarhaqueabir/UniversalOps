package app

import (
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
		Connections: 0, // fetched on-demand via NetOps
		Alerts:      d.app.alerts.AlertCount(),
		Uptime:      d.app.GetAppInfo().Uptime,
	}
}
