package masking_test

import (
	"testing"

	"github.com/yourorg/logslice/internal/masking"
)

func TestNew_Valid(t *testing.T) {
	_, err := masking.New([]masking.Config{
		{Field: "email", Pattern: `[^@]+@`, Replacement: "***@"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNew_NoRules(t *testing.T) {
	_, err := masking.New(nil)
	if err == nil {
		t.Fatal("expected error for empty rules")
	}
}

func TestNew_EmptyField(t *testing.T) {
	_, err := masking.New([]masking.Config{
		{Field: "", Pattern: `\d+`, Replacement: "***"},
	})
	if err == nil {
		t.Fatal("expected error for empty field name")
	}
}

func TestNew_InvalidPattern(t *testing.T) {
	_, err := masking.New([]masking.Config{
		{Field: "token", Pattern: `[invalid`, Replacement: "***"},
	})
	if err == nil {
		t.Fatal("expected error for invalid regex")
	}
}

func TestApply_MasksMatchingField(t *testing.T) {
	m, _ := masking.New([]masking.Config{
		{Field: "email", Pattern: `[^@]+@`, Replacement: "***@"},
	})
	entry := map[string]any{"email": "user@example.com", "level": "info"}
	out := m.Apply(entry)
	if got := out["email"]; got != "***@example.com" {
		t.Errorf("expected ***@example.com, got %v", got)
	}
	if out["level"] != "info" {
		t.Error("unrelated field should be unchanged")
	}
}

func TestApply_SkipsMissingField(t *testing.T) {
	m, _ := masking.New([]masking.Config{
		{Field: "ssn", Pattern: `\d{4}$`, Replacement: "****"},
	})
	entry := map[string]any{"message": "hello"}
	out := m.Apply(entry)
	if _, ok := out["ssn"]; ok {
		t.Error("ssn should not be added to entry")
	}
}

func TestApply_SkipsNonStringValue(t *testing.T) {
	m, _ := masking.New([]masking.Config{
		{Field: "count", Pattern: `\d+`, Replacement: "***"},
	})
	entry := map[string]any{"count": 42}
	out := m.Apply(entry)
	if out["count"] != 42 {
		t.Errorf("non-string value should be unchanged, got %v", out["count"])
	}
}

func TestApply_MultipleRules(t *testing.T) {
	m, _ := masking.New([]masking.Config{
		{Field: "card", Pattern: `\d{12}(\d{4})`, Replacement: "************$1"},
		{Field: "phone", Pattern: `\d{7}(\d{4})`, Replacement: "*******$1"},
	})
	entry := map[string]any{
		"card":  "1234567890121234",
		"phone": "5551234567",
	}
	out := m.Apply(entry)
	if out["card"] != "************1234" {
		t.Errorf("unexpected card value: %v", out["card"])
	}
	if out["phone"] != "*******4567" {
		t.Errorf("unexpected phone value: %v", out["phone"])
	}
}
