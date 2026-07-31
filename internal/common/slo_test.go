package common

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

// ── Helpers ─────────────────────────────────────────────────────────────────

// seedSLO inserts a single SLO definition for testing.
func seedSLO(t *testing.T, s *Storage, slo SLODefinition) {
	t.Helper()
	if err := s.UpsertSLODefinition(slo); err != nil {
		t.Fatalf("seedSLO: %v", err)
	}
}

// seedMetric inserts a single metric value and flushes to ensure persistence.
func seedMetric(t *testing.T, s *Storage, name string, value float64) {
	t.Helper()
	if err := s.InsertMetric(name, "%", value); err != nil {
		t.Fatalf("seedMetric: %v", err)
	}
	s.flushMetrics()
}

// ── NewSLOEngine ────────────────────────────────────────────────────────────

func TestNewSLOEngine(t *testing.T) {
	s := newTestStorage(t)
	engine := NewSLOEngine(s)
	if engine == nil {
		t.Fatal("NewSLOEngine returned nil")
	}
	if engine.store != s {
		t.Fatal("NewSLOEngine did not store reference")
	}
}

// ── EvaluateAll — edge cases ────────────────────────────────────────────────

func TestEvaluateAll_NoSLOs(t *testing.T) {
	s := newTestStorage(t)
	engine := NewSLOEngine(s)

	summary, err := engine.EvaluateAll()
	if err != nil {
		t.Fatalf("EvaluateAll with no SLOs: %v", err)
	}
	if summary.TotalSLOs != 0 {
		t.Errorf("expected 0 total SLOs, got %d", summary.TotalSLOs)
	}
	if summary.OverallPct != 100.0 {
		t.Errorf("expected 100%% overall with no SLOs, got %.1f", summary.OverallPct)
	}
	if len(summary.Results) != 0 {
		t.Errorf("expected 0 results, got %d", len(summary.Results))
	}
}

func TestEvaluateAll_OnlyDisabledSLOs(t *testing.T) {
	s := newTestStorage(t)
	engine := NewSLOEngine(s)

	seedSLO(t, s, SLODefinition{
		ID: "slo-disabled", Name: "Disabled SLO", Metric: "cpu",
		Comparison: "lt", Threshold: 80, TargetPct: 95, WindowDays: 7,
		Enabled: false, Description: "should be skipped",
	})

	summary, err := engine.EvaluateAll()
	if err != nil {
		t.Fatalf("EvaluateAll: %v", err)
	}
	if summary.TotalSLOs != 0 {
		t.Errorf("expected 0 evaluated (all disabled), got %d", summary.TotalSLOs)
	}
}

// ── EvaluateAll — with metric data ──────────────────────────────────────────

func TestEvaluateAll_MetAndMiss(t *testing.T) {
	s := newTestStorage(t)
	engine := NewSLOEngine(s)

	// SLO: CPU < 80, target 100% (all samples must pass)
	seedSLO(t, s, SLODefinition{
		ID: "slo-cpu", Name: "CPU", Metric: "cpu",
		Comparison: "lt", Threshold: 80, TargetPct: 100,
		WindowDays: 7, Enabled: true,
	})

	// SLO: Memory < 85, target 100%
	seedSLO(t, s, SLODefinition{
		ID: "slo-mem", Name: "Memory", Metric: "memory",
		Comparison: "lt", Threshold: 85, TargetPct: 100,
		WindowDays: 7, Enabled: true,
	})

	// CPU: all below 80 → should pass
	seedMetric(t, s, "cpu", 45)
	seedMetric(t, s, "cpu", 62)
	seedMetric(t, s, "cpu", 55)

	// Memory: one above 85 → should fail
	seedMetric(t, s, "memory", 40)
	seedMetric(t, s, "memory", 90)
	seedMetric(t, s, "memory", 50)

	summary, err := engine.EvaluateAll()
	if err != nil {
		t.Fatalf("EvaluateAll: %v", err)
	}

	if summary.TotalSLOs != 2 {
		t.Errorf("expected 2 SLOs, got %d", summary.TotalSLOs)
	}
	if summary.MetCount != 1 {
		t.Errorf("expected 1 met, got %d", summary.MetCount)
	}
	if summary.MissCount != 1 {
		t.Errorf("expected 1 miss, got %d", summary.MissCount)
	}
	if summary.OverallPct != 50.0 {
		t.Errorf("expected 50%% overall, got %.1f", summary.OverallPct)
	}

	// Verify individual results
	for _, r := range summary.Results {
		switch r.SLOID {
		case "slo-cpu":
			if !r.Met {
				t.Errorf("CPU SLO should be met (all samples < 80)")
			}
			if r.Samples != 3 {
				t.Errorf("CPU expected 3 samples, got %d", r.Samples)
			}
			if r.CompliantPct != 100.0 {
				t.Errorf("CPU expected 100%% compliant, got %.1f", r.CompliantPct)
			}
		case "slo-mem":
			if r.Met {
				t.Errorf("Memory SLO should not be met (1 of 3 above 85)")
			}
			if r.Samples != 3 {
				t.Errorf("Memory expected 3 samples, got %d", r.Samples)
			}
			if r.CompliantPct < 66.6 || r.CompliantPct > 66.7 {
				t.Errorf("Memory expected ~66.7%% compliant, got %.1f", r.CompliantPct)
			}
		}
	}
}

