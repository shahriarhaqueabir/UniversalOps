package app

// ── Event Types ──────────────────────────────────────────────────────────────
//
// These structs are emitted to the frontend via runtime.EventsEmit.
// All fields must be JSON-serializable (no channels, no complex types).

// MetricsEvent is emitted on each tick with the latest system snapshot.
type MetricsEvent struct {
	CPU         GaugeMetric   `json:"cpu"`
	Memory      GaugeMetric   `json:"memory"`
	Disk        GaugeMetric   `json:"disk"`
	Network     NetworkMetric `json:"network"`
	Processes   int           `json:"processes"`
	Connections int           `json:"connections"`
}

// AlertEvent is emitted when an alert fires or is resolved.
type AlertEvent struct {
	Action     string    `json:"action"` // "fired" or "resolved"
	Alert      AlertInfo `json:"alert"`
	AlertCount int       `json:"alert_count"`
}

// LogEvent is emitted when a new log entry is written.
type LogEvent struct {
	Level string `json:"level"` // "INFO", "WARN", "ERROR"
	Line  string `json:"line"`
	Time  string `json:"time"`
}

// PipelineEvent is emitted when pipeline status changes.
type PipelineEvent struct {
	Status      string `json:"status"` // "collecting", "idle", "error"
	SeriesCount int    `json:"series_count"`
	Error       string `json:"error,omitempty"`
}

// Event names used with runtime.EventsEmit.
const (
	EventMetrics  = "metrics"  // payload: MetricsEvent
	EventAlert    = "alert"    // payload: AlertEvent
	EventLog      = "log"      // payload: LogEvent
	EventPipeline = "pipeline" // payload: PipelineEvent
	EventCmdLine  = "cmd:line" // payload: string (live command output)
	EventCmdDone  = "cmd:done" // payload: string (command id)
)
