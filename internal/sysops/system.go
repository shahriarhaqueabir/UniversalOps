package sysops

import (
	"fmt"

	"github.com/shirou/gopsutil/v4/host"
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
	Percent     int     `json:"percent"`
	Charging    bool    `json:"charging"`
	DesignCap   uint32  `json:"design_cap"`
	FullCap     uint32  `json:"full_cap"`
	WearLevel   float64 `json:"wear_level"`
	CycleCount  uint32  `json:"cycle_count"`
	Chemistry   string  `json:"chemistry"`
	Temperature float64 `json:"temperature"`
}

type PhysicalDisk struct {
	DeviceID     string `json:"device_id"`
	Model        string `json:"model"`
	MediaType    string `json:"media_type"`
	Size         uint64 `json:"size"`
	Status       string `json:"status"`       // OK, Degraded, etc
	PredictFail  bool   `json:"predict_fail"` // SMART predictive failure
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

// LoggedInUser holds information about a logged-in user.
type LoggedInUser struct {
	User     string `json:"user"`
	Terminal string `json:"terminal"`
	Host     string `json:"host"`
	Started  string `json:"started"`
}

// GetLoggedInUsers returns all currently logged-in users (cross-platform via gopsutil).
func GetLoggedInUsers() ([]LoggedInUser, error) {
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
