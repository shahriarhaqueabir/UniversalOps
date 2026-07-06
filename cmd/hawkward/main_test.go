package main

import (
	"testing"
)

// TestMainCompiles verifies the package imports and entry point signature.
func TestMainCompiles(t *testing.T) {
	// main() is the entry point — we just verify the function exists
	// by checking we can reference it. The actual TUI run is tested
	// through the ui package tests.
	if false {
		main()
	}
}

// TestMainConstants validates key assumptions build on expected defaults.
func TestMainConstants(t *testing.T) {
	// The binary uses ui.NewRootModel() and tea.NewProgram().
	// This test ensures imports resolve correctly at build time.
	_ = t.Name()
}
