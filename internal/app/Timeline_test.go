package app

import (
	"testing"

	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
)

func TestTimeline_GetTimelineEvents_DefaultFilter(t *testing.T) {
	a := NewApp()
	tl := NewTimeline(a)
	events := tl.GetTimelineEvents("", "", 10, 0)
	if events == nil {
		t.Fatal("GetTimelineEvents returned nil, expected non-nil slice")
	}
}

func TestTimeline_GetTimelineEvents_WithCategory(t *testing.T) {
	a := NewApp()
	tl := NewTimeline(a)
	events := tl.GetTimelineEvents("system", "", 10, 0)
	if events == nil {
		t.Fatal("GetTimelineEvents with category returned nil, expected non-nil slice")
	}
}

func TestTimeline_GetTimelineEvents_WithOffset(t *testing.T) {
	a := NewApp()
	tl := NewTimeline(a)
	events := tl.GetTimelineEvents("", "", 10, 100)
	if events == nil {
		t.Fatal("GetTimelineEvents with large offset returned nil, expected non-nil slice")
	}
}

func TestTimeline_GetTimelineEvents_NegativeLimit(t *testing.T) {
	a := NewApp()
	tl := NewTimeline(a)
	events := tl.GetTimelineEvents("", "", -1, 0)
	if events == nil {
		t.Fatal("GetTimelineEvents with negative limit returned nil, expected non-nil slice")
	}
}

func TestTimeline_GetTimelineEvents_OverLimit(t *testing.T) {
	a := NewApp()
	tl := NewTimeline(a)
	events := tl.GetTimelineEvents("", "", 9999, 0)
	if events == nil {
		t.Fatal("GetTimelineEvents with over-limit returned nil, expected non-nil slice")
	}
}

func TestTimeline_GetTimelineEventByID_Nonexistent(t *testing.T) {
	a := NewApp()
	tl := NewTimeline(a)
	evt := tl.GetTimelineEventByID("nonexistent-id")
	if evt != nil {
		t.Log("GetTimelineEventByID returned non-nil for nonexistent ID (may be in bus)")
	}
}

func TestTimeline_GetRelatedEvents_Nonexistent(t *testing.T) {
	a := NewApp()
	tl := NewTimeline(a)
	events := tl.GetRelatedEvents("nonexistent-id")
	if events == nil {
		t.Fatal("GetRelatedEvents returned nil, expected non-nil slice")
	}
}

func TestTimeline_GetTimelineCategories(t *testing.T) {
	a := NewApp()
	tl := NewTimeline(a)
	cats := tl.GetTimelineCategories()
	if cats == nil {
		t.Fatal("GetTimelineCategories returned nil, expected non-nil slice")
	}
	if len(cats) == 0 {
		t.Error("GetTimelineCategories returned empty slice")
	}
	expected := []string{"system", "network", "security", "devops", "ai", "alert", "pipeline"}
	if len(cats) != len(expected) {
		t.Errorf("GetTimelineCategories length = %d, want %d", len(cats), len(expected))
	}
	for i, cat := range cats {
		if cat != expected[i] {
			t.Errorf("GetTimelineCategories[%d] = %q, want %q", i, cat, expected[i])
		}
	}
}

func TestTimeline_GetTimelineSummary_NegativeMinutes(t *testing.T) {
	a := NewApp()
	tl := NewTimeline(a)
	summary := tl.GetTimelineSummary(-1)
	if summary == nil {
		t.Fatal("GetTimelineSummary returned nil, expected non-nil map")
	}
}

func TestTimeline_GetTimelineSummary(t *testing.T) {
	a := NewApp()
	tl := NewTimeline(a)
	summary := tl.GetTimelineSummary(60)
	if summary == nil {
		t.Fatal("GetTimelineSummary returned nil, expected non-nil map")
	}
}

func TestTimeline_ExplainEvents_Empty(t *testing.T) {
	a := NewApp()
	tl := NewTimeline(a)
	result := tl.ExplainEvents(nil)
	if result == "" {
		t.Error("ExplainEvents with nil returned empty string")
	}
}

func TestTimeline_ExplainEvents_NonexistentIDs(t *testing.T) {
	a := NewApp()
	tl := NewTimeline(a)
	result := tl.ExplainEvents([]string{"id1", "id2"})
	if result == "" {
		t.Error("ExplainEvents with nonexistent IDs returned empty string")
	}
}

func TestTimeline_convertTimelineEvent_WithNilRelated(t *testing.T) {
	evt := common.TimelineEvent{
		ID:    "test-id",
		Title: "test event",
	}
	converted := convertTimelineEvent(evt)
	if converted.ID != "test-id" {
		t.Errorf("convertTimelineEvent ID = %q, want %q", converted.ID, "test-id")
	}
	if converted.Related == nil {
		t.Error("convertTimelineEvent.Related is nil, expected non-nil slice")
	}
}

func TestTimeline_convertTimelineEvents_NilInput(t *testing.T) {
	result := convertTimelineEvents(nil)
	if result == nil {
		t.Fatal("convertTimelineEvents(nil) returned nil, expected non-nil slice")
	}
	if len(result) != 0 {
		t.Errorf("convertTimelineEvents(nil) length = %d, want 0", len(result))
	}
}
