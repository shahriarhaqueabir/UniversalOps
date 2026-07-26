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

// DetectPipelineAnomalies performs anomaly detection on pipeline metrics.
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
		stddev := (mf.Stats.Max - mf.Stats.Min) / 2
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
