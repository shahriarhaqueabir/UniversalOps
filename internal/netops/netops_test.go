package netops

import (
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
)

func TestParsePingOutput(t *testing.T) {
	tests := []struct {
		name    string
		target  string
		output  string
		count   int
		want    *PingResult
		wantErr bool
	}{
		{
			name:   "windows-english-success",
			target: "8.8.8.8",
			output: "Pinging 8.8.8.8 with 32 bytes of data:\n" +
				"Reply from 8.8.8.8: bytes=32 time=10ms TTL=117\n" +
				"Reply from 8.8.8.8: bytes=32 time=15ms TTL=117\n" +
				"Reply from 8.8.8.8: bytes=32 time=20ms TTL=117\n" +
				"Reply from 8.8.8.8: bytes=32 time=15ms TTL=117\n" +
				"\n" +
				"Ping statistics for 8.8.8.8:\n" +
				"    Packets: Sent = 4, Received = 4, Lost = 0 (0% loss),\n" +
				"Approximate round trip times in milli-seconds:\n" +
				"    Minimum = 10ms, Maximum = 20ms, Average = 15ms\n",
			count: 4,
			want: &PingResult{
				Target:   "8.8.8.8",
				IP:       "8.8.8.8",
				Sent:     4,
				Received: 4,
				Lost:     0,
				Min:      10 * time.Millisecond,
				Max:      20 * time.Millisecond,
				Avg:      15 * time.Millisecond,
				TTL:      117,
			},
		},
		{
			name:   "partial-loss",
			target: "8.8.8.8",
			output: "Pinging 8.8.8.8 with 32 bytes of data:\n" +
				"Reply from 8.8.8.8: bytes=32 time=10ms TTL=117\n" +
				"Request timed out.\n" +
				"Reply from 8.8.8.8: bytes=32 time=15ms TTL=117\n" +
				"Request timed out.\n" +
				"\n" +
				"Ping statistics for 8.8.8.8:\n" +
				"    Packets: Sent = 4, Received = 2, Lost = 2 (50% loss),\n" +
				"Approximate round trip times in milli-seconds:\n" +
				"    Minimum = 10ms, Maximum = 15ms, Average = 12ms\n",
			count: 4,
			want: &PingResult{
				Target:   "8.8.8.8",
				IP:       "8.8.8.8",
				Sent:     4,
				Received: 2,
				Lost:     2,
				Min:      10 * time.Millisecond,
				Max:      15 * time.Millisecond,
				Avg:      12 * time.Millisecond,
				TTL:      117,
			},
		},
		{
			name:   "ipv6-format",
			target: "google.com",
			output: "Pinging google.com [2607:f8b0:4004:800::200e] with 32 bytes of data:\n" +
				"Reply from 2607:f8b0:4004:800::200e: time=14ms\n" +
				"Reply from 2607:f8b0:4004:800::200e: time=12ms\n" +
				"Reply from 2607:f8b0:4004:800::200e: time=13ms\n" +
				"Reply from 2607:f8b0:4004:800::200e: time=11ms\n" +
				"\n" +
				"Ping statistics for 2607:f8b0:4004:800::200e:\n" +
				"    Packets: Sent = 4, Received = 4, Lost = 0 (0% loss),\n" +
				"Approximate round trip times in milli-seconds:\n" +
				"    Minimum = 11ms, Maximum = 14ms, Average = 12ms\n",
			count: 4,
			want: &PingResult{
				Target:   "google.com",
				IP:       "", // IPv6 addresses are not parsed by current regex
				Sent:     4,
				Received: 4,
				Lost:     0,
				Min:      11 * time.Millisecond,
				Max:      14 * time.Millisecond,
				Avg:      12 * time.Millisecond,
				TTL:      0, // IPv6 replies don't include TTL in this format
			},
		},
		{
			name:   "ip-extracted-from-brackets",
			target: "google.com",
			output: "Pinging google.com [8.8.8.8] with 32 bytes of data:\n" +
				"Reply from 8.8.8.8: bytes=32 time=5ms TTL=118\n" +
				"Reply from 8.8.8.8: bytes=32 time=6ms TTL=118\n" +
				"\n" +
				"Ping statistics for 8.8.8.8:\n" +
				"    Packets: Sent = 2, Received = 2, Lost = 0 (0% loss),\n" +
				"Approximate round trip times in milli-seconds:\n" +
				"    Minimum = 5ms, Maximum = 6ms, Average = 5ms\n",
			count: 2,
			want: &PingResult{
				Target:   "google.com",
				IP:       "8.8.8.8",
				Sent:     2,
				Received: 2,
				Lost:     0,
				Min:      5 * time.Millisecond,
				Max:      6 * time.Millisecond,
				Avg:      5 * time.Millisecond,
				TTL:      118,
			},
		},
		{
			name:   "unparseable-output",
			target: "unknown-host",
			output: "Ping request could not find host unknown-host. Please check the name and try again.\n",
			count:  4,
			want: &PingResult{
				Target: "unknown-host",
			},
		},
		{
			name:   "reply-from-fallback",
			target: "example.com",
			output: "Reply from 93.184.216.34: bytes=32 time=42ms TTL=55\n" +
				"Reply from 93.184.216.34: bytes=32 time=43ms TTL=55\n" +
				"\n" +
				"Ping statistics for 93.184.216.34:\n" +
				"    Packets: Sent = 2, Received = 2, Lost = 0 (0% loss),\n" +
				"Approximate round trip times in milli-seconds:\n" +
				"    Minimum = 42ms, Maximum = 43ms, Average = 42ms\n",
			count: 2,
			want: &PingResult{
				Target:   "example.com",
				IP:       "93.184.216.34",
				Sent:     2,
				Received: 2,
				Lost:     0,
				Min:      42 * time.Millisecond,
				Max:      43 * time.Millisecond,
				Avg:      42 * time.Millisecond,
				TTL:      55,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parsePingOutput(tt.target, tt.output, tt.count)
			if (err != nil) != tt.wantErr {
				t.Errorf("parsePingOutput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got == nil {
				t.Fatal("parsePingOutput() returned nil")
			}

			if got.Target != tt.want.Target {
				t.Errorf("Target = %q, want %q", got.Target, tt.want.Target)
			}
			if got.IP != tt.want.IP {
				t.Errorf("IP = %q, want %q", got.IP, tt.want.IP)
			}
			if got.Sent != tt.want.Sent {
				t.Errorf("Sent = %d, want %d", got.Sent, tt.want.Sent)
			}
			if got.Received != tt.want.Received {
				t.Errorf("Received = %d, want %d", got.Received, tt.want.Received)
			}
			if got.Lost != tt.want.Lost {
				t.Errorf("Lost = %d, want %d", got.Lost, tt.want.Lost)
			}
			if got.Min != tt.want.Min {
				t.Errorf("Min = %v, want %v", got.Min, tt.want.Min)
			}
			if got.Max != tt.want.Max {
				t.Errorf("Max = %v, want %v", got.Max, tt.want.Max)
			}
			if got.Avg != tt.want.Avg {
				t.Errorf("Avg = %v, want %v", got.Avg, tt.want.Avg)
			}
			if got.TTL != tt.want.TTL {
				t.Errorf("TTL = %d, want %d", got.TTL, tt.want.TTL)
			}
		})
	}
}

func TestParsePingOutputErrors(t *testing.T) {
	tests := []struct {
		name   string
		target string
		output string
		count  int
	}{
		{
			name:   "empty-output",
			target: "8.8.8.8",
			output: "",
			count:  4,
		},
		{
			name:   "garbage-output",
			target: "8.8.8.8",
			output: "Lorem ipsum dolor sit amet.\nSome random output without ping data.\n",
			count:  4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parsePingOutput(tt.target, tt.output, tt.count)
			if err != nil {
				t.Fatalf("parsePingOutput() returned unexpected error: %v", err)
			}
			if got == nil {
				t.Fatal("parsePingOutput() returned nil")
			}
			if got.Target != tt.target {
				t.Errorf("Target = %q, want %q", got.Target, tt.target)
			}
			if got.Sent != 0 || got.Received != 0 || got.Lost != 0 {
				t.Errorf("expected zero stats, got Sent=%d Received=%d Lost=%d",
					got.Sent, got.Received, got.Lost)
			}
		})
	}
}

