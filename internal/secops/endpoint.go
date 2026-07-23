package secops

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
)

// DiskEncryption holds disk encryption status.
type DiskEncryption struct {
	Volume    string `json:"volume"`
	Encrypted bool   `json:"encrypted"`
	Method    string `json:"method"`
	Status    string `json:"status"`
}

// SecureBoot holds secure boot status.
type SecureBoot struct {
	Enabled bool   `json:"enabled"`
	State   string `json:"state"`
}

// SystemService holds a system service entry.
type SystemService struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Status      string `json:"status"`
	StartupType string `json:"startup_type"`
}

// GetDiskEncryptionStatus retrieves disk encryption status.
func GetDiskEncryptionStatus() ([]DiskEncryption, error) {
	if common.IsWindows() {
		return getDiskEncryptionWindows()
	}
	return getDiskEncryptionLinux()
}

func getDiskEncryptionWindows() ([]DiskEncryption, error) {
	out, err := common.HiddenCommand("powershell", "-NoProfile", "-Command",
		"Get-BitLockerVolume | Select-Object MountPoint,ProtectionStatus,EncryptionMethod | ConvertTo-Json -As Array -Depth 2").Output()
	if err != nil {
		return []DiskEncryption{{Volume: "C:", Encrypted: false, Method: "Unknown", Status: "Unavailable"}}, nil
	}
	return parseBitLockerJSON(string(out)), nil
}

type bitLockerVolume struct {
	MountPoint          string `json:"MountPoint"`
	ProtectionStatus    *int   `json:"ProtectionStatus"`
	EncryptionMethod    string `json:"EncryptionMethod"`
	ConversionStatus    *int   `json:"ConversionStatus"`
	PercentageEncrypted *int   `json:"PercentageEncrypted"`
}

// bitLockerProtection maps BitLocker ProtectionStatus values to boolean.
func bitLockerProtection(status *int) bool {
	return status != nil && *status == 1
}

// bitLockerMethod normalises the encryption method string.
func bitLockerMethod(method string) string {
	if method == "" || method == "None" {
		return "None"
	}
	return method
}

func parseBitLockerJSON(jsonStr string) []DiskEncryption {
	// PowerShell may return a single object or an array.
	// Try array first, then single object.
	var volumes []bitLockerVolume
	if err := json.Unmarshal([]byte(jsonStr), &volumes); err != nil {
		var single bitLockerVolume
		if err2 := json.Unmarshal([]byte(jsonStr), &single); err2 != nil {
			return []DiskEncryption{{Volume: "C:", Encrypted: false, Method: "Unknown", Status: "ParseError"}}
		}
		volumes = []bitLockerVolume{single}
	}

	var disks []DiskEncryption
	for _, v := range volumes {
		if v.MountPoint == "" {
			continue
		}
		status := "Active"
		if v.ProtectionStatus == nil || *v.ProtectionStatus == 0 {
			status = "Disabled"
		}
		disks = append(disks, DiskEncryption{
			Volume:    v.MountPoint,
			Encrypted: bitLockerProtection(v.ProtectionStatus),
			Method:    bitLockerMethod(v.EncryptionMethod),
			Status:    status,
		})
	}
	if len(disks) == 0 {
		disks = append(disks, DiskEncryption{Volume: "C:", Encrypted: false, Method: "Unknown", Status: "No BitLocker volumes found"})
	}
	return disks
}

func getDiskEncryptionLinux() ([]DiskEncryption, error) {
	out, err := common.HiddenCommand("lsblk", "--discard", "-o", "NAME,TYPE,FSTYPE,MOUNTPOINT").Output()
	if err != nil {
		return nil, fmt.Errorf("lsblk failed: %w", err)
	}
	var disks []DiskEncryption
	for _, line := range strings.Split(string(out), "\n") {
		fields := strings.Fields(strings.TrimSpace(line))
		if len(fields) >= 3 && fields[1] == "part" && fields[2] == "crypto" {
			disks = append(disks, DiskEncryption{
				Volume:    fields[0],
				Encrypted: true,
				Method:    "LUKS",
				Status:    "Active",
			})
		}
	}
	if len(disks) == 0 {
		disks = append(disks, DiskEncryption{Volume: "N/A", Encrypted: false, Method: "None", Status: "No encryption detected"})
	}
	return disks, nil
}

// GetSecureBootStatus retrieves secure boot status.
func GetSecureBootStatus() (*SecureBoot, error) {
	if common.IsWindows() {
		return getSecureBootWindows()
	}
	return getSecureBootLinux()
}

func getSecureBootWindows() (*SecureBoot, error) {
	out, err := common.HiddenCommand("powershell", "-Command", "Confirm-SecureBootUEFI").Output()
	if err != nil {
		return &SecureBoot{Enabled: false, State: "Unable to determine"}, nil
	}
	output := strings.TrimSpace(string(out))
	return &SecureBoot{
		Enabled: strings.EqualFold(output, "True"),
		State:   output,
	}, nil
}

