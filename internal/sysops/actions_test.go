package sysops

import "testing"

func TestRunSystemAction_Unknown(t *testing.T) {
	_, err := RunSystemAction("nonexistent_action")
	if err == nil {
		t.Error("Expected error for unknown action")
	}
}
