# Hawkward Command Palette Design

> A global search-and-navigate overlay inspired by hackingtool's `/` search, accessible from any screen in the TUI.

---

## Design Principles

1. **Always accessible** — `/` opens from any screen, any state (except onboarding)
2. **Fuzzy search** — Fast, forgiving matching on operation names, descriptions, keywords, and tags
3. **Keyboard-first** — Full keyboard navigation: type to search, ↑↓ to navigate, Enter to select
4. **Minimal visual disruption** — Overlay at the top of the screen, does not replace the current view
5. **Discoverable** — Shows available operations with tags, descriptions, and key hints

---

## Component Architecture

### CommandPaletteModel

```go
// CommandPaletteModel manages the search overlay state.
type CommandPaletteModel struct {
    visible     bool              // Is the palette open?
    query       string            // Current search text
    results     []Operation       // Filtered operation list
    selected    int               // Index in results (for ↑↓ navigation)
    history     []string          // Previous searches (↑ cycles through)
    historyIdx  int               // Position in history
    mode        PaletteMode       // Search, TagFilter, Recommend
    tagFilter   string            // Current tag filter (mode=TagFilter)
    textInput   tea.Model         // Bubble tea text input model
}

type PaletteMode int

const (
    PaletteSearch    PaletteMode = iota // Default: type to search
    PaletteTagFilter                     // 't': filter by tag
    PaletteRecommend                    // 'r': natural language recommend
)
```

### Operation Registry

```go
// Operation describes a navigable action in the TUI.
type Operation struct {
    Name        string      // Display name (e.g., "Ping Tool")
    Description string      // Brief description (e.g., "ICMP ping to a target host")
    Tags        []string    // Categories (e.g., ["netops", "diagnostic", "icmp"])
    Screen      common.Screen  // Target screen to navigate to
    Action      func() tea.Cmd // Optional action (for non-navigation ops)
    Keywords    []string    // Extra search keywords not in name/desc
    Icon        string      // Emoji icon for display
}

// Global operation registry — populated by each ops layer at init.
var OperationRegistry []Operation

// RegisterOperations adds operations to the global registry.
func RegisterOperations(ops ...Operation) {
    OperationRegistry = append(OperationRegistry, ops...)
}
```

### Registration Example

```go
// In internal/sysops/init.go or at package init time:
func init() {
    ui.RegisterOperations(
        Operation{
            Name:        "CPU Dashboard",
            Description: "Real-time CPU usage, per-core breakdown, forecast",
            Tags:        []string{"sysops", "monitoring", "performance"},
            Screen:      common.ScreenSysOps,
            Keywords:    []string{"processor", "cores", "load"},
            Icon:        "🖥",
        },
        Operation{
            Name:        "Process List",
            Description: "Top processes by CPU/memory usage with drill-down",
            Tags:        []string{"sysops", "processes", "monitoring"},
            Screen:      common.ScreenSysOps,
            Keywords:    []string{"tasks", "kill", "threads"},
            Icon:        "⚙",
        },
        // ... more operations from each ops layer
    )
}
```

---

## Search Algorithm

### Fuzzy Search

```go
// SearchOps filters the operation registry by a fuzzy query.
func SearchOps(query string, ops []Operation) []Operation {
    if query == "" {
        return ops // No query → show all
    }
    
    query = strings.ToLower(strings.TrimSpace(query))
    queryWords := strings.Fields(query)
    
    type scored struct {
        op    Operation
        score int
    }
    
    var scoredOps []scored
    for _, op := range ops {
        score := scoreOperation(op, queryWords)
        if score > 0 {
            scoredOps = append(scoredOps, scored{op, score})
        }
    }
    
    // Sort by score descending
    sort.Slice(scoredOps, func(i, j int) bool {
        return scoredOps[i].score > scoredOps[j].score
    })
    
    result := make([]Operation, len(scoredOps))
    for i, s := range scoredOps {
        result[i] = s.op
    }
    return result
}

// scoreOperation calculates a relevance score for a single operation.
func scoreOperation(op Operation, queryWords []string) int {
    score := 0
    searchTarget := strings.ToLower(op.Name + " " + op.Description + " " + strings.Join(op.Keywords, " "))
    
    for _, word := range queryWords {
        if strings.Contains(searchTarget, word) {
            score += 10 // Full word match anywhere
        }
        if strings.HasPrefix(searchTarget, word) {
            score += 5 // Prefix match
        }
        for _, tag := range op.Tags {
            if strings.EqualFold(tag, word) {
                score += 15 // Exact tag match (highest priority)
            }
        }
        if strings.Contains(strings.ToLower(op.Name), word) {
            score += 8 // Name match bonus
        }
    }
    
    return score
}
```

### Tag Filter

```go
// FilterByTag filters operations by a specific tag.
func FilterByTag(tag string, ops []Operation) []Operation {
    tag = strings.ToLower(tag)
    var filtered []Operation
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

// AllTags returns all unique tags across registered operations.
func AllTags(ops []Operation) []string {
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
```

