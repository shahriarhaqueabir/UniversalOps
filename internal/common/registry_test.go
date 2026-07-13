package common

import (
	"context"
	"testing"
	"time"
)

type testCollector struct {
	info  CollectorInfo
	collect func(ctx context.Context) ([]MetricSample, error)
}

func (c *testCollector) Info() CollectorInfo { return c.info }
func (c *testCollector) Collect(ctx context.Context) ([]MetricSample, error) { return c.collect(ctx) }

func TestRegistryRegisterAndSnapshot(t *testing.T) {
	r := NewCollectorRegistry()
	r.Register(&testCollector{
		info: CollectorInfo{
			ID:              CollectorCPU,
			Name:            "CPU",
			Description:     "test",
			DefaultInterval: 3 * time.Second,
			DefaultEnabled:  true,
		},
		collect: func(ctx context.Context) ([]MetricSample, error) {
			return []MetricSample{{Name: "cpu.pct", Unit: "%", Value: 42}}, nil
		},
	})

	snap := r.Snapshot()
	if len(snap) != 1 {
		t.Fatalf("expected 1 collector, got %d", len(snap))
	}
	if snap[0].ID != CollectorCPU {
		t.Fatalf("expected CPU, got %s", snap[0].ID)
	}
	if !snap[0].Enabled {
		t.Fatal("expected enabled by default")
	}
	if snap[0].IntervalMs != 3000 {
		t.Fatalf("expected 3000ms interval, got %d", snap[0].IntervalMs)
	}
}

func TestRegistryEnableDisable(t *testing.T) {
	r := NewCollectorRegistry()
	r.Register(&testCollector{
		info: CollectorInfo{
			ID:              CollectorMem,
			Name:            "Memory",
			DefaultInterval: time.Second,
			DefaultEnabled:  false,
		},
		collect: func(ctx context.Context) ([]MetricSample, error) { return nil, nil },
	})

	if r.IsEnabled(CollectorMem) {
		t.Fatal("expected disabled by default")
	}

	if err := r.Enable(CollectorMem); err != nil {
		t.Fatal(err)
	}
	if !r.IsEnabled(CollectorMem) {
		t.Fatal("expected enabled after Enable")
	}

	if err := r.Disable(CollectorMem); err != nil {
		t.Fatal(err)
	}
	if r.IsEnabled(CollectorMem) {
		t.Fatal("expected disabled after Disable")
	}
}

func TestRegistryEnableUnknown(t *testing.T) {
	r := NewCollectorRegistry()
	if err := r.Enable("nonexistent"); err == nil {
		t.Fatal("expected error for unknown collector")
	}
	if err := r.Disable("nonexistent"); err == nil {
		t.Fatal("expected error for unknown collector")
	}
}

func TestRegistrySetInterval(t *testing.T) {
	r := NewCollectorRegistry()
	r.Register(&testCollector{
		info: CollectorInfo{
			ID:              CollectorDisk,
			DefaultInterval: 5 * time.Second,
			DefaultEnabled:  true,
		},
		collect: func(ctx context.Context) ([]MetricSample, error) { return nil, nil },
	})

	if err := r.SetInterval(CollectorDisk, 10*time.Second); err != nil {
		t.Fatal(err)
	}
	if got := r.GetInterval(CollectorDisk); got != 10*time.Second {
		t.Fatalf("expected 10s, got %v", got)
	}

	if err := r.SetInterval(CollectorDisk, 0); err == nil {
		t.Fatal("expected error for zero interval")
	}
	if err := r.SetInterval(CollectorDisk, -time.Second); err == nil {
		t.Fatal("expected error for negative interval")
	}
}

func TestRegistrySetIntervalUnknown(t *testing.T) {
	r := NewCollectorRegistry()
	if err := r.SetInterval("bogus", time.Second); err == nil {
		t.Fatal("expected error for unknown collector")
	}
}

func TestRegistryCollectNow(t *testing.T) {
	r := NewCollectorRegistry()
	r.Register(&testCollector{
		info: CollectorInfo{ID: CollectorTemp, DefaultInterval: time.Second, DefaultEnabled: true},
		collect: func(ctx context.Context) ([]MetricSample, error) {
			return []MetricSample{{Name: "temp", Unit: "C", Value: 36.5}}, nil
		},
	})

	samples, err := r.CollectNow(context.Background(), CollectorTemp)
	if err != nil {
		t.Fatal(err)
	}
	if len(samples) != 1 || samples[0].Value != 36.5 {
		t.Fatalf("unexpected samples: %+v", samples)
	}

	// lastRun should be updated
	snap := r.Snapshot()
	if snap[0].LastRun == "" {
		t.Fatal("expected non-empty last_run after CollectNow")
	}
}

func TestRegistryCollectNowUnknown(t *testing.T) {
	r := NewCollectorRegistry()
	if _, err := r.CollectNow(context.Background(), "bogus"); err == nil {
		t.Fatal("expected error")
	}
}

func TestRegistryAllCollectors(t *testing.T) {
	r := NewCollectorRegistry()
	r.Register(&testCollector{info: CollectorInfo{ID: "a", DefaultEnabled: true}})
	r.Register(&testCollector{info: CollectorInfo{ID: "b", DefaultEnabled: true}})
	all := r.AllCollectors()
	if len(all) != 2 {
		t.Fatalf("expected 2, got %d", len(all))
	}
}

func TestRegistryGetCollector(t *testing.T) {
	r := NewCollectorRegistry()
	r.Register(&testCollector{info: CollectorInfo{ID: CollectorCPU, DefaultEnabled: true}})
	if c := r.GetCollector(CollectorCPU); c == nil {
		t.Fatal("expected collector")
	}
	if c := r.GetCollector("bogus"); c != nil {
		t.Fatal("expected nil for unknown")
	}
}

func TestRegistryGetIntervalZeroForUnknown(t *testing.T) {
	r := NewCollectorRegistry()
	if got := r.GetInterval("bogus"); got != 0 {
		t.Fatalf("expected 0, got %v", got)
	}
}

func TestRegistryMarkRunNoPanic(t *testing.T) {
	r := NewCollectorRegistry()
	r.Register(&testCollector{
		info: CollectorInfo{ID: CollectorCPU, DefaultEnabled: true},
		collect: func(ctx context.Context) ([]MetricSample, error) {
			return []MetricSample{{Name: "x", Unit: "y", Value: 1}}, nil
		},
	})

	_, _ = r.CollectNow(context.Background(), CollectorCPU)
	snap := r.Snapshot()
	if len(snap) != 1 || snap[0].LastRun == "" {
		t.Fatal("last_run should be set after CollectNow")
	}
}
