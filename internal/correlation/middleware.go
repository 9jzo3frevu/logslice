package correlation

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

// Middleware reads or generates a correlation ID for each incoming POST
// request, injects it into the JSON body, and propagates it via response header.
func Middleware(inj *Injector, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			next.ServeHTTP(w, r)
			return
		}

		ctype := r.Header.Get("Content-Type")
		if ctype != "application/json" {
			next.ServeHTTP(w, r)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil || len(body) == 0 {
			next.ServeHTTP(w, r)
			return
		}
		_ = r.Body.Close()

		var entry map[string]interface{}
		if err := json.Unmarshal(body, &entry); err != nil {
			// Not valid JSON — pass through unchanged.
			r.Body = io.NopCloser(bytes.NewReader(body))
			next.ServeHTTP(w, r)
			return
		}

		id := inj.FromRequest(r)
		inj.Inject(entry, id)

		modified, err := json.Marshal(entry)
		if err != nil {
			r.Body = io.NopCloser(bytes.NewReader(body))
			next.ServeHTTP(w, r)
			return
		}

		r.Body = io.NopCloser(bytes.NewReader(modified))
		r.ContentLength = int64(len(modified))
		w.Header().Set(HeaderName, id)

		next.ServeHTTP(w, r)
	})
}
