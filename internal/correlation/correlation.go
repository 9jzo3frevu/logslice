// Package correlation provides request correlation ID injection and propagation
// for structured log entries passing through the proxy pipeline.
package correlation

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

const (
	// HeaderName is the HTTP header used to carry the correlation ID.
	HeaderName = "X-Correlation-ID"

	// FieldName is the JSON field injected into log entries.
	FieldName = "correlation_id"
)

// Generator produces correlation IDs.
type Generator interface {
	Generate() string
}

// DefaultGenerator produces 16-byte random hex correlation IDs.
type DefaultGenerator struct{}

// Generate returns a new random correlation ID.
func (g *DefaultGenerator) Generate() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "unknown"
	}
	return hex.EncodeToString(b)
}

// Injector reads or generates a correlation ID from an HTTP request,
// then injects it into a log entry map.
type Injector struct {
	gen Generator
}

// New creates an Injector with the given Generator. If gen is nil,
// DefaultGenerator is used.
func New(gen Generator) *Injector {
	if gen == nil {
		gen = &DefaultGenerator{}
	}
	return &Injector{gen: gen}
}

// FromRequest returns the correlation ID from the request header,
// generating one if the header is absent.
func (inj *Injector) FromRequest(r *http.Request) string {
	if id := r.Header.Get(HeaderName); id != "" {
		return id
	}
	return inj.gen.Generate()
}

// Inject sets the correlation ID field on the log entry map.
func (inj *Injector) Inject(entry map[string]interface{}, id string) {
	entry[FieldName] = id
}
