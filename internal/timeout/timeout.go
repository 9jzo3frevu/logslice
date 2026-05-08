// Package timeout provides HTTP middleware that enforces a maximum duration
// for request processing. Requests that exceed the deadline receive a 503
// response and the downstream handler's context is cancelled.
package timeout

import (
	"context"
	"net/http"
	"time"
)

// defaultTimeout is used when New is called with a zero or negative duration.
const defaultTimeout = 5 * time.Second

// Limiter holds the configured request timeout.
type Limiter struct {
	duration time.Duration
}

// New returns a Limiter with the given timeout duration.
// If d is zero or negative, defaultTimeout is used.
func New(d time.Duration) *Limiter {
	if d <= 0 {
		d = defaultTimeout
	}
	return &Limiter{duration: d}
}

// Duration returns the configured timeout.
func (l *Limiter) Duration() time.Duration {
	return l.duration
}

// Middleware returns an http.Handler that wraps next with a per-request
// deadline equal to l.duration. If the deadline is exceeded before next
// returns, the client receives HTTP 503 and a plain-text error body.
func (l *Limiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), l.duration)
		defer cancel()

		done := make(chan struct{})
		pw := &panicWriter{ResponseWriter: w}

		go func() {
			defer close(done)
			next.ServeHTTP(pw, r.WithContext(ctx))
		}()

		select {
		case <-done:
			// handler completed in time
		case <-ctx.Done():
			if !pw.written {
				w.Header().Set("Content-Type", "text/plain; charset=utf-8")
				w.WriteHeader(http.StatusServiceUnavailable)
				_, _ = w.Write([]byte("request timeout\n"))
			}
		}
	})
}

// panicWriter guards against writing headers after the timeout path has
// already responded.
type panicWriter struct {
	http.ResponseWriter
	written bool
}

func (pw *panicWriter) WriteHeader(code int) {
	pw.written = true
	pw.ResponseWriter.WriteHeader(code)
}

func (pw *panicWriter) Write(b []byte) (int, error) {
	pw.written = true
	return pw.ResponseWriter.Write(b)
}
