package secops

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
)

// ScheduledTask represents a scheduled task.
type ScheduledTask struct {
	Name    string
	Status  string // Ready, Running, Disabled, etc.
	NextRun string
	LastRun string
	Author  string
	Trigger string
}

// GetScheduledTasks retrieves scheduled tasks.
func GetScheduledTasks() ([]ScheduledTask, error) {
	if common.IsWindows() {
		// Use HiddenCommand for Windows system tools because Get-ScheduledTask
		// requires access often restricted by sandboxing.
		cmd := common.HiddenCommand("powershell", "-Command",
			"Get-ScheduledTask | Select-Object TaskName,State,NextRunTime,LastRunTime,Author,Triggers | ConvertTo-Json -Compress -Depth 1")
		output, err := cmd.Output()
		if err == nil {
			tasks, parseErr := parseTasksJSON(string(output))
			if parseErr == nil && len(tasks) > 0 {
				return tasks, nil
			}
		}

		cmd2 := common.HiddenCommand("powershell", "-Command",
			"Get-ScheduledTask | Select-Object TaskName,State,NextRunTime,LastRunTime | ConvertTo-Json -Compress")
		output2, err2 := cmd2.Output()
		if err2 == nil {
			tasks, parseErr := parseTasksSimpleJSON(string(output2))
			if parseErr == nil && len(tasks) > 0 {
				return tasks, nil
			}
		}

		// Fallback: try schtasks (available in all Windows environments, unlike PowerShell)
		tasks, err := tasksSchTasksFallback()
		if err == nil && len(tasks) > 0 {
			return tasks, nil
		}

		return nil, fmt.Errorf("failed to query Windows scheduled tasks: all approaches failed")
	}

	if common.IsLinux() {
		return getTasksLinux()
	}

	return nil, fmt.Errorf("scheduled task query not supported on this platform")
}

// ---------------------------------------------------------------------------
// schtasks fallback (used when PowerShell is unavailable on Windows)
// ---------------------------------------------------------------------------

// tasksSchTasksFallback queries scheduled tasks via schtasks.exe as a fallback
// when PowerShell is unavailable.
func tasksSchTasksFallback() ([]ScheduledTask, error) {
	// Try verbose format first (more fields). Using HiddenCommand.
	cmd := common.HiddenCommand("cmd", "/c", "schtasks /query /v /fo csv")
	output, err := cmd.Output()
	if err == nil {
		tasks, parseErr := parseTasksSchTasksCSV(string(output))
		if parseErr == nil && len(tasks) > 0 {
			return tasks, nil
		}
	}

	// Fallback: simpler format without /v
	cmd2 := common.HiddenCommand("cmd", "/c", "schtasks /query /fo csv")
	output2, err2 := cmd2.Output()
	if err2 == nil {
		tasks, parseErr := parseTasksSchTasksSimpleCSV(string(output2))
		if parseErr == nil && len(tasks) > 0 {
			return tasks, nil
		}
	}

	return nil, fmt.Errorf("schtasks query failed")
}

// parseTasksSchTasksCSV parses the verbose CSV output from "schtasks /query /v /fo csv".
// Columns vary by locale, but common headers include:
// HostName, TaskName, Next Run Time, Status, Logon Mode, Last Run Time, Last Result, Author, ...
func parseTasksSchTasksCSV(csvStr string) ([]ScheduledTask, error) {
	reader := csv.NewReader(strings.NewReader(csvStr))
	records, err := reader.ReadAll()
	if err != nil || len(records) < 2 {
		return nil, fmt.Errorf("failed to parse schtasks CSV output")
	}

	// Build column name index (case-insensitive)
	headers := make([]string, len(records[0]))
	for i, h := range records[0] {
		headers[i] = strings.ToLower(strings.TrimSpace(h))
	}

	// Find column indices
	taskNameIdx := findCSVColumn(headers, "taskname")
	statusIdx := findCSVColumn(headers, "status")
	nextRunIdx := findCSVColumn(headers, "next runtime")
	lastRunIdx := findCSVColumn(headers, "last runtime")
	authorIdx := findCSVColumn(headers, "author")

	if taskNameIdx < 0 {
		return nil, fmt.Errorf("taskname column not found in schtasks output")
	}

	var tasks []ScheduledTask
	for i := 1; i < len(records); i++ {
		row := records[i]
		if len(row) <= taskNameIdx {
			continue
		}
		name := strings.TrimSpace(row[taskNameIdx])
		if name == "" {
			continue
		}

		task := ScheduledTask{
			Name:    name,
			Status:  "Ready",
			NextRun: "N/A",
			LastRun: "N/A",
			Author:  "N/A",
		}

		if statusIdx >= 0 && statusIdx < len(row) {
			status := strings.TrimSpace(row[statusIdx])
			if status != "" {
				task.Status = status
			}
		}
		if nextRunIdx >= 0 && nextRunIdx < len(row) {
			next := strings.TrimSpace(row[nextRunIdx])
			if next != "" && !strings.Contains(next, "N/A") {
				task.NextRun = next
			}
		}
		if lastRunIdx >= 0 && lastRunIdx < len(row) {
			last := strings.TrimSpace(row[lastRunIdx])
			if last != "" && !strings.Contains(last, "N/A") {
				task.LastRun = last
			}
		}
		if authorIdx >= 0 && authorIdx < len(row) {
			author := strings.TrimSpace(row[authorIdx])
			if author != "" {
				task.Author = author
			}
		}

		tasks = append(tasks, task)
	}

	if len(tasks) == 0 {
		return nil, fmt.Errorf("no scheduled tasks found in schtasks output")
	}

	return tasks, nil
}

