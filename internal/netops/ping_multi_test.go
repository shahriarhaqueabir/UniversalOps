package netops

import "testing"

func TestPingMultiTarget(t *testing.T) {
	targets := []string{"127.0.0.1", "8.8.8.8"}
	results := PingMultiTarget(targets, 3)
	if len(results) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(results))
	}
	for _, r := range results {
		t.Logf("Target=%s Success=%v AvgMs=%.1f Loss=%.0f%%", r.Target, r.Success, r.AvgMs, r.PacketLoss)
	}
	stats := GetPingStats(results)
	t.Logf("Stats: AvgLatency=%.1f MaxLatency=%.1f Worst=%s", stats.AvgLatency, stats.MaxLatency, stats.WorstTarget)
}
