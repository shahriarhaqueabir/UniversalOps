package secops

import (
	"testing"
)

func TestExtractPort(t *testing.T) {
	tests := []struct {
		addr string
		want string
	}{
		{"0.0.0.0:80", "80"},
		{"127.0.0.1:443", "443"},
		{"[::]:22", "22"},
		{"[::1]:3389", "3389"},
		{"0.0.0.0:0", "0"},
		{"", ""},
		{"no port here", ""}, // no colon, returns empty
	}
	for _, tt := range tests {
		got := extractPort(tt.addr)
		if got != tt.want {
			t.Errorf("extractPort(%q) = %q, want %q", tt.addr, got, tt.want)
		}
	}
}

func TestExtractPortFromSS(t *testing.T) {
	tests := []struct {
		addr string
		want string
	}{
		{"0.0.0.0:80", "80"},
		{"[::]:22", "22"},
		{"*:443", "443"},
		{"127.0.0.1:8080", "8080"},
		{"", ""},
	}
	for _, tt := range tests {
		got := extractPortFromSS(tt.addr)
		if got != tt.want {
			t.Errorf("extractPortFromSS(%q) = %q, want %q", tt.addr, got, tt.want)
		}
	}
}

func TestRiskForPort(t *testing.T) {
	tests := []struct {
		port int
		want string
	}{
		{22, "low"},      // SSH
		{23, "high"},     // Telnet
		{3389, "high"},   // RDP
		{21, "high"},     // FTP
		{5900, "medium"}, // VNC (not in map, defaults to medium)
		{80, "medium"},   // HTTP
		{443, "low"},     // HTTPS
		{8080, "medium"}, // HTTP-Alt
		{8443, "low"},    // HTTPS-Alt
		{3306, "high"},   // MySQL
		{5432, "high"},   // PostgreSQL
		{6379, "high"},   // Redis
		{27017, "high"},  // MongoDB
		{53, "medium"},   // DNS
		{123, "medium"},  // NTP (not in map, defaults to medium)
		{3000, "medium"}, // Common dev port (not in map)
		{9999, "medium"}, // Unknown port
	}
	for _, tt := range tests {
		got := riskForPort(tt.port)
		if got != tt.want {
			t.Errorf("riskForPort(%d) = %q, want %q", tt.port, got, tt.want)
		}
	}
}

func TestServiceNameForPort(t *testing.T) {
	tests := []struct {
		port int
		want string
	}{
		{22, "SSH"},
		{80, "HTTP"},
		{443, "HTTPS"},
		{3389, "RDP"},
		{3306, "MySQL"},
		{5432, "PostgreSQL"},
		{6379, "Redis"},
		{27017, "MongoDB"},
		{9999, ""},
	}
	for _, tt := range tests {
		got := serviceNameForPort(tt.port)
		if got != tt.want {
			t.Errorf("serviceNameForPort(%d) = %q, want %q", tt.port, got, tt.want)
		}
	}
}

func TestParseListeningPorts(t *testing.T) {
	output := `
  Proto  Local Address          Foreign Address        State           PID
  TCP    0.0.0.0:80             0.0.0.0:0              LISTENING       1234
  TCP    0.0.0.0:443            0.0.0.0:0              LISTENING       5678
  TCP    [::]:22                [::]:0                 LISTENING       9012
  UDP    0.0.0.0:53             *:*                                    1111
`
	ports := parseListeningPorts(output)
	if len(ports) == 0 {
		t.Fatal("parseListeningPorts returned no ports")
	}
	// Verify we found the listening TCP ports
	found := false
	for _, p := range ports {
		if p.Port == 80 && p.Protocol == "TCP" && p.PID == 1234 {
			found = true
			break
		}
	}
	if !found {
		t.Error("parseListeningPorts did not find TCP port 80 with PID 1234")
	}
}

func TestParseListeningPorts_Empty(t *testing.T) {
	ports := parseListeningPorts("")
	if len(ports) != 0 {
		t.Errorf("parseListeningPorts(empty) returned %d ports, want 0", len(ports))
	}
}

func TestParseTasklistCSV(t *testing.T) {
	output := `"Image Name","PID","Session Name","Session#","Mem Usage"
"System Idle Process",0,"Services",0,"24 K"
"System",4,"Services",0,"8 K"
"chrome.exe",1234,"Console",1,"200 K"
"notepad.exe",5678,"Console",1,"50 K"
`
	procs := parseTasklistCSV(output)
	if len(procs) != 4 {
		t.Errorf("parseTasklistCSV returned %d processes, want 4", len(procs))
	}
	if procs[1234] != "chrome" {
		t.Errorf("parseTasklistCSV[1234] = %q, want chrome", procs[1234])
	}
	if procs[5678] != "notepad" {
		t.Errorf("parseTasklistCSV[5678] = %q, want notepad", procs[5678])
	}
}

func TestParseTasklistCSV_Empty(t *testing.T) {
	procs := parseTasklistCSV("")
	if len(procs) != 0 {
		t.Errorf("parseTasklistCSV(empty) returned %d processes, want 0", len(procs))
	}
}

func TestParseSSOutput(t *testing.T) {
	output := `Netid  State   Recv-Q  Send-Q  Local Address:Port   Peer Address:Port   Process
tcp    LISTEN  0       128     0.0.0.0:22           0.0.0.0:*           users:(("sshd",pid=123,fd=3))
`
	ports, err := parseSSOutput(output)
	if err != nil {
		t.Fatalf("parseSSOutput failed: %v", err)
	}
	if len(ports) == 0 {
		t.Fatal("parseSSOutput returned no ports")
	}
	if ports[0].Port != 22 {
		t.Errorf("parseSSOutput port = %d, want 22", ports[0].Port)
	}
	if ports[0].Protocol != "TCP" {
		t.Errorf("parseSSOutput protocol = %q, want TCP", ports[0].Protocol)
	}
}

func TestParseSSOutput_Empty(t *testing.T) {
	ports, err := parseSSOutput("")
	if err != nil {
		t.Fatalf("parseSSOutput(empty) failed: %v", err)
	}
	if len(ports) != 0 {
		t.Errorf("parseSSOutput(empty) returned %d ports, want 0", len(ports))
	}
}
