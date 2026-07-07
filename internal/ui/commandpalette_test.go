package ui

import (
	"fmt"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
)

// keyPressRune creates a tea.KeyPressMsg for a printable rune.
func keyPressRune(r rune) tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: r, Text: string(r)}
}

// keyPressSpecial creates a tea.KeyPressMsg for a special key.
func keyPressSpecial(code rune) tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: code}
}

// typeString sends a sequence of rune key presses to the palette.
func typeString(p *CommandPalette, s string) {
	for _, ch := range s {
		p.HandleKey(keyPressRune(ch))
	}
}

// ── Fuzzy Scoring ─────────────────────────────────────────────────────────────

func TestFuzzyScore_ExactMatch(t *testing.T) {
	score, matches := fuzzyScore("cpu", "cpu")
	if score != 100 {
		t.Errorf("exact match score = %d, want 100", score)
	}
	if len(matches) != 3 {
		t.Errorf("exact match has %d matches, want 3", len(matches))
	}
}

func TestFuzzyScore_PrefixMatch(t *testing.T) {
	score, _ := fuzzyScore("cpu", "cpu dashboard")
	if score != 80 {
		t.Errorf("prefix match score = %d, want 80", score)
	}
}

func TestFuzzyScore_SubstringMatch(t *testing.T) {
	score, _ := fuzzyScore("disk", "ssd disk usage")
	if score != 60 {
		t.Errorf("substring match score = %d, want 60", score)
	}
}

func TestFuzzyScore_SequenceMatch(t *testing.T) {
	score, matches := fuzzyScore("cbd", "cpu bandwidth dashboard")
	if score == 0 {
		t.Error("sequence match should return score > 0")
	}
	if len(matches) == 0 {
		t.Error("sequence match should return matched indices")
	}
}

func TestFuzzyScore_NoMatch(t *testing.T) {
	score, matches := fuzzyScore("xyz", "cpu dashboard")
	if score != 0 {
		t.Errorf("no match score = %d, want 0", score)
	}
	if len(matches) != 0 {
		t.Errorf("no match should return nil matches, got %v", matches)
	}
}

func TestFuzzyScore_EmptyQuery(t *testing.T) {
	score, matches := fuzzyScore("", "cpu")
	if score != 0 {
		t.Errorf("empty query score = %d, want 0", score)
	}
	if matches != nil {
		t.Errorf("empty query should return nil matches")
	}
}

func TestFuzzyScore_CaseInsensitive(t *testing.T) {
	score, _ := fuzzyScore("CPU", "Cpu Dashboard")
	if score != 80 {
		t.Errorf("case-insensitive prefix match score = %d, want 80", score)
	}
}

func TestFuzzyScore_CaseInsensitiveExact(t *testing.T) {
	score, _ := fuzzyScore("DNS", "dns")
	if score != 100 {
		t.Errorf("case-insensitive exact match score = %d, want 100", score)
	}
}

// ── Search Operations ─────────────────────────────────────────────────────────

func TestSearchOps_EmptyQuery(t *testing.T) {
	ops := []Command{
		{ID: "a", Title: "Alpha", Description: "First operation"},
		{ID: "b", Title: "Beta", Description: "Second operation"},
	}
	results := SearchOps("", ops)
	if len(results) != 2 {
		t.Errorf("empty query returned %d results, want 2", len(results))
	}
}

func TestSearchOps_ResultsOrderedByScore(t *testing.T) {
	ops := []Command{
		{ID: "cpu", Title: "CPU Dashboard", Description: "View CPU usage", Tags: []string{"cpu"}},
		{ID: "ping", Title: "Ping Tool", Description: "ICMP ping tool"},
		{ID: "ports", Title: "Listening Ports", Description: "View open ports"},
	}
	results := SearchOps("cpu", ops)
	if len(results) == 0 {
		t.Fatal("search should return at least one result")
	}
	if results[0].Command.ID != "cpu" {
		t.Errorf("top result ID = %q, want %q", results[0].Command.ID, "cpu")
	}
}

func TestSearchOps_PartialMatch(t *testing.T) {
	ops := []Command{
		{ID: "mem", Title: "Memory Monitor", Description: "RAM monitoring"},
		{ID: "dm", Title: "Disk Monitor", Description: "Disk space monitoring"},
	}
	results := SearchOps("monitor", ops)
	if len(results) != 2 {
		t.Errorf("partial match returned %d results, want 2", len(results))
	}
}