// parseTasksSchTasksSimpleCSV parses the basic CSV output from "schtasks /query /fo csv".
// Format: TaskName, Next Run Time, Status
func parseTasksSchTasksSimpleCSV(csvStr string) ([]ScheduledTask, error) {
	reader := csv.NewReader(strings.NewReader(csvStr))
	records, err := reader.ReadAll()
	if err != nil || len(records) < 2 {
		return nil, fmt.Errorf("failed to parse simple schtasks CSV output")
	}

	headers := make([]string, len(records[0]))
	for i, h := range records[0] {
		headers[i] = strings.ToLower(strings.TrimSpace(h))
	}

	taskNameIdx := findCSVColumn(headers, "taskname")
	nextRunIdx := findCSVColumn(headers, "next runtime")
	statusIdx := findCSVColumn(headers, "status")

	if taskNameIdx < 0 {
		return nil, fmt.Errorf("taskname column not found in schtasks output")
	}

	var tasks []ScheduledTask
	for i := 1; i < len(records); i++ {
		row := records[i]
		if len(row) <= taskNameIdx {
			continue
		}
		name := strings.TrimSpace(row[taskNameIdx])
		if name == "" {
			continue
		}

		task := ScheduledTask{
			Name:    name,
			Author:  "N/A",
			Trigger: "N/A",
		}

		if nextRunIdx >= 0 && nextRunIdx < len(row) {
			next := strings.TrimSpace(row[nextRunIdx])
			if next != "" && !strings.Contains(next, "N/A") {
				task.NextRun = next
			} else {
				task.NextRun = "N/A"
			}
		} else {
			task.NextRun = "N/A"
		}

		if statusIdx >= 0 && statusIdx < len(row) {
			status := strings.TrimSpace(row[statusIdx])
			if status != "" {
				task.Status = status
			} else {
				task.Status = "Ready"
			}
		} else {
			task.Status = "Ready"
		}

		tasks = append(tasks, task)
	}

	if len(tasks) == 0 {
		return nil, fmt.Errorf("no scheduled tasks found in simple schtasks output")
	}

	return tasks, nil
}

// findCSVColumn finds the index of a column by name (case-insensitive, partial match).
func findCSVColumn(headers []string, name string) int {
	for i, h := range headers {
		if h == name {
			return i
		}
	}
	// Try partial match for locale variations
	for i, h := range headers {
		if strings.Contains(h, name) || strings.Contains(name, h) {
			return i
		}
	}
	return -1
}

// ---------------------------------------------------------------------------
// Linux scheduled tasks via systemd timers and cron
// ---------------------------------------------------------------------------

// getTasksLinux retrieves scheduled tasks via systemd timers and cron.
func getTasksLinux() ([]ScheduledTask, error) {
	var tasks []ScheduledTask

	// Try systemd timers first
	cmd := common.SandboxedCommandWithConfig(common.SystemQuerySandbox(), "systemctl", "list-timers", "--all")
	output, err := cmd.Output()
	if err == nil {
		tasks = parseSystemdTimers(string(output))
		if len(tasks) > 0 {
			return tasks, nil
		}
	}

	// Fallback: parse /etc/crontab and /etc/cron.d/*
	cronTasks := parseCronTab()
	if len(cronTasks) > 0 {
		tasks = append(tasks, cronTasks...)
	}

	if len(tasks) == 0 {
		return nil, fmt.Errorf("no scheduled tasks found via systemd or cron")
	}

	return tasks, nil
}