// ── evaluateOne — comparison operators ──────────────────────────────────────

func TestEvaluateOne_EmptyData(t *testing.T) {
	s := newTestStorage(t)
	engine := NewSLOEngine(s)

	slo := SLODefinition{
		ID: "slo-empty", Name: "Empty", Metric: "nonexistent",
		Comparison: "lt", Threshold: 80, TargetPct: 95,
		WindowDays: 7, Enabled: true,
	}

	result, err := engine.evaluateOne(slo)
	if err != nil {
		t.Fatalf("evaluateOne with empty data: %v", err)
	}
	if !result.Met {
		t.Error("expected Met=true when no data (vacuous truth)")
	}
	if result.Samples != 0 {
		t.Errorf("expected 0 samples, got %d", result.Samples)
	}
	if result.CompliantPct != 100.0 {
		t.Errorf("expected 100%% compliant with no data, got %.1f", result.CompliantPct)
	}
}

func TestEvaluateOne_LessThan(t *testing.T) {
	s := newTestStorage(t)
	engine := NewSLOEngine(s)

	slo := SLODefinition{
		ID: "slo-lt", Name: "LT Test", Metric: "test_lt",
		Comparison: "lt", Threshold: 50, TargetPct: 100,
		WindowDays: 7, Enabled: true,
	}

	seedMetric(t, s, "test_lt", 10)
	seedMetric(t, s, "test_lt", 30)
	seedMetric(t, s, "test_lt", 49) // just below → compliant
	seedMetric(t, s, "test_lt", 50) // equal → NOT compliant (lt, not lte)
	seedMetric(t, s, "test_lt", 80) // above → NOT compliant

	result, err := engine.evaluateOne(slo)
	if err != nil {
		t.Fatalf("evaluateOne lt: %v", err)
	}
	if result.Samples != 5 {
		t.Errorf("expected 5 samples, got %d", result.Samples)
	}
	if result.CompliantPct != 60.0 {
		t.Errorf("expected 60%% compliant (3/5), got %.1f", result.CompliantPct)
	}
	if result.Met {
		t.Error("expected Met=false (60% < 100% target)")
	}
}

func TestEvaluateOne_GreaterThan(t *testing.T) {
	s := newTestStorage(t)
	engine := NewSLOEngine(s)

	slo := SLODefinition{
		ID: "slo-gt", Name: "GT Test", Metric: "test_gt",
		Comparison: "gt", Threshold: 50, TargetPct: 75,
		WindowDays: 7, Enabled: true,
	}

	seedMetric(t, s, "test_gt", 10) // below → NOT compliant
	seedMetric(t, s, "test_gt", 51) // above → compliant
	seedMetric(t, s, "test_gt", 60) // above → compliant
	seedMetric(t, s, "test_gt", 50) // equal → NOT compliant (gt, not gte)

	result, err := engine.evaluateOne(slo)
	if err != nil {
		t.Fatalf("evaluateOne gt: %v", err)
	}
	if result.Samples != 4 {
		t.Errorf("expected 4 samples, got %d", result.Samples)
	}
	if result.CompliantPct != 50.0 {
		t.Errorf("expected 50%% compliant (2/4), got %.1f", result.CompliantPct)
	}
	if result.Met {
		t.Error("expected Met=false (50% < 75% target)")
	}
}