func getSecureBootLinux() (*SecureBoot, error) {
	_, err := os.Stat("/sys/firmware/efi")
	if err != nil {
		return &SecureBoot{Enabled: false, State: "Legacy BIOS"}, nil
	}
	out, err := common.HiddenCommand("mokutil", "--sb-state").Output()
	if err != nil {
		return &SecureBoot{Enabled: false, State: "UEFI (status unknown)"}, nil
	}
	output := strings.TrimSpace(string(out))
	return &SecureBoot{
		Enabled: strings.Contains(strings.ToLower(output), "enabled"),
		State:   output,
	}, nil
}

// GetRunningServices retrieves running system services.
func GetRunningServices() ([]SystemService, error) {
	if common.IsWindows() {
		return getRunningServicesWindows()
	}
	return getRunningServicesLinux()
}

func getRunningServicesWindows() ([]SystemService, error) {
	// Try PowerShell (omit -As Array for PS 5.1 compatibility)
	out, err := common.HiddenCommand("powershell", "-Command",
		"Get-Service | Select-Object Name,DisplayName,Status,StartType | ConvertTo-Json -Depth 2").Output()
	if err == nil {
		services := parseServicesJSON(string(out))
		if len(services) > 0 {
			return services, nil
		}
	}

	// Fallback to sc query (works on all Windows versions)
	cmd := common.HiddenCommand("cmd", "/c", "sc query type= service state= all bufsize= 262144")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("query Windows services: %w", err)
	}
	return parseServicesSCQuery(string(output)), nil
}

// parseServicesJSON parses Get-Service JSON output (handles both PS5.1 numeric
// values and PS7+ string values, and both single-object and array formats).
func parseServicesJSON(jsonStr string) []SystemService {
	jsonStr = strings.TrimSpace(jsonStr)
	if jsonStr == "" {
		return nil
	}

	// Handle both array and single-object JSON
	var raw []map[string]interface{}
	if strings.HasPrefix(jsonStr, "{") {
		var single map[string]interface{}
		if err := json.Unmarshal([]byte(jsonStr), &single); err != nil {
			return nil
		}
		raw = []map[string]interface{}{single}
	} else if err := json.Unmarshal([]byte(jsonStr), &raw); err != nil {
		return nil
	}

	services := make([]SystemService, 0, len(raw))
	for _, item := range raw {
		s := SystemService{
			Name:        stringOrEmpty(item["Name"]),
			DisplayName: stringOrEmpty(item["DisplayName"]),
			Status:      serviceStatusString(item["Status"]),
			StartupType: serviceStartTypeString(item["StartType"]),
		}
		if s.Name != "" {
			services = append(services, s)
		}
	}
	return services
}

// stringOrEmpty extracts a string from a JSON value, returning "" on nil.
func stringOrEmpty(v interface{}) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return strings.TrimSpace(val)
	case float64:
		return fmt.Sprintf("%.0f", val)
	}
	return strings.TrimSpace(fmt.Sprintf("%v", v))
}

// serviceStatusString converts numeric or string service status to display form.
// PowerShell 5.1 returns int (1=Stopped, 4=Running).
func serviceStatusString(v interface{}) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case float64:
		switch int(val) {
		case 1:
			return "Stopped"
		case 4:
			return "Running"
		default:
			return fmt.Sprintf("Unknown(%d)", int(val))
		}
	}
	return fmt.Sprintf("%v", v)
}

// serviceStartTypeString converts numeric or string start type to display form.
// PowerShell 5.1 returns int (2=Auto, 3=Manual, 4=Disabled).
func serviceStartTypeString(v interface{}) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case float64:
		switch int(val) {
		case 1:
			return "Other"
		case 2:
			return "Automatic"
		case 3:
			return "Manual"
		case 4:
			return "Disabled"
		default:
			return fmt.Sprintf("Unknown(%d)", int(val))
		}
	}
	return fmt.Sprintf("%v", v)
}

// parseServicesSCQuery parses the output of "sc query type= service state= all".
// sc query does not include START_TYPE, so that field will be empty.
func parseServicesSCQuery(output string) []SystemService {
	var services []SystemService
	blocks := strings.Split(output, "\n\n")
	for _, block := range blocks {
		lines := strings.Split(block, "\n")
		var s SystemService
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "SERVICE_NAME:") {
				s.Name = strings.TrimSpace(strings.TrimPrefix(line, "SERVICE_NAME:"))
			} else if strings.HasPrefix(line, "DISPLAY_NAME:") {
				s.DisplayName = strings.TrimSpace(strings.TrimPrefix(line, "DISPLAY_NAME:"))
			} else if strings.HasPrefix(line, "STATE") {
				stateParts := strings.Fields(line)
				if len(stateParts) >= 4 {
					s.Status = stateParts[3]
				}
			}
		}
		if s.Name != "" {
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

func getRunningServicesLinux() ([]SystemService, error) {
	out, err := common.HiddenCommand("systemctl", "list-units", "--type=service", "--all", "--no-pager", "--plain", "--no-legend").Output()
	if err != nil {
		return nil, fmt.Errorf("systemctl failed: %w", err)
	}
	var services []SystemService
	for _, line := range strings.Split(string(out), "\n") {
		fields := strings.Fields(strings.TrimSpace(line))
		if len(fields) >= 4 {
			services = append(services, SystemService{
				Name:        strings.TrimSuffix(fields[0], ".service"),
				DisplayName: strings.Join(fields[3:], " "),
				Status:      fields[2],
				StartupType: fields[1],
			})
		}
	}
	return services, nil
}
