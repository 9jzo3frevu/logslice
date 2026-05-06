package ratelimit_test

import (
	"testing"
	"time"

	"github.com/logslice/logslice/internal/ratelimit"
)

func TestNew_DefaultsToOne(t *testing.T) {
	l := ratelimit.New(0)
	if l.Rate() != 1 {
		t.Fatalf("expected rate 1, got %f", l.Rate())
	}
}

func TestAllow_WithinBurst(t *testing.T) {
	l := ratelimit.New(5)
	for i := 0; i < 5; i++ {
		if !l.Allow() {
			t.Fatalf("expected Allow()=true on call %d", i+1)
		}
	}
}

func TestAllow_ExceedsBurst(t *testing.T) {
	l := ratelimit.New(3)
	for i := 0; i < 3; i++ {
		l.Allow()
	}
	if l.Allow() {
		t.Fatal("expected Allow()=false after burst exhausted")
	}
}

func TestAllow_Refill(t *testing.T) {
	l := ratelimit.New(10)
	for i := 0; i < 10; i++ {
		l.Allow()
	}
	// Wait for tokens to refill.
	time.Sleep(200 * time.Millisecond)
	if !l.Allow() {
		t.Fatal("expected Allow()=true after refill window")
	}
}

func TestRate_ReturnsConfigured(t *testing.T) {
	l := ratelimit.New(42)
	if l.Rate() != 42 {
		t.Fatalf("expected rate 42, got %f", l.Rate())
	}
}