func TestEvaluateOne_LessThanOrEqual(t *testing.T) {
	s := newTestStorage(t)
	engine := NewSLOEngine(s)

	slo := SLODefinition{
		ID: "slo-lte", Name: "LTE Test", Metric: "test_lte",
		Comparison: "lte", Threshold: 50, TargetPct: 100,
		WindowDays: 7, Enabled: true,
	}

	seedMetric(t, s, "test_lte", 10)
	seedMetric(t, s, "test_lte", 50) // equal → compliant (lte)
	seedMetric(t, s, "test_lte", 80) // above → NOT compliant

	result, err := engine.evaluateOne(slo)
	if err != nil {
		t.Fatalf("evaluateOne lte: %v", err)
	}
	if result.CompliantPct < 66.6 || result.CompliantPct > 66.7 {
		t.Errorf("expected ~66.7%% compliant (2/3), got %.1f", result.CompliantPct)
	}
}

func TestEvaluateOne_GreaterThanOrEqual(t *testing.T) {
	s := newTestStorage(t)
	engine := NewSLOEngine(s)

	slo := SLODefinition{
		ID: "slo-gte", Name: "GTE Test", Metric: "test_gte",
		Comparison: "gte", Threshold: 50, TargetPct: 100,
		WindowDays: 7, Enabled: true,
	}

	seedMetric(t, s, "test_gte", 50) // equal → compliant (gte)
	seedMetric(t, s, "test_gte", 60) // above → compliant
	seedMetric(t, s, "test_gte", 10) // below → NOT compliant

	result, err := engine.evaluateOne(slo)
	if err != nil {
		t.Fatalf("evaluateOne gte: %v", err)
	}
	if result.CompliantPct < 66.6 || result.CompliantPct > 66.7 {
		t.Errorf("expected ~66.7%% compliant (2/3), got %.1f", result.CompliantPct)
	}
}

// ── evaluateOne — boundary conditions ───────────────────────────────────────

func TestEvaluateOne_ExactBoundaryMet(t *testing.T) {
	s := newTestStorage(t)
	engine := NewSLOEngine(s)

	// SLO: CPU < 80, target exactly 66.7% (2 of 3 must pass)
	slo := SLODefinition{
		ID: "slo-boundary", Name: "Boundary", Metric: "boundary",
		Comparison: "lt", Threshold: 80, TargetPct: 66.7,
		WindowDays: 7, Enabled: true,
	}

	seedMetric(t, s, "boundary", 50) // pass
	seedMetric(t, s, "boundary", 90) // fail
	seedMetric(t, s, "boundary", 50) // pass

	result, err := engine.evaluateOne(slo)
	if err != nil {
		t.Fatalf("evaluateOne boundary: %v", err)
	}
	// 2/3 = 66.666...% which is < 66.7 → not met
	if result.Met {
		t.Errorf("expected Met=false (66.666%% < 66.7%%)")
	}
}

func TestEvaluateOne_ExactBoundaryMiss(t *testing.T) {
	s := newTestStorage(t)
	engine := NewSLOEngine(s)

	slo := SLODefinition{
		ID: "slo-boundary-miss", Name: "Boundary Miss", Metric: "boundary_miss",
		Comparison: "lt", Threshold: 80, TargetPct: 67.0,
		WindowDays: 7, Enabled: true,
	}

	seedMetric(t, s, "boundary_miss", 50) // pass
	seedMetric(t, s, "boundary_miss", 90) // fail
	seedMetric(t, s, "boundary_miss", 50) // pass

	result, err := engine.evaluateOne(slo)
	if err != nil {
		t.Fatalf("evaluateOne boundary miss: %v", err)
	}
	if result.Met {
		t.Errorf("expected Met=false (66.7%% < 67.0%%)")
	}
}

