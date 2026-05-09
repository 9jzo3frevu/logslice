package replay

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

type entryView struct {
	Sink     string    `json:"sink"`
	FailedAt time.Time `json:"failed_at"`
	Attempts int       `json:"attempts"`
	Size     int       `json:"payload_bytes"`
}

// Handler exposes the dead-letter queue over HTTP.
// GET  /replay  — list queued entries (non-destructive).
// POST /replay  — trigger a replay attempt using the provided senders.
type Handler struct {
	store   *Store
	senders map[string]Sender
}

// NewHandler creates an HTTP handler backed by the given store and senders.
func NewHandler(store *Store, senders map[string]Sender) *Handler {
	return &Handler{store: store, senders: senders}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.list(w)
	case http.MethodPost:
		h.replay(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) list(w http.ResponseWriter) {
	h.store.mu.Lock()
	views := make([]entryView, len(h.store.entries))
	for i, e := range h.store.entries {
		views[i] = entryView{
			Sink:     e.Sink,
			FailedAt: e.FailedAt,
			Attempts: e.Attempts,
			Size:     len(e.Payload),
		}
	}
	h.store.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(views)
}

func (h *Handler) replay(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	h.store.Replay(ctx, h.senders)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]int{"remaining": h.store.Len()})
}