func TestParseNetstatOutput(t *testing.T) {
	tests := []struct {
		name    string
		output  string
		wantLen int
		check   func(*testing.T, []ConnectionInfo)
	}{
		{
			name: "tcp-and-udp",
			output: "Active Connections\n" +
				"\n" +
				"  Proto  Local Address          Foreign Address        State           PID\n" +
				"  TCP    0.0.0.0:135            0.0.0.0:0              LISTENING       1234\n" +
				"  TCP    0.0.0.0:445            0.0.0.0:0              LISTENING       4\n" +
				"  TCP    192.168.1.100:54321    8.8.8.8:443            ESTABLISHED     9012\n" +
				"  TCP    192.168.1.100:49684    20.42.65.90:443         ESTABLISHED     1111\n" +
				"  UDP    0.0.0.0:123            *:*                                    4567\n" +
				"  UDP    0.0.0.0:1900           *:*                                    7890\n",
			wantLen: 6,
			check: func(t *testing.T, conns []ConnectionInfo) {
				// First connection: TCP LISTENING on port 135
				if conns[0].LocalAddr != "0.0.0.0:135" {
					t.Errorf("conns[0].LocalAddr = %q, want %q", conns[0].LocalAddr, "0.0.0.0:135")
				}
				if conns[0].LocalPort != 135 {
					t.Errorf("conns[0].LocalPort = %d, want %d", conns[0].LocalPort, 135)
				}
				if conns[0].State != "LISTENING" {
					t.Errorf("conns[0].State = %q, want %q", conns[0].State, "LISTENING")
				}
				if conns[0].PID != 1234 {
					t.Errorf("conns[0].PID = %d, want %d", conns[0].PID, 1234)
				}

				// Established TCP connection
				if conns[2].RemoteAddr != "8.8.8.8:443" {
					t.Errorf("conns[2].RemoteAddr = %q, want %q", conns[2].RemoteAddr, "8.8.8.8:443")
				}
				if conns[2].RemotePort != 443 {
					t.Errorf("conns[2].RemotePort = %d, want %d", conns[2].RemotePort, 443)
				}
				if conns[2].State != "ESTABLISHED" {
					t.Errorf("conns[2].State = %q, want %q", conns[2].State, "ESTABLISHED")
				}

				// UDP connection: state should be "UDP" for UDP entries
				if conns[4].State != "UDP" {
					t.Errorf("conns[4].State = %q, want %q", conns[4].State, "UDP")
				}
				if conns[4].RemoteAddr != "*:*" {
					t.Errorf("conns[4].RemoteAddr = %q, want %q", conns[4].RemoteAddr, "*:*")
				}
			},
		},
		{
			name: "ipv6-addresses",
			output: "Active Connections\n" +
				"\n" +
				"  Proto  Local Address          Foreign Address        State           PID\n" +
				"  TCP    [::1]:49685            [2607:f8b0:4004:800::200e]:443  ESTABLISHED  2222\n" +
				"  TCP    [fe80::1%eth0]:546      [ff02::2]:546           ESTABLISHED     3333\n",
			wantLen: 2,
			check: func(t *testing.T, conns []ConnectionInfo) {
				if conns[0].LocalAddr != "[::1]:49685" {
					t.Errorf("conns[0].LocalAddr = %q, want %q", conns[0].LocalAddr, "[::1]:49685")
				}
				if conns[0].LocalPort != 49685 {
					t.Errorf("conns[0].LocalPort = %d, want %d", conns[0].LocalPort, 49685)
				}
				if conns[0].RemoteAddr != "[2607:f8b0:4004:800::200e]:443" {
					t.Errorf("conns[0].RemoteAddr = %q, want %q", conns[0].RemoteAddr, "[2607:f8b0:4004:800::200e]:443")
				}
				if conns[0].RemotePort != 443 {
					t.Errorf("conns[0].RemotePort = %d, want %d", conns[0].RemotePort, 443)
				}
				if conns[0].PID != 2222 {
					t.Errorf("conns[0].PID = %d, want %d", conns[0].PID, 2222)
				}
			},
		},
		{
			name:    "empty-output",
			output:  "",
			wantLen: 0,
			check:   func(t *testing.T, conns []ConnectionInfo) {},
		},
		{
			name: "header-only",
			output: "Active Connections\n" +
				"\n" +
				"  Proto  Local Address          Foreign Address        State           PID\n",
			wantLen: 0,
			check:   func(t *testing.T, conns []ConnectionInfo) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseNetstatOutput(tt.output)
			if err != nil {
				t.Fatalf("parseNetstatOutput() unexpected error: %v", err)
			}
			if len(got) != tt.wantLen {
				t.Errorf("len(connections) = %d, want %d", len(got), tt.wantLen)
			}
			if tt.check != nil {
				tt.check(t, got)
			}
		})
	}
}

