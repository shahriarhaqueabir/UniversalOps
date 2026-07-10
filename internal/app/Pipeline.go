package app

import (
	"time"

	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
)

// PipelineAPI exposes data pipeline bindings to the frontend.
type PipelineAPI struct {
	app *App
}

// NewPipelineAPI creates a new PipelineAPI facade.
func NewPipelineAPI(app *App) *PipelineAPI {
	return &PipelineAPI{app: app}
}

// GetMetricHistory returns the last n values for a named metric.
func (p *PipelineAPI) GetMetricHistory(name string, n int) []float64 {
	ts := p.app.pipeline.GetTimeSeries(name)
	if ts == nil {
		return nil
	}
	values := ts.Values()
	if n > 0 && len(values) > n {
		return values[len(values)-n:]
	}
	return values
}

// GetMetricHistoryWithTimestamps returns the last n values with timestamps for a named metric.
func (p *PipelineAPI) GetMetricHistoryWithTimestamps(name string, n int) []common.DataPoint {
	ts := p.app.pipeline.GetTimeSeries(name)
	if ts == nil {
		return nil
	}
	points := ts.DataPoints()
	if n > 0 && len(points) > n {
		points = points[len(points)-n:]
	}
	return points
}

// GetForecast returns predicted values for a named metric.
func (p *PipelineAPI) GetForecast(name string, steps int) []float64 {
	return p.app.pipeline.GetForecast(name, steps)
}

// GetTrend returns the current trend for a named metric.
func (p *PipelineAPI) GetTrend(name string) TrendInfo {
	trend := p.app.pipeline.GetTrend(name)
	return TrendInfo{
		Direction:   trendDirectionString(trend.Direction),
		ChangePct:   trend.ChangePct,
		Slope:       trend.Slope,
		Correlation: trend.Correlation,
	}
}

// GetWindowStats returns rolling statistics for a named metric.
func (p *PipelineAPI) GetWindowStats(name string) StatsInfo {
	stats := p.app.pipeline.GetWindowStats(name)
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
	mf := p.app.pipeline.GetMetricWithForecast(name)
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

// MetricDef describes a tracked metric.
type MetricDef struct {
	Name  string `json:"name"`
	Unit  string `json:"unit"`
	Label string `json:"label"`
}

// ClearPipeline resets all stored data and forecasts.
func (p *PipelineAPI) ClearPipeline() {
	p.app.pipeline.Clear()
}

// UpdateSettings updates the pipeline configuration and tick interval.
func (p *PipelineAPI) UpdateSettings(intervalMs int, capacity int) {
	newInterval := time.Duration(intervalMs) * time.Millisecond
	cfg := p.app.pipeline.Config()
	cfg.TickInterval = newInterval
	if capacity > 0 {
		cfg.Capacity = capacity
	}
	p.app.pipeline.UpdateConfig(cfg)

	// Signal the tick loop to use the new interval
	select {
	case p.app.tickIntervalCh <- newInterval:
	default:
		// Channel full, loop will pick up next time or ignore if same
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
