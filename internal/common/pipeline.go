package common

import (
	"sync"
	"time"
)

// ── CollectionConfig ────────────────────────────────────────────────────────

// CollectionConfig controls how the data pipeline collects and stores metrics.
type CollectionConfig struct {
	// Capacity is the number of samples retained per time series (ring buffer).
	// Default 240 = 4 min at 1 s tick interval.
	Capacity int

	// TickInterval is the nominal collection interval.
	TickInterval time.Duration

	// ForecastSteps is the default number of steps to predict ahead.
	ForecastSteps int

	// ForecastWindow is the lookback window (in samples) used by the forecast engine.
	// If zero, ForecastWindow defaults to Capacity/4.
	ForecastWindow int
}

// DefaultCollectionConfig returns a sensible configuration.
func DefaultCollectionConfig() CollectionConfig {
	return CollectionConfig{
		Capacity:       240,
		TickInterval:   1 * time.Second,
		ForecastSteps:  12,
		ForecastWindow: 60,
	}
}

// ── MetricForecast ──────────────────────────────────────────────────────────

// MetricForecast bundles the raw time series, the forecast projection,
// the detected trend, and rolling statistics for a single metric.
type MetricForecast struct {
	Name      string      `json:"name"`
	Unit      string      `json:"unit"`
	Values    []float64   `json:"values"`
	Forecast  []float64   `json:"forecast"`
	Trend     TrendInfo   `json:"trend"`
	Stats     WindowStats `json:"stats"`
	LastValue float64     `json:"last_value"`
	LastTime  time.Time   `json:"last_time"`
}

// ── DataPipeline ────────────────────────────────────────────────────────────

// DataPipeline is the central manager for metric time-series ingestion,
// forecasting, and retrieval. It fans every PushMetric call into both the
// TimeSeriesStore and the corresponding ForecastEngine.
type DataPipeline struct {
	mu       sync.RWMutex
	store    *TimeSeriesStore
	forecast map[string]*ForecastEngine
	config   CollectionConfig
}

// NewDataPipeline creates a pipeline with the given configuration.
func NewDataPipeline(cfg CollectionConfig) *DataPipeline {
	if cfg.Capacity <= 0 {
		cfg.Capacity = DefaultCollectionConfig().Capacity
	}
	if cfg.ForecastSteps <= 0 {
		cfg.ForecastSteps = DefaultCollectionConfig().ForecastSteps
	}
	if cfg.ForecastWindow <= 0 {
		cfg.ForecastWindow = cfg.Capacity / 4
		if cfg.ForecastWindow < 10 {
			cfg.ForecastWindow = 10
		}
	}
	if cfg.TickInterval <= 0 {
		cfg.TickInterval = DefaultCollectionConfig().TickInterval
	}

	return &DataPipeline{
		store:    NewTimeSeriesStore(cfg.Capacity),
		forecast: make(map[string]*ForecastEngine),
		config:   cfg,
	}
}

// ── Ingestion ───────────────────────────────────────────────────────────────

// PushMetric pushes a single value to both the time-series store and the
// forecast engine for the named metric. The unit is used only when creating
// a new time series (subsequent calls with the same name reuse the original
// unit).
func (dp *DataPipeline) PushMetric(name, unit string, value float64) {
	dp.mu.Lock()
	defer dp.mu.Unlock()

	now := time.Now()
	dp.store.Push(name, unit, now, value)

	fe, ok := dp.forecast[name]
	if !ok {
		fe = NewForecastEngine(dp.config.ForecastWindow)
		dp.forecast[name] = fe
	}
	fe.Push(value)

	// Persist to database
	if s := GetStorage(); s != nil {
		s.InsertMetric(name, unit, value)
	}
}

// ── Retrieval ───────────────────────────────────────────────────────────────

// GetTimeSeries returns the raw time series for the named metric, or nil if
// no data has been pushed yet.
func (dp *DataPipeline) GetTimeSeries(name string) *TimeSeries {
	dp.mu.RLock()
	defer dp.mu.RUnlock()

	// Since the store auto-creates series on Get, but we only want to return
	// ones that have data, retrieve from the underlying map.
	if ts := dp.store.Get(name, ""); ts != nil && ts.Count() > 0 {
		return ts
	}
	return nil
}

