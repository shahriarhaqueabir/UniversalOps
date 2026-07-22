package sysops

import (
	"fmt"
	"strings"

	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/yusufpapurcu/wmi"
)

// SystemInfo holds general system information.
type SystemInfo struct {
	Hostname        string
	OS              string
	Platform        string
	PlatformVersion string
	KernelVersion   string
	KernelArch      string
	UptimeSeconds   uint64
	ProcessCount    int
	Virtualization  string
}

// BaseboardInfo holds hardware-level motherboard information.
type BaseboardInfo struct {
	Manufacturer string `json:"manufacturer"`
	Product      string `json:"product"`
	Version      string `json:"version"`
	SerialNumber string `json:"serial_number"`
	ChassisType  string `json:"chassis_type"`
}

// GetBaseboardInfo queries WMI for motherboard details.
func GetBaseboardInfo() *BaseboardInfo {
	if !common.IsWindows() {
		return nil
	}

	type Win32_BaseBoard struct {
		Manufacturer string
		Product      string
		Version      string
		SerialNumber string
	}
	var dst []Win32_BaseBoard
	if err := wmi.Query("SELECT Manufacturer, Product, Version, SerialNumber FROM Win32_BaseBoard", &dst); err != nil || len(dst) == 0 {
		return nil
	}

	type Win32_SystemEnclosure struct {
		ChassisTypes []uint16
	}
	var enc []Win32_SystemEnclosure
	chassis := "Unknown"
	if err := wmi.Query("SELECT ChassisTypes FROM Win32_SystemEnclosure", &enc); err == nil && len(enc) > 0 {
		if len(enc[0].ChassisTypes) > 0 {
			chassis = mapChassisType(enc[0].ChassisTypes[0])
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

func mapChassisType(t uint16) string {
	types := map[uint16]string{
		1: "Other", 2: "Unknown", 3: "Desktop", 4: "Low Profile Desktop", 5: "Pizza Box",
		6: "Mini Tower", 7: "Tower", 8: "Portable", 9: "Laptop", 10: "Notebook",
		11: "Hand Held", 12: "Docking Station", 13: "All in One", 14: "Sub Notebook",
		15: "Space-saving", 16: "Lunch Box", 17: "Main System Chassis",
		18: "Expansion Chassis", 19: "SubChassis", 20: "Bus Expansion Chassis",
		21: "Peripheral Chassis", 22: "Storage Chassis", 23: "Rack Mount Chassis",
		24: "Sealed-case PC",
	}
	if v, ok := types[t]; ok {
		return v
	}
	return "Unknown"
}

// BatteryHealth holds detailed battery wear data.
type BatteryHealth struct {
	Percent      int     `json:"percent"`
	Charging     bool    `json:"charging"`
	DesignCap    uint32  `json:"design_cap"`
	FullCap      uint32  `json:"full_cap"`
	WearLevel    float64 `json:"wear_level"`
	CycleCount   uint32  `json:"cycle_count"`
	Chemistry    string  `json:"chemistry"`
	Temperature  float64 `json:"temperature"`
}

// PhysicalDisk holds hardware-level disk information.
type PhysicalDisk struct {
	DeviceID     string `json:"device_id"`
	Model        string `json:"model"`
	MediaType    string `json:"media_type"`
	Size         uint64 `json:"size"`
	Status       string `json:"status"`        // OK, Degraded, etc
	PredictFail  bool   `json:"predict_fail"`  // SMART predictive failure
	SerialNumber string `json:"serial_number"`
}

// GetSystemInfo returns general system information.
func GetSystemInfo() (*SystemInfo, error) {
	info, err := host.Info()
	if err != nil {
		return nil, err
	}

	virt := ""
	if info.VirtualizationSystem != "" {
		virt = info.VirtualizationSystem + " (" + info.VirtualizationRole + ")"
	}

	return &SystemInfo{
		Hostname:        info.Hostname,
		OS:              info.OS,
		Platform:        info.Platform,
		PlatformVersion: info.PlatformVersion,
		KernelVersion:   info.KernelVersion,
		KernelArch:      info.KernelArch,
		UptimeSeconds:   info.Uptime,
		ProcessCount:    GetProcessCount(),
		Virtualization:  virt,
	}, nil
}

// GetPhysicalDisks queries WMI for physical drive health and SMART status.
func GetPhysicalDisks() ([]PhysicalDisk, error) {
	if !common.IsWindows() {
		return []PhysicalDisk{}, nil
	}

	type Win32_DiskDrive struct {
		DeviceID     string
		Model        string
		MediaType    string
		Size         uint64
		Status       string
		SerialNumber string
	}
	var dst []Win32_DiskDrive
	q := "SELECT DeviceID, Model, MediaType, Size, Status, SerialNumber FROM Win32_DiskDrive"
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}

	// Secondary query for Predictive Failure (SMART)
	type MSStorageDriver_FailurePredictStatus struct {
		InstanceName string
		PredictFailure bool
	}
	var smart []MSStorageDriver_FailurePredictStatus
	_ = wmi.QueryNamespace("SELECT InstanceName, PredictFailure FROM MSStorageDriver_FailurePredictStatus", &smart, "root\\wmi")

	res := make([]PhysicalDisk, len(dst))
	for i, d := range dst {
		res[i] = PhysicalDisk{
			DeviceID:     d.DeviceID,
			Model:        d.Model,
			MediaType:    d.MediaType,
			Size:         d.Size,
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
	if !common.IsWindows() {
		return nil, nil
	}

	type Win32_Battery struct {
		EstimatedChargeRemaining uint32
		BatteryStatus            uint16
		DesignCapacity           uint32
		FullChargeCapacity       uint32
		CycleCount               uint32
		Chemistry                uint16
	}
	var dst []Win32_Battery
	if err := wmi.Query("SELECT EstimatedChargeRemaining, BatteryStatus, DesignCapacity, FullChargeCapacity, CycleCount, Chemistry FROM Win32_Battery", &dst); err != nil || len(dst) == 0 {
		return nil, nil
	}

	b := dst[0]
	health := &BatteryHealth{
		Percent:    int(b.EstimatedChargeRemaining),
		Charging:   b.BatteryStatus == 2 || b.BatteryStatus == 6 || b.BatteryStatus == 7,
		DesignCap:  b.DesignCapacity,
		FullCap:    b.FullChargeCapacity,
		CycleCount: b.CycleCount,
	}

	if b.DesignCapacity > 0 {
		health.WearLevel = 100.0 - (float64(b.FullChargeCapacity) / float64(b.DesignCapacity) * 100.0)
	}

	return health, nil
}

// LoggedInUser holds information about a logged-in user.
type LoggedInUser struct {
	User     string `json:"user"`
	Terminal string `json:"terminal"`
	Host     string `json:"host"`
	Started  string `json:"started"`
}

// GetLoggedInUsers returns all currently logged-in users.
func GetLoggedInUsers() ([]LoggedInUser, error) {
	// Priority: WMI (Most reliable on Windows 11 for interactive sessions)
	if common.IsWindows() {
		users, err := getLoggedInUsersWMI()
		if err == nil && len(users) > 0 {
			return users, nil
		}
		// Fallback to query user if WMI fails or returns nothing
		return getLoggedInUsersWindows()
	}

	users, err := host.Users()
	if err != nil {
		return nil, err
	}

	var result []LoggedInUser
	for _, u := range users {
		result = append(result, LoggedInUser{
			User:     u.User,
			Terminal: u.Terminal,
			Host:     u.Host,
			Started:  fmt.Sprintf("%d", u.Started),
		})
	}
	return result, nil
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
		// Typical format: " username   console   1  Active  .none   2025-01-15 09:30"
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		user := fields[0]
		terminal := ""
		host := ""
		started := ""

		// Find the session name and state
		// Typical: username console 1 Active .none 2025-01-15 09:30
		// or:      username rdp-tcp#0 1 Active .none 2025-01-15 09:30
		if len(fields) >= 4 {
			terminal = fields[1]
			// Check if session contains '#' which indicates remote
			if strings.Contains(terminal, "#") {
				host = "remote"
			}
		}
		// The last fields contain the logon time
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

// Win32_ComputerSystem holds minimal WMI data for user detection.
type Win32_ComputerSystem struct {
	Name     string
	UserName string // Format: "DOMAIN\username"
}

// getLoggedInUsersWMI queries Win32_ComputerSystem.UserName for the active interactive user.
// This is the reliable fallback when `query user` / `quser` is unavailable.
func getLoggedInUsersWMI() ([]LoggedInUser, error) {
	var dst []Win32_ComputerSystem
	q := wmi.CreateQuery(&dst, "")
	if err := wmi.Query(q, &dst); err != nil {
		return []LoggedInUser{}, nil
	}

	var result []LoggedInUser
	for _, cs := range dst {
		uname := strings.TrimSpace(cs.UserName)
		if uname == "" {
			continue
		}
		// Strip DOMAIN\ prefix
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

