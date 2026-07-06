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

func listWindowsServices() ([]ServiceEntry, error) {
	cmd := common.SandboxedCommandWithConfig(common.SystemQuerySandbox(), "powershell", "-Command",
		"Get-Service | Sort-Object Status,Name | Select-Object Name,DisplayName,Status,StartType | ConvertTo-Json -Compress")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("query Windows services: %w", err)
	}
	return parseWindowsServicesJSON(string(output))
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

	var raw []map[string]interface{}
	if strings.HasPrefix(jsonStr, "{") {
		var single map[string]interface{}
		if err := json.Unmarshal([]byte(jsonStr), &single); err != nil {
			return nil, fmt.Errorf("parse service json: %w", err)
		}
		raw = []map[string]interface{}{single}
	} else if err := json.Unmarshal([]byte(jsonStr), &raw); err != nil {
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
