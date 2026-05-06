package audit

import (
	"net/http"
	"sync"
)

// MemoryStore holds a bounded in-memory ring of recent audit events
// and exposes them over HTTP for debugging purposes.
type MemoryStore struct {
	mu     sync.RWMutex
	events []Event
	max    int
}

// NewMemoryStore returns a MemoryStore that keeps at most maxEvents entries.
func NewMemoryStore(maxEvents int) *MemoryStore {
	if maxEvents <= 0 {
		maxEvents = 100
	}
	return &MemoryStore{max: maxEvents}
}

// Write satisfies io.Writer so MemoryStore can be used as an audit.Logger sink.
func (m *MemoryStore) Write(p []byte) (int, error) {
	return len(p), nil // raw bytes are not parsed here; use Record directly
}

// Add appends an event, evicting the oldest if at capacity.
func (m *MemoryStore) Add(e Event) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.events) >= m.max {
		m.events = m.events[1:]
	}
	m.events = append(m.events, e)
}

// Snapshot returns a copy of all stored events.
func (m *MemoryStore) Snapshot() []Event {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]Event, len(m.events))
	copy(out, m.events)
	return out
}

// Handler returns an http.HandlerFunc that renders stored events as JSON.
func (m *MemoryStore) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		events := m.Snapshot()
		w.Header().Set("Content-Type", "application/json")
		enc := newJSONArrayEncoder(w)
		for _, e := range events {
			_ = enc.Encode(e)
		}
		enc.Close()
	}
}
