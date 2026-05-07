package enrichment

import (
	"testing"
)

func TestNew_Valid(t *testing.T) {
	e, err := New(map[string]string{"env": "prod", "region": "us-east-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if e == nil {
		t.Fatal("expected non-nil enricher")
	}
}

func TestNew_NilFields(t *testing.T) {
	_, err := New(nil)
	if err == nil {
		t.Fatal("expected error for nil fields")
	}
}

func TestNew_EmptyKey(t *testing.T) {
	_, err := New(map[string]string{"": "value"})
	if err == nil {
		t.Fatal("expected error for empty key")
	}
}

func TestApply_AddsFields(t *testing.T) {
	e, _ := New(map[string]string{"env": "prod"})
	entry := map[string]any{"message": "hello", "level": "info"}
	out, err := e.Apply(entry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out["env"] != "prod" {
		t.Errorf("expected env=prod, got %v", out["env"])
	}
	if out["message"] != "hello" {
		t.Error("original fields should be preserved")
	}
}

func TestApply_OverwritesConflictingKeys(t *testing.T) {
	e, _ := New(map[string]string{"env": "prod"})
	entry := map[string]any{"env": "dev", "level": "warn"}
	out, _ := e.Apply(entry)
	if out["env"] != "prod" {
		t.Errorf("expected enricher value to win, got %v", out["env"])
	}
}

func TestApply_DoesNotMutateOriginal(t *testing.T) {
	e, _ := New(map[string]string{"env": "prod"})
	entry := map[string]any{"message": "test"}
	e.Apply(entry)
	if _, ok := entry["env"]; ok {
		t.Error("original entry should not be mutated")
	}
}

func TestApply_NilEntry(t *testing.T) {
	e, _ := New(map[string]string{"env": "prod"})
	_, err := e.Apply(nil)
	if err == nil {
		t.Fatal("expected error for nil entry")
	}
}

func TestFields_ReturnsCopy(t *testing.T) {
	e, _ := New(map[string]string{"env": "prod"})
	f := e.Fields()
	f["env"] = "mutated"
	if e.Fields()["env"] != "prod" {
		t.Error("Fields should return a defensive copy")
	}
}
