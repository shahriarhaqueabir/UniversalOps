package netops

import "testing"

func TestGetFirewallRules(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping firewall test in short mode")
	}
	rules, err := GetFirewallRules()
	if err != nil {
		t.Logf("GetFirewallRules error (may need admin): %v", err)
		return
	}
	t.Logf("Found %d firewall rules", len(rules))
	for i, r := range rules {
		if i >= 5 {
			break
		}
		t.Logf("  %s: %s %s %s (enabled=%v)", r.Name, r.Action, r.Direction, r.Protocol, r.Enabled)
	}
}
