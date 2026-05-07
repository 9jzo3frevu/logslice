// Package dedup provides log entry deduplication using a sliding time window.
// Duplicate entries with identical message and level within the window are suppressed.
package dedup

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// Entry represents the minimal fields used for deduplication keying.
type Entry interface {
	GetMessage() string
	GetLevel() string
}

// Deduplicator tracks recently seen log entries and suppresses duplicates.
type Deduplicator struct {
	mu      sync.Mutex
	seen    map[string]time.Time
	window  time.Duration
	now     func() time.Time
}

// New creates a Deduplicator with the given deduplication window.
// A zero or negative window defaults to 5 seconds.
func New(window time.Duration) *Deduplicator {
	if window <= 0 {
		window = 5 * time.Second
	}
	return &Deduplicator{
		seen:   make(map[string]time.Time),
		window: window,
		now:    time.Now,
	}
}

// IsDuplicate returns true if an equivalent entry was seen within the window.
// It also records the entry if it is not a duplicate.
func (d *Deduplicator) IsDuplicate(msg, level string) bool {
	key := fingerprint(msg, level)

	d.mu.Lock()
	defer d.mu.Unlock()

	now := d.now()
	d.evict(now)

	if _, exists := d.seen[key]; exists {
		return true
	}

	d.seen[key] = now
	return false
}

// WindowSize returns the configured deduplication window.
func (d *Deduplicator) WindowSize() time.Duration {
	return d.window
}

// evict removes entries older than the window. Must be called with mu held.
func (d *Deduplicator) evict(now time.Time) {
	for key, ts := range d.seen {
		if now.Sub(ts) > d.window {
			delete(d.seen, key)
		}
	}
}

func fingerprint(msg, level string) string {
	h := sha256.Sum256([]byte(fmt.Sprintf("%s|%s", level, msg)))
	return hex.EncodeToString(h[:])
}
