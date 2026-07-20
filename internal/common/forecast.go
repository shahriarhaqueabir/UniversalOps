package common

import (
	"math"
	"sync"
)

// ForecastEngine provides trend prediction using linear regression
// and exponential smoothing.
type ForecastEngine struct {
	values   []float64
	capacity int
	head     int // index to the next insert position
	full     bool
	mu       sync.RWMutex

	// Cache for expensive trend/forecast calculations
	cacheValid bool
	lastTrend  TrendInfo
}

// NewForecastEngine creates a forecast engine with a lookback window.
func NewForecastEngine(capacity int) *ForecastEngine {
	if capacity <= 0 {
		capacity = 10 // fallback
	}
	return &ForecastEngine{
		values:   make([]float64, capacity),
		capacity: capacity,
	}
}

// Push adds a new observed value and invalidates the cache.
func (fe *ForecastEngine) Push(v float64) {
	fe.mu.Lock()
	defer fe.mu.Unlock()
	fe.values[fe.head] = v
	fe.head = (fe.head + 1) % fe.capacity
	if fe.head == 0 {
		fe.full = true
	}
	fe.cacheValid = false
}

// Values returns the current observation window.
func (fe *ForecastEngine) Values() []float64 {
	fe.mu.RLock()
	defer fe.mu.RUnlock()
	size := fe.head
	if fe.full {
		size = fe.capacity
	}
	out := make([]float64, size)

	if !fe.full {
		copy(out, fe.values[:fe.head])
	} else {
		// [head...capacity-1] [0...head-1]
		copy(out, fe.values[fe.head:])
		copy(out[fe.capacity-fe.head:], fe.values[:fe.head])
	}
	return out
}

// DetectTrend performs linear regression on the current values with memoization.
func (fe *ForecastEngine) DetectTrend() TrendInfo {
	fe.mu.RLock()
	if fe.cacheValid {
		trend := fe.lastTrend
		fe.mu.RUnlock()
		return trend
	}
	fe.mu.RUnlock()

	values := fe.Values()
	n := len(values)
	if n < 2 {
		return TrendInfo{Direction: TrendStable}
	}

	// Compute means
	var sumX, sumY float64
	for i, y := range values {
		x := float64(i)
		sumX += x
		sumY += y
	}
	meanX := sumX / float64(n)
	meanY := sumY / float64(n)

	// Compute slope (least squares)
	var num, den float64
	for i, y := range values {
		x := float64(i)
		dx := x - meanX
		dy := y - meanY
		num += dx * dy
		den += dx * dx
	}

	slope := 0.0
	if den != 0 {
		slope = num / den
	}
	intercept := meanY - slope*meanX

	// Correlation (Pearson R)
	var ssRes, ssTot float64
	for i, y := range values {
		x := float64(i)
		pred := slope*x + intercept
		dy := y - meanY
		ssRes += (y - pred) * (y - pred)
		ssTot += dy * dy
	}

	corr := 0.0
	if ssTot != 0 {
		r2 := 1.0 - ssRes/ssTot
		if r2 > 0 {
			corr = math.Sqrt(r2)
			if slope < 0 {
				corr = -corr
			}
		}
	}

	// Percent change (first value to last)
	changePct := 0.0
	first := values[0]
	if first != 0 {
		changePct = (values[n-1] - first) / math.Abs(first) * 100
	}

	dir := TrendStable
	if slope > 0.01 {
		dir = TrendRising
	} else if slope < -0.01 {
		dir = TrendFalling
	}

	res := TrendInfo{
		Direction:   dir,
		ChangePct:   changePct,
		Slope:       slope,
		Intercept:   intercept,
		Correlation: corr,
	}

	fe.mu.Lock()
	fe.lastTrend = res
	fe.cacheValid = true
	fe.mu.Unlock()

	return res
}

// Predict estimates the value at a future step using the linear trend.
func (fe *ForecastEngine) Predict(stepsAhead int) float64 {
	trend := fe.DetectTrend()
	fe.mu.RLock()
	size := fe.head
	if fe.full {
		size = fe.capacity
	}
	fe.mu.RUnlock()
	nextX := float64(size + stepsAhead - 1)
	return trend.Slope*nextX + trend.Intercept
}

// PredictSeries generates a forecast series for the given number of steps.
func (fe *ForecastEngine) PredictSeries(steps int) []float64 {
	trend := fe.DetectTrend()
	fe.mu.RLock()
	size := fe.head
	if fe.full {
		size = fe.capacity
	}
	fe.mu.RUnlock()
	out := make([]float64, steps)
	for i := 0; i < steps; i++ {
		x := float64(size + i)
		out[i] = trend.Slope*x + trend.Intercept
	}
	return out
}

// SimpleMovingAverage computes SMA over the given window.
func SimpleMovingAverage(values []float64, window int) []float64 {
	if len(values) == 0 || window <= 0 {
		return nil
	}
	if window > len(values) {
		window = len(values)
	}
	out := make([]float64, len(values)-window+1)
	for i := range out {
		sum := 0.0
		for j := 0; j < window; j++ {
			sum += values[i+j]
		}
		out[i] = sum / float64(window)
	}
	return out
}

// ExponentialMovingAverage computes EMA with the given smoothing factor (alpha).
// Alpha = 2 / (window + 1) is a common choice.
func ExponentialMovingAverage(values []float64, alpha float64) []float64 {
	if len(values) == 0 {
		return nil
	}
	out := make([]float64, len(values))
	out[0] = values[0]
	for i := 1; i < len(values); i++ {
		out[i] = alpha*values[i] + (1-alpha)*out[i-1]
	}
	return out
}

// TimeToThreshold estimates how many steps until the trend crosses a threshold.
// Returns -1 if the trend never reaches the threshold.
func (fe *ForecastEngine) TimeToThreshold(threshold float64) int {
	trend := fe.DetectTrend()
	if trend.Slope == 0 {
		return -1
	}
	values := fe.Values()
	n := len(values)
	if n == 0 {
		return -1
	}
	lastValue := values[n-1]
	// Only compute if moving toward threshold
	if trend.Slope > 0 && threshold <= lastValue {
		return 0 // already past
	}
	if trend.Slope < 0 && threshold >= lastValue {
		return 0
	}
	steps := int((threshold - lastValue) / trend.Slope)
	if steps < 0 {
		return -1
	}
	return steps
}
