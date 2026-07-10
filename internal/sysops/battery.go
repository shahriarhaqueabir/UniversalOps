package sysops

import (
	"github.com/yusufpapurcu/wmi"
)

// win32Battery represents a WMI battery entry.
type win32Battery struct {
	BatteryStatus            uint16
	EstimatedChargeRemaining uint16
	EstimatedRunTime         uint32
	Status                   string
	DeviceID                 string
}

// BatteryInfo holds battery information.
type BatteryInfo struct {
	Percent     float64
	Charging    bool
	TimeLeftSec int64 // estimated seconds remaining; -1 if unknown
	Status      string
	Detected    bool
}

// GetBatteryInfo queries WMI for battery information.
// Falls back gracefully on desktops or systems without a battery.
func GetBatteryInfo() *BatteryInfo {
	var dst []win32Battery
	q := wmi.CreateQuery(&dst, "")
	if err := wmi.Query(q, &dst); err != nil {
		return &BatteryInfo{Detected: false}
	}

	if len(dst) == 0 {
		return &BatteryInfo{Detected: false}
	}

	b := dst[0]

	// BatteryStatus bitmask: bit 1 = charging, bit 2 = fully charged, bit 3 = short-circuit
	charging := b.BatteryStatus&2 != 0 || b.BatteryStatus&6 != 0
	percent := float64(b.EstimatedChargeRemaining)

	timeLeft := int64(b.EstimatedRunTime)
	// 0x7FFFFFFF means unknown on Windows
	if timeLeft == 0x7FFFFFFF {
		timeLeft = -1
	}

	statusStr := "Unknown"
	switch b.BatteryStatus {
	case 1:
		statusStr = "Discharging"
	case 2:
		statusStr = "Charging"
	case 3:
		statusStr = "On AC"
	case 4:
		statusStr = "Fully Charged"
	case 5:
		statusStr = "Low"
	case 6:
		statusStr = "Critical"
	case 7:
		statusStr = "Charging"
	case 8:
		statusStr = "Charging"
	default:
		statusStr = b.Status
	}

	return &BatteryInfo{
		Percent:     percent,
		Charging:    charging,
		TimeLeftSec: timeLeft,
		Status:      statusStr,
		Detected:    true,
	}
}
