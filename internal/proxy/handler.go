package proxy

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/yourorg/logslice/internal/filter"
	"github.com/yourorg/logslice/internal/sink"
)

// Handler receives incoming log payloads, filters them, tags them,
// and fans out to all configured sinks.
type Handler struct {
	filter  *filter.Filter
	fanout  *sink.Fanout
}

// NewHandler constructs a Handler with the given filter and fanout.
func NewHandler(f *filter.Filter, fo *sink.Fanout) *Handler {
	return &Handler{filter: f, fanout: fo}
}

// ServeHTTP implements http.Handler.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var entry map[string]interface{}
	if err := json.Unmarshal(body, &entry); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if !h.filter.Allow(entry) {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	tagged := h.filter.Tag(entry)

	payload, err := json.Marshal(tagged)
	if err != nil {
		http.Error(w, "failed to encode entry", http.StatusInternalServerError)
		return
	}

	if err := h.fanout.Send(r.Context(), payload); err != nil {
		http.Error(w, "failed to forward log", http.StatusBadGateway)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}
