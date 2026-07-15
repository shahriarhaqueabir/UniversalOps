package secops

import (
	"testing"
)

func TestActionResultStruct(t *testing.T) {
	result := ActionResult{
		Success: true,
		Message: "Action completed",
		Error:   "",
	}
	if !result.Success {
		t.Error("expected Success to be true")
	}
	if result.Error != "" {
		t.Errorf("expected empty Error, got %s", result.Error)
	}
}
