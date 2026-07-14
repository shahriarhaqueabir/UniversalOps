package netops

import (
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

// WiFiNetwork holds a detected WiFi network.
type WiFiNetwork struct {
	SSID      string `json:"ssid"`
	Signal    int    `json:"signal"`
	Channel   int    `json:"channel"`
	Security  string `json:"security"`
	BSSID     string `json:"bssid"`
	Frequency string `json:"frequency"`
}

// WiFiInfo holds current WiFi connection info.
type WiFiInfo struct {
	Interface string `json:"interface"`
	SSID      string `json:"ssid"`
	Signal    int    `json:"signal"`
	Speed     string `json:"speed"`
	Channel   int    `json:"channel"`
}

// ScanWiFiNetworks scans for available WiFi networks.
func ScanWiFiNetworks() ([]WiFiNetwork, error) {
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("netsh", "wlan", "show", "networks", "mode=bssid")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return nil, err
		}
		return parseWiFiWindows(string(output))
	case "linux":
		cmd := exec.Command("nmcli", "-f", "SSID,SIGNAL,CHAN,SECURITY", "dev", "wifi", "list", "--rescan", "yes")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return nil, err
		}
		return parseWiFiNmcli(string(output))
	default:
		return nil, fmt.Errorf("unsupported platform")
	}
}

func parseWiFiWindows(output string) ([]WiFiNetwork, error) {
	var networks []WiFiNetwork
	var current *WiFiNetwork
	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "SSID") && strings.Contains(trimmed, ":") {
			if current != nil {
				networks = append(networks, *current)
			}
			current = &WiFiNetwork{SSID: strings.TrimSpace(strings.SplitN(trimmed, ":", 2)[1])}
		} else if current != nil {
			if strings.HasPrefix(trimmed, "Signal") {
				pct := strings.TrimSuffix(strings.TrimSpace(strings.SplitN(trimmed, ":", 2)[1]), "%")
				current.Signal, _ = strconv.Atoi(pct)
			} else if strings.HasPrefix(trimmed, "Channel") {
				current.Channel, _ = strconv.Atoi(strings.TrimSpace(strings.SplitN(trimmed, ":", 2)[1]))
			} else if strings.HasPrefix(trimmed, "Authentication") {
				current.Security = strings.TrimSpace(strings.SplitN(trimmed, ":", 2)[1])
			}
		}
	}
	if current != nil {
		networks = append(networks, *current)
	}
	return networks, nil
}

func parseWiFiNmcli(output string) ([]WiFiNetwork, error) {
	var networks []WiFiNetwork
	for _, line := range strings.Split(output, "\n")[1:] {
		fields := strings.Fields(strings.TrimSpace(line))
		if len(fields) >= 2 {
			networks = append(networks, WiFiNetwork{
				SSID:   fields[0],
				Signal: parseIntSafe(fields[1]),
			})
		}
	}
	return networks, nil
}

// GetWiFiInfo returns info about the current WiFi connection.
func GetWiFiInfo() (*WiFiInfo, error) {
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("netsh", "wlan", "show", "interfaces")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return nil, err
		}
		return parseWiFiInfoWindows(string(output))
	case "linux":
		cmd := exec.Command("nmcli", "-t", "-f", "ACTIVE,SSID,SIGNAL,SPEED,CHAN", "dev", "wifi")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return nil, err
		}
		for _, line := range strings.Split(string(output), "\n") {
			if strings.HasPrefix(line, "yes:") {
				parts := strings.Split(line, ":")
				if len(parts) >= 5 {
					return &WiFiInfo{SSID: parts[1], Signal: parseIntSafe(parts[2]), Speed: parts[3], Channel: parseIntSafe(parts[4])}, nil
				}
			}
		}
		return nil, fmt.Errorf("not connected to WiFi")
	default:
		return nil, fmt.Errorf("unsupported platform")
	}
}

func parseWiFiInfoWindows(output string) (*WiFiInfo, error) {
	info := &WiFiInfo{}
	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(trimmed, "SSID") && strings.Contains(trimmed, ":"):
			info.SSID = strings.TrimSpace(strings.SplitN(trimmed, ":", 2)[1])
		case strings.HasPrefix(trimmed, "Signal"):
			pct := strings.TrimSuffix(strings.TrimSpace(strings.SplitN(trimmed, ":", 2)[1]), "%")
			info.Signal, _ = strconv.Atoi(pct)
		case strings.HasPrefix(trimmed, "Speed"):
			info.Speed = strings.TrimSpace(strings.SplitN(trimmed, ":", 2)[1])
		case strings.HasPrefix(trimmed, "Channel"):
			info.Channel, _ = strconv.Atoi(strings.TrimSpace(strings.SplitN(trimmed, ":", 2)[1]))
		}
	}
	if info.SSID == "" {
		return nil, fmt.Errorf("not connected to WiFi")
	}
	return info, nil
}

func parseIntSafe(s string) int {
	n, _ := strconv.Atoi(strings.TrimSpace(s))
	return n
}
