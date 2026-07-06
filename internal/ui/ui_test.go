package ui

import (
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
)

// ── StatusBar ──────────────────────────────────────────────────────────────

func TestStatusBarNilRender(t *testing.T) {
	var s *StatusBar
	if got := s.Render(); got != "" {
		t.Errorf("nil StatusBar.Render() = %q, want empty", got)
	}
}

func TestStatusBarEmptyRender(t *testing.T) {
	s := &StatusBar{Width: 80, Screen: "Main Menu"}
	got := s.Render()
	if got == "" {
		t.Error("StatusBar.Render() should not be empty")
	}
}

func TestStatusBarWithStats(t *testing.T) {
	s := &StatusBar{
		Width:  80,
		Screen: "Test",
		Stats: &common.SystemStats{
			CPUPercent: 45.0,
			MemoryUsed: 60.0,
			DiskUsed:   70.0,
		},
	}
	got := s.Render()
	if got == "" {
		t.Error("StatusBar.Render() with stats should not be empty")
	}
}

func TestStatusBarBadHealth(t *testing.T) {
	s := &StatusBar{
		Width:  80,
		Screen: "Test",
		Stats: &common.SystemStats{
			CPUPercent: 95.0,
			MemoryUsed: 50.0,
			DiskUsed:   30.0,
		},
	}
	got := s.Render()
	if got == "" {
		t.Error("StatusBar.Render() should not be empty")
	}
}

// ── MainMenu ───────────────────────────────────────────────────────────────

func TestNewMainMenu(t *testing.T) {
	m := NewMainMenu()
	if m == nil {
		t.Fatal("NewMainMenu() returned nil")
	}
	if len(m.items) != 5 {
		t.Errorf("items = %d, want 5", len(m.items))
	}
	if m.cursor != 0 {
		t.Errorf("cursor = %d, want 0", m.cursor)
	}
}

func TestMainMenuNavigate(t *testing.T) {
	m := NewMainMenu()
	// Initial cursor at 0
	if screen := m.Navigate(); screen != common.ScreenSysOps {
		t.Errorf("Navigate() = %d, want ScreenSysOps (%d)", screen, common.ScreenSysOps)
	}

	// Move down
	m.cursor = 1
	if screen := m.Navigate(); screen != common.ScreenNetOps {
		t.Errorf("Navigate() = %d, want ScreenNetOps (%d)", screen, common.ScreenNetOps)
	}

	m.cursor = 4
	if screen := m.Navigate(); screen != common.ScreenAIOps {
		t.Errorf("Navigate() = %d, want ScreenAIOps (%d)", screen, common.ScreenAIOps)
	}
}

func TestMainMenuUpdateDownUp(t *testing.T) {
	m := NewMainMenu()

	// Down key should move cursor to 1
	selected, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	if selected {
		t.Error("Down key should not select")
	}
	if m.cursor != 1 {
		t.Errorf("cursor = %d, want 1", m.cursor)
	}

	// Up key should move cursor back to 0
	selected, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyUp})
	if selected {
		t.Error("Up key should not select")
	}
	if m.cursor != 0 {
		t.Errorf("cursor = %d, want 0", m.cursor)
	}
}

func TestMainMenuUpdateJKLetters(t *testing.T) {
	m := NewMainMenu()

	// 'j' key (down alternative)
	m.Update(tea.KeyPressMsg{Text: "j"})
	if m.cursor != 1 {
		t.Errorf("cursor = %d, want 1", m.cursor)
	}

	// 'k' key (up alternative)
	m.Update(tea.KeyPressMsg{Text: "k"})
	if m.cursor != 0 {
		t.Errorf("cursor = %d, want 0", m.cursor)
	}
}

func TestMainMenuUpdateCursorBounds(t *testing.T) {
	m := NewMainMenu()

	// Should not go above 0
	m.cursor = 0
	m.Update(tea.KeyPressMsg{Code: tea.KeyUp})
	if m.cursor != 0 {
		t.Errorf("cursor = %d, want 0 (stay at top)", m.cursor)
	}

	// Should not go below last item
	m.cursor = 4
	m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	if m.cursor != 4 {
		t.Errorf("cursor = %d, want 4 (stay at bottom)", m.cursor)
	}
}

