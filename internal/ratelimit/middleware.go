package ratelimit

import (
	"net/http"
)

// Middleware returns an http.Handler that enforces the given Limiter.
// Requests exceeding the rate limit receive 429 Too Many Requests.
func Middleware(l *Limiter, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !l.Allow() {
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
