package ui

import (
	"fmt"
	"sort"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
)

// ── Palette Mode ──────────────────────────────────────────────────────────────

// PaletteMode represents the current search mode of the command palette.
type PaletteMode int

const (
	PaletteSearch    PaletteMode = iota // Default: fuzzy search all operations
	PaletteTagFilter                    // 't' mode: filter by tag
	PaletteRecommend                    // 'r' mode: natural language → operations
)

const maxSearchHistory = 20

// ── Command ───────────────────────────────────────────────────────────────────

// Command describes a navigable action in the TUI.
type Command struct {
	ID          string
	Title       string
	Description string
	Tags        []string
	Icon        string
	Screen      common.Screen
	Action      func() tea.Cmd
	Key         string
}

// ── SearchResult ──────────────────────────────────────────────────────────────

// SearchResult holds a command with its fuzzy match score and matched positions.
type SearchResult struct {
	Command Command
	Score   int
	Matches []int
}

// ── CommandPalette ────────────────────────────────────────────────────────────

// CommandPalette manages the global search overlay state.
type CommandPalette struct {
	visible   bool
	query     string
	cursor    int // index into results for ↑↓ navigation
	results   []SearchResult
	history   []string // search history ring buffer
	histIdx   int      // position in history
	mode      PaletteMode
	tagFilter string

	onSelect func(cmd Command) tea.Cmd
	onClose  func() tea.Cmd

	// Visual tracking for scroll
	scrollOffset int
}

// NewCommandPalette creates a new command palette with the given callbacks.
func NewCommandPalette(onSelect func(cmd Command) tea.Cmd, onClose func() tea.Cmd) *CommandPalette {
	return &CommandPalette{
		results:  make([]SearchResult, 0),
		history:  make([]string, 0, maxSearchHistory),
		histIdx:  -1,
		onSelect: onSelect,
		onClose:  onClose,
	}
}

// ── Operation Registry ────────────────────────────────────────────────────────

var registeredOps []Command

// RegisterOperation adds an operation to the global registry.
func RegisterOperation(cmd Command) {
	registeredOps = append(registeredOps, cmd)
}

// AllOperations returns a copy of all registered operations.
func AllOperations() []Command {
	ops := make([]Command, len(registeredOps))
	copy(ops, registeredOps)
	return ops
}

