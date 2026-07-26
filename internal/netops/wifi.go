package netops

import (
	"bufio"
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/shahriarhaqueabir/UniversalOps/internal/common"
)

// WiFiNetwork represents a visible WiFi access point.
type WiFiNetwork struct {
	SSID      string `json:"ssid"`
	Signal    int    `json:"signal"`
	Channel   int    `json:"channel"`
	Security  string `json:"security"`
	BSSID     string `json:"bssid"`
	Frequency string `json:"frequency"`
}

// WiFiInfo represents the current WiFi adapter connection state.
type WiFiInfo struct {
	Interface string `json:"interface"`
	SSID      string `json:"ssid"`
	Signal    int    `json:"signal"`
	Speed     string `json:"speed"`
	Channel   int    `json:"channel"`
}

// ScanWiFiNetworks scans for visible WiFi networks using netsh (Windows) or iwlist (Linux).
func ScanWiFiNetworks() ([]WiFiNetwork, error) {
	if runtime.GOOS != "windows" {
		return scanWiFiLinux()
	}
	return scanWiFiWindows()
}

// scanWiFiWindows uses netsh wlan show networks mode=bssid to list visible networks.
func scanWiFiWindows() ([]WiFiNetwork, error) {
	cmd := common.HiddenCommand("netsh", "wlan", "show", "networks", "mode=bssid")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("netsh wlan scan failed: %w", err)
	}
	return parseNetshBssidOutput(string(output))
}

// parseNetshBssidOutput parses the output of "netsh wlan show networks mode=bssid".
func parseNetshBssidOutput(output string) ([]WiFiNetwork, error) {
	var networks []WiFiNetwork
	var currentSSID, currentAuth string
	var inNetwork bool

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Detect "SSID N : Name" — start of a new network
		if strings.HasPrefix(trimmed, "SSID") && strings.Contains(trimmed, ":") {
			// Save any previously parsed network
			if currentSSID != "" && inNetwork {
				// A new SSID means we changed networks; the previous one is complete
			}
			parts := strings.SplitN(trimmed, ":", 2)
			if len(parts) == 2 {
				currentSSID = strings.TrimSpace(parts[1])
				inNetwork = true
			}
			continue
		}

		if !inNetwork {
			continue
		}

		// Capture Authentication (security type)
		if strings.HasPrefix(trimmed, "Authentication") {
			parts := strings.SplitN(trimmed, ":", 2)
			if len(parts) == 2 {
				currentAuth = strings.TrimSpace(parts[1])
			}
			continue
		}

		// Capture BSSID line: "BSSID N : aa:bb:cc:dd:ee:ff"
		if strings.HasPrefix(trimmed, "BSSID") && strings.Contains(trimmed, ":") {
			parts := strings.SplitN(trimmed, ":", 2)
			if len(parts) == 2 {
				bssid := strings.TrimSpace(parts[1])
				// BSSIDs are colon-separated hex; if not, it's not a BSSID line
				if !isMACAddress(bssid) {
					continue
				}
				// Parse signal, channel, frequency from indented lines below
				network := WiFiNetwork{
					SSID:     currentSSID,
					Security: currentAuth,
					BSSID:    bssid,
				}
				networks = append(networks, network)
			}
			continue
		}

		// Parse Signal (%) — only for lines indented under a BSSID
		if strings.HasPrefix(trimmed, "Signal") && len(networks) > 0 {
			parts := strings.SplitN(trimmed, ":", 2)
			if len(parts) == 2 {
				sigStr := strings.TrimSpace(parts[1])
				sigStr = strings.TrimSuffix(sigStr, "%")
				sigStr = strings.TrimSpace(sigStr)
				if sig, err := strconv.Atoi(sigStr); err == nil {
					networks[len(networks)-1].Signal = sig
				}
			}
			continue
		}

		// Parse Channel
		if strings.HasPrefix(trimmed, "Channel") && len(networks) > 0 {
			parts := strings.SplitN(trimmed, ":", 2)
			if len(parts) == 2 {
				chStr := strings.TrimSpace(parts[1])
				if ch, err := strconv.Atoi(chStr); err == nil {
					networks[len(networks)-1].Channel = ch
				}
			}
			continue
		}

		// Parse Band (frequency)
		if strings.HasPrefix(trimmed, "Band") && len(networks) > 0 {
			parts := strings.SplitN(trimmed, ":", 2)
			if len(parts) == 2 {
				networks[len(networks)-1].Frequency = strings.TrimSpace(parts[1])
			}
			continue
		}
	}

	if len(networks) == 0 {
		return networks, nil
	}

	// Deduplicate by SSID (take the best signal per SSID)
	seen := make(map[string]bool)
	deduped := make([]WiFiNetwork, 0, len(networks))
	for _, n := range networks {
		if !seen[n.SSID] {
			seen[n.SSID] = true
			deduped = append(deduped, n)
		}
	}
	return deduped, nil
}

func isMACAddress(s string) bool {
	parts := strings.Split(s, ":")
	if len(parts) != 6 {
		return false
	}
	for _, p := range parts {
		if len(p) != 2 {
			return false
		}
		for _, c := range p {
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
				return false
			}
		}
	}
	return true
}