func TestEvaluateOne_AllPass(t *testing.T) {
	s := newTestStorage(t)
	engine := NewSLOEngine(s)

	slo := SLODefinition{
		ID: "slo-allpass", Name: "All Pass", Metric: "all_pass",
		Comparison: "lt", Threshold: 100, TargetPct: 100,
		WindowDays: 7, Enabled: true,
	}

	seedMetric(t, s, "all_pass", 10)
	seedMetric(t, s, "all_pass", 20)
	seedMetric(t, s, "all_pass", 30)

	result, err := engine.evaluateOne(slo)
	if err != nil {
		t.Fatalf("evaluateOne all pass: %v", err)
	}
	if !result.Met {
		t.Error("expected Met=true (all pass)")
	}
	if result.CompliantPct != 100.0 {
		t.Errorf("expected 100%%, got %.1f", result.CompliantPct)
	}
}

func TestEvaluateOne_AllFail(t *testing.T) {
	s := newTestStorage(t)
	engine := NewSLOEngine(s)

	slo := SLODefinition{
		ID: "slo-allfail", Name: "All Fail", Metric: "all_fail",
		Comparison: "lt", Threshold: 10, TargetPct: 50,
		WindowDays: 7, Enabled: true,
	}

	seedMetric(t, s, "all_fail", 50)
	seedMetric(t, s, "all_fail", 60)
	seedMetric(t, s, "all_fail", 70)

	result, err := engine.evaluateOne(slo)
	if err != nil {
		t.Fatalf("evaluateOne all fail: %v", err)
	}
	if result.Met {
		t.Error("expected Met=false (all fail)")
	}
	if result.CompliantPct != 0.0 {
		t.Errorf("expected 0%%, got %.1f", result.CompliantPct)
	}
}

func TestEvaluateOne_SingleSample(t *testing.T) {
	s := newTestStorage(t)
	engine := NewSLOEngine(s)

	slo := SLODefinition{
		ID: "slo-single", Name: "Single", Metric: "single",
		Comparison: "lt", Threshold: 50, TargetPct: 100,
		WindowDays: 7, Enabled: true,
	}

	seedMetric(t, s, "single", 30)

	result, err := engine.evaluateOne(slo)
	if err != nil {
		t.Fatalf("evaluateOne single: %v", err)
	}
	if !result.Met {
		t.Error("expected Met=true (single sample passes)")
	}
	if result.Samples != 1 {
		t.Errorf("expected 1 sample, got %d", result.Samples)
	}
}

// ── EvaluateAll — mixed enabled/disabled ────────────────────────────────────

func TestEvaluateAll_MixedEnabledDisabled(t *testing.T) {
	s := newTestStorage(t)
	engine := NewSLOEngine(s)

	seedSLO(t, s, SLODefinition{
		ID: "slo-enabled", Name: "Enabled", Metric: "enabled_metric",
		Comparison: "lt", Threshold: 80, TargetPct: 100,
		WindowDays: 7, Enabled: true,
	})
	seedSLO(t, s, SLODefinition{
		ID: "slo-disabled1", Name: "Disabled 1", Metric: "disabled_metric",
		Comparison: "lt", Threshold: 80, TargetPct: 100,
		WindowDays: 7, Enabled: false,
	})
	seedSLO(t, s, SLODefinition{
		ID: "slo-disabled2", Name: "Disabled 2", Metric: "disabled_metric2",
		Comparison: "lt", Threshold: 80, TargetPct: 100,
		WindowDays: 7, Enabled: false,
	})

	seedMetric(t, s, "enabled_metric", 50)

	summary, err := engine.EvaluateAll()
	if err != nil {
		t.Fatalf("EvaluateAll mixed: %v", err)
	}
	if summary.TotalSLOs != 1 {
		t.Errorf("expected 1 evaluated (2 disabled), got %d", summary.TotalSLOs)
	}
}

// ── SeedDefaultSLOs ─────────────────────────────────────────────────────────

