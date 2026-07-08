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
	ThemeDark         ThemeName = "dark"
	ThemeLight        ThemeName = "light"
	ThemeHighContrast ThemeName = "high-contrast"
	ThemeSquibDark    ThemeName = "squib-dark"
	ThemeSquibLight   ThemeName = "squib-light"
	ThemeAmber        ThemeName = "amber"
	ThemeGreen        ThemeName = "green"
	ThemeDracula      ThemeName = "dracula"
	ThemeNord         ThemeName = "nord"
)

// Palette contains the full application color palette including chart and card colors.
type Palette struct {
	// Base semantic colors
	Primary    string
	Secondary  string
	Warning    string
	Danger     string
	Info       string
	Muted      string
	Text       string
	Background string
	Border     string

	// Chart-specific colors
	ChartLine1 string // First series line color
	ChartLine2 string // Second series line color
	ChartLine3 string // Third series line color
	ChartGrid  string // Grid line color
	ChartText  string // Chart label color

	// Card colors
	CardBg            string
	CardBorder        string
	CardFocusedBg     string
	CardFocusedBorder string
}

// allThemeOrder defines the canonical cycling order. Keep this in sync with
// the themes below and update NextTheme when adding or removing entries.
var allThemeOrder = []ThemeName{
	ThemeDefault,
	ThemeDark,
	ThemeLight,
	ThemeHighContrast,
	ThemeSquibDark,
	ThemeSquibLight,
	ThemeAmber,
	ThemeGreen,
	ThemeDracula,
	ThemeNord,
}

