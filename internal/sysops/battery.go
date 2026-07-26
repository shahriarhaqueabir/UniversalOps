//go:build windows

package sysops

import (
	"time"

	"github.com/shahriarhaqueabir/UniversalOps/internal/common"
	"github.com/yusufpapurcu/wmi"
)

// win32Battery represents a WMI battery entry.
// Uses signed int fields to avoid the reflect.Value.Uint panic
// in yusufpapurcu/wmi when WMI returns signed values.
type win32Battery struct {
	BatteryStatus            int32
	EstimatedChargeRemaining int16
	EstimatedRunTime         int32
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
func GetBatteryInfo() (result *BatteryInfo) {
	// Recover from any panic in WMI query construction (reflection-based)
	// or WMI execution, preventing a crash in the calling goroutine.
	defer func() {
		if r := recover(); r != nil {
			result = &BatteryInfo{Detected: false}
		}
	}()

	var dst []win32Battery
	q := wmi.CreateQuery(&dst, "")
	if err := common.WMIQueryWithTimeout(q, &dst, 2*time.Second); err != nil {
		return &BatteryInfo{Detected: false}
	}

	if len(dst) == 0 {
		return &BatteryInfo{Detected: false}
	}

	b := dst[0]

	// BatteryStatus bitmask: bit 1 = charging, bit 2 = fully charged, bit 3 = short-circuit
	charging := b.BatteryStatus&2 != 0 || b.BatteryStatus&6 != 0
	percent := float64(uint16(b.EstimatedChargeRemaining))

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
