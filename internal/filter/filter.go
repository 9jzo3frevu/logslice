package filter

import (
	"strings"
)

// Level represents a log severity level.
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

var levelNames = map[string]Level{
	"debug": LevelDebug,
	"info":  LevelInfo,
	"warn":  LevelWarn,
	"error": LevelError,
}

// ParseLevel converts a string level to a Level value.
// Returns LevelDebug and false if the level is unrecognised.
func ParseLevel(s string) (Level, bool) {
	l, ok := levelNames[strings.ToLower(s)]
	return l, ok
}

// Entry represents a single structured log entry.
type Entry struct {
	Level   string            `json:"level"`
	Message string            `json:"message"`
	Tags    map[string]string `json:"tags,omitempty"`
}

// Filter decides which log entries should be forwarded.
type Filter struct {
	minLevel Level
}

// New creates a Filter that passes entries at or above minLevelStr.
// If minLevelStr is empty or unrecognised, LevelDebug is used.
func New(minLevelStr string) *Filter {
	l, _ := ParseLevel(minLevelStr)
	return &Filter{minLevel: l}
}

// Allow returns true when the entry's level meets the minimum threshold.
func (f *Filter) Allow(e Entry) bool {
	l, ok := ParseLevel(e.Level)
	if !ok {
		// Unknown levels are always forwarded.
		return true
	}
	return l >= f.minLevel
}

// Tag merges the provided extra tags into the entry, returning a new Entry.
// Existing tags on the entry are preserved; extra tags take precedence on
// key collisions.
func Tag(e Entry, extra map[string]string) Entry {
	merged := make(map[string]string, len(e.Tags)+len(extra))
	for k, v := range e.Tags {
		merged[k] = v
	}
	for k, v := range extra {
		merged[k] = v
	}
	e.Tags = merged
	return e
}
