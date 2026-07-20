package common

import (
	"context"
	"runtime"
	"sync"
	"time"
)

const (
	maxBackoff       = 30 * time.Second
	backoffInitial   = time.Second
	minIntervalGuard = 100 * time.Millisecond
)

type CollectorScheduler struct {
	registry *CollectorRegistry
	pipeline *DataPipeline
	quit     chan struct{}
	wg       sync.WaitGroup
}

func NewCollectorScheduler(registry *CollectorRegistry, pipeline *DataPipeline) *CollectorScheduler {
	return &CollectorScheduler{
		registry: registry,
		pipeline: pipeline,
		quit:     make(chan struct{}),
	}
}

func (s *CollectorScheduler) Start() {
	for _, c := range s.registry.AllCollectors() {
		c := c
		s.wg.Add(1)
		go s.run(c)
	}
}

func (s *CollectorScheduler) Stop(timeout time.Duration) {
	close(s.quit)

	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(timeout):
		LogWarn("CollectorScheduler: %v timeout waiting for goroutines", timeout)
	}
}

func (s *CollectorScheduler) run(c Collector) {
	defer RecoverPanic()
	defer s.wg.Done()

	info := c.Info()
	backoff := backoffInitial
	consecutiveErrors := 0

	for {
		interval := s.registry.GetInterval(info.ID)
		if interval < minIntervalGuard {
			interval = minIntervalGuard
		}

		select {
		case <-s.quit:
			return
		case <-time.After(interval):
		}

		if !s.registry.IsEnabled(info.ID) {
			backoff = backoffInitial
			consecutiveErrors = 0
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), interval)

		// Phase 0 Instrumentation: Start
		var m1, m2 runtime.MemStats
		runtime.ReadMemStats(&m1)
		start := time.Now()

		samples, err := c.Collect(ctx)

		duration := time.Since(start)
		runtime.ReadMemStats(&m2)
		allocs := m2.Mallocs - m1.Mallocs
		bytesAlloc := m2.TotalAlloc - m1.TotalAlloc

		cancel()

		if err != nil {
			consecutiveErrors++
			LogWarn("Collector %q failed (attempt %d): %v", info.ID, consecutiveErrors, err)
			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
			// Sleep during backoff — check quit channel periodically
			t := time.NewTimer(backoff)
			select {
			case <-s.quit:
				t.Stop()
				return
			case <-t.C:
			}
			continue
		}

		// Phase 0 Instrumentation: Log telemetry
		LogDebug("COLLECTOR_METRICS | id=%s | duration=%v | mallocs=%d | total_alloc=%d | samples=%d",
			info.ID, duration, allocs, bytesAlloc, len(samples))

		consecutiveErrors = 0
		backoff = backoffInitial

		for _, sample := range samples {
			s.pipeline.PushMetric(sample.Name, sample.Unit, sample.Value)
		}

		s.registry.markRun(info.ID)
	}
}