func TestSplitAddrPort(t *testing.T) {
	tests := []struct {
		name     string
		addrPort string
		wantHost string
		wantPort string
	}{
		{
			name:     "ipv4-with-port",
			addrPort: "0.0.0.0:135",
			wantHost: "0.0.0.0",
			wantPort: "135",
		},
		{
			name:     "ipv4-host-with-port",
			addrPort: "192.168.1.100:443",
			wantHost: "192.168.1.100",
			wantPort: "443",
		},
		{
			name:     "ipv6-loopback-with-port",
			addrPort: "[::1]:49685",
			wantHost: "[::1]",
			wantPort: "49685",
		},
		{
			name:     "ipv6-global-with-port",
			addrPort: "[2607:f8b0:4004:800::200e]:443",
			wantHost: "[2607:f8b0:4004:800::200e]",
			wantPort: "443",
		},
		{
			name:     "ipv6-link-local-with-zone",
			addrPort: "[fe80::1%eth0]:546",
			wantHost: "[fe80::1%eth0]",
			wantPort: "546",
		},
		{
			name:     "wildcard-remote",
			addrPort: "*:*",
			wantHost: "*",
			wantPort: "*",
		},
		{
			name:     "no-port",
			addrPort: "localhost",
			wantHost: "localhost",
			wantPort: "",
		},
		{
			name:     "empty-string",
			addrPort: "",
			wantHost: "",
			wantPort: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotHost, gotPort := splitAddrPort(tt.addrPort)
			if gotHost != tt.wantHost {
				t.Errorf("splitAddrPort(%q) host = %q, want %q", tt.addrPort, gotHost, tt.wantHost)
			}
			if gotPort != tt.wantPort {
				t.Errorf("splitAddrPort(%q) port = %q, want %q", tt.addrPort, gotPort, tt.wantPort)
			}
		})
	}
}

