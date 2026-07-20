package app

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
	"github.com/shahriarhaqueabir/AllOpsFull/internal/netops"
	"github.com/shahriarhaqueabir/AllOpsFull/internal/sysops"
)

// DiscoverEnvironment gathers comprehensive information about the host environment.
func (a *App) DiscoverEnvironment() (EnvReport, error) {
	common.LogInfo("DiscoverEnvironment: starting workstation audit")

	// Default values in case of discovery failure
	report := EnvReport{
		Hostname:    "Unknown-Host",
		OS:          runtime.GOOS,
		Arch:        runtime.GOARCH,
		Interfaces:  []string{},
		PackageMgrs: []string{},
		Shells:      []string{},
	}

	// 1. Basic System Info
	sysInfo, err := sysops.GetSystemInfo()
	if err != nil {
		common.LogWarn("DiscoverEnvironment: GetSystemInfo failed: %v", err)
	} else if sysInfo != nil {
		report.Hostname = sysInfo.Hostname
		if sysInfo.Platform != "" {
			report.OS = sysInfo.Platform
		}
		if sysInfo.KernelArch != "" {
			report.Arch = sysInfo.KernelArch
		}
	}

	// 2. CPU Info
	cpuStats, err := sysops.GetCPUStats()
	if err != nil {
		common.LogWarn("DiscoverEnvironment: GetCPUStats failed: %v", err)
		report.CPU = "Unknown CPU"
	} else if cpuStats != nil {
		report.CPU = cpuStats.ModelName
		report.Cores = cpuStats.LogicalCores
	}

	// 3. Memory Info
	memStats, err := sysops.GetMemoryStats()
	if err != nil {
		common.LogWarn("DiscoverEnvironment: GetMemoryStats failed: %v", err)
		report.Memory = "Unknown RAM"
	} else if memStats != nil {
		report.Memory = fmt.Sprintf("%.1f GB", float64(memStats.TotalBytes)/(1024*1024*1024))
	}

	// 4. Network Interfaces
	ifaceResult, err := netops.GetInterfaces(nil, 0)
	if err != nil {
		common.LogWarn("DiscoverEnvironment: GetInterfaces failed: %v", err)
	} else {
		for _, iface := range ifaceResult.Interfaces {
			if iface.IsUp && len(iface.IPs) > 0 {
				report.Interfaces = append(report.Interfaces, fmt.Sprintf("%s (%s)", iface.Name, strings.Join(iface.IPs, ", ")))
			}
		}
	}

	// 5. Package Managers
	pkgs := sysops.GetInstalledPackages()
	for _, pm := range pkgs {
		if pm.Found {
			report.PackageMgrs = append(report.PackageMgrs, pm.Name)
		}
	}

	// 6. Shells
	report.Shells = discoverShells()

	common.LogInfo("DiscoverEnvironment: audit complete (OS: %s, CPU: %s)", report.OS, report.CPU)
	return report, nil
}

func discoverShells() []string {
	var shells []string
	candidates := []string{"powershell", "pwsh", "cmd", "bash", "zsh", "sh", "fish"}

	if runtime.GOOS == "windows" {
		// Windows specific: check for PowerShell, CMD, Git Bash
		for _, s := range []string{"powershell", "pwsh", "cmd"} {
			if _, err := exec.LookPath(s); err == nil {
				shells = append(shells, s)
			}
		}
		// Check for Git Bash common location
		gitBash := "C:\\Program Files\\Git\\bin\\bash.exe"
		if _, err := os.Stat(gitBash); err == nil {
			shells = append(shells, "git-bash")
		}
	} else {
		// Unix specific
		for _, s := range candidates {
			if _, err := exec.LookPath(s); err == nil {
				shells = append(shells, s)
			}
		}
	}

	return shells
}

// DiscoverEnvironmentStandalone (standalone for testing or non-App use if needed)
func DiscoverEnvironmentStandalone() (EnvReport, error) {
	return (&App{}).DiscoverEnvironment()
}
