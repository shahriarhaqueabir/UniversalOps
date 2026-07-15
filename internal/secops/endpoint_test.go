package secops

import (
	"testing"
)

func TestDiskEncryptionStruct(t *testing.T) {
	de := DiskEncryption{
		Volume:    "C:",
		Encrypted: true,
		Method:    "AES-256",
		Status:    "Protection On",
	}
	if !de.Encrypted {
		t.Error("expected Encrypted to be true")
	}
	if de.Method != "AES-256" {
		t.Errorf("expected Method AES-256, got %s", de.Method)
	}
}

func TestSecureBootStruct(t *testing.T) {
	sb := SecureBoot{Enabled: true, State: "Secure Boot is enabled"}
	if !sb.Enabled {
		t.Error("expected Enabled to be true")
	}
}

func TestSystemServiceStruct(t *testing.T) {
	ss := SystemService{
		Name:        "wuauserv",
		DisplayName: "Windows Update",
		Status:      "Running",
		StartupType: "Automatic",
	}
	if ss.Status != "Running" {
		t.Errorf("expected Status Running, got %s", ss.Status)
	}
}
