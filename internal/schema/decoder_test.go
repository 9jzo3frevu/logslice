package schema_test

import (
	"strings"
	"testing"

	"github.com/logslice/logslice/internal/schema"
)

func TestDecoder_ValidEntry(t *testing.T) {
	body := `{"level":"info","message":"hello","service":"svc"}`
	d := schema.NewDecoder(strings.NewReader(body))
	e, err := d.Decode()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if e.Message != "hello" {
		t.Errorf("expected message 'hello', got %q", e.Message)
	}
	if e.Service != "svc" {
		t.Errorf("expected service 'svc', got %q", e.Service)
	}
	if e.Timestamp.IsZero() {
		t.Error("expected timestamp to be set")
	}
}

func TestDecoder_InvalidJSON(t *testing.T) {
	d := schema.NewDecoder(strings.NewReader(`not json`))
	_, err := d.Decode()
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestDecoder_MissingRequiredField(t *testing.T) {
	body := `{"level":"error"}`
	d := schema.NewDecoder(strings.NewReader(body))
	_, err := d.Decode()
	if err == nil {
		t.Fatal("expected error for missing message")
	}
}

func TestDecoder_MultipleEntries(t *testing.T) {
	body := `{"level":"info","message":"first"}
{"level":"warn","message":"second"}`
	d := schema.NewDecoder(strings.NewReader(body))

	for i, want := range []string{"first", "second"} {
		e, err := d.Decode()
		if err != nil {
			t.Fatalf("entry %d: unexpected error: %v", i, err)
		}
		if e.Message != want {
			t.Errorf("entry %d: expected %q, got %q", i, want, e.Message)
		}
	}
}
