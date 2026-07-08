package netops

import (
	"fmt"
	"net"
	"strings"
	"time"

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

	interfaces, err := net.Interfaces()
	if err != nil {
		return BandwidthResult{}, fmt.Errorf("get interfaces: %w", err)
	}

	var rates map[string]bandwidthRate
	if lastCounters != nil && elapsed > 0 {
		rates = calculateBandwidthRates(lastCounters, current, elapsed)
	}

	var result []InterfaceInfo
	for _, iface := range interfaces {
		mACStr := ""
		if iface.HardwareAddr != nil {
			mACStr = iface.HardwareAddr.String()
		}

		info := InterfaceInfo{
			Name:  iface.Name,
			MAC:   mACStr,
			IsUp:  iface.Flags&net.FlagUp != 0,
			MTU:   iface.MTU,
			Flags: iface.Flags.String(),
		}

		if rate, ok := rates[iface.Name]; ok {
			info.RXBytes = rate.RXBytes
			info.TXBytes = rate.TXBytes
			info.RXRateBps = rate.RXRateBps
			info.TXRateBps = rate.TXRateBps
		}

		// Get IP addresses
		addrs, err := iface.Addrs()
		if err == nil {
			for _, addr := range addrs {
				info.IPs = append(info.IPs, addr.String())
			}
		}

		// Speed is not directly available via stdlib, estimate from flags
		if iface.Flags&net.FlagLoopback != 0 {
			info.Speed = "N/A (loopback)"
		} else if info.IsUp {
			info.Speed = "unknown"
		}

		// Heuristic labeler for UI
		if iface.Flags&net.FlagLoopback != 0 {
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
