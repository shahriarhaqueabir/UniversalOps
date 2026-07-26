package secops

import (
	"strings"
	"testing"
)

func TestBoolStr(t *testing.T) {
	if boolStr(true) != "enabled" {
		t.Error("boolStr(true) should return 'enabled'")
	}
	if boolStr(false) != "disabled" {
		t.Error("boolStr(false) should return 'disabled'")
	}
}

func TestSecurityReportString_Empty(t *testing.T) {
	r := &SecurityReport{}
	output := r.String()
	if !strings.Contains(output, "Security Audit Report") {
		t.Error("String() should include header")
	}
}

func TestSecurityReportString_WithFirewallDisabled(t *testing.T) {
	r := &SecurityReport{
		FirewallRules: []FirewallRule{
			{Name: "Rule1", Enabled: false},
			{Name: "Rule2", Enabled: true},
			{Name: "Rule3", Enabled: false},
		},
		Users: []UserInfo{
			{Username: "admin", IsAdmin: true, IsEnabled: true},
			{Username: "user", IsAdmin: false, IsEnabled: true},
		},
		Defender: &DefenderStatus{
			Enabled:            true,
			RealTimeProtection: true,
			SignatureAge:       "1 day old",
		},
		ListeningPorts: []ListeningPort{
			{Port: 80, Protocol: "TCP"},
		},
		ScheduledTasks: []ScheduledTask{
			{Name: "Task1"},
		},
		SecurityEvents: []SecurityEvent{
			{ID: 4625, Important: true},
		},
	}
	output := r.String()
	checks := []string{
		"Security Audit Report",
		"3 rules",
		"2 accounts",
		"enabled",
		"real-time=enabled",
		"signatures=1 day old",
		"LISTENING PORTS: 1",
		"SCHEDULED TASKS: 1",
		"SECURITY EVENTS: 1",
		"WARNING: 2 firewall rules are disabled",
		"1 users have administrator privileges",
		"1 important security events found",
	}
	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Errorf("String() missing: %s", check)
		}
	}
}

func TestSecurityReportString_DefenderNil(t *testing.T) {
	r := &SecurityReport{
		FirewallRules: []FirewallRule{{Name: "Rule1", Enabled: true}},
		Users:         []UserInfo{{Username: "user", IsAdmin: false}},
	}
	output := r.String()
	if strings.Contains(output, "DEFENDER") {
		t.Error("String() should not contain DEFENDER section when Defender is nil")
	}
	if !strings.Contains(output, "1 rules") {
		t.Error("String() should include rule count")
	}
}

func TestSecurityReportMarkdown_Empty(t *testing.T) {
	r := &SecurityReport{}
	output := r.Markdown()
	if !strings.Contains(output, "Security Audit Report") {
		t.Error("Markdown() should include header")
	}
	if !strings.Contains(output, "| Firewall Rules | 0 |") {
		t.Error("Markdown() should show 0 firewall rules")
	}
}

func TestSecurityReportMarkdown_WithData(t *testing.T) {
	r := &SecurityReport{
		FirewallRules: []FirewallRule{
			{Name: "Allow HTTP", Direction: "In", Action: "Allow", Protocol: "TCP", LocalPort: "80", Enabled: true},
		},
		Users: []UserInfo{
			{Username: "admin", IsAdmin: true, IsEnabled: true, Group: "Administrators"},
			{Username: "guest", IsAdmin: false, IsEnabled: false, Group: "Guests"},
		},
		ListeningPorts: []ListeningPort{
			{Protocol: "TCP", Port: 443, ProcessName: "nginx", PID: 1234},
		},
		Defender: &DefenderStatus{
			Enabled: true,
		},
		ScheduledTasks: []ScheduledTask{
			{Name: "Cleanup"},
		},
		SecurityEvents: []SecurityEvent{
			{ID: 4625, Time: "2024-01-01T00:00:00Z"},
		},
	}
	output := r.Markdown()
	checks := []string{
		"Security Audit Report",
		"| Firewall Rules | 1 |",
		"| User Accounts | 2 |",
		"| Listening Ports | 1 |",
		"| Defender | enabled |",
		"Allow HTTP",
		"admin",
		"guest",
		"| TCP | 443 | nginx | 1234 |",
	}
	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Errorf("Markdown() missing: %s", check)
		}
	}
}

func TestSecurityReportMarkdown_NoDefender(t *testing.T) {
	r := &SecurityReport{
		FirewallRules: []FirewallRule{{Name: "SSH", Direction: "In", Action: "Allow", Protocol: "TCP", LocalPort: "22", Enabled: true}},
	}
	output := r.Markdown()
	// Defender section should not be present
	if strings.Contains(output, "| Defender |") {
		t.Error("Markdown() should not include Defender row when Defender is nil")
	}
}
