// Package correlation provides request correlation ID injection and propagation.
package correlation

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// contextKey is an unexported type for context keys in this package.
type contextKey struct{}

// DefaultHeader is the HTTP header used to carry the correlation ID.
const DefaultHeader = "X-Correlation-ID"

// Correlator generates and attaches correlation IDs to log entries.
type Correlator struct {
	header string
}

// New returns a Correlator that reads/writes the given header name.
// If header is empty, DefaultHeader is used.
func New(header string) (*Correlator, error) {
	if header == "" {
		header = DefaultHeader
	}
	return &Correlator{header: header}, nil
}

// Header returns the configured header name.
func (c *Correlator) Header() string {
	return c.header
}

// Generate creates a new random correlation ID.
func Generate() (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("correlation: generate id: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// FromContext retrieves the correlation ID stored in ctx, or empty string.
func FromContext(ctx context.Context) string {
	v, _ := ctx.Value(contextKey{}).(string)
	return v
}

// WithContext returns a new context carrying the given correlation ID.
func WithContext(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, contextKey{}, id)
}

// Inject adds the correlation ID to the provided log entry map.
// If the entry already contains the field, it is not overwritten.
func (c *Correlator) Inject(entry map[string]interface{}, id string) {
	if _, exists := entry["correlation_id"]; !exists {
		entry["correlation_id"] = id
	}
}
