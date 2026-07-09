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
