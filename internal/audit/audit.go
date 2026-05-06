// Package audit provides an append-only audit log for recording
// pipeline processing events such as received, filtered, forwarded,
// and dropped log entries.
package audit

import (
	"encoding/json"
	"io"
	"sync"
	"time"
)

// EventType classifies what happened to a log entry.
type EventType string

const (
	EventReceived  EventType = "received"
	EventForwarded EventType = "forwarded"
	EventFiltered  EventType = "filtered"
	EventDropped   EventType = "dropped"
)

// Event represents a single audit record.
type Event struct {
	Timestamp time.Time `json:"timestamp"`
	Type      EventType `json:"type"`
	Sink      string    `json:"sink,omitempty"`
	Message   string    `json:"message,omitempty"`
	Error     string    `json:"error,omitempty"`
}

// Logger writes audit events to an io.Writer as newline-delimited JSON.
type Logger struct {
	mu  sync.Mutex
	out io.Writer
}

// New returns a new audit Logger that writes to w.
func New(w io.Writer) *Logger {
	return &Logger{out: w}
}

// Record encodes and writes a single audit Event.
func (l *Logger) Record(e Event) error {
	if e.Timestamp.IsZero() {
		e.Timestamp = time.Now().UTC()
	}
	b, err := json.Marshal(e)
	if err != nil {
		return err
	}
	b = append(b, '\n')
	l.mu.Lock()
	defer l.mu.Unlock()
	_, err = l.out.Write(b)
	return err
}

// Received is a convenience wrapper for EventReceived.
func (l *Logger) Received(msg string) error {
	return l.Record(Event{Type: EventReceived, Message: msg})
}

// Forwarded is a convenience wrapper for EventForwarded.
func (l *Logger) Forwarded(sink, msg string) error {
	return l.Record(Event{Type: EventForwarded, Sink: sink, Message: msg})
}

// Filtered is a convenience wrapper for EventFiltered.
func (l *Logger) Filtered(msg string) error {
	return l.Record(Event{Type: EventFiltered, Message: msg})
}

// Dropped is a convenience wrapper for EventDropped.
func (l *Logger) Dropped(msg, errMsg string) error {
	return l.Record(Event{Type: EventDropped, Message: msg, Error: errMsg})
}
