//go:build windows

package sysops

import (
	"strings"
	"time"

	"github.com/shahriarhaqueabir/UniversalOps/internal/common"
	"github.com/yusufpapurcu/wmi"
)

// Win32_ComputerSystem holds minimal WMI data for user detection.
type Win32_ComputerSystem struct {
	Name     string
	UserName string // Format: "DOMAIN\username"
}

// win32BatteryWMI is the raw WMI struct for Win32_Battery.
// Uses signed int32 fields to avoid the reflect.Value.Uint panic.
type win32BatteryWMI struct {
	EstimatedChargeRemaining int32
	BatteryStatus            int32
	DesignCapacity           int32
	FullChargeCapacity       int32
	CycleCount               int32
	Chemistry                int32
}

// GetBaseboardInfo queries WMI for motherboard details.
func GetBaseboardInfo() *BaseboardInfo {
	type Win32_BaseBoard struct {
		Manufacturer string
		Product      string
		Version      string
		SerialNumber string
	}
	var dst []Win32_BaseBoard
	if err := common.WMIQueryWithTimeout("SELECT Manufacturer, Product, Version, SerialNumber FROM Win32_BaseBoard", &dst, 2*time.Second); err != nil || len(dst) == 0 {
		return nil
	}

	type Win32_SystemEnclosure struct {
		ChassisTypes []int32
	}
	var enc []Win32_SystemEnclosure
	chassis := "Unknown"
	if err := common.WMIQueryWithTimeout("SELECT ChassisTypes FROM Win32_SystemEnclosure", &enc, 2*time.Second); err == nil && len(enc) > 0 {
		if len(enc[0].ChassisTypes) > 0 {
			chassis = mapChassisType(uint16(enc[0].ChassisTypes[0]))
		}
	}

	return &BaseboardInfo{
		Manufacturer: strings.TrimSpace(dst[0].Manufacturer),
		Product:      strings.TrimSpace(dst[0].Product),
		Version:      strings.TrimSpace(dst[0].Version),
		SerialNumber: strings.TrimSpace(dst[0].SerialNumber),
		ChassisType:  chassis,
	}
}

// GetPhysicalDisks queries WMI for physical drive health and SMART status.
func GetPhysicalDisks() ([]PhysicalDisk, error) {
	type Win32_DiskDrive struct {
		DeviceID     string
		Model        string
		MediaType    string
		Size         int64
		Status       string
		SerialNumber string
	}
	var dst []Win32_DiskDrive
	q := "SELECT DeviceID, Model, MediaType, Size, Status, SerialNumber FROM Win32_DiskDrive"
	if err := common.WMIQueryWithTimeout(q, &dst, 3*time.Second); err != nil {
		return nil, err
	}

	// Secondary query for Predictive Failure (SMART)
	type MSStorageDriver_FailurePredictStatus struct {
		InstanceName   string
		PredictFailure bool
	}
	var smart []MSStorageDriver_FailurePredictStatus
	_ = common.WMIQueryNamespaceWithTimeout("SELECT InstanceName, PredictFailure FROM MSStorageDriver_FailurePredictStatus", &smart, "root\\wmi", 2*time.Second)

	res := make([]PhysicalDisk, len(dst))
	for i, d := range dst {
		res[i] = PhysicalDisk{
			DeviceID:     d.DeviceID,
			Model:        d.Model,
			MediaType:    d.MediaType,
			Size:         uint64(d.Size),
			Status:       d.Status,
			SerialNumber: strings.TrimSpace(d.SerialNumber),
		}
		// Match SMART status if available
		for _, s := range smart {
			if strings.Contains(strings.ToLower(s.InstanceName), strings.ToLower(d.DeviceID)) {
				res[i].PredictFail = s.PredictFailure
			}
		}
	}
	return res, nil
}

// GetDetailedBatteryHealth returns wear levels and cycle counts from WMI.
func GetDetailedBatteryHealth() (*BatteryHealth, error) {
	var dst []win32BatteryWMI
	if err := common.WMIQueryWithTimeout("SELECT EstimatedChargeRemaining, BatteryStatus, DesignCapacity, FullChargeCapacity, CycleCount, Chemistry FROM Win32_Battery", &dst, 2*time.Second); err != nil || len(dst) == 0 {
		return nil, nil
	}

	b := dst[0]
	health := &BatteryHealth{
		Percent:    int(b.EstimatedChargeRemaining),
		Charging:   b.BatteryStatus == 2 || b.BatteryStatus == 6 || b.BatteryStatus == 7,
		DesignCap:  uint32(b.DesignCapacity),
		FullCap:    uint32(b.FullChargeCapacity),
		CycleCount: uint32(b.CycleCount),
	}

	if b.DesignCapacity > 0 {
		health.WearLevel = 100.0 - (float64(b.FullChargeCapacity) / float64(b.DesignCapacity) * 100.0)
	}

	return health, nil
}

// getLoggedInUsersPlatform returns logged-in users on Windows via `query user` with WMI fallback.
func getLoggedInUsersPlatform() ([]LoggedInUser, error) {
	return getLoggedInUsersWindows()
}

// getLoggedInUsersWindows uses the Windows `query user` command to list logged-in users.
// Falls back to WMI if the command is not available (e.g. Windows Home edition).
func getLoggedInUsersWindows() ([]LoggedInUser, error) {
	cmd := common.SandboxedCommandWithConfig(common.SystemQuerySandbox(), "query", "user")
	output, err := cmd.Output()
	if err != nil {
		// `query user` not available (Windows Home, sandboxed, etc.) — use WMI instead
		return getLoggedInUsersWMI()
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) < 2 {
		return []LoggedInUser{}, nil
	}

	var result []LoggedInUser
	// Skip header line (line 0); parse data lines
	for _, line := range lines[1:] {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// query user output is fixed-width: USERNAME  SESSIONNAME  ID  STATE  IDLE TIME  LOGON TIME
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		user := fields[0]
		terminal := ""
		host := ""
		started := ""

		if len(fields) >= 4 {
			terminal = fields[1]
			if strings.Contains(terminal, "#") {
				host = "remote"
			}
		}
		if len(fields) >= 6 {
			started = strings.Join(fields[len(fields)-2:], " ")
		}

		result = append(result, LoggedInUser{
			User:     user,
			Terminal: terminal,
			Host:     host,
			Started:  started,
		})
	}

	if result == nil {
		result = []LoggedInUser{}
	}
	return result, nil
}

// getLoggedInUsersWMI queries Win32_ComputerSystem.UserName for the active interactive user.
func getLoggedInUsersWMI() ([]LoggedInUser, error) {
	var dst []Win32_ComputerSystem
	q := wmi.CreateQuery(&dst, "")
	if err := common.WMIQueryWithTimeout(q, &dst, 2*time.Second); err != nil {
		return []LoggedInUser{}, nil
	}

	var result []LoggedInUser
	for _, cs := range dst {
		uname := strings.TrimSpace(cs.UserName)
		if uname == "" {
			continue
		}
		if idx := strings.LastIndex(uname, "\\"); idx >= 0 {
			uname = uname[idx+1:]
		}
		result = append(result, LoggedInUser{
			User:     uname,
			Terminal: "console",
			Host:     cs.Name,
			Started:  "active",
		})
	}

	if result == nil {
		result = []LoggedInUser{}
	}
	return result, nil
}
