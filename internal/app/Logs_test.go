package app

import (
	"testing"
	"time"

	"github.com/shahriarhaqueabir/UniversalOps/internal/common"
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

func TestGetLogTimeline(t *testing.T) {
	common.InitStorage(":memory:")
	defer common.GetStorage().Close()

	store := common.GetStorage()
	for i := 0; i < 10; i++ {
		store.InsertLog("INFO", "test", "test info")
		store.InsertLog("ERROR", "test", "test error")
		store.InsertLog("WARN", "test", "test warn")
	}

	// Wait for async writer's 1-second ticker to flush pending logs
	time.Sleep(1500 * time.Millisecond)

	l := NewLogs()
	points := l.GetLogTimeline(24)
	if len(points) == 0 {
		t.Fatal("GetLogTimeline(24) returned empty — expected at least 1 bucket")
	}

	// Verify each point has expected fields
	for _, p := range points {
		if p.Timestamp == "" {
			t.Error("found point with empty timestamp")
		}
	}

	// Verify ordering (ascending)
	for i := 1; i < len(points); i++ {
		if points[i].Timestamp < points[i-1].Timestamp {
			t.Errorf("points not sorted ascending: %q before %q", points[i-1].Timestamp, points[i].Timestamp)
		}
	}
}

func TestGetLogTimeline_Empty(t *testing.T) {
	common.InitStorage(":memory:")
	defer common.GetStorage().Close()

	l := NewLogs()
	points := l.GetLogTimeline(24)
	if points == nil {
		t.Error("GetLogTimeline(24) on empty storage returned nil, expected empty slice")
	}
	if len(points) != 0 {
		t.Errorf("GetLogTimeline(24) returned %d points on empty storage, expected 0", len(points))
	}

	// Also verify that the defaulting (hours <= 0 --> 24) doesn't crash
	if p := l.GetLogTimeline(0); p == nil {
		t.Error("GetLogTimeline(0) returned nil")
	}
	if p := l.GetLogTimeline(-1); p == nil {
		t.Error("GetLogTimeline(-1) returned nil")
	}
}
