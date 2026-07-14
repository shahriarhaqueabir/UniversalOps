package netops

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// RouteEntry holds a single routing table entry.
type RouteEntry struct {
	Destination string `json:"destination"`
	Mask        string `json:"mask"`
	Gateway     string `json:"gateway"`
	Interface   string `json:"interface"`
	Metric      int    `json:"metric"`
	IsDefault   bool   `json:"is_default"`
}

// GetRoutingTable returns the system routing table.
func GetRoutingTable() ([]RouteEntry, error) {
	switch runtime.GOOS {
	case "windows":
		return getRoutingTableWindows()
	case "linux":
		return getRoutingTableLinux()
	default:
		return nil, fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

func getRoutingTableWindows() ([]RouteEntry, error) {
	cmd := exec.Command("netstat", "-rn")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	var entries []RouteEntry
	inIPv4 := false
	for _, line := range strings.Split(string(output), "\n") {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "IPv4") {
			inIPv4 = true
			continue
		}
		if strings.Contains(line, "IPv6") {
			inIPv4 = false
			continue
		}
		if !inIPv4 || line == "" || strings.HasPrefix(line, "Network") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 4 {
			e := RouteEntry{Destination: fields[0], Mask: fields[1], Gateway: fields[2], Interface: fields[3]}
			if len(fields) >= 5 {
				fmt.Sscanf(fields[4], "%d", &e.Metric)
			}
			if e.Destination == "0.0.0.0" {
				e.IsDefault = true
			}
			entries = append(entries, e)
		}
	}
	return entries, nil
}

func getRoutingTableLinux() ([]RouteEntry, error) {
	cmd := exec.Command("ip", "route", "show")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	var entries []RouteEntry
	for _, line := range strings.Split(string(output), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "default") {
			entries = append(entries, RouteEntry{Destination: "0.0.0.0/0", IsDefault: true})
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 1 {
			continue
		}
		e := RouteEntry{Destination: fields[0]}
		for i := 1; i < len(fields)-1; i++ {
			switch fields[i] {
			case "via":
				e.Gateway = fields[i+1]
			case "dev":
				e.Interface = fields[i+1]
			case "metric":
				if i+1 < len(fields) {
					fmt.Sscanf(fields[i+1], "%d", &e.Metric)
				}
			}
		}
		entries = append(entries, e)
	}
	return entries, nil
}

// ManageStaticRoutes adds or deletes a static route.
func ManageStaticRoutes(action, dest, mask, gateway string) error {
	switch runtime.GOOS {
	case "windows":
		switch action {
		case "add":
			_, err := exec.Command("netsh", "interface", "ipv4", "add", "route", dest, mask, gateway).CombinedOutput()
			return err
		case "delete":
			_, err := exec.Command("netsh", "interface", "ipv4", "delete", "route", dest, mask, gateway).CombinedOutput()
			return err
		}
	case "linux":
		switch action {
		case "add":
			_, err := exec.Command("sudo", "ip", "route", "add", dest, "via", gateway).CombinedOutput()
			return err
		case "delete":
			_, err := exec.Command("sudo", "ip", "route", "del", dest, "via", gateway).CombinedOutput()
			return err
		}
	}
	return fmt.Errorf("unknown action or platform: %s/%s", action, runtime.GOOS)
}
