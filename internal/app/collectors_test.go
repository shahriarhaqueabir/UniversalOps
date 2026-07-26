package app

import (
	"context"
	"testing"

	"github.com/shahriarhaqueabir/UniversalOps/internal/common"
)

func TestRegisterCollectors(t *testing.T) {
	a := NewApp()
	registry := common.NewCollectorRegistry()
	RegisterCollectors(registry, a)

	snapshot := registry.Snapshot()
	if len(snapshot) != 12 {
		t.Fatalf("expected 12 registered collectors, got %d", len(snapshot))
	}

	// Verify all expected IDs are present
	expectedIDs := map[common.CollectorID]bool{
		common.CollectorCPU:    false,
		common.CollectorMem:    false,
		common.CollectorDisk:   false,
		common.CollectorNet:    false,
		common.CollectorTemp:   false,
		common.CollectorProc:   false,
		common.CollectorUptime: false,
		common.CollectorLoad:   false,
		common.CollectorSwap:   false,
		common.CollectorDiskIO: false,
		common.CollectorOpenFD: false,
		"gpu":                  false,
	}

	for _, s := range snapshot {
		if _, ok := expectedIDs[s.ID]; !ok {
			t.Errorf("unexpected collector ID: %s", s.ID)
		}
		expectedIDs[s.ID] = true
	}

	for id, found := range expectedIDs {
		if !found {
			t.Errorf("expected collector %s not registered", id)
		}
	}
}

func TestUptimeCollector_Info(t *testing.T) {
	c := &uptimeCollector{}
	info := c.Info()
	if info.ID != common.CollectorUptime {
		t.Errorf("expected ID %s, got %s", common.CollectorUptime, info.ID)
	}
	if info.DefaultInterval == 0 {
		t.Error("expected non-zero default interval")
	}
}

func TestUptimeCollector_Collect(t *testing.T) {
	c := &uptimeCollector{}
	samples, err := c.Collect(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(samples) != 1 {
		t.Fatalf("expected 1 sample, got %d", len(samples))
	}
	if samples[0].Name != "system.uptime" {
		t.Errorf("expected name 'system.uptime', got %q", samples[0].Name)
	}
	if samples[0].Unit != "s" {
		t.Errorf("expected unit 's', got %q", samples[0].Unit)
	}
}

func TestLoadAvgCollector_Info(t *testing.T) {
	c := &loadAvgCollector{}
	info := c.Info()
	if info.ID != common.CollectorLoad {
		t.Errorf("expected ID %s, got %s", common.CollectorLoad, info.ID)
	}
	if !info.DefaultEnabled {
		t.Error("expected load avg collector to be enabled by default")
	}
}

func TestLoadAvgCollector_Collect(t *testing.T) {
	c := &loadAvgCollector{}
	samples, err := c.Collect(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(samples) != 3 {
		t.Fatalf("expected 3 samples (1m, 5m, 15m), got %d", len(samples))
	}
	expectedNames := []string{"load.1m", "load.5m", "load.15m"}
	for i, name := range expectedNames {
		if samples[i].Name != name {
			t.Errorf("sample %d: expected name %q, got %q", i, name, samples[i].Name)
		}
	}
}

func TestSwapCollector_Info(t *testing.T) {
	c := &swapCollector{}
	info := c.Info()
	if info.ID != common.CollectorSwap {
		t.Errorf("expected ID %s, got %s", common.CollectorSwap, info.ID)
	}
}

func TestSwapCollector_Collect(t *testing.T) {
	c := &swapCollector{}
	samples, err := c.Collect(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(samples) != 1 {
		t.Fatalf("expected 1 sample, got %d", len(samples))
	}
	if samples[0].Name != "swap.percent" {
		t.Errorf("expected name 'swap.percent', got %q", samples[0].Name)
	}
	if samples[0].Unit != "%" {
		t.Errorf("expected unit '%%', got %q", samples[0].Unit)
	}
}

func TestDiskIOCollector_Info(t *testing.T) {
	c := &diskIOCollector{}
	info := c.Info()
	if info.ID != common.CollectorDiskIO {
		t.Errorf("expected ID %s, got %s", common.CollectorDiskIO, info.ID)
	}
}

func TestDiskIOCollector_Collect(t *testing.T) {
	c := &diskIOCollector{}
	samples, err := c.Collect(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(samples) != 2 {
		t.Fatalf("expected 2 samples, got %d", len(samples))
	}
	expectedNames := []string{"disk.io.read", "disk.io.write"}
	for i, name := range expectedNames {
		if samples[i].Name != name {
			t.Errorf("sample %d: expected name %q, got %q", i, name, samples[i].Name)
		}
	}
}

func TestOpenFDCollector_Info(t *testing.T) {
	c := &openFDCollector{}
	info := c.Info()
	if info.ID != common.CollectorOpenFD {
		t.Errorf("expected ID %s, got %s", common.CollectorOpenFD, info.ID)
	}
	if info.DefaultEnabled {
		t.Error("expected open FD collector to be disabled by default (expensive on Windows)")
	}
}

func TestOpenFDCollector_Collect(t *testing.T) {
	c := &openFDCollector{}
	samples, err := c.Collect(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(samples) != 1 {
		t.Fatalf("expected 1 sample, got %d", len(samples))
	}
	if samples[0].Name != "system.open_fds" {
		t.Errorf("expected name 'system.open_fds', got %q", samples[0].Name)
	}
	if samples[0].Value < 0 {
		t.Errorf("expected non-negative FD count, got %f", samples[0].Value)
	}
}

func TestProcessCountCollector_Collect(t *testing.T) {
	c := &processCountCollector{}
	samples, err := c.Collect(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(samples) != 1 {
		t.Fatalf("expected 1 sample, got %d", len(samples))
	}
	if samples[0].Name != "process.count" {
		t.Errorf("expected name 'process.count', got %q", samples[0].Name)
	}
	if samples[0].Value <= 0 {
		t.Errorf("expected positive process count, got %f", samples[0].Value)
	}
}

func TestTemperatureCollector_Info(t *testing.T) {
	c := &temperatureCollector{}
	info := c.Info()
	if info.ID != common.CollectorTemp {
		t.Errorf("expected ID %s, got %s", common.CollectorTemp, info.ID)
	}
	if info.DefaultEnabled {
		t.Error("expected temperature collector to be disabled by default")
	}
}