func TestServiceFromPort(t *testing.T) {
	tests := []struct {
		name string
		port int
		want string
	}{
		{name: "ssh", port: 22, want: "SSH"},
		{name: "http", port: 80, want: "HTTP"},
		{name: "https", port: 443, want: "HTTPS"},
		{name: "dns", port: 53, want: "DNS"},
		{name: "smtp", port: 25, want: "SMTP"},
		{name: "rdp", port: 3389, want: "RDP"},
		{name: "mysql", port: 3306, want: "MySQL"},
		{name: "postgresql", port: 5432, want: "PostgreSQL"},
		{name: "redis", port: 6379, want: "Redis"},
		{name: "unknown-low", port: 1, want: "unknown"},
		{name: "unknown-high", port: 65535, want: "unknown"},
		{name: "unknown-mid", port: 9999, want: "unknown"},
		{name: "zero", port: 0, want: "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := serviceFromPort(tt.port)
			if got != tt.want {
				t.Errorf("serviceFromPort(%d) = %q, want %q", tt.port, got, tt.want)
			}
		})
	}
}

func TestDefaultScanPorts(t *testing.T) {
	ports := DefaultScanPorts()

	if len(ports) == 0 {
		t.Error("DefaultScanPorts() returned empty slice")
	}

	// Verify ports are sorted
	for i := 1; i < len(ports); i++ {
		if ports[i-1] >= ports[i] {
			t.Errorf("DefaultScanPorts() not sorted: ports[%d]=%d >= ports[%d]=%d",
				i-1, ports[i-1], i, ports[i])
			break
		}
	}

	// Verify known ports are present
	known := map[int]bool{22: false, 80: false, 443: false, 53: false}
	for _, p := range ports {
		if _, ok := known[p]; ok {
			known[p] = true
		}
	}
	for port, found := range known {
		if !found {
			t.Errorf("DefaultScanPorts() missing known port %d", port)
		}
	}
}

