// Package ratelimit implements a token-bucket rate limiter and an
// HTTP middleware wrapper for logslice's ingestion endpoint.
//
// Usage:
//
//	limiter := ratelimit.New(100) // 100 req/s with burst of 100
//	http.Handle("/logs", ratelimit.Middleware(limiter, myHandler))
package ratelimit
