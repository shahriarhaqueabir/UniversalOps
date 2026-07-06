package common

import (
	"testing"
)

func TestParseThemeName(t *testing.T) {
	tests := []struct {
		input    string
		expected ThemeName
	}{
		{"default", ThemeDefault},
		{"LIGHT", ThemeLight},
		{"dark ", ThemeDark},
		{"high-contrast", ThemeHighContrast},
		{"invalid", ThemeDefault},
	}

	for _, tt := range tests {
		result := ParseThemeName(tt.input)
		if result != tt.expected {
			t.Errorf("ParseThemeName(%q) = %v; want %v", tt.input, result, tt.expected)
		}
	}
}

func TestRenderSparkline(t *testing.T) {
	// Simple sanity check for sparkline rendering
	values := []float64{0, 10, 20, 30, 40, 50}
	spark := RenderSparkline(values, 10)
	if len(spark) != 10 {
		t.Errorf("Expected sparkline width 10, got %d", len(spark))
	}
}
