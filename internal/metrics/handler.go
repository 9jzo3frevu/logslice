package metrics

import (
	"encoding/json"
	"net/http"
)

// Handler returns an http.HandlerFunc that serves a JSON snapshot of c.
func Handler(c *Counters) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		snap := c.Snap()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(snap); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
		}
	}
}
