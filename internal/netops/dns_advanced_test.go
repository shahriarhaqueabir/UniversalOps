package netops

import "testing"

func TestReverseLookup(t *testing.T) {
	result, err := ReverseLookup("8.8.8.8")
	if err != nil {
		t.Logf("ReverseLookup error: %v", err)
		return
	}
	t.Logf("PTR for 8.8.8.8: %s", result)
}

func TestFlushDNSCache(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping DNS flush in short mode")
	}
	err := FlushDNSCache()
	if err != nil {
		t.Logf("FlushDNSCache error (may need admin): %v", err)
	}
}
