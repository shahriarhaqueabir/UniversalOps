package app

import (
	"testing"

	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
)

func TestGetLogs(t *testing.T) {
	// Initialize in-memory storage for test
	common.InitStorage(":memory:")
	defer common.GetStorage().Close()

	l := NewLogs()
	logs := l.GetLogs("", "", 10)

	// Should return empty slice if no logs yet, not nil
	if logs == nil {
		t.Error("GetLogs returned nil, expected empty slice")
	}
}
