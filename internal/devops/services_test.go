package devops

import (
	"runtime"
	"testing"
)

func TestListServices(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Services test only for Windows")
	}

	services, err := ListServices(0)
	if err != nil {
		t.Fatalf("ListServices failed: %v", err)
	}

	// Windows always has services
	if len(services) == 0 {
		t.Error("No services found")
	}

	foundSpooler := false
	for _, s := range services {
		if s.Name == "Spooler" {
			foundSpooler = true
			break
		}
	}

	if !foundSpooler {
		t.Log("Spooler service not found (unusual for Windows)")
	}
}
