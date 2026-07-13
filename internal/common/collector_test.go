package common

import (
	"context"
	"testing"
	"time"
)

func TestCollectorInfo(t *testing.T) {
	c := &struct{ Collector }{&testCollector{
		info: CollectorInfo{
			ID:              CollectorCPU,
			Name:            "CPU",
			Description:     "test",
			DefaultInterval: 3 * time.Second,
			DefaultEnabled:  true,
		},
		collect: func(ctx context.Context) ([]MetricSample, error) {
			return []MetricSample{{Name: "cpu.percent", Unit: "%", Value: 50}}, nil
		},
	}}

	info := c.Info()
	if info.ID != CollectorCPU {
		t.Fatalf("expected CPU, got %s", info.ID)
	}
	if !info.DefaultEnabled {
		t.Fatal("expected enabled")
	}
	if info.DefaultInterval != 3*time.Second {
		t.Fatalf("expected 3s, got %v", info.DefaultInterval)
	}

	samples, err := c.Collect(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(samples) != 1 || samples[0].Value != 50 {
		t.Fatalf("unexpected samples: %+v", samples)
	}
}

func TestMetricSample(t *testing.T) {
	s := MetricSample{Name: "test.metric", Unit: "count", Value: 123}
	if s.Name != "test.metric" || s.Unit != "count" || s.Value != 123 {
		t.Fatalf("unexpected MetricSample: %+v", s)
	}
}