// getForecastEngine is a helper that returns the forecast engine for a metric.
// Caller must hold at least a read lock.
func (dp *DataPipeline) getForecastEngine(name string) *ForecastEngine {
	fe, ok := dp.forecast[name]
	if !ok {
		return nil
	}
	return fe
}

// GetForecast returns a forecast for the named metric. If steps <= 0 the
// default from CollectionConfig is used. Returns nil when there is
// insufficient data.
func (dp *DataPipeline) GetForecast(name string, steps int) []float64 {
	dp.mu.RLock()
	defer dp.mu.RUnlock()

	fe := dp.getForecastEngine(name)
	if fe == nil {
		return nil
	}
	if steps <= 0 {
		steps = dp.config.ForecastSteps
	}
	return fe.PredictSeries(steps)
}

// GetTrend returns the current trend for the named metric.
func (dp *DataPipeline) GetTrend(name string) TrendInfo {
	dp.mu.RLock()
	defer dp.mu.RUnlock()

	fe := dp.getForecastEngine(name)
	if fe == nil {
		return TrendInfo{}
	}
	return fe.DetectTrend()
}

// GetWindowStats returns rolling statistics over the current time-series
// window for the named metric.
func (dp *DataPipeline) GetWindowStats(name string) WindowStats {
	dp.mu.RLock()
	defer dp.mu.RUnlock()

	ts := dp.store.Get(name, "")
	if ts == nil || ts.Count() == 0 {
		return WindowStats{}
	}
	return ComputeWindowStats(ts.Values())
}

// GetMetricWithForecast returns a combined struct with raw values, forecast
// projection, trend, and statistics – suitable for chart rendering.
func (dp *DataPipeline) GetMetricWithForecast(name string) MetricForecast {
	dp.mu.RLock()
	defer dp.mu.RUnlock()

	ts := dp.store.Get(name, "")
	fe := dp.getForecastEngine(name)

	mf := MetricForecast{Name: name}
	if ts != nil {
		mf.Unit = ts.Unit
		mf.LastValue = ts.Last()
		mf.Values = ts.Values()
		// Capture last timestamp from data points
		if pts := ts.DataPoints(); len(pts) > 0 {
			mf.LastTime = pts[len(pts)-1].Time
		}
		if ts.Count() > 0 {
			mf.Stats = ComputeWindowStats(mf.Values)
		}
	}

	if fe != nil {
		mf.Trend = fe.DetectTrend()
		mf.Forecast = fe.PredictSeries(dp.config.ForecastSteps)
	} else if len(mf.Values) >= 2 {
		// Even without a dedicated forecast engine, compute a basic trend
		fe2 := NewForecastEngine(dp.config.ForecastWindow)
		for _, v := range mf.Values {
			fe2.Push(v)
		}
		mf.Trend = fe2.DetectTrend()
		mf.Forecast = fe2.PredictSeries(dp.config.ForecastSteps)
	}

	return mf
}

// AllSeries returns a snapshot of every tracked time series.
func (dp *DataPipeline) AllSeries() map[string]*TimeSeries {
	dp.mu.RLock()
	defer dp.mu.RUnlock()

	return dp.store.Series()
}

// ── Metadata ────────────────────────────────────────────────────────────────

// Config returns a copy of the pipeline configuration.
func (dp *DataPipeline) Config() CollectionConfig {
	dp.mu.RLock()
	defer dp.mu.RUnlock()
	return dp.config
}

// UpdateConfig updates the pipeline configuration.
func (dp *DataPipeline) UpdateConfig(cfg CollectionConfig) {
	dp.mu.Lock()
	defer dp.mu.Unlock()
	dp.config = cfg
}

// NumSeries returns the number of distinct metrics being tracked.
func (dp *DataPipeline) NumSeries() int {
	dp.mu.RLock()
	defer dp.mu.RUnlock()
	return len(dp.store.Series())
}

// Clear resets all stored data and forecast engines.
func (dp *DataPipeline) Clear() {
	dp.mu.Lock()
	defer dp.mu.Unlock()

	dp.store.ClearAll()
	dp.forecast = make(map[string]*ForecastEngine)
}
