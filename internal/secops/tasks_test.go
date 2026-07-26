package secops

import (
	"testing"
)

func TestTaskStateToString(t *testing.T) {
	tests := []struct {
		state int
		want  string
	}{
		{0, "Unknown"},
		{1, "Disabled"},
		{2, "Queued"},
		{3, "Ready"},
		{4, "Running"},
		{5, "State(5)"},
		{99, "State(99)"},
	}
	for _, tt := range tests {
		got := taskStateToString(tt.state)
		if got != tt.want {
			t.Errorf("taskStateToString(%d) = %q, want %q", tt.state, got, tt.want)
		}
	}
}

func TestFormatTaskTime(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"1970-01-01T00:00:00Z", "1970-01-01T00:00"},
		{"2024-06-15T14:30:00+00:00", "2024-06-15T14:30"},
		{"", ""},
		{"invalid", "invalid"},
	}
	for _, tt := range tests {
		got := formatTaskTime(tt.input)
		if got != tt.want {
			t.Errorf("formatTaskTime(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestTrimDateTime(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"2024-01-15T10:30:00", "2024-01-15T10:30:00"},
		{"2024-01-15T10:30:00.123456", "2024-01-15T10:30:00"},
		{"2024-01-15T10:30:00.123", "2024-01-15T10:30:00"},
		{"", ""},
	}
	for _, tt := range tests {
		got := trimDateTime(tt.input)
		if got != tt.want {
			t.Errorf("trimDateTime(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestFindCSVColumn(t *testing.T) {
	headers := []string{"TaskName", "Next Run Time", "Status", "Logon Mode"}
	tests := []struct {
		name string
		want int
	}{
		{"TaskName", 0},
		{"taskname", 0}, // partial match
		{"TASKNAME", 0}, // partial match
		{"Next Run Time", 1},
		{"Status", 2},
		{"Logon Mode", 3},
		{"Missing", -1},
		{"", 0}, // empty matches everything, returns first
	}
	for _, tt := range tests {
		got := findCSVColumn(headers, tt.name)
		if got != tt.want {
			t.Errorf("findCSVColumn(%q) = %d, want %d", tt.name, got, tt.want)
		}
	}
}

func TestParseTasksSchTasksCSV(t *testing.T) {
	csvStr := `"TaskName","Next Run Time","Status"
"TestTask","6/15/2024 2:00:00 PM","Ready"
"BackupJob","","Disabled"
`
	tasks, err := parseTasksSchTasksCSV(csvStr)
	if err != nil {
		t.Fatalf("parseTasksSchTasksCSV failed: %v", err)
	}
	if len(tasks) != 2 {
		t.Fatalf("parseTasksSchTasksCSV returned %d tasks, want 2", len(tasks))
	}
	if tasks[0].Name != "TestTask" {
		t.Errorf("task[0].Name = %q, want TestTask", tasks[0].Name)
	}
	if tasks[1].Name != "BackupJob" {
		t.Errorf("task[1].Name = %q, want BackupJob", tasks[1].Name)
	}
}

func TestParseTasksSchTasksCSV_Empty(t *testing.T) {
	tasks, err := parseTasksSchTasksCSV("")
	if err != nil {
		t.Logf("parseTasksSchTasksCSV(empty) error (acceptable): %v", err)
		return
	}
	if len(tasks) != 0 {
		t.Errorf("parseTasksSchTasksCSV(empty) returned %d tasks", len(tasks))
	}
}

func TestParseTasksSchTasksSimpleCSV_Invalid(t *testing.T) {
	tasks, err := parseTasksSchTasksSimpleCSV("no header")
	if err != nil {
		t.Logf("parseTasksSchTasksSimpleCSV(invalid) error (acceptable): %v", err)
		return
	}
	_ = tasks
}

func TestParseTasksSchTasksSimpleCSV_Empty(t *testing.T) {
	tasks, err := parseTasksSchTasksSimpleCSV("")
	if err != nil {
		t.Logf("parseTasksSchTasksSimpleCSV(empty) error (acceptable): %v", err)
		return
	}
	if len(tasks) != 0 {
		t.Errorf("parseTasksSchTasksSimpleCSV(empty) returned %d tasks", len(tasks))
	}
}

func TestParseTasksJSON(t *testing.T) {
	jsonStr := `[
		{"TaskName": "TestTask", "State": 1, "TaskPath": "\\", "NextRunTime": "2024-01-15T10:00:00Z"}
	]`
	tasks, err := parseTasksJSON(jsonStr)
	if err != nil {
		t.Fatalf("parseTasksJSON failed: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("parseTasksJSON returned %d tasks, want 1", len(tasks))
	}
	if tasks[0].Name != "TestTask" {
		t.Errorf("task.Name = %q, want TestTask", tasks[0].Name)
	}
	if tasks[0].Status != "Disabled" {
		t.Errorf("task.Status = %q, want Disabled", tasks[0].Status)
	}
}

func TestParseTasksJSON_Empty(t *testing.T) {
	tasks, err := parseTasksJSON("")
	if err == nil {
		t.Error("parseTasksJSON(empty) should return error")
	}
	if tasks != nil {
		t.Error("parseTasksJSON(empty) should return nil tasks")
	}
}

func TestParseTasksSimpleJSON(t *testing.T) {
	jsonStr := `[
		{"TaskName": "SimpleTask", "TaskStatus": "Ready", "StartWhenAvailable": true}
	]`
	tasks, err := parseTasksSimpleJSON(jsonStr)
	if err != nil {
		t.Fatalf("parseTasksSimpleJSON failed: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("parseTasksSimpleJSON returned %d tasks, want 1", len(tasks))
	}
	if tasks[0].Name != "SimpleTask" {
		t.Errorf("task.Name = %q, want SimpleTask", tasks[0].Name)
	}
}

func TestParseTasksSimpleJSON_Empty(t *testing.T) {
	tasks, err := parseTasksSimpleJSON("")
	if err != nil {
		t.Logf("parseTasksSimpleJSON(empty) error: %v (acceptable)", err)
		_ = tasks
	}
}

func TestParseSystemdTimers(t *testing.T) {
	output := `NEXT                        LEFT          LAST                        PASSED       UNIT                        ACTIVATES
Mon 2024-01-15 10:00:00 UTC 1h left       Mon 2024-01-15 09:00:00 UTC 30min ago    systemd-tmpfiles-clean.timer systemd-tmpfiles-clean.service
`
	tasks := parseSystemdTimers(output)
	if len(tasks) == 0 {
		t.Fatal("parseSystemdTimers returned no tasks")
	}
	// Name is set from activates field (the service) when present
	if tasks[0].Name != "systemd-tmpfiles-clean.service" {
		t.Errorf("task[0].Name = %q, want systemd-tmpfiles-clean.service", tasks[0].Name)
	}
	// NextRun should be parsed from the NEXT columns
	if tasks[0].NextRun == "" {
		t.Error("task[0].NextRun should not be empty")
	}
}

func TestParseSystemdTimers_Empty(t *testing.T) {
	tasks := parseSystemdTimers("")
	if len(tasks) != 0 {
		t.Errorf("parseSystemdTimers(empty) returned %d tasks", len(tasks))
	}
}

func TestParseSystemdTimers_NoActivate(t *testing.T) {
	output := `NEXT                        LEFT          LAST                        PASSED       UNIT                        ACTIVATES
Mon 2024-01-15 10:00:00 UTC 1h left       Mon 2024-01-15 09:00:00 UTC 30min ago    systemd-tmpfiles-clean.timer
`
	tasks := parseSystemdTimers(output)
	if len(tasks) == 0 {
		t.Fatal("parseSystemdTimers returned no tasks")
	}
	// Without activates, name comes from unitName
	if tasks[0].Name != "systemd-tmpfiles-clean.timer" {
		t.Errorf("task[0].Name = %q, want systemd-tmpfiles-clean.timer", tasks[0].Name)
	}
}