// init registers all default operations across all five ops layers.
func init() {
	// ── SysOps ──
	RegisterOperation(Command{
		ID: "sys.cpu", Title: "CPU Dashboard",
		Description: "Real-time CPU usage, per-core breakdown, and load forecast",
		Tags:        []string{"sysops", "monitoring", "performance"},
		Icon:        "🖥", Screen: common.ScreenSysOps,
	})
	RegisterOperation(Command{
		ID: "sys.memory", Title: "Memory Monitor",
		Description: "RAM usage, swap, and memory pressure details",
		Tags:        []string{"sysops", "monitoring", "memory"},
		Icon:        "🧠", Screen: common.ScreenSysOps,
	})
	RegisterOperation(Command{
		ID: "sys.disk", Title: "Disk Usage",
		Description: "Disk space, I/O stats, and partition information",
		Tags:        []string{"sysops", "monitoring", "disk"},
		Icon:        "💾", Screen: common.ScreenSysOps,
	})
	RegisterOperation(Command{
		ID: "sys.process", Title: "Process List",
		Description: "Top processes by CPU and memory usage with drill-down",
		Tags:        []string{"sysops", "processes", "monitoring"},
		Icon:        "⚙", Screen: common.ScreenSysOps,
	})
	RegisterOperation(Command{
		ID: "sys.info", Title: "System Info",
		Description: "OS version, hardware specs, kernel, and uptime details",
		Tags:        []string{"sysops", "diagnostic"},
		Icon:        "ℹ", Screen: common.ScreenSysOps,
	})

	// ── NetOps ──
	RegisterOperation(Command{
		ID: "net.ping", Title: "Ping Tool",
		Description: "ICMP ping with RTT statistics and packet loss analysis",
		Tags:        []string{"netops", "diagnostic", "icmp"},
		Icon:        "🌐", Screen: common.ScreenNetOps,
	})
	RegisterOperation(Command{
		ID: "net.dns", Title: "DNS Lookup",
		Description: "Query DNS records: A, AAAA, MX, NS, TXT, CNAME",
		Tags:        []string{"netops", "diagnostic", "dns"},
		Icon:        "📡", Screen: common.ScreenNetOps,
	})
	RegisterOperation(Command{
		ID: "net.portscan", Title: "Port Scanner",
		Description: "TCP/UDP port scanning with service detection",
		Tags:        []string{"netops", "scan", "diagnostic"},
		Icon:        "🔍", Screen: common.ScreenNetOps,
	})
	RegisterOperation(Command{
		ID: "net.connections", Title: "Connection Table",
		Description: "Active network connections, listening sockets, and state",
		Tags:        []string{"netops", "monitoring", "network"},
		Icon:        "🔗", Screen: common.ScreenNetOps,
	})
	RegisterOperation(Command{
		ID: "net.traceroute", Title: "Traceroute",
		Description: "Route tracing to target hosts with hop-by-hop latency",
		Tags:        []string{"netops", "diagnostic", "network"},
		Icon:        "🗺", Screen: common.ScreenNetOps,
	})
	RegisterOperation(Command{
		ID: "net.bandwidth", Title: "Bandwidth Graph",
		Description: "Real-time network bandwidth usage and throughput chart",
		Tags:        []string{"netops", "monitoring", "performance"},
		Icon:        "📊", Screen: common.ScreenNetOps,
	})

	// ── SecOps ──
	RegisterOperation(Command{
		ID: "sec.firewall", Title: "Firewall Rules",
		Description: "View and manage Windows firewall rules",
		Tags:        []string{"secops", "firewall", "security"},
		Icon:        "🔒", Screen: common.ScreenSecOps,
	})
	RegisterOperation(Command{
		ID: "sec.users", Title: "User Audit",
		Description: "Local user accounts, groups, and security identifiers",
		Tags:        []string{"secops", "audit", "users"},
		Icon:        "👤", Screen: common.ScreenSecOps,
	})
	RegisterOperation(Command{
		ID: "sec.ports", Title: "Listening Ports",
		Description: "Open TCP/UDP ports and associated processes",
		Tags:        []string{"secops", "scan", "network"},
		Icon:        "🔌", Screen: common.ScreenSecOps,
	})
	RegisterOperation(Command{
		ID: "sec.defender", Title: "Defender Status",
		Description: "Windows Defender health, signature age, and protection status",
		Tags:        []string{"secops", "security", "monitoring"},
		Icon:        "🛡", Screen: common.ScreenSecOps,
	})
	RegisterOperation(Command{
		ID: "sec.tasks", Title: "Scheduled Tasks",
		Description: "List and inspect Windows scheduled tasks",
		Tags:        []string{"secops", "audit", "tasks"},
		Icon:        "📋", Screen: common.ScreenSecOps,
	})
	RegisterOperation(Command{
		ID: "sec.report", Title: "Security Report",
		Description: "Generate a comprehensive security posture report",
		Tags:        []string{"secops", "report", "audit"},
		Icon:        "📄", Screen: common.ScreenSecOps,
	})

	// ── DevOps ──
	RegisterOperation(Command{
		ID: "dev.runner", Title: "Command Runner",
		Description: "Execute shell commands and capture output",
		Tags:        []string{"devops", "shell", "tools"},
		Icon:        "⌨", Screen: common.ScreenDevOps,
	})
	RegisterOperation(Command{
		ID: "dev.logs", Title: "Log Tailer",
		Description: "Tail and filter application and system log files",
		Tags:        []string{"devops", "monitoring", "logs"},
		Icon:        "📝", Screen: common.ScreenDevOps,
	})
	RegisterOperation(Command{
		ID: "dev.files", Title: "File Browser",
		Description: "Browse, search, and preview files in the filesystem",
		Tags:        []string{"devops", "files", "tools"},
		Icon:        "📁", Screen: common.ScreenDevOps,
	})
	RegisterOperation(Command{
		ID: "dev.manager", Title: "Process Manager",
		Description: "Manage running processes: view, filter, and terminate",
		Tags:        []string{"devops", "processes", "tools"},
		Icon:        "🛠", Screen: common.ScreenDevOps,
	})

	// ── AIOps ──
	RegisterOperation(Command{
		ID: "ai.chat", Title: "AI Chat",
		Description: "Conversational AI assistant for system queries and help",
		Tags:        []string{"aiops", "ai", "assistant"},
		Icon:        "🤖", Screen: common.ScreenAIOps,
	})
	RegisterOperation(Command{
		ID: "ai.report", Title: "Report Generator",
		Description: "Generate AI-powered system and diagnostic reports",
		Tags:        []string{"aiops", "report", "ai"},
		Icon:        "📊", Screen: common.ScreenAIOps,
	})
	RegisterOperation(Command{
		ID: "ai.summary", Title: "System Summary",
		Description: "AI-generated overview of current system health and status",
		Tags:        []string{"aiops", "ai", "monitoring"},
		Icon:        "📋", Screen: common.ScreenAIOps,
	})
	RegisterOperation(Command{
		ID: "ai.trends", Title: "Trend Analysis",
		Description: "Detect anomalies and trends in system metrics over time",
		Tags:        []string{"aiops", "ai", "forecast"},
		Icon:        "📈", Screen: common.ScreenAIOps,
	})
}

