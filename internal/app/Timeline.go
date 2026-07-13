package app

import (
	"fmt"
	"sort"
	"time"

	"github.com/shahriarhaqueabir/AllOpsFull/internal/common"
)

// ── Timeline Binding ─────────────────────────────────────────────────────────

// Timeline exposes timeline operations to the frontend.
type Timeline struct {
	app *App
}

// NewTimeline creates a new Timeline facade.
func NewTimeline(app *App) *Timeline {
	return &Timeline{app: app}
}

// GetTimelineEvents returns timeline events with optional filters.
// category and level can be empty to include all.
func (t *Timeline) GetTimelineEvents(category, level string, limit, offset int) []TimelineEvent {
	if limit <= 0 || limit > 500 {
		limit = 100
	}

	// Try persistent store first
	storage := common.GetStorage()
	if storage != nil {
		events, err := storage.QueryEvents(category, level, limit, offset)
		if err == nil && len(events) > 0 {
			return convertTimelineEvents(events)
		}
	}

	// Fall back to in-memory bus
	all := t.app.eventBus.All()
	var filtered []common.TimelineEvent
	for _, e := range all {
		if category != "" && string(e.Category) != category {
			continue
		}
		if level != "" && e.Level.String() != level {
			continue
		}
		filtered = append(filtered, e)
	}

	// Sort newest first
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Timestamp.After(filtered[j].Timestamp)
	})

	// Apply offset + limit
	if offset >= len(filtered) {
		return []TimelineEvent{}
	}
	filtered = filtered[offset:]
	if len(filtered) > limit {
		filtered = filtered[:limit]
	}

	return convertTimelineEvents(filtered)
}

// GetTimelineEventByID returns a single event by ID.
func (t *Timeline) GetTimelineEventByID(id string) *TimelineEvent {
	storage := common.GetStorage()
	if storage != nil {
		evt, err := storage.GetEventByID(id)
		if err == nil && evt != nil {
			converted := convertTimelineEvent(*evt)
			return &converted
		}
	}
	return nil
}

// GetRelatedEvents returns events related to a given event.
func (t *Timeline) GetRelatedEvents(eventID string) []TimelineEvent {
	evt := t.GetTimelineEventByID(eventID)
	if evt == nil || len(evt.Related) == 0 {
		return []TimelineEvent{}
	}

	storage := common.GetStorage()
	if storage == nil {
		return []TimelineEvent{}
	}

	var results []TimelineEvent
	for _, rid := range evt.Related {
		related, err := storage.GetEventByID(rid)
		if err == nil && related != nil {
			results = append(results, convertTimelineEvent(*related))
		}
	}
	return results
}

// GetTimelineCategories returns all distinct event categories.
func (t *Timeline) GetTimelineCategories() []string {
	return []string{"system", "network", "security", "devops", "ai", "alert", "pipeline"}
}

// GetTimelineSummary returns a summary of events grouped by category.
func (t *Timeline) GetTimelineSummary(sinceMinutes int) map[string]int {
	if sinceMinutes <= 0 {
		sinceMinutes = 60
	}
	since := time.Now().Add(-time.Duration(sinceMinutes) * time.Minute)

	storage := common.GetStorage()
	if storage == nil {
		return map[string]int{}
	}

	summary := make(map[string]int)
	for _, cat := range t.GetTimelineCategories() {
		events, err := storage.QueryEvents(cat, "", 1000, 0)
		if err != nil {
			continue
		}
		count := 0
		for _, e := range events {
			if e.Timestamp.After(since) {
				count++
			}
		}
		if count > 0 {
			summary[cat] = count
		}
	}
	return summary
}

// ── AI-powered Explanation ───────────────────────────────────────────────────

// ExplainEvents asks the AI to analyse a set of timeline events.
func (t *Timeline) ExplainEvents(eventIDs []string) string {
	if len(eventIDs) == 0 {
		return "No events to analyse."
	}

	storage := common.GetStorage()
	if storage == nil {
		return "Storage not available."
	}

	// Gather events
	var events []common.TimelineEvent
	for _, id := range eventIDs {
		evt, err := storage.GetEventByID(id)
		if err == nil && evt != nil {
			events = append(events, *evt)
		}
	}

	if len(events) == 0 {
		return "No events found for the given IDs."
	}

	// Try AI analysis
	if t.app.AIOps != nil {
		prompt := buildTimelineAnalysisPrompt(events)
		ctx, cancel := t.app.AIOps.WithTimeout(30 * time.Second)
		defer cancel()
		analysis, err := t.app.AIOps.AskAI(ctx, prompt)
		if err == nil && analysis != "" {
			return analysis
		}
	}

	// Fallback: generate a basic textual summary
	return buildTimelineFallbackSummary(events)
}

// ── Internal Helpers ─────────────────────────────────────────────────────────

func convertTimelineEvent(e common.TimelineEvent) TimelineEvent {
	related := e.Related
	if related == nil {
		related = []string{}
	}
	return TimelineEvent{
		ID:        e.ID,
		Timestamp: e.Timestamp.Format(time.RFC3339),
		Category:  string(e.Category),
		Level:     e.Level.String(),
		Title:     e.Title,
		Detail:    e.Detail,
		Module:    e.Module,
		Related:   related,
		Metadata:  e.Metadata,
	}
}

func convertTimelineEvents(events []common.TimelineEvent) []TimelineEvent {
	out := make([]TimelineEvent, len(events))
	for i, e := range events {
		out[i] = convertTimelineEvent(e)
	}
	return out
}

func buildTimelineAnalysisPrompt(events []common.TimelineEvent) string {
	prompt := "Analyse these operations events and explain what's happening, whether they're related, and what actions to recommend:\n\n"
	for _, e := range events {
		prompt += fmt.Sprintf("[%s] [%s] [%s] %s — %s\n",
			e.Timestamp.Format("15:04:05"),
			e.Level.String(),
			e.Category,
			e.Title,
			e.Detail,
		)
	}
	prompt += "\nProvide a concise analysis."
	return prompt
}

func buildTimelineFallbackSummary(events []common.TimelineEvent) string {
	levels := make(map[string]int)
	cats := make(map[string]int)
	for _, e := range events {
		levels[e.Level.String()]++
		cats[string(e.Category)]++
	}

	summary := fmt.Sprintf("Analysed %d events. ", len(events))
	for lvl, count := range levels {
		summary += fmt.Sprintf("%d %s-level, ", count, lvl)
	}
	summary = summary[:len(summary)-2] + ". "
	summary += "Categories: "
	for cat, count := range cats {
		summary += fmt.Sprintf("%d %s, ", count, cat)
	}
	summary = summary[:len(summary)-2] + "."
	return summary
}
