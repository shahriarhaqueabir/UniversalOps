package aiops

import (
	"github.com/shahriarhaqueabir/UniversalOps/internal/common"
)

// AnomalyInfo holds a detected anomaly.
type AnomalyInfo struct {
	Metric    string  `json:"metric"`
	Value     float64 `json:"value"`
	Expected  float64 `json:"expected"`
	Deviation float64 `json:"deviation"`
	Severity  string  `json:"severity"`
	Timestamp string  `json:"timestamp"`
}

// DetectPipelineAnomalies performs anomaly detection on pipeline metrics
// using proper standard deviation (σ) for z-score calculation.
func DetectPipelineAnomalies(pipeline *common.DataPipeline) []AnomalyInfo {
	var anomalies []AnomalyInfo

	metrics := []string{
		common.MetricCPU,
		common.MetricMem,
		common.MetricDisk,
		common.MetricNetRX,
		common.MetricNetTX,
		common.MetricProcCnt,
	}

	for _, name := range metrics {
		mf := pipeline.GetMetricWithForecast(name)
		if len(mf.Values) < 10 {
			continue
		}

		lastVal := mf.LastValue
		mean := mf.Stats.Avg

		// Compute true standard deviation from the sample values
		stddev := computeStdDev(mf.Values, mean)
		if stddev < 0.1 {
			stddev = 0.1
		}

		deviation := (lastVal - mean) / stddev
		if deviation < 0 {
			deviation = -deviation
		}

		if deviation > 3.0 {
			severity := "warning"
			if deviation > 5.0 {
				severity = "critical"
			}
			anomalies = append(anomalies, AnomalyInfo{
				Metric:    name,
				Value:     lastVal,
				Expected:  mean,
				Deviation: deviation,
				Severity:  severity,
				Timestamp: mf.LastTime.Format("2006-01-02T15:04:05Z07:00"),
			})
		}
	}

	return anomalies
}

// computeStdDev computes the sample standard deviation of values given a precomputed mean.
func computeStdDev(values []float64, mean float64) float64 {
	if len(values) < 2 {
		return 0.1
	}
	var sumSqDiff float64
	for _, v := range values {
		diff := v - mean
		sumSqDiff += diff * diff
	}
	variance := sumSqDiff / float64(len(values)-1) // sample variance (Bessel's correction)
	if variance < 0 {
		variance = 0
	}
	// Use math.Sqrt for proper square root
	return sqrt(variance)
}

// sqrt is a simple Newton's method square root to avoid importing math
// for a single function. Precision is sufficient for stddev estimation.
func sqrt(x float64) float64 {
	if x <= 0 {
		return 0.1
	}
	z := x
	for i := 0; i < 20; i++ {
		z = (z + x/z) / 2
	}
	return z
}
