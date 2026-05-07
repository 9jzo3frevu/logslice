package enrichment

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

// Middleware returns an HTTP middleware that enriches every incoming JSON
// log payload with the static fields from e before passing it downstream.
// Non-POST requests and non-JSON bodies are forwarded unchanged.
func Middleware(e *Enricher) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				next.ServeHTTP(w, r)
				return
			}
			ct := r.Header.Get("Content-Type")
			if ct != "application/json" {
				next.ServeHTTP(w, r)
				return
			}
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "failed to read body", http.StatusBadRequest)
				return
			}
			defer r.Body.Close()

			var entry map[string]any
			if err := json.Unmarshal(body, &entry); err != nil {
				// Not valid JSON — pass through as-is
				r.Body = io.NopCloser(bytes.NewReader(body))
				next.ServeHTTP(w, r)
				return
			}

			enriched, err := e.Apply(entry)
			if err != nil {
				http.Error(w, "enrichment failed", http.StatusInternalServerError)
				return
			}

			encoded, err := json.Marshal(enriched)
			if err != nil {
				http.Error(w, "failed to encode enriched entry", http.StatusInternalServerError)
				return
			}

			r.Body = io.NopCloser(bytes.NewReader(encoded))
			r.ContentLength = int64(len(encoded))
			next.ServeHTTP(w, r)
		})
	}
}