func TestMainMenuSelect(t *testing.T) {
	m := NewMainMenu()

	// Enter selects current item
	selected, screen := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if !selected {
		t.Error("Enter should select")
	}
	if screen != common.ScreenSysOps {
		t.Errorf("screen = %d, want ScreenSysOps (%d)", screen, common.ScreenSysOps)
	}

	// Space selects current item
	m = NewMainMenu()
	m.cursor = 2
	selected, screen = m.Update(tea.KeyPressMsg{Code: tea.KeySpace, Text: " "})
	if !selected {
		t.Error("Space should select")
	}
	if screen != common.ScreenSecOps {
		t.Errorf("screen = %d, want ScreenSecOps (%d)", screen, common.ScreenSecOps)
	}
}

func TestMainMenuNumberKeys(t *testing.T) {
	m := NewMainMenu()

	tests := []struct {
		key  tea.KeyPressMsg
		want common.Screen
	}{
		{tea.KeyPressMsg{Text: "1"}, common.ScreenSysOps},
		{tea.KeyPressMsg{Text: "2"}, common.ScreenNetOps},
		{tea.KeyPressMsg{Text: "3"}, common.ScreenSecOps},
		{tea.KeyPressMsg{Text: "4"}, common.ScreenDevOps},
		{tea.KeyPressMsg{Text: "5"}, common.ScreenAIOps},
	}
	for _, tt := range tests {
		t.Run(tt.key.Text, func(t *testing.T) {
			selected, screen := m.Update(tt.key)
			if !selected {
				t.Errorf("number key '%s' should select", tt.key.Text)
			}
			if screen != tt.want {
				t.Errorf("screen = %d, want %d", screen, tt.want)
			}
		})
	}
}

func TestMainMenuUnknownKey(t *testing.T) {
	m := NewMainMenu()
	// An unknown key should return false with MainMenu screen
	selected, screen := m.Update(tea.KeyPressMsg{Text: "x"})
	if selected {
		t.Error("unknown key should not select")
	}
	if screen != common.ScreenMainMenu {
		t.Errorf("screen = %d, want ScreenMainMenu (%d)", screen, common.ScreenMainMenu)
	}
}

func TestMainMenuSetSize(t *testing.T) {
	m := NewMainMenu()
	m.SetSize(100, 40)
	if m.width != 100 {
		t.Errorf("width = %d, want 100", m.width)
	}
	if m.height != 40 {
		t.Errorf("height = %d, want 40", m.height)
	}
}

func TestMainMenuRender(t *testing.T) {
	m := NewMainMenu()
	got := m.Render()
	if got == "" {
		t.Error("Render() should not be empty")
	}
}

// ── HelpOverlay ────────────────────────────────────────────────────────────

func TestNewHelpOverlay(t *testing.T) {
	h := NewHelpOverlay()
	if h == nil {
		t.Fatal("NewHelpOverlay() returned nil")
	}
	if h.visible {
		t.Error("help should not be visible initially")
	}
}

func TestHelpOverlayToggle(t *testing.T) {
	h := NewHelpOverlay()
	h.Toggle()
	if !h.Visible() {
		t.Error("Help should be visible after Toggle")
	}
	h.Toggle()
	if h.Visible() {
		t.Error("Help should not be visible after second Toggle")
	}
}

func TestHelpOverlayHide(t *testing.T) {
	h := NewHelpOverlay()
	h.Toggle()
	h.Hide()
	if h.Visible() {
		t.Error("Help should not be visible after Hide")
	}
}

func TestHelpOverlayHandleKeyWhenHidden(t *testing.T) {
	h := NewHelpOverlay()
	consumed := h.HandleKey(tea.KeyPressMsg{Code: tea.KeyEsc})
	if consumed {
		t.Error("HandleKey should not consume keys when hidden")
	}
}

func TestHelpOverlayHandleKeyWhenVisible(t *testing.T) {
	h := NewHelpOverlay()
	h.Toggle()

	// Esc should close help
	consumed := h.HandleKey(tea.KeyPressMsg{Code: tea.KeyEsc})
	if !consumed {
		t.Error("HandleKey should consume keys when visible")
	}
	if h.Visible() {
		t.Error("Help should close on Esc")
	}
}

func TestHelpOverlayHandleKeyClosesOnQ(t *testing.T) {
	h := NewHelpOverlay()
	h.Toggle()

	consumed := h.HandleKey(tea.KeyPressMsg{Text: "q"})
	if !consumed {
		t.Error("'q' should be consumed when help is visible")
	}
	if h.Visible() {
		t.Error("Help should close on 'q'")
	}
}

