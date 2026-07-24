package sysops

import (
	"encoding/json"
	"os/exec"
	"runtime"
	"strings"

	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
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

// execLookPath and execCommand are package-level vars so tests can inject fakes.
var execLookPath = exec.LookPath
var execCommand = exec.Command

// GetInstalledPackages detects available package managers and lists installed packages.
// On Windows, falls back to reading installed apps from the registry if neither
// winget nor choco is available.
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

		anyFound := false
		for _, m := range managers {
			if m.Found {
				anyFound = true
				break
			}
		}
		if !anyFound {
			managers = append(managers, getWindowsInstalledApps())
		}
	}

	return managers
}

// ── Windows Registry fallback ────────────────────────────────────────────────

// getWindowsInstalledApps reads installed applications from the Windows registry
// via PowerShell as a fallback when no package manager is detected.
func getWindowsInstalledApps() PackageManagerInfo {
	result := PackageManagerInfo{Name: "windows-installed", Found: false}

	ps, err := execLookPath("powershell")
	if err != nil {
		return result
	}

	script := `Get-ItemProperty HKLM:\Software\Microsoft\Windows\CurrentVersion\Uninstall\*,
	                 HKLM:\Software\WOW6432Node\Microsoft\Windows\CurrentVersion\Uninstall\*,
	                 HKCU:\Software\Microsoft\Windows\CurrentVersion\Uninstall\* 2>$null |
	             Where-Object { $_.DisplayName } |
	             Select-Object DisplayName, DisplayVersion |
	             ConvertTo-Json -Compress`

	out, err := execCommand(ps, "-NoProfile", "-NonInteractive", "-Command", script).CombinedOutput()
	if err != nil {
		return result
	}

	return parseWindowsInstalledApps(string(out))
}

// parseWindowsInstalledApps parses the JSON output from the registry PowerShell script.
func parseWindowsInstalledApps(rawJSON string) PackageManagerInfo {
	result := PackageManagerInfo{Name: "windows-installed", Found: true}

	if len(rawJSON) > 0 && rawJSON[0] != '[' {
		rawJSON = "[" + rawJSON + "]"
	}

	type regEntry struct {
		DisplayName    string `json:"DisplayName"`
		DisplayVersion string `json:"DisplayVersion"`
	}

	var entries []regEntry
	if err := json.Unmarshal([]byte(rawJSON), &entries); err != nil {
		return PackageManagerInfo{Name: "windows-installed", Found: false}
	}

	for _, e := range entries {
		if e.DisplayName == "" {
			continue
		}
		result.Packages = append(result.Packages, PackageInfo{
			Name:    e.DisplayName,
			Version: e.DisplayVersion,
		})
	}

	return result
}

// ── apt (dpkg) ───────────────────────────────────────────────────────────────

func getAptPackages() PackageManagerInfo {
	cmd := common.HiddenCommand("dpkg", "--get-selections")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return PackageManagerInfo{Name: "apt", Found: false}
	}
	return PackageManagerInfo{Name: "apt", Found: true, Packages: parseAptOutput(string(output))}
}

// parseAptOutput parses "dpkg --get-selections" output.
func parseAptOutput(output string) []PackageInfo {
	var pkgs []PackageInfo
	for _, line := range strings.Split(output, "\n") {
		parts := strings.Fields(line)
		if len(parts) >= 2 && parts[1] == "install" {
			pkgs = append(pkgs, PackageInfo{Name: parts[0]})
		}
	}
	return pkgs
}

// ── dnf ──────────────────────────────────────────────────────────────────────

func getDnfPackages() PackageManagerInfo {
	cmd := common.HiddenCommand("dnf", "list", "installed", "-q")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return PackageManagerInfo{Name: "dnf", Found: false}
	}
	return PackageManagerInfo{Name: "dnf", Found: true, Packages: parseDnfOutput(string(output))}
}

// parseDnfOutput parses "dnf list installed -q" output (skips header line).
func parseDnfOutput(output string) []PackageInfo {
	var pkgs []PackageInfo
	lines := strings.Split(output, "\n")
	if len(lines) < 2 {
		return pkgs
	}
	for _, line := range lines[1:] {
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			pkgs = append(pkgs, PackageInfo{Name: parts[0], Version: parts[1]})
		}
	}
	return pkgs
}

