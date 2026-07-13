package common

import (
	"os"
	"path/filepath"
	"runtime"
)

// Platform holds OS detection information.
type PlatformInfo struct {
	OS   string // "windows", "linux", "darwin"
	Arch string // "amd64", "arm64", "386"
}

// DetectPlatform returns the current OS and architecture.
func DetectPlatform() PlatformInfo {
	return PlatformInfo{
		OS:   runtime.GOOS,
		Arch: runtime.GOARCH,
	}
}

// IsWindows returns true if running on Windows.
func IsWindows() bool {
	return runtime.GOOS == "windows"
}

// IsLinux returns true if running on Linux.
func IsLinux() bool {
	return runtime.GOOS == "linux"
}

// IsMacOS returns true if running on macOS.
func IsMacOS() bool {
	return runtime.GOOS == "darwin"
}

// IsAdminRequired returns true if the current operation typically needs
// elevated privileges on the current OS.
func IsAdminRequired() bool {
	// Most system operations on Windows and Linux need elevated privileges
	return IsWindows() || IsLinux()
}

// ConfigDir returns the application config directory path.
func ConfigDir() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "opsforall"), nil
}

// IsOnboarded checks if the onboarding marker file exists.
func IsOnboarded() bool {
	dir, err := ConfigDir()
	if err != nil {
		return false
	}
	_, err = os.Stat(filepath.Join(dir, ".onboarded"))
	return err == nil
}

// MarkOnboarded creates the onboarding marker file.
func MarkOnboarded() error {
	dir, err := ConfigDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, ".onboarded"), []byte{}, 0644)
}

// ClearOnboarded removes the onboarding marker file.
func ClearOnboarded() error {
	dir, err := ConfigDir()
	if err != nil {
		return err
	}
	return os.Remove(filepath.Join(dir, ".onboarded"))
}
