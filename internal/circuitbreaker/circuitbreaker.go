// Package circuitbreaker provides a simple circuit breaker for sink forwarding.
// It transitions between Closed, Open, and HalfOpen states based on consecutive
// failures, preventing repeated attempts to an unhealthy downstream.
package circuitbreaker

import (
	"errors"
	"sync"
	"time"
)

// ErrOpen is returned when the circuit is open and requests are rejected.
var ErrOpen = errors.New("circuit breaker is open")

// State represents the circuit breaker state.
type State int

const (
	StateClosed   State = iota // normal operation
	StateOpen                  // failing; requests rejected
	StateHalfOpen              // probe request allowed
)

// Breaker is a circuit breaker instance.
type Breaker struct {
	mu           sync.Mutex
	state        State
	failures     int
	threshold    int
	resetTimeout time.Duration
	openedAt     time.Time
}

// New creates a Breaker that opens after threshold consecutive failures and
// attempts recovery after resetTimeout.
func New(threshold int, resetTimeout time.Duration) *Breaker {
	if threshold <= 0 {
		threshold = 3
	}
	if resetTimeout <= 0 {
		resetTimeout = 30 * time.Second
	}
	return &Breaker{
		threshold:    threshold,
		resetTimeout: resetTimeout,
	}
}

// Allow reports whether the caller may proceed. It returns ErrOpen when the
// circuit is open and the reset timeout has not elapsed.
func (b *Breaker) Allow() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	switch b.state {
	case StateClosed:
		return nil
	case StateOpen:
		if time.Since(b.openedAt) >= b.resetTimeout {
			b.state = StateHalfOpen
			return nil
		}
		return ErrOpen
	case StateHalfOpen:
		return nil
	}
	return nil
}

// RecordSuccess records a successful operation. If in HalfOpen state the
// circuit closes and the failure counter resets.
func (b *Breaker) RecordSuccess() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.failures = 0
	b.state = StateClosed
}

// RecordFailure records a failed operation. If the failure count reaches the
// threshold the circuit opens.
func (b *Breaker) RecordFailure() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.failures++
	if b.state == StateHalfOpen || b.failures >= b.threshold {
		b.state = StateOpen
		b.openedAt = time.Now()
	}
}

// State returns the current circuit state.
func (b *Breaker) CurrentState() State {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.state
}
