package ratelimit

import (
	"net/http"
	"strings"
	"time"
)

// KeyedMiddleware wraps an http.Handler and applies per-IP rate limiting
// using a KeyedLimiter. Requests that exceed the limit receive 429.
func KeyedMiddleware(rate float64, burst int, ttl time.Duration, next http.Handler) http.Handler {
	kl := NewKeyedLimiter(rate, burst, ttl)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := remoteIP(r)
		if !kl.Allow(key) {
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// remoteIP extracts the client IP from X-Forwarded-For or RemoteAddr.
func remoteIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.SplitN(xff, ",", 2)
		return strings.TrimSpace(parts[0])
	}
	addr := r.RemoteAddr
	if i := strings.LastIndex(addr, ":"); i != -1 {
		return addr[:i]
	}
	return addr
}
