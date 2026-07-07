package app

import (
	"os"
	"strings"
	"time"

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

// GetLogs returns filtered log entries from the hawkward log file.
// level filters by log level ("info", "warn", "error"). Empty string returns all.
// since filters by ISO 8601 timestamp (optional).
// n limits the number of entries returned.
func (l *Logs) GetLogs(level string, since string, n int) []LogEntry {
	// Read the last 500 lines from the main log file
	lines, err := devops.TailLog("hawkward-gui.log", 500)
	if err != nil {
		// Try the default hawkward.log as fallback
		lines, err = devops.TailLog("hawkward.log", 500)
		if err != nil {
			common.LogWarn("GetLogs failed: %v", err)
			return nil
		}
	}

	var entries []LogEntry
	levelUpper := strings.ToUpper(level)

	for _, line := range lines {
		// Parse the log line format: [HAWKWARD] YYYY/MM/DD HH:MM:SS [LEVEL] message
		entry := parseLogLine(line)

		// Apply level filter
		if levelUpper != "" && entry.Level != levelUpper {
			continue
		}

		// Apply time filter
		if since != "" {
			sinceTime, err := time.Parse(time.RFC3339, since)
			if err == nil {
				entryTime, err := time.Parse("2006/01/02 15:04:05", entry.Timestamp)
				if err == nil && entryTime.Before(sinceTime) {
					continue
				}
			}
		}

		entries = append(entries, entry)

		// Apply count limit
		if n > 0 && len(entries) >= n {
			break
		}
	}

	return entries
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

// ── Log parsing ──────────────────────────────────────────────────────────────

// parseLogLine parses a hawkward log line into a LogEntry.
// Expected format: [HAWKWARD] YYYY/MM/DD HH:MM:SS [LEVEL] message
func parseLogLine(line string) LogEntry {
	entry := LogEntry{Line: line}

	// Try to extract timestamp and level
	parts := strings.SplitN(line, " ", 5)
	if len(parts) >= 4 {
		// parts[0] = [HAWKWARD]
		// parts[1] = date
		// parts[2] = time
		// parts[3] = [LEVEL]
		entry.Timestamp = strings.TrimPrefix(parts[1], "[") + " " + strings.TrimSuffix(parts[2], "]")
		level := strings.Trim(parts[3], "[]")
		entry.Level = level

		if len(parts) >= 5 {
			entry.Source = level
		}
	}

	return entry
}
