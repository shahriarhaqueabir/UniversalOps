package common

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type managedCollector struct {
	Collector
	enabled  bool
	interval time.Duration
	lastRun  time.Time
}

type CollectorStatus struct {
	ID                CollectorID `json:"id"`
	Name              string      `json:"name"`
	Description       string      `json:"description"`
	Enabled           bool        `json:"enabled"`
	IntervalMs        int         `json:"interval_ms"`
	DefaultIntervalMs int         `json:"default_interval_ms"`
	LastRun           string      `json:"last_run"`
}

type CollectorRegistry struct {
	mu         sync.RWMutex
	collectors map[CollectorID]*managedCollector
}

func NewCollectorRegistry() *CollectorRegistry {
	return &CollectorRegistry{
		collectors: make(map[CollectorID]*managedCollector),
	}
}

func (r *CollectorRegistry) Register(c Collector) {
	r.mu.Lock()
	defer r.mu.Unlock()
	info := c.Info()
	r.collectors[info.ID] = &managedCollector{
		Collector: c,
		enabled:   info.DefaultEnabled,
		interval:  info.DefaultInterval,
	}
}

func (r *CollectorRegistry) Enable(id CollectorID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	mc, ok := r.collectors[id]
	if !ok {
		return fmt.Errorf("collector %q not registered", id)
	}
	mc.enabled = true
	return nil
}

func (r *CollectorRegistry) Disable(id CollectorID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	mc, ok := r.collectors[id]
	if !ok {
		return fmt.Errorf("collector %q not registered", id)
	}
	mc.enabled = false
	return nil
}

func (r *CollectorRegistry) IsEnabled(id CollectorID) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	mc, ok := r.collectors[id]
	return ok && mc.enabled
}

func (r *CollectorRegistry) SetInterval(id CollectorID, d time.Duration) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	mc, ok := r.collectors[id]
	if !ok {
		return fmt.Errorf("collector %q not registered", id)
	}
	if d <= 0 {
		return fmt.Errorf("interval must be positive")
	}
	mc.interval = d
	return nil
}

func (r *CollectorRegistry) GetInterval(id CollectorID) time.Duration {
	r.mu.RLock()
	defer r.mu.RUnlock()
	mc, ok := r.collectors[id]
	if !ok {
		return 0
	}
	return mc.interval
}

func (r *CollectorRegistry) CollectNow(ctx context.Context, id CollectorID) ([]MetricSample, error) {
	r.mu.RLock()
	mc, ok := r.collectors[id]
	r.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("collector %q not registered", id)
	}
	samples, err := mc.Collect(ctx)
	if err != nil {
		return nil, err
	}
	r.mu.Lock()
	mc.lastRun = time.Now()
	r.mu.Unlock()
	return samples, nil
}

func (r *CollectorRegistry) Snapshot() []CollectorStatus {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]CollectorStatus, 0, len(r.collectors))
	for _, mc := range r.collectors {
		info := mc.Info()
		lastRun := ""
		if !mc.lastRun.IsZero() {
			lastRun = mc.lastRun.Format(time.RFC3339)
		}
		out = append(out, CollectorStatus{
			ID:                info.ID,
			Name:              info.Name,
			Description:       info.Description,
			Enabled:           mc.enabled,
			IntervalMs:        int(mc.interval.Milliseconds()), //nolint:gosec
			DefaultIntervalMs: int(info.DefaultInterval.Milliseconds()),
			LastRun:           lastRun,
		})
	}
	return out
}

func (r *CollectorRegistry) AllCollectors() []Collector {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Collector, 0, len(r.collectors))
	for _, mc := range r.collectors {
		out = append(out, mc.Collector)
	}
	return out
}

func (r *CollectorRegistry) GetCollector(id CollectorID) Collector {
	r.mu.RLock()
	defer r.mu.RUnlock()
	mc, ok := r.collectors[id]
	if !ok {
		return nil
	}
	return mc.Collector
}

// markRun updates the lastRun timestamp for a collector.
// Used internally by the scheduler after a successful collect cycle.
func (r *CollectorRegistry) markRun(id CollectorID) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if mc, ok := r.collectors[id]; ok {
		mc.lastRun = time.Now()
	}
}
