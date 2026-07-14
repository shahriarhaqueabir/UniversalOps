package netops

import "testing"

func TestGetRoutingTable(t *testing.T) {
	routes, err := GetRoutingTable()
	if err != nil {
		t.Fatalf("GetRoutingTable error: %v", err)
	}
	if len(routes) == 0 {
		t.Error("Expected at least one route")
	}
	for _, r := range routes {
		if r.IsDefault {
			t.Logf("Default route: %s via %s dev %s", r.Destination, r.Gateway, r.Interface)
		}
	}
}