// ── Fuzzy Search ──────────────────────────────────────────────────────────────

// fuzzyScore computes a fuzzy match score between query and target.
// Returns a score and the indices of matched characters (for highlighting).
func fuzzyScore(query, target string) (score int, matches []int) {
	query = strings.ToLower(query)
	target = strings.ToLower(target)

	if query == "" {
		return 0, nil
	}

	// Exact match = 100
	if target == query {
		matches = make([]int, len(query))
		for i := 0; i < len(query); i++ {
			matches[i] = i
		}
		return 100, matches
	}

	// Contains prefix = 80
	if strings.HasPrefix(target, query) {
		matches = make([]int, len(query))
		for i := 0; i < len(query); i++ {
			matches[i] = i
		}
		return 80, matches
	}

	// Contains substring = 60
	if idx := strings.Index(target, query); idx >= 0 {
		matches = make([]int, len(query))
		for i := 0; i < len(query); i++ {
			matches[i] = idx + i
		}
		return 60, matches
	}

	// Character sequence match: score based on density
	// Walk through target finding each query character in order
	qi := 0
	densityScore := 0
	matchedIdxs := make([]int, 0, len(query))
	for ti, tc := range target {
		if qi < len(query) && byte(tc) == query[qi] {
			matchedIdxs = append(matchedIdxs, ti)
			if len(matchedIdxs) > 1 {
				gap := ti - matchedIdxs[len(matchedIdxs)-2]
				if gap <= 2 {
					densityScore += 10 // adjacent or near-adjacent
				} else if gap <= 5 {
					densityScore += 5
				} else {
					densityScore += 2
				}
			}
			qi++
		}
	}

	if qi == len(query) {
		// All characters matched sequentially
		score = 30 + densityScore
		if score > 55 {
			score = 55
		}
		return score, matchedIdxs
	}

	return 0, nil
}

