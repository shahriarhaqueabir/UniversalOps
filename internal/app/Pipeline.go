package app

import (
	"time"

	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
)

// PipelineAPI exposes data pipeline bindings to the frontend.
type PipelineAPI struct {
	pipeline *common.DataPipeline
}

// NewPipelineAPI creates a new PipelineAPI facade.
func NewPipelineAPI(pipeline *common.DataPipeline) *PipelineAPI {
	return &PipelineAPI{pipeline: pipeline}
}

// GetMetricHistory returns the last n values for a named metric.
func (p *PipelineAPI) GetMetricHistory(name string, n int) []float64 {
	ts := p.pipeline.GetTimeSeries(name)
	if ts == nil {
		return []float64{}
	}
	values := ts.Values()
	if n > 0 && len(values) > n {
		return values[len(values)-n:]
	}
	return values
}

// GetMetricHistoryWithTimestamps returns the last n values with timestamps for a named metric.
func (p *PipelineAPI) GetMetricHistoryWithTimestamps(name string, n int) []DataPoint {
	ts := p.pipeline.GetTimeSeries(name)
	if ts == nil {
		return []DataPoint{}
	}
	points := ts.DataPoints()
	if n > 0 && len(points) > n {
		points = points[len(points)-n:]
	}
	out := make([]DataPoint, len(points))
	for i, pt := range points {
		out[i] = DataPoint{
			Time:  pt.Time.Format(time.RFC3339),
			Value: pt.Value,
		}
	}
	return out
}

// GetForecast returns predicted values for a named metric.
func (p *PipelineAPI) GetForecast(name string, steps int) []float64 {
	return p.pipeline.GetForecast(name, steps)
}

// GetTrend returns the current trend for a named metric.
func (p *PipelineAPI) GetTrend(name string) TrendInfo {
	trend := p.pipeline.GetTrend(name)
	return TrendInfo{
		Direction:   trendDirectionString(trend.Direction),
		ChangePct:   trend.ChangePct,
		Slope:       trend.Slope,
		Correlation: trend.Correlation,
	}
}

// GetWindowStats returns rolling statistics for a named metric.
func (p *PipelineAPI) GetWindowStats(name string) StatsInfo {
	stats := p.pipeline.GetWindowStats(name)
	return StatsInfo{
		Min:   stats.Min,
		Max:   stats.Max,
		Avg:   stats.Avg,
		P50:   stats.P50,
		P95:   stats.P95,
		P99:   stats.P99,
		Count: stats.Count,
	}

}

// GetMetricWithForecast returns a complete metric snapshot with forecast.
func (p *PipelineAPI) GetMetricWithForecast(name string) MetricHistory {
	mf := p.pipeline.GetMetricWithForecast(name)
	return MetricHistory{
		Name:      mf.Name,
		Unit:      mf.Unit,
		Values:    mf.Values,
		Forecast:  mf.Forecast,
		Trend:     convertTrendInfo(mf.Trend),
		Stats:     convertStatsInfo(mf.Stats),
		LastValue: mf.LastValue,
	}
}

// AllMetricNames returns all tracked metric names and units.
func (p *PipelineAPI) AllMetricNames() []MetricDef {
	metrics := common.DefaultMetrics
	out := make([]MetricDef, 0, len(metrics))
	for _, m := range metrics {
		out = append(out, MetricDef{Name: m.Name, Unit: m.Unit, Label: m.Label})
	}
	return out
}

// ClearPipeline resets all stored data and forecasts.
func (p *PipelineAPI) ClearPipeline() {
	p.pipeline.Clear()
}

// UpdateSettings updates the pipeline configuration and network defaults.
func (p *PipelineAPI) UpdateSettings(intervalMs int, capacity int, pingCount int, dnsTimeout int) {
	cfg := p.pipeline.Config()

	if intervalMs > 0 {
		cfg.TickInterval = time.Duration(intervalMs) * time.Millisecond
	}

	if capacity > 0 {
		cfg.Capacity = capacity
	}

	if pingCount > 0 {
		cfg.PingCount = pingCount
	}

	if dnsTimeout > 0 {
		cfg.DNSTimeout = dnsTimeout
	}

	p.pipeline.UpdateConfig(cfg)
}

// GetCurrentSettings returns the current operational settings of the data pipeline.
func (p *PipelineAPI) GetCurrentSettings() map[string]interface{} {
	cfg := p.pipeline.Config()
	return map[string]interface{}{
		"refreshInterval": int(cfg.TickInterval.Milliseconds()),
		"pingCount":       cfg.PingCount,
		"dnsTimeout":      cfg.DNSTimeout,
	}
}

// ── Converters ───────────────────────────────────────────────────────────────

func convertTrendInfo(t common.TrendInfo) TrendInfo {
	return TrendInfo{
		Direction:   trendDirectionString(t.Direction),
		ChangePct:   t.ChangePct,
		Slope:       t.Slope,
		Correlation: t.Correlation,
	}
}

func convertStatsInfo(s common.WindowStats) StatsInfo {
	return StatsInfo{
		Min:   s.Min,
		Max:   s.Max,
		Avg:   s.Avg,
		P50:   s.P50,
		P95:   s.P95,
		P99:   s.P99,
		Count: s.Count,
	}
}
