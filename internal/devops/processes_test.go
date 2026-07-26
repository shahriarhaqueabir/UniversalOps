package devops

import (
	"os"
	"testing"

	"github.com/shirou/gopsutil/v4/process"
)

func TestValidatePID(t *testing.T) {
	tests := []struct {
		name    string
		pid     int32
		wantErr bool
		errMsg  string
	}{
		{"zero pid", 0, true, "must be positive"},
		{"negative pid", -1, true, "must be positive"},
		{"pid 1", 1, true, "refusing to target pid 1"},
		{"our own pid", int32(os.Getpid()), true, "refusing to target our own process"},
		{"valid pid", 12345, false, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePID(tt.pid)
			if tt.wantErr {
				if err == nil {
					t.Errorf("validatePID(%d) = nil, want error", tt.pid)
				} else if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("validatePID(%d) error = %q, want %q", tt.pid, err.Error(), tt.errMsg)
				}
			} else if err != nil {
				t.Errorf("validatePID(%d) = %v, want nil", tt.pid, err)
			}
		})
	}
}

func TestKillProcess_ValidatePID(t *testing.T) {
	// KillProcess calls validatePID first; test that it rejects invalid PIDs
	// without attempting to find/kill the process.
	err := KillProcess(0)
	if err == nil {
		t.Fatal("KillProcess(0) should return error")
	}
	if !contains(err.Error(), "must be positive") {
		t.Errorf("KillProcess(0) error = %q, want 'must be positive'", err.Error())
	}

	err = KillProcess(1)
	if err == nil {
		t.Fatal("KillProcess(1) should return error")
	}
	if !contains(err.Error(), "refusing to target pid 1") {
		t.Errorf("KillProcess(1) error = %q, want 'refusing to target pid 1'", err.Error())
	}

	self := int32(os.Getpid())
	err = KillProcess(self)
	if err == nil {
		t.Fatal("KillProcess(self) should return error")
	}
	if !contains(err.Error(), "refusing to target our own process") {
		t.Errorf("KillProcess(self) error = %q, want 'refusing to target our own process'", err.Error())
	}
}

func TestRestartProcess_ValidatePID(t *testing.T) {
	// RestartProcess calls validatePID first; test that it rejects invalid PIDs.
	err := RestartProcess(0)
	if err == nil {
		t.Fatal("RestartProcess(0) should return error")
	}
	if !contains(err.Error(), "must be positive") {
		t.Errorf("RestartProcess(0) error = %q, want 'must be positive'", err.Error())
	}

	err = RestartProcess(1)
	if err == nil {
		t.Fatal("RestartProcess(1) should return error")
	}
	if !contains(err.Error(), "refusing to target pid 1") {
		t.Errorf("RestartProcess(1) error = %q, want 'refusing to target pid 1'", err.Error())
	}

	self := int32(os.Getpid())
	err = RestartProcess(self)
	if err == nil {
		t.Fatal("RestartProcess(self) should return error")
	}
	if !contains(err.Error(), "refusing to target our own process") {
		t.Errorf("RestartProcess(self) error = %q, want 'refusing to target our own process'", err.Error())
	}
}

// contains is a helper for substring matching.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && containsStr(s, substr)
}

// containsStr is a simple readable substring check.
func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestProcessEntryFromProcess(t *testing.T) {
	// Create a real process object for the current process
	proc, err := process.NewProcess(int32(os.Getpid()))
	if err != nil {
		t.Fatalf("NewProcess failed: %v", err)
	}

	entry, ok := processEntryFromProcess(proc)
	if !ok {
		t.Fatal("processEntryFromProcess returned false")
	}
	if entry.PID != int32(os.Getpid()) {
		t.Errorf("entry.PID = %d, want %d", entry.PID, os.Getpid())
	}
	if entry.Name == "" {
		t.Error("entry.Name should not be empty")
	}
	t.Logf("Process entry: PID=%d Name=%s CPU=%.1f Mem=%.1f Status=%s",
		entry.PID, entry.Name, entry.CPU, entry.Memory, entry.Status)
}

func TestListProcesses(t *testing.T) {
	procs, err := ListProcesses(5)
	if err != nil {
		t.Fatalf("ListProcesses failed: %v", err)
	}
	if len(procs) == 0 {
		t.Error("ListProcesses returned no processes")
	}
	if len(procs) > 5 {
		t.Errorf("ListProcesses(5) returned %d entries, want <=5", len(procs))
	}
	t.Logf("Got %d processes, first: PID=%d Name=%s", len(procs), procs[0].PID, procs[0].Name)
}

func TestListProcesses_NoLimit(t *testing.T) {
	procs, err := ListProcesses(0)
	if err != nil {
		t.Fatalf("ListProcesses(0) failed: %v", err)
	}
	if len(procs) == 0 {
		t.Error("ListProcesses(0) returned no processes")
	}
	t.Logf("Got %d processes (no limit)", len(procs))
}
