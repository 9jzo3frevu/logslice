package redact

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

// Middleware returns an HTTP middleware that redacts sensitive fields from
// JSON request bodies before passing them to the next handler.
// The body is replaced with the redacted version so downstream handlers
// receive clean, masked content.
func Middleware(r *Redactor) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if req.Body == nil || req.Method != http.MethodPost {
				next.ServeHTTP(w, req)
				return
			}

			body, err := io.ReadAll(req.Body)
			if err != nil {
				http.Error(w, "failed to read body", http.StatusBadRequest)
				return
			}
			_ = req.Body.Close()

			var entry map[string]any
			if err := json.Unmarshal(body, &entry); err != nil {
				// Not JSON — pass through unchanged.
				req.Body = io.NopCloser(strings.NewReader(string(body)))
				next.ServeHTTP(w, req)
				return
			}

			r.Apply(entry)

			redacted, err := json.Marshal(entry)
			if err != nil {
				http.Error(w, "failed to encode redacted body", http.StatusInternalServerError)
				return
			}

			req.Body = io.NopCloser(strings.NewReader(string(redacted)))
			req.ContentLength = int64(len(redacted))
			next.ServeHTTP(w, req)
		})
	}
}
