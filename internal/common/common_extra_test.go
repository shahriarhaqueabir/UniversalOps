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
		{"high_contrast", ThemeHighContrast},
		{"contrast", ThemeHighContrast},
		{"squib-dark", ThemeSquibDark},
		{"squib_dark", ThemeSquibDark},
		{"squib-light", ThemeSquibLight},
		{"squib_light", ThemeSquibLight},
		{"amber", ThemeAmber},
		{"green", ThemeGreen},
		{"dracula", ThemeDracula},
		{"nord", ThemeNord},
		{"invalid", ThemeDefault},
		{"", ThemeDefault},
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

func TestAllThemePalettesAreValid(t *testing.T) {
	for _, name := range ThemeNames() {
		p := PaletteForTheme(name)
		if p.Primary == "" || p.Secondary == "" {
			t.Errorf("Theme %q missing basic colors", name)
		}
		if p.Text == "" || p.Background == "" {
			t.Errorf("Theme %q missing text/background colors", name)
		}
	}
}
