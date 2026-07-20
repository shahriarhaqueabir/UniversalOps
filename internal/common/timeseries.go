package common

import (
	"math"
	"sort"
	"sync"
	"time"
)

// DataPoint is a timestamped value.
type DataPoint struct {
	Time  time.Time
	Value float64
}

// TimeSeries is a ring buffer of timestamped values.
type TimeSeries struct {
	Name    string
	Unit    string
	data    []DataPoint
	maxSize int
	head    int
	count   int
	mu      sync.RWMutex

	// Performance Cache: Prevents re-sorting the whole slice every 2s
	statsCache     WindowStats
	statsDirty     bool
	lastValuesView []float64 // Cached chronological view

	// Incremental Stats
	runningSum float64
}

// NewTimeSeries creates a ring-buffer time series with the given capacity.
func NewTimeSeries(name, unit string, capacity int) *TimeSeries {
	return &TimeSeries{
		Name:       name,
		Unit:       unit,
		data:       make([]DataPoint, capacity),
		maxSize:    capacity,
		statsDirty: true,
	}
}

// Push adds a data point, overwriting the oldest if full.
func (ts *TimeSeries) Push(t time.Time, v float64) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	// Update running sum
	if ts.count == ts.maxSize {
		ts.runningSum -= ts.data[ts.head].Value
	}
	ts.runningSum += v

	ts.data[ts.head] = DataPoint{Time: t, Value: v}
	ts.head = (ts.head + 1) % ts.maxSize
	if ts.count < ts.maxSize {
		ts.count++
	}
	ts.statsDirty = true
	ts.lastValuesView = nil // Invalidate view
}

// Values returns the stored values in chronological order.
// PERFORMANCE: Uses a cached view if the series hasn't changed.
func (ts *TimeSeries) Values() []float64 {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.lastValuesView != nil {
		return ts.lastValuesView
	}

	n := ts.count
	if n == 0 {
		return nil
	}
	out := make([]float64, n)
	start := (ts.head - n + ts.maxSize) % ts.maxSize
	for i := 0; i < n; i++ {
		out[i] = ts.data[(start+i)%ts.maxSize].Value
	}
	ts.lastValuesView = out
	return out
}

// ... (skipping some lines) ...

// GetStats returns rolling statistics over the current window.
// PERFORMANCE: Uses a cached calculation if the series hasn't changed.
func (ts *TimeSeries) GetStats() WindowStats {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if !ts.statsDirty {
		return ts.statsCache
	}

	n := ts.count
	if n == 0 {
		return WindowStats{}
	}

	// 1. Compute Avg in O(1)
	avg := ts.runningSum / float64(n)

	// 2. We still need a sort for percentiles, but we can avoid one allocation
	// if we are careful. However, ComputeWindowStats currently handles this.

	// We need a chronological view to sort
	var values []float64
	if ts.lastValuesView != nil {
		values = ts.lastValuesView
	} else {
		values = make([]float64, n)
		start := (ts.head - n + ts.maxSize) % ts.maxSize
		for i := 0; i < n; i++ {
			values[i] = ts.data[(start+i)%ts.maxSize].Value
		}
		ts.lastValuesView = values
	}

	ts.statsCache = ComputeWindowStats(values)
	ts.statsCache.Avg = avg // Use the more precise running sum
	ts.statsDirty = false
	return ts.statsCache
}

// DataPoints returns all stored points in chronological order.
func (ts *TimeSeries) DataPoints() []DataPoint {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	n := ts.count
	if n == 0 {
		return nil
	}
	out := make([]DataPoint, n)
	start := (ts.head - n + ts.maxSize) % ts.maxSize
	for i := 0; i < n; i++ {
		out[i] = ts.data[(start+i)%ts.maxSize]
	}
	return out
}

// Last returns the most recent value, or 0 if empty.
func (ts *TimeSeries) Last() float64 {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	if ts.count == 0 {
		return 0
	}
	idx := (ts.head - 1 + ts.maxSize) % ts.maxSize
	return ts.data[idx].Value
}

// Count returns the number of stored points.
func (ts *TimeSeries) Count() int {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return ts.count
}

// Capacity returns the maximum capacity.
func (ts *TimeSeries) Capacity() int {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return ts.maxSize
}

// Clear resets the series.
func (ts *TimeSeries) Clear() {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.head = 0
	ts.count = 0
	ts.runningSum = 0
	ts.statsDirty = true
	ts.lastValuesView = nil
}

// WindowStats holds rolling window statistics.
type WindowStats struct {
	Min   float64
	Max   float64
	Avg   float64
	P50   float64
	P95   float64
	P99   float64
	Count int
}

// ComputeWindowStats calculates statistics over the values.
func ComputeWindowStats(values []float64) WindowStats {
	if len(values) == 0 {
		return WindowStats{}
	}

	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)

	n := len(sorted)
	sum := 0.0
	for _, v := range sorted {
		sum += v
	}

	return WindowStats{
		Min:   sorted[0],
		Max:   sorted[n-1],
		Avg:   sum / float64(n),
		P50:   percentile(sorted, 50),
		P95:   percentile(sorted, 95),
		P99:   percentile(sorted, 99),
		Count: n,
	}
}

func percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	idx := int(math.Ceil(float64(len(sorted))*p/100.0)) - 1
	if idx < 0 {
		idx = 0
	}
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}

// TimeSeriesStore manages multiple named time series.
type TimeSeriesStore struct {
	series map[string]*TimeSeries
	cap    int
	mu     sync.RWMutex
}

// NewTimeSeriesStore creates a store where each series has the given capacity.
func NewTimeSeriesStore(capacity int) *TimeSeriesStore {
	return &TimeSeriesStore{
		series: make(map[string]*TimeSeries),
		cap:    capacity,
	}
}

// Get returns the named series, creating it if needed.
func (s *TimeSeriesStore) Get(name, unit string) *TimeSeries {
	s.mu.RLock()
	ts, ok := s.series[name]
	s.mu.RUnlock()
	if ok {
		return ts
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	// Check again in case it was created between locks
	if ts, ok := s.series[name]; ok {
		return ts
	}
	ts = NewTimeSeries(name, unit, s.cap)
	s.series[name] = ts
	return ts
}

// Push adds a data point to the named series.
func (s *TimeSeriesStore) Push(name, unit string, t time.Time, v float64) {
	s.Get(name, unit).Push(t, v)
}

// Series returns a map of all series. Note: the map is a shallow copy
// but the TimeSeries pointers are the same.
func (s *TimeSeriesStore) Series() map[string]*TimeSeries {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make(map[string]*TimeSeries, len(s.series))
	for k, v := range s.series {
		out[k] = v
	}
	return out
}

// Exists returns true if the named series already exists (non-creating).
func (s *TimeSeriesStore) Exists(name string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.series[name]
	return ok
}

// NumSeries returns the number of series in the store.
func (s *TimeSeriesStore) NumSeries() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.series)
}

// ClearAll resets all series.
func (s *TimeSeriesStore) ClearAll() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.series = make(map[string]*TimeSeries)
}
