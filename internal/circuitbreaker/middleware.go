package circuitbreaker

import (
	"encoding/json"
	"net/http"
)

// Middleware wraps an http.Handler and rejects requests with 503 when the
// circuit is open. On success it records a success; on a 5xx response it
// records a failure.
func Middleware(b *Breaker, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := b.Allow(); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error": "service unavailable: circuit open",
			})
			return
		}

		rec := &statusRecorder{ResponseWriter: w, code: http.StatusOK}
		next.ServeHTTP(rec, r)

		if rec.code >= 500 {
			b.RecordFailure()
		} else {
			b.RecordSuccess()
		}
	})
}

// statusRecorder captures the HTTP status code written by a handler.
type statusRecorder struct {
	http.ResponseWriter
	code int
}

func (sr *statusRecorder) WriteHeader(code int) {
	sr.code = code
	sr.ResponseWriter.WriteHeader(code)
}
