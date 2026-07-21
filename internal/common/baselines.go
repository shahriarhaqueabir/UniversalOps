package common

import (
	"math"
	"sync"
	"time"
)

// DriftInfo describes a detected deviation from the learned baseline.
type DriftInfo struct {
	Metric    string    `json:"metric"`
	Baseline  float64   `json:"baseline"`
	Current   float64   `json:"current"`
	Deviation float64   `json:"deviation"` // In standard deviations (σ)
	Severity  string    `json:"severity"`  // "low", "med", "high"
}

// BaselinesEngine handles the calculation and persistence of long-term
// statistical norms for system metrics.
type BaselinesEngine struct {
	pipeline *DataPipeline
	mu       sync.RWMutex
	cache    map[string]BaselineEntry
}

func NewBaselinesEngine(p *DataPipeline) *BaselinesEngine {
	return &BaselinesEngine{
		pipeline: p,
		cache:    make(map[string]BaselineEntry),
	}
}

// UpdateFromStorage loads all persisted baselines into the engine's cache.
func (e *BaselinesEngine) UpdateFromStorage() {
	s := GetStorage()
	if s == nil {
		return
	}

	// We'll need a ListBaselines method in storage.go to implement this efficiently.
	// For now, we'll focus on the core drift detection logic.
}

// DetectDrift compares the current rolling average vs the persistent baseline.
func (e *BaselinesEngine) DetectDrift(metric string) (*DriftInfo, bool) {
	mf := e.pipeline.GetMetricWithForecast(metric)
	if len(mf.Values) < 10 {
		return nil, false
	}

	s := GetStorage()
	if s == nil {
		return nil, false
	}

	baseline, err := s.GetBaseline(metric)
	if err != nil || baseline == nil {
		// No baseline yet? Use current stats as the first baseline.
		if len(mf.Values) >= 100 {
			_ = s.UpsertBaseline(BaselineEntry{
				Metric: metric,
				Avg:    mf.Stats.Avg,
				StdDev: calculateStdDev(mf.Values, mf.Stats.Avg),
			})
		}
		return nil, false
	}

	currentAvg := mf.Stats.Avg
	diff := math.Abs(currentAvg - baseline.Avg)

	// σ (Sigma) check: How many standard deviations is the current average away from the baseline?
	sigma := baseline.StdDev
	if sigma < 0.1 { sigma = 0.1 } // Prevent div by zero

	deviation := diff / sigma

	// DRIFT THRESHOLD: > 2.5σ is statistically significant "Drift"
	if deviation > 2.5 {
		severity := "low"
		if deviation > 5 {
			severity = "high"
		} else if deviation > 3.5 {
			severity = "med"
		}

		return &DriftInfo{
			Metric:    metric,
			Baseline:  baseline.Avg,
			Current:   currentAvg,
			Deviation: deviation,
			Severity:  severity,
		}, true
	}

	return nil, false
}

// RecalculateBaselines performs a deep-dive into historical data to update norms.
// Intended to be called every 1-4 hours.
func (e *BaselinesEngine) RecalculateBaselines() {
	s := GetStorage()
	if s == nil {
		return
	}

	metrics := []string{MetricCPU, MetricMem, MetricDisk, MetricNetRX, MetricNetTX}
	for _, m := range metrics {
		// Retrieve longer window (e.g., 500 samples)
		history, err := s.GetMetricHistory(m, 500)
		if err != nil || len(history) < 100 {
			continue
		}

		stats := ComputeWindowStats(history)
		_ = s.UpsertBaseline(BaselineEntry{
			Metric: m,
			Avg:    stats.Avg,
			StdDev: calculateStdDev(history, stats.Avg),
		})
	}
}

func calculateStdDev(values []float64, avg float64) float64 {
	if len(values) == 0 {
		return 0
	}
	var sum float64
	for _, v := range values {
		sum += math.Pow(v-avg, 2)
	}
	return math.Sqrt(sum / float64(len(values)))
}
