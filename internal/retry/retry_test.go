package retry_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/your-org/logslice/internal/retry"
)

var errTemp = errors.New("temporary error")

func TestDo_SuccessOnFirstAttempt(t *testing.T) {
	p := retry.DefaultPolicy()
	calls := 0
	err := p.Do(context.Background(), func() error {
		calls++
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}
}

func TestDo_RetriesAndSucceeds(t *testing.T) {
	p := retry.Policy{MaxAttempts: 3, BaseDelay: time.Millisecond, MaxDelay: 10 * time.Millisecond}
	calls := 0
	err := p.Do(context.Background(), func() error {
		calls++
		if calls < 3 {
			return errTemp
		}
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil after retries, got %v", err)
	}
	if calls != 3 {
		t.Fatalf("expected 3 calls, got %d", calls)
	}
}

func TestDo_ExhaustsAttempts(t *testing.T) {
	p := retry.Policy{MaxAttempts: 2, BaseDelay: time.Millisecond, MaxDelay: 10 * time.Millisecond}
	calls := 0
	err := p.Do(context.Background(), func() error {
		calls++
		return errTemp
	})
	if !errors.Is(err, retry.ErrExhausted) {
		t.Fatalf("expected ErrExhausted, got %v", err)
	}
	if !errors.Is(err, errTemp) {
		t.Fatalf("expected wrapped errTemp, got %v", err)
	}
	if calls != 2 {
		t.Fatalf("expected 2 calls, got %d", calls)
	}
}

func TestDo_ContextCancelled(t *testing.T) {
	p := retry.Policy{MaxAttempts: 5, BaseDelay: 50 * time.Millisecond, MaxDelay: time.Second}
	ctx, cancel := context.WithCancel(context.Background())
	calls := 0
	err := p.Do(ctx, func() error {
		calls++
		cancel() // cancel after first failure
		return errTemp
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestDefaultPolicy(t *testing.T) {
	p := retry.DefaultPolicy()
	if p.MaxAttempts != 3 {
		t.Errorf("expected MaxAttempts=3, got %d", p.MaxAttempts)
	}
	if p.BaseDelay != 100*time.Millisecond {
		t.Errorf("unexpected BaseDelay: %v", p.BaseDelay)
	}
}
