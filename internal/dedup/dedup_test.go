package dedup

import (
	"testing"
	"time"
)

func TestNew_DefaultWindow(t *testing.T) {
	d := New(0)
	if d.WindowSize() != 5*time.Second {
		t.Fatalf("expected default window 5s, got %v", d.WindowSize())
	}
}

func TestNew_CustomWindow(t *testing.T) {
	d := New(10 * time.Second)
	if d.WindowSize() != 10*time.Second {
		t.Fatalf("expected 10s, got %v", d.WindowSize())
	}
}

func TestIsDuplicate_FirstSeen(t *testing.T) {
	d := New(5 * time.Second)
	if d.IsDuplicate("hello world", "info") {
		t.Fatal("first occurrence should not be a duplicate")
	}
}

func TestIsDuplicate_SecondSeen(t *testing.T) {
	d := New(5 * time.Second)
	d.IsDuplicate("hello world", "info")
	if !d.IsDuplicate("hello world", "info") {
		t.Fatal("second occurrence within window should be a duplicate")
	}
}

func TestIsDuplicate_DifferentLevel(t *testing.T) {
	d := New(5 * time.Second)
	d.IsDuplicate("hello world", "info")
	if d.IsDuplicate("hello world", "error") {
		t.Fatal("same message but different level should not be a duplicate")
	}
}

func TestIsDuplicate_AfterWindowExpires(t *testing.T) {
	base := time.Now()
	d := New(2 * time.Second)

	// Inject a controllable clock.
	d.now = func() time.Time { return base }
	d.IsDuplicate("expiring entry", "warn")

	// Advance clock beyond the window.
	d.now = func() time.Time { return base.Add(3 * time.Second) }
	if d.IsDuplicate("expiring entry", "warn") {
		t.Fatal("entry should have been evicted after window expired")
	}
}

func TestIsDuplicate_MultipleEntries(t *testing.T) {
	d := New(5 * time.Second)

	pairs := [][2]string{
		{"msg one", "info"},
		{"msg two", "debug"},
		{"msg three", "error"},
	}

	for _, p := range pairs {
		if d.IsDuplicate(p[0], p[1]) {
			t.Fatalf("first occurrence of (%s, %s) should not be duplicate", p[0], p[1])
		}
	}

	for _, p := range pairs {
		if !d.IsDuplicate(p[0], p[1]) {
			t.Fatalf("second occurrence of (%s, %s) should be duplicate", p[0], p[1])
		}
	}
}
