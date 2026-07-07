package app

import (
	"testing"
)

func TestSysOps_GetCPUInfo(t *testing.T) {
	a := NewApp()
	s := NewSysOps(a)

	info := s.GetCPUInfo()
	// On most systems this should return something, on CI it might be empty
	// but we check it doesn't crash and returns the type.
	if info.ModelName == "" && info.Percent == 0 && info.CoreCount == 0 {
		t.Log("CPU info might be empty on this environment")
	}
}

func TestSysOps_GetMemoryInfo(t *testing.T) {
	a := NewApp()
	s := NewSysOps(a)

	info := s.GetMemoryInfo()
	if info.TotalBytes == 0 {
		t.Log("Memory info might be empty on this environment")
	}
}

func TestSysOps_GetDiskInfo(t *testing.T) {
	a := NewApp()
	s := NewSysOps(a)

	info := s.GetDiskInfo()
	// Should at least return the struct
	_ = info
}

func TestSysOps_GetTopProcesses(t *testing.T) {
	a := NewApp()
	s := NewSysOps(a)

	procs := s.GetTopProcesses(5)
	if len(procs) > 5 {
		t.Errorf("len(procs) = %d, want <= 5", len(procs))
	}
}

func TestSysOps_GetSystemInfo(t *testing.T) {
	a := NewApp()
	s := NewSysOps(a)

	info := s.GetSystemInfo()
	if info.Hostname == "" {
		t.Log("Hostname might be empty on this environment")
	}
}
