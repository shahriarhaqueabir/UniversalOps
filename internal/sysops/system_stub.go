//go:build !windows

package sysops

// GetBaseboardInfo is not available on non-Windows platforms.
func GetBaseboardInfo() *BaseboardInfo {
	return nil
}

// GetPhysicalDisks is not available on non-Windows platforms.
func GetPhysicalDisks() ([]PhysicalDisk, error) {
	return []PhysicalDisk{}, nil
}

// GetDetailedBatteryHealth is not available on non-Windows platforms.
func GetDetailedBatteryHealth() (*BatteryHealth, error) {
	return nil, nil
}
