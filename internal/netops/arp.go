package netops

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
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
	// Try PowerShell Get-NetNeighbor first (more reliable on modern Windows)
	entries, err := getARPTablePowerShell()
	if err == nil && len(entries) > 0 {
		return entries, nil
	}

	// Fallback to arp -a command
	cmd := common.HiddenCommand("arp", "-a")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("arp command failed: %w", err)
	}

	outputStr := string(output)
	if strings.TrimSpace(outputStr) == "" {
		return []ARPEntry{}, nil
	}

	var arpEntries []ARPEntry
	for _, line := range strings.Split(outputStr, "\n") {
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
			arpEntries = append(arpEntries, e)
		}
	}
	return arpEntries, nil
}

// getARPTablePowerShell uses Get-NetNeighbor for reliable ARP data on modern Windows.
func getARPTablePowerShell() ([]ARPEntry, error) {
	cmd := common.HiddenCommand("powershell", "-Command",
		"Get-NetNeighbor -State Reachable,Stale,Unreachable | Select-Object IPAddress, LinkLayerAddress, InterfaceAlias | Where-Object { $_.IPAddress -match '^\\d+\\.\\d+\\.\\d+\\.\\d+$' } | ConvertTo-Json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("Get-NetNeighbor failed: %w", err)
	}

	outputStr := strings.TrimSpace(string(output))
	if outputStr == "" {
		return []ARPEntry{}, nil
	}

	var entries []ARPEntry
	// Handle both single object and array JSON
	if !strings.HasPrefix(outputStr, "[") {
		outputStr = "[" + outputStr + "]"
	}

	// Parse JSON manually to avoid import cycle — simple line-based extraction
	for _, line := range strings.Split(outputStr, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || line == "[" || line == "]" || line == "{" || line == "}" || strings.HasPrefix(line, "[{") || strings.HasPrefix(line, "}]") {
			continue
		}

		var ip, mac, iface string
		if idx := strings.Index(line, "\"IPAddress\""); idx >= 0 {
			val := extractJSONString(line[idx:])
			if val != "" {
				ip = val
			}
		}
		if idx := strings.Index(line, "\"LinkLayerAddress\""); idx >= 0 {
			val := extractJSONString(line[idx:])
			if val != "" {
				mac = val
			}
		}
		if idx := strings.Index(line, "\"InterfaceAlias\""); idx >= 0 {
			val := extractJSONString(line[idx:])
			if val != "" {
				iface = val
			}
		}

		if ip != "" && mac != "" {
			e := ARPEntry{IP: ip, MAC: mac, Interface: iface}
			e.Vendor = LookupVendor(mac)
			entries = append(entries, e)
		}
	}
	return entries, nil
}

// extractJSONString extracts a string value from a JSON line like "key": "value".
func extractJSONString(line string) string {
	idx := strings.Index(line, ":")
	if idx < 0 {
		return ""
	}
	rest := strings.TrimSpace(line[idx+1:])
	if len(rest) < 2 || rest[0] != '"' {
		return ""
	}
	rest = rest[1:]
	end := strings.Index(rest, "\"")
	if end < 0 {
		return ""
	}
	return rest[:end]
}

func getARPTableLinux() ([]ARPEntry, error) {
	cmd := common.HiddenCommand("ip", "neigh", "show")
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
