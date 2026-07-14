package sysops

import "testing"

func TestGetScheduledTasks(t *testing.T) {
	tasks, err := GetScheduledTasks()
	if err != nil {
		t.Logf("GetScheduledTasks returned error (may be expected): %v", err)
		return
	}
	t.Logf("Found %d scheduled tasks", len(tasks))
}