// scoreOperation calculates a relevance score for a single operation against
// a set of query words. Combines fuzzy name/description matching with tag hits.
func scoreOperation(op Command, queryWords []string) int {
	score := 0
	searchTarget := strings.ToLower(op.Title + " " + op.Description)

	for _, word := range queryWords {
		// Fuzzy match on title (weighted highest)
		if s, _ := fuzzyScore(word, op.Title); s > 0 {
			score += s
		}

		// Contains match in description
		if strings.Contains(searchTarget, word) {
			score += 10
		}

		// Prefix match on title
		if strings.HasPrefix(strings.ToLower(op.Title), word) {
			score += 8
		}

		// Exact tag match (highest per-word bonus)
		for _, tag := range op.Tags {
			if strings.EqualFold(tag, word) {
				score += 15
			}
		}

		// ID partial match
		if strings.Contains(strings.ToLower(op.ID), word) {
			score += 5
		}
	}

	return score
}

// SearchOps filters the operation registry by a fuzzy query, returning
// results sorted by descending score.
func SearchOps(query string, ops []Command) []SearchResult {
	if query == "" {
		// Empty query returns all operations in registration order with score 0
		results := make([]SearchResult, len(ops))
		for i, op := range ops {
			results[i] = SearchResult{Command: op, Score: 0}
		}
		return results
	}

	query = strings.ToLower(strings.TrimSpace(query))
	queryWords := strings.Fields(query)

	type scored struct {
		op    Command
		score int
	}

	var scoredOps []scored
	for _, op := range ops {
		s := scoreOperation(op, queryWords)
		if s > 0 {
			scoredOps = append(scoredOps, scored{op, s})
		}
	}

	// Sort by score descending
	sort.Slice(scoredOps, func(i, j int) bool {
		return scoredOps[i].score > scoredOps[j].score
	})

	results := make([]SearchResult, len(scoredOps))
	for i, s := range scoredOps {
		results[i] = SearchResult{Command: s.op, Score: s.score}
	}
	return results
}

// ── Tag Filter ────────────────────────────────────────────────────────────────

// FilterByTag filters commands by exact tag match (case-insensitive).
func FilterByTag(tag string, ops []Command) []Command {
	tag = strings.ToLower(tag)
	var filtered []Command
	for _, op := range ops {
		for _, t := range op.Tags {
			if strings.ToLower(t) == tag {
				filtered = append(filtered, op)
				break
			}
		}
	}
	return filtered
}

// AllTags returns all unique tags across the given commands, sorted.
func AllTags(ops []Command) []string {
	tagSet := make(map[string]bool)
	for _, op := range ops {
		for _, tag := range op.Tags {
			tagSet[tag] = true
		}
	}
	tags := make([]string, 0, len(tagSet))
	for tag := range tagSet {
		tags = append(tags, tag)
	}
	sort.Strings(tags)
	return tags
}

// ── Recommend Mode ────────────────────────────────────────────────────────────

// Recommend interprets a natural language query and returns matching operations
// by mapping common intent words to tag-based results.
func Recommend(query string, ops []Command) []SearchResult {
	query = strings.ToLower(query)
	queryWords := strings.Fields(query)

	// Map common intent words to tags
	intentMap := map[string]string{
		"network":     "netops",
		"internet":    "netops",
		"connect":     "netops",
		"ping":        "icmp",
		"dns":         "dns",
		"security":    "secops",
		"firewall":    "firewall",
		"memory":      "memory",
		"cpu":         "cpu",
		"disk":        "disk",
		"process":     "processes",
		"performance": "performance",
		"monitor":     "monitoring",
		"scan":        "scan",
		"ai":          "ai",
		"llm":         "ai",
		"ollama":      "ai",
		"system":      "sysops",
		"developer":   "devops",
		"dev":         "devops",
		"code":        "devops",
		"shell":       "shell",
		"report":      "report",
		"diagnose":    "diagnostic",
		"diagnostic":  "diagnostic",
	}

	// Score operations by intent match
	type scored struct {
		op    Command
		score int
	}
	scoredMap := make(map[string]int)
	for _, word := range queryWords {
		if tag, ok := intentMap[word]; ok {
			for _, op := range ops {
				for _, t := range op.Tags {
					if strings.EqualFold(t, tag) {
						scoredMap[op.ID] += 20
					}
				}
			}
		}
	}

	// Fall back to standard search if no intent matched
	if len(scoredMap) == 0 {
		return SearchOps(query, ops)
	}

	var scoredOps []scored
	for _, op := range ops {
		if s, ok := scoredMap[op.ID]; ok {
			scoredOps = append(scoredOps, scored{op, s})
		}
	}

	sort.Slice(scoredOps, func(i, j int) bool {
		return scoredOps[i].score > scoredOps[j].score
	})

	results := make([]SearchResult, len(scoredOps))
	for i, s := range scoredOps {
		results[i] = SearchResult{Command: s.op, Score: s.score}
	}
	return results
}

