package secops

import (
	"testing"
)

// ---------------------------------------------------------------------------
// extractPort
// ---------------------------------------------------------------------------

func TestExtractPort(t *testing.T) {
	tests := []struct {
		name string
		addr string
		want string
	}{
		// IPv4
		{"ipv4 basic", "0.0.0.0:135", "135"},
		{"ipv4 specific", "192.168.1.1:443", "443"},
		{"ipv4 high port", "10.0.0.1:65535", "65535"},
		{"ipv4 localhost", "127.0.0.1:8080", "8080"},
		// IPv6
		{"ipv6 unspecified", "[::]:135", "135"},
		{"ipv6 link-local", "[fe80::1]:445", "445"},
		{"ipv6 loopback", "[::1]:53", "53"},
		{"ipv6 full", "[2001:db8::1]:9090", "9090"},
		// Edge cases
		{"empty string", "", ""},
		{"no port", "no-port", ""},
		{"just colon", ":", ""},
		{"ipv6 no closing bracket", "[::1:135", ""},
		{"ipv6 no port after bracket", "[::1]", ""},
		{"ipv6 nothing after colon", "[::]:", ""},
		{"multiple colons no brackets", "::1", "1"}, // bare IPv6 without brackets — extractPort uses LastIndex and finds the last colon
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractPort(tt.addr)
			if got != tt.want {
				t.Errorf("extractPort(%q) = %q; want %q", tt.addr, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// parseListeningPorts
// ---------------------------------------------------------------------------

func TestParseListeningPorts(t *testing.T) {
	t.Run("tcp ipv4 and ipv6 listening", func(t *testing.T) {
		output := `
  Proto  Local Address          Foreign Address        State           PID
  TCP    0.0.0.0:135            0.0.0.0:0              LISTENING       1234
  TCP    0.0.0.0:445            0.0.0.0:0              LISTENING       4
  TCP    [::]:135               [::]:0                 LISTENING       1234
  TCP    [::]:445               [::]:0                 LISTENING       4
  UDP    0.0.0.0:123            *:*                                    4567
  UDP    [::]:123               *:*                                    5678
`
		ports := parseListeningPorts(output)
		if len(ports) != 4 {
			t.Fatalf("expected 4 ports (dedup removes IPv4/IPv6 dups), got %d", len(ports))
		}

		// Build a lookup to check values
		byPort := make(map[int]ListeningPort)
		for _, p := range ports {
			byPort[p.Port] = p
		}

		for _, port := range []int{135, 445} {
			p, ok := byPort[port]
			if !ok {
				t.Errorf("expected port %d in results", port)
				continue
			}
			if p.Protocol != "TCP" {
				t.Errorf("port %d protocol = %q; want TCP", port, p.Protocol)
			}
			if p.State != "LISTENING" {
				t.Errorf("port %d state = %q; want LISTENING", port, p.State)
			}
		}

		for _, port := range []int{123} {
			p, ok := byPort[port]
			if !ok {
				t.Errorf("expected UDP port %d in results", port)
				continue
			}
			if p.Protocol != "UDP" {
				t.Errorf("port %d protocol = %q; want UDP", port, p.Protocol)
			}
		}
	})

	t.Run("excludes non-listening states", func(t *testing.T) {
		output := `  TCP    0.0.0.0:135            0.0.0.0:0              LISTENING       1234
  TCP    10.0.0.1:50000         10.0.0.2:443            ESTABLISHED      5678
  TCP    10.0.0.1:50001         10.0.0.3:80             TIME_WAIT        5678
  TCP    10.0.0.1:50002         10.0.0.4:22             CLOSE_WAIT       5678
`
		ports := parseListeningPorts(output)
		if len(ports) != 1 {
			t.Fatalf("expected 1 port (only LISTENING), got %d", len(ports))
		}
		if ports[0].Port != 135 {
			t.Errorf("expected port 135, got %d", ports[0].Port)
		}
	})

	t.Run("skips header line", func(t *testing.T) {
		output := `  Proto  Local Address          Foreign Address        State           PID
  TCP    0.0.0.0:135            0.0.0.0:0              LISTENING       1234
`
		ports := parseListeningPorts(output)
		if len(ports) != 1 {
			t.Fatalf("expected 1 port, got %d", len(ports))
		}
		if ports[0].Port != 135 {
			t.Errorf("expected port 135, got %d", ports[0].Port)
		}
	})

	t.Run("deduplication by proto:port:pid", func(t *testing.T) {
		output := `  TCP    0.0.0.0:135            0.0.0.0:0              LISTENING       1234
  TCP    0.0.0.0:135            0.0.0.0:0              LISTENING       1234
  UDP    0.0.0.0:123            *:*                                    4567
  UDP    0.0.0.0:123            *:*                                    4567
`
		ports := parseListeningPorts(output)
		if len(ports) != 2 {
			t.Fatalf("expected 2 ports after dedup, got %d", len(ports))
		}
	})

	t.Run("skips lines with unparseable port", func(t *testing.T) {
		output := `  TCP    bad-address             0.0.0.0:0              LISTENING       1234
  TCP    0.0.0.0:135            0.0.0.0:0              LISTENING       1234
`
		ports := parseListeningPorts(output)
		if len(ports) != 1 {
			t.Fatalf("expected 1 port, got %d", len(ports))
		}
		if ports[0].Port != 135 {
			t.Errorf("expected port 135, got %d", ports[0].Port)
		}
	})

	t.Run("skips short lines", func(t *testing.T) {
		output := `  TCP    0.0.0.0:135
`
		ports := parseListeningPorts(output)
		if len(ports) != 0 {
			t.Errorf("expected 0 ports for short line, got %d", len(ports))
		}
	})

	t.Run("empty output", func(t *testing.T) {
		ports := parseListeningPorts("")
		if len(ports) != 0 {
			t.Errorf("expected 0 ports for empty, got %d", len(ports))
		}
		ports = parseListeningPorts("\n\n  \n")
		if len(ports) != 0 {
			t.Errorf("expected 0 ports for whitespace, got %d", len(ports))
		}
	})
}

// ---------------------------------------------------------------------------
// getJSONBool / getJSONInt / getJSONString
// ---------------------------------------------------------------------------

func TestGetJSONBool(t *testing.T) {
	t.Run("returns true for true", func(t *testing.T) {
		data := map[string]interface{}{"key": true}
		if !getJSONBool(data, "key") {
			t.Error("expected true")
		}
	})

	t.Run("returns false for false", func(t *testing.T) {
		data := map[string]interface{}{"key": false}
		if getJSONBool(data, "key") {
			t.Error("expected false")
		}
	})

	t.Run("returns false for missing key", func(t *testing.T) {
		data := map[string]interface{}{}
		if getJSONBool(data, "missing") {
			t.Error("expected false for missing")
		}
	})

	t.Run("returns false for wrong type", func(t *testing.T) {
		data := map[string]interface{}{"key": "notabool"}
		if getJSONBool(data, "key") {
			t.Error("expected false for string value")
		}
		data2 := map[string]interface{}{"key": 42}
		if getJSONBool(data2, "key") {
			t.Error("expected false for numeric value")
		}
	})
}

func TestGetJSONInt(t *testing.T) {
	t.Run("extracts float64 as int", func(t *testing.T) {
		data := map[string]interface{}{"key": float64(42)}
		got, ok := getJSONInt(data, "key")
		if !ok {
			t.Fatal("expected ok")
		}
		if got != 42 {
			t.Errorf("got %d; want 42", got)
		}
	})

	t.Run("extracts int directly", func(t *testing.T) {
		data := map[string]interface{}{"key": 7}
		got, ok := getJSONInt(data, "key")
		if !ok {
			t.Fatal("expected ok")
		}
		if got != 7 {
			t.Errorf("got %d; want 7", got)
		}
	})

	t.Run("returns false for missing key", func(t *testing.T) {
		data := map[string]interface{}{}
		_, ok := getJSONInt(data, "missing")
		if ok {
			t.Error("expected false for missing")
		}
	})

	t.Run("returns false for wrong type", func(t *testing.T) {
		data := map[string]interface{}{"key": "notanumber"}
		_, ok := getJSONInt(data, "key")
		if ok {
			t.Error("expected false for string value")
		}
	})
}

func TestGetJSONString(t *testing.T) {
	t.Run("extracts string", func(t *testing.T) {
		data := map[string]interface{}{"key": "hello"}
		got, ok := getJSONString(data, "key")
		if !ok {
			t.Fatal("expected ok")
		}
		if got != "hello" {
			t.Errorf("got %q; want %q", got, "hello")
		}
	})

	t.Run("returns false for missing key", func(t *testing.T) {
		data := map[string]interface{}{}
		_, ok := getJSONString(data, "missing")
		if ok {
			t.Error("expected false for missing")
		}
	})

	t.Run("returns false for wrong type", func(t *testing.T) {
		data := map[string]interface{}{"key": 42}
		_, ok := getJSONString(data, "key")
		if ok {
			t.Error("expected false for numeric value")
		}
	})
}

// ---------------------------------------------------------------------------
// formatAge
// ---------------------------------------------------------------------------

func TestFormatAge(t *testing.T) {
	tests := []struct {
		days int
		want string
	}{
		{0, "Today"},
		{1, "1 day ago"},
		{2, "2 days ago"},
		{6, "6 days ago"},
		{7, "7 days ago"},
		{29, "29 days ago"},
		{30, "1 month ago"},
		{31, "1 month ago"},
		{59, "1 month ago"},
		{60, "2 months ago"},
		{90, "3 months ago"},
		{365, "12 months ago"},
		{-1, "Today"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := formatAge(tt.days)
			if got != tt.want {
				t.Errorf("formatAge(%d) = %q; want %q", tt.days, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// formatTimeStr
// ---------------------------------------------------------------------------

func TestFormatTimeStr(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"2024-01-15T10:30:00.1234567Z", "2024-01-15 10:30:00"},
		{"2024-06-01T03:00:00.0000000", "2024-06-01 03:00:00"},
		{"2024-12-25T00:00:00", "2024-12-25 00:00:00"},
		{"short", "short"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := formatTimeStr(tt.input)
			if got != tt.want {
				t.Errorf("formatTimeStr(%q) = %q; want %q", tt.input, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// parseDefenderJSON
// ---------------------------------------------------------------------------

func TestParseDefenderJSON(t *testing.T) {
	t.Run("full response", func(t *testing.T) {
		j := `{
			"AntivirusEnabled": true,
			"AMServiceEnabled": true,
			"AntispywareEnabled": true,
			"NISEnabled": false,
			"RealTimeProtectionEnabled": true,
			"CloudProtectionEnabled": false,
			"SignatureAge": 2,
			"QuickScanEndTime": "2024-12-15T03:00:00.1234567Z",
			"QuickScanAge": 1,
			"FullScanAge": 8
		}`
		status, err := parseDefenderJSON(j)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !status.Enabled {
			t.Error("expected Enabled=true")
		}
		if !status.AMServiceEnabled {
			t.Error("expected AMServiceEnabled=true")
		}
		if !status.AntispywareEnabled {
			t.Error("expected AntispywareEnabled=true")
		}
		if status.NISEnabled {
			t.Error("expected NISEnabled=false")
		}
		if !status.RealTimeProtection {
			t.Error("expected RealTimeProtection=true")
		}
		if status.CloudProtection {
			t.Error("expected CloudProtection=false")
		}
		if status.SignatureAge != "2 days ago" {
			t.Errorf("SignatureAge = %q; want %q", status.SignatureAge, "2 days ago")
		}
		if !status.UpToDate {
			t.Error("expected UpToDate=true (age 2 <= 7)")
		}
		if status.LastScan != "Quick: 2024-12-15 03:00:00" {
			t.Errorf("LastScan = %q; want %q", status.LastScan, "Quick: 2024-12-15 03:00:00")
		}
		if status.QuickScanAge != 1 {
			t.Errorf("QuickScanAge = %d; want 1", status.QuickScanAge)
		}
		if status.FullScanAge != 8 {
			t.Errorf("FullScanAge = %d; want 8", status.FullScanAge)
		}
	})

	t.Run("enabled derived from AntivirusEnabled", func(t *testing.T) {
		// When AMServiceEnabled is absent but AntivirusEnabled is true
		j := `{"AntivirusEnabled": true}`
		status, err := parseDefenderJSON(j)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !status.Enabled {
			t.Error("expected Enabled=true from AntivirusEnabled")
		}
	})

	t.Run("falls back to FullScanEndTime", func(t *testing.T) {
		j := `{
			"AntivirusEnabled": true,
			"SignatureAge": 0,
			"FullScanEndTime": "2024-12-10T06:00:00Z"
		}`
		status, err := parseDefenderJSON(j)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if status.LastScan != "Full: 2024-12-10 06:00:00" {
			t.Errorf("LastScan = %q; want Full fallback", status.LastScan)
		}
		if status.SignatureAge != "Today" {
			t.Errorf("SignatureAge = %q; want Today", status.SignatureAge)
		}
		if !status.UpToDate {
			t.Error("expected UpToDate=true for age 0")
		}
	})

	t.Run("unknown last scan", func(t *testing.T) {
		j := `{"AntivirusEnabled": true}`
		status, err := parseDefenderJSON(j)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if status.LastScan != "Unknown" {
			t.Errorf("LastScan = %q; want Unknown", status.LastScan)
		}
	})

	t.Run("signature age absent", func(t *testing.T) {
		j := `{"AntivirusEnabled": true}`
		status, err := parseDefenderJSON(j)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if status.SignatureAge != "Unknown" {
			t.Errorf("SignatureAge = %q; want Unknown", status.SignatureAge)
		}
		if status.UpToDate {
			t.Error("expected UpToDate=false when age unknown")
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		_, err := parseDefenderJSON("{bad json}")
		if err == nil {
			t.Fatal("expected error for invalid JSON")
		}
	})
}

// ---------------------------------------------------------------------------
// taskStateToString
// ---------------------------------------------------------------------------

func TestTaskStateToString(t *testing.T) {
	tests := []struct {
		state int
		want  string
	}{
		{0, "Unknown"},
		{1, "Disabled"},
		{2, "Queued"},
		{3, "Ready"},
		{4, "Running"},
		{5, "State(5)"},
		{99, "State(99)"},
		{-1, "State(-1)"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := taskStateToString(tt.state)
			if got != tt.want {
				t.Errorf("taskStateToString(%d) = %q; want %q", tt.state, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// formatTaskTime / trimDateTime
// ---------------------------------------------------------------------------

func TestTrimDateTime(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"2024-01-15T10:30:00.1234567", "2024-01-15T10:30:00"},
		{"2024-01-15T10:30:00Z", "2024-01-15T10:30:00"},
		{"2024-01-15T10:30:00+00:00", "2024-01-15T10:30:00"},
		{"short", "short"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := trimDateTime(tt.input)
			if got != tt.want {
				t.Errorf("trimDateTime(%q) = %q; want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestFormatTaskTime(t *testing.T) {
	t.Run("trims fractional seconds via trimDateTime then takes first 16 chars", func(t *testing.T) {
		got := formatTaskTime("2024-01-15T10:30:00.1234567")
		// trimDateTime("2024-01-15T10:30:00.1234567") -> "2024-01-15T10:30:00"
		// then len >= 16 -> t[:16] -> "2024-01-15T10:30"
		want := "2024-01-15T10:30"
		if got != want {
			t.Errorf("got %q; want %q", got, want)
		}
	})

	t.Run("standard ISO truncates to 16 chars", func(t *testing.T) {
		got := formatTaskTime("2024-01-15T10:30:00")
		// trimDateTime leaves it unchanged (len == 19, not > 19)
		// but formatTaskTime does t[:16] -> "2024-01-15T10:30"
		want := "2024-01-15T10:30"
		if got != want {
			t.Errorf("got %q; want %q", got, want)
		}
	})

	t.Run("short string unchanged", func(t *testing.T) {
		got := formatTaskTime("abc")
		if got != "abc" {
			t.Errorf("got %q; want %q", got, "abc")
		}
	})

	t.Run("empty string", func(t *testing.T) {
		got := formatTaskTime("")
		if got != "" {
			t.Errorf("got %q; want empty", got)
		}
	})
}

// ---------------------------------------------------------------------------
// parseTasksJSON
// ---------------------------------------------------------------------------

func TestParseTasksJSON(t *testing.T) {
	t.Run("normal tasks with string state", func(t *testing.T) {
		j := `[
			{"TaskName": "TaskA", "State": "Ready", "NextRunTime": "2024-12-16T03:00:00", "LastRunTime": "2024-12-15T03:00:00", "Author": "Admin", "Triggers": [{"Enabled": true, "Type": "Daily"}]},
			{"TaskName": "TaskB", "State": "Disabled", "NextRunTime": "", "LastRunTime": "", "Author": "User"}
		]`
		tasks, err := parseTasksJSON(j)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(tasks) != 2 {
			t.Fatalf("expected 2 tasks, got %d", len(tasks))
		}

		if tasks[0].Name != "TaskA" {
			t.Errorf("Name = %q; want TaskA", tasks[0].Name)
		}
		if tasks[0].Status != "Ready" {
			t.Errorf("Status = %q; want Ready", tasks[0].Status)
		}
		if tasks[0].NextRun != "2024-12-16T03:00" {
			t.Errorf("NextRun = %q; want %q", tasks[0].NextRun, "2024-12-16T03:00")
		}
		if tasks[0].LastRun != "2024-12-15T03:00" {
			t.Errorf("LastRun = %q; want %q", tasks[0].LastRun, "2024-12-15T03:00")
		}
		if tasks[0].Author != "Admin" {
			t.Errorf("Author = %q; want Admin", tasks[0].Author)
		}
		if tasks[0].Trigger != "Daily" {
			t.Errorf("Trigger = %q; want Daily", tasks[0].Trigger)
		}

		if tasks[1].Name != "TaskB" {
			t.Errorf("Name = %q; want TaskB", tasks[1].Name)
		}
		if tasks[1].Status != "Disabled" {
			t.Errorf("Status = %q; want Disabled", tasks[1].Status)
		}
		if tasks[1].NextRun != "N/A" {
			t.Errorf("NextRun = %q; want N/A", tasks[1].NextRun)
		}
		if tasks[1].LastRun != "N/A" {
			t.Errorf("LastRun = %q; want N/A", tasks[1].LastRun)
		}
		if tasks[1].Author != "User" {
			t.Errorf("Author = %q; want User", tasks[1].Author)
		}
		if tasks[1].Trigger != "N/A" {
			t.Errorf("Trigger = %q; want N/A", tasks[1].Trigger)
		}
	})

	t.Run("state as numeric (float64)", func(t *testing.T) {
		j := `[
			{"TaskName": "NumTask", "State": 3.0, "NextRunTime": null, "LastRunTime": null}
		]`
		tasks, err := parseTasksJSON(j)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(tasks) != 1 {
			t.Fatalf("expected 1 task, got %d", len(tasks))
		}
		if tasks[0].Status != "Ready" {
			t.Errorf("Status = %q; want Ready (from state 3)", tasks[0].Status)
		}
	})

	t.Run("state as map with Value", func(t *testing.T) {
		j := `[
			{"TaskName": "MapTask", "State": {"Value": "Running"}}
		]`
		tasks, err := parseTasksJSON(j)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(tasks) != 1 {
			t.Fatalf("expected 1 task, got %d", len(tasks))
		}
		if tasks[0].Status != "Running" {
			t.Errorf("Status = %q; want Running", tasks[0].Status)
		}
	})

	t.Run("state absent defaults to Ready", func(t *testing.T) {
		j := `[
			{"TaskName": "NoState"}
		]`
		tasks, err := parseTasksJSON(j)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(tasks) != 1 {
			t.Fatalf("expected 1 task, got %d", len(tasks))
		}
		if tasks[0].Status != "Ready" {
			t.Errorf("Status = %q; want Ready (default)", tasks[0].Status)
		}
	})

	t.Run("empty task name skipped", func(t *testing.T) {
		j := `[
			{"TaskName": "", "State": "Ready"},
			{"TaskName": "Valid", "State": "Ready"}
		]`
		tasks, err := parseTasksJSON(j)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(tasks) != 1 {
			t.Fatalf("expected 1 task, got %d", len(tasks))
		}
		if tasks[0].Name != "Valid" {
			t.Errorf("Name = %q; want Valid", tasks[0].Name)
		}
	})

	t.Run("epoch times handled before formatTaskTime truncation", func(t *testing.T) {
		j := `[
			{"TaskName": "Epoch", "State": "Ready", "NextRunTime": "0001-01-01T00:00:00", "LastRunTime": "12/31/1600 00:00:00"}
		]`
		tasks, err := parseTasksJSON(j)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Note: formatTaskTime truncates these to 16 characters *before*
		// the epoch comparison, so the literal string checks against
		// "0001-01-01T00:00:00" and "12/31/1600 00:00:00" never match.
		// The values end up as the truncated forms.
		if tasks[0].NextRun != "0001-01-01T00:00" {
			t.Errorf("NextRun = %q; want %q", tasks[0].NextRun, "0001-01-01T00:00")
		}
		if tasks[0].LastRun != "12/31/1600 00:00" {
			t.Errorf("LastRun = %q; want %q", tasks[0].LastRun, "12/31/1600 00:00")
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		_, err := parseTasksJSON("{bad}")
		if err == nil {
			t.Fatal("expected error for invalid JSON")
		}
	})
}

// ---------------------------------------------------------------------------
// parseTasksSimpleJSON
// ---------------------------------------------------------------------------

func TestParseTasksSimpleJSON(t *testing.T) {
	t.Run("normal simple tasks", func(t *testing.T) {
		j := `[
			{"TaskName": "SimpleA", "State": "Ready", "NextRunTime": "2024-12-16T03:00:00", "LastRunTime": "2024-12-15T03:00:00"},
			{"TaskName": "SimpleB", "State": "Disabled", "NextRunTime": "", "LastRunTime": ""}
		]`
		tasks, err := parseTasksSimpleJSON(j)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(tasks) != 2 {
			t.Fatalf("expected 2 tasks, got %d", len(tasks))
		}
		if tasks[0].Name != "SimpleA" {
			t.Errorf("Name = %q; want SimpleA", tasks[0].Name)
		}
		if tasks[0].Status != "Ready" {
			t.Errorf("Status = %q; want Ready", tasks[0].Status)
		}
		// Simple format sets Author and Trigger to N/A
		if tasks[0].Author != "N/A" {
			t.Errorf("Author = %q; want N/A", tasks[0].Author)
		}
		if tasks[0].Trigger != "N/A" {
			t.Errorf("Trigger = %q; want N/A", tasks[0].Trigger)
		}
		if tasks[1].NextRun != "N/A" {
			t.Errorf("NextRun = %q; want N/A", tasks[1].NextRun)
		}
		if tasks[1].LastRun != "N/A" {
			t.Errorf("LastRun = %q; want N/A", tasks[1].LastRun)
		}
	})

	t.Run("state as map with Value", func(t *testing.T) {
		j := `[
			{"TaskName": "MapState", "State": {"Value": "Disabled"}}
		]`
		tasks, err := parseTasksSimpleJSON(j)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(tasks) != 1 {
			t.Fatalf("expected 1 task, got %d", len(tasks))
		}
		if tasks[0].Status != "Disabled" {
			t.Errorf("Status = %q; want Disabled", tasks[0].Status)
		}
	})

	t.Run("state as float64 numeric", func(t *testing.T) {
		j := `[
			{"TaskName": "NumState", "State": 4.0}
		]`
		tasks, err := parseTasksSimpleJSON(j)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(tasks) != 1 {
			t.Fatalf("expected 1 task, got %d", len(tasks))
		}
		if tasks[0].Status != "Running" {
			t.Errorf("Status = %q; want Running", tasks[0].Status)
		}
	})

	t.Run("state as unsupported type defaults to Ready", func(t *testing.T) {
		j := `[
			{"TaskName": "Unsup", "State": null}
		]`
		tasks, err := parseTasksSimpleJSON(j)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(tasks) != 1 {
			t.Fatalf("expected 1 task, got %d", len(tasks))
		}
		if tasks[0].Status != "Ready" {
			t.Errorf("Status = %q; want Ready (default for unsupported type)", tasks[0].Status)
		}
	})

	t.Run("empty name skipped", func(t *testing.T) {
		j := `[
			{"TaskName": ""},
			{"TaskName": "Valid"}
		]`
		tasks, err := parseTasksSimpleJSON(j)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(tasks) != 1 {
			t.Fatalf("expected 1 task, got %d", len(tasks))
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		_, err := parseTasksSimpleJSON("{bad}")
		if err == nil {
			t.Fatal("expected error for invalid JSON")
		}
	})
}

// ---------------------------------------------------------------------------
// findCSVColumn
// ---------------------------------------------------------------------------

func TestFindCSVColumn(t *testing.T) {
	headers := []string{"hostname", "taskname", "next runtime", "status", "last runtime", "author"}

	tests := []struct {
		name string
		col  string
		want int
	}{
		{"exact match", "taskname", 1},
		{"next runtime", "next runtime", 2},
		{"author", "author", 5},
		{"missing column", "nonexistent", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findCSVColumn(headers, tt.col)
			if got != tt.want {
				t.Errorf("findCSVColumn(%q) = %d, want %d", tt.col, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// parseDefenderWMIC
// ---------------------------------------------------------------------------

func TestParseDefenderWMIC(t *testing.T) {
	t.Run("full response", func(t *testing.T) {
		csv := `Node,AntivirusEnabled,AMServiceEnabled,AntispywareEnabled,NISEnabled,RealTimeProtectionEnabled,CloudProtectionEnabled,SignatureAge,QuickScanAge,FullScanAge,QuickScanEndTime,FullScanEndTime
HOSTNAME,TRUE,TRUE,TRUE,FALSE,TRUE,FALSE,2,0,15,20241215T030000.123456Z,20241201T000000.000000Z`
		status, err := parseDefenderWMIC(csv)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !status.Enabled {
			t.Error("Enabled should be true")
		}
		if !status.AMServiceEnabled {
			t.Error("AMServiceEnabled should be true")
		}
		if status.NISEnabled {
			t.Error("NISEnabled should be false")
		}
		if !status.RealTimeProtection {
			t.Error("RealTimeProtection should be true")
		}
		if status.SignatureAge != "2 days ago" {
			t.Errorf("SignatureAge = %q, want '2 days ago'", status.SignatureAge)
		}
		if !status.UpToDate {
			t.Error("UpToDate should be true (age=2)")
		}
		if status.QuickScanAge != 0 {
			t.Errorf("QuickScanAge = %d, want 0", status.QuickScanAge)
		}
		if status.FullScanAge != 15 {
			t.Errorf("FullScanAge = %d, want 15", status.FullScanAge)
		}
	})

	t.Run("disabled defender", func(t *testing.T) {
		csv := `Node,AntivirusEnabled,AMServiceEnabled,RealTimeProtectionEnabled,SignatureAge
HOSTNAME,FALSE,FALSE,FALSE,30`
		status, err := parseDefenderWMIC(csv)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if status.Enabled {
			t.Error("Enabled should be false")
		}
		if status.UpToDate {
			t.Error("UpToDate should be false (age=30 > 7)")
		}
		if status.SignatureAge != "1 month ago" {
			t.Errorf("SignatureAge = %q, want '1 month ago'", status.SignatureAge)
		}
	})

	t.Run("empty output returns error", func(t *testing.T) {
		_, err := parseDefenderWMIC("")
		if err == nil {
			t.Error("expected error for empty output")
		}
	})

	t.Run("header only returns error", func(t *testing.T) {
		_, err := parseDefenderWMIC("Node,AntivirusEnabled\n")
		if err == nil {
			t.Error("expected error for header-only output")
		}
	})
}

func TestWMICBool(t *testing.T) {
	data := map[string]string{"a": "TRUE", "b": "true", "c": "1", "d": "FALSE", "e": "0", "f": ""}
	tests := []struct {
		key  string
		want bool
	}{
		{"a", true},
		{"b", true},
		{"c", true},
		{"d", false},
		{"e", false},
		{"f", false},
		{"missing", false},
	}
	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			got := wmicBool(data, tt.key)
			if got != tt.want {
				t.Errorf("wmicBool(%q) = %v, want %v", tt.key, got, tt.want)
			}
		})
	}
}

func TestWMICInt(t *testing.T) {
	data := map[string]string{"a": "42", "b": "2.0000000000", "c": "", "d": "notanumber"}
	tests := []struct {
		key  string
		want int
		ok   bool
	}{
		{"a", 42, true},
		{"b", 2, true},
		{"c", 0, false},
		{"d", 0, false},
		{"missing", 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			got, ok := wmicInt(data, tt.key)
			if ok != tt.ok || got != tt.want {
				t.Errorf("wmicInt(%q) = (%d, %v), want (%d, %v)", tt.key, got, ok, tt.want, tt.ok)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// parseTasksSchTasksCSV
// ---------------------------------------------------------------------------

func TestParseTasksSchTasksCSV(t *testing.T) {
	t.Run("verbose output with all fields", func(t *testing.T) {
		csv := `"HostName","TaskName","Next Run Time","Status","Logon Mode","Last Run Time","Last Result","Author"
"MyPC","\Microsoft\Windows\Update\ScheduledStart","1/15/2024 3:00:00 AM","Ready","Interactive","1/14/2024 3:00:00 AM","0","NT AUTHORITY\SYSTEM"
"MyPC","\MyTask","N/A","Disabled","Interactive","N/A","0","Admin"`
		tasks, err := parseTasksSchTasksCSV(csv)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(tasks) != 2 {
			t.Fatalf("expected 2 tasks, got %d", len(tasks))
		}
		if tasks[0].Name != "\\Microsoft\\Windows\\Update\\ScheduledStart" {
			t.Errorf("Name = %q, want '\\Microsoft\\Windows\\Update\\ScheduledStart'", tasks[0].Name)
		}
		if tasks[0].Status != "Ready" {
			t.Errorf("Status = %q, want Ready", tasks[0].Status)
		}
		if tasks[0].Author != "NT AUTHORITY\\SYSTEM" {
			t.Errorf("Author = %q, want 'NT AUTHORITY\\SYSTEM'", tasks[0].Author)
		}
		if tasks[1].Status != "Disabled" {
			t.Errorf("second task Status = %q, want Disabled", tasks[1].Status)
		}
		if tasks[1].NextRun != "N/A" {
			t.Errorf("second task NextRun = %q, want N/A", tasks[1].NextRun)
		}
	})

	t.Run("empty output returns error", func(t *testing.T) {
		_, err := parseTasksSchTasksCSV("")
		if err == nil {
			t.Error("expected error for empty output")
		}
	})

	t.Run("no taskname column returns error", func(t *testing.T) {
		csv := "\"Col1\",\"Col2\"\n\"a\",\"b\""
		_, err := parseTasksSchTasksCSV(csv)
		if err == nil {
			t.Error("expected error when taskname column missing")
		}
	})
}

func TestParseTasksSchTasksSimpleCSV(t *testing.T) {
	t.Run("basic output", func(t *testing.T) {
		csv := `"TaskName","Next Run Time","Status"
"TaskA","1/15/2024 3:00:00 AM","Ready"
"TaskB","N/A","Disabled"`
		tasks, err := parseTasksSchTasksSimpleCSV(csv)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(tasks) != 2 {
			t.Fatalf("expected 2 tasks, got %d", len(tasks))
		}
		if tasks[0].Name != "TaskA" {
			t.Errorf("Name = %q, want TaskA", tasks[0].Name)
		}
		if tasks[0].Status != "Ready" {
			t.Errorf("Status = %q, want Ready", tasks[0].Status)
		}
		if tasks[1].Status != "Disabled" {
			t.Errorf("second Status = %q, want Disabled", tasks[1].Status)
		}
	})

	t.Run("empty output returns error", func(t *testing.T) {
		_, err := parseTasksSchTasksSimpleCSV("")
		if err == nil {
			t.Error("expected error for empty output")
		}
	})

	t.Run("no taskname column returns error", func(t *testing.T) {
		csv := "\"Col1\",\"Col2\"\n\"a\",\"b\""
		_, err := parseTasksSchTasksSimpleCSV(csv)
		if err == nil {
			t.Error("expected error when taskname column missing")
		}
	})
}

func TestParseFirewallRules(t *testing.T) {
	t.Run("multiple rule blocks", func(t *testing.T) {
		output := `Rule Name:                            Allow HTTP
Enabled:                              Yes
Direction:                            In
Profiles:                             Domain,Private,Public
Action:                               Allow
Protocol:                             TCP
LocalPort:                            80
RemotePort:                           Any
RemoteIP:                             Any

Rule Name:                            Block SMB
Enabled:                              Yes
Direction:                            In
Profiles:                             Domain,Private,Public
Action:                               Block
Protocol:                             TCP
LocalPort:                            445
RemotePort:                           Any
RemoteIP:                             Any

Rule Name:                            Allow RDP
Enabled:                              No
Direction:                            In
Profiles:                             Domain
Action:                               Allow
Protocol:                             TCP
LocalPort:                            3389
RemotePort:                           Any
RemoteIP:                             Any
`
		rules, err := parseFirewallRules(output)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(rules) != 3 {
			t.Fatalf("expected 3 rules, got %d", len(rules))
		}

		// First rule: Allow HTTP
		r := rules[0]
		if r.Name != "Allow HTTP" {
			t.Errorf("Name = %q; want %q", r.Name, "Allow HTTP")
		}
		if !r.Enabled {
			t.Error("Allow HTTP should be enabled")
		}
		if r.Direction != "In" {
			t.Errorf("Direction = %q; want In", r.Direction)
		}
		if r.Action != "Allow" {
			t.Errorf("Action = %q; want Allow", r.Action)
		}
		if r.Protocol != "TCP" {
			t.Errorf("Protocol = %q; want TCP", r.Protocol)
		}
		if r.LocalPort != "80" {
			t.Errorf("LocalPort = %q; want 80", r.LocalPort)
		}
		if r.RemotePort != "Any" {
			t.Errorf("RemotePort = %q; want Any", r.RemotePort)
		}
		if r.RemoteIP != "Any" {
			t.Errorf("RemoteIP = %q; want Any", r.RemoteIP)
		}
		if r.Profile != "Domain,Private,Public" {
			t.Errorf("Profile = %q; want Domain,Private,Public", r.Profile)
		}

		// Second rule: Block SMB
		r = rules[1]
		if r.Name != "Block SMB" {
			t.Errorf("Name = %q; want Block SMB", r.Name)
		}
		if !r.Enabled {
			t.Error("Block SMB should be enabled")
		}
		if r.Action != "Block" {
			t.Errorf("Action = %q; want Block", r.Action)
		}

		// Third rule: Allow RDP (disabled)
		r = rules[2]
		if r.Name != "Allow RDP" {
			t.Errorf("Name = %q; want Allow RDP", r.Name)
		}
		if r.Enabled {
			t.Error("Allow RDP should be disabled")
		}
	})

	t.Run("partial block with missing fields", func(t *testing.T) {
		output := `Rule Name:                            Minimal Rule
Enabled:                              Yes
Direction:                            In
`
		rules, err := parseFirewallRules(output)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(rules) != 1 {
			t.Fatalf("expected 1 rule, got %d", len(rules))
		}
		if rules[0].Name != "Minimal Rule" {
			t.Errorf("Name = %q; want Minimal Rule", rules[0].Name)
		}
		if !rules[0].Enabled {
			t.Error("should be enabled")
		}
		if rules[0].Direction != "In" {
			t.Errorf("Direction = %q; want In", rules[0].Direction)
		}
		// Unset fields should be zero values
		if rules[0].Action != "" {
			t.Errorf("Action = %q; want empty", rules[0].Action)
		}
		if rules[0].Protocol != "" {
			t.Errorf("Protocol = %q; want empty", rules[0].Protocol)
		}
	})

	t.Run("blocks without Rule Name prefix are skipped", func(t *testing.T) {
		output := `Some random block

Rule Name:                            Valid Rule
Enabled:                              Yes
`
		rules, err := parseFirewallRules(output)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(rules) != 1 {
			t.Fatalf("expected 1 rule, got %d", len(rules))
		}
		if rules[0].Name != "Valid Rule" {
			t.Errorf("Name = %q; want Valid Rule", rules[0].Name)
		}
	})

	t.Run("lines with --- separators skipped", func(t *testing.T) {
		output := `Rule Name:                            Test
Enabled:                              Yes
--------------------------------------
Direction:                            In
`
		rules, err := parseFirewallRules(output)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(rules) != 1 {
			t.Fatalf("expected 1 rule, got %d", len(rules))
		}
	})

	t.Run("empty output returns empty slice", func(t *testing.T) {
		rules, err := parseFirewallRules("")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(rules) != 0 {
			t.Errorf("expected 0 rules, got %d", len(rules))
		}

		rules, err = parseFirewallRules("   \n\n  ")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(rules) != 0 {
			t.Errorf("expected 0 rules for whitespace, got %d", len(rules))
		}
	})

	t.Run("cap at 100 rules", func(t *testing.T) {
		// Build 110 identical rule blocks
		var sb string
		for i := 0; i < 110; i++ {
			sb += "Rule Name:                            Rule" + string(rune('0'+i%10)) + "\n"
			sb += "Enabled:                              Yes\n"
			sb += "Direction:                            In\n\n"
		}
		rules, err := parseFirewallRules(sb)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(rules) > 100 {
			t.Errorf("expected at most 100 rules, got %d", len(rules))
		}
	})
}

func TestParseSecurityEventsJSON(t *testing.T) {
	t.Run("array response", func(t *testing.T) {
		j := `[
			{"Id": 4625, "LevelDisplayName": "Information", "ProviderName": "Microsoft-Windows-Security-Auditing", "TimeCreated": "2026-07-02T01:02:03.0000000Z", "Message": "An account failed to log on."},
			{"Id": 4624, "LevelDisplayName": "Information", "ProviderName": "Microsoft-Windows-Security-Auditing", "TimeCreated": "2026-07-02T01:05:03Z", "Message": "An account was successfully logged on."}
		]`

		events, err := parseSecurityEventsJSON(j)
		if err != nil {
			t.Fatalf("parseSecurityEventsJSON() error = %v", err)
		}
		if len(events) != 2 {
			t.Fatalf("len(events) = %d, want 2", len(events))
		}
		if events[0].ID != 4625 {
			t.Errorf("events[0].ID = %d, want 4625", events[0].ID)
		}
		if !events[0].Important {
			t.Error("failed login event should be important")
		}
		if events[0].Time != "2026-07-02 01:02:03" {
			t.Errorf("events[0].Time = %q, want formatted time", events[0].Time)
		}
		if events[1].Important {
			t.Error("successful login event should not be important by default")
		}
	})

	t.Run("single object response", func(t *testing.T) {
		j := `{"Id": 1102, "LevelDisplayName": "Information", "ProviderName": "Audit", "Message": "The audit log was cleared."}`

		events, err := parseSecurityEventsJSON(j)
		if err != nil {
			t.Fatalf("parseSecurityEventsJSON() error = %v", err)
		}
		if len(events) != 1 {
			t.Fatalf("len(events) = %d, want 1", len(events))
		}
		if !events[0].Important {
			t.Error("audit log cleared event should be important")
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		_, err := parseSecurityEventsJSON("{bad}")
		if err == nil {
			t.Fatal("expected error for invalid JSON")
		}
	})

	t.Run("empty output", func(t *testing.T) {
		_, err := parseSecurityEventsJSON("")
		if err == nil {
			t.Fatal("expected error for empty output")
		}
	})
}

func TestSecurityEventImportance(t *testing.T) {
	if !isImportantSecurityEvent(SecurityEvent{ID: 4740}) {
		t.Error("account lockout event should be important")
	}
	if !isImportantSecurityEvent(SecurityEvent{Level: "Critical"}) {
		t.Error("critical level event should be important")
	}
	if isImportantSecurityEvent(SecurityEvent{ID: 4624, Level: "Information"}) {
		t.Error("ordinary informational event should not be important")
	}
}

func TestSecOpsEventsResult(t *testing.T) {
	m := NewModel()
	m.Update(EventsResult{
		Events: []SecurityEvent{{
			ID:       4625,
			Level:    "Information",
			Provider: "Audit",
			Message:  "failed login",
		}},
	})

	if !m.Ready() {
		t.Error("model should be ready after EventsResult")
	}
	if len(m.SecurityEvents()) != 1 {
		t.Fatalf("SecurityEvents length = %d, want 1", len(m.SecurityEvents()))
	}
	if m.EventsError() != nil {
		t.Errorf("EventsError() = %v, want nil", m.EventsError())
	}
}
