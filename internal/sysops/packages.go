package sysops

import (
	"os/exec"
	"runtime"
	"strings"
)

// PackageInfo holds info about a single installed package.
type PackageInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// PackageManagerInfo holds info about a detected package manager.
type PackageManagerInfo struct {
	Name     string        `json:"name"`
	Found    bool          `json:"found"`
	Packages []PackageInfo `json:"packages"`
}

// GetInstalledPackages detects available package managers and lists installed packages.
func GetInstalledPackages() []PackageManagerInfo {
	var managers []PackageManagerInfo

	switch runtime.GOOS {
	case "linux":
		managers = append(managers, getAptPackages())
		managers = append(managers, getDnfPackages())
		managers = append(managers, getPacmanPackages())
	case "windows":
		managers = append(managers, getWingetPackages())
		managers = append(managers, getChocoPackages())
	}

	return managers
}

func getAptPackages() PackageManagerInfo {
	result := PackageManagerInfo{Name: "apt", Found: false}
	cmd := exec.Command("dpkg", "--get-selections")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return result
	}
	result.Found = true

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) >= 2 && parts[1] == "install" {
			result.Packages = append(result.Packages, PackageInfo{Name: parts[0]})
		}
	}
	return result
}

func getDnfPackages() PackageManagerInfo {
	result := PackageManagerInfo{Name: "dnf", Found: false}
	cmd := exec.Command("dnf", "list", "installed", "-q")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return result
	}
	result.Found = true

	lines := strings.Split(string(output), "\n")
	for _, line := range lines[1:] {
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			result.Packages = append(result.Packages, PackageInfo{Name: parts[0], Version: parts[1]})
		}
	}
	return result
}

func getPacmanPackages() PackageManagerInfo {
	result := PackageManagerInfo{Name: "pacman", Found: false}
	cmd := exec.Command("pacman", "-Q")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return result
	}
	result.Found = true

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			result.Packages = append(result.Packages, PackageInfo{Name: parts[0], Version: parts[1]})
		}
	}
	return result
}

func getWingetPackages() PackageManagerInfo {
	result := PackageManagerInfo{Name: "winget", Found: false}
	cmd := exec.Command("winget", "list", "--accept-source-agreements")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return result
	}
	result.Found = true

	lines := strings.Split(string(output), "\n")
	for _, line := range lines[3:] {
		parts := strings.Fields(line)
		if len(parts) >= 3 {
			result.Packages = append(result.Packages, PackageInfo{Name: parts[0], Version: parts[1]})
		}
	}
	return result
}

func getChocoPackages() PackageManagerInfo {
	result := PackageManagerInfo{Name: "choco", Found: false}
	cmd := exec.Command("choco", "list", "--local-only")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return result
	}
	result.Found = true

	lines := strings.Split(string(output), "\n")
	for _, line := range lines[2:] {
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			result.Packages = append(result.Packages, PackageInfo{Name: parts[0], Version: parts[1]})
		}
	}
	return result
}