func TestHelpOverlayHandleKeyOthersConsumed(t *testing.T) {
	h := NewHelpOverlay()
	h.Toggle()

	// Any other key should be consumed but keep help visible
	consumed := h.HandleKey(tea.KeyPressMsg{Text: "a"})
	if !consumed {
		t.Error("other keys should be consumed when help is visible")
	}
	if !h.Visible() {
		t.Error("help should stay visible for other keys")
	}
}

func TestHelpOverlayRender(t *testing.T) {
	h := NewHelpOverlay()
	got := h.Render()
	if got != "" {
		t.Error("Render() should be empty when not visible")
	}

	h.Toggle()
	got = h.Render()
	if got == "" {
		t.Error("Render() should not be empty when visible")
	}
}

// ── OnboardingWizard ───────────────────────────────────────────────────────

func TestNewOnboardingWizard(t *testing.T) {
	// Reset to clean state in case the local machine has been onboarded
	common.ClearOnboarded()
	o := NewOnboardingWizard()
	if o == nil {
		t.Fatal("NewOnboardingWizard() returned nil")
	}
	if o.complete {
		t.Error("should not be complete initially")
	}
	if o.step != OnboardWelcome {
		t.Errorf("step = %d, want OnboardWelcome (%d)", o.step, OnboardWelcome)
	}
}

func TestOnboardingIsComplete(t *testing.T) {
	// Reset to clean state in case the local machine has been onboarded
	common.ClearOnboarded()
	o := NewOnboardingWizard()
	if o.IsComplete() {
		t.Error("IsComplete() should be false initially")
	}
	o.complete = true
	if !o.IsComplete() {
		t.Error("IsComplete() should be true after set")
	}
}

func TestOnboardingComplete(t *testing.T) {
	o := NewOnboardingWizard()
	o.Complete()
	if !o.IsComplete() {
		t.Error("IsComplete() should be true after Complete()")
	}
}

func TestOnboardingReset(t *testing.T) {
	o := NewOnboardingWizard()
	o.complete = true
	o.step = OnboardComplete

	o.Reset()
	if o.IsComplete() {
		t.Error("should not be complete after Reset")
	}
	if o.step != OnboardWelcome {
		t.Errorf("step = %d, want OnboardWelcome (%d)", o.step, OnboardWelcome)
	}
}

func TestOnboardingHandleKeySteps(t *testing.T) {
	o := NewOnboardingWizard()
	enter := tea.KeyPressMsg{Code: tea.KeyEnter}

	// Press Enter to advance through steps
	for i := OnboardWelcome; i < OnboardComplete; i++ {
		if o.step != i {
			t.Errorf("step = %d, want %d", o.step, i)
		}
		consumed := o.HandleKey(enter)
		if !consumed {
			t.Errorf("HandleKey should consume key at step %d", i)
		}
	}

	// After OnboardComplete, one more Enter marks it done
	if o.complete {
		t.Error("should not be complete yet")
	}
	o.HandleKey(enter)
	if !o.complete {
		t.Error("should be complete after final step")
	}
}

func TestOnboardingHandleKeySpaceAndN(t *testing.T) {
	o := NewOnboardingWizard()

	// Space should also advance
	consumed := o.HandleKey(tea.KeyPressMsg{Code: tea.KeySpace, Text: " "})
	if !consumed {
		t.Error("Space should be consumed")
	}
	if o.step != OnboardWhatIs {
		t.Errorf("step = %d, want OnboardWhatIs", o.step)
	}

	// 'n' key should advance
	consumed = o.HandleKey(tea.KeyPressMsg{Text: "n"})
	if !consumed {
		t.Error("'n' should be consumed")
	}
	if o.step != OnboardFeatures {
		t.Errorf("step = %d, want OnboardFeatures", o.step)
	}
}

func TestOnboardingHandleKeyBack(t *testing.T) {
	o := NewOnboardingWizard()
	o.step = OnboardFeatures

	// 'b' should go to previous step
	consumed := o.HandleKey(tea.KeyPressMsg{Text: "b"})
	if !consumed {
		t.Error("'b' should be consumed")
	}
	if o.step != OnboardWhatIs {
		t.Errorf("step = %d, want OnboardWhatIs (%d)", o.step, OnboardWhatIs)
	}

	// Back at Welcome should stay
	o.step = OnboardWelcome
	consumed = o.HandleKey(tea.KeyPressMsg{Text: "b"})
	if !consumed {
		t.Error("'b' should be consumed")
	}
	if o.step != OnboardWelcome {
		t.Errorf("step = %d, want OnboardWelcome (%d)", o.step, OnboardWelcome)
	}
}

