package common

import (
	"crypto/rand"
	"fmt"
	"sync"
	"time"
)

// ── Event Types ──────────────────────────────────────────────────────────────

// EventLevel represents the severity of a timeline event.
type EventLevel int

const (
	EventInfo EventLevel = iota
	EventWarning
	EventCritical
)

func (l EventLevel) String() string {
	switch l {
	case EventInfo:
		return "info"
	case EventWarning:
		return "warning"
	case EventCritical:
		return "critical"
	default:
		return "info"
	}
}

// EventCategory categorises events by domain.
type EventCategory string

const (
	CatSystem   EventCategory = "system"
	CatNetwork  EventCategory = "network"
	CatSecurity EventCategory = "security"
	CatDevOps   EventCategory = "devops"
	CatAI       EventCategory = "ai"
	CatAlert    EventCategory = "alert"
	CatPipeline EventCategory = "pipeline"
)

// TimelineEvent is a single entry in the unified operations timeline.
type TimelineEvent struct {
	ID        string            `json:"id"`
	Timestamp time.Time         `json:"timestamp"`
	Category  EventCategory     `json:"category"`
	Level     EventLevel        `json:"level"`
	Title     string            `json:"title"`
	Detail    string            `json:"detail"`
	Module    string            `json:"module"`
	Related   []string          `json:"related,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// ── EventBus ─────────────────────────────────────────────────────────────────

// EventHandler is a callback for timeline events.
type EventHandler func(TimelineEvent)

// EventBus is a simple pub-sub bus for timeline events.
type EventBus struct {
	mu       sync.RWMutex
	handlers []EventHandler
	history  []TimelineEvent // bounded ring buffer for recent events
	capacity int
}

// NewEventBus creates an event bus that retains up to capacity events.
func NewEventBus(capacity int) *EventBus {
	return &EventBus{
		capacity: capacity,
		history:  make([]TimelineEvent, 0, capacity),
	}
}

// Subscribe registers a handler that receives all future events.
func (eb *EventBus) Subscribe(h EventHandler) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	eb.handlers = append(eb.handlers, h)
}

// Emit publishes an event to all subscribers and stores it in the ring buffer.
func (eb *EventBus) Emit(evt TimelineEvent) {
	eb.mu.Lock()

	// Ring-buffer append
	if len(eb.history) >= eb.capacity {
		eb.history = append(eb.history[1:], evt)
	} else {
		eb.history = append(eb.history, evt)
	}

	// Snapshot handlers under lock
	handlers := make([]EventHandler, len(eb.handlers))
	copy(handlers, eb.handlers)
	eb.mu.Unlock()

	// Notify outside lock
	for _, h := range handlers {
		h(evt)
	}
}

// Recent returns the most recent events, newest first.
func (eb *EventBus) Recent(n int) []TimelineEvent {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	if n <= 0 || n > len(eb.history) {
		n = len(eb.history)
	}

	// Return newest first
	out := make([]TimelineEvent, n)
	for i := 0; i < n; i++ {
		out[i] = eb.history[len(eb.history)-1-i]
	}
	return out
}

// All returns all stored events oldest first.
func (eb *EventBus) All() []TimelineEvent {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	out := make([]TimelineEvent, len(eb.history))
	copy(out, eb.history)
	return out
}

// FilterByCategory returns events matching the given category.
func (eb *EventBus) FilterByCategory(cat EventCategory, n int) []TimelineEvent {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	var filtered []TimelineEvent
	for i := len(eb.history) - 1; i >= 0 && len(filtered) < n; i-- {
		if eb.history[i].Category == cat {
			filtered = append(filtered, eb.history[i])
		}
	}
	return filtered
}

// ── UUID Helper ───────────────────────────────────────────────────────────────

// NewUUID returns a random 16-byte UUID v4 hex string.
func NewUUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	// Set version 4 and variant bits
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// ── Event Helpers ─────────────────────────────────────────────────────────────

// NewEvent creates a properly initialised TimelineEvent.
func NewEvent(category EventCategory, level EventLevel, module, title, detail string) TimelineEvent {
	return TimelineEvent{
		ID:        NewUUID(),
		Timestamp: time.Now(),
		Category:  category,
		Level:     level,
		Module:    module,
		Title:     title,
		Detail:    detail,
	}
}

// NewEventWithMeta creates an event with metadata key-value pairs.
func NewEventWithMeta(category EventCategory, level EventLevel, module, title, detail string, meta map[string]string) TimelineEvent {
	e := NewEvent(category, level, module, title, detail)
	e.Metadata = meta
	return e
}