// ── pacman ───────────────────────────────────────────────────────────────────

func getPacmanPackages() PackageManagerInfo {
	cmd := common.HiddenCommand("pacman", "-Q")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return PackageManagerInfo{Name: "pacman", Found: false}
	}
	return PackageManagerInfo{Name: "pacman", Found: true, Packages: parsePacmanOutput(string(output))}
}

// parsePacmanOutput parses "pacman -Q" output.
func parsePacmanOutput(output string) []PackageInfo {
	var pkgs []PackageInfo
	for _, line := range strings.Split(output, "\n") {
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			pkgs = append(pkgs, PackageInfo{Name: parts[0], Version: parts[1]})
		}
	}
	return pkgs
}

// ── winget ───────────────────────────────────────────────────────────────────

func getWingetPackages() PackageManagerInfo {
	cmd := common.HiddenCommand("winget", "list", "--accept-source-agreements")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return PackageManagerInfo{Name: "winget", Found: false}
	}
	return PackageManagerInfo{Name: "winget", Found: true, Packages: parseWingetOutput(string(output))}
}

// parseWingetOutput parses the fixed-width table emitted by "winget list".
// Column offsets come from the header so application names containing spaces
// are kept intact and the version column is not confused with the package ID.
func parseWingetOutput(output string) []PackageInfo {
	var pkgs []PackageInfo
	lines := strings.Split(output, "\n")
	headerIndex := -1
	idOffset := -1
	versionOffset := -1
	versionEnd := -1
	for i, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		hasName, hasID, hasVersion := false, false, false
		for _, field := range fields {
			switch field {
			case "Name":
				hasName = true
			case "Id":
				hasID = true
			case "Version":
				hasVersion = true
			}
		}
		if !hasName || !hasID || !hasVersion {
			continue
		}

		headerIndex = i
		idOffset = strings.Index(line, "Id")
		versionOffset = strings.Index(line, "Version")
		versionEnd = len(line)
		for _, field := range fields {
			if field == "Name" || field == "Id" || field == "Version" {
				continue
			}
			pos := strings.Index(line[versionOffset+len("Version"):], field)
			if pos >= 0 {
				versionEnd = versionOffset + len("Version") + pos
				break
			}
		}
		break
	}
	if headerIndex < 0 || idOffset < 0 || versionOffset < 0 {
		return pkgs
	}

	for _, line := range lines[headerIndex+1:] {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.Trim(trimmed, "- ") == "" {
			continue
		}

		// Fixed-width output is the normal case. Fall back to token parsing for
		// compact output used by older winget versions and test fixtures.
		if len(line) > versionOffset {
			name := strings.TrimSpace(line[:min(idOffset, len(line))])
			version := strings.TrimSpace(line[versionOffset:min(versionEnd, len(line))])
			if name != "" && version != "" && !strings.Contains(version, " ") {
				pkgs = append(pkgs, PackageInfo{Name: name, Version: version})
				continue
			}
		}

		parts := strings.Fields(line)
		if len(parts) >= 3 {
			pkgs = append(pkgs, PackageInfo{Name: parts[0], Version: parts[2]})
		}
	}
	return pkgs
}

// ── Chocolatey ───────────────────────────────────────────────────────────────

func getChocoPackages() PackageManagerInfo {
	cmd := common.HiddenCommand("choco", "list", "--local-only")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return PackageManagerInfo{Name: "choco", Found: false}
	}
	return PackageManagerInfo{Name: "choco", Found: true, Packages: parseChocoOutput(string(output))}
}

// parseChocoOutput parses "choco list --local-only" output (skips 2 header lines).
func parseChocoOutput(output string) []PackageInfo {
	var pkgs []PackageInfo
	lines := strings.Split(output, "\n")
	if len(lines) < 3 {
		return pkgs
	}
	for _, line := range lines[2:] {
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			pkgs = append(pkgs, PackageInfo{Name: parts[0], Version: parts[1]})
		}
	}
	return pkgs
}
