package app

import (
	"os"
	"strings"

	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"

	"github.com/shahriarhaqueabir/AllOpsFull/internal/devops"
)

// Logs exposes log management bindings to the frontend.
type Logs struct {
	app *App
}

// NewLogs creates a new Logs facade.
func NewLogs(app *App) *Logs {
	return &Logs{app: app}
}

// GetLogs returns filtered log entries from the hawkward database.
func (l *Logs) GetLogs(level string, since string, n int) []LogEntry {
	if n <= 0 {
		n = 200
	}

	storage := common.GetStorage()
	if storage == nil {
		return nil
	}

	data, err := storage.QueryLogs(level, "", n)
	if err != nil {
		common.LogWarn("QueryLogs failed: %v", err)
		return nil
	}

	out := make([]LogEntry, 0, len(data))
	for _, d := range data {
		out = append(out, LogEntry{
			Timestamp: d.Timestamp,
			Level:     d.Level,
			Module:    d.Module,
			Message:   d.Message,
			Line:      d.Message,
		})
	}
	return out
}

// ExportLogs returns the log contents as a single formatted string.
// format can be "text", "json", or "csv".
func (l *Logs) ExportLogs(format string) string {
	lines, err := devops.TailLog("hawkward-gui.log", 2000)
	if err != nil {
		lines, err = devops.TailLog("hawkward.log", 2000)
		if err != nil {
			return ""
		}
	}

	switch strings.ToLower(format) {
	case "json":
		// Simple JSON array of lines
		var b strings.Builder
		b.WriteString("[")
		for i, line := range lines {
			if i > 0 {
				b.WriteString(",")
			}
			b.WriteString("\"" + strings.ReplaceAll(line, "\"", "\\\"") + "\"")
		}
		b.WriteString("]")
		return b.String()
	case "csv":
		// Simple CSV: line per row
		return strings.Join(lines, "\n")
	default:
		return strings.Join(lines, "\n")
	}
}

// SaveLogsToFile exports logs to a file on disk.
func (l *Logs) SaveLogsToFile(path string, format string) string {
	content := l.ExportLogs(format)
	if content == "" {
		return "No log content to export"
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return "Failed to save logs: " + err.Error()
	}
	return "Logs saved to " + path
}