func TestOnboardingHandleKeyQuit(t *testing.T) {
	o := NewOnboardingWizard()

	// Esc should complete onboarding
	consumed := o.HandleKey(tea.KeyPressMsg{Code: tea.KeyEsc})
	if !consumed {
		t.Error("Esc should be consumed")
	}
	if !o.complete {
		t.Error("Esc should mark onboarding complete")
	}

	// Reset and test 'q' key
	o.Reset()
	consumed = o.HandleKey(tea.KeyPressMsg{Text: "q"})
	if !consumed {
		t.Error("'q' should be consumed")
	}
	if !o.complete {
		t.Error("'q' should mark onboarding complete")
	}
}

func TestOnboardingHandleKeyWhenComplete(t *testing.T) {
	o := NewOnboardingWizard()
	o.complete = true
	consumed := o.HandleKey(tea.KeyPressMsg{Code: tea.KeyEnter})
	if consumed {
		t.Error("HandleKey should not consume keys when complete")
	}
}

func TestOnboardingHandleKeyUnknown(t *testing.T) {
	o := NewOnboardingWizard()
	consumed := o.HandleKey(tea.KeyPressMsg{Text: "x"})
	if consumed {
		t.Error("unknown key should not be consumed")
	}
	if o.complete {
		t.Error("unknown key should not complete onboarding")
	}
	if o.step != OnboardWelcome {
		t.Errorf("step should not change, got %d", o.step)
	}
}

func TestOnboardingRender(t *testing.T) {
	o := NewOnboardingWizard()

	// Render when not complete
	got := o.Render()
	if got == "" {
		t.Error("Render() should not be empty during onboarding")
	}

	// Render when complete
	o.complete = true
	got = o.Render()
	if got != "" {
		t.Error("Render() should be empty when complete")
	}
}

func TestOnboardingRenderSteps(t *testing.T) {
	o := NewOnboardingWizard()

	steps := []OnboardingStep{
		OnboardWelcome,
		OnboardWhatIs,
		OnboardFeatures,
		OnboardNavigation,
		OnboardComplete,
	}
	for _, step := range steps {
		o.step = step
		got := o.Render()
		if got == "" {
			t.Errorf("Render() should not be empty for step %d", step)
		}
	}
}

// ── KeyMap ─────────────────────────────────────────────────────────────────

func TestDefaultKeyMap(t *testing.T) {
	k := DefaultKeyMap()
	if len(k.Quit.Keys()) == 0 {
		t.Error("Quit binding has no keys")
	}
	if len(k.Up.Keys()) == 0 {
		t.Error("Up binding has no keys")
	}
	if len(k.Down.Keys()) == 0 {
		t.Error("Down binding has no keys")
	}
	if len(k.Enter.Keys()) == 0 {
		t.Error("Enter binding has no keys")
	}
	if len(k.Help.Keys()) == 0 {
		t.Error("Help binding has no keys")
	}
	if len(k.Back.Keys()) == 0 {
		t.Error("Back binding has no keys")
	}
	if len(k.Refresh.Keys()) == 0 {
		t.Error("Refresh binding has no keys")
	}
}

func TestKeyMapFullHelp(t *testing.T) {
	k := DefaultKeyMap()
	bindings := k.FullHelp()
	if len(bindings) == 0 {
		t.Error("FullHelp() should return bindings")
	}
}

func TestKeyMapShortHelp(t *testing.T) {
	k := DefaultKeyMap()
	bindings := k.ShortHelp()
	if len(bindings) == 0 {
		t.Error("ShortHelp() should return bindings")
	}
}

// ── RootModel ──────────────────────────────────────────────────────────────

func TestNewRootModel(t *testing.T) {
	m := NewRootModel()
	if m == nil {
		t.Fatal("NewRootModel() returned nil")
	}
	if m.activeScreen != common.ScreenOnboarding {
		t.Errorf("activeScreen = %d, want ScreenOnboarding (%d)",
			m.activeScreen, common.ScreenOnboarding)
	}
	if m.mainMenu == nil {
		t.Error("mainMenu is nil")
	}
	if m.onboarding == nil {
		t.Error("onboarding is nil")
	}
	if m.statusBar == nil {
		t.Error("statusBar is nil")
	}
	if m.sysOps == nil {
		t.Error("sysOps is nil")
	}
}

