package common

import (
	"crypto/rand"
	"fmt"
	"regexp"
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
	history  []TimelineEvent // circular buffer for recent events
	capacity int
	head     int  // index to the next insert position
	full     bool // true if history has been filled at least once
}

// NewEventBus creates an event bus that retains up to capacity events.
func NewEventBus(capacity int) *EventBus {
	if capacity <= 0 {
		capacity = 100 // fallback
	}
	return &EventBus{
		capacity: capacity,
		history:  make([]TimelineEvent, capacity),
	}
}

// Subscribe registers a handler that receives all future events.
// It returns a function that can be called to unsubscribe.
func (eb *EventBus) Subscribe(h EventHandler) func() {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	eb.handlers = append(eb.handlers, h)

	return func() {
		eb.mu.Lock()
		defer eb.mu.Unlock()
		for i, existing := range eb.handlers {
			// In Go, comparing functions is tricky.
			// For this simple bus, we'll just filter out the first match.
			if fmt.Sprintf("%v", existing) == fmt.Sprintf("%v", h) {
				eb.handlers = append(eb.handlers[:i], eb.handlers[i+1:]...)
				return
			}
		}
	}
}

// Emit publishes an event to all subscribers and stores it in the ring buffer.
func (eb *EventBus) Emit(evt TimelineEvent) {
	eb.mu.Lock()

	// Ring-buffer insert
	eb.history[eb.head] = evt
	eb.head = (eb.head + 1) % eb.capacity
	if eb.head == 0 {
		eb.full = true
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

	size := eb.head
	if eb.full {
		size = eb.capacity
	}

	if n <= 0 || n > size {
		n = size
	}

	out := make([]TimelineEvent, n)
	for i := 0; i < n; i++ {
		// Newest is at (head - 1 - i) mod capacity
		idx := (eb.head - 1 - i)
		if idx < 0 {
			idx += eb.capacity
		}
		out[i] = eb.history[idx]
	}
	return out
}

// All returns all stored events oldest first.
func (eb *EventBus) All() []TimelineEvent {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	size := eb.head
	if eb.full {
		size = eb.capacity
	}

	out := make([]TimelineEvent, size)
	if !eb.full {
		copy(out, eb.history[:eb.head])
	} else {
		// [head...capacity-1] [0...head-1]
		copy(out, eb.history[eb.head:])
		copy(out[eb.capacity-eb.head:], eb.history[:eb.head])
	}
	return out
}

// FilterByCategory returns events matching the given category.
func (eb *EventBus) FilterByCategory(cat EventCategory, n int) []TimelineEvent {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	size := eb.head
	if eb.full {
		size = eb.capacity
	}

	var filtered []TimelineEvent
	for i := 0; i < size && len(filtered) < n; i++ {
		// Traverse newest to oldest
		idx := (eb.head - 1 - i)
		if idx < 0 {
			idx += eb.capacity
		}
		if eb.history[idx].Category == cat {
			filtered = append(filtered, eb.history[idx])
		}
	}
	return filtered
}

// ── UUID Helper ───────────────────────────────────────────────────────────────

// NewUUID returns a random 16-byte UUID v4 hex string.
func NewUUID() string {
	b := make([]byte, 16)
	_ = rand.Read(b)
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
	// SANITIZATION: Scrub potential secrets from detail and metadata before storage
	scrubbedDetail := ScrubSecrets(detail)
	scrubbedMeta := make(map[string]string)
	for k, v := range meta {
		scrubbedMeta[k] = ScrubSecrets(v)
	}

	e := NewEvent(category, level, module, title, scrubbedDetail)
	e.Metadata = scrubbedMeta
	return e
}

// ScrubSecrets redacts sensitive patterns (passwords, keys) using regex.
func ScrubSecrets(s string) string {
	if s == "" {
		return ""
	}

	// 1. Redact common CLI password flags: -p, --password, -pass
	// Patterns: -p [pass], --password=[pass], --password [pass], -pass=[pass]
	pwdFlags := regexp.MustCompile(`(?i)(-(?:-password|pass|p))(?:\s*=| +)([^\s]+)`)
	s = pwdFlags.ReplaceAllString(s, "$1 [REDACTED]")

	// 2. Redact common KV patterns: password=xxx, pwd=xxx, api_key=xxx
	kvPatterns := regexp.MustCompile(`(?i)(password|pwd|pass|api_key|secret|token|apikey)(?:\s*[:=]\s*)([^\s,"'\}]+)`)
	s = kvPatterns.ReplaceAllString(s, "$1=[REDACTED]")

	// 3. Redact common URL auth
	urlAuth := regexp.MustCompile(`(?i)(://)([^:/]+):([^@/]+)(@)`)
	s = urlAuth.ReplaceAllString(s, "$1$2:[REDACTED]$4")

	return s
}
