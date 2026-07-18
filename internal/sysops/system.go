package sysops

import (
	"fmt"
	"strings"

	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/process"
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

// GetSystemInfo returns general system information.
func GetSystemInfo() (*SystemInfo, error) {
	info, err := host.Info()
	if err != nil {
		return nil, err
	}

	// Count processes
	procs, err := process.Processes()
	procCount := 0
	if err == nil {
		procCount = len(procs)
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
		ProcessCount:    procCount,
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

// GetLoggedInUsers returns all currently logged-in users.
func GetLoggedInUsers() ([]LoggedInUser, error) {
	users, err := host.Users()
	if err != nil {
		// gopsutil host.Users() is not implemented on Windows; fall back to `query user`
		if common.IsWindows() {
			return getLoggedInUsersWindows()
		}
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

// getLoggedInUsersWindows uses the Windows `query user` command to list logged-in users.
func getLoggedInUsersWindows() ([]LoggedInUser, error) {
	cmd := common.SandboxedCommandWithConfig(common.SystemQuerySandbox(), "query", "user")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("query user failed: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) < 2 {
		return []LoggedInUser{}, nil
	}

	var result []LoggedInUser
	// Skip header line (line 0); parse data lines
	for _, line := range lines[1:] {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// query user output is fixed-width: USERNAME  SESSIONNAME  ID  STATE  IDLE TIME  LOGON TIME
		// Typical format: " username   console   1  Active  .none   2025-01-15 09:30"
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		user := fields[0]
		terminal := ""
		host := ""
		started := ""

		// Find the session name and state
		// Typical: username console 1 Active .none 2025-01-15 09:30
		// or:      username rdp-tcp#0 1 Active .none 2025-01-15 09:30
		if len(fields) >= 4 {
			terminal = fields[1]
			// Check if session contains '#' which indicates remote
			if strings.Contains(terminal, "#") {
				host = "remote"
			}
		}
		// The last fields contain the logon time
		if len(fields) >= 6 {
			started = strings.Join(fields[len(fields)-2:], " ")
		}

		result = append(result, LoggedInUser{
			User:     user,
			Terminal: terminal,
			Host:     host,
			Started:  started,
		})
	}

	if result == nil {
		result = []LoggedInUser{}
	}
	return result, nil
}
