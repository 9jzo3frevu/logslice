package schema

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestEncoder_ValidEntry(t *testing.T) {
	entry := &LogEntry{Message: "hello", Level: "info"}
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	if err := enc.Encode(entry); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	line := strings.TrimSpace(buf.String())
	var out LogEntry
	if err := json.Unmarshal([]byte(line), &out); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if out.Message != "hello" {
		t.Errorf("expected message 'hello', got %q", out.Message)
	}
	if out.Timestamp == "" {
		t.Error("expected timestamp to be set by Normalize")
	}
}

func TestEncoder_InvalidEntry(t *testing.T) {
	entry := &LogEntry{Level: "info"} // missing Message
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	if err := enc.Encode(entry); err == nil {
		t.Fatal("expected validation error, got nil")
	}
	if buf.Len() != 0 {
		t.Error("expected no bytes written on validation failure")
	}
}

func TestEncoder_MultipleEntries(t *testing.T) {
	entries := []*LogEntry{
		{Message: "first", Level: "debug"},
		{Message: "second", Level: "warn"},
	}
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	for _, e := range entries {
		if err := enc.Encode(e); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}
}

func TestEncodeToBytes_Valid(t *testing.T) {
	entry := &LogEntry{Message: "payload", Level: "error"}
	b, err := EncodeToBytes(entry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if bytes.Contains(b, []byte("\n")) {
		t.Error("EncodeToBytes should not include trailing newline")
	}
	var out LogEntry
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatalf("result is not valid JSON: %v", err)
	}
	if out.Message != "payload" {
		t.Errorf("expected message 'payload', got %q", out.Message)
	}
}

func TestEncodeToBytes_Invalid(t *testing.T) {
	entry := &LogEntry{Message: "no-level"} // missing Level
	_, err := EncodeToBytes(entry)
	if err == nil {
		t.Fatal("expected error for missing level")
	}
}