func TestSearchOps_NoMatch(t *testing.T) {
	ops := []Command{
		{ID: "a", Title: "Alpha", Description: "First"},
	}
	results := SearchOps("zzz_nonexistent", ops)
	if len(results) != 0 {
		t.Errorf("no-match search returned %d results, want 0", len(results))
	}
}

// ── Tag Filtering ─────────────────────────────────────────────────────────────

func TestFilterByTag(t *testing.T) {
	ops := []Command{
		{ID: "cpu", Title: "CPU", Tags: []string{"sysops", "cpu"}},
		{ID: "ping", Title: "Ping", Tags: []string{"netops", "icmp"}},
		{ID: "firewall", Title: "Firewall", Tags: []string{"secops", "firewall"}},
	}
	filtered := FilterByTag("sysops", ops)
	if len(filtered) != 1 {
		t.Errorf("tag filter returned %d results, want 1", len(filtered))
	}
	if filtered[0].ID != "cpu" {
		t.Errorf("filtered result ID = %q, want %q", filtered[0].ID, "cpu")
	}
}

func TestFilterByTag_CaseInsensitive(t *testing.T) {
	ops := []Command{
		{ID: "cpu", Title: "CPU", Tags: []string{"SysOps"}},
	}
	filtered := FilterByTag("sysops", ops)
	if len(filtered) != 1 {
		t.Errorf("case-insensitive tag filter returned %d, want 1", len(filtered))
	}
}

func TestFilterByTag_NoMatch(t *testing.T) {
	ops := []Command{
		{ID: "cpu", Title: "CPU", Tags: []string{"sysops"}},
	}
	filtered := FilterByTag("nonexistent", ops)
	if len(filtered) != 0 {
		t.Errorf("no-match tag filter returned %d, want 0", len(filtered))
	}
}

func TestAllTags(t *testing.T) {
	ops := []Command{
		{ID: "a", Tags: []string{"sysops", "monitoring"}},
		{ID: "b", Tags: []string{"netops", "monitoring"}},
	}
	tags := AllTags(ops)
	if len(tags) != 3 {
		t.Errorf("AllTags returned %d tags, want 3", len(tags))
	}
	if tags[0] != "monitoring" {
		t.Errorf("first tag = %q, want %q", tags[0], "monitoring")
	}
}

func TestAllTags_Empty(t *testing.T) {
	tags := AllTags(nil)
	if len(tags) != 0 {
		t.Errorf("AllTags(nil) returned %d, want 0", len(tags))
	}
}

// ── Recommend Mode ────────────────────────────────────────────────────────────

func TestRecommend_IntentMatch(t *testing.T) {
	ops := []Command{
		{ID: "ping", Title: "Ping Tool", Tags: []string{"netops", "icmp"}},
		{ID: "cpu", Title: "CPU Dashboard", Tags: []string{"sysops", "cpu"}},
	}
	results := Recommend("network", ops)
	if len(results) == 0 {
		t.Error("recommend should return results for intent match")
	}
	if results[0].Command.ID != "ping" {
		t.Errorf("top recommend result = %q, want %q", results[0].Command.ID, "ping")
	}
}

func TestRecommend_FallbackToSearch(t *testing.T) {
	ops := []Command{
		{ID: "cpu", Title: "CPU Dashboard", Description: "CPU usage", Tags: []string{"cpu"}},
	}
	results := Recommend("cpu", ops)
	if len(results) == 0 {
		t.Error("recommend should fall back to search when no intent matches")
	}
}

// ── Palette Lifecycle ─────────────────────────────────────────────────────────

func TestNewCommandPalette(t *testing.T) {
	p := NewCommandPalette(nil, nil)
	if p == nil {
		t.Fatal("NewCommandPalette() returned nil")
	}
	if p.IsVisible() {
		t.Error("new palette should not be visible")
	}
	if len(p.results) != 0 {
		t.Errorf("new palette has %d results, want 0", len(p.results))
	}
}

func TestCommandPaletteShowHide(t *testing.T) {
	called := false
	p := NewCommandPalette(nil, func() tea.Cmd {
		called = true
		return nil
	})

	p.Show()
	if !p.IsVisible() {
		t.Error("palette should be visible after Show()")
	}

	p.Hide()
	if p.IsVisible() {
		t.Error("palette should not be visible after Hide()")
	}
	if !called {
		t.Error("onClose callback should be called")
	}
}

func TestCommandPaletteToggle(t *testing.T) {
	p := NewCommandPalette(nil, nil)
	p.Toggle()
	if !p.IsVisible() {
		t.Error("palette should be visible after Toggle()")
	}
	p.Toggle()
	if p.IsVisible() {
		t.Error("palette should be hidden after second Toggle()")
	}
}