// parseSystemdTimers parses the output of "systemctl list-timers --all".
// Format:
// NEXT                        LEFT     LAST                        PASSED  UNIT                        ACTIVATES
// Mon 2024-01-15 03:00:00 UTC 8h left  Sun 2024-01-14 03:00:00 UTC 16h ago systemd-tmpfiles-clean.timer systemd-tmpfiles-clean.service
func parseSystemdTimers(output string) []ScheduledTask {
	var tasks []ScheduledTask
	scanner := bufio.NewScanner(strings.NewReader(output))
	inHeader := true

	for scanner.Scan() {
		line := scanner.Text()
		if inHeader {
			if strings.HasPrefix(line, "NEXT") || strings.HasPrefix(line, "UNIT") {
				inHeader = false
			}
			continue
		}
		if strings.TrimSpace(line) == "" {
			continue
		}
		if strings.HasPrefix(line, "(") || strings.Contains(line, "timers listed") {
			continue
		}

		// Parse line: columns are whitespace-separated
		fields := strings.Fields(line)
		if len(fields) < 7 {
			continue
		}

		// NEXT is fields[0..2], LEFT is fields[3..4], LAST is fields[5..7], PASED is fields[8..9], UNIT is fields[10], ACTIVATES is fields[11]
		nextRun := ""
		lastRun := ""
		unitName := ""
		activates := ""

		// The format varies; try common patterns
		if strings.Contains(line, "left") && strings.Contains(line, "ago") {
			// Full format with all columns
			parts := splitSystemdTimerLine(fields)
			nextRun = parts.next
			lastRun = parts.last
			unitName = parts.unit
			activates = parts.activates
		} else {
			// Sparse format - fewer columns
			if len(fields) >= 7 {
				unitName = fields[len(fields)-2]
				activates = fields[len(fields)-1]
			}
		}

		name := unitName
		if activates != "" {
			name = activates
		}
		if name == "" {
			name = unitName
		}

		tasks = append(tasks, ScheduledTask{
			Name:    name,
			Status:  "Ready",
			NextRun: nextRun,
			LastRun: lastRun,
			Author:  "systemd",
			Trigger: "timer",
		})
	}

	return tasks
}

// systemdTimerParts holds parsed fields from a systemctl list-timers line.
type systemdTimerParts struct {
	next      string
	last      string
	unit      string
	activates string
}

// splitSystemdTimerLine attempts to extract fields from a systemd timer line.
func splitSystemdTimerLine(fields []string) systemdTimerParts {
	var result systemdTimerParts
	if len(fields) < 7 {
		return result
	}

	// The UNIT and ACTIVATES are the last two fields
	result.unit = fields[len(fields)-2]
	result.activates = fields[len(fields)-1]

	// NEXT is the first field (e.g., "Mon" or "2024-")
	if len(fields) > 11 {
		// Full format
		result.next = fmt.Sprintf("%s %s %s", fields[0], fields[1], fields[2])
		result.last = fmt.Sprintf("%s %s %s", fields[5], fields[6], fields[7])
	}

	return result
}

// parseCronTab parses /etc/crontab and /etc/cron.d/* for scheduled tasks.
func parseCronTab() []ScheduledTask {
	var tasks []ScheduledTask

	// Parse main crontab
	tasks = append(tasks, parseCronFile("/etc/crontab")...)

	// Parse cron.d directory
	entries, err := os.ReadDir("/etc/cron.d")
	if err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				path := "/etc/cron.d/" + entry.Name()
				tasks = append(tasks, parseCronFile(path)...)
			}
		}
	}

	// Parse cron.hourly, daily, weekly, monthly
	for _, dir := range []string{"/etc/cron.hourly", "/etc/cron.daily", "/etc/cron.weekly", "/etc/cron.monthly"} {
		entries, err := os.ReadDir(dir)
		if err == nil {
			for _, entry := range entries {
				if !entry.IsDir() {
					tasks = append(tasks, ScheduledTask{
						Name:    entry.Name(),
						Status:  "Ready",
						NextRun: "N/A",
						LastRun: "N/A",
						Author:  "root",
						Trigger: dir,
					})
				}
			}
		}
	}

	return tasks
}

// parseCronFile parses a single crontab file.
func parseCronFile(path string) []ScheduledTask {
	var tasks []ScheduledTask
	file, err := os.Open(path)
	if err != nil {
		return tasks
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Skip environment variable lines
		if strings.Contains(line, "=") && !strings.Contains(line, " ") {
			continue
		}

		// Extract command as task name
		fields := strings.Fields(line)
		if len(fields) >= 6 {
			// Format: minute hour day month weekday command
			trigger := fmt.Sprintf("%s %s %s %s %s", fields[0], fields[1], fields[2], fields[3], fields[4])
			command := strings.Join(fields[5:], " ")

			name := command
			if len(name) > 40 {
				name = name[:37] + "..."
			}

			tasks = append(tasks, ScheduledTask{
				Name:    name,
				Status:  "Ready",
				NextRun: "N/A",
				LastRun: "N/A",
				Author:  "cron",
				Trigger: trigger,
			})
		}
	}

	return tasks
}

// ---------------------------------------------------------------------------
// PowerShell JSON parsing
// ---------------------------------------------------------------------------

