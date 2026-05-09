// Package replay provides dead-letter queue functionality for logslice.
// Failed log entries are stored and can be replayed to sinks at a later time.
package replay

import (
	"context"
	"sync"
	"time"
)

// Entry holds a log payload that failed to be forwarded.
type Entry struct {
	Payload   []byte
	Sink      string
	FailedAt  time.Time
	Attempts  int
}

// Store is a bounded in-memory dead-letter queue.
type Store struct {
	mu      sync.Mutex
	entries []*Entry
	cap     int
}

// New creates a Store with the given maximum capacity.
// If cap is <= 0 it defaults to 100.
func New(cap int) *Store {
	if cap <= 0 {
		cap = 100
	}
	return &Store{cap: cap}
}

// Add enqueues a failed entry. If the store is full the oldest entry is evicted.
func (s *Store) Add(sink string, payload []byte, attempts int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	e := &Entry{
		Payload:  payload,
		Sink:     sink,
		FailedAt: time.Now().UTC(),
		Attempts: attempts,
	}
	if len(s.entries) >= s.cap {
		s.entries = s.entries[1:]
	}
	s.entries = append(s.entries, e)
}

// Drain removes and returns all current entries.
func (s *Store) Drain() []*Entry {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]*Entry, len(s.entries))
	copy(out, s.entries)
	s.entries = s.entries[:0]
	return out
}

// Len returns the current number of queued entries.
func (s *Store) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.entries)
}

// Sender is the interface used to forward a replayed payload.
type Sender interface {
	Send(ctx context.Context, payload []byte) error
}

// Replay drains the store and attempts to re-send each entry via the provided
// senders map (keyed by sink name). Entries that fail again are re-queued.
func (s *Store) Replay(ctx context.Context, senders map[string]Sender) {
	entries := s.Drain()
	for _, e := range entries {
		sender, ok := senders[e.Sink]
		if !ok {
			continue
		}
		if err := sender.Send(ctx, e.Payload); err != nil {
			s.Add(e.Sink, e.Payload, e.Attempts+1)
		}
	}
}