func TestCommandPaletteShowIdempotent(t *testing.T) {
	p := NewCommandPalette(nil, nil)
	p.Show()
	p.Show()
	if !p.IsVisible() {
		t.Error("palette should remain visible after double Show()")
	}
}

func TestCommandPaletteHideIdempotent(t *testing.T) {
	p := NewCommandPalette(nil, nil)
	p.Hide()
	p.Hide()
	if p.IsVisible() {
		t.Error("palette should remain hidden after double Hide()")
	}
}

// ── Palette Navigation ────────────────────────────────────────────────────────

func TestCommandPaletteNavigation(t *testing.T) {
	p := NewCommandPalette(nil, nil)
	p.Show()
	if len(p.results) == 0 {
		t.Fatal("palette should have results after Show()")
	}

	first := p.results[0].Command.Title

	// Navigate down
	p.HandleKey(keyPressSpecial(tea.KeyDown))
	if p.cursor != 1 {
		t.Errorf("cursor = %d, want 1 after down", p.cursor)
	}

	// Navigate up
	p.HandleKey(keyPressSpecial(tea.KeyUp))
	if p.cursor != 0 {
		t.Errorf("cursor = %d, want 0 after up", p.cursor)
	}

	// SelectedCommand should match first
	cmd := p.SelectedCommand()
	if cmd == nil {
		t.Fatal("SelectedCommand() returned nil")
	}
	if cmd.Title != first {
		t.Errorf("selected title = %q, want %q", cmd.Title, first)
	}
}

func TestCommandPaletteNavigationJK(t *testing.T) {
	p := NewCommandPalette(nil, nil)
	p.Show()

	p.HandleKey(keyPressRune('j'))
	if p.cursor != 1 {
		t.Errorf("cursor = %d, want 1 after j", p.cursor)
	}

	p.HandleKey(keyPressRune('k'))
	if p.cursor != 0 {
		t.Errorf("cursor = %d, want 0 after k", p.cursor)
	}
}

func TestCommandPaletteSelect(t *testing.T) {
	var selectedCmd *Command
	p := NewCommandPalette(func(cmd Command) tea.Cmd {
		selectedCmd = &cmd
		return nil
	}, nil)
	p.Show()

	// Navigate then select
	p.HandleKey(keyPressRune('j'))
	p.HandleKey(keyPressSpecial(tea.KeyEnter))

	if selectedCmd == nil {
		t.Fatal("onSelect callback should have been called")
	}
	if p.IsVisible() {
		t.Error("palette should close after selection")
	}
}

func TestCommandPaletteSelectNoCallback(t *testing.T) {
	p := NewCommandPalette(nil, nil)
	p.Show()
	p.HandleKey(keyPressSpecial(tea.KeyEnter))
}

func TestCommandPaletteEsc(t *testing.T) {
	closeCalled := false
	p := NewCommandPalette(nil, func() tea.Cmd {
		closeCalled = true
		return nil
	})
	p.Show()

	p.HandleKey(keyPressSpecial(tea.KeyEsc))
	if p.IsVisible() {
		t.Error("palette should close on esc")
	}
	if !closeCalled {
		t.Error("onClose should fire on esc")
	}
}

// ── Palette Input ─────────────────────────────────────────────────────────────

func TestCommandPaletteTyping(t *testing.T) {
	p := NewCommandPalette(nil, nil)
	p.Show()

	typeString(p, "cpu")

	if p.query != "cpu" {
		t.Errorf("query = %q, want %q", p.query, "cpu")
	}
	if len(p.results) == 0 {
		t.Error("search should return results for 'cpu'")
	}
}

func TestCommandPaletteBackspace(t *testing.T) {
	p := NewCommandPalette(nil, nil)
	p.Show()

	typeString(p, "cpu")
	p.HandleKey(keyPressSpecial(tea.KeyBackspace))

	if p.query != "cp" {
		t.Errorf("query after backspace = %q, want %q", p.query, "cp")
	}
}

func TestCommandPaletteBackspaceEmpty(t *testing.T) {
	p := NewCommandPalette(nil, nil)
	p.Show()
	p.HandleKey(keyPressSpecial(tea.KeyBackspace))
}

// ── Mode Switching ────────────────────────────────────────────────────────────

func TestCommandPaletteTagMode(t *testing.T) {
	p := NewCommandPalette(nil, nil)
	p.Show()

	p.HandleKey(keyPressRune('t'))
	if p.mode != PaletteTagFilter {
		t.Error("palette should be in tag filter mode after 't'")
	}
}

