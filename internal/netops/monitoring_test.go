package netops

import (
	"testing"
	"time"
)

func TestGetBandwidthHistory(t *testing.T) {
	StartMonitoring(1)
	time.Sleep(3 * time.Second)
	StopMonitoring()
	history := GetBandwidthHistory()
	t.Logf("Got %d bandwidth samples", len(history))
	for _, s := range history {
		t.Logf("  %.0f Bps rx, %.0f Bps tx", s.RxBytesPerSec, s.TxBytesPerSec)
	}
}
