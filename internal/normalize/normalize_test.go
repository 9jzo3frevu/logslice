package normalize

import (
	"testing"
)

func TestNew_Valid(t *testing.T) {
	_, err := New([]Mapping{
		{From: "msg", To: "message"},
		{From: "lvl", To: "level"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNew_EmptyFrom(t *testing.T) {
	_, err := New([]Mapping{{From: "", To: "message"}})
	if err == nil {
		t.Fatal("expected error for empty 'from' field")
	}
}

func TestNew_EmptyTo(t *testing.T) {
	_, err := New([]Mapping{{From: "msg", To: ""}})
	if err == nil {
		t.Fatal("expected error for empty 'to' field")
	}
}

func TestApply_RenamesKey(t *testing.T) {
	n, _ := New([]Mapping{{From: "msg", To: "message"}})
	entry := map[string]any{"msg": "hello", "level": "info"}
	n.Apply(entry)

	if _, ok := entry["msg"]; ok {
		t.Error("original key 'msg' should have been removed")
	}
	if entry["message"] != "hello" {
		t.Errorf("expected entry[message]='hello', got %v", entry["message"])
	}
}

func TestApply_SkipsMissingSource(t *testing.T) {
	n, _ := New([]Mapping{{From: "msg", To: "message"}})
	entry := map[string]any{"level": "warn"}
	n.Apply(entry)

	if _, ok := entry["message"]; ok {
		t.Error("'message' key should not have been created")
	}
}

func TestApply_DoesNotOverwriteExistingDest(t *testing.T) {
	n, _ := New([]Mapping{{From: "msg", To: "message"}})
	entry := map[string]any{"msg": "alias", "message": "canonical"}
	n.Apply(entry)

	if entry["message"] != "canonical" {
		t.Errorf("existing 'message' value should be preserved, got %v", entry["message"])
	}
	if entry["msg"] != "alias" {
		t.Errorf("source key 'msg' should remain when dest exists, got %v", entry["msg"])
	}
}

func TestApply_MultipleRules(t *testing.T) {
	n, _ := New([]Mapping{
		{From: "msg", To: "message"},
		{From: "lvl", To: "level"},
	})
	entry := map[string]any{"msg": "hi", "lvl": "debug"}
	n.Apply(entry)

	if entry["message"] != "hi" {
		t.Errorf("expected message='hi', got %v", entry["message"])
	}
	if entry["level"] != "debug" {
		t.Errorf("expected level='debug', got %v", entry["level"])
	}
}

func TestMappings_ReturnsCopy(t *testing.T) {
	orig := []Mapping{{From: "msg", To: "message"}}
	n, _ := New(orig)
	copy1 := n.Mappings()
	copy1[0].From = "mutated"

	if n.Mappings()[0].From != "msg" {
		t.Error("Mappings should return an independent copy")
	}
}