### Recommend Mode (Natural Language)

```go
// Recommend interprets a natural language query and returns matching operations.
func Recommend(query string, ops []Operation) []Operation {
    query = strings.ToLower(query)
    queryWords := strings.Fields(query)
    
    // Map common intent words to tags
    intentMap := map[string]string{
        "network":    "network",
        "internet":   "network",
        "connect":    "network",
        "ping":       "icmp",
        "dns":        "dns",
        "security":   "security",
        "firewall":   "firewall",
        "memory":     "memory",
        "cpu":        "cpu",
        "disk":       "disk",
        "process":    "process",
        "performance":"performance",
        "monitor":    "monitoring",
        "scan":       "scan",
        "ai":         "ai",
        "llm":        "ai",
        "ollama":     "ai",
    }
    
    // Score operations by intent match
    scored := make(map[string]int)
    for _, word := range queryWords {
        if tag, ok := intentMap[word]; ok {
            for _, op := range ops {
                for _, t := range op.Tags {
                    if t == tag {
                        scored[op.Name] += 20
                    }
                }
            }
        }
    }
    
    // Fall back to standard search if no intent matched
    if len(scored) == 0 {
        return SearchOps(query, ops)
    }
    
    // Sort by score and return
    // ...
}
```

---

## UI Rendering

### Palette Layout

```
┌──────────────────────────────────────────────────────────────────┐
│ / ████████████████                                        [2/15] │  ← Query bar
├──────────────────────────────────────────────────────────────────┤
│ 🖥 CPU Dashboard           sysops, monitoring, perf       ⏎     │  ← Results
│ 🌐 Ping Tool               network, icmp, diagnostic     ⏎     │
│ 🔒 Firewall Rules          security, firewall            ⏎     │
│ ⚙ Process List             sysops, processes             ⏎     │
│ ...                                                             │
│                                                                  │
│ Tags: [sysops] [network] [security] [ai] [all]                  │  ← Tag chips
│                                                                  │
│ ? help  ↑↓ navigate  enter select  t filter  r recommend  esc   │  ← Footer
└──────────────────────────────────────────────────────────────────┘
```

```go
// View renders the command palette overlay.
func (m *CommandPaletteModel) View(width, height int) string {
    if !m.visible {
        return ""
    }
    
    var b strings.Builder
    
    // Search bar
    b.WriteString(m.renderSearchBar(width))
    b.WriteString("\n")
    
    // Results
    b.WriteString(m.renderResults(width))
    b.WriteString("\n")
    
    // Tag chips (if any results)
    if len(m.results) > 0 {
        b.WriteString(m.renderTagChips(width))
        b.WriteString("\n")
    }
    
    // Footer help
    b.WriteString(m.renderFooter(width))
    
    // Wrap in overlay style with max height
    return overlayStyle.Width(width).
        MaxHeight(12).  // Show at most 12 lines
        Render(b.String())
}
```

### Search Bar Rendering

```go
func (m *CommandPaletteModel) renderSearchBar(width int) string {
    prefix := "/ "
    cursor := "█"
    if m.mode == PaletteTagFilter {
        prefix = "t "
    } else if m.mode == PaletteRecommend {
        prefix = "r "
    }
    
    queryText := prefix + m.query + cursor
    countText := fmt.Sprintf(" [%d/%d]", m.selected+1, len(m.results))
    
    return searchBarStyle.Width(width).
        Render(queryText + lipgloss.NewStyle().Width(width-lipgloss.Width(queryText)-lipgloss.Width(countText)).Render("") + countText)
}
```

### Result Rendering

```go
func (m *CommandPaletteModel) renderResults(width int) string {
    if len(m.results) == 0 {
        return mutedStyle.Render("  No matching operations. Try a different search or press 't' to browse by tag.")
    }
    
    var b strings.Builder
    visibleCount := min(len(m.results), 8) // Show at most 8 at a time
    start := max(0, m.selected - 3)        // Keep selected visible
    end := min(start+visibleCount, len(m.results))
    
    for i := start; i < end; i++ {
        op := m.results[i]
        line := fmt.Sprintf("  %s %s", op.Icon, op.Name)
        
        // Description
        desc := op.Description
        if len(desc) > 40 {
            desc = desc[:37] + "..."
        }
        line += strings.Repeat(" ", max(1, 30 - len(op.Name)))
        line += mutedStyle.Render(desc)
        
        // Tags
        if len(op.Tags) > 0 {
            line += "  "
            for _, tag := range op.Tags {
                line += tagStyle.Render("["+tag+"]") + " "
            }
        }
        
        if i == m.selected {
            b.WriteString(selectedStyle.Render(line))
        } else {
            b.WriteString(resultStyle.Render(line))
        }
        b.WriteString("\n")
    }
    
    return b.String()
}
```

---

## Key Bindings

