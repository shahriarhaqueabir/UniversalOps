package sysops

import (
	"strings"
	"testing"
)

func TestRunHealthCheck(t *testing.T) {
	report, err := RunHealthCheck()
	if err != nil {
		t.Logf("RunHealthCheck returned error (may be expected): %v", err)
		return
	}
	if report == nil {
		t.Fatal("RunHealthCheck returned nil report")
	}
	t.Logf("CPU=%v Mem=%v Disk=%v Sys=%v Procs=%d",
		report.CPU != nil, report.Memory != nil, report.Disk != nil, report.System != nil, len(report.Procs))
}

func TestHealthReportString_Full(t *testing.T) {
	r := &HealthReport{
		CPU: &CPUStats{Percent: 45.2, CoreCount: 8},
		Memory: &MemoryStats{
			TotalBytes:     16 * 1024 * 1024 * 1024,
			AvailableBytes: 8 * 1024 * 1024 * 1024,
			UsedPercent:    50.0,
		},
		Disk: &DiskUsage{
			Mountpoint:  "C:\\",
			TotalBytes:  500 * 1024 * 1024 * 1024,
			UsedBytes:   250 * 1024 * 1024 * 1024,
			FreeBytes:   250 * 1024 * 1024 * 1024,
			UsedPercent: 50.0,
		},
		System: &SystemInfo{
			Hostname:        "test-pc",
			Platform:        "Windows",
			PlatformVersion: "10.0.22631",
			KernelVersion:   "10.0.22631",
			UptimeSeconds:   86400,
			ProcessCount:    200,
		},
		Procs: []ProcessInfo{
			{PID: 1234, Name: "chrome.exe", CPU: 15.2, Memory: 450.5},
			{PID: 5678, Name: "code.exe", CPU: 5.1, Memory: 320.0},
		},
	}

	output := r.String()

	checks := []string{
		"System Health Report",
		"CPU: 45.2% across 8 cores",
		"MEM: 50.0% used (8.0/16.0 GB)",
		"DISK: 50.0% used",
		"UPTIME: 1d 0h 0m",
		"PROCS: 200 running",
		"chrome.exe",
		"code.exe",
	}
	for _, c := range checks {
		if !strings.Contains(output, c) {
			t.Errorf("String() missing expected text %q", c)
		}
	}
}

func TestHealthReportString_Partial(t *testing.T) {
	t.Run("nil CPU", func(t *testing.T) {
		r := &HealthReport{CPU: nil, Memory: nil, Disk: nil, System: nil}
		output := r.String()
		if !strings.Contains(output, "System Health Report") {
			t.Error("String() should still include header")
		}
	})

	t.Run("only processes", func(t *testing.T) {
		r := &HealthReport{
			CPU: &CPUStats{Percent: 10.0, CoreCount: 2},
			Procs: []ProcessInfo{
				{PID: 1, Name: "init", CPU: 0.5, Memory: 10.0},
			},
		}
		output := r.String()
		if !strings.Contains(output, "init") {
			t.Errorf("String() should include process name, got: %s", output)
		}
	})
}

func TestHealthReportMarkdown_Full(t *testing.T) {
	r := &HealthReport{
		CPU: &CPUStats{Percent: 30.0, CoreCount: 4},
		Memory: &MemoryStats{
			TotalBytes:     8 * 1024 * 1024 * 1024,
			AvailableBytes: 4 * 1024 * 1024 * 1024,
			UsedPercent:    50.0,
		},
		Disk: &DiskUsage{
			Mountpoint:  "/",
			TotalBytes:  256 * 1024 * 1024 * 1024,
			UsedBytes:   128 * 1024 * 1024 * 1024,
			FreeBytes:   128 * 1024 * 1024 * 1024,
			UsedPercent: 50.0,
		},
		System: &SystemInfo{
			Hostname:        "srv-01",
			Platform:        "Linux",
			PlatformVersion: "5.15.0",
			KernelVersion:   "5.15.0-generic",
			UptimeSeconds:   172800,
			ProcessCount:    500,
		},
		Procs: []ProcessInfo{
			{PID: 100, Name: "nginx", CPU: 2.5, Memory: 64.0},
		},
	}

	output := r.Markdown()

	checks := []string{
		"System Health Report",
		"Hostname | srv-01",
		"Linux 5.15.0",
		"5.15.0-generic",
		"Uptime",
		"2d 0h 0m",
		"500",
		"30.0",
		"50.0",
		"Disk",
		"nginx",
	}
	for _, c := range checks {
		if !strings.Contains(output, c) {
			t.Errorf("Markdown() missing expected text %q", c)
		}
	}
}

func TestHealthReportMarkdown_Partial(t *testing.T) {
	t.Run("empty report", func(t *testing.T) {
		r := &HealthReport{}
		output := r.Markdown()
		if !strings.Contains(output, "System Health Report") {
			t.Error("Markdown() should still include header even with empty report")
		}
		if !strings.Contains(output, "Resource Usage") {
			t.Error("Markdown() should include Resource Usage header")
		}
	})

	t.Run("only system info", func(t *testing.T) {
		r := &HealthReport{
			System: &SystemInfo{
				Hostname:        "mini-pc",
				Platform:        "macOS",
				PlatformVersion: "14.0",
				KernelVersion:   "23.0.0",
				UptimeSeconds:   3600,
				ProcessCount:    150,
			},
		}
		output := r.Markdown()
		if !strings.Contains(output, "mini-pc") {
			t.Errorf("Markdown() should include hostname, got: %s", output)
		}
		if !strings.Contains(output, "macOS") {
			t.Errorf("Markdown() should include platform, got: %s", output)
		}
	})
}
