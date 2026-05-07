// Package ratelimit provides token-bucket rate limiting for logslice.
//
// It exposes three main constructs:
//
//   - Limiter: a single token-bucket rate limiter wrapping golang.org/x/time/rate.
//   - KeyedLimiter: manages independent per-key Limiters with automatic TTL-based
//     eviction, suitable for per-IP or per-tenant rate limiting.
//   - Middleware / KeyedMiddleware: http.Handler wrappers that enforce limits and
//     return HTTP 429 when a request exceeds its allowance.
//
// Example — global limit:
//
//	l := ratelimit.New(100, 10)
//	http.Handle("/ingest", ratelimit.Middleware(l, next))
//
// Example — per-IP limit:
//
//	http.Handle("/ingest", ratelimit.KeyedMiddleware(50, 5, 5*time.Minute, next))
package ratelimit
