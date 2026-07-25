//go:build !prometheus

package common

// Stub implementations — compiled when the "prometheus" build tag is not set.
// These no-ops allow the rest of the codebase to call SetCPUMetric, etc.
// without importing the prometheus client library, keeping the dependency
// graph lean for default builds.

// InitMetricsExporter is a no-op when Prometheus support is disabled.
func InitMetricsExporter(int) {}

// SetCPUMetric is a no-op when Prometheus support is disabled.
func SetCPUMetric(float64) {}

// SetMemoryMetric is a no-op when Prometheus support is disabled.
func SetMemoryMetric(float64) {}

// SetDiskMetric is a no-op when Prometheus support is disabled.
func SetDiskMetric(float64) {}

// SetProcessCountMetric is a no-op when Prometheus support is disabled.
func SetProcessCountMetric(float64) {}

// SetAlertCountMetric is a no-op when Prometheus support is disabled.
func SetAlertCountMetric(float64) {}

// IncPipelineTick is a no-op when Prometheus support is disabled.
func IncPipelineTick() {}

// MetricsExporterPort returns the default port string when Prometheus is
// disabled, so callers that display the port in the UI still get a sane value.
func MetricsExporterPort() string { return "9210" }
