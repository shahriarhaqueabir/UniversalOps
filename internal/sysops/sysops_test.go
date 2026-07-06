package sysops

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
)

func TestNewModel(t *testing.T) {
	m := NewModel()
	if m == nil {
		t.Fatal("NewModel() returned nil")
	}
	if m.tabIndex != 0 {
		t.Errorf("tabIndex = %d, want 0", m.tabIndex)
	}
	if m.ready {
		t.Error("ready should be false initially")
	}
	if m.stats != nil {
		t.Error("stats should be nil initially")
	}
}

func TestModelState(t *testing.T) {
	m := NewModel()

	if m.Ready() {
		t.Error("Ready() should be false initially")
	}

	if m.TabIndex() != 0 {
		t.Errorf("TabIndex() = %d, want 0", m.TabIndex())
	}

	if m.Error() != nil {
		t.Errorf("Error() = %v, want nil", m.Error())
	}

	// Mark ready
	m.ready = true
	if !m.Ready() {
		t.Error("Ready() should be true")
	}
}

func TestModelString(t *testing.T) {
	m := NewModel()
	s := m.String()
	if s != "SysOps: no data" {
		t.Errorf("String() = %q, want %q", s, "SysOps: no data")
	}
}

func TestRenderBar(t *testing.T) {
	tests := []struct {
		name  string
		value float64
		max   float64
		width float64
		want  string
	}{
		{"zero value", 0, 100, 10, "░░░░░░░░░░"},
		{"full value", 100, 100, 10, "██████████"},
		{"half value", 50, 100, 10, "█████░░░░░"},
		{"quarter value", 25, 100, 10, "██░░░░░░░░"},
		{"exceeds max", 120, 100, 10, "██████████"},
		{"zero max", 50, 0, 10, "░░░░░░░░░░"},
		{"narrow bar", 50, 100, 5, "██░░░"},
		{"wide bar", 50, 100, 20, "██████████░░░░░░░░░░"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := renderBar(tt.value, tt.max, tt.width)
			if got != tt.want {
				t.Errorf("renderBar(%v, %v, %v) = %q, want %q",
					tt.value, tt.max, tt.width, got, tt.want)
			}
		})
	}
}

func TestGetHealthColor(t *testing.T) {
	tests := []struct {
		name    string
		pct     float64
		wantHex string
	}{
		{"low usage", 30, "#10B981"},
		{"moderate usage", 50, "#10B981"},
		{"borderline warning", 70, "#10B981"},
		{"warning level", 75, "#F59E0B"},
		{"high warning", 85, "#F59E0B"},
		{"critical level", 91, "#EF4444"},
		{"maximum", 100, "#EF4444"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := common.GetHealthColor(tt.pct)
			// Convert to RGBA for comparison
			gotR, gotG, gotB, _ := got.RGBA()
			// Parse wanted hex
			wantR, wantG, wantB := parseHexColor(tt.wantHex)

			if gotR != wantR || gotG != wantG || gotB != wantB {
				t.Errorf("common.GetHealthColor(%v) = rgba(%d,%d,%d), want %s (rgba(%d,%d,%d))",
					tt.pct, gotR>>8, gotG>>8, gotB>>8, tt.wantHex, wantR>>8, wantG>>8, wantB>>8)
			}
		})
	}
}

func TestGetHealthColor_Boundaries(t *testing.T) {
	// Test exact boundary values (70 and 90)
	green := common.GetHealthColor(70)
	amber := common.GetHealthColor(71)
	amber2 := common.GetHealthColor(90)
	red := common.GetHealthColor(91)

	gr, gg, gb, _ := green.RGBA()
	ar, ag, ab, _ := amber.RGBA()
	ar2, ag2, ab2, _ := amber2.RGBA()
	rr, rg, rb, _ := red.RGBA()

	if gr == ar && gg == ag && gb == ab {
		t.Error("70 and 71 should return different colors")
	}
	if ar2 == rr && ag2 == rg && ab2 == rb {
		t.Error("90 and 91 should return different colors")
	}
}

// TestDiskUsageStruct tests DiskUsage struct creation.
func TestDiskUsageStruct(t *testing.T) {
	d := DiskUsage{
		Mountpoint:  "/",
		TotalBytes:  500000000000,
		FreeBytes:   100000000000,
		UsedBytes:   400000000000,
		UsedPercent: 80.0,
		FSType:      "ext4",
		Device:      "/dev/sda1",
	}
	if d.Mountpoint != "/" {
		t.Errorf("Mountpoint = %q, want %q", d.Mountpoint, "/")
	}
	if d.UsedPercent != 80.0 {
		t.Errorf("UsedPercent = %v, want 80.0", d.UsedPercent)
	}
}

// TestCPUStatsStruct tests CPUStats struct creation.
func TestCPUStatsStruct(t *testing.T) {
	c := CPUStats{
		Percent:   45.5,
		ModelName: "Intel Core i7",
		CoreCount: 8,
	}
	if c.Percent != 45.5 {
		t.Errorf("Percent = %v, want 45.5", c.Percent)
	}
	if c.CoreCount != 8 {
		t.Errorf("CoreCount = %d, want 8", c.CoreCount)
	}
}