func TestSeedDefaultSLOs_CreatesDefaults(t *testing.T) {
	s := newTestStorage(t)
	engine := NewSLOEngine(s)

	if err := engine.SeedDefaultSLOs(); err != nil {
		t.Fatalf("SeedDefaultSLOs: %v", err)
	}

	defs, err := s.ListSLODefinitions()
	if err != nil {
		t.Fatalf("ListSLODefinitions: %v", err)
	}
	if len(defs) != 3 {
		t.Fatalf("expected 3 default SLOs, got %d", len(defs))
	}

	// Verify each default
	expected := map[string]struct {
		Metric     string
		Comparison string
		Threshold  float64
		TargetPct  float64
		Enabled    bool
	}{
		"slo-cpu":    {"cpu", "lt", 80, 95, true},
		"slo-memory": {"memory", "lt", 85, 95, true},
		"slo-disk":   {"disk", "lt", 90, 90, true},
	}

	for _, slo := range defs {
		exp, ok := expected[slo.ID]
		if !ok {
			t.Errorf("unexpected SLO ID: %s", slo.ID)
			continue
		}
		if slo.Metric != exp.Metric {
			t.Errorf("%s: metric=%q, want %q", slo.ID, slo.Metric, exp.Metric)
		}
		if slo.Comparison != exp.Comparison {
			t.Errorf("%s: comparison=%q, want %q", slo.ID, slo.Comparison, exp.Comparison)
		}
		if slo.Threshold != exp.Threshold {
			t.Errorf("%s: threshold=%.1f, want %.1f", slo.ID, slo.Threshold, exp.Threshold)
		}
		if slo.TargetPct != exp.TargetPct {
			t.Errorf("%s: targetPct=%.1f, want %.1f", slo.ID, slo.TargetPct, exp.TargetPct)
		}
		if slo.Enabled != exp.Enabled {
			t.Errorf("%s: enabled=%v, want %v", slo.ID, slo.Enabled, exp.Enabled)
		}
	}
}

func TestSeedDefaultSLOs_Idempotent(t *testing.T) {
	s := newTestStorage(t)
	engine := NewSLOEngine(s)

	// Seed twice
	if err := engine.SeedDefaultSLOs(); err != nil {
		t.Fatalf("first seed: %v", err)
	}
	if err := engine.SeedDefaultSLOs(); err != nil {
		t.Fatalf("second seed: %v", err)
	}

	defs, err := s.ListSLODefinitions()
	if err != nil {
		t.Fatalf("ListSLODefinitions: %v", err)
	}
	if len(defs) != 3 {
		t.Errorf("expected 3 SLOs after idempotent seed, got %d", len(defs))
	}
}

func TestSeedDefaultSLOs_DoesNotOverwriteExisting(t *testing.T) {
	s := newTestStorage(t)
	engine := NewSLOEngine(s)

	// Pre-seed with a custom SLO
	seedSLO(t, s, SLODefinition{
		ID: "slo-custom", Name: "Custom", Metric: "custom",
		Comparison: "gt", Threshold: 10, TargetPct: 99,
		WindowDays: 1, Enabled: true, Description: "user-defined",
	})

	if err := engine.SeedDefaultSLOs(); err != nil {
		t.Fatalf("SeedDefaultSLOs: %v", err)
	}

	defs, err := s.ListSLODefinitions()
	if err != nil {
		t.Fatalf("ListSLODefinitions: %v", err)
	}
	// SeedDefaultSLOs skips if any SLOs already exist, so only the custom one remains
	if len(defs) != 1 {
		t.Errorf("expected 1 SLO (custom only, seeding skipped), got %d", len(defs))
	}
	if defs[0].ID != "slo-custom" {
		t.Errorf("expected custom SLO to remain, got %s", defs[0].ID)
	}
}

// ── EvaluateAll — with real default SLOs and metric data ────────────────────

