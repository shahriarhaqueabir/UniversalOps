package common

import (
	"testing"
)

func TestCapabilityRegistry_LHM(t *testing.T) {
	r := NewCapabilityRegistry()

	// Since LHM depends on a running service, we can't easily assert 'true' in CI
	// but we can assert the ID exists in the list.
	found := false
	for _, info := range r.List() {
		if info.ID == CapLHM {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("CapLHM not found in registry list")
	}
}

func TestCapabilityRegistry_RefreshBatch(t *testing.T) {
	r := NewCapabilityRegistry()

	// Refresh a specific tool
	r.RefreshBatch([]CapabilityID{CapOllama})

	info, ok := r.tools[CapOllama]
	if !ok {
		t.Fatalf("CapOllama not found in registry")
	}

	// We can't guarantee availability, but we can verify the ID matches
	if info.ID != CapOllama {
		t.Errorf("Expected ID %s, got %s", CapOllama, info.ID)
	}
}