// TestMemoryStatsStruct tests MemoryStats struct creation.
func TestMemoryStatsStruct(t *testing.T) {
	m := MemoryStats{
		TotalBytes:     17179869184, // 16 GB
		AvailableBytes: 8589934592,  // 8 GB
		UsedPercent:    50.0,
	}
	if m.TotalBytes != 17179869184 {
		t.Errorf("TotalBytes = %d, want 17179869184", m.TotalBytes)
	}
	if m.UsedPercent != 50.0 {
		t.Errorf("UsedPercent = %v, want 50.0", m.UsedPercent)
	}
}

func TestBuildSystemStatsPartialData(t *testing.T) {
	stats := buildSystemStats(
		&CPUStats{Percent: 0},
		&MemoryStats{
			TotalBytes:  8 * 1024 * 1024 * 1024,
			UsedBytes:   4 * 1024 * 1024 * 1024,
			UsedPercent: 50,
		},
		nil,
		&SystemInfo{
			UptimeSeconds: 3600,
			ProcessCount:  42,
		},
	)

	if stats.CPUPercent != 0 {
		t.Errorf("CPUPercent = %v, want 0", stats.CPUPercent)
	}
	if stats.MemoryUsed != 50 {
		t.Errorf("MemoryUsed = %v, want 50", stats.MemoryUsed)
	}
	if stats.MemoryTotalGB != 8 {
		t.Errorf("MemoryTotalGB = %v, want 8", stats.MemoryTotalGB)
	}
	if stats.DiskUsed != 0 {
		t.Errorf("DiskUsed = %v, want 0 when disk stats unavailable", stats.DiskUsed)
	}
	if stats.ProcessCount != 42 {
		t.Errorf("ProcessCount = %d, want 42", stats.ProcessCount)
	}
	if stats.Uptime == "" {
		t.Error("Uptime should be formatted when system info is available")
	}
}

func TestPrimaryDiskUsage(t *testing.T) {
	t.Run("prefers root mount", func(t *testing.T) {
		used, free := primaryDiskUsage(&DiskStats{Usage: []DiskUsage{
			{Mountpoint: "/mnt/data", UsedPercent: 80, FreeBytes: 20},
			{Mountpoint: "/", UsedPercent: 55, FreeBytes: 45},
		}})
		if used != 55 || free != 45 {
			t.Errorf("primaryDiskUsage() = (%v, %d), want (55, 45)", used, free)
		}
	})

	t.Run("falls back to first disk", func(t *testing.T) {
		used, free := primaryDiskUsage(&DiskStats{Usage: []DiskUsage{
			{Mountpoint: "/mnt/data", UsedPercent: 80, FreeBytes: 20},
		}})
		if used != 80 || free != 20 {
			t.Errorf("primaryDiskUsage() = (%v, %d), want (80, 20)", used, free)
		}
	})

	t.Run("handles nil disk stats", func(t *testing.T) {
		used, free := primaryDiskUsage(nil)
		if used != 0 || free != 0 {
			t.Errorf("primaryDiskUsage(nil) = (%v, %d), want zero values", used, free)
		}
	})
}

// TestProcessInfoStruct tests ProcessInfo struct creation.
func TestProcessInfoStruct(t *testing.T) {
	p := ProcessInfo{
		PID:    1234,
		Name:   "test.exe",
		CPU:    12.5,
		Memory: 45.6,
		Status: "running",
	}
	if p.PID != 1234 {
		t.Errorf("PID = %d, want 1234", p.PID)
	}
	if p.Name != "test.exe" {
		t.Errorf("Name = %q, want %q", p.Name, "test.exe")
	}
}

// TestSystemInfoStruct tests SystemInfo struct creation.
func TestSystemInfoStruct(t *testing.T) {
	s := SystemInfo{
		Hostname:      "myhost",
		OS:            "windows",
		Platform:      "Windows 11",
		KernelVersion: "10.0.22631",
		UptimeSeconds: 3600,
		ProcessCount:  120,
	}
	if s.Hostname != "myhost" {
		t.Errorf("Hostname = %q, want %q", s.Hostname, "myhost")
	}
	if s.ProcessCount != 120 {
		t.Errorf("ProcessCount = %d, want 120", s.ProcessCount)
	}
}

// parseHexColor converts a hex color string to RGBA values.
func parseHexColor(hex string) (r, g, b uint32) {
	// Strip # prefix
	if len(hex) > 0 && hex[0] == '#' {
		hex = hex[1:]
	}
	if len(hex) != 6 {
		return 0, 0, 0
	}
	r = uint32(hexToByte(hex[0:2])) * 257 // scale to 16-bit
	g = uint32(hexToByte(hex[2:4])) * 257
	b = uint32(hexToByte(hex[4:6])) * 257
	return
}

