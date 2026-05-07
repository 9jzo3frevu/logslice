package redact_test

import (
	"testing"

	"github.com/yourorg/logslice/internal/redact"
)

func TestNew_Valid(t *testing.T) {
	r, err := redact.New([]string{"password", "token"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r == nil {
		t.Fatal("expected non-nil Redactor")
	}
}

func TestNew_NoFields(t *testing.T) {
	_, err := redact.New(nil)
	if err == nil {
		t.Fatal("expected error for empty fields")
	}
}

func TestNew_EmptyFieldName(t *testing.T) {
	_, err := redact.New([]string{"token", ""})
	if err == nil {
		t.Fatal("expected error for empty field name")
	}
}

func TestApply_RedactsTargetedFields(t *testing.T) {
	r, _ := redact.New([]string{"password", "secret"})
	entry := map[string]any{
		"message":  "user login",
		"password": "hunter2",
		"secret":   "abc123",
		"user":     "alice",
	}
	r.Apply(entry)

	if entry["password"] != "[REDACTED]" {
		t.Errorf("expected password to be redacted, got %v", entry["password"])
	}
	if entry["secret"] != "[REDACTED]" {
		t.Errorf("expected secret to be redacted, got %v", entry["secret"])
	}
	if entry["message"] != "user login" {
		t.Errorf("expected message to be unchanged, got %v", entry["message"])
	}
	if entry["user"] != "alice" {
		t.Errorf("expected user to be unchanged, got %v", entry["user"])
	}
}

func TestApply_CaseInsensitiveKey(t *testing.T) {
	r, _ := redact.New([]string{"Authorization"})
	entry := map[string]any{
		"authorization": "Bearer xyz",
	}
	r.Apply(entry)
	if entry["authorization"] != "[REDACTED]" {
		t.Errorf("expected case-insensitive redaction, got %v", entry["authorization"])
	}
}

func TestApply_CustomMask(t *testing.T) {
	r, _ := redact.New([]string{"token"}, redact.WithMask("***"))
	entry := map[string]any{"token": "secret-value"}
	r.Apply(entry)
	if entry["token"] != "***" {
		t.Errorf("expected custom mask, got %v", entry["token"])
	}
}

func TestApply_NonStringValueRedacted(t *testing.T) {
	r, _ := redact.New([]string{"pin"})
	entry := map[string]any{"pin": 1234}
	r.Apply(entry)
	if entry["pin"] != "[REDACTED]" {
		t.Errorf("expected numeric field to be redacted, got %v", entry["pin"])
	}
}

func TestFields_ReturnsList(t *testing.T) {
	r, _ := redact.New([]string{"password"})
	fields := r.Fields()
	if len(fields) != 1 {
		t.Errorf("expected 1 field, got %d", len(fields))
	}
}
