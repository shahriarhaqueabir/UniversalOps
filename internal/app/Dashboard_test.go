package app

import (
	"testing"

	"github.com/shahriarhaqueabir/UniversalOps/internal/common"
)

func TestGetDashboardData(t *testing.T) {
	common.InitStorage(":memory:")
	defer common.GetStorage().Close()

	app := NewApp()

	// Pre-populate pipeline with a metric to ensure units are set
	app.pipeline.PushMetric("cpu.percent", "%", 10.0)

	d := NewDashboard(app.pipeline, app.alerts, app.SysOps, app.NetOps, app.SecOps, app.DevOps, app.AIOps, app.Timeline, nil, func() string { return "" })
	data := d.GetDashboardData()

	// Initial data might be empty but shouldn't crash
	if data.Uptime == "" {
		t.Log("Uptime is empty (expected on uninitialized app)")
	}

	if data.CPU.Unit != "%" {
		t.Errorf("Expected CPU unit %%, got %s", data.CPU.Unit)
	}
}

func TestDashboardLayoutPersistence(t *testing.T) {
	common.InitStorage(":memory:")
	defer common.GetStorage().Close()

	d := NewDashboard(nil, nil, nil, nil, nil, nil, nil, nil, nil, func() string { return "" })

	// Initially empty.
	if got := d.GetDashboardLayout(); got != "" {
		t.Fatalf("expected empty layout, got %q", got)
	}

	layout := `[{"id":"aiops"},{"id":"secops"},{"id":"slo"},{"id":"devops"}]`
	if err := d.SaveDashboardLayout(layout); err != nil {
		t.Fatalf("SaveDashboardLayout failed: %v", err)
	}

	if got := d.GetDashboardLayout(); got != layout {
		t.Fatalf("expected layout %q, got %q", layout, got)
	}

	// Empty layout should be rejected.
	if err := d.SaveDashboardLayout(""); err == nil {
		t.Fatal("expected error saving empty layout")
	}

	// Reset clears it.
	if err := d.ResetDashboardLayout(); err != nil {
		t.Fatalf("ResetDashboardLayout failed: %v", err)
	}
	if got := d.GetDashboardLayout(); got != "" {
		t.Fatalf("expected empty layout after reset, got %q", got)
	}
}
