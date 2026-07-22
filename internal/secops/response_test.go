package secops

import (
	"testing"
	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
)

func TestActionResultStruct(t *testing.T) {
	result := common.SecActionResult{
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
