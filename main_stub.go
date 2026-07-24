//go:build !windows

package main

import (
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

// checkWindowsPrereqs is a no-op on non-Windows platforms.
func checkWindowsPrereqs() {}

// backdropType returns Auto on non-Windows platforms (Mica is Windows-only).
func backdropType() windows.BackdropType {
	return windows.Auto
}