// parseTasksJSON parses the JSON array from Get-ScheduledTask.
func parseTasksJSON(jsonStr string) ([]ScheduledTask, error) {
	var raw []map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &raw); err != nil {
		return nil, fmt.Errorf("failed to parse tasks JSON: %w", err)
	}

	var tasks []ScheduledTask
	for _, item := range raw {
		task := ScheduledTask{}

		if name, ok := getJSONString(item, "TaskName"); ok {
			task.Name = name
		}
		if task.Name == "" {
			continue
		}

		if state, ok := item["State"]; ok {
			switch s := state.(type) {
			case string:
				task.Status = s
			case map[string]interface{}:
				if v, ok := s["Value"]; ok {
					task.Status = fmt.Sprintf("%v", v)
				}
			case float64:
				task.Status = taskStateToString(int(s))
			}
		}
		if task.Status == "" {
			task.Status = "Ready"
		}

		if next, ok := getJSONString(item, "NextRunTime"); ok {
			task.NextRun = formatTaskTime(next)
		}
		if task.NextRun == "" || task.NextRun == "12/31/1600 00:00:00" || task.NextRun == "0001-01-01T00:00:00" {
			task.NextRun = "N/A"
		}

		if last, ok := getJSONString(item, "LastRunTime"); ok {
			task.LastRun = formatTaskTime(last)
		}
		if task.LastRun == "" || task.LastRun == "12/31/1600 00:00:00" || task.LastRun == "0001-01-01T00:00:00" {
			task.LastRun = "N/A"
		}

		if author, ok := getJSONString(item, "Author"); ok {
			task.Author = author
		}

		if triggers, ok := item["Triggers"]; ok {
			if tArr, ok := triggers.([]interface{}); ok && len(tArr) > 0 {
				if tMap, ok := tArr[0].(map[string]interface{}); ok {
					if enabled, ok := tMap["Enabled"]; ok {
						if enabledBool, ok := enabled.(bool); ok && enabledBool {
							if typeStr, ok := getJSONString(tMap, "Type"); ok {
								task.Trigger = typeStr
							}
						}
					}
				}
			}
		}
		if task.Trigger == "" {
			task.Trigger = "N/A"
		}

		tasks = append(tasks, task)
	}

	return tasks, nil
}

// parseTasksSimpleJSON parses a simpler JSON output from Get-ScheduledTask.
func parseTasksSimpleJSON(jsonStr string) ([]ScheduledTask, error) {
	var raw []map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &raw); err != nil {
		return nil, fmt.Errorf("failed to parse simple tasks JSON: %w", err)
	}

	var tasks []ScheduledTask
	for _, item := range raw {
		task := ScheduledTask{}

		if name, ok := getJSONString(item, "TaskName"); ok {
			task.Name = name
		}
		if task.Name == "" {
			continue
		}

		if state, ok := item["State"]; ok {
			switch s := state.(type) {
			case string:
				task.Status = s
			case map[string]interface{}:
				if v, ok := s["Value"]; ok {
					task.Status = fmt.Sprintf("%v", v)
				}
			case float64:
				task.Status = taskStateToString(int(s))
			default:
				task.Status = "Ready"
			}
		} else {
			task.Status = "Ready"
		}

		if next, ok := getJSONString(item, "NextRunTime"); ok {
			task.NextRun = formatTaskTime(next)
		}
		if task.NextRun == "" || task.NextRun == "12/31/1600 00:00:00" || task.NextRun == "0001-01-01T00:00:00" {
			task.NextRun = "N/A"
		}

		if last, ok := getJSONString(item, "LastRunTime"); ok {
			task.LastRun = formatTaskTime(last)
		}
		if task.LastRun == "" || task.LastRun == "12/31/1600 00:00:00" || task.LastRun == "0001-01-01T00:00:00" {
			task.LastRun = "N/A"
		}

		task.Author = "N/A"
		task.Trigger = "N/A"

		tasks = append(tasks, task)
	}

	return tasks, nil
}

// formatTaskTime cleans up a DateTime string for display.
func formatTaskTime(t string) string {
	t = trimDateTime(t)
	if len(t) >= 16 {
		return t[:16]
	}
	return t
}

// trimDateTime removes trailing fractional seconds and timezone from ISO dates.
func trimDateTime(t string) string {
	if len(t) > 19 && t[10] == 'T' {
		return t[:19]
	}
	return t
}

// taskStateToString converts the numeric task state to a display string.
func taskStateToString(state int) string {
	switch state {
	case 0:
		return "Unknown"
	case 1:
		return "Disabled"
	case 2:
		return "Queued"
	case 3:
		return "Ready"
	case 4:
		return "Running"
	default:
		return fmt.Sprintf("State(%d)", state)
	}
}
