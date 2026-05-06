// Package healthcheck provides a simple HTTP health check handler
// that reports the liveness and readiness of the logslice proxy.
package healthcheck

import (
	"encoding/json"
	"net/http"
	"sync/atomic"
	"time"
)

// Status holds the health check response payload.
type Status struct {
	OK        bool      `json:"ok"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version,omitempty"`
}

// Checker tracks readiness state and exposes an HTTP handler.
type Checker struct {
	ready   atomic.Bool
	version string
}

// New creates a new Checker. The service starts in a not-ready state.
func New(version string) *Checker {
	return &Checker{version: version}
}

// SetReady marks the service as ready to receive traffic.
func (c *Checker) SetReady(ready bool) {
	c.ready.Store(ready)
}

// writeJSON encodes v as JSON into w with the given HTTP status code.
func (c *Checker) writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

// newStatus constructs a Status with the current UTC timestamp and checker version.
func (c *Checker) newStatus(ok bool) Status {
	return Status{
		OK:        ok,
		Timestamp: time.Now().UTC(),
		Version:   c.version,
	}
}

// LivenessHandler always returns 200 OK as long as the process is running.
func (c *Checker) LivenessHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	c.writeJSON(w, http.StatusOK, c.newStatus(true))
}

// ReadinessHandler returns 200 when ready, 503 otherwise.
func (c *Checker) ReadinessHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !c.ready.Load() {
		c.writeJSON(w, http.StatusServiceUnavailable, c.newStatus(false))
		return
	}
	c.writeJSON(w, http.StatusOK, c.newStatus(true))
}
