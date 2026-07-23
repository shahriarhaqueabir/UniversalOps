package devops

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
)

// ServiceEntry holds service status information.
type ServiceEntry struct {
	Name        string
	DisplayName string
	Status      string
	StartType   string
}

// ListServices returns service status information for the current platform.
func ListServices(limit int) ([]ServiceEntry, error) {
	var services []ServiceEntry
	var err error

	switch {
	case common.IsWindows():
		services, err = listWindowsServices()
	case common.IsLinux():
		services, err = listLinuxServices()
	default:
		return nil, fmt.Errorf("service status query not supported on this platform")
	}
	if err != nil {
		return nil, err
	}
	if limit > 0 && len(services) > limit {
		return services[:limit], nil
	}
	return services, nil
}

// isValidServiceName checks that the service name contains only safe characters.
func isValidServiceName(name string) bool {
	if len(name) == 0 || len(name) > 256 {
		return false
	}
	for _, r := range name {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '.' || r == '_') {
			return false
		}
	}
	return true
}

// ControlService manages a service state (start, stop, restart).
func ControlService(name, action string) error {
	if !isValidServiceName(name) {
		return fmt.Errorf("invalid service name: %q", name)
	}

	if common.IsWindows() {
		var netAction string
		switch action {
		case "start":
			netAction = "start"
		case "stop":
			netAction = "stop"
		case "restart":
			stopCmd := common.SandboxedCommandWithConfig(common.SystemQuerySandbox(), "net", "stop", name)
			if err := stopCmd.Run(); err != nil {
				return fmt.Errorf("stop service %s: %w", name, err)
			}
			startCmd := common.SandboxedCommandWithConfig(common.SystemQuerySandbox(), "net", "start", name)
			return startCmd.Run()
		default:
			return fmt.Errorf("invalid service action: %s", action)
		}
		cmd := common.SandboxedCommandWithConfig(common.SystemQuerySandbox(), "net", netAction, name)
		return cmd.Run()
	}

	if common.IsLinux() {
		cmd := common.SandboxedCommandWithConfig(common.SystemQuerySandbox(), "systemctl", action, name)
		return cmd.Run()
	}

	return fmt.Errorf("service control not supported on this platform")
}

func listWindowsServices() ([]ServiceEntry, error) {
	// Try PowerShell first (best detail)
	cmd := common.SandboxedCommandWithConfig(common.SystemQuerySandbox(), "powershell", "-NoProfile", "-Command",
		"Get-Service -ErrorAction SilentlyContinue | Sort-Object Status,Name | Select-Object Name,DisplayName,Status,StartType | ConvertTo-Json -As Array -Depth 2")
	output, err := cmd.Output()
	if err == nil {
		services, parseErr := parseWindowsServicesJSON(string(output))
		if parseErr == nil && len(services) > 0 {
			return services, nil
		}
	}

	// Fallback to sc query type= service (legacy but works with lower permissions)
	cmd = common.SandboxedCommandWithConfig(common.SystemQuerySandbox(), "cmd", "/c", "sc query type= service state= all bufsize= 262144")
	output, err = cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("query Windows services (sc fallback): %w", err)
	}

	return parseSCQuery(string(output)), nil
}

// parseSCQuery parses the output of "sc query".
func parseSCQuery(output string) []ServiceEntry {
	var services []ServiceEntry
	blocks := strings.Split(output, "\n\n")

	for _, block := range blocks {
		lines := strings.Split(block, "\n")
		var s ServiceEntry
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "SERVICE_NAME:") {
				s.Name = strings.TrimSpace(strings.TrimPrefix(line, "SERVICE_NAME:"))
			} else if strings.HasPrefix(line, "DISPLAY_NAME:") {
				s.DisplayName = strings.TrimSpace(strings.TrimPrefix(line, "DISPLAY_NAME:"))
			} else if strings.HasPrefix(line, "STATE") {
				stateParts := strings.Fields(line)
				if len(stateParts) >= 4 {
					s.Status = stateParts[3] // e.g., RUNNING, STOPPED
				}
			}
		}
		if s.Name != "" {
			// Normalize status to match PowerShell
			switch s.Status {
			case "RUNNING":
				s.Status = "Running"
			case "STOPPED":
				s.Status = "Stopped"
			}
			services = append(services, s)
		}
	}
	return services
}

func listLinuxServices() ([]ServiceEntry, error) {
	cmd := common.SandboxedCommandWithConfig(common.SystemQuerySandbox(), "systemctl", "list-units", "--type=service", "--all", "--no-legend", "--no-pager")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("query Linux services: %w", err)
	}
	return parseSystemctlServices(string(output)), nil
}

func parseWindowsServicesJSON(jsonStr string) ([]ServiceEntry, error) {
	jsonStr = strings.TrimSpace(jsonStr)
	if jsonStr == "" {
		return nil, fmt.Errorf("empty service output")
	}

	// Clean malformed JSON before parsing — strip control characters
	cleaned := common.CleanJSON(jsonStr)

	// Additional cleaning for PowerShell edge cases:
	// Replace isolated "-" (dash) values used for unset/null fields
	cleaned = common.FixPowerShellDashes(cleaned)

	var raw []map[string]interface{}
	if strings.HasPrefix(cleaned, "{") {
		var single map[string]interface{}
		if err := json.Unmarshal([]byte(cleaned), &single); err != nil {
			return nil, fmt.Errorf("parse service json: %w", err)
		}
		raw = []map[string]interface{}{single}
	} else if err := json.Unmarshal([]byte(cleaned), &raw); err != nil {
		return nil, fmt.Errorf("parse service json: %w", err)
	}

	services := make([]ServiceEntry, 0, len(raw))
	for _, item := range raw {
		service := ServiceEntry{
			Name:        stringifyServiceField(item["Name"]),
			DisplayName: stringifyServiceField(item["DisplayName"]),
			Status:      stringifyServiceField(item["Status"]),
			StartType:   stringifyServiceField(item["StartType"]),
		}
		if service.Name != "" {
			services = append(services, service)
		}
	}
	return services, nil
}

func stringifyServiceField(value interface{}) string {
	switch v := value.(type) {
	case string:
		return strings.TrimSpace(v)
	case float64:
		return serviceStatusFromNumber(int(v))
	case map[string]interface{}:
		if value, ok := v["Value"]; ok {
			return stringifyServiceField(value)
		}
	}
	if value == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprintf("%v", value))
}

func serviceStatusFromNumber(status int) string {
	switch status {
	case 1:
		return "Stopped"
	case 4:
		return "Running"
	default:
		return fmt.Sprintf("Status(%d)", status)
	}
}

func parseSystemctlServices(output string) []ServiceEntry {
	var services []ServiceEntry
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}

		name := fields[0]
		status := fields[3]
		description := ""
		if len(fields) > 4 {
			description = strings.Join(fields[4:], " ")
		}
		services = append(services, ServiceEntry{
			Name:        name,
			DisplayName: description,
			Status:      status,
			StartType:   fields[2],
		})
	}
	return services
}
