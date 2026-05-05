package filter

import (
	"testing"
)

func TestParseLevel(t *testing.T) {
	cases := []struct {
		input string
		want  Level
		ok    bool
	}{
		{"debug", LevelDebug, true},
		{"INFO", LevelInfo, true},
		{"Warn", LevelWarn, true},
		{"ERROR", LevelError, true},
		{"unknown", LevelDebug, false},
		{"", LevelDebug, false},
	}
	for _, c := range cases {
		got, ok := ParseLevel(c.input)
		if ok != c.ok || got != c.want {
			t.Errorf("ParseLevel(%q) = (%v, %v); want (%v, %v)", c.input, got, ok, c.want, c.ok)
		}
	}
}

func TestFilter_Allow(t *testing.T) {
	f := New("warn")

	allowed := []Entry{
		{Level: "warn", Message: "disk almost full"},
		{Level: "error", Message: "connection refused"},
		{Level: "WARN", Message: "slow query"},
	}
	for _, e := range allowed {
		if !f.Allow(e) {
			t.Errorf("expected Allow(%+v) = true", e)
		}
	}

	blocked := []Entry{
		{Level: "debug", Message: "verbose trace"},
		{Level: "info", Message: "server started"},
	}
	for _, e := range blocked {
		if f.Allow(e) {
			t.Errorf("expected Allow(%+v) = false", e)
		}
	}
}

func TestFilter_Allow_UnknownLevel(t *testing.T) {
	f := New("error")
	e := Entry{Level: "critical", Message: "unknown severity"}
	if !f.Allow(e) {
		t.Error("expected unknown level to be allowed through")
	}
}

func TestTag(t *testing.T) {
	e := Entry{
		Level:   "info",
		Message: "hello",
		Tags:    map[string]string{"app": "logslice", "env": "dev"},
	}
	extra := map[string]string{"env": "prod", "region": "us-east-1"}
	tagged := Tag(e, extra)

	if tagged.Tags["env"] != "prod" {
		t.Errorf("expected env=prod, got %s", tagged.Tags["env"])
	}
	if tagged.Tags["app"] != "logslice" {
		t.Errorf("expected app=logslice, got %s", tagged.Tags["app"])
	}
	if tagged.Tags["region"] != "us-east-1" {
		t.Errorf("expected region=us-east-1, got %s", tagged.Tags["region"])
	}
	// Original entry must not be mutated.
	if e.Tags["env"] != "dev" {
		t.Error("original entry tags were mutated")
	}
}
