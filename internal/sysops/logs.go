package sysops

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/shahriarhaqueabir/UniversalOps/internal/common"
)

// dotNetDateRe matches .NET JSON date format: /Date(1704067200000)/
var dotNetDateRe = regexp.MustCompile(`/Date\((\d+)\)/`)

// PowerShellEventLogEntry maps the fields from Get-EventLog / Get-WinEvent output.
// TimeGenerated uses json.RawMessage to handle DateTime serialisation from both
// classic Get-EventLog (string) and Get-WinEvent (object or null).
type PowerShellEventLogEntry struct {
	TimeGenerated json.RawMessage `json:"TimeGenerated"`
	EntryType     string          `json:"EntryType"`
	Source        string          `json:"Source"`
	Message       string          `json:"Message"`
}

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

		// Try Get-EventLog first (classic API, works on most systems)
		output, cmdErr := common.HiddenCommand("powershell", "-Command",
			fmt.Sprintf("Get-EventLog -LogName '%s' -Newest %d | Select-Object TimeGenerated, EntryType, Source, Message | ConvertTo-Json -Depth 5", logName, n)).CombinedOutput()
		if cmdErr == nil && len(strings.TrimSpace(string(output))) > 0 && !strings.HasPrefix(strings.TrimSpace(string(output)), "[]") {
			entries := parseLogOutput(string(output), logSource, "windows")
			return &SystemLogsResult{Entries: entries, Source: logSource, Total: len(entries)}, nil
		}

		// Fallback: Get-WinEvent (modern cmdlet, works on PowerShell 3+/Windows 8+).
		// Calculated properties map Get-WinEvent fields to Get-EventLog-compatible JSON
		// names so the same parseLogOutput code handles both.
		cmd = common.HiddenCommand("powershell", "-Command",
			fmt.Sprintf("Get-WinEvent -LogName '%s' -MaxEvents %d -ErrorAction SilentlyContinue | Select-Object @{N='TimeGenerated';E={$_.TimeCreated}}, @{N='EntryType';E={$_.LevelDisplayName}}, @{N='Source';E={$_.ProviderName}}, Message | ConvertTo-Json -Depth 5", logName, n))
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
	// Non-Windows: fall back to line-by-line parsing (dmesg / journalctl)
	if goos != "windows" {
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

	// Windows: parse JSON array from Get-EventLog | ConvertTo-Json
	output = strings.TrimSpace(output)
	if output == "" || output == "[]" || output == "null" {
		return nil
	}

	// Strip UTF-16LE null bytes (Windows PowerShell defaults to UTF-16LE
	// when stdout is redirected). Valid UTF-8/ASCII JSON has no null bytes,
	// so this safely normalises the encoding without adding dependencies.
	output = strings.ReplaceAll(output, "\x00", "")

	// Handle both JSON array and single-object cases
	var rawEntries []json.RawMessage
	if strings.HasPrefix(output, "[") {
		if err := json.Unmarshal([]byte(output), &rawEntries); err != nil {
			return nil
		}
	} else {
		rawEntries = []json.RawMessage{json.RawMessage(output)}
	}

	var entries []LogEntry
	for _, raw := range rawEntries {
		var pe PowerShellEventLogEntry
		if err := json.Unmarshal(raw, &pe); err != nil {
			continue
		}
		entry := LogEntry{
			Timestamp: parseDotNetDate(pe.TimeGenerated),
			Source:    pe.Source,
			Message:   pe.Message,
		}
		switch strings.ToLower(pe.EntryType) {
		case "error", "critical":
			entry.Level = "error"
		case "warning":
			entry.Level = "warning"
		case "verbose", "debug":
			entry.Level = "debug"
		case "information", "successaudit", "failureaudit":
			entry.Level = "info"
		default:
			entry.Level = "info"
		}
		entries = append(entries, entry)
	}
	return entries
}

// parseDotNetDate converts a .NET JSON date (/Date(millis)/) or an ISO‑8601
// string from a json.RawMessage to ISO 8601. Returns "" on empty / null input.
func parseDotNetDate(raw json.RawMessage) string {
	if len(raw) == 0 || string(raw) == "null" {
		return ""
	}
	// Unquote the JSON string value.
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return ""
	}
	if s == "" {
		return ""
	}
	// Try .NET /Date(millis)/ format first
	matches := dotNetDateRe.FindStringSubmatch(s)
	if len(matches) >= 2 {
		millis, err := strconv.ParseInt(matches[1], 10, 64)
		if err == nil {
			return time.UnixMilli(millis).UTC().Format(time.RFC3339)
		}
	}
	// ISO‑8601 or whatever the string contains — return as-is
	return s
}
