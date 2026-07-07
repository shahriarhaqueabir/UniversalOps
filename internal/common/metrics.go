package common

// ── Standard Metric Identifiers ─────────────────────────────────────────────

const (
	MetricCPU     = "cpu.percent"
	MetricMem     = "memory.percent"
	MetricDisk    = "disk.percent"
	MetricNetRX   = "network.rx.rate"
	MetricNetTX   = "network.tx.rate"
	MetricProcCnt = "process.count"
	MetricConnCnt = "connection.count"
	MetricCPUTemp = "cpu.temperature"
)

// MetricDef describes a trackable metric.
type MetricDef struct {
	Name  string // canonical identifier, e.g. "cpu.percent"
	Unit  string // display unit, e.g. "%", "bps", "count"
	Label string // human-readable short name, e.g. "CPU"
}

// DefaultMetrics is the list of all metrics tracked by the pipeline.
var DefaultMetrics = []MetricDef{
	{Name: MetricCPU, Unit: "%", Label: "CPU"},
	{Name: MetricMem, Unit: "%", Label: "Memory"},
	{Name: MetricDisk, Unit: "%", Label: "Disk"},
	{Name: MetricNetRX, Unit: "bps", Label: "Net RX"},
	{Name: MetricNetTX, Unit: "bps", Label: "Net TX"},
	{Name: MetricProcCnt, Unit: "count", Label: "Processes"},
	{Name: MetricConnCnt, Unit: "count", Label: "Connections"},
	{Name: MetricCPUTemp, Unit: "°C", Label: "CPU Temp"},
}

// DefaultAlertThresholds returns standard warning/critical thresholds
// used by the alert engine.
func DefaultAlertThresholds() map[string]struct {
	Warn float64
	Crit float64
	Unit string
} {
	return map[string]struct {
		Warn float64
		Crit float64
		Unit string
	}{
		MetricCPU:     {Warn: 70, Crit: 90, Unit: "%"},
		MetricMem:     {Warn: 80, Crit: 90, Unit: "%"},
		MetricDisk:    {Warn: 85, Crit: 95, Unit: "%"},
		MetricCPUTemp: {Warn: 70, Crit: 85, Unit: "°C"},
	}
}