// ── Lifecycle Methods ─────────────────────────────────────────────────────────

// Show opens the palette and initializes the result list.
func (p *CommandPalette) Show() {
	if p.visible {
		return
	}
	p.visible = true
	p.query = ""
	p.cursor = 0
	p.scrollOffset = 0
	p.mode = PaletteSearch
	p.tagFilter = ""
	p.results = SearchOps("", AllOperations())
	common.LogInfo("Command palette opened")
}

// Hide closes the palette.
func (p *CommandPalette) Hide() {
	if !p.visible {
		return
	}
	p.visible = false
	if p.onClose != nil {
		p.onClose()
	}
	common.LogInfo("Command palette closed")
}

// IsVisible returns whether the palette is currently open.
func (p *CommandPalette) IsVisible() bool {
	return p.visible
}

// Toggle opens or closes the palette.
func (p *CommandPalette) Toggle() {
	if p.visible {
		p.Hide()
	} else {
		p.Show()
	}
}

// SelectedCommand returns the currently selected command, or nil.
func (p *CommandPalette) SelectedCommand() *Command {
	if len(p.results) == 0 || p.cursor < 0 || p.cursor >= len(p.results) {
		return nil
	}
	return &p.results[p.cursor].Command
}

// ── Key Handling ──────────────────────────────────────────────────────────────

// HandleKey processes a key press when the palette is visible.
func (p *CommandPalette) HandleKey(msg tea.KeyPressMsg) {
	key := msg.String()

	switch key {
	case "esc":
		p.Hide()
		return

	case "enter":
		cmd := p.SelectedCommand()
		if cmd != nil && p.onSelect != nil {
			p.addToHistory(p.query)
			p.Hide()
			p.onSelect(*cmd)
		}
		return

	case "up", "k":
		if len(p.results) > 0 {
			p.cursor--
			if p.cursor < 0 {
				p.cursor = len(p.results) - 1
			}
			p.ensureCursorVisible()
		}
		return

	case "down", "j":
		if len(p.results) > 0 {
			p.cursor++
			if p.cursor >= len(p.results) {
				p.cursor = 0
			}
			p.ensureCursorVisible()
		}
		return

	case "tab":
		// Next page of results
		if len(p.results) > 0 {
			p.cursor += 8
			if p.cursor >= len(p.results) {
				p.cursor = len(p.results) - 1
			}
			p.ensureCursorVisible()
		}
		return

	case "shift+tab":
		// Previous page of results
		if len(p.results) > 0 {
			p.cursor -= 8
			if p.cursor < 0 {
				p.cursor = 0
			}
			p.ensureCursorVisible()
		}
		return

	case "t":
		// Toggle tag filter mode
		if p.mode == PaletteTagFilter {
			p.mode = PaletteSearch
		} else {
			p.mode = PaletteTagFilter
		}
		p.query = ""
		p.cursor = 0
		p.scrollOffset = 0
		if p.mode == PaletteTagFilter {
			// Show all available tags as results
			ops := AllOperations()
			tags := AllTags(ops)
			p.results = make([]SearchResult, len(tags))
			common.LogInfo("Number of tags: %d", len(tags))
			for i, tag := range tags {
				// Create a display-only command for each tag
				p.results[i] = SearchResult{
					Command: Command{
						ID:          "tag." + tag,
						Title:       tag,
						Description: fmt.Sprintf("Filter by tag: %s (%d ops)", tag, countByTag(tag, ops)),
						Tags:        []string{tag},
						Icon:        "#",
					},
					Score: 0,
				}
			}
		} else {
			p.results = SearchOps("", AllOperations())
		}
		return

	case "r":
		// Toggle recommend mode
		if p.mode == PaletteRecommend {
			p.mode = PaletteSearch
			p.results = SearchOps(p.query, AllOperations())
		} else {
			p.mode = PaletteRecommend
			if p.query != "" {
				p.results = Recommend(p.query, AllOperations())
			} else {
				p.results = SearchOps(p.query, AllOperations())
			}
		}
		p.cursor = 0
		p.scrollOffset = 0
		return

	case "backspace":
		if len(p.query) > 0 {
			p.query = p.query[:len(p.query)-1]
			p.cursor = 0
			p.scrollOffset = 0
			p.performSearch()
		} else if p.mode == PaletteSearch && len(p.history) > 0 {
			// On empty backspace, cycle history up
			p.historyUp()
		}
		return

	case "space":
		p.query += " "
		p.cursor = 0
		p.scrollOffset = 0
		p.performSearch()
		return

	default:
		// Regular character input
		if len(key) == 1 {
			p.query += key
			p.cursor = 0
			p.scrollOffset = 0
			p.performSearch()
		}
	}
}

