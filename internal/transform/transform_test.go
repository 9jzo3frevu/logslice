package transform

import (
	"testing"
)

func TestNew_Valid(t *testing.T) {
	rules := []Rule{
		{Op: "add", Field: "env", Value: "prod"},
		{Op: "remove", Field: "secret"},
		{Op: "rename", Field: "msg", NewField: "message"},
	}
	tr, err := New(rules)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tr == nil {
		t.Fatal("expected non-nil Transformer")
	}
}

func TestNew_InvalidOp(t *testing.T) {
	_, err := New([]Rule{{Op: "upsert", Field: "x"}})
	if err == nil {
		t.Fatal("expected error for unknown op")
	}
}

func TestNew_AddMissingField(t *testing.T) {
	_, err := New([]Rule{{Op: "add", Value: "v"}})
	if err == nil {
		t.Fatal("expected error when add has no field")
	}
}

func TestNew_RenameMissingNewField(t *testing.T) {
	_, err := New([]Rule{{Op: "rename", Field: "old"}})
	if err == nil {
		t.Fatal("expected error when rename has no new_field")
	}
}

func TestApply_Add(t *testing.T) {
	tr, _ := New([]Rule{{Op: "add", Field: "env", Value: "staging"}})
	out := tr.Apply(map[string]any{"level": "info"})
	if out["env"] != "staging" {
		t.Errorf("expected env=staging, got %v", out["env"])
	}
	if out["level"] != "info" {
		t.Error("original field should be preserved")
	}
}

func TestApply_Remove(t *testing.T) {
	tr, _ := New([]Rule{{Op: "remove", Field: "secret"}})
	out := tr.Apply(map[string]any{"level": "warn", "secret": "abc"})
	if _, ok := out["secret"]; ok {
		t.Error("secret field should have been removed")
	}
}

func TestApply_Rename(t *testing.T) {
	tr, _ := New([]Rule{{Op: "rename", Field: "msg", NewField: "message"}})
	out := tr.Apply(map[string]any{"msg": "hello"})
	if out["message"] != "hello" {
		t.Errorf("expected message=hello, got %v", out["message"])
	}
	if _, ok := out["msg"]; ok {
		t.Error("old field name should be absent after rename")
	}
}

func TestApply_RenameAbsentField(t *testing.T) {
	tr, _ := New([]Rule{{Op: "rename", Field: "missing", NewField: "present"}})
	out := tr.Apply(map[string]any{"level": "debug"})
	if _, ok := out["present"]; ok {
		t.Error("present should not exist when source field is absent")
	}
}

func TestApply_DoesNotMutateOriginal(t *testing.T) {
	tr, _ := New([]Rule{{Op: "add", Field: "env", Value: "test"}})
	orig := map[string]any{"level": "info"}
	tr.Apply(orig)
	if _, ok := orig["env"]; ok {
		t.Error("original map should not be mutated")
	}
}
