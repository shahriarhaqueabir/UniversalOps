package sysops

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
)

// LogEntry holds a single log line.
type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Source    string `json:"source"`
	Message   string `json:"message"`
}

// SystemLogsResult holds system log output.
type SystemLogsResult struct {
	Entries []LogEntry `json:"entries"`
	Source  string     `json:"source"`
	Total   int        `json:"total"`
}

// GetSystemLogs retrieves OS system logs.
// source can be "system", "application", "security", or empty (defaults to "system").
func GetSystemLogs(n int, source string) (*SystemLogsResult, error) {
	if n <= 0 {
		n = 50
	}

	var cmd *exec.Cmd
	logSource := source

	switch runtime.GOOS {
	case "windows":
		// Map source to Windows event log name
		logName := "System"
		switch strings.ToLower(source) {
		case "application", "app":
			logName = "Application"
		case "security", "sec":
			logName = "Security"
		case "system", "":
			logName = "System"
		default:
			logName = source // Allow raw log names like "Setup", "Windows PowerShell"
		}
		logSource = logName
		cmd = common.HiddenCommand("powershell", "-Command",
			fmt.Sprintf("Get-EventLog -LogName '%s' -Newest %d | Select-Object TimeGenerated, EntryType, Source, Message | ConvertTo-Json", logName, n))
	case "linux":
		if source == "dmesg" {
			logSource = "dmesg"
			cmd = exec.Command("dmesg", "--time-format=iso", "-T")
		} else {
			logSource = "journald"
			cmd = exec.Command("journalctl", "-n", fmt.Sprintf("%d", n), "--no-pager", "-o", "short-iso")
		}
	default:
		return &SystemLogsResult{Source: "unsupported"}, nil
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	entries := parseLogOutput(string(output), logSource, runtime.GOOS)

	return &SystemLogsResult{
		Entries: entries,
		Source:  logSource,
		Total:   len(entries),
	}, nil
}

func parseLogOutput(output, source, goos string) []LogEntry {
	lines := strings.Split(output, "\n")
	var entries []LogEntry

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		entry := LogEntry{
			Source:  source,
			Message: line,
		}

		lower := strings.ToLower(line)
		switch {
		case strings.Contains(lower, "error") || strings.Contains(lower, "fatal") || strings.Contains(lower, "critical"):
			entry.Level = "error"
		case strings.Contains(lower, "warn"):
			entry.Level = "warning"
		case strings.Contains(lower, "debug") || strings.Contains(lower, "trace"):
			entry.Level = "debug"
		default:
			entry.Level = "info"
		}

		entries = append(entries, entry)
	}

	return entries
}
