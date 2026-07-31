package app

import (
	"testing"
	"time"

	"github.com/shahriarhaqueabir/UniversalOps/internal/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewApp(t *testing.T) {
	a := NewApp()
	require.NotNil(t, a, "NewApp() returned nil")
	assert.NotNil(t, a.pipeline, "pipeline is nil")
	assert.NotNil(t, a.alerts, "alerts is nil")
	assert.NotNil(t, a.SysOps, "SysOps is nil")
	assert.NotNil(t, a.NetOps, "NetOps is nil")
}

func TestGetAppInfo(t *testing.T) {
	a := NewApp()
	a.startedAt = time.Now().Add(-1 * time.Hour)
	info := a.GetAppInfo()

	assert.Equal(t, "Universal-Ops Operations Platform", info.Name)
	assert.NotEmpty(t, info.Uptime, "Uptime should not be empty")
}

func TestSafeLastN(t *testing.T) {
	values := []float64{1, 2, 3, 4, 5}

	tests := []struct {
		name string
		n    int
		want int
	}{
		{"n smaller than len", 3, 3},
		{"n larger than len", 10, 5},
		{"n equal to len", 5, 5},
		{"n is zero", 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := safeLastN(values, tt.n)
			assert.Len(t, got, tt.want)
		})
	}
}

func TestLastValue(t *testing.T) {
	tests := []struct {
		name   string
		values []float64
		want   float64
	}{
		{"non-empty", []float64{1, 2, 3}, 3},
		{"single value", []float64{5}, 5},
		{"empty", []float64{}, 0},
		{"nil", nil, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := lastValue(tt.values)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTrendDirectionString(t *testing.T) {
	tests := []struct {
		dir  common.TrendDirection
		want string
	}{
		{common.TrendRising, "rising"},
		{common.TrendFalling, "falling"},
		{common.TrendStable, "stable"},
		{99, "stable"}, // Unknown
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := trendDirectionString(tt.dir)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestValidateOllamaEnv(t *testing.T) {
	validateOllamaEnv()
}

func TestValidateOllamaEnv_Custom(t *testing.T) {
	t.Setenv("OLLAMA_HOST", "http://example.com:11434")
	t.Setenv("OLLAMA_MODEL", "test-model")
	validateOllamaEnv()
}

func TestValidateOllamaEnv_InvalidURL(t *testing.T) {
	t.Setenv("OLLAMA_HOST", "://invalid")
	validateOllamaEnv()
}

func TestConvertAlert(t *testing.T) {
	now := time.Now()
	alert := common.Alert{
		ID:        "test-1",
		Level:     common.AlertCritical,
		Metric:    "CPU",
		Message:   "Too high",
		Value:     95,
		Threshold: 90,
		Timestamp: now,
		Resolved:  false,
	}

	got := convertAlert(alert)

	assert.Equal(t, alert.ID, got.ID)
	assert.Equal(t, "CRITICAL", got.Level)
	assert.NotEmpty(t, got.Timestamp, "Timestamp should not be empty")
}
