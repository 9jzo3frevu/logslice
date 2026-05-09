package quota

import (
	"testing"
	"time"
)

func newLimiter(t *testing.T, cfg Config) *Limiter {
	t.Helper()
	l, err := New(cfg)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return l
}

func TestNew_Defaults(t *testing.T) {
	l, err := New(Config{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if l.cfg.MaxBytes != 10*1024*1024 {
		t.Errorf("MaxBytes default: got %d", l.cfg.MaxBytes)
	}
	if l.cfg.MaxEntries != 10_000 {
		t.Errorf("MaxEntries default: got %d", l.cfg.MaxEntries)
	}
	if l.cfg.Window != time.Minute {
		t.Errorf("Window default: got %v", l.cfg.Window)
	}
}

func TestAllow_WithinLimits(t *testing.T) {
	l := newLimiter(t, Config{MaxBytes: 100, MaxEntries: 5, Window: time.Minute})
	if err := l.Allow("svc-a", 10); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestAllow_ExceedsEntryLimit(t *testing.T) {
	l := newLimiter(t, Config{MaxBytes: 1_000_000, MaxEntries: 2, Window: time.Minute})
	_ = l.Allow("svc", 1)
	_ = l.Allow("svc", 1)
	if err := l.Allow("svc", 1); err != ErrQuotaExceeded {
		t.Fatalf("expected ErrQuotaExceeded, got %v", err)
	}
}

func TestAllow_ExceedsByteLimit(t *testing.T) {
	l := newLimiter(t, Config{MaxBytes: 20, MaxEntries: 1000, Window: time.Minute})
	_ = l.Allow("svc", 15)
	if err := l.Allow("svc", 10); err != ErrQuotaExceeded {
		t.Fatalf("expected ErrQuotaExceeded, got %v", err)
	}
}

func TestAllow_DifferentSources_Independent(t *testing.T) {
	l := newLimiter(t, Config{MaxBytes: 100, MaxEntries: 1, Window: time.Minute})
	_ = l.Allow("a", 10)
	if err := l.Allow("b", 10); err != nil {
		t.Fatalf("sources should be independent, got %v", err)
	}
}

func TestAllow_ResetsAfterWindow(t *testing.T) {
	l := newLimiter(t, Config{MaxBytes: 100, MaxEntries: 1, Window: 10 * time.Millisecond})
	_ = l.Allow("svc", 10)
	time.Sleep(20 * time.Millisecond)
	if err := l.Allow("svc", 10); err != nil {
		t.Fatalf("counter should have reset, got %v", err)
	}
}

func TestReset_ClearsSource(t *testing.T) {
	l := newLimiter(t, Config{MaxBytes: 100, MaxEntries: 1, Window: time.Minute})
	_ = l.Allow("svc", 10)
	l.Reset("svc")
	if err := l.Allow("svc", 10); err != nil {
		t.Fatalf("after Reset expected nil, got %v", err)
	}
}
