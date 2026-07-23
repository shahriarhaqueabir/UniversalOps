package netops

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
	gopsnet "github.com/shirou/gopsutil/v4/net"
)

// InterfaceInfo holds information about a network interface.
type InterfaceInfo struct {
	Name      string
	MAC       string
	IPs       []string
	IsUp      bool
	Speed     string
	MTU       int
	Flags     string
	RXBytes   uint64
	TXBytes   uint64
	RXRateBps float64
	TXRateBps float64
	RXHistory []float64
	TXHistory []float64
}

// BandwidthResult holds the outcome of a non-blocking bandwidth capture.
type BandwidthResult struct {
	Interfaces []InterfaceInfo
	Counters   map[string]BandwidthCounter
}

// GetInterfaces returns information about all network interfaces without blocking.
// It uses the provided lastCounters and elapsed duration to calculate rates.
func GetInterfaces(lastCounters map[string]BandwidthCounter, elapsed time.Duration) (BandwidthResult, error) {
	current, err := GetBandwidthCounters()
	if err != nil {
		return BandwidthResult{}, err
	}

	// Use gopsutil for interface list to ensure name consistency with IOCounters
	gopsIfaces, err := gopsnet.Interfaces()
	if err != nil {
		return BandwidthResult{}, fmt.Errorf("get interfaces: %w", err)
	}

	var rates map[string]bandwidthRate
	if lastCounters != nil && elapsed > 0 {
		rates = calculateBandwidthRates(lastCounters, current, elapsed)
	}

	// Get link speeds via PowerShell (gopsutil doesn't expose speed on all platforms)
	speeds := getLinkSpeeds()

	var result []InterfaceInfo
	for _, iface := range gopsIfaces {
		mACStr := iface.HardwareAddr

		isUp := false
		isLoopback := false
		for _, flag := range iface.Flags {
			switch flag {
			case "up":
				isUp = true
			case "loopback":
				isLoopback = true
			}
		}

		info := InterfaceInfo{
			Name:  iface.Name,
			MAC:   mACStr,
			IsUp:  isUp,
			MTU:   iface.MTU,
			Flags: strings.Join(iface.Flags, ","),
		}

		if rate, ok := rates[iface.Name]; ok {
			info.RXBytes = rate.RXBytes
			info.TXBytes = rate.TXBytes
			info.RXRateBps = rate.RXRateBps
			info.TXRateBps = rate.TXRateBps
		}

		// Get IP addresses
		for _, addr := range iface.Addrs {
			info.IPs = append(info.IPs, addr.String())
		}

		// Speed from PowerShell Get-NetAdapter
		if isLoopback {
			info.Speed = "N/A (loopback)"
		} else if spd, ok := speeds[iface.Name]; ok && spd > 0 {
			if spd >= 1_000_000_000 {
				info.Speed = fmt.Sprintf("%.0f Gbps", float64(spd)/1_000_000_000)
			} else if spd >= 1_000_000 {
				info.Speed = fmt.Sprintf("%.0f Mbps", float64(spd)/1_000_000)
			} else {
				info.Speed = fmt.Sprintf("%.0f Kbps", float64(spd)/1_000)
			}
		} else if isUp {
			info.Speed = "unknown"
		}

		// Heuristic labeler for UI
		if isLoopback {
			info.Name = "[Loopback] " + iface.Name
		} else if strings.Contains(strings.ToLower(iface.Name), "wi-fi") || strings.Contains(strings.ToLower(iface.Name), "wlan") {
			info.Name = "[WiFi] " + iface.Name
		} else if strings.Contains(strings.ToLower(iface.Name), "ethernet") {
			info.Name = "[Wired] " + iface.Name
		}

		result = append(result, info)
	}

	return BandwidthResult{Interfaces: result, Counters: current}, nil
}

type BandwidthCounter struct {
	Name    string
	RXBytes uint64
	TXBytes uint64
}

