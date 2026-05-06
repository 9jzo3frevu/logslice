package schema_test

import (
	"testing"
	"time"

	"github.com/logslice/logslice/internal/schema"
)

func TestValidate_Valid(t *testing.T) {
	e := &schema.Entry{Level: "info", Message: "hello"}
	if err := schema.Validate(e); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidate_Nil(t *testing.T) {
	if err := schema.Validate(nil); err == nil {
		t.Fatal("expected error for nil entry")
	}
}

func TestValidate_MissingMessage(t *testing.T) {
	e := &schema.Entry{Level: "info"}
	if err := schema.Validate(e); err == nil {
		t.Fatal("expected error for missing message")
	}
}

func TestValidate_MissingLevel(t *testing.T) {
	e := &schema.Entry{Message: "hello"}
	if err := schema.Validate(e); err == nil {
		t.Fatal("expected error for missing level")
	}
}

func TestNormalize_SetsTimestamp(t *testing.T) {
	e := &schema.Entry{Level: "warn", Message: "test"}
	if err := schema.Normalize(e); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if e.Timestamp.IsZero() {
		t.Error("expected timestamp to be set")
	}
}

func TestNormalize_PreservesTimestamp(t *testing.T) {
	ts := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	e := &schema.Entry{Level: "info", Message: "hi", Timestamp: ts}
	_ = schema.Normalize(e)
	if !e.Timestamp.Equal(ts) {
		t.Errorf("expected %v, got %v", ts, e.Timestamp)
	}
}

func TestNormalize_InitializesFields(t *testing.T) {
	e := &schema.Entry{Level: "debug", Message: "msg"}
	_ = schema.Normalize(e)
	if e.Fields == nil {
		t.Error("expected Fields map to be initialized")
	}
}

func TestNormalize_InvalidEntry(t *testing.T) {
	e := &schema.Entry{Level: "info"} // missing message
	if err := schema.Normalize(e); err == nil {
		t.Fatal("expected error for invalid entry")
	}
}