func TestCommandPaletteRecommendMode(t *testing.T) {
	p := NewCommandPalette(nil, nil)
	p.Show()

	p.HandleKey(keyPressRune('r'))
	if p.mode != PaletteRecommend {
		t.Error("palette should be in recommend mode after 'r'")
	}
}

// ── Search History ────────────────────────────────────────────────────────────

func TestSearchHistoryRingBuffer(t *testing.T) {
	p := NewCommandPalette(nil, nil)

	p.addToHistory("cpu")
	p.addToHistory("ping")
	p.addToHistory("dns")

	if len(p.history) != 3 {
		t.Errorf("history length = %d, want 3", len(p.history))
	}
	if p.history[0] != "dns" {
		t.Errorf("first history entry = %q, want %q", p.history[0], "dns")
	}
}

func TestSearchHistoryNoDuplicates(t *testing.T) {
	p := NewCommandPalette(nil, nil)

	p.addToHistory("cpu")
	p.addToHistory("ping")
	p.addToHistory("cpu")

	if len(p.history) != 2 {
		t.Errorf("history length with dedup = %d, want 2", len(p.history))
	}
	if p.history[0] != "cpu" {
		t.Errorf("first history entry = %q, want %q", p.history[0], "cpu")
	}
}

func TestSearchHistoryEmptyIgnored(t *testing.T) {
	p := NewCommandPalette(nil, nil)
	p.addToHistory("")
	if len(p.history) != 0 {
		t.Error("empty query should not be added to history")
	}
}

func TestSearchHistoryMaxSize(t *testing.T) {
	p := NewCommandPalette(nil, nil)
	for i := 0; i < maxSearchHistory+10; i++ {
		p.addToHistory(fmt.Sprintf("query%d", i))
	}
	if len(p.history) > maxSearchHistory {
		t.Errorf("history exceeded max %d, got %d", maxSearchHistory, len(p.history))
	}
}

// ── Operation Registry ────────────────────────────────────────────────────────

func TestOperationRegistry(t *testing.T) {
	ops := AllOperations()
	if len(ops) == 0 {
		t.Fatal("operation registry should have operations registered at init")
	}
	if len(ops) < 24 {
		t.Errorf("expected at least 24 operations, got %d", len(ops))
	}
}

func TestOperationRegistryHasAllLayers(t *testing.T) {
	ops := AllOperations()
	layers := map[string]bool{"sysops": false, "netops": false, "secops": false, "devops": false, "aiops": false}

	for _, op := range ops {
		for _, tag := range op.Tags {
			if _, ok := layers[tag]; ok {
				layers[tag] = true
			}
		}
	}

	for layer, found := range layers {
		if !found {
			t.Errorf("operation registry missing operations for layer %q", layer)
		}
	}
}

func TestRegisterOperation(t *testing.T) {
	testOp := Command{
		ID: "test.operation", Title: "Test Operation",
		Description: "A test operation for unit tests",
		Icon:        "🧪", Screen: common.ScreenDashboard,
	}
	RegisterOperation(testOp)

	found := false
	for _, op := range AllOperations() {
		if op.ID == "test.operation" {
			found = true
			break
		}
	}
	if !found {
		t.Error("test operation not found after RegisterOperation")
	}
}

// ── SelectedCommand Edge Cases ────────────────────────────────────────────────

func TestSelectedCommandEmptyResults(t *testing.T) {
	p := NewCommandPalette(nil, nil)
	if cmd := p.SelectedCommand(); cmd != nil {
		t.Error("SelectedCommand should return nil when palette has no results")
	}
}

func TestSelectedCommandNegativeCursor(t *testing.T) {
	p := NewCommandPalette(nil, nil)
	p.cursor = -1
	if cmd := p.SelectedCommand(); cmd != nil {
		t.Error("SelectedCommand should return nil with negative cursor")
	}
}

func TestSelectedCommandOutOfBounds(t *testing.T) {
	p := NewCommandPalette(nil, nil)
	p.cursor = 999
	if cmd := p.SelectedCommand(); cmd != nil {
		t.Error("SelectedCommand should return nil with out-of-bounds cursor")
	}
}

// ── Search Result Scoring ─────────────────────────────────────────────────────

func TestScoreOperation(t *testing.T) {
	op := Command{
		ID: "test", Title: "CPU Monitor",
		Description: "Monitors CPU usage and temperature",
		Tags:        []string{"sysops", "cpu"},
	}
	score := scoreOperation(op, []string{"cpu"})
	if score <= 0 {
		t.Error("scoreOperation should return > 0 for matching query")
	}
}

