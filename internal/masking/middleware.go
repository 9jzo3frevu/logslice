package masking

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

// Middleware returns an HTTP middleware that applies the Masker to every
// inbound POST request body containing a JSON log entry before passing it
// downstream. Non-POST requests and non-JSON bodies are forwarded unchanged.
func Middleware(m *Masker) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				next.ServeHTTP(w, r)
				return
			}

			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "failed to read body", http.StatusBadRequest)
				return
			}
			_ = r.Body.Close()

			var entry map[string]any
			if err := json.Unmarshal(body, &entry); err != nil {
				// Not valid JSON — pass through unchanged.
				r.Body = io.NopCloser(bytes.NewReader(body))
				next.ServeHTTP(w, r)
				return
			}

			masked := m.Apply(entry)

			encoded, err := json.Marshal(masked)
			if err != nil {
				http.Error(w, "failed to encode masked entry", http.StatusInternalServerError)
				return
			}

			r.Body = io.NopCloser(bytes.NewReader(encoded))
			r.ContentLength = int64(len(encoded))
			next.ServeHTTP(w, r)
		})
	}
}
