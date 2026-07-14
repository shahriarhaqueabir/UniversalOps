package netops

import (
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/net"
)

// BandwidthSample holds a single bandwidth measurement.
type BandwidthSample struct {
	Timestamp     time.Time `json:"timestamp"`
	RxBytesPerSec float64   `json:"rx_bytes_per_sec"`
	TxBytesPerSec float64   `json:"tx_bytes_per_sec"`
	Interface     string    `json:"interface"`
}

var (
	bandwidthHistory []BandwidthSample
	bandwidthMu      sync.Mutex
	monitoringTicker *time.Ticker
	monitoringDone   chan struct{}
)

// GetBandwidthHistory returns the recorded bandwidth samples.
func GetBandwidthHistory() []BandwidthSample {
	bandwidthMu.Lock()
	defer bandwidthMu.Unlock()
	result := make([]BandwidthSample, len(bandwidthHistory))
	copy(result, bandwidthHistory)
	return result
}

// StartMonitoring begins periodic bandwidth sampling.
func StartMonitoring(intervalSec int) {
	bandwidthMu.Lock()
	defer bandwidthMu.Unlock()
	if monitoringTicker != nil {
		return
	}
	if intervalSec <= 0 {
		intervalSec = 5
	}
	monitoringTicker = time.NewTicker(time.Duration(intervalSec) * time.Second)
	monitoringDone = make(chan struct{})
	go func() {
		var prevBytes net.IOCountersStat
		first := true
		for {
			select {
			case <-monitoringTicker.C:
				counters, err := net.IOCounters(false)
				if err != nil || len(counters) == 0 {
					continue
				}
				total := counters[0]
				if !first {
					bandwidthMu.Lock()
					bandwidthHistory = append(bandwidthHistory, BandwidthSample{
						Timestamp:     time.Now(),
						RxBytesPerSec: float64(total.BytesRecv-prevBytes.BytesRecv) / float64(intervalSec),
						TxBytesPerSec: float64(total.BytesSent-prevBytes.BytesSent) / float64(intervalSec),
						Interface:     "total",
					})
					if len(bandwidthHistory) > 360 {
						bandwidthHistory = bandwidthHistory[len(bandwidthHistory)-360:]
					}
					bandwidthMu.Unlock()
				}
				prevBytes = total
				first = false
			case <-monitoringDone:
				return
			}
		}
	}()
}

// StopMonitoring stops the bandwidth monitoring loop.
func StopMonitoring() {
	bandwidthMu.Lock()
	defer bandwidthMu.Unlock()
	if monitoringTicker != nil {
		monitoringTicker.Stop()
		close(monitoringDone)
		monitoringTicker = nil
		monitoringDone = nil
	}
}