```go
// Command palette key bindings (active when palette is open):
PaletteOpen     = key.NewBinding(key.WithKeys("/"),       key.WithHelp("/", "command palette"))
PaletteClose    = key.NewBinding(key.WithKeys("esc"),     key.WithHelp("esc", "close palette"))
PaletteUp       = key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "previous result"))
PaletteDown     = key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "next result"))
PaletteSelect   = key.NewBinding(key.WithKeys("enter"),   key.WithHelp("enter", "navigate to operation"))
PaletteTag      = key.NewBinding(key.WithKeys("t"),       key.WithHelp("t", "filter by tag"))
PaletteRecommend= key.NewBinding(key.WithKeys("r"),       key.WithHelp("r", "recommend mode"))
PaletteHistoryUp= key.NewBinding(key.WithKeys("up"),      key.WithHelp("↑", "search history"))
```

### Integration with RootModel

```go
// In internal/ui/root.go Update():
func (m *RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyPressMsg:
        // Command palette takes priority when visible
        if m.commandPalette.IsVisible() {
            switch {
            case matches(msg, PaletteClose):
                m.commandPalette.Hide()
            case matches(msg, PaletteSelect):
                op := m.commandPalette.SelectedOperation()
                if op != nil {
                    if op.Action != nil {
                        cmd = op.Action()
                    } else {
                        m.pushScreen(op.Screen)
                    }
                }
                m.commandPalette.Hide()
            default:
                m.commandPalette.HandleKey(msg)
            }
            return m, cmd
        }
        
        // Open palette from any screen
        if matches(msg, PaletteOpen) && m.onboarding.IsComplete() {
            m.commandPalette.Show()
            return m, nil
        }
        // ... rest of key handling
    }
}
```

---

## Search History

```go
// Search history management:
// - ↑ (when query is empty): cycles through previous searches
// - Each submitted search (Enter on non-registered query) adds to history
// - History is stored in the palette model (not persisted)

const maxSearchHistory = 20

func (m *CommandPaletteModel) addToHistory(query string) {
    if query == "" {
        return
    }
    // Remove duplicate if exists
    for i, h := range m.history {
        if h == query {
            m.history = append(m.history[:i], m.history[i+1:]...)
            break
        }
    }
    // Add to front
    m.history = append([]string{query}, m.history...)
    // Trim
    if len(m.history) > maxSearchHistory {
        m.history = m.history[:maxSearchHistory]
    }
}

func (m *CommandPaletteModel) historyUp() {
    if len(m.history) == 0 {
        return
    }
    m.historyIdx = (m.historyIdx - 1 + len(m.history)) % len(m.history)
    m.query = m.history[m.historyIdx]
    m.results = SearchOps(m.query, OperationRegistry)
    m.selected = 0
}
```

---

## Integration with Ops Layers

### Operation Registration by Layer

Each ops layer registers its operations at package init time:

**SysOps** operations:
- CPU Dashboard, Memory Monitor, Disk Usage, Process List, System Info, Service Status

**NetOps** operations:
- Ping Tool, DNS Lookup, Port Scanner, Connection Table, Traceroute, Interface Monitor, Bandwidth Graph

**SecOps** operations:
- Firewall Rules, User Audit, Listening Ports, Defender Status, Scheduled Tasks, Security Report

**DevOps** operations:
- Command Runner, Log Tailer, File Browser, Process Manager

**AI Ops** operations:
- AI Chat, Report Generator, System Summary, Trend Analysis, Anomaly Detection

### Tags Reference

| Tag | Description | Example Operations |
|-----|-------------|-------------------|
| `sysops` | System operations | CPU Dashboard, Memory Monitor |
| `netops` | Network operations | Ping, DNS, Port Scan |
| `secops` | Security operations | Firewall, Users, Defender |
| `devops` | Dev operations | Shell, Logs, Files |
| `aiops` | AI operations | Chat, Reports, Analysis |
| `monitoring` | Real-time monitoring | All dashboards |
| `diagnostic` | Troubleshooting tools | Ping, Traceroute, DNS |
| `scan` | Scanning tools | Port Scanner, User Audit |
| `report` | Report generation | All workflows |
| `forecast` | Predictive features | CPU/MEM/DISK forecast |
| `search` | Search/filter | Command palette itself |

---

## Implementation Plan

### Files
- `internal/ui/commandpalette.go` — CommandPaletteModel, search algorithm, rendering, integration

### Dependencies
- `github.com/charmbracelet/bubbles/textinput` — Text input widget for query entry
- `internal/common/types.go` — Requires `Operation` struct in shared types

### Tests

| Test | Description |
|------|-------------|
| `TestSearchOps_EmptyQuery` | Empty query returns all operations |
| `TestSearchOps_Fuzzy` | Partial word matches return ranked results |
| `TestSearchOps_ExactTag` | Tag filter returns only matching ops |
| `TestSearchOps_CaseInsensitive` | Search is case-insensitive |
| `TestRecommend_Intent` | Recommend mode maps intent words to tags |
| `TestSearchHistory` | History tracks correctly, no duplicates |
| `TestOperationRegistry` | All layers register operations at init |

---

*Last updated: 2026-07-07*