func hexToByte(s string) byte {
	var v byte
	for i := 0; i < 2; i++ {
		c := s[i]
		switch {
		case c >= '0' && c <= '9':
			v = v*16 + byte(c-'0')
		case c >= 'a' && c <= 'f':
			v = v*16 + byte(c-'a'+10)
		case c >= 'A' && c <= 'F':
			v = v*16 + byte(c-'A'+10)
		}
	}
	return v
}

// TestStyleConstants verifies that shared common styles are initialized.
func TestStyleConstants(t *testing.T) {
	panelRendered := common.Panel.Render("test")
	if panelRendered == "" {
		t.Error("common.Panel.Render('test') returned empty")
	}

	titleRendered := common.PanelTitle.Render("test")
	if titleRendered == "" {
		t.Error("common.PanelTitle.Render('test') returned empty")
	}

	labelRendered := common.Label.Render("test")
	if labelRendered == "" {
		t.Error("common.Label.Render('test') returned empty")
	}

	valueRendered := common.Value.Render("test")
	if valueRendered == "" {
		t.Error("common.Value.Render('test') returned empty")
	}
}

func TestModelUpdate_StatsResult(t *testing.T) {
	m := NewModel()
	stats := &common.SystemStats{
		CPUPercent: 15.0,
		MemoryUsed: 40.0,
	}

	m.Update(StatsResult{Stats: stats, Err: nil})

	if !m.ready {
		t.Error("model should be ready after stats update")
	}
	if m.stats.CPUPercent != 15.0 {
		t.Errorf("CPUPercent = %v, want 15.0", m.stats.CPUPercent)
	}
	if m.err != nil {
		t.Errorf("err = %v, want nil", m.err)
	}
}

func TestModelUpdate_StatsResultError(t *testing.T) {
	m := NewModel()
	err := fmt.Errorf("stats error")

	m.Update(StatsResult{Stats: nil, Err: err})

	if m.ready {
		t.Error("model should not be ready after failed stats update")
	}
	if m.err != err {
		t.Errorf("err = %v, want %v", m.err, err)
	}
}

func TestModelUpdate_WorkflowResultMsg(t *testing.T) {
	m := NewModel()
	report := "All systems normal"

	m.Update(WorkflowResultMsg{Report: report, Err: nil})

	if !m.ready {
		t.Error("model should be ready after workflow report")
	}
	if !m.showReport {
		t.Error("showReport should be true")
	}
	if m.workflowReport != report {
		t.Errorf("workflowReport = %q, want %q", m.workflowReport, report)
	}
}

func TestModelHandleKeyPress(t *testing.T) {
	m := NewModel()

	// Tab navigation
	m.handleKeyPress(tea.KeyPressMsg{Code: tea.KeyTab})
	if m.tabIndex != 1 {
		t.Errorf("tabIndex = %d, want 1", m.tabIndex)
	}

	m.handleKeyPress(tea.KeyPressMsg{Code: tea.KeyRight})
	if m.tabIndex != 2 {
		t.Errorf("tabIndex = %d, want 2", m.tabIndex)
	}

	m.handleKeyPress(tea.KeyPressMsg{Text: "l"})
	if m.tabIndex != 0 {
		t.Errorf("tabIndex = %d, want 0 (wrapped)", m.tabIndex)
	}

	// Direct jump
	m.handleKeyPress(tea.KeyPressMsg{Text: "3"})
	if m.tabIndex != 2 {
		t.Errorf("tabIndex = %d, want 2", m.tabIndex)
	}

	// Backwards
	m.handleKeyPress(tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift})
	if m.tabIndex != 1 {
		t.Errorf("tabIndex = %d, want 1", m.tabIndex)
	}

	// Refresh
	cmd := m.handleKeyPress(tea.KeyPressMsg{Text: "r"})
	if cmd == nil {
		t.Error("refresh should return a command")
	}
	if m.ready {
		t.Error("ready should be false after refresh")
	}
}

func TestModelView(t *testing.T) {
	m := NewModel()
	stats := &common.SystemStats{
		CPUPercent: 45.0,
		MemoryUsed: 60.0,
		Uptime:     "2 hours",
	}
	m.stats = stats
	m.ready = true

	view := m.View(80, 24, nil)
	if !strings.Contains(view, "System Operations Dashboard") {
		t.Error("view missing title")
	}
	if !strings.Contains(view, "45.0%") {
		t.Error("view missing CPU stats")
	}
	if !strings.Contains(view, "2 hours") {
		t.Error("view missing uptime")
	}
}

func TestModelView_NoStats(t *testing.T) {
	m := NewModel()
	view := m.View(80, 24, nil)
	if !strings.Contains(view, "Collecting data") {
		t.Error("view should show collecting data message")
	}
}

func TestModelView_WorkflowReport(t *testing.T) {
	m := NewModel()
	m.showReport = true
	m.workflowReport = "Critical health issue"

	view := m.View(80, 24, nil)
	if !strings.Contains(view, "System Health Report") {
		t.Error("view missing health report title")
	}
	if !strings.Contains(view, "Critical health issue") {
		t.Error("view missing report content")
	}
}