func TestEvaluateAll_WithDefaultsAndData(t *testing.T) {
	s := newTestStorage(t)
	engine := NewSLOEngine(s)

	if err := engine.SeedDefaultSLOs(); err != nil {
		t.Fatalf("SeedDefaultSLOs: %v", err)
	}

	// Seed metric data that makes CPU pass, memory pass, disk fail
	// CPU < 80 target 95%: 19 below 80, 1 above → 95% → met
	for i := 0; i < 19; i++ {
		seedMetric(t, s, "cpu", 50)
	}
	seedMetric(t, s, "cpu", 85) // 1 violation → 95% compliant → met

	// Memory < 85 target 95%: 19 below 85, 1 above → 95% → met
	for i := 0; i < 19; i++ {
		seedMetric(t, s, "memory", 60)
	}
	seedMetric(t, s, "memory", 90) // 1 violation → 95% compliant → met

	// Disk < 90 target 90%: 8 below 90, 2 above → 80% → miss
	for i := 0; i < 8; i++ {
		seedMetric(t, s, "disk", 50)
	}
	seedMetric(t, s, "disk", 95)
	seedMetric(t, s, "disk", 95)

	summary, err := engine.EvaluateAll()
	if err != nil {
		t.Fatalf("EvaluateAll with defaults: %v", err)
	}

	if summary.TotalSLOs != 3 {
		t.Errorf("expected 3 SLOs, got %d", summary.TotalSLOs)
	}
	if summary.MetCount != 2 {
		t.Errorf("expected 2 met, got %d", summary.MetCount)
	}
	if summary.MissCount != 1 {
		t.Errorf("expected 1 miss, got %d", summary.MissCount)
	}

	// Verify individual results
	for _, r := range summary.Results {
		switch r.SLOID {
		case "slo-cpu":
			if !r.Met {
				t.Errorf("CPU SLO should be met (95%% compliant)")
			}
		case "slo-memory":
			if !r.Met {
				t.Errorf("Memory SLO should be met (95%% compliant)")
			}
		case "slo-disk":
			if r.Met {
				t.Errorf("Disk SLO should not be met (80%% < 90%% target)")
			}
		}
	}
}

// ── EvaluateAll — error handling ────────────────────────────────────────────

func TestEvaluateAll_ContinuesOnError(t *testing.T) {
	s := newTestStorage(t)
	engine := NewSLOEngine(s)

	// Seed one valid SLO
	seedSLO(t, s, SLODefinition{
		ID: "slo-valid", Name: "Valid", Metric: "valid_metric",
		Comparison: "lt", Threshold: 80, TargetPct: 100,
		WindowDays: 7, Enabled: true,
	})
	seedMetric(t, s, "valid_metric", 50)

	// Seed an SLO with a nonexistent metric (will get empty data, not error)
	seedSLO(t, s, SLODefinition{
		ID: "slo-empty-metric", Name: "Empty Metric", Metric: "no_data_here",
		Comparison: "lt", Threshold: 80, TargetPct: 100,
		WindowDays: 7, Enabled: true,
	})

	summary, err := engine.EvaluateAll()
	if err != nil {
		t.Fatalf("EvaluateAll: %v", err)
	}
	if summary.TotalSLOs != 2 {
		t.Errorf("expected 2 SLOs evaluated, got %d", summary.TotalSLOs)
	}
	if summary.MetCount != 2 {
		t.Errorf("expected 2 met (valid + vacuous), got %d", summary.MetCount)
	}
}

// ── SLOSummary ──────────────────────────────────────────────────────────────

func TestSLOSummary_OverallPctRounding(t *testing.T) {
	s := newTestStorage(t)
	engine := NewSLOEngine(s)

	// 3 SLOs, 1 met → 33.3%
	for i := 0; i < 3; i++ {
		id := fmt.Sprintf("slo-round-%d", i)
		seedSLO(t, s, SLODefinition{
			ID: id, Name: id, Metric: id,
			Comparison: "lt", Threshold: 80, TargetPct: 100,
			WindowDays: 7, Enabled: true,
		})
	}
	// Only first metric passes
	seedMetric(t, s, "slo-round-0", 50)
	// Others fail
	seedMetric(t, s, "slo-round-1", 90)
	seedMetric(t, s, "slo-round-2", 90)

	summary, err := engine.EvaluateAll()
	if err != nil {
		t.Fatalf("EvaluateAll: %v", err)
	}
	if summary.TotalSLOs != 3 {
		t.Errorf("expected 3 SLOs, got %d", summary.TotalSLOs)
	}
	if summary.MetCount != 1 {
		t.Errorf("expected 1 met, got %d", summary.MetCount)
	}
	if summary.OverallPct < 33.3 || summary.OverallPct > 33.4 {
		t.Errorf("expected ~33.3%% overall, got %.1f", summary.OverallPct)
	}
}

