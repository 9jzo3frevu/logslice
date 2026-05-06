// Package ratelimit provides a simple token-bucket rate limiter
// for controlling inbound log ingestion throughput.
package ratelimit

import (
	"sync"
	"time"
)

// Limiter is a token-bucket rate limiter.
type Limiter struct {
	mu       sync.Mutex
	tokens   float64
	max      float64
	rate     float64 // tokens per second
	lastTick time.Time
}

// New creates a Limiter that allows up to rps requests per second
// with a burst capacity equal to rps.
func New(rps float64) *Limiter {
	if rps <= 0 {
		rps = 1
	}
	return &Limiter{
		tokens:   rps,
		max:      rps,
		rate:     rps,
		lastTick: time.Now(),
	}
}

// Allow returns true if the request is within the rate limit.
func (l *Limiter) Allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(l.lastTick).Seconds()
	l.lastTick = now

	l.tokens += elapsed * l.rate
	if l.tokens > l.max {
		l.tokens = l.max
	}

	if l.tokens < 1 {
		return false
	}
	l.tokens--
	return true
}

// Rate returns the configured rate in requests per second.
func (l *Limiter) Rate() float64 {
	return l.rate
}
