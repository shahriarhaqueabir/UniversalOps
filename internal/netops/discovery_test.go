package netops

import "testing"

func TestRunNetworkDiscovery(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping discovery test in short mode")
	}
	result := RunNetworkDiscovery("192.168.1")
	t.Logf("Found %d devices on %s in %dms", len(result.Devices), result.Subnet, result.ScanTimeMs)
	for _, d := range result.Devices {
		t.Logf("  IP=%s MAC=%s Vendor=%s", d.IP, d.MAC, d.Vendor)
	}
}
