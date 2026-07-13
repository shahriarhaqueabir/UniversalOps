package common

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	metricsOnce     sync.Once
	metricsPort     = 9210
	cpuGauge        = prometheus.NewGauge(prometheus.GaugeOpts{Name: "opsforall_cpu_percent", Help: "Current CPU usage percentage"})
	memGauge        = prometheus.NewGauge(prometheus.GaugeOpts{Name: "opsforall_memory_percent", Help: "Current memory usage percentage"})
	diskGauge       = prometheus.NewGauge(prometheus.GaugeOpts{Name: "opsforall_disk_percent", Help: "Current disk usage percentage"})
	procGauge       = prometheus.NewGauge(prometheus.GaugeOpts{Name: "opsforall_process_count", Help: "Number of running processes"})
	alertsGauge     = prometheus.NewGauge(prometheus.GaugeOpts{Name: "opsforall_alerts_total", Help: "Total number of active alerts"})
	pipelineTickCnt = prometheus.NewCounter(prometheus.CounterOpts{Name: "opsforall_pipeline_ticks_total", Help: "Number of pipeline collection ticks"})
)

// InitMetricsExporter registers Prometheus metrics and starts an HTTP server
// on the specified port. The server is started only once (idempotent).
func InitMetricsExporter(port int) {
	metricsOnce.Do(func() {
		if port > 0 {
			metricsPort = port
		}

		prometheus.MustRegister(cpuGauge, memGauge, diskGauge, procGauge, alertsGauge, pipelineTickCnt)

		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":    "ok",
				"app":       "OpsForAll",
				"version":   "1.3.0",
				"timestamp": time.Now().UTC().Format(time.RFC3339),
				"port":      metricsPort,
			})
		})

		addr := fmt.Sprintf(":%d", metricsPort)
		go func() {
			defer RecoverPanic()
			LogInfo("Metrics exporter listening on %s", addr)
			if err := http.ListenAndServe(addr, mux); err != nil {
				LogWarn("Metrics exporter stopped: %v", err)
			}
		}()
	})
}

// SetCPUMetric sets the current CPU usage gauge.
func SetCPUMetric(val float64) {
	cpuGauge.Set(val)
}

// SetMemoryMetric sets the current memory usage gauge.
func SetMemoryMetric(val float64) {
	memGauge.Set(val)
}

// SetDiskMetric sets the current disk usage gauge.
func SetDiskMetric(val float64) {
	diskGauge.Set(val)
}

// SetProcessCountMetric sets the process count gauge.
func SetProcessCountMetric(val float64) {
	procGauge.Set(val)
}

// SetAlertCountMetric sets the alert count gauge.
func SetAlertCountMetric(val float64) {
	alertsGauge.Set(val)
}

// IncPipelineTick increments the pipeline tick counter.
func IncPipelineTick() {
	pipelineTickCnt.Inc()
}

// MetricsExporterPort returns the configured metrics port.
func MetricsExporterPort() string {
	return strconv.Itoa(metricsPort)
}