// scanWiFiLinux uses iwlist to scan for WiFi networks (basic support).
func scanWiFiLinux() ([]WiFiNetwork, error) {
	cmd := exec.Command("sudo", "iwlist", "scan")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("iwlist scan failed: %w", err)
	}
	return parseIwlistOutput(string(output))
}

func parseIwlistOutput(output string) ([]WiFiNetwork, error) {
	var networks []WiFiNetwork
	var current netopsWiFiBuilder

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, "Cell") && strings.Contains(line, "Address:") {
			if current.SSID != "" {
				networks = append(networks, current.build())
			}
			current = netopsWiFiBuilder{}
			parts := strings.SplitN(line, "Address:", 2)
			if len(parts) == 2 {
				current.BSSID = strings.TrimSpace(parts[1])
			}
			continue
		}

		if strings.HasPrefix(line, "ESSID:") {
			essid := strings.TrimPrefix(line, "ESSID:")
			current.SSID = strings.Trim(strings.TrimSpace(essid), `"`)
			continue
		}

		if strings.HasPrefix(line, "Quality=") {
			// Quality=70/70  Signal level=-40 dBm
			for _, part := range strings.Fields(line) {
				if strings.HasPrefix(part, "Signal level=") {
					sigStr := strings.TrimPrefix(part, "Signal level=")
					if idx := strings.Index(sigStr, "/"); idx > 0 {
						sigStr = sigStr[:idx]
					}
					if idx := strings.Index(sigStr, " "); idx > 0 {
						sigStr = sigStr[:idx]
					}
					// Signal level is typically negative dBm, convert to 0-100 scale
					sigVal, err := strconv.Atoi(sigStr)
					if err == nil {
						if sigVal < 0 {
							current.Signal = min(100, max(0, sigVal+100))
						} else {
							current.Signal = sigVal
						}
					}
				}
			}
			continue
		}

		if strings.HasPrefix(line, "Encryption key:") {
			val := strings.TrimPrefix(line, "Encryption key:")
			if strings.TrimSpace(val) == "on" {
				current.Security = "WPA" // approximate
			} else {
				current.Security = "Open"
			}
			continue
		}

		if strings.HasPrefix(line, "IE: IEEE 802.11i/WPA2") || strings.HasPrefix(line, "IE: WPA2") {
			if current.Security == "" || current.Security == "WPA" {
				current.Security = "WPA2"
			}
			continue
		}

		if strings.HasPrefix(line, "Channel:") {
			chStr := strings.TrimPrefix(line, "Channel:")
			if ch, err := strconv.Atoi(strings.TrimSpace(chStr)); err == nil {
				current.Channel = ch
			}
			continue
		}
	}

	if current.SSID != "" {
		networks = append(networks, current.build())
	}

	return networks, nil
}

type netopsWiFiBuilder struct {
	SSID     string
	Signal   int
	Channel  int
	Security string
	BSSID    string
}

func (b netopsWiFiBuilder) build() WiFiNetwork {
	return WiFiNetwork{
		SSID:      b.SSID,
		Signal:    b.Signal,
		Channel:   b.Channel,
		Security:  b.Security,
		BSSID:     b.BSSID,
		Frequency: "",
	}
}

// GetWiFiInfo returns info about the current WiFi connection.
func GetWiFiInfo() (WiFiInfo, error) {
	if runtime.GOOS != "windows" {
		return WiFiInfo{}, fmt.Errorf("wifi info not implemented on %s", runtime.GOOS)
	}
	return getWiFiInfoWindows()
}

// getWiFiInfoWindows uses netsh wlan show interfaces to get current connection info.
func getWiFiInfoWindows() (WiFiInfo, error) {
	cmd := common.HiddenCommand("netsh", "wlan", "show", "interfaces")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return WiFiInfo{}, fmt.Errorf("netsh wlan interfaces failed: %w", err)
	}
	return parseNetshInterfacesOutput(string(output))
}

// parseNetshInterfacesOutput parses "netsh wlan show interfaces" output.
func parseNetshInterfacesOutput(output string) (WiFiInfo, error) {
	info := WiFiInfo{}
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.Contains(line, ":") {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		switch key {
		case "Name":
			if info.Interface == "" {
				info.Interface = val
			}
		case "SSID":
			if val != "" {
				info.SSID = val
			}
		case "Signal":
			sigStr := strings.TrimSuffix(val, "%")
			sigStr = strings.TrimSpace(sigStr)
			if sig, err := strconv.Atoi(sigStr); err == nil {
				info.Signal = sig
			}
		case "Channel":
			if ch, err := strconv.Atoi(val); err == nil {
				info.Channel = ch
			}
		case "Receive rate (Mbps)":
			if info.Speed == "" {
				info.Speed = val + " Mbps"
			}
		case "Transmit rate (Mbps)":
			if info.Speed != "" && !strings.Contains(info.Speed, "/") {
				info.Speed = info.Speed + " / " + val + " Mbps"
			}
		}
	}
	return info, nil
}
