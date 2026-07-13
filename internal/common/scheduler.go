package common

import (
	"context"
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
		samples, err := c.Collect(ctx)
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

		consecutiveErrors = 0
		backoff = backoffInitial

		for _, sample := range samples {
			s.pipeline.PushMetric(sample.Name, sample.Unit, sample.Value)
		}

		s.registry.markRun(info.ID)
	}
}
