package netops

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// ARPEntry holds a single ARP cache entry.
type ARPEntry struct {
	IP        string `json:"ip"`
	MAC       string `json:"mac"`
	Vendor    string `json:"vendor"`
	Interface string `json:"interface"`
}

// GetARPTable returns the system ARP table with vendor resolution.
func GetARPTable() ([]ARPEntry, error) {
	switch runtime.GOOS {
	case "windows":
		return getARPTableWindows()
	case "linux":
		return getARPTableLinux()
	default:
		return nil, fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

func getARPTableWindows() ([]ARPEntry, error) {
	cmd := exec.Command("arp", "-a")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("arp command failed: %w", err)
	}
	var entries []ARPEntry
	for _, line := range strings.Split(string(output), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Interface") || strings.HasPrefix(line, "---") || strings.HasPrefix(line, "Internet") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			e := ARPEntry{IP: fields[0], MAC: fields[1]}
			if len(fields) >= 3 {
				e.Interface = fields[2]
			}
			e.Vendor = LookupVendor(e.MAC)
			entries = append(entries, e)
		}
	}
	return entries, nil
}

func getARPTableLinux() ([]ARPEntry, error) {
	cmd := exec.Command("ip", "neigh", "show")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("ip neigh failed: %w", err)
	}
	var entries []ARPEntry
	for _, line := range strings.Split(string(output), "\n") {
		fields := strings.Fields(strings.TrimSpace(line))
		if len(fields) < 5 {
			continue
		}
		e := ARPEntry{IP: fields[0]}
		for i, f := range fields {
			if f == "lladdr" && i+1 < len(fields) {
				e.MAC = fields[i+1]
			}
			if f == "dev" && i+1 < len(fields) {
				e.Interface = fields[i+1]
			}
		}
		if e.MAC != "" {
			e.Vendor = LookupVendor(e.MAC)
		}
		entries = append(entries, e)
	}
	return entries, nil
}