type bandwidthRate struct {
	RXBytes   uint64
	TXBytes   uint64
	RXRateBps float64
	TXRateBps float64
}

func GetBandwidthCounters() (map[string]BandwidthCounter, error) {
	counters, err := gopsnet.IOCounters(true)
	if err != nil {
		return nil, fmt.Errorf("get network counters: %w", err)
	}

	result := make(map[string]BandwidthCounter, len(counters))
	for _, counter := range counters {
		result[counter.Name] = BandwidthCounter{Name: counter.Name, RXBytes: counter.BytesRecv, TXBytes: counter.BytesSent}
	}
	return result, nil
}

func calculateBandwidthRates(before, after map[string]BandwidthCounter, elapsed time.Duration) map[string]bandwidthRate {
	rates := make(map[string]bandwidthRate, len(after))
	if elapsed <= 0 {
		return rates
	}

	seconds := elapsed.Seconds()
	for name, current := range after {
		previous, ok := before[name]
		if !ok {
			rates[name] = bandwidthRate{RXBytes: current.RXBytes, TXBytes: current.TXBytes}
			continue
		}

		rxDelta := counterDelta(previous.RXBytes, current.RXBytes)
		txDelta := counterDelta(previous.TXBytes, current.TXBytes)
		rates[name] = bandwidthRate{
			RXBytes:   current.RXBytes,
			TXBytes:   current.TXBytes,
			RXRateBps: float64(rxDelta) / seconds,
			TXRateBps: float64(txDelta) / seconds,
		}
	}
	return rates
}

func counterDelta(before, after uint64) uint64 {
	if after < before {
		return 0
	}
	return after - before
}

// getLinkSpeeds returns a map of interface name → link speed in bits/sec
// by querying Windows PowerShell Get-NetAdapter. Returns empty map on non-Windows or error.
func getLinkSpeeds() map[string]int64 {
	speeds := make(map[string]int64)
	if runtime.GOOS != "windows" {
		return speeds
	}

	cmd := common.HiddenCommand("powershell", "-Command",
		"Get-NetAdapter | Where-Object { $_.Status -eq 'Up' } | Select-Object Name, LinkSpeed | ConvertTo-Json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return speeds
	}

	outputStr := strings.TrimSpace(string(output))
	if outputStr == "" {
		return speeds
	}

	// Handle both single object and array
	if !strings.HasPrefix(outputStr, "[") {
		outputStr = "[" + outputStr + "]"
	}

	// Simple line-based JSON extraction
	var currentName string
	for _, line := range strings.Split(outputStr, "\n") {
		line = strings.TrimSpace(line)
		if idx := strings.Index(line, "\"Name\""); idx >= 0 {
			currentName = extractJSONString(line[idx:])
		}
		if idx := strings.Index(line, "\"LinkSpeed\""); idx >= 0 {
			val := extractJSONString(line[idx:])
			if currentName != "" && val != "" {
				speeds[currentName] = parseSpeed(val)
				currentName = ""
			}
		}
	}
	return speeds
}

// parseSpeed converts "1 Gbps", "100 Mbps", etc. to bits/sec.
func parseSpeed(s string) int64 {
	s = strings.TrimSpace(s)
	s = strings.ToUpper(s)
	var multiplier int64 = 1
	var value float64

	if strings.HasSuffix(s, "GBPS") {
		s = strings.TrimSuffix(s, "GBPS")
		multiplier = 1_000_000_000
	} else if strings.HasSuffix(s, "MBPS") {
		s = strings.TrimSuffix(s, "MBPS")
		multiplier = 1_000_000
	} else if strings.HasSuffix(s, "KBPS") {
		s = strings.TrimSuffix(s, "KBPS")
		multiplier = 1_000
	}

	fmt.Sscanf(strings.TrimSpace(s), "%f", &value)
	return int64(value * float64(multiplier))
}
