package secops

import (
	"testing"
)

func TestHardeningCheckStruct(t *testing.T) {
	hc := HardeningCheck{
		Category:    "Firewall",
		Check:       "Firewall enabled",
		Passed:      true,
		Severity:    "high",
		Remediation: "Enable firewall",
	}
	if !hc.Passed {
		t.Error("expected Passed to be true")
	}
	if hc.Severity != "high" {
		t.Errorf("expected severity high, got %s", hc.Severity)
	}
}

func TestSSHConfigStruct(t *testing.T) {
	sc := SSHConfig{
		PermitRootLogin:        "no",
		PasswordAuthentication: "yes",
		PubkeyAuthentication:   "yes",
		X11Forwarding:          "no",
		MaxAuthTries:           "3",
	}
	if sc.PermitRootLogin != "no" {
		t.Errorf("expected PermitRootLogin no, got %s", sc.PermitRootLogin)
	}
}
