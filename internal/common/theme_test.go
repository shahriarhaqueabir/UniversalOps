package common

import (
	"os"
	"testing"
)

func TestThemeCycling(t *testing.T) {
	// Reset to default
	SetTheme(ThemeDefault)
	if CurrentTheme() != ThemeDefault {
		t.Errorf("CurrentTheme() = %q, want %q", CurrentTheme(), ThemeDefault)
	}

	// Cycle through all themes and verify each step
	allThemes := ThemeNames()
	startIdx := 0
	for i, name := range allThemes {
		if name == ThemeDefault {
			startIdx = i
			break
		}
	}

	for step := 0; step < len(allThemes); step++ {
		expected := allThemes[(startIdx+step)%len(allThemes)]
		if CurrentTheme() != expected {
			t.Errorf("Step %d: CurrentTheme() = %q, want %q", step, CurrentTheme(), expected)
		}
		NextTheme()
	}

	// After a full cycle we should be back at ThemeDefault
	if CurrentTheme() != ThemeDefault {
		t.Errorf("After full cycle, CurrentTheme() = %q, want %q", CurrentTheme(), ThemeDefault)
	}
}

func TestAllThemesAreUnique(t *testing.T) {
	seen := make(map[ThemeName]bool)
	for _, name := range ThemeNames() {
		if seen[name] {
			t.Errorf("Duplicate theme name: %q", name)
		}
		seen[name] = true
	}
	if len(ThemeNames()) != 10 {
		t.Errorf("Expected 10 themes, got %d", len(ThemeNames()))
	}
}

func TestAllPalettesHaveChartColors(t *testing.T) {
	for _, name := range ThemeNames() {
		p := PaletteForTheme(name)
		if p.ChartLine1 == "" || p.ChartLine2 == "" || p.ChartLine3 == "" {
			t.Errorf("Theme %q missing chart line colors", name)
		}
		if p.ChartGrid == "" {
			t.Errorf("Theme %q missing chart grid color", name)
		}
		if p.ChartText == "" {
			t.Errorf("Theme %q missing chart text color", name)
		}
		if p.CardBg == "" || p.CardBorder == "" {
			t.Errorf("Theme %q missing card colors", name)
		}
		if p.CardFocusedBg == "" || p.CardFocusedBorder == "" {
			t.Errorf("Theme %q missing card focused colors", name)
		}
	}
}

func TestThemeCyclingWrapsToDefault(t *testing.T) {
	SetTheme(ThemeDefault)
	// Cycle 10 times (full cycle)
	for i := 0; i < 10; i++ {
		NextTheme()
	}
	if CurrentTheme() != ThemeDefault {
		t.Errorf("After 10 cycles, CurrentTheme() = %q, want %q", CurrentTheme(), ThemeDefault)
	}
}

func TestThemePersistence(t *testing.T) {
	// First ensure clean state
	origTheme := LoadTheme()
	defer SetTheme(origTheme)

	testTheme := ThemeHighContrast
	if err := SaveTheme(testTheme); err != nil {
		t.Fatalf("SaveTheme() error = %v", err)
	}

	loaded := LoadTheme()
	if loaded != testTheme {
		t.Errorf("LoadTheme() = %q, want %q", loaded, testTheme)
	}

	// Test SetTheme persistence
	SetTheme(ThemeLight)
	if LoadTheme() != ThemeLight {
		t.Errorf("SetTheme() did not persist theme")
	}
}

func TestThemeFromEnv(t *testing.T) {
	os.Setenv("HAWKWARD_THEME", "dark")
	defer os.Unsetenv("HAWKWARD_THEME")

	if ThemeFromEnv() != ThemeDark {
		t.Errorf("ThemeFromEnv() = %q, want dark", ThemeFromEnv())
	}
}
