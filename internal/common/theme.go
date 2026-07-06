package common

import (
	"os"
	"path/filepath"
	"strings"
)

// ThemeName identifies a supported color theme.
type ThemeName string

const (
	ThemeDefault      ThemeName = "default"
	ThemeLight        ThemeName = "light"
	ThemeDark         ThemeName = "dark"
	ThemeHighContrast ThemeName = "high-contrast"
)

// Palette contains the shared application colors.
type Palette struct {
	Primary   string
	Secondary string
	Warning   string
	Danger    string
	Info      string
	Muted     string
	Text      string
	Background string
	Border    string
}

var palettes = map[ThemeName]Palette{
	ThemeDefault: {
		Primary:    ColorPrimary,
		Secondary:  ColorSecondary,
		Warning:    ColorWarning,
		Danger:     ColorDanger,
		Info:       ColorInfo,
		Muted:      ColorMuted,
		Text:       ColorText,
		Background: ColorDarkBg,
		Border:     ColorBorder,
	},
	ThemeDark: {
		Primary:    "#A78BFA",
		Secondary:  "#34D399",
		Warning:    "#FBBF24",
		Danger:     "#F87171",
		Info:       "#60A5FA",
		Muted:      "#9CA3AF",
		Text:       "#F9FAFB",
		Background: "#111827",
		Border:     "#4B5563",
	},
	ThemeLight: {
		Primary:    "#5B21B6",
		Secondary:  "#047857",
		Warning:    "#B45309",
		Danger:     "#B91C1C",
		Info:       "#1D4ED8",
		Muted:      "#4B5563",
		Text:       "#111827",
		Background: "#E5E7EB",
		Border:     "#9CA3AF",
	},
	ThemeHighContrast: {
		Primary:    "#00FFFF",
		Secondary:  "#00FF00",
		Warning:    "#FFFF00",
		Danger:     "#FF5555",
		Info:       "#55AAFF",
		Muted:      "#FFFFFF",
		Text:       "#FFFFFF",
		Background: "#000000",
		Border:     "#FFFFFF",
	},
}

var activeTheme = ThemeDefault

// ParseThemeName normalizes user supplied theme names.
func ParseThemeName(name string) ThemeName {
	switch ThemeName(strings.ToLower(strings.TrimSpace(name))) {
	case ThemeLight:
		return ThemeLight
	case ThemeDark:
		return ThemeDark
	case ThemeHighContrast, "high_contrast", "contrast":
		return ThemeHighContrast
	default:
		return ThemeDefault
	}
}

// ThemeFromEnv returns the configured theme from HAWKWARD_THEME.
func ThemeFromEnv() ThemeName {
	return ParseThemeName(os.Getenv("HAWKWARD_THEME"))
}

// SetTheme updates the shared common styles for a theme.
func SetTheme(name ThemeName) ThemeName {
	activeTheme = ParseThemeName(string(name))
	rebuildCommonStyles(CurrentPalette())
	_ = SaveTheme(activeTheme) // ignore persistence error
	return activeTheme
}

// LoadTheme loads the saved theme from the config directory.
func LoadTheme() ThemeName {
	dir, err := ConfigDir()
	if err != nil {
		return ThemeDefault
	}
	data, err := os.ReadFile(filepath.Join(dir, ".theme"))
	if err != nil {
		return ThemeDefault
	}
	return ParseThemeName(string(data))
}

// SaveTheme saves the theme to the config directory.
func SaveTheme(name ThemeName) error {
	dir, err := ConfigDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, ".theme"), []byte(name), 0644)
}

// CurrentTheme returns the active common theme.
func CurrentTheme() ThemeName {
	return activeTheme
}

// CurrentPalette returns the active common color palette.
func CurrentPalette() Palette {
	palette, ok := palettes[activeTheme]
	if !ok {
		return palettes[ThemeDefault]
	}
	return palette
}

// PaletteForTheme returns a palette for a theme name, defaulting safely.
func PaletteForTheme(name ThemeName) Palette {
	palette, ok := palettes[ParseThemeName(string(name))]
	if !ok {
		return palettes[ThemeDefault]
	}
	return palette
}

// NextTheme cycles to the next available theme.
func NextTheme() ThemeName {
	allThemes := []ThemeName{ThemeDefault, ThemeDark, ThemeLight, ThemeHighContrast}
	for i, name := range allThemes {
		if name == activeTheme {
			nextIdx := (i + 1) % len(allThemes)
			return SetTheme(allThemes[nextIdx])
		}
	}
	return SetTheme(ThemeDefault)
}
