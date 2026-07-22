package secops

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
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
	out, err := exec.Command("powershell", "-Command",
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
	out, err := exec.Command("lsblk", "--discard", "-o", "NAME,TYPE,FSTYPE,MOUNTPOINT").Output()
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
	out, err := exec.Command("powershell", "-Command", "Confirm-SecureBootUEFI").Output()
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
	out, err := exec.Command("mokutil", "--sb-state").Output()
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
	out, err := exec.Command("powershell", "-Command",
		"Get-Service | Select-Object Name,DisplayName,Status,StartType | ConvertTo-Json -As Array -Depth 2").Output()
	if err != nil {
		return nil, fmt.Errorf("Get-Service failed: %w", err)
	}
	return parseServicesJSON(string(out)), nil
}

func parseServicesJSON(jsonStr string) []SystemService {
	// Simplified: return empty
	return []SystemService{}
}

func getRunningServicesLinux() ([]SystemService, error) {
	out, err := exec.Command("systemctl", "list-units", "--type=service", "--all", "--no-pager", "--plain", "--no-legend").Output()
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