// performSearch runs the current query against all operations based on mode.
func (p *CommandPalette) performSearch() {
	ops := AllOperations()
	switch p.mode {
	case PaletteSearch:
		p.results = SearchOps(p.query, ops)
	case PaletteTagFilter:
		p.results = SearchOps(p.query, ops)
	case PaletteRecommend:
		if p.query != "" {
			p.results = Recommend(p.query, ops)
		} else {
			p.results = SearchOps("", ops)
		}
	}
}

// ensureCursorVisible adjusts scrollOffset so the cursor is visible.
func (p *CommandPalette) ensureCursorVisible() {
	const visibleResults = 8
	if p.cursor < p.scrollOffset {
		p.scrollOffset = p.cursor
	}
	if p.cursor >= p.scrollOffset+visibleResults {
		p.scrollOffset = p.cursor - visibleResults + 1
	}
}

// countByTag returns how many ops have the given tag.
func countByTag(tag string, ops []Command) int {
	count := 0
	for _, op := range ops {
		for _, t := range op.Tags {
			if strings.EqualFold(t, tag) {
				count++
				break
			}
		}
	}
	return count
}

// ── Search History ────────────────────────────────────────────────────────────

// addToHistory adds a query to the search history, deduplicating and trimming.
func (p *CommandPalette) addToHistory(query string) {
	if query == "" {
		return
	}
	// Remove duplicate if exists
	for i, h := range p.history {
		if h == query {
			p.history = append(p.history[:i], p.history[i+1:]...)
			break
		}
	}
	// Add to front
	p.history = append([]string{query}, p.history...)
	// Trim
	if len(p.history) > maxSearchHistory {
		p.history = p.history[:maxSearchHistory]
	}
	p.histIdx = -1
}

// historyUp cycles through search history from newest to oldest.
func (p *CommandPalette) historyUp() {
	if len(p.history) == 0 {
		return
	}
	p.histIdx = (p.histIdx + 1) % len(p.history)
	p.query = p.history[p.histIdx]
	p.cursor = 0
	p.scrollOffset = 0
	p.performSearch()
}

// ── Rendering ─────────────────────────────────────────────────────────────────

