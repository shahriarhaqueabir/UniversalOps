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

	// Cycle to Dark
	NextTheme()
	if CurrentTheme() != ThemeDark {
		t.Errorf("NextTheme() should go to dark, got %q", CurrentTheme())
	}

	// Cycle to Light
	NextTheme()
	if CurrentTheme() != ThemeLight {
		t.Errorf("NextTheme() should go to light, got %q", CurrentTheme())
	}

	// Cycle to High Contrast
	NextTheme()
	if CurrentTheme() != ThemeHighContrast {
		t.Errorf("NextTheme() should go to high-contrast, got %q", CurrentTheme())
	}

	// Cycle back to Default
	NextTheme()
	if CurrentTheme() != ThemeDefault {
		t.Errorf("NextTheme() should wrap to default, got %q", CurrentTheme())
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
