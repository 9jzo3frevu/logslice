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

// LivenessHandler always returns 200 OK as long as the process is running.
func (c *Checker) LivenessHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(Status{
		OK:        true,
		Timestamp: time.Now().UTC(),
		Version:   c.version,
	})
}

// ReadinessHandler returns 200 when ready, 503 otherwise.
func (c *Checker) ReadinessHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if !c.ready.Load() {
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(Status{
			OK:        false,
			Timestamp: time.Now().UTC(),
			Version:   c.version,
		})
		return
	}
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(Status{
		OK:        true,
		Timestamp: time.Now().UTC(),
		Version:   c.version,
	})
}
