package netops

import (
	"fmt"
	"net"
	"time"

	gopsnet "github.com/shirou/gopsutil/v4/net"
)

const (
	bandwidthHistoryLimit   = 24
	bandwidthSparklineWidth = 12
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
	Counters   map[string]bandwidthCounter
}

// GetInterfaces returns information about all network interfaces without blocking.
// It uses the provided lastCounters and elapsed duration to calculate rates.
func GetInterfaces(lastCounters map[string]bandwidthCounter, elapsed time.Duration) (BandwidthResult, error) {
	current, err := getBandwidthCounters()
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

		result = append(result, info)
	}

	return BandwidthResult{Interfaces: result, Counters: current}, nil
}

type bandwidthCounter struct {
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

func getBandwidthCounters() (map[string]bandwidthCounter, error) {
	counters, err := gopsnet.IOCounters(true)
	if err != nil {
		return nil, fmt.Errorf("get network counters: %w", err)
	}

	result := make(map[string]bandwidthCounter, len(counters))
	for _, counter := range counters {
		result[counter.Name] = bandwidthCounter{
			Name:    counter.Name,
			RXBytes: counter.BytesRecv,
			TXBytes: counter.BytesSent,
		}
	}
	return result, nil
}

func calculateBandwidthRates(before, after map[string]bandwidthCounter, elapsed time.Duration) map[string]bandwidthRate {
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

func appendRateHistory(history []float64, value float64, limit int) []float64 {
	if limit <= 0 {
		return nil
	}

	next := append(append([]float64(nil), history...), value)
	if len(next) > limit {
		next = next[len(next)-limit:]
	}
	return next
}

func mergeInterfaceBandwidthHistory(previous, current []InterfaceInfo) []InterfaceInfo {
	histories := make(map[string]InterfaceInfo, len(previous))
	for _, iface := range previous {
		histories[iface.Name] = iface
	}

	for i := range current {
		if previousIface, ok := histories[current[i].Name]; ok {
			current[i].RXHistory = appendRateHistory(previousIface.RXHistory, current[i].RXRateBps, bandwidthHistoryLimit)
			current[i].TXHistory = appendRateHistory(previousIface.TXHistory, current[i].TXRateBps, bandwidthHistoryLimit)
		}
	}
	return current
}
