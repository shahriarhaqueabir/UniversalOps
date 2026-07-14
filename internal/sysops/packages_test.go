package sysops

import "testing"

func TestGetInstalledPackages(t *testing.T) {
	managers := GetInstalledPackages()
	if len(managers) == 0 {
		t.Error("Expected at least one package manager")
	}
	for _, m := range managers {
		t.Logf("Package manager %s: found=%v, packages=%d", m.Name, m.Found, len(m.Packages))
	}
}