func TestParseTracerouteOutputWindows(t *testing.T) {
	output := "Tracing route to dns.google [8.8.8.8]\n" +
		"over a maximum of 30 hops:\n\n" +
		"  1     1 ms    <1 ms     1 ms  192.168.1.1\n" +
		"  2    12 ms    11 ms    10 ms  10.0.0.1\n" +
		"  3     *        *        *     Request timed out.\n"

	got := parseTracerouteOutput("8.8.8.8", output)
	if got.Target != "8.8.8.8" {
		t.Errorf("Target = %q, want 8.8.8.8", got.Target)
	}
	if len(got.Hops) != 3 {
		t.Fatalf("len(Hops) = %d, want 3", len(got.Hops))
	}
	if got.Hops[0].Number != 1 || got.Hops[0].IP != "192.168.1.1" {
		t.Errorf("first hop = %+v, want hop 1 at 192.168.1.1", got.Hops[0])
	}
	if len(got.Hops[0].RTTs) != 3 {
		t.Errorf("first hop RTT count = %d, want 3", len(got.Hops[0].RTTs))
	}
	if !got.Hops[2].Timed {
		t.Error("third hop should be marked timed out")
	}
}

func TestParseTracerouteOutputUnix(t *testing.T) {
	output := "traceroute to google.com (142.250.191.142), 30 hops max\n" +
		" 1  192.168.1.1  1.215 ms  1.010 ms  0.982 ms\n" +
		" 2  10.0.0.1  9.321 ms  8.500 ms  8.123 ms\n"

	got := parseTracerouteOutput("google.com", output)
	if len(got.Hops) != 2 {
		t.Fatalf("len(Hops) = %d, want 2", len(got.Hops))
	}
	if got.Hops[0].IP != "192.168.1.1" {
		t.Errorf("first hop IP = %q, want 192.168.1.1", got.Hops[0].IP)
	}
	if got.Hops[1].RTTs[0] != 9321*time.Microsecond {
		t.Errorf("second hop first RTT = %v, want 9.321ms", got.Hops[1].RTTs[0])
	}
}

func TestTracerouteCommand(t *testing.T) {
	name, args := tracerouteCommand("example.com", 12)
	if name == "" {
		t.Fatal("command name should not be empty")
	}
	joined := strings.Join(args, " ")
	if !strings.Contains(joined, "example.com") {
		t.Errorf("args = %q, want target included", joined)
	}
	if !strings.Contains(joined, "12") {
		t.Errorf("args = %q, want max hop included", joined)
	}
}

func TestCalculateBandwidthRates(t *testing.T) {
	before := map[string]bandwidthCounter{
		"eth0": {Name: "eth0", RXBytes: 1_000, TXBytes: 2_000},
		"wifi": {Name: "wifi", RXBytes: 4_000, TXBytes: 8_000},
	}
	after := map[string]bandwidthCounter{
		"eth0": {Name: "eth0", RXBytes: 3_000, TXBytes: 2_500},
		"wifi": {Name: "wifi", RXBytes: 3_000, TXBytes: 9_000},
		"new0": {Name: "new0", RXBytes: 100, TXBytes: 200},
	}

	got := calculateBandwidthRates(before, after, 2*time.Second)

	if got["eth0"].RXRateBps != 1000 {
		t.Errorf("eth0 RXRateBps = %.1f, want 1000", got["eth0"].RXRateBps)
	}
	if got["eth0"].TXRateBps != 250 {
		t.Errorf("eth0 TXRateBps = %.1f, want 250", got["eth0"].TXRateBps)
	}
	if got["wifi"].RXRateBps != 0 {
		t.Errorf("wifi RXRateBps = %.1f, want 0 after counter reset", got["wifi"].RXRateBps)
	}
	if got["wifi"].TXRateBps != 500 {
		t.Errorf("wifi TXRateBps = %.1f, want 500", got["wifi"].TXRateBps)
	}
	if got["new0"].RXRateBps != 0 || got["new0"].TXRateBps != 0 {
		t.Errorf("new interface rates = %.1f/%.1f, want zero rates", got["new0"].RXRateBps, got["new0"].TXRateBps)
	}
	if got["new0"].RXBytes != 100 || got["new0"].TXBytes != 200 {
		t.Errorf("new interface totals = %d/%d, want 100/200", got["new0"].RXBytes, got["new0"].TXBytes)
	}
}

