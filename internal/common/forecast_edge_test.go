package common

import (
	"math"
	"testing"
)

func TestForecastEngine_NaN(t *testing.T) {
	fe := NewForecastEngine(10)
	fe.Push(10.0)
	fe.Push(math.NaN())
	fe.Push(20.0)

	trend := fe.DetectTrend()
	// Regression with NaN values usually results in NaN slope/intercept.
	// We want to ensure it doesn't panic.
	if math.IsNaN(trend.Slope) {
		t.Log("Trend slope is NaN as expected when input contains NaN")
	}

	pred := fe.Predict(1)
	if math.IsNaN(pred) {
		t.Log("Prediction is NaN as expected")
	}
}

func TestForecastEngine_Inf(t *testing.T) {
	fe := NewForecastEngine(10)
	fe.Push(10.0)
	fe.Push(math.Inf(1))
	fe.Push(20.0)

	trend := fe.DetectTrend()
	if math.IsNaN(trend.Slope) || math.IsInf(trend.Slope, 0) {
		t.Logf("Trend slope is %v as expected when input contains Inf", trend.Slope)
	}

	pred := fe.Predict(1)
	if math.IsNaN(pred) || math.IsInf(pred, 0) {
		t.Logf("Prediction is %v as expected", pred)
	}
}

func TestForecastEngine_AllSame(t *testing.T) {
	fe := NewForecastEngine(10)
	for i := 0; i < 5; i++ {
		fe.Push(42.0)
	}

	trend := fe.DetectTrend()
	if trend.Slope != 0 {
		t.Errorf("Expected 0 slope for constant values, got %f", trend.Slope)
	}
	if trend.Direction != TrendStable {
		t.Errorf("Expected TrendStable, got %v", trend.Direction)
	}
}

func TestForecastEngine_Empty(t *testing.T) {
	fe := NewForecastEngine(10)
	trend := fe.DetectTrend()
	if trend.Direction != TrendStable {
		t.Errorf("Expected TrendStable for empty engine, got %v", trend.Direction)
	}

	pred := fe.Predict(1)
	if pred != 0 && !math.IsNaN(pred) {
		t.Logf("Prediction for empty: %f", pred)
	}
}
