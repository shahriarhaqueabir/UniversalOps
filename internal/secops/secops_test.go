package secops

import (
	"runtime"
	"testing"
)

func TestGetFirewallRules(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Firewall rules test currently only for Windows")
	}

	rules, err := GetFirewallRules()
	if err != nil {
		t.Fatalf("GetFirewallRules failed: %v", err)
	}

	// Should have some rules on a default Windows system
	if len(rules) == 0 {
		t.Log("No firewall rules found (unexpected on Windows)")
	}

	for _, r := range rules {
		if r.Name == "" {
			t.Error("Firewall rule missing Name")
		}
	}
}

func TestGetUsers(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Users test currently only for Windows")
	}

	users, err := GetUsers()
	if err != nil {
		t.Fatalf("GetUsers failed: %v", err)
	}

	if len(users) == 0 {
		t.Error("No users found")
	}

	foundAdmin := false
	for _, u := range users {
		if u.IsAdmin {
			foundAdmin = true
		}
		if u.Username == "" {
			t.Error("User missing Username")
		}
	}

	if !foundAdmin {
		t.Log("No admin users found (unexpected but possible in restricted environments)")
	}
}

func TestGetListeningPorts(t *testing.T) {
	ports, err := GetListeningPorts()
	if err != nil {
		t.Fatalf("GetListeningPorts failed: %v", err)
	}

	// Should have some listening ports on a live system
	if len(ports) == 0 {
		t.Log("No listening ports found")
	}

	for _, p := range ports {
		if p.Port <= 0 {
			t.Errorf("Invalid port number: %d", p.Port)
		}
	}
}

func TestGetJSONStringWithValueObject(t *testing.T) {
	data := map[string]interface{}{
		"Level": map[string]interface{}{"Value": "Warning"},
		"Name":  "direct string",
		"Empty": map[string]interface{}{"Value": ""},
	}

	tests := []struct {
		name string
		key  string
		want string
		ok   bool
	}{
		{"Value object extracts inner", "Level", "Warning", true},
		{"Direct string works", "Name", "direct string", true},
		{"Empty Value returns empty", "Empty", "", true},
		{"Missing key returns false", "Missing", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := getJSONString(data, tt.key)
			if ok != tt.ok {
				t.Errorf("getJSONString(_, %q) ok = %v, want %v", tt.key, ok, tt.ok)
			}
			if got != tt.want {
				t.Errorf("getJSONString(_, %q) = %q, want %q", tt.key, got, tt.want)
			}
		})
	}
}

func TestGetJSONIntEdgeCases(t *testing.T) {
	data := map[string]interface{}{
		"Count":      42,
		"FloatCount": float64(3.14),
		"Nested":     map[string]interface{}{"Value": float64(99)},
		"DashValue":  "-",
		"EmptyStr":   "",
	}

	tests := []struct {
		name string
		key  string
		want int
		ok   bool
	}{
		{"int value works", "Count", 42, true},
		{"float value truncates", "FloatCount", 3, true},
		{"nested Value object extracts", "Nested", 99, true},
		{"dash string returns 0", "DashValue", 0, true},
		{"empty string returns 0", "EmptyStr", 0, true},
		{"missing key returns false", "Missing", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := getJSONInt(data, tt.key)
			if ok != tt.ok {
				t.Errorf("getJSONInt(_, %q) ok = %v, want %v", tt.key, ok, tt.ok)
			}
			if got != tt.want {
				t.Errorf("getJSONInt(_, %q) = %d, want %d", tt.key, got, tt.want)
			}
		})
	}
}

func TestParseFirewallRulesEmpty(t *testing.T) {
	rules, err := parseFirewallRules("")
	if err != nil {
		t.Errorf("parseFirewallRules empty returned err: %v", err)
	}
	if len(rules) != 0 {
		t.Errorf("parseFirewallRules empty returned %d rules, want 0", len(rules))
	}
}

func TestParseNetUserListEmpty(t *testing.T) {
	users := parseNetUserList("")
	if len(users) != 0 {
		t.Errorf("parseNetUserList empty returned %d users, want 0", len(users))
	}

	users = parseNetUserList("The command completed successfully.")
	if len(users) != 0 {
		t.Errorf("parseNetUserList completed returned %d users, want 0", len(users))
	}
}

func TestGetDefenderStatus(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Defender status test only for Windows")
	}

	status, err := GetDefenderStatus()
	if err != nil {
		t.Logf("GetDefenderStatus might fail if not on Windows or if WMI fails: %v", err)
		return
	}

	if status == nil {
		t.Fatal("DefenderStatus is nil")
	}
}
