package correlation

import (
	"context"
	"testing"
)

func TestNew_DefaultHeader(t *testing.T) {
	c, err := New("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Header() != DefaultHeader {
		t.Errorf("expected %q, got %q", DefaultHeader, c.Header())
	}
}

func TestNew_CustomHeader(t *testing.T) {
	c, err := New("X-Request-ID")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Header() != "X-Request-ID" {
		t.Errorf("expected X-Request-ID, got %q", c.Header())
	}
}

func TestGenerate_UniqueIDs(t *testing.T) {
	a, err := Generate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	b, err := Generate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a == b {
		t.Error("expected unique IDs, got duplicates")
	}
	if len(a) != 16 {
		t.Errorf("expected 16 hex chars, got %d", len(a))
	}
}

func TestContextRoundtrip(t *testing.T) {
	ctx := context.Background()
	if id := FromContext(ctx); id != "" {
		t.Errorf("expected empty, got %q", id)
	}
	ctx = WithContext(ctx, "abc123")
	if id := FromContext(ctx); id != "abc123" {
		t.Errorf("expected abc123, got %q", id)
	}
}

func TestInject_AddsField(t *testing.T) {
	c, _ := New("")
	entry := map[string]interface{}{"message": "hello"}
	c.Inject(entry, "test-id")
	if entry["correlation_id"] != "test-id" {
		t.Errorf("expected test-id, got %v", entry["correlation_id"])
	}
}

func TestInject_DoesNotOverwrite(t *testing.T) {
	c, _ := New("")
	entry := map[string]interface{}{"correlation_id": "existing"}
	c.Inject(entry, "new-id")
	if entry["correlation_id"] != "existing" {
		t.Errorf("expected existing, got %v", entry["correlation_id"])
	}
}
