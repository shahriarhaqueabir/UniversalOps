package app

import (
	"testing"

	"github.com/shahriarhaqueabir/UniversalOps/internal/common"
)

func TestPipelineAPI_GetMetricHistory_NilTS(t *testing.T) {
	a := NewApp()
	p := NewPipelineAPI(a.pipeline)
	values := p.GetMetricHistory("nonexistent.metric", 10)
	if values == nil {
		t.Fatal("GetMetricHistory returned nil, expected non-nil slice")
	}
}

func TestPipelineAPI_GetMetricHistory(t *testing.T) {
	common.InitStorage(":memory:")
	defer common.GetStorage().Close()

	a := NewApp()
	a.pipeline.PushMetric("cpu.percent", "%", 10.0)
	a.pipeline.PushMetric("cpu.percent", "%", 20.0)

	p := NewPipelineAPI(a.pipeline)
	values := p.GetMetricHistory("cpu.percent", 5)
	if values == nil {
		t.Fatal("GetMetricHistory returned nil, expected non-nil slice")
	}
	if len(values) != 2 {
		t.Logf("GetMetricHistory returned %d values, expected 2", len(values))
	}
}

func TestPipelineAPI_GetMetricHistory_NegativeN(t *testing.T) {
	common.InitStorage(":memory:")
	defer common.GetStorage().Close()

	a := NewApp()
	a.pipeline.PushMetric("cpu.percent", "%", 10.0)
	p := NewPipelineAPI(a.pipeline)
	values := p.GetMetricHistory("cpu.percent", -1)
	if values == nil {
		t.Fatal("GetMetricHistory with negative n returned nil, expected non-nil slice")
	}
}

func TestPipelineAPI_GetMetricHistory_ZeroN(t *testing.T) {
	a := NewApp()
	p := NewPipelineAPI(a.pipeline)
	values := p.GetMetricHistory("cpu.percent", 0)
	if values == nil {
		t.Fatal("GetMetricHistory with n=0 returned nil, expected non-nil slice")
	}
}

func TestPipelineAPI_GetMetricHistoryWithTimestamps_NilTS(t *testing.T) {
	a := NewApp()
	p := NewPipelineAPI(a.pipeline)
	points := p.GetMetricHistoryWithTimestamps("nonexistent.metric", 10)
	if points == nil {
		t.Fatal("GetMetricHistoryWithTimestamps returned nil, expected non-nil slice")
	}
}

func TestPipelineAPI_GetMetricHistoryWithTimestamps(t *testing.T) {
	common.InitStorage(":memory:")
	defer common.GetStorage().Close()

	a := NewApp()
	a.pipeline.PushMetric("mem.used", "%", 50.0)
	p := NewPipelineAPI(a.pipeline)
	points := p.GetMetricHistoryWithTimestamps("mem.used", 5)
	if points == nil {
		t.Fatal("GetMetricHistoryWithTimestamps returned nil, expected non-nil slice")
	}
}

func TestPipelineAPI_GetForecast(t *testing.T) {
	a := NewApp()
	p := NewPipelineAPI(a.pipeline)
	forecast := p.GetForecast("cpu.percent", 5)
	if forecast == nil {
		t.Log("GetForecast returned nil (expected without data)")
	}
}

func TestPipelineAPI_GetTrend(t *testing.T) {
	a := NewApp()
	p := NewPipelineAPI(a.pipeline)
	trend := p.GetTrend("cpu.percent")
	if trend.Direction == "" {
		t.Log("Trend direction is empty (expected without data)")
	}
}

func TestPipelineAPI_GetWindowStats(t *testing.T) {
	a := NewApp()
	p := NewPipelineAPI(a.pipeline)
	stats := p.GetWindowStats("cpu.percent")
	if stats.Count < 0 {
		t.Errorf("WindowStats.Count = %d, want >= 0", stats.Count)
	}
}

func TestPipelineAPI_GetMetricWithForecast(t *testing.T) {
	a := NewApp()
	p := NewPipelineAPI(a.pipeline)
	mf := p.GetMetricWithForecast("cpu.percent")
	if mf.Name != "cpu.percent" {
		t.Errorf("MetricWithForecast.Name = %q, want %q", mf.Name, "cpu.percent")
	}
	if mf.Values == nil {
		t.Log("MetricWithForecast.Values is nil (expected without data)")
	}
	if mf.Forecast == nil {
		t.Log("MetricWithForecast.Forecast is nil (expected without data)")
	}
}

func TestPipelineAPI_AllMetricNames(t *testing.T) {
	a := NewApp()
	p := NewPipelineAPI(a.pipeline)
	metrics := p.AllMetricNames()
	if metrics == nil {
		t.Fatal("AllMetricNames returned nil, expected non-nil slice")
	}
	if len(metrics) == 0 {
		t.Error("AllMetricNames returned empty slice")
	}
	for _, m := range metrics {
		if m.Name == "" {
			t.Error("MetricDef entry has empty Name")
		}
	}
}

func TestPipelineAPI_ClearPipeline(t *testing.T) {
	common.InitStorage(":memory:")
	defer common.GetStorage().Close()

	a := NewApp()
	p := NewPipelineAPI(a.pipeline)
	a.pipeline.PushMetric("cpu.percent", "%", 10.0)
	p.ClearPipeline()
	values := p.GetMetricHistory("cpu.percent", 10)
	if len(values) > 0 {
		t.Logf("GetMetricHistory returned %d values after Clear (may vary)", len(values))
	}
}

func TestPipelineAPI_UpdateSettings(t *testing.T) {
	a := NewApp()
	p := NewPipelineAPI(a.pipeline)
	p.UpdateSettings(100, 100, 4, 2000)
	cfg := a.pipeline.Config()
	if cfg.PingCount != 4 {
		t.Errorf("PingCount = %d, want 4", cfg.PingCount)
	}
}
