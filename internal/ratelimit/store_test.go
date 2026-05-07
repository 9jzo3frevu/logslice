package ratelimit

import (
	"testing"
	"time"
)

func TestKeyedLimiter_SeparateKeys(t *testing.T) {
	kl := NewKeyedLimiter(10, 1, time.Minute)

	if !kl.Allow("a") {
		t.Fatal("expected first Allow for key 'a' to succeed")
	}
	if !kl.Allow("b") {
		t.Fatal("expected first Allow for key 'b' to succeed")
	}
}

func TestKeyedLimiter_BlocksAfterBurst(t *testing.T) {
	kl := NewKeyedLimiter(0.001, 1, time.Minute)

	if !kl.Allow("x") {
		t.Fatal("expected first token to be available")
	}
	if kl.Allow("x") {
		t.Fatal("expected second Allow to be blocked after burst exhausted")
	}
}

func TestKeyedLimiter_IndependentBursts(t *testing.T) {
	kl := NewKeyedLimiter(0.001, 1, time.Minute)

	kl.Allow("p") // exhaust p
	if !kl.Allow("q") {
		t.Fatal("key 'q' should have its own independent burst")
	}
}

func TestKeyedLimiter_Len(t *testing.T) {
	kl := NewKeyedLimiter(10, 5, time.Minute)

	kl.Allow("one")
	kl.Allow("two")
	kl.Allow("three")

	if got := kl.Len(); got != 3 {
		t.Fatalf("expected Len 3, got %d", got)
	}
}

func TestKeyedLimiter_Eviction(t *testing.T) {
	ttl := 10 * time.Millisecond
	kl := NewKeyedLimiter(10, 5, ttl)

	kl.Allow("evictme")
	if kl.Len() != 1 {
		t.Fatalf("expected 1 entry before eviction")
	}

	time.Sleep(ttl + 5*time.Millisecond)

	// Trigger eviction via a new Allow call.
	kl.Allow("trigger")

	if kl.Len() != 1 {
		t.Fatalf("expected 1 entry after eviction (only 'trigger'), got %d", kl.Len())
	}
}
