package netops

import "testing"

func TestGetARPTable(t *testing.T) {
	entries, err := GetARPTable()
	if err != nil {
		t.Logf("GetARPTable error (may need admin): %v", err)
		return
	}
	t.Logf("Found %d ARP entries", len(entries))
	for i, e := range entries {
		if i >= 3 {
			break
		}
		t.Logf("  IP=%s MAC=%s Vendor=%s", e.IP, e.MAC, e.Vendor)
	}
}