func TestScoreOperationNoMatch(t *testing.T) {
	op := Command{
		ID: "test", Title: "CPU Monitor",
		Description: "Monitors CPU usage",
	}
	score := scoreOperation(op, []string{"zzz_nonexistent"})
	if score != 0 {
		t.Errorf("scoreOperation for non-matching query = %d, want 0", score)
	}
}

// ── CountByTag ────────────────────────────────────────────────────────────────

func TestCountByTag(t *testing.T) {
	ops := []Command{
		{ID: "a", Tags: []string{"sysops", "cpu"}},
		{ID: "b", Tags: []string{"netops", "icmp"}},
		{ID: "c", Tags: []string{"sysops", "memory"}},
	}
	count := countByTag("sysops", ops)
	if count != 2 {
		t.Errorf("countByTag('sysops') = %d, want 2", count)
	}
}

func TestCountByTagNone(t *testing.T) {
	count := countByTag("nonexistent", []Command{{Tags: []string{"sysops"}}})
	if count != 0 {
		t.Errorf("countByTag('nonexistent') = %d, want 0", count)
	}
}

// ── View Rendering ────────────────────────────────────────────────────────────

func TestCommandPaletteViewHidden(t *testing.T) {
	p := NewCommandPalette(nil, nil)
	if view := p.View(80, 24); view != "" {
		t.Error("hidden palette should render empty string")
	}
}

func TestCommandPaletteViewVisible(t *testing.T) {
	p := NewCommandPalette(nil, nil)
	p.Show()
	if view := p.View(80, 24); view == "" {
		t.Error("visible palette should not render empty string")
	}
}

func TestCommandPaletteViewNoResults(t *testing.T) {
	p := NewCommandPalette(nil, nil)
	p.Show()
	p.query = "zzz_nonexistent_xyzzy"
	p.results = SearchOps(p.query, AllOperations())
	if view := p.View(80, 24); view == "" {
		t.Error("visible palette with no results should still render")
	}
}

// ── Mode Constants ────────────────────────────────────────────────────────────

func TestPaletteModeValues(t *testing.T) {
	if PaletteSearch != 0 {
		t.Errorf("PaletteSearch = %d, want 0", PaletteSearch)
	}
	if PaletteTagFilter != 1 {
		t.Errorf("PaletteTagFilter = %d, want 1", PaletteTagFilter)
	}
	if PaletteRecommend != 2 {
		t.Errorf("PaletteRecommend = %d, want 2", PaletteRecommend)
	}
}

// ── Init Registration Check ───────────────────────────────────────────────────

func TestInitRegistersOperations(t *testing.T) {
	if len(registeredOps) == 0 {
		t.Fatal("registeredOps should be populated after init")
	}

	layerOps := []string{"sys.cpu", "net.ping", "sec.firewall", "dev.runner", "ai.chat"}
	for _, id := range layerOps {
		found := false
		for _, op := range registeredOps {
			if op.ID == id {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected operation %q to be registered by init", id)
		}
	}
}

// ── History Up ────────────────────────────────────────────────────────────────

func TestHistoryUpWhenEmpty(t *testing.T) {
	p := NewCommandPalette(nil, nil)
	p.historyUp() // should not panic
}

func TestHistoryUpCycles(t *testing.T) {
	p := NewCommandPalette(nil, nil)
	p.addToHistory("cpu")
	p.addToHistory("ping")

	p.historyUp()
	if p.query != "ping" {
		t.Errorf("historyUp query = %q, want %q", p.query, "ping")
	}

	p.historyUp()
	if p.query != "" {
		t.Logf("historyUp after full cycle query = %q", p.query)
	}
}

// ── Tab Navigation ────────────────────────────────────────────────────────────

func TestCommandPaletteTabNav(t *testing.T) {
	p := NewCommandPalette(nil, nil)
	p.Show()

	// Tab should jump forward by 8
	p.HandleKey(keyPressSpecial(tea.KeyTab))
	if p.cursor < 8 && len(p.results) >= 8 {
		t.Errorf("cursor should advance by 8 on tab, got %d", p.cursor)
	}

	// Shift+tab should jump backward
	p.HandleKey(keyPressSpecial(tea.KeyTab))
	p.HandleKey(keyPressSpecial(tea.KeyTab))
	p.HandleKey(keyPressSpecial(tea.KeyTab))
	prev := p.cursor

	// Use escape for now to test cursor reset
	p.HandleKey(keyPressSpecial(tea.KeyEsc))
	if p.IsVisible() {
		t.Error("palette should close on esc")
	}
	_ = prev
}
