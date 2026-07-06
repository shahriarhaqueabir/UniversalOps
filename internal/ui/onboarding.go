package ui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
)

// OnboardingStep represents a step in the first-run wizard.
type OnboardingStep int

const (
	OnboardWelcome OnboardingStep = iota
	OnboardWhatIs
	OnboardFeatures
	OnboardNavigation
	OnboardComplete
)

// OnboardingWizard guides first-time users through the application.
type OnboardingWizard struct {
	step     OnboardingStep
	complete bool
}

// NewOnboardingWizard creates a new onboarding wizard.
func NewOnboardingWizard() *OnboardingWizard {
	return &OnboardingWizard{
		step:     OnboardWelcome,
		complete: common.IsOnboarded(),
	}
}

// IsComplete returns true if the user has finished onboarding.
func (o *OnboardingWizard) IsComplete() bool {
	return o.complete
}

// Complete marks the wizard as done (for returning users).
func (o *OnboardingWizard) Complete() {
	o.complete = true
	_ = common.MarkOnboarded() // ignore error — non-fatal
}

// Reset restarts the onboarding (for first-run).
func (o *OnboardingWizard) Reset() {
	o.step = OnboardWelcome
	o.complete = false
	_ = common.ClearOnboarded() // ignore error — non-fatal
}

// HandleKey processes keyboard input during onboarding.
func (o *OnboardingWizard) HandleKey(msg tea.KeyPressMsg) bool {
	if o.complete {
		return false
	}

	switch msg.String() {
	case "enter", " ", "space", "n":
		if o.step < OnboardComplete {
			o.step++
		} else {
			o.complete = true
		}
		return true
	case "q", "esc":
		o.complete = true
		return true
	case "b":
		if o.step > OnboardWelcome {
			o.step--
		}
		return true
	}
	return false
}

// Render returns the current onboarding step view.
func (o *OnboardingWizard) Render() string {
	if o.complete {
		return ""
	}

	var b strings.Builder

	switch o.step {
	case OnboardWelcome:
		b.WriteString(OnboardingTitleStyle.Render("👋 Welcome to Hawkward!"))
		b.WriteString("\n\n")
		b.WriteString(OnboardingBodyStyle.Render(
			"Hawkward is a terminal-based operations platform that puts " +
				"powerful system, network, security, and development tools " +
				"at your fingertips.\n\n" +
				"No terminal experience? No problem. Just use the arrow keys " +
				"and follow the prompts.",
		))

	case OnboardWhatIs:
		b.WriteString(OnboardingTitleStyle.Render("🤔 What is Hawkward?"))
		b.WriteString("\n\n")
		b.WriteString(OnboardingBodyStyle.Render(
			"Hawkward combines five operations layers into one keyboard-friendly interface:\n\n" +
				"  🖥  SysOps — Monitor your system (CPU, RAM, disk, processes)\n" +
				"  🌐  NetOps — Diagnose your network (ping, ports, DNS)\n" +
				"  🔒  SecOps — Audit your security (firewall, users, defender)\n" +
				"  ⚙️  DevOps — Run commands, tail logs, browse files\n" +
				"  🤖  AI Ops — Ask an AI assistant about your system",
		))

	case OnboardFeatures:
		b.WriteString(OnboardingTitleStyle.Render("✨ What You Can Do"))
		b.WriteString("\n\n")
		b.WriteString(OnboardingBodyStyle.Render(
			"• Check your system health at a glance with live dashboards\n" +
				"• Ping servers and scan ports without remembering command-line flags\n" +
				"• Review firewall rules and user accounts in seconds\n" +
				"• Run commands and tail logs without switching to another terminal\n" +
				"• Generate detailed reports with one key press\n" +
				"• Everything updates in real-time — no manual refreshes needed",
		))

	case OnboardNavigation:
		b.WriteString(OnboardingTitleStyle.Render("⌨ Navigation at a Glance"))
		b.WriteString("\n\n")
		b.WriteString(OnboardingBodyStyle.Render(
			"  ↑↓ or k/j    Move through menus\n" +
				"  Enter/Space  Select an option\n" +
				"  1-5          Jump directly to any operations layer\n" +
				"  ?            Open help (always available)\n" +
				"  q / Ctrl+C   Quit at any time\n" +
				"  Esc          Go back\n\n" +
				"  [b]ack  [n]ext  [q]uit tutorial",
		))

	case OnboardComplete:
		b.WriteString(OnboardingTitleStyle.Render("✅ You're All Set!"))
		b.WriteString("\n\n")
		b.WriteString(OnboardingBodyStyle.Render(
			"You're ready to start using Hawkward.\n\n" +
				"The main menu will appear next. Press [1] to view your system dashboard, " +
				"or explore any layer using the number keys.\n\n" +
				"Press [?] at any time for help.\n" +
				"Press [Enter] to begin!",
		))
	}

	// Progress indicator
	progress := fmt.Sprintf("\n\n[Step %d/%d]  [b]ack  [n]ext  [q]uit",
		o.step+1, OnboardComplete+1)
	b.WriteString(SubtitleStyle.Render(progress))

	return PanelStyle.Render(b.String())
}
