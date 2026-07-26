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

func TestBitLockerProtection(t *testing.T) {
	on := 1
	off := 0
	var nilPtr *int
	if !bitLockerProtection(&on) {
		t.Error("bitLockerProtection(1) should return true")
	}
	if bitLockerProtection(&off) {
		t.Error("bitLockerProtection(0) should return false")
	}
	if bitLockerProtection(nilPtr) {
		t.Error("bitLockerProtection(nil) should return false")
	}
}

func TestBitLockerMethod(t *testing.T) {
	if bitLockerMethod("AES 128") != "AES 128" {
		t.Errorf("bitLockerMethod(AES 128) = %q, want AES 128", bitLockerMethod("AES 128"))
	}
	if bitLockerMethod("AES 256") != "AES 256" {
		t.Errorf("bitLockerMethod(AES 256) = %q, want AES 256", bitLockerMethod("AES 256"))
	}
	if bitLockerMethod("XTS-AES 128") != "XTS-AES 128" {
		t.Errorf("bitLockerMethod(XTS-AES 128) = %q, want XTS-AES 128", bitLockerMethod("XTS-AES 128"))
	}
	if bitLockerMethod("") != "None" {
		t.Errorf("bitLockerMethod(empty) = %q, want None", bitLockerMethod(""))
	}
	if bitLockerMethod("None") != "None" {
		t.Errorf("bitLockerMethod(None) = %q, want None", bitLockerMethod("None"))
	}
}

func TestStringOrEmpty(t *testing.T) {
	if stringOrEmpty("hello") != "hello" {
		t.Error("stringOrEmpty(string) should return the string")
	}
	if stringOrEmpty(42) != "42" {
		t.Errorf("stringOrEmpty(int) = %q, want 42", stringOrEmpty(42))
	}
	if stringOrEmpty(nil) != "" {
		t.Errorf("stringOrEmpty(nil) = %q, want empty", stringOrEmpty(nil))
	}
}

func TestServiceStatusString(t *testing.T) {
	if serviceStatusString("Running") != "Running" {
		t.Errorf("serviceStatusString(string) = %q", serviceStatusString("Running"))
	}
	if serviceStatusString(float64(4)) != "Running" {
		t.Errorf("serviceStatusString(float64(4)) = %q, want Running", serviceStatusString(float64(4)))
	}
	if serviceStatusString(float64(1)) != "Stopped" {
		t.Errorf("serviceStatusString(float64(1)) = %q, want Stopped", serviceStatusString(float64(1)))
	}
	if serviceStatusString(nil) != "" {
		t.Errorf("serviceStatusString(nil) = %q, want empty", serviceStatusString(nil))
	}
}

func TestServiceStartTypeString(t *testing.T) {
	if serviceStartTypeString("Auto") != "Auto" {
		t.Errorf("serviceStartTypeString(string) = %q", serviceStartTypeString("Auto"))
	}
	if serviceStartTypeString(float64(2)) != "Automatic" {
		t.Errorf("serviceStartTypeString(float64(2)) = %q, want Automatic", serviceStartTypeString(float64(2)))
	}
	if serviceStartTypeString(float64(3)) != "Manual" {
		t.Errorf("serviceStartTypeString(float64(3)) = %q, want Manual", serviceStartTypeString(float64(3)))
	}
	if serviceStartTypeString(float64(4)) != "Disabled" {
		t.Errorf("serviceStartTypeString(float64(4)) = %q, want Disabled", serviceStartTypeString(float64(4)))
	}
	if serviceStartTypeString(nil) != "" {
		t.Errorf("serviceStartTypeString(nil) = %q, want empty", serviceStartTypeString(nil))
	}
}

func TestParseServicesSCQuery(t *testing.T) {
	output := `SERVICE_NAME: wuauserv
DISPLAY_NAME: Windows Update
TYPE               : 20  WIN32_SHARE_PROCESS
STATE              : 4  RUNNING
START_TYPE         : 2   AUTO_START
`
	svcs := parseServicesSCQuery(output)
	if len(svcs) == 0 {
		t.Fatal("parseServicesSCQuery returned no services")
	}
	if svcs[0].Name != "wuauserv" {
		t.Errorf("service.Name = %q, want wuauserv", svcs[0].Name)
	}
}
