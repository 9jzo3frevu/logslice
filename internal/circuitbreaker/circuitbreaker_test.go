package circuitbreaker

import (
	"testing"
	"time"
)

func TestNew_Defaults(t *testing.T) {
	b := New(0, 0)
	if b.threshold != 3 {
		t.Fatalf("expected default threshold 3, got %d", b.threshold)
	}
	if b.resetTimeout != 30*time.Second {
		t.Fatalf("expected default reset timeout 30s, got %v", b.resetTimeout)
	}
}

func TestAllow_ClosedByDefault(t *testing.T) {
	b := New(3, time.Second)
	if err := b.Allow(); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestOpensAfterThreshold(t *testing.T) {
	b := New(3, time.Minute)
	for i := 0; i < 3; i++ {
		b.RecordFailure()
	}
	if b.CurrentState() != StateOpen {
		t.Fatal("expected state Open")
	}
	if err := b.Allow(); err != ErrOpen {
		t.Fatalf("expected ErrOpen, got %v", err)
	}
}

func TestHalfOpenAfterTimeout(t *testing.T) {
	b := New(1, 10*time.Millisecond)
	b.RecordFailure()
	time.Sleep(20 * time.Millisecond)
	if err := b.Allow(); err != nil {
		t.Fatalf("expected nil in half-open, got %v", err)
	}
	if b.CurrentState() != StateHalfOpen {
		t.Fatal("expected HalfOpen state")
	}
}

func TestRecordSuccess_CloseFromHalfOpen(t *testing.T) {
	b := New(1, 10*time.Millisecond)
	b.RecordFailure()
	time.Sleep(20 * time.Millisecond)
	_ = b.Allow() // transitions to HalfOpen
	b.RecordSuccess()
	if b.CurrentState() != StateClosed {
		t.Fatal("expected Closed after success")
	}
}

func TestRecordFailure_ReopensFromHalfOpen(t *testing.T) {
	b := New(1, 10*time.Millisecond)
	b.RecordFailure()
	time.Sleep(20 * time.Millisecond)
	_ = b.Allow()
	b.RecordFailure()
	if b.CurrentState() != StateOpen {
		t.Fatal("expected Open after failure in HalfOpen")
	}
}

func TestRecordSuccess_ResetFailures(t *testing.T) {
	b := New(3, time.Minute)
	b.RecordFailure()
	b.RecordFailure()
	b.RecordSuccess()
	if b.failures != 0 {
		t.Fatalf("expected failures reset to 0, got %d", b.failures)
	}
	if b.CurrentState() != StateClosed {
		t.Fatal("expected Closed")
	}
}
