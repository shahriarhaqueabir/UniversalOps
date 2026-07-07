package common

import (
	"math"
	"testing"
	"time"
)

func TestNewDataPipeline(t *testing.T) {
	dp := NewDataPipeline(DefaultCollectionConfig())
	if dp == nil {
		t.Fatal("NewDataPipeline returned nil")
	}
	if dp.NumSeries() != 0 {
		t.Errorf("expected 0 series, got %d", dp.NumSeries())
	}
}

func TestPipelinePushAndGetTimeSeries(t *testing.T) {
	dp := NewDataPipeline(CollectionConfig{Capacity: 10})
	dp.PushMetric("cpu.percent", "%", 42.5)
	dp.PushMetric("cpu.percent", "%", 43.0)
	dp.PushMetric("cpu.percent", "%", 44.0)

	ts := dp.GetTimeSeries("cpu.percent")
	if ts == nil {
		t.Fatal("GetTimeSeries returned nil")
	}
	if ts.Count() != 3 {
		t.Errorf("expected count 3, got %d", ts.Count())
	}
	if ts.Last() != 44.0 {
		t.Errorf("expected last 44.0, got %f", ts.Last())
	}
}

func TestPipelinePushAndGetForecast(t *testing.T) {
	dp := NewDataPipeline(CollectionConfig{
		Capacity:       20,
		ForecastSteps:  3,
		ForecastWindow: 10,
	})

	// Push a steadily rising sequence
	for i := 0; i < 10; i++ {
		dp.PushMetric("cpu.percent", "%", float64(50+i))
	}

	forecast := dp.GetForecast("cpu.percent", 3)
	if forecast == nil {
		t.Fatal("GetForecast returned nil")
	}
	if len(forecast) != 3 {
		t.Errorf("expected 3 forecast values, got %d", len(forecast))
	}

	// With a rising trend, predictions should be > last value
	lastVal := dp.GetTimeSeries("cpu.percent").Last()
	for i, v := range forecast {
		if v < lastVal {
			t.Errorf("forecast[%d] = %.1f < lastValue %.1f; expected rising", i, v, lastVal)
		}
	}
}

func TestPipelineGetTrend(t *testing.T) {
	dp := NewDataPipeline(CollectionConfig{Capacity: 10, ForecastWindow: 10})

	// Falling trend
	for i := 0; i < 5; i++ {
		dp.PushMetric("memory.percent", "%", float64(90-i*5))
	}

	trend := dp.GetTrend("memory.percent")
	if trend.Direction != TrendFalling {
		t.Errorf("expected TrendFalling, got %v", trend.Direction)
	}
	if trend.Slope >= 0 {
		t.Errorf("expected negative slope, got %f", trend.Slope)
	}
}

func TestPipelineGetWindowStats(t *testing.T) {
	dp := NewDataPipeline(CollectionConfig{Capacity: 10})
	dp.PushMetric("test.metric", "count", 10)
	dp.PushMetric("test.metric", "count", 20)
	dp.PushMetric("test.metric", "count", 30)

	stats := dp.GetWindowStats("test.metric")
	if stats.Count != 3 {
		t.Errorf("expected count 3, got %d", stats.Count)
	}
	if stats.Min != 10 {
		t.Errorf("expected min 10, got %f", stats.Min)
	}
	if stats.Max != 30 {
		t.Errorf("expected max 30, got %f", stats.Max)
	}
	if math.Abs(stats.Avg-20) > 0.01 {
		t.Errorf("expected avg 20, got %f", stats.Avg)
	}
}

func TestPipelineGetMetricWithForecast(t *testing.T) {
	dp := NewDataPipeline(CollectionConfig{Capacity: 10, ForecastSteps: 3, ForecastWindow: 10})

	for i := 0; i < 5; i++ {
		dp.PushMetric("cpu.percent", "%", float64(50+i*2))
	}

	mf := dp.GetMetricWithForecast("cpu.percent")
	if mf.Name != "cpu.percent" {
		t.Errorf("expected name cpu.percent, got %s", mf.Name)
	}
	if mf.Unit != "%" {
		t.Errorf("expected unit %%, got %s", mf.Unit)
	}
	if len(mf.Values) != 5 {
		t.Errorf("expected 5 values, got %d", len(mf.Values))
	}
	if len(mf.Forecast) != 3 {
		t.Errorf("expected 3 forecast values, got %d", len(mf.Forecast))
	}
	if mf.Trend.Direction != TrendRising {
		t.Errorf("expected TrendRising, got %v", mf.Trend.Direction)
	}
}

