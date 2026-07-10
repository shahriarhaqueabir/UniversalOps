package netops

import (
	"os/exec"
	"runtime"
	"testing"

	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
)

func TestPing(t *testing.T) {
	// On Linux CI (GitHub Actions, etc.), ping often requires CAP_NET_RAW or root.
	// Test using ICMP socket first (needs root), fall back to ping exec binary (also needs
	// elevated on some systems). Skip if neither works.
	if runtime.GOOS == "linux" {
		// Check if ping binary exists at all
		if _, err := exec.LookPath("ping"); err != nil {
			t.Skipf("Skipping ping test: ping binary not found on this system: %v", err)
		}
		// Use the same sandbox configuration as the real Ping() function
		cfg := common.SystemQuerySandbox()
		cfg.DenyNetworkAccess = false
		cmd := common.SandboxedCommandWithConfig(cfg, "ping", "-c", "1", "-W", "1", "127.0.0.1")
		if err := cmd.Run(); err != nil {
			t.Skipf("Skipping ping test: ping sandbox exec failed on this system: %v", err)
		}
	}

	target := "127.0.0.1"
	count := 1

	result, err := Ping(target, count)
	if err != nil {
		t.Fatalf("Ping failed: %v", err)
	}

	if result.Target != target {
		t.Errorf("Expected target %s, got %s", target, result.Target)
	}

	if result.Sent != count {
		t.Errorf("Expected sent %d, got %d", count, result.Sent)
	}

	if result.Received > count {
		t.Errorf("Received count %d exceeds sent count %d", result.Received, result.Sent)
	}
}

func TestParsePingOutput_Windows(t *testing.T) {
	output := `
Pinging 8.8.8.8 with 32 bytes of data:
Reply from 8.8.8.8: bytes=32 time=12ms TTL=117

Ping statistics for 8.8.8.8:
    Packets: Sent = 1, Received = 1, Lost = 0 (0% loss),
Approximate round trip times in milli-seconds:
    Minimum = 12ms, Maximum = 12ms, Average = 12ms
`
	result, err := parsePingOutput("8.8.8.8", output, 1)
	if err != nil {
		t.Fatalf("parsePingOutput failed: %v", err)
	}

	if result.IP != "8.8.8.8" {
		t.Errorf("Expected IP 8.8.8.8, got %s", result.IP)
	}
	if result.Sent != 1 {
		t.Errorf("Expected sent 1, got %d", result.Sent)
	}
	if result.Received != 1 {
		t.Errorf("Expected received 1, got %d", result.Received)
	}
	if result.Avg.Milliseconds() != 12 {
		t.Errorf("Expected avg 12ms, got %v", result.Avg)
	}
	if result.TTL != 117 {
		t.Errorf("Expected TTL 117, got %d", result.TTL)
	}
}

func TestResolveTarget(t *testing.T) {
	// Test with IP
	ip := "127.0.0.1"
	resolved, err := resolveTarget(ip)
	if err != nil {
		t.Errorf("resolveTarget(IP) failed: %v", err)
	}
	if resolved != ip {
		t.Errorf("Expected %s, got %s", ip, resolved)
	}

	// Test with hostname (google.com should resolve if online)
	host := "localhost"
	resolved, err = resolveTarget(host)
	if err != nil {
		t.Logf("Skipping hostname resolution test (offline?): %v", err)
		return
	}
	if resolved == "" {
		t.Error("Expected non-empty resolved IP for localhost")
	}
}

func TestLookupDNS(t *testing.T) {
	// Skip if offline
	host := "google.com"
	result, err := LookupDNS(host)
	if err != nil {
		t.Logf("Skipping DNS lookup test (offline?): %v", err)
		return
	}

	if result.Hostname != host {
		t.Errorf("Expected hostname %s, got %s", host, result.Hostname)
	}

	if len(result.A) == 0 && len(result.AAAA) == 0 {
		t.Error("Expected at least one A or AAAA record for google.com")
	}

	// Test with specific server
	result, err = LookupDNS(host, "8.8.8.8")
	if err != nil {
		t.Logf("Skipping DNS lookup with specific server test (offline?): %v", err)
		return
	}
	if len(result.A) == 0 {
		t.Error("Expected A records for google.com using 8.8.8.8")
	}
}

func TestScanPorts(t *testing.T) {
	// Test localhost on some common ports
	target := "127.0.0.1"
	ports := []int{80, 443, 3389}

	results, err := ScanPorts(target, ports)
	if err != nil {
		t.Fatalf("PortScan failed: %v", err)
	}

	if len(results) != len(ports) {
		t.Errorf("Expected %d results, got %d", len(ports), len(results))
	}

	for i, res := range results {
		if res.Port != ports[i] {
			t.Errorf("Expected port %d, got %d", ports[i], res.Port)
		}
	}
}

func TestGetConnections(t *testing.T) {
	conns, err := GetConnections()
	if err != nil {
		t.Fatalf("GetConnections failed: %v", err)
	}

	// Should have some connections on a live system
	if len(conns) == 0 {
		t.Log("No connections found (expected on most systems)")
	}

	for _, c := range conns {
		if c.LocalAddr == "" {
			t.Error("Connection missing LocalAddr")
		}
	}
}

func TestTraceRoute(t *testing.T) {
	// Traceroute is slow and might require privileges, test parsing or short run
	target := "8.8.8.8"
	result, err := TraceRoute(target)
	if err != nil {
		t.Logf("Traceroute failed (expected if not admin or offline): %v", err)
		return
	}

	if result.Target != target {
		t.Errorf("Expected target %s, got %s", target, result.Target)
	}
}
