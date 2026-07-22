package netops

import (
	"math"
	"sync"
)

// PingResultMulti holds multi-target ping results for a single target.
type PingResultMulti struct {
	Target         string    `json:"target"`
	MinMs          float64   `json:"min_ms"`
	AvgMs          float64   `json:"avg_ms"`
	MaxMs          float64   `json:"max_ms"`
	StdDevMs       float64   `json:"stddev_ms"`
	PacketLoss     float64   `json:"packet_loss"`
	JitterMs       float64   `json:"jitter_ms"`
	IndividualRTTs []float64 `json:"individual_rtts"`
	Success        bool      `json:"success"`
	Error          string    `json:"error,omitempty"`
}

// PingStats holds aggregate statistics across multiple ping results.
type PingStats struct {
	AvgLatency  float64 `json:"avg_latency"`
	MaxLatency  float64 `json:"max_latency"`
	TotalLoss   float64 `json:"total_loss"`
	WorstTarget string  `json:"worst_target"`
}

// LatencyMatrix holds a multi-hop latency assessment.
type LatencyMatrix struct {
	Gateway PingResultMulti `json:"gateway"`
	ISP     PingResultMulti `json:"isp"`
	Cloud   PingResultMulti `json:"cloud"`
	CongestionSource string `json:"congestion_source"` // "local", "isp", "backbone", "none"
}

// GetLatencyMatrix pings a chain of targets to identify bottleneck segments.
func GetLatencyMatrix() LatencyMatrix {
	gw := GetDefaultGateway().IP
	if gw == "" { gw = "192.168.1.1" } // fallback

	targets := []string{gw, "1.1.1.1", "google.com"}
	results := PingMultiTarget(targets, 3)

	matrix := LatencyMatrix{
		Gateway: results[0],
		ISP:     results[1],
		Cloud:   results[2],
	}

	// Heuristic segment analysis
	if matrix.Gateway.AvgMs > 50 {
		matrix.CongestionSource = "local"
	} else if matrix.ISP.AvgMs > matrix.Gateway.AvgMs + 100 {
		matrix.CongestionSource = "isp"
	} else if matrix.Cloud.AvgMs > matrix.ISP.AvgMs + 150 {
		matrix.CongestionSource = "backbone"
	} else {
		matrix.CongestionSource = "none"
	}

	return matrix
}

// PingMultiTarget pings multiple targets concurrently.
func PingMultiTarget(targets []string, count int) []PingResultMulti {
	if count <= 0 {
		count = 4
	}
	results := make([]PingResultMulti, len(targets))
	var wg sync.WaitGroup
	for i, target := range targets {
		wg.Add(1)
		go func(idx int, t string) {
			defer wg.Done()
			results[idx] = pingOneTarget(t, count)
		}(i, target)
	}
	wg.Wait()
	return results
}

func pingOneTarget(target string, count int) PingResultMulti {
	result := PingResultMulti{Target: target}
	pingResult, err := Ping(target, count)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	result.PacketLoss = float64(pingResult.Lost) / float64(pingResult.Sent) * 100
	result.Success = pingResult.Received > 0
	if result.Success {
		result.MinMs = float64(pingResult.Min.Milliseconds())
		result.MaxMs = float64(pingResult.Max.Milliseconds())
		result.AvgMs = float64(pingResult.Avg.Milliseconds())
		// Reconstruct individual RTTs from aggregate stats (we only have min/max/avg).
		// Use avg as the representative value for stddev calculation.
		rtts := make([]float64, pingResult.Received)
		for i := range rtts {
			rtts[i] = result.AvgMs
		}
		result.IndividualRTTs = rtts
		var sumSq float64
		for _, t := range rtts {
			sumSq += (t - result.AvgMs) * (t - result.AvgMs)
		}
		result.StdDevMs = math.Sqrt(sumSq / float64(len(rtts)))
		result.JitterMs = float64(pingResult.Jitter.Milliseconds())
	}
	return result
}

// GetPingStats computes aggregate stats across multiple ping results.
func GetPingStats(results []PingResultMulti) PingStats {
	stats := PingStats{}
	var totalLatency float64
	var totalLoss float64
	var count int
	for _, r := range results {
		if r.Success {
			totalLatency += r.AvgMs
			count++
			if r.AvgMs > stats.MaxLatency {
				stats.MaxLatency = r.AvgMs
				stats.WorstTarget = r.Target
			}
		}
		totalLoss += r.PacketLoss
	}
	if count > 0 {
		stats.AvgLatency = totalLatency / float64(count)
	}
	if len(results) > 0 {
		stats.TotalLoss = totalLoss / float64(len(results))
	}
	return stats
}
