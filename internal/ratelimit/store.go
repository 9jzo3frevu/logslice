package ratelimit

import (
	"sync"
	"time"
)

// KeyedLimiter manages per-key rate limiters, allowing independent rate
// limiting for distinct sources such as client IPs or API keys.
type KeyedLimiter struct {
	mu      sync.Mutex
	limiters map[string]*Limiter
	rate    float64
	burst   int
	ttl     time.Duration
	lastSeen map[string]time.Time
}

// NewKeyedLimiter creates a KeyedLimiter where each key gets its own
// Limiter configured with rate and burst. Entries not seen for ttl are
// evicted on the next Allow call.
func NewKeyedLimiter(rate float64, burst int, ttl time.Duration) *KeyedLimiter {
	return &KeyedLimiter{
		limiters:  make(map[string]*Limiter),
		lastSeen:  make(map[string]time.Time),
		rate:      rate,
		burst:     burst,
		ttl:       ttl,
	}
}

// Allow reports whether the given key is within its rate limit.
func (kl *KeyedLimiter) Allow(key string) bool {
	kl.mu.Lock()
	defer kl.mu.Unlock()
	kl.evict()
	l, ok := kl.limiters[key]
	if !ok {
		l = New(kl.rate, kl.burst)
		kl.limiters[key] = l
	}
	kl.lastSeen[key] = time.Now()
	return l.Allow()
}

// evict removes entries that have not been seen within the TTL window.
// Must be called with kl.mu held.
func (kl *KeyedLimiter) evict() {
	now := time.Now()
	for k, t := range kl.lastSeen {
		if now.Sub(t) > kl.ttl {
			delete(kl.limiters, k)
			delete(kl.lastSeen, k)
		}
	}
}

// Len returns the number of active per-key limiters.
func (kl *KeyedLimiter) Len() int {
	kl.mu.Lock()
	defer kl.mu.Unlock()
	return len(kl.limiters)
}
