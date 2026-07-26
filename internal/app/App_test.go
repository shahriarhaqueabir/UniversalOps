package app

import (
	"testing"
	"time"

	"github.com/shahriarhaqueabir/UniversalOps/internal/common"
)

func TestNewApp(t *testing.T) {
	a := NewApp()
	if a == nil {
		t.Fatal("NewApp() returned nil")
	}
	if a.pipeline == nil {
		t.Error("pipeline is nil")
	}
	if a.alerts == nil {
		t.Error("alerts is nil")
	}
	if a.SysOps == nil {
		t.Error("SysOps is nil")
	}
	if a.NetOps == nil {
		t.Error("NetOps is nil")
	}
}

func TestGetAppInfo(t *testing.T) {
	a := NewApp()
	a.startedAt = time.Now().Add(-1 * time.Hour)
	info := a.GetAppInfo()

	if info.Name != "Universal-Ops Operations Platform" {
		t.Errorf("Name = %q, want %q", info.Name, "Universal-Ops Operations Platform")
	}
	if info.Uptime == "" {
		t.Error("Uptime is empty")
	}
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
			if len(got) != tt.want {
				t.Errorf("len(safeLastN) = %d, want %d", len(got), tt.want)
			}
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
			if got != tt.want {
				t.Errorf("lastValue() = %v, want %v", got, tt.want)
			}
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
			if got != tt.want {
				t.Errorf("trendDirectionString(%v) = %q, want %q", tt.dir, got, tt.want)
			}
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

	if got.ID != alert.ID {
		t.Errorf("ID = %q, want %q", got.ID, alert.ID)
	}
	if got.Level != "CRITICAL" {
		t.Errorf("Level = %q, want %q", got.Level, "CRITICAL")
	}
	if got.Timestamp == "" {
		t.Error("Timestamp is empty")
	}
}
