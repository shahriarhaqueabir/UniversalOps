package secops

import (
	"testing"
)

func TestFormatEventTime(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"2024-01-15T10:30:00Z", "2024-01-15 10:30:00"},
		{"2024-01-15T10:30:00.000Z", "2024-01-15 10:30:00"},
		{"2024-01-15T10:30:00.123456Z", "2024-01-15 10:30:00"},
		{"invalid", "invalid"},
		{"", ""},
	}
	for _, tt := range tests {
		got := formatEventTime(tt.input)
		if got != tt.want {
			t.Errorf("formatEventTime(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestIsImportantSecurityEvent(t *testing.T) {
	importantIDs := []int{4625, 4720, 4726, 4732, 4733, 4740, 1102}
	unimportantIDs := []int{4624, 4634, 4647, 4672, 4722, 4723, 4724, 4725, 4728, 4756, 4768, 4769, 4776, 4778, 4779, 4798, 4799, 4800, 4801, 4802, 4803}

	for _, id := range importantIDs {
		evt := SecurityEvent{ID: id}
		if !isImportantSecurityEvent(evt) {
			t.Errorf("isImportantSecurityEvent(%d) should be true", id)
		}
	}
	for _, id := range unimportantIDs {
		evt := SecurityEvent{ID: id}
		if isImportantSecurityEvent(evt) {
			t.Errorf("isImportantSecurityEvent(%d) should be false", id)
		}
	}
}

func TestParseSecurityEventsJSON(t *testing.T) {
	jsonStr := `[
		{"Id": 4625, "TimeCreated": "2024-01-15T10:30:00Z", "LevelDisplayName": "Failure", "Message": "Failed login attempt"}
	]`
	events, err := parseSecurityEventsJSON(jsonStr)
	if err != nil {
		t.Fatalf("parseSecurityEventsJSON failed: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("parseSecurityEventsJSON returned %d events, want 1", len(events))
	}
	if events[0].ID != 4625 {
		t.Errorf("event ID = %d, want 4625", events[0].ID)
	}
	if events[0].Level != "Failure" {
		t.Errorf("event Level = %q, want Failure", events[0].Level)
	}
}

func TestParseSecurityEventsJSON_Empty(t *testing.T) {
	events, err := parseSecurityEventsJSON("")
	if err == nil {
		t.Error("parseSecurityEventsJSON(empty) should return error")
	}
	if events != nil {
		t.Error("parseSecurityEventsJSON(empty) should return nil events")
	}
}

func TestParseSecurityEventsJSON_Invalid(t *testing.T) {
	events, err := parseSecurityEventsJSON("{invalid}")
	if err == nil {
		t.Error("parseSecurityEventsJSON(invalid) should return error")
	}
	if events != nil {
		t.Error("parseSecurityEventsJSON(invalid) should return nil events")
	}
}

func TestParsePrivilegeEventsJSON(t *testing.T) {
	events, err := parsePrivilegeEventsJSON("")
	if err != nil {
		t.Fatalf("parsePrivilegeEventsJSON failed: %v", err)
	}
	if len(events) != 0 {
		t.Errorf("parsePrivilegeEventsJSON returned %d events, want 0", len(events))
	}
}

func TestGetPrivilegeEventsLinux(t *testing.T) {
	events, err := getPrivilegeEventsLinux()
	if err != nil {
		t.Fatalf("getPrivilegeEventsLinux failed: %v", err)
	}
	if len(events) != 0 {
		t.Errorf("getPrivilegeEventsLinux returned %d events, want 0", len(events))
	}
}

func TestGetEventTimeline_EmptyData(t *testing.T) {
	// This calls GetSecurityEvents which shells out, but that should work on any Windows box
	events, err := GetEventTimeline()
	if err != nil {
		t.Logf("GetEventTimeline failed (expected if no security events): %v", err)
		return
	}
	_ = events
}
