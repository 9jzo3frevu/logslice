package routing_test

import (
	"testing"

	"github.com/logslice/logslice/internal/routing"
)

func TestNew_Valid(t *testing.T) {
	r, err := routing.New([]routing.Rule{
		{Field: "level", Value: "error", Sinks: []string{"pagerduty"}},
	}, []string{"default"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r == nil {
		t.Fatal("expected non-nil router")
	}
}

func TestNew_EmptyDefaultSinks(t *testing.T) {
	_, err := routing.New(nil, nil)
	if err == nil {
		t.Fatal("expected error for empty defaultSinks")
	}
}

func TestNew_EmptyRuleField(t *testing.T) {
	_, err := routing.New([]routing.Rule{
		{Field: "", Value: "error", Sinks: []string{"s1"}},
	}, []string{"default"})
	if err == nil {
		t.Fatal("expected error for empty rule field")
	}
}

func TestNew_EmptyRuleSinks(t *testing.T) {
	_, err := routing.New([]routing.Rule{
		{Field: "level", Value: "error", Sinks: nil},
	}, []string{"default"})
	if err == nil {
		t.Fatal("expected error for empty rule sinks")
	}
}

func TestMatch_FirstRuleWins(t *testing.T) {
	r, _ := routing.New([]routing.Rule{
		{Field: "level", Value: "error", Sinks: []string{"critical"}},
		{Field: "level", Value: "warn", Sinks: []string{"warnings"}},
	}, []string{"default"})

	got := r.Match(map[string]string{"level": "error"})
	if len(got) != 1 || got[0] != "critical" {
		t.Fatalf("expected [critical], got %v", got)
	}
}

func TestMatch_FallsBackToDefault(t *testing.T) {
	r, _ := routing.New([]routing.Rule{
		{Field: "level", Value: "error", Sinks: []string{"critical"}},
	}, []string{"default"})

	got := r.Match(map[string]string{"level": "info"})
	if len(got) != 1 || got[0] != "default" {
		t.Fatalf("expected [default], got %v", got)
	}
}

func TestMatch_CaseInsensitiveValue(t *testing.T) {
	r, _ := routing.New([]routing.Rule{
		{Field: "level", Value: "ERROR", Sinks: []string{"critical"}},
	}, []string{"default"})

	got := r.Match(map[string]string{"level": "error"})
	if len(got) != 1 || got[0] != "critical" {
		t.Fatalf("expected [critical], got %v", got)
	}
}

func TestMatch_NoRules(t *testing.T) {
	r, _ := routing.New(nil, []string{"fallback"})
	got := r.Match(map[string]string{"level": "debug"})
	if len(got) != 1 || got[0] != "fallback" {
		t.Fatalf("expected [fallback], got %v", got)
	}
}

func TestMatch_MissingField(t *testing.T) {
	r, _ := routing.New([]routing.Rule{
		{Field: "level", Value: "error", Sinks: []string{"critical"}},
	}, []string{"default"})

	// A log entry that does not contain the rule's field should fall back to default.
	got := r.Match(map[string]string{"service": "api"})
	if len(got) != 1 || got[0] != "default" {
		t.Fatalf("expected [default], got %v", got)
	}
}

func TestDefaultSinks(t *testing.T) {
	r, _ := routing.New(nil, []string{"a", "b"})
	ds := r.DefaultSinks()
	if len(ds) != 2 {
		t.Fatalf("expected 2 default sinks, got %d", len(ds))
	}
}
