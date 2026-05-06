// Package retry provides configurable retry logic with exponential backoff
// for use when forwarding logs to downstream sinks.
package retry

import (
	"context"
	"errors"
	"math"
	"time"
)

// Policy defines the retry behaviour.
type Policy struct {
	// MaxAttempts is the total number of attempts (including the first).
	MaxAttempts int
	// BaseDelay is the initial backoff duration.
	BaseDelay time.Duration
	// MaxDelay caps the exponential growth.
	MaxDelay time.Duration
}

// DefaultPolicy returns a sensible out-of-the-box retry policy.
func DefaultPolicy() Policy {
	return Policy{
		MaxAttempts: 3,
		BaseDelay:   100 * time.Millisecond,
		MaxDelay:    2 * time.Second,
	}
}

// ErrExhausted is returned when all attempts have been consumed.
var ErrExhausted = errors.New("retry: all attempts exhausted")

// Do calls fn up to p.MaxAttempts times. It returns nil on the first success.
// Between attempts it sleeps an exponentially growing delay, bounded by
// p.MaxDelay. If ctx is cancelled the function returns ctx.Err() immediately.
func (p Policy) Do(ctx context.Context, fn func() error) error {
	if p.MaxAttempts <= 0 {
		p.MaxAttempts = 1
	}

	var lastErr error
	for attempt := 0; attempt < p.MaxAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}

		if lastErr = fn(); lastErr == nil {
			return nil
		}

		if attempt == p.MaxAttempts-1 {
			break
		}

		delay := p.backoff(attempt)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
	}

	return errors.Join(ErrExhausted, lastErr)
}

// backoff computes the delay for a given attempt index.
func (p Policy) backoff(attempt int) time.Duration {
	exp := math.Pow(2, float64(attempt))
	d := time.Duration(exp * float64(p.BaseDelay))
	if d > p.MaxDelay {
		d = p.MaxDelay
	}
	return d
}
