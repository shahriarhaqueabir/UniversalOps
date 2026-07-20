package app

import (
	"testing"

	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
)

func TestGetDashboardData(t *testing.T) {
	common.InitStorage(":memory:")
	defer common.GetStorage().Close()

	app := NewApp()

	// Pre-populate pipeline with a metric to ensure units are set
	app.pipeline.PushMetric("cpu.percent", "%", 10.0)

	d := NewDashboard(app.pipeline, app.alerts, app.SysOps, app.NetOps, app.Timeline, func() string { return "" })
	data := d.GetDashboardData()

	// Initial data might be empty but shouldn't crash
	if data.Uptime == "" {
		t.Log("Uptime is empty (expected on uninitialized app)")
	}

	if data.CPU.Unit != "%" {
		t.Errorf("Expected CPU unit %%, got %s", data.CPU.Unit)
	}
}
