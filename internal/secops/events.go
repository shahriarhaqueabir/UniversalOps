package secops

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
)

// SecurityEvent represents a security event log entry.
type SecurityEvent struct {
	ID        int
	Level     string
	Provider  string
	Time      string
	Message   string
	Important bool
}

// GetSecurityEvents retrieves recent security events from the current platform.
func GetSecurityEvents() ([]SecurityEvent, error) {
	if common.IsWindows() {
		// Try Security log first (requires Admin)
		cmd := common.SandboxedCommandWithConfig(common.SystemQuerySandbox(), "powershell", "-Command",
			"Get-WinEvent -LogName Security -MaxEvents 25 -ErrorAction Stop | Select-Object Id,LevelDisplayName,ProviderName,TimeCreated,Message | ConvertTo-Json -As Array -Depth 2")
		output, err := cmd.Output()
		if err != nil {
			// Fallback: Try System log (usually accessible by non-admins)
			common.LogInfo("Security log inaccessible, falling back to System log: %v", err)
			cmd = common.SandboxedCommandWithConfig(common.SystemQuerySandbox(), "powershell", "-Command",
				"Get-WinEvent -LogName System -MaxEvents 25 -ErrorAction SilentlyContinue | Select-Object Id,LevelDisplayName,ProviderName,TimeCreated,Message | ConvertTo-Json -As Array -Depth 2")
			output, err = cmd.Output()
			if err != nil {
				return nil, fmt.Errorf("failed to query Windows security/system event logs: %w", err)
			}
		}
		return parseSecurityEventsJSON(string(output))
	}

	return nil, fmt.Errorf("security event log query not supported on this platform")
}

func parseSecurityEventsJSON(jsonStr string) ([]SecurityEvent, error) {
	jsonStr = strings.TrimSpace(jsonStr)
	if jsonStr == "" {
		return nil, fmt.Errorf("empty security event log output")
	}

	// Clean malformed JSON before parsing
	cleaned := common.CleanJSON(jsonStr)
	// Fix PowerShell dash values for numeric fields
	cleaned = common.FixPowerShellDashes(cleaned)

	var raw []map[string]interface{}
	if strings.HasPrefix(cleaned, "{") {
		var single map[string]interface{}
		if err := json.Unmarshal([]byte(cleaned), &single); err != nil {
			return nil, fmt.Errorf("failed to parse security events JSON: %w", err)
		}
		raw = []map[string]interface{}{single}
	} else if err := json.Unmarshal([]byte(cleaned), &raw); err != nil {
		return nil, fmt.Errorf("failed to parse security events JSON: %w", err)
	}

	events := make([]SecurityEvent, 0, len(raw))
	for _, item := range raw {
		event := SecurityEvent{}

		if id, ok := getJSONInt(item, "Id"); ok {
			event.ID = id
		}
		if level, ok := getJSONString(item, "LevelDisplayName"); ok {
			event.Level = level
		}
		if provider, ok := getJSONString(item, "ProviderName"); ok {
			event.Provider = provider
		}
		if created, ok := getJSONString(item, "TimeCreated"); ok {
			event.Time = formatEventTime(created)
		}
		if message, ok := getJSONString(item, "Message"); ok {
			event.Message = strings.TrimSpace(message)
		}
		event.Important = isImportantSecurityEvent(event)

		if event.ID != 0 || event.Message != "" {
			events = append(events, event)
		}
	}

	return events, nil
}

func formatEventTime(value string) string {
	value = trimDateTime(value)
	if len(value) >= 19 {
		return strings.ReplaceAll(value[:19], "T", " ")
	}
	return value
}

func isImportantSecurityEvent(event SecurityEvent) bool {
	switch event.ID {
	case 4625, 4720, 4726, 4732, 4733, 4740, 1102:
		return true
	}
	return strings.Contains(strings.ToLower(event.Level), "error") ||
		strings.Contains(strings.ToLower(event.Level), "critical")
}