func TestPipelineAllSeries(t *testing.T) {
	dp := NewDataPipeline(CollectionConfig{Capacity: 10})
	dp.PushMetric("cpu.percent", "%", 50)
	dp.PushMetric("memory.percent", "%", 60)

	series := dp.AllSeries()
	if len(series) != 2 {
		t.Errorf("expected 2 series, got %d", len(series))
	}
	if _, ok := series["cpu.percent"]; !ok {
		t.Error("expected cpu.percent in series")
	}
	if _, ok := series["memory.percent"]; !ok {
		t.Error("expected memory.percent in series")
	}
}

func TestPipelineClear(t *testing.T) {
	dp := NewDataPipeline(CollectionConfig{Capacity: 10})
	dp.PushMetric("cpu.percent", "%", 50)
	dp.Clear()

	if dp.NumSeries() != 0 {
		t.Errorf("expected 0 series after clear, got %d", dp.NumSeries())
	}
	// Should return nil after clear
	if ts := dp.GetTimeSeries("cpu.percent"); ts != nil {
		t.Error("expected nil time series after clear")
	}
}

func TestPipelineMissingMetric(t *testing.T) {
	dp := NewDataPipeline(DefaultCollectionConfig())

	ts := dp.GetTimeSeries("nonexistent")
	if ts != nil {
		t.Error("expected nil for nonexistent metric")
	}

	forecast := dp.GetForecast("nonexistent", 5)
	if forecast != nil {
		t.Error("expected nil forecast for nonexistent metric")
	}

	trend := dp.GetTrend("nonexistent")
	if trend.Direction != TrendStable {
		t.Errorf("expected TrendStable for nonexistent, got %v", trend.Direction)
	}

	mf := dp.GetMetricWithForecast("nonexistent")
	if len(mf.Values) != 0 {
		t.Error("expected empty values for nonexistent metric")
	}
}

func TestPipelineRingBufferCapacity(t *testing.T) {
	capacity := 3
	dp := NewDataPipeline(CollectionConfig{Capacity: capacity})

	// Push more values than capacity
	for i := 0; i < 10; i++ {
		dp.PushMetric("test", "count", float64(i))
	}

	ts := dp.GetTimeSeries("test")
	if ts == nil {
		t.Fatal("GetTimeSeries returned nil")
	}
	if ts.Count() != capacity {
		t.Errorf("expected count %d, got %d", capacity, ts.Count())
	}

	vals := ts.Values()
	if len(vals) != capacity {
		t.Fatalf("expected %d values, got %d", capacity, len(vals))
	}
	// Last 3 values pushed: 7, 8, 9
	if vals[0] != 7 || vals[1] != 8 || vals[2] != 9 {
		t.Errorf("expected [7,8,9], got %v", vals)
	}
}

func TestPipelineEmptyStats(t *testing.T) {
	dp := NewDataPipeline(DefaultCollectionConfig())

	stats := dp.GetWindowStats("nonexistent")
	if stats.Count != 0 {
		t.Errorf("expected count 0 for empty stats, got %d", stats.Count)
	}
}

func TestPipelineConfigDefaults(t *testing.T) {
	dp := NewDataPipeline(CollectionConfig{})
	cfg := dp.Config()

	if cfg.Capacity <= 0 {
		t.Error("expected positive capacity default")
	}
	if cfg.ForecastSteps <= 0 {
		t.Error("expected positive ForecastSteps default")
	}
	if cfg.ForecastWindow <= 0 {
		t.Error("expected positive ForecastWindow default")
	}
	if cfg.TickInterval <= 0 {
		t.Error("expected positive TickInterval default")
	}
}

func TestPipelineConcurrency(t *testing.T) {
	dp := NewDataPipeline(CollectionConfig{Capacity: 100, ForecastWindow: 20})
	done := make(chan bool, 2)

	// Writer goroutine
	go func() {
		for i := 0; i < 50; i++ {
			dp.PushMetric("cpu.percent", "%", float64(50+i%20))
		}
		done <- true
	}()

	// Reader goroutine
	go func() {
		for i := 0; i < 50; i++ {
			dp.GetMetricWithForecast("cpu.percent")
			dp.GetTrend("cpu.percent")
			dp.AllSeries()
			time.Sleep(time.Microsecond)
		}
		done <- true
	}()

	<-done
	<-done
}

func TestTimeToThreshold(t *testing.T) {
	dp := NewDataPipeline(CollectionConfig{Capacity: 10, ForecastWindow: 10})

	// Rising trend: 50, 52, 54, 56, 58
	for i := 0; i < 5; i++ {
		dp.PushMetric("cpu.percent", "%", float64(50+i*2))
	}

	mf := dp.GetMetricWithForecast("cpu.percent")
	if mf.Trend.Direction != TrendRising {
		t.Errorf("expected rising trend, got %v", mf.Trend.Direction)
	}

	// Forecast values should exist
	if len(mf.Forecast) == 0 {
		t.Error("expected non-empty forecast")
	}
}