// View renders the command palette overlay.
func (p *CommandPalette) View(width, height int) string {
	if !p.visible {
		return ""
	}

	pal := common.CurrentPalette()
	var b strings.Builder

	// ── Search bar ──
	b.WriteString(p.renderSearchBar(width, pal))
	b.WriteString("\n")

	// ── Results ──
	b.WriteString(p.renderResults(width, pal))
	b.WriteString("\n")

	// ── Footer help ──
	b.WriteString(p.renderFooter(width, pal))

	// Wrap everything in a bordered box
	overlayStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(pal.Primary)).
		Padding(0, 1).
		Width(width - 6). // account for border + padding
		Background(lipgloss.Color(pal.CardBg))

	return overlayStyle.Render(b.String())
}

// renderSearchBar renders the query input line.
func (p *CommandPalette) renderSearchBar(width int, pal common.Palette) string {
	prefix := "/ "
	if p.mode == PaletteTagFilter {
		prefix = "t "
	} else if p.mode == PaletteRecommend {
		prefix = "r "
	}

	cursor := "█"
	countText := ""
	if len(p.results) > 0 {
		countText = fmt.Sprintf(" [%d/%d]", p.cursor+1, len(p.results))
	}

	// Style the search bar
	searchStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(pal.Text)).
		Background(lipgloss.Color(pal.CardBg)).
		Bold(true)

	// Mode prefix color
	modeStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(pal.Secondary)).
		Bold(true)

	countStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(pal.Muted))

	return modeStyle.Render(prefix) + searchStyle.Render(p.query) +
		lipgloss.NewStyle().Foreground(lipgloss.Color(pal.Text)).Render(cursor) +
		" " + countStyle.Render(countText)
}

// renderResults renders the filtered operation list.
func (p *CommandPalette) renderResults(width int, pal common.Palette) string {
	if len(p.results) == 0 {
		mutedStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(pal.Muted)).
			Padding(0, 1)
		if p.mode == PaletteTagFilter {
			return mutedStyle.Render("  No operations found for this tag filter.")
		}
		return mutedStyle.Render("  No matching operations. Try a different search, press [t] to browse tags, or [r] for recommend mode.")
	}

	var b strings.Builder
	visibleCount := 8
	start := p.scrollOffset
	end := start + visibleCount
	if end > len(p.results) {
		end = len(p.results)
	}

	resultStyle := lipgloss.NewStyle().
		Padding(0, 1).
		Foreground(lipgloss.Color(pal.Text))

	selectedStyle := lipgloss.NewStyle().
		Padding(0, 1).
		Foreground(lipgloss.Color(pal.Text)).
		Background(lipgloss.Color(pal.Primary)).
		Bold(true)

	mutedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(pal.Muted))

	tagStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(pal.Info)).
		Padding(0, 0)

	for i := start; i < end; i++ {
		op := p.results[i].Command
		line := fmt.Sprintf("  %s  %s", op.Icon, op.Title)

		// Description (truncated)
		desc := op.Description
		if len(desc) > 36 {
			desc = desc[:33] + "..."
		}
		// Pad title column to align descriptions
		titleWidth := lipgloss.Width(op.Title)
		padding := 28 - titleWidth
		if padding < 1 {
			padding = 1
		}
		line += strings.Repeat(" ", padding)
		line += mutedStyle.Render(desc)

		// Tags
		if len(op.Tags) > 0 {
			line += "  "
			for _, tag := range op.Tags {
				line += tagStyle.Render("["+tag+"]") + " "
			}
		}

		if i == p.cursor {
			b.WriteString(selectedStyle.Render(line))
		} else {
			b.WriteString(resultStyle.Render(line))
		}
		b.WriteString("\n")
	}

	return b.String()
}

// renderFooter renders the help footer line.
func (p *CommandPalette) renderFooter(width int, pal common.Palette) string {
	modeLabel := "search"
	if p.mode == PaletteTagFilter {
		modeLabel = "tag filter"
	} else if p.mode == PaletteRecommend {
		modeLabel = "recommend"
	}

	help := fmt.Sprintf("  %d results  |  [t] tags  [r] recommend  [esc] close  • mode: %s",
		len(p.results), modeLabel)

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(pal.Muted)).
		Padding(0, 1)

	return helpStyle.Render(help)
}
