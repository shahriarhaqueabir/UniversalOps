package common

import (
	"context"
	"sync"
	"testing"
	"time"
)

type recordingCollector struct {
	mu       sync.Mutex
	info     CollectorInfo
	calls    []time.Time
	results  func() ([]MetricSample, error)
}

func (c *recordingCollector) Info() CollectorInfo { return c.info }
func (c *recordingCollector) Collect(ctx context.Context) ([]MetricSample, error) {
	c.mu.Lock()
	c.calls = append(c.calls, time.Now())
	c.mu.Unlock()
	if c.results != nil {
		return c.results()
	}
	return []MetricSample{{Name: "test", Unit: "x", Value: 1}}, nil
}
func (c *recordingCollector) CallCount() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.calls)
}

func TestSchedulerStartStop(t *testing.T) {
	reg := NewCollectorRegistry()
	c := &recordingCollector{
		info: CollectorInfo{
			ID:              "test_collector",
			Name:            "Test",
			DefaultInterval: 50 * time.Millisecond,
			DefaultEnabled:  true,
		},
	}
	reg.Register(c)

	pipeline := NewDataPipeline(DefaultCollectionConfig())
	sched := NewCollectorScheduler(reg, pipeline)
	sched.Start()

	// Let it collect a couple times
	time.Sleep(120 * time.Millisecond)

	sched.Stop(time.Second)

	if c.CallCount() < 1 {
		t.Fatalf("expected at least 1 collect call, got %d", c.CallCount())
	}
}

func TestSchedulerDisabledCollector(t *testing.T) {
	reg := NewCollectorRegistry()
	c := &recordingCollector{
		info: CollectorInfo{
			ID:              "disabled_test",
			Name:            "Disabled",
			DefaultInterval: 50 * time.Millisecond,
			DefaultEnabled:  false,
		},
	}
	reg.Register(c)

	pipeline := NewDataPipeline(DefaultCollectionConfig())
	sched := NewCollectorScheduler(reg, pipeline)
	sched.Start()

	time.Sleep(100 * time.Millisecond)

	sched.Stop(time.Second)

	if c.CallCount() > 0 {
		t.Fatalf("expected 0 calls for disabled collector, got %d", c.CallCount())
	}
}

func TestSchedulerStopTimeout(t *testing.T) {
	reg := NewCollectorRegistry()
	c := &recordingCollector{
		info: CollectorInfo{
			ID:              "timeout_test",
			DefaultInterval: time.Millisecond,
			DefaultEnabled:  true,
		},
	}
	reg.Register(c)

	pipeline := NewDataPipeline(DefaultCollectionConfig())
	sched := NewCollectorScheduler(reg, pipeline)
	sched.Start()

	// Stop with a very short timeout — should not hang
	done := make(chan struct{})
	go func() {
		sched.Stop(time.Millisecond)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Stop with timeout hung")
	}
}

func TestSchedulerPipelineIntegration(t *testing.T) {
	reg := NewCollectorRegistry()
	c := &recordingCollector{
		info: CollectorInfo{
			ID:              "integrated",
			DefaultInterval: 30 * time.Millisecond,
			DefaultEnabled:  true,
		},
	}
	reg.Register(c)

	pipeline := NewDataPipeline(DefaultCollectionConfig())
	sched := NewCollectorScheduler(reg, pipeline)
	sched.Start()

	time.Sleep(100 * time.Millisecond)
	sched.Stop(time.Second)

	// Pipeline should have the metric
	mf := pipeline.GetMetricWithForecast("test")
	if mf.LastValue != 1 {
		t.Fatalf("expected pipeline value 1, got %f", mf.LastValue)
	}
}
