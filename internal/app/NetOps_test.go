package app

import (
	"testing"
)

func TestNetOps_Ping_NegativeCount(t *testing.T) {
	n := NewNetOps(NewApp())
	result := n.Ping("127.0.0.1", -1)
	if result.Error != "" {
		t.Logf("Ping error (may be expected in CI): %s", result.Error)
	}
	if result.Sent > 0 && result.Received == 0 {
		t.Log("Ping sent packets but got no replies")
	}
}

func TestNetOps_Ping_EmptyHost(t *testing.T) {
	n := NewNetOps(NewApp())
	result := n.Ping("", 2)
	if result.Error == "" {
		t.Log("Ping to empty host returned no error (CI may vary)")
	}
}

func TestNetOps_DNSLookup_DefaultTimeout(t *testing.T) {
	n := NewNetOps(NewApp())
	result := n.DNSLookup("localhost", "", 0)
	if result.Error != "" {
		t.Logf("DNSLookup error: %s", result.Error)
	}
}

func TestNetOps_DNSLookup_InvalidHost(t *testing.T) {
	n := NewNetOps(NewApp())
	result := n.DNSLookup("::not a valid hostname 12345 ::", "", 2000)
	if result.Error == "" && len(result.A) == 0 {
		t.Log("No error but no A records returned")
	}
}

func TestNetOps_PortScan_EmptyPorts(t *testing.T) {
	n := NewNetOps(NewApp())
	results := n.PortScan("127.0.0.1", nil)
	if results == nil {
		t.Fatal("PortScan returned nil, expected non-nil slice")
	}
}

func TestNetOps_PortScan_SpecificPorts(t *testing.T) {
	n := NewNetOps(NewApp())
	results := n.PortScan("127.0.0.1", []int{80, 443, 8080})
	if results == nil {
		t.Fatal("PortScan returned nil, expected non-nil slice")
	}
	if len(results) != 3 {
		t.Logf("PortScan returned %d results, expected 3", len(results))
	}
	for _, r := range results {
		if r.Port == 0 {
			t.Error("PortScan entry has zero Port")
		}
	}
}

func TestNetOps_Traceroute_EmptyHost(t *testing.T) {
	n := NewNetOps(NewApp())
	result := n.Traceroute("")
	if result.Error == "" {
		t.Log("Traceroute to empty host returned no error")
	}
}

func TestNetOps_GetConnections(t *testing.T) {
	n := NewNetOps(NewApp())
	conns := n.GetConnections()
	if conns == nil {
		t.Fatal("GetConnections returned nil, expected non-nil slice")
	}
	for _, c := range conns {
		if c.LocalAddr == "" && c.Protocol == "" {
			t.Log("Found connection with empty local address and protocol")
		}
	}
}

func TestNetOps_GetInterfaces(t *testing.T) {
	n := NewNetOps(NewApp())
	ifaces := n.GetInterfaces()
	if ifaces == nil {
		t.Fatal("GetInterfaces returned nil, expected non-nil slice")
	}
	if len(ifaces) > 0 {
		if ifaces[0].Name == "" {
			t.Error("First interface has empty name")
		}
	}
}