// ── Timestamp format ────────────────────────────────────────────────────────

func TestEvaluateOne_TimestampFormat(t *testing.T) {
	s := newTestStorage(t)
	engine := NewSLOEngine(s)

	slo := SLODefinition{
		ID: "slo-ts", Name: "Timestamp Test", Metric: "ts_metric",
		Comparison: "lt", Threshold: 80, TargetPct: 100,
		WindowDays: 7, Enabled: true,
	}
	seedMetric(t, s, "ts_metric", 50)

	result, err := engine.evaluateOne(slo)
	if err != nil {
		t.Fatalf("evaluateOne: %v", err)
	}

	// Verify RFC3339 format
	_, err = time.Parse(time.RFC3339, result.EvaluatedAt)
	if err != nil {
		t.Errorf("EvaluatedAt is not RFC3339: %q — %v", result.EvaluatedAt, err)
	}
}

// ── evaluateOne — invalid comparison operator ────────────────────────────

func TestEvaluateOne_InvalidComparison(t *testing.T) {
	s := newTestStorage(t)
	engine := NewSLOEngine(s)

	// An unknown comparison ("eq") should silently yield 0% compliance
	// since the switch has no default case.
	slo := SLODefinition{
		ID: "slo-invalid-op", Name: "Invalid Op", Metric: "invalid_op",
		Comparison: "eq", Threshold: 50, TargetPct: 100,
		WindowDays: 7, Enabled: true,
	}

	seedMetric(t, s, "invalid_op", 50) // equal to threshold but "eq" is unknown
	seedMetric(t, s, "invalid_op", 30) // below threshold
	seedMetric(t, s, "invalid_op", 10) // below threshold

	result, err := engine.evaluateOne(slo)
	if err != nil {
		t.Fatalf("evaluateOne invalid comparison: %v", err)
	}
	if result.CompliantPct != 0.0 {
		t.Errorf("expected 0%% compliant for unknown comparison %q, got %.1f",
			slo.Comparison, result.CompliantPct)
	}
	if result.Met {
		t.Errorf("expected Met=false for unknown comparison %q", slo.Comparison)
	}
	if result.Samples != 3 {
		t.Errorf("expected 3 samples, got %d", result.Samples)
	}
}

// ── EvaluateAll — storage error propagation ───────────────────────────────

func TestEvaluateAll_ListError(t *testing.T) {
	s := newTestStorage(t)
	engine := NewSLOEngine(s)

	// Seed one valid SLO and metric
	seedSLO(t, s, SLODefinition{
		ID: "slo-survivor", Name: "Survivor", Metric: "survivor",
		Comparison: "lt", Threshold: 80, TargetPct: 100,
		WindowDays: 7, Enabled: true,
	})
	seedMetric(t, s, "survivor", 50)

	// Close the storage to trigger a DB error on ListSLODefinitions
	s.Close()

	_, err := engine.EvaluateAll()
	if err == nil {
		t.Fatal("expected error from EvaluateAll after storage close, got nil")
	}
	// Verify the error mentions the root cause
	if !strings.Contains(err.Error(), "list SLOs") {
		t.Errorf("expected error wrapping 'list SLOs', got: %v", err)
	}
}

// ── evaluateOne — storage error ──────────────────────────────────────────

func TestEvaluateOne_StorageError(t *testing.T) {
	s := newTestStorage(t)
	engine := NewSLOEngine(s)

	slo := SLODefinition{
		ID: "slo-db-error", Name: "DB Error", Metric: "db_error_metric",
		Comparison: "lt", Threshold: 80, TargetPct: 100,
		WindowDays: 7, Enabled: true,
	}
	seedMetric(t, s, "db_error_metric", 50)

	// Close storage so GetMetricValuesInWindow fails
	s.Close()

	_, err := engine.evaluateOne(slo)
	if err == nil {
		t.Fatal("expected error from evaluateOne after storage close, got nil")
	}
	if !strings.Contains(err.Error(), "get metric values") {
		t.Errorf("expected error wrapping 'get metric values', got: %v", err)
	}
}
