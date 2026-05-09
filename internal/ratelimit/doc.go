// Package ratelimit provides token-bucket rate limiting for incoming log
// entries, both globally and per-client IP. It exposes standard
// http.Handler middleware as well as a keyed limiter backed by an LRU-style
// eviction store.
package ratelimit