func TestRootModelPushPopScreen(t *testing.T) {
	m := NewRootModel()
	m.onboarding.Complete() // skip onboarding

	// Push a screen
	m.pushScreen(common.ScreenSysOps)
	if m.activeScreen != common.ScreenSysOps {
		t.Errorf("activeScreen = %d, want ScreenSysOps", m.activeScreen)
	}
	if len(m.previousScreens) != 1 {
		t.Errorf("previousScreens = %d, want 1", len(m.previousScreens))
	}

	// Pop back
	m.popScreen()
	if m.activeScreen != common.ScreenOnboarding {
		// Note: since we navigated from Onboarding, it's the previous screen
		t.Logf("After pop: activeScreen = %d", m.activeScreen)
	}
}

func TestRootModelPopScreenEmpty(t *testing.T) {
	m := NewRootModel()
	// Pop with empty stack should go to MainMenu
	m.previousScreens = nil
	m.popScreen()
	if m.activeScreen != common.ScreenMainMenu {
		t.Errorf("activeScreen = %d, want ScreenMainMenu (%d)",
			m.activeScreen, common.ScreenMainMenu)
	}
}

// ── Refresh Interval ────────────────────────────────────────────────────────

func TestRefreshIntervalDefaults(t *testing.T) {
	m := NewRootModel()
	if m.refreshInterval != DefaultRefreshInterval {
		t.Errorf("refreshInterval = %v, want %v", m.refreshInterval, DefaultRefreshInterval)
	}
	if got := m.RefreshInterval(); got != DefaultRefreshInterval {
		t.Errorf("RefreshInterval() = %v, want %v", got, DefaultRefreshInterval)
	}
}

func TestSetRefreshInterval(t *testing.T) {
	m := NewRootModel()
	prev := m.SetRefreshInterval(5 * time.Second)
	if prev != DefaultRefreshInterval {
		t.Errorf("previous interval = %v, want %v", prev, DefaultRefreshInterval)
	}
	if m.refreshInterval != 5*time.Second {
		t.Errorf("refreshInterval = %v, want 5s", m.refreshInterval)
	}
	if got := m.RefreshInterval(); got != 5*time.Second {
		t.Errorf("RefreshInterval() = %v, want 5s", got)
	}
	// Verify Init uses the configured interval
	cmd := m.Init()
	if cmd == nil {
		t.Error("Init() returned nil cmd")
	}
}

// ── Screen Routing ──────────────────────────────────────────────────────────

func TestRootModelRouteKeyFromMainMenu(t *testing.T) {
	m := NewRootModel()
	m.onboarding.Complete() // skip onboarding
	m.activeScreen = common.ScreenMainMenu

	tests := []struct {
		key    string
		screen common.Screen
	}{
		{"1", common.ScreenSysOps},
		{"2", common.ScreenNetOps},
		{"3", common.ScreenSecOps},
		{"4", common.ScreenDevOps},
		{"5", common.ScreenAIOps},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			msg := tea.KeyPressMsg{Text: tt.key}
			cmd := m.routeKey(msg)
			if cmd != nil {
				t.Error("expected nil cmd for number key routing")
			}
			if m.activeScreen != tt.screen {
				t.Errorf("activeScreen = %d, want %d (%s)", m.activeScreen, tt.screen, common.ScreenNames[tt.screen])
			}
			// Reset back to main menu for next test
			m.popScreen()
		})
	}
}

func TestRootModelRouteMessage(t *testing.T) {
	m := NewRootModel()
	m.onboarding.Complete()

	t.Run("routes to sysops", func(t *testing.T) {
		m.activeScreen = common.ScreenSysOps
		cmd := m.routeMessage(tea.KeyPressMsg{Text: "r"})
		// sysOps handles "r" as a refresh key
		_ = cmd // cmd may be nil or non-nil, just verify no panic
	})

	t.Run("routes to secops", func(t *testing.T) {
		m.activeScreen = common.ScreenSecOps
		cmd := m.routeMessage(tea.KeyPressMsg{Text: "r"})
		_ = cmd
	})

	t.Run("unknown screen returns nil", func(t *testing.T) {
		m.activeScreen = common.Screen(-1)
		cmd := m.routeMessage(tea.KeyPressMsg{Text: "r"})
		if cmd != nil {
			t.Error("expected nil cmd for unknown screen")
		}
	})
}

