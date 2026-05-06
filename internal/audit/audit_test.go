package audit

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestRecord_WritesJSON(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf)

	now := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	err := l.Record(Event{
		Timestamp: now,
		Type:      EventReceived,
		Message:   "hello",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var got Event
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}
	if got.Type != EventReceived {
		t.Errorf("expected type %q, got %q", EventReceived, got.Type)
	}
	if got.Message != "hello" {
		t.Errorf("expected message %q, got %q", "hello", got.Message)
	}
}

func TestRecord_SetsTimestampIfZero(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf)

	before := time.Now().UTC()
	_ = l.Record(Event{Type: EventForwarded})
	after := time.Now().UTC()

	var got Event
	_ = json.Unmarshal(buf.Bytes(), &got)
	if got.Timestamp.Before(before) || got.Timestamp.After(after) {
		t.Errorf("timestamp %v not in expected range [%v, %v]", got.Timestamp, before, after)
	}
}

func TestConvenienceWrappers(t *testing.T) {
	tests := []struct {
		name     string
		fn       func(*Logger) error
		wantType EventType
	}{
		{"Received", func(l *Logger) error { return l.Received("msg") }, EventReceived},
		{"Forwarded", func(l *Logger) error { return l.Forwarded("s1", "msg") }, EventForwarded},
		{"Filtered", func(l *Logger) error { return l.Filtered("msg") }, EventFiltered},
		{"Dropped", func(l *Logger) error { return l.Dropped("msg", "err") }, EventDropped},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			l := New(&buf)
			if err := tc.fn(l); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			var got Event
			_ = json.Unmarshal(buf.Bytes(), &got)
			if got.Type != tc.wantType {
				t.Errorf("expected type %q, got %q", tc.wantType, got.Type)
			}
		})
	}
}

func TestRecord_NewlineDelimited(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf)
	_ = l.Received("first")
	_ = l.Received("second")

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}
	for i, line := range lines {
		var e Event
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			t.Errorf("line %d is not valid JSON: %v", i, err)
		}
	}
}