func TestAppendRateHistory(t *testing.T) {
	got := appendRateHistory([]float64{1, 2, 3}, 4, 3)
	want := []float64{2, 3, 4}

	if len(got) != len(want) {
		t.Fatalf("len(history) = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("history[%d] = %.1f, want %.1f", i, got[i], want[i])
		}
	}
}

func TestRenderSparkline(t *testing.T) {
	tests := []struct {
		name   string
		values []float64
		width  int
		want   string
	}{
		{name: "empty", values: nil, width: 4, want: "...."},
		{name: "zero-values", values: []float64{0, 0}, width: 4, want: "...."},
		{name: "rising-values", values: []float64{0, 25, 50, 75, 100}, width: 5, want: "._-=#"},
		{name: "left-padded", values: []float64{10, 20}, width: 4, want: "..-#"},
		{name: "truncated", values: []float64{1, 2, 3, 4}, width: 2, want: "=#"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := common.RenderSparkline(tt.values, tt.width)
			if got != tt.want {
				t.Errorf("common.RenderSparkline() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNetOpsInterfaceUpdateMergesBandwidthHistory(t *testing.T) {
	m := NewModel()
	m.InterfaceData = []InterfaceInfo{{
		Name:      "eth0",
		RXHistory: []float64{10, 20},
		TXHistory: []float64{5},
	}}

	m.Update(InterfacesResultMsg{Interfaces: []InterfaceInfo{{
		Name:      "eth0",
		RXRateBps: 30,
		TXRateBps: 15,
	}}})

	if !m.Ready() {
		t.Error("model should be ready after interfaces result")
	}
	if len(m.InterfaceData) != 1 {
		t.Fatalf("len(InterfaceData) = %d, want 1", len(m.InterfaceData))
	}
	if got := m.InterfaceData[0].RXHistory; len(got) != 3 || got[0] != 10 || got[1] != 20 || got[2] != 30 {
		t.Errorf("RXHistory = %#v, want [10 20 30]", got)
	}
	if got := m.InterfaceData[0].TXHistory; len(got) != 2 || got[0] != 5 || got[1] != 15 {
		t.Errorf("TXHistory = %#v, want [5 15]", got)
	}
}

func TestNetOpsTraceRouteUpdate(t *testing.T) {
	m := NewModel()
	m.Update(TraceRouteResultMsg{
		Result: &TraceRouteResult{
			Target: "example.com",
			Hops: []TraceHop{{
				Number: 1,
				IP:     "192.168.1.1",
				RTTs:   []time.Duration{time.Millisecond},
			}},
		},
	})

	if !m.Ready() {
		t.Error("model should be ready after traceroute result")
	}
	if m.TraceResult == nil || len(m.TraceResult.Hops) != 1 {
		t.Fatalf("TraceResult = %+v, want one hop", m.TraceResult)
	}
	if got := m.String(); got != "NetOps: tab=0" {
		t.Errorf("String() = %q, want tab summary", got)
	}
}

func TestModelUpdate_Messages(t *testing.T) {
	m := NewModel()

	// PingResultMsg
	pingRes := &PingResult{Target: "8.8.8.8", Sent: 4, Received: 4}
	m.Update(PingResultMsg{Result: pingRes, Err: nil})
	if m.PingResult != pingRes || !m.ready {
		t.Error("PingResultMsg not handled correctly")
	}

	// DNSResultMsg
	dnsRes := &DNSResult{A: []string{"1.1.1.1"}}
	m.Update(DNSResultMsg{Result: dnsRes, Err: nil})
	if m.DNSResult != dnsRes {
		t.Error("DNSResultMsg not handled correctly")
	}

	// PortScanResultMsg
	portRes := []PortResult{{Port: 80, Open: true}}
	m.Update(PortScanResultMsg{Results: portRes, Err: nil})
	if len(m.PortResults) != 1 || m.PortResults[0].Port != 80 {
		t.Error("PortScanResultMsg not handled correctly")
	}

	// ConnectionsResultMsg
	connRes := []ConnectionInfo{{LocalAddr: "127.0.0.1:8080", State: "LISTENING"}}
	m.Update(ConnectionsResultMsg{Connections: connRes, Err: nil})
	if len(m.Connections) != 1 || m.Connections[0].LocalAddr != "127.0.0.1:8080" {
		t.Error("ConnectionsResultMsg not handled correctly")
	}

	// WorkflowResultMsg
	m.Update(WorkflowResultMsg{Report: "Net report", Err: nil})
	if m.workflowReport != "Net report" || !m.showReport {
		t.Error("WorkflowResultMsg not handled correctly")
	}
}

func TestModelHandleKeyPress(t *testing.T) {
	m := NewModel()

	// Tab navigation
	m.handleKeyPress(tea.KeyPressMsg{Code: tea.KeyTab})
	if m.tabIndex != 1 {
		t.Errorf("tabIndex = %d, want 1", m.tabIndex)
	}

	m.handleKeyPress(tea.KeyPressMsg{Code: tea.KeyRight})
	if m.tabIndex != 2 {
		t.Errorf("tabIndex = %d, want 2", m.tabIndex)
	}

	// Direct jump
	m.handleKeyPress(tea.KeyPressMsg{Text: "5"})
	if m.tabIndex != 4 {
		t.Errorf("tabIndex = %d, want 4", m.tabIndex)
	}

	// Backwards
	m.handleKeyPress(tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift})
	if m.tabIndex != 3 {
		t.Errorf("tabIndex = %d, want 3", m.tabIndex)
	}

	// Interface selection (tab 5)
	m.tabIndex = 5
	m.InterfaceData = []InterfaceInfo{{Name: "eth0"}, {Name: "wlan0"}}
	m.selectedIndex = 0
	m.handleKeyPress(tea.KeyPressMsg{Code: tea.KeyDown})
	if m.selectedIndex != 1 {
		t.Errorf("selectedIndex = %d, want 1", m.selectedIndex)
	}
	m.handleKeyPress(tea.KeyPressMsg{Text: "k"})
	if m.selectedIndex != 0 {
		t.Errorf("selectedIndex = %d, want 0", m.selectedIndex)
	}
}

func TestModelView(t *testing.T) {
	m := NewModel()
	m.ready = true

	t.Run("ping-tab", func(t *testing.T) {
		m.tabIndex = 0
		m.PingResult = &PingResult{Target: "8.8.8.8", IP: "8.8.8.8", Sent: 1, Received: 1}
		view := m.View(80, 24)
		if !strings.Contains(view, "Ping") || !strings.Contains(view, "8.8.8.8") {
			t.Error("Ping tab view incorrect")
		}
	})

	t.Run("dns-tab", func(t *testing.T) {
		m.tabIndex = 1
		m.DNSResult = &DNSResult{A: []string{"1.1.1.1"}}
		view := m.View(80, 24)
		if !strings.Contains(view, "DNS Lookup") || !strings.Contains(view, "1.1.1.1") {
			t.Error("DNS tab view incorrect")
		}
	})

	t.Run("portscan-tab", func(t *testing.T) {
		m.tabIndex = 2
		m.PortResults = []PortResult{{Port: 80, Open: true, Service: "HTTP"}}
		view := m.View(80, 24)
		if !strings.Contains(view, "Port Scan") || !strings.Contains(view, "80") || !strings.Contains(view, "open") {
			t.Error("Port Scan tab view incorrect")
		}
	})

	t.Run("connections-tab", func(t *testing.T) {
		m.tabIndex = 4
		m.Connections = []ConnectionInfo{{LocalAddr: "127.0.0.1:80", State: "LISTENING"}}
		view := m.View(80, 24)
		if !strings.Contains(view, "Network Connections") || !strings.Contains(view, "LISTENING") {
			t.Error("Connections tab view incorrect")
		}
	})

	t.Run("interfaces-tab", func(t *testing.T) {
		m.tabIndex = 5
		m.InterfaceData = []InterfaceInfo{{Name: "eth0", IsUp: true, IPs: []string{"192.168.1.5"}}}
		view := m.View(80, 24)
		if !strings.Contains(view, "Network Interfaces") || !strings.Contains(view, "eth0") {
			t.Error("Interfaces tab view incorrect")
		}
	})

	t.Run("report-view", func(t *testing.T) {
		m.showReport = true
		m.workflowReport = "Net health OK"
		view := m.View(80, 24)
		if !strings.Contains(view, "Network Diagnostic Report") || !strings.Contains(view, "Net health OK") {
			t.Error("Report view incorrect")
		}
	})
}
