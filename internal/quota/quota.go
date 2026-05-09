// Package quota enforces per-source byte and entry quotas over a rolling
// time window, allowing logslice operators to cap the volume of logs
// accepted from any single named source.
package quota

import (
	"errors"
	"sync"
	"time"
)

// ErrQuotaExceeded is returned when a source has exhausted its allowance.
var ErrQuotaExceeded = errors.New("quota: source quota exceeded")

// Config holds the limits applied to every source unless overridden.
type Config struct {
	// MaxBytes is the maximum number of bytes accepted per window.
	MaxBytes int64
	// MaxEntries is the maximum number of log entries accepted per window.
	MaxEntries int64
	// Window is the rolling duration after which counters reset.
	Window time.Duration
}

type bucket struct {
	bytes   int64
	entries int64
	reset   time.Time
}

// Limiter tracks per-source usage and enforces configured quotas.
type Limiter struct {
	cfg     Config
	mu      sync.Mutex
	sources map[string]*bucket
}

// New creates a Limiter with the provided Config.
// Defaults: MaxBytes=10 MiB, MaxEntries=10 000, Window=1 minute.
func New(cfg Config) (*Limiter, error) {
	if cfg.MaxBytes <= 0 {
		cfg.MaxBytes = 10 * 1024 * 1024
	}
	if cfg.MaxEntries <= 0 {
		cfg.MaxEntries = 10_000
	}
	if cfg.Window <= 0 {
		cfg.Window = time.Minute
	}
	return &Limiter{
		cfg:     cfg,
		sources: make(map[string]*bucket),
	}, nil
}

// Allow records n bytes for source and returns ErrQuotaExceeded when either
// the byte or entry limit for the current window is breached.
func (l *Limiter) Allow(source string, n int64) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	b, ok := l.sources[source]
	if !ok || now.After(b.reset) {
		b = &bucket{reset: now.Add(l.cfg.Window)}
		l.sources[source] = b
	}

	if b.entries+1 > l.cfg.MaxEntries {
		return ErrQuotaExceeded
	}
	if b.bytes+n > l.cfg.MaxBytes {
		return ErrQuotaExceeded
	}

	b.bytes += n
	b.entries++
	return nil
}

// Reset clears the usage counters for source immediately.
func (l *Limiter) Reset(source string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.sources, source)
}
