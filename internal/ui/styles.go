package ui

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
)

// Color palette aliases kept for compatibility with existing callers.
const (
	ColorPrimary   = common.ColorPrimary
	ColorSecondary = common.ColorSecondary
	ColorWarning   = common.ColorWarning
	ColorDanger    = common.ColorDanger
	ColorInfo      = common.ColorInfo
	ColorMuted     = common.ColorMuted
	ColorText      = common.ColorText
	ColorDarkBg    = common.ColorDarkBg
	ColorBorder    = common.ColorBorder
)

// Common styles.
var (
	// App styles
	AppStyle lipgloss.Style

	TitleStyle lipgloss.Style

	SubtitleStyle lipgloss.Style

	// Menu styles
	MenuItemStyle lipgloss.Style

	MenuSelectedStyle lipgloss.Style

	MenuKeyStyle lipgloss.Style

	MenuDescStyle lipgloss.Style

	// Status bar
	StatusBarStyle lipgloss.Style

	StatusGoodStyle lipgloss.Style

	StatusWarnStyle lipgloss.Style

	StatusBadStyle lipgloss.Style

	// Panels
	PanelStyle lipgloss.Style

	PanelTitleStyle lipgloss.Style

	// Help
	HelpKeyStyle lipgloss.Style

	HelpDescStyle lipgloss.Style

	// Onboarding
	OnboardingTitleStyle lipgloss.Style

	OnboardingBodyStyle lipgloss.Style

	// Value highlighting
	ValueStyle lipgloss.Style

	LabelStyle lipgloss.Style

	// Card styles
	CardStyle lipgloss.Style

	CardFocusedStyle lipgloss.Style

	CardTitleStyle lipgloss.Style

	CardBodyStyle lipgloss.Style

	DashboardTitleStyle lipgloss.Style

	GridContainerStyle lipgloss.Style

	DashboardSectionStyle lipgloss.Style

	// Divider
	Divider string

	// Chart style helpers
	ChartLineStyle lipgloss.Style
	ChartGridStyle lipgloss.Style
	ChartTextStyle lipgloss.Style
)

func init() {
	ApplyTheme(common.ThemeFromEnv())
}

// ApplyTheme updates UI and shared common styles for a theme.
func ApplyTheme(name common.ThemeName) common.ThemeName {
	applied := common.SetTheme(name)
	rebuildStyles(common.CurrentPalette())
	return applied
}

func rebuildStyles(p common.Palette) {
	AppStyle = lipgloss.NewStyle().
		Padding(1, 2)

	TitleStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(p.Primary)).
		Bold(true).
		Padding(0, 0, 1, 0)

	SubtitleStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(p.Muted)).
		Padding(0, 0, 1, 0)

	MenuItemStyle = lipgloss.NewStyle().
		Padding(0, 1).
		Foreground(lipgloss.Color(p.Text))

	MenuSelectedStyle = lipgloss.NewStyle().
		Padding(0, 1).
		Foreground(lipgloss.Color(p.Secondary)).
		Bold(true)

	MenuKeyStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(p.Primary)).
		Bold(true).
		Width(4)

	MenuDescStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(p.Muted))

	StatusBarStyle = lipgloss.NewStyle().
		Background(lipgloss.Color(p.Background)).
		Foreground(lipgloss.Color(p.Text)).
		Padding(0, 1).
		Width(80)

	StatusGoodStyle = lipgloss.NewStyle().
		Background(lipgloss.Color(p.Secondary)).
		Foreground(lipgloss.Color(p.Text)).
		Padding(0, 1)

	StatusWarnStyle = lipgloss.NewStyle().
		Background(lipgloss.Color(p.Warning)).
		Foreground(lipgloss.Color(p.Text)).
		Padding(0, 1)

	StatusBadStyle = lipgloss.NewStyle().
		Background(lipgloss.Color(p.Danger)).
		Foreground(lipgloss.Color(p.Text)).
		Padding(0, 1)

	PanelStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(p.Border)).
		Padding(1, 2)

	PanelTitleStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(p.Primary)).
		Bold(true).
		Padding(0, 0, 1, 0)

	HelpKeyStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(p.Primary)).
		Bold(true)

	HelpDescStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(p.Muted))

	OnboardingTitleStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(p.Primary)).
		Bold(true).
		Padding(0, 0, 1, 0).
		Width(60).
		Align(lipgloss.Center)

	OnboardingBodyStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(p.Text)).
		Padding(0, 2)

	ValueStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(p.Secondary)).
		Bold(true)

	LabelStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(p.Muted))

	// ── Card styles ──

	CardStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(p.CardBorder)).
		Background(lipgloss.Color(p.CardBg)).
		Padding(0, 1).
		Margin(0, 0, 1, 0)

	CardFocusedStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(p.CardFocusedBorder)).
		Background(lipgloss.Color(p.CardFocusedBg)).
		Padding(0, 1).
		Margin(0, 0, 1, 0)

	CardTitleStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(p.Primary)).
		Bold(true)

	CardBodyStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(p.Text))

	DashboardTitleStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(p.Primary)).
		Bold(true).
		Padding(0, 0, 1, 0)

	GridContainerStyle = lipgloss.NewStyle().
		Padding(0, 0, 1, 0)

	DashboardSectionStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(p.Secondary)).
		Bold(true).
		Padding(0, 0, 1, 0)

	Divider = lipgloss.NewStyle().
		Foreground(lipgloss.Color(p.Border)).
		Render(strings.Repeat("-", 40))

	// Chart style helpers (used by dashboard/sysops views)
	ChartLineStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(p.ChartLine1))

	ChartGridStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(p.ChartGrid))

	ChartTextStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(p.ChartText))
}
