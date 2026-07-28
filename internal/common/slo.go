package common

import (
	"fmt"
	"time"
)

// ── SLO/SLI Types ────────────────────────────────────────────────────────────

// SLODefinition defines a service level objective for a tracked metric.
type SLODefinition struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Metric      string  `json:"metric"`     // e.g., "cpu", "memory", "disk"
	Comparison  string  `json:"comparison"` // "lt" (less than), "gt" (greater than)
	Threshold   float64 `json:"threshold"`  // e.g., 80.0 for CPU < 80%
	TargetPct   float64 `json:"targetPct"`  // e.g., 95.0 means 95% of time
	WindowDays  int     `json:"windowDays"` // evaluation window in days
	Enabled     bool    `json:"enabled"`
	Description string  `json:"description"`
}

// SLIResult holds the result of evaluating an SLO against actual metric data.
type SLIResult struct {
	SLOID        string  `json:"sloId"`
	SLOName      string  `json:"sloName"`
	CompliantPct float64 `json:"compliantPct"` // % of time within threshold
	TargetPct    float64 `json:"targetPct"`
	Met          bool    `json:"met"`     // compliantPct >= targetPct
	Samples      int     `json:"samples"` // total data points evaluated
	EvaluatedAt  string  `json:"evaluatedAt"`
}

// SLOSummary holds the overall SLO/SLI dashboard summary.
type SLOSummary struct {
	TotalSLOs  int         `json:"totalSLOs"`
	MetCount   int         `json:"metCount"`
	MissCount  int         `json:"missCount"`
	OverallPct float64     `json:"overallPct"`
	Results    []SLIResult `json:"results"`
}

// ── SLO Engine ───────────────────────────────────────────────────────────────

// SLOEngine evaluates SLO definitions against stored metric data.
type SLOEngine struct {
	store *Storage
}

// NewSLOEngine creates a new SLO evaluation engine.
func NewSLOEngine(store *Storage) *SLOEngine {
	return &SLOEngine{store: store}
}

// EvaluateAll evaluates all enabled SLO definitions and returns a summary.
func (e *SLOEngine) EvaluateAll() (SLOSummary, error) {
	defs, err := e.store.ListSLODefinitions()
	if err != nil {
		return SLOSummary{}, fmt.Errorf("list SLOs: %w", err)
	}

	var results []SLIResult
	metCount := 0

	for _, slo := range defs {
		if !slo.Enabled {
			continue
		}
		result, err := e.evaluateOne(slo)
		if err != nil {
			LogWarn("SLO evaluation failed for %q: %v", slo.Name, err)
			continue
		}
		if result.Met {
			metCount++
		}
		results = append(results, result)
	}

	total := len(results)
	overallPct := 100.0
	if total > 0 {
		overallPct = float64(metCount) / float64(total) * 100.0
	}

	return SLOSummary{
		TotalSLOs:  total,
		MetCount:   metCount,
		MissCount:  total - metCount,
		OverallPct: overallPct,
		Results:    results,
	}, nil
}

// evaluateOne evaluates a single SLO definition.
func (e *SLOEngine) evaluateOne(slo SLODefinition) (SLIResult, error) {
	values, err := e.store.GetMetricValuesInWindow(slo.Metric, slo.WindowDays)
	if err != nil {
		return SLIResult{}, fmt.Errorf("get metric values: %w", err)
	}

	if len(values) == 0 {
		return SLIResult{
			SLOID:        slo.ID,
			SLOName:      slo.Name,
			CompliantPct: 100.0,
			TargetPct:    slo.TargetPct,
			Met:          true,
			Samples:      0,
			EvaluatedAt:  time.Now().UTC().Format(time.RFC3339),
		}, nil
	}

	compliant := 0
	for _, v := range values {
		switch slo.Comparison {
		case "lt":
			if v < slo.Threshold {
				compliant++
			}
		case "gt":
			if v > slo.Threshold {
				compliant++
			}
		case "lte":
			if v <= slo.Threshold {
				compliant++
			}
		case "gte":
			if v >= slo.Threshold {
				compliant++
			}
		}
	}

	compliantPct := float64(compliant) / float64(len(values)) * 100.0

	return SLIResult{
		SLOID:        slo.ID,
		SLOName:      slo.Name,
		CompliantPct: compliantPct,
		TargetPct:    slo.TargetPct,
		Met:          compliantPct >= slo.TargetPct,
		Samples:      len(values),
		EvaluatedAt:  time.Now().UTC().Format(time.RFC3339),
	}, nil
}

// SeedDefaultSLOs creates a set of default SLO definitions if none exist.
func (e *SLOEngine) SeedDefaultSLOs() error {
	existing, err := e.store.ListSLODefinitions()
	if err != nil {
		return err
	}
	if len(existing) > 0 {
		return nil // already seeded
	}

	defaults := []SLODefinition{
		{
			ID:          "slo-cpu",
			Name:        "CPU Utilization",
			Metric:      "cpu",
			Comparison:  "lt",
			Threshold:   80.0,
			TargetPct:   95.0,
			WindowDays:  7,
			Enabled:     true,
			Description: "CPU should stay below 80% at least 95% of the time",
		},
		{
			ID:          "slo-memory",
			Name:        "Memory Utilization",
			Metric:      "memory",
			Comparison:  "lt",
			Threshold:   85.0,
			TargetPct:   95.0,
			WindowDays:  7,
			Enabled:     true,
			Description: "Memory should stay below 85% at least 95% of the time",
		},
		{
			ID:          "slo-disk",
			Name:        "Disk Utilization",
			Metric:      "disk",
			Comparison:  "lt",
			Threshold:   90.0,
			TargetPct:   90.0,
			WindowDays:  7,
			Enabled:     true,
			Description: "Disk should stay below 90% at least 90% of the time",
		},
	}

	for _, slo := range defaults {
		if err := e.store.UpsertSLODefinition(slo); err != nil {
			return fmt.Errorf("seed SLO %q: %w", slo.Name, err)
		}
	}

	LogInfo("Seeded %d default SLO definitions", len(defaults))
	return nil
}
