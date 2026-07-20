package common

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// TestPipelineEndToEnd simulates what the scheduler + dashboard do:
// push metrics → GetMetricWithForecast → verify values.
func TestPipelineEndToEnd(t *testing.T) {
	// Use realistic config
	cfg := DefaultCollectionConfig()
	cfg.Capacity = 300
	cfg.MaxSeries = 100
	cfg.ForecastWindow = 60
	p := NewDataPipeline(cfg)

	// Push like the scheduler's collectors do
	samples := []MetricSample{
		{Name: MetricCPU, Unit: "%", Value: 45.2},
		{Name: MetricMem, Unit: "%", Value: 67.8},
		{Name: MetricDisk, Unit: "%", Value: 52.1},
		{Name: MetricNetRX, Unit: "bps", Value: 1_500_000},
		{Name: MetricNetTX, Unit: "bps", Value: 800_000},
		{Name: MetricProcCnt, Unit: "count", Value: 187},
		{Name: "cpu.temperature", Unit: "C", Value: 52.0},
		{Name: "system.uptime", Unit: "s", Value: 3600},
		{Name: "load.1m", Unit: "", Value: 1.5},
		{Name: "swap.percent", Unit: "%", Value: 12.3},
		{Name: "gpu.memory.total", Unit: "GB", Value: 8.0},
		{Name: "system.open_fds", Unit: "count", Value: 42},
	}

	for _, s := range samples {
		p.PushMetric(s.Name, s.Unit, s.Value)
	}

	// Also push more data as the scheduler would over time
	for i := 0; i < 5; i++ {
		time.Sleep(5 * time.Millisecond)
		for _, s := range samples {
			p.PushMetric(s.Name, s.Unit, s.Value+float64(i))
		}
	}

	// Now read back - exactly like GetDashboardData does
	metrics := []string{MetricCPU, MetricMem, MetricDisk, MetricNetRX, MetricNetTX, MetricProcCnt}
	for _, name := range metrics {
		mf := p.GetMetricWithForecast(name)
		trend := "stable"
		if mf.Trend.Direction > 0 {
			trend = "rising"
		} else if mf.Trend.Direction < 0 {
			trend = "falling"
		}
		t.Logf("METRIC %s: LastValue=%.2f Unit=%q Values=%d Forecast=%.2f Trend=%s",
			name, mf.LastValue, mf.Unit, len(mf.Values), mf.Forecast, trend)

		if len(mf.Values) == 0 {
			t.Errorf("METRIC %s: zero-length values array (LastValue=%.2f)", name, mf.LastValue)
		}
		if mf.LastValue == 0 && name != MetricNetRX && name != MetricNetTX {
			// Network metrics can be 0 on some systems, CPU/Mem/Disk should not be 0
			t.Errorf("METRIC %s: LastValue is 0, expected >0 after pushing data", name)
		}
	}

	// Check the store directly
	t.Logf("Store has %d series", p.store.NumSeries())
	for name, ts := range p.store.Series() {
		t.Logf("  SERIES %s: count=%d unit=%q last=%.2f", name, ts.Count(), ts.Unit, ts.Last())
	}
}

// TestPipelineConcurrentPush reads while pushing to catch races
func TestPipelineConcurrentPush(t *testing.T) {
	cfg := DefaultCollectionConfig()
	p := NewDataPipeline(cfg)

	var wg sync.WaitGroup
	done := make(chan struct{})

	// Writer goroutines
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				select {
				case <-done:
					return
				default:
				}
				name := fmt.Sprintf("test.metric.%d", id)
				p.PushMetric(name, "units", float64(j))
				time.Sleep(time.Millisecond)
			}
		}(i)
	}

	// Reader goroutines
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 30; j++ {
				select {
				case <-done:
					return
				default:
				}
				mf := p.GetMetricWithForecast("test.metric.0")
				if mf.LastValue < 0 {
					t.Logf("unexpected negative: %f", mf.LastValue)
				}
				time.Sleep(2 * time.Millisecond)
			}
		}()
	}

	wg.Wait()
	close(done)

	t.Logf("Concurrent test completed, store has %d series", p.store.NumSeries())
}

// TestMaxSeriesGuard verifies PushMetric respects MaxSeries limit
func TestMaxSeriesGuard(t *testing.T) {
	cfg := DefaultCollectionConfig()
	cfg.MaxSeries = 3 // very tight limit
	p := NewDataPipeline(cfg)

	// First 3 should succeed
	p.PushMetric("alpha", "x", 1.0)
	p.PushMetric("beta", "x", 2.0)
	p.PushMetric("gamma", "x", 3.0)

	if n := p.store.NumSeries(); n != 3 {
		t.Errorf("expected 3 series, got %d", n)
	}

	// 4th should be silently dropped
	p.PushMetric("delta", "x", 4.0)

	if n := p.store.NumSeries(); n != 3 {
		t.Errorf("expected 3 series after drop, got %d", n)
	}

	// But existing series should still be pushed to
	p.PushMetric("alpha", "x", 1.5)
	mf := p.GetMetricWithForecast("alpha")
	if mf.LastValue != 1.5 {
		t.Errorf("alpha last value: expected 1.5, got %f", mf.LastValue)
	}
}
