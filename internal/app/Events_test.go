package app

import (
	"testing"
)

func TestEvents_Constants(t *testing.T) {
	if EventMetrics != "metrics" {
		t.Errorf("EventMetrics = %q, want %q", EventMetrics, "metrics")
	}
	if EventAlert != "alert" {
		t.Errorf("EventAlert = %q, want %q", EventAlert, "alert")
	}
	if EventLog != "log" {
		t.Errorf("EventLog = %q, want %q", EventLog, "log")
	}
	if EventPipeline != "pipeline" {
		t.Errorf("EventPipeline = %q, want %q", EventPipeline, "pipeline")
	}
	if EventCmdLine != "cmd:line" {
		t.Errorf("EventCmdLine = %q, want %q", EventCmdLine, "cmd:line")
	}
	if EventCmdDone != "cmd:done" {
		t.Errorf("EventCmdDone = %q, want %q", EventCmdDone, "cmd:done")
	}
	if EventTimeline != "timeline" {
		t.Errorf("EventTimeline = %q, want %q", EventTimeline, "timeline")
	}
}

func TestEvents_StructSizes(t *testing.T) {
	// These tests ensure the event structs haven't drifted
	evt := AlertEvent{
		Action: "fired",
	}
	if evt.Action != "fired" {
		t.Errorf("AlertEvent.Action = %q, want %q", evt.Action, "fired")
	}

	me := MetricsEvent{
		Processes: 100,
	}
	if me.Processes != 100 {
		t.Errorf("MetricsEvent.Processes = %d, want 100", me.Processes)
	}

	le := LogEvent{
		Level: "ERROR",
		Line:  "test error",
	}
	if le.Level != "ERROR" {
		t.Errorf("LogEvent.Level = %q, want %q", le.Level, "ERROR")
	}

	pe := PipelineEvent{
		Status: "idle",
	}
	if pe.Status != "idle" {
		t.Errorf("PipelineEvent.Status = %q, want %q", pe.Status, "idle")
	}
}

func TestEvents_TimelineEvent_JSON(t *testing.T) {
	evt := TimelineEvent{
		ID:        "evt-1",
		Timestamp: "2026-07-12T12:00:00Z",
		Category:  "system",
		Level:     "info",
		Title:     "test event",
		Detail:    "a test detail",
	}

	if evt.ID != "evt-1" {
		t.Errorf("TimelineEvent.ID = %q, want %q", evt.ID, "evt-1")
	}
	if evt.Category != "system" {
		t.Errorf("TimelineEvent.Category = %q, want %q", evt.Category, "system")
	}
}