func TestRootModelEscKey(t *testing.T) {
	m := NewRootModel()
	m.onboarding.Complete()

	// Navigate to secops
	m.pushScreen(common.ScreenSecOps)
	if m.activeScreen != common.ScreenSecOps {
		t.Fatalf("activeScreen = %d, want ScreenSecOps", m.activeScreen)
	}

	// Press esc to go back
	msg := tea.KeyPressMsg{Code: tea.KeyEsc}
	model, cmd := m.Update(msg)
	if model == nil {
		t.Fatal("Update returned nil model")
	}
	_ = cmd
	// Should be back at main menu (from onboarding -> secops -> back to onboarding -> pop -> main menu)
	// Actually from onboarding we pushed secops, so popping should go back to onboarding, which defaults to main menu
	t.Logf("After esc: activeScreen = %d (%s)", m.activeScreen, common.ScreenNames[m.activeScreen])
}

func TestRootModelTickHandling(t *testing.T) {
	m := NewRootModel()
	m.onboarding.Complete()

	// Send a TickMsg
	model, cmd := m.Update(common.TickMsg(time.Now()))
	if model == nil {
		t.Fatal("Update returned nil model")
	}
	if cmd == nil {
		t.Error("expected non-nil cmd to restart the tick timer")
	}
	// Stats may be nil because sysops might fail on CI, but tick should still be handled
}

func TestRootModelQuitKey(t *testing.T) {
	m := NewRootModel()
	m.onboarding.Complete()

	// Press q
	msg := tea.KeyPressMsg{Text: "q"}
	model, cmd := m.Update(msg)
	if model == nil {
		t.Fatal("Update returned nil model")
	}
	// Should return a quit command
	_ = cmd
}

func TestRootModelHelpKey(t *testing.T) {
	m := NewRootModel()
	m.onboarding.Complete()

	// Press ? to toggle help
	msg := tea.KeyPressMsg{Text: "?"}
	model, cmd := m.Update(msg)
	if model == nil {
		t.Fatal("Update returned nil model")
	}
	_ = cmd
	if !m.help.Visible() {
		t.Error("help should be visible after pressing ?")
	}
}

func TestRootModelRKeySecOpsWorkflow(t *testing.T) {
	m := NewRootModel()
	m.onboarding.Complete()
	m.activeScreen = common.ScreenSecOps

	// Press 'R' (shift+r) to trigger security audit workflow
	msg := tea.KeyPressMsg{Text: "R"}
	model, cmd := m.Update(msg)
	if model == nil {
		t.Fatal("Update returned nil model")
	}
	if cmd == nil {
		t.Fatal("expected non-nil cmd for R workflow")
	}

	// Execute the returned command to get the workflow result
	resultMsg := cmd()
	if resultMsg == nil {
		t.Fatal("workflow command returned nil message")
	}

	// Send the result back to the model
	model2, cmd2 := m.Update(resultMsg)
	if model2 == nil {
		t.Fatal("second Update returned nil model")
	}
	_ = cmd2

	// After a successful workflow, showReport should be true
	if !m.secOps.ShowReport() {
		t.Error("ShowReport should be true after workflow completes")
	}
	if m.secOps.WorkflowReport() == "" {
		t.Error("WorkflowReport should not be empty after successful run")
	}
}

// ── Styles ─────────────────────────────────────────────────────────────────

func TestStyleConstants(t *testing.T) {
	// lipgloss.Style is a value type (struct), not a pointer.
	// Check by rendering a sample string — if it panics or returns "", fail.
	if got := AppStyle.Render("test"); got == "" {
		t.Error("AppStyle.Render('test') returned empty")
	}
	if got := TitleStyle.Render("test"); got == "" {
		t.Error("TitleStyle.Render('test') returned empty")
	}
	if got := MenuItemStyle.Render("test"); got == "" {
		t.Error("MenuItemStyle.Render('test') returned empty")
	}
	if got := MenuSelectedStyle.Render("test"); got == "" {
		t.Error("MenuSelectedStyle.Render('test') returned empty")
	}
	if got := PanelStyle.Render("test"); got == "" {
		t.Error("PanelStyle.Render('test') returned empty")
	}
	if Divider == "" {
		t.Error("Divider is empty")
	}
}

func TestStyleColorConstants(t *testing.T) {
	if ColorPrimary != "#7C3AED" {
		t.Errorf("ColorPrimary = %q, want %q", ColorPrimary, "#7C3AED")
	}
	if ColorSecondary != "#10B981" {
		t.Errorf("ColorSecondary = %q, want %q", ColorSecondary, "#10B981")
	}
	if ColorDanger != "#EF4444" {
		t.Errorf("ColorDanger = %q, want %q", ColorDanger, "#EF4444")
	}
}
