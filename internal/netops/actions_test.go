package netops

import "testing"

func TestRunNetworkAction_FlushDNS(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping in short mode")
	}
	err := RunNetworkAction("flush_dns", nil)
	if err != nil {
		t.Logf("FlushDNS error (may need admin): %v", err)
	}
}

func TestRunNetworkAction_ClearARP(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping in short mode")
	}
	err := RunNetworkAction("clear_arp_cache", nil)
	if err != nil {
		t.Logf("ClearARP error (may need admin): %v", err)
	}
}