var palettes = map[ThemeName]Palette{
	// ── default — Purple/green dark ──
	ThemeDefault: {
		Primary:    "#7C3AED",
		Secondary:  "#10B981",
		Warning:    "#FBBF24",
		Danger:     "#F87171",
		Info:       "#60A5FA",
		Muted:      "#9CA3AF",
		Text:       "#F9FAFB",
		Background: "#111827",
		Border:     "#4B5563",

		ChartLine1: "#7C3AED",
		ChartLine2: "#10B981",
		ChartLine3: "#3B82F6",
		ChartGrid:  "#374151",
		ChartText:  "#9CA3AF",

		CardBg:            "#374151",
		CardBorder:        "#4B5563",
		CardFocusedBg:     "#4B5563",
		CardFocusedBorder: "#7C3AED",
	},

	// ── dark — Purple/green dark variant ──
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

		ChartLine1: "#A78BFA",
		ChartLine2: "#34D399",
		ChartLine3: "#60A5FA",
		ChartGrid:  "#4B5563",
		ChartText:  "#9CA3AF",

		CardBg:            "#1F2937",
		CardBorder:        "#374151",
		CardFocusedBg:     "#374151",
		CardFocusedBorder: "#A78BFA",
	},

	// ── light — Light mode ──
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

		ChartLine1: "#5B21B6",
		ChartLine2: "#047857",
		ChartLine3: "#1D4ED8",
		ChartGrid:  "#9CA3AF",
		ChartText:  "#4B5563",

		CardBg:            "#FFFFFF",
		CardBorder:        "#D1D5DB",
		CardFocusedBg:     "#F3F4F6",
		CardFocusedBorder: "#5B21B6",
	},

	// ── high-contrast — Accessibility ──
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

		ChartLine1: "#00FFFF",
		ChartLine2: "#00FF00",
		ChartLine3: "#55AAFF",
		ChartGrid:  "#FFFFFF",
		ChartText:  "#FFFFFF",

		CardBg:            "#000000",
		CardBorder:        "#FFFFFF",
		CardFocusedBg:     "#000000",
		CardFocusedBorder: "#00FFFF",
	},

	// ── squib-dark — Slate/sky dark (squib inspired) ──
	ThemeSquibDark: {
		Primary:    "#38BDF8",
		Secondary:  "#4ADE80",
		Warning:    "#FBBF24",
		Danger:     "#F87171",
		Info:       "#38BDF8",
		Muted:      "#94A3B8",
		Text:       "#F8FAFC",
		Background: "#0F172A",
		Border:     "#334155",

		ChartLine1: "#38BDF8",
		ChartLine2: "#4ADE80",
		ChartLine3: "#FBBF24",
		ChartGrid:  "#334155",
		ChartText:  "#94A3B8",

		CardBg:            "#1E293B",
		CardBorder:        "#334155",
		CardFocusedBg:     "#334155",
		CardFocusedBorder: "#38BDF8",
	},

	// ── squib-light — Slate/sky light (squib inspired) ──
	ThemeSquibLight: {
		Primary:    "#0EA5E9",
		Secondary:  "#22C55E",
		Warning:    "#EAB308",
		Danger:     "#EF4444",
		Info:       "#0EA5E9",
		Muted:      "#64748B",
		Text:       "#0F172A",
		Background: "#F8FAFC",
		Border:     "#E2E8F0",

		ChartLine1: "#0EA5E9",
		ChartLine2: "#22C55E",
		ChartLine3: "#EAB308",
		ChartGrid:  "#E2E8F0",
		ChartText:  "#64748B",

		CardBg:            "#FFFFFF",
		CardBorder:        "#E2E8F0",
		CardFocusedBg:     "#F1F5F9",
		CardFocusedBorder: "#0EA5E9",
	},

	// ── amber — Amber-on-black retro terminal ──
	ThemeAmber: {
		Primary:    "#FF8800",
		Secondary:  "#FBBF24",
		Warning:    "#FFAA00",
		Danger:     "#FF4444",
		Info:       "#FFAA00",
		Muted:      "#888888",
		Text:       "#FFCC88",
		Background: "#000000",
		Border:     "#333333",

		ChartLine1: "#FF8800",
		ChartLine2: "#FBBF24",
		ChartLine3: "#FF6600",
		ChartGrid:  "#333333",
		ChartText:  "#888888",

		CardBg:            "#111111",
		CardBorder:        "#333333",
		CardFocusedBg:     "#222222",
		CardFocusedBorder: "#FF8800",
	},

	// ── green — Green-on-black classic terminal ──
	ThemeGreen: {
		Primary:    "#00FF00",
		Secondary:  "#4ADE80",
		Warning:    "#AAFF00",
		Danger:     "#FF4444",
		Info:       "#00CCFF",
		Muted:      "#00AA00",
		Text:       "#CCFFCC",
		Background: "#000000",
		Border:     "#003300",

		ChartLine1: "#00FF00",
		ChartLine2: "#4ADE80",
		ChartLine3: "#22C55E",
		ChartGrid:  "#003300",
		ChartText:  "#00AA00",

		CardBg:            "#001100",
		CardBorder:        "#003300",
		CardFocusedBg:     "#002200",
		CardFocusedBorder: "#00FF00",
	},

	// ── dracula — Dracula-inspired ──
	ThemeDracula: {
		Primary:    "#BD93F9",
		Secondary:  "#50FA7B",
		Warning:    "#FFB86C",
		Danger:     "#FF5555",
		Info:       "#8BE9FD",
		Muted:      "#6272A4",
		Text:       "#F8F8F2",
		Background: "#282A36",
		Border:     "#44475A",

		ChartLine1: "#BD93F9",
		ChartLine2: "#50FA7B",
		ChartLine3: "#FF79C6",
		ChartGrid:  "#44475A",
		ChartText:  "#6272A4",

		CardBg:            "#44475A",
		CardBorder:        "#6272A4",
		CardFocusedBg:     "#555577",
		CardFocusedBorder: "#BD93F9",
	},

	// ── nord — Nord-inspired ──
	ThemeNord: {
		Primary:    "#88C0D0",
		Secondary:  "#A3BE8C",
		Warning:    "#EBCB8B",
		Danger:     "#BF616A",
		Info:       "#81A1C1",
		Muted:      "#616E88",
		Text:       "#ECEFF4",
		Background: "#2E3440",
		Border:     "#4C566A",

		ChartLine1: "#88C0D0",
		ChartLine2: "#A3BE8C",
		ChartLine3: "#EBCB8B",
		ChartGrid:  "#434C5E",
		ChartText:  "#616E88",

		CardBg:            "#3B4252",
		CardBorder:        "#4C566A",
		CardFocusedBg:     "#434C5E",
		CardFocusedBorder: "#88C0D0",
	},
}

var activeTheme = ThemeDefault

// ParseThemeName normalizes user supplied theme names.
func ParseThemeName(name string) ThemeName {
	switch ThemeName(strings.ToLower(strings.TrimSpace(name))) {
	case ThemeDefault:
		return ThemeDefault
	case ThemeDark:
		return ThemeDark
	case ThemeLight:
		return ThemeLight
	case ThemeHighContrast, "high_contrast", "contrast":
		return ThemeHighContrast
	case ThemeSquibDark, "squib_dark":
		return ThemeSquibDark
	case ThemeSquibLight, "squib_light":
		return ThemeSquibLight
	case ThemeAmber:
		return ThemeAmber
	case ThemeGreen:
		return ThemeGreen
	case ThemeDracula:
		return ThemeDracula
	case ThemeNord:
		return ThemeNord
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
	for i, name := range allThemeOrder {
		if name == activeTheme {
			nextIdx := (i + 1) % len(allThemeOrder)
			return SetTheme(allThemeOrder[nextIdx])
		}
	}
	return SetTheme(ThemeDefault)
}

// ThemeNames returns the ordered list of available theme names.
func ThemeNames() []ThemeName {
	out := make([]ThemeName, len(allThemeOrder))
	copy(out, allThemeOrder)
	return out
}
