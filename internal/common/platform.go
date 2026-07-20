package common

import (
	"os"
	"os/exec"
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

// IsAdmin returns true if the process has administrative/root privileges.
func IsAdmin() bool {
	if runtime.GOOS == "windows" {
		// Windows check via shell command as fallback if syscalls aren't easily bundled
		cmd := exec.Command("net", "session")
		err := cmd.Run()
		return err == nil
	}
	// Unix/Linux/Darwin
	return os.Geteuid() == 0
}


// GetBaseDir returns the application base directory. It prefers a local 'data'
// directory for portable mode, but falls back to the OS-standard UserConfigDir.
func GetBaseDir() string {
	// Check for portable mode first
	if _, err := os.Stat("data"); err == nil {
		return "."
	}

	// Fallback to standard OS directory
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "." // Last resort
	}
	return filepath.Join(configDir, "allopsfull")
}

// ConfigDir returns the application data directory path.
func ConfigDir() (string, error) {
	base := GetBaseDir()
	if base == "." {
		return "data", nil
	}
	return filepath.Join(base, "data"), nil
}

// LogsDir returns the application logs directory path.
func LogsDir() (string, error) {
	base := GetBaseDir()
	if base == "." {
		return "logs", nil
	}
	return filepath.Join(base, "logs"), nil
}

// IsOnboarded checks if the onboarding marker file exists in the data dir.
func IsOnboarded() bool {
	dir, _ := ConfigDir()
	_, err := os.Stat(filepath.Join(dir, ".onboarded"))
	return err == nil
}

// MarkOnboarded creates the onboarding marker file.
func MarkOnboarded() error {
	dir, _ := ConfigDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, ".onboarded"), []byte{}, 0644)
}

// ClearOnboarded removes the onboarding marker.
func ClearOnboarded() error {
	dir, _ := ConfigDir()
	return os.Remove(filepath.Join(dir, ".onboarded"))
}
