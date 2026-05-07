// Package redact provides field-level redaction for structured log entries.
// It replaces sensitive field values with a configurable mask string before
// logs are forwarded to downstream sinks.
package redact

import (
	"errors"
	"strings"
)

const defaultMask = "[REDACTED]"

// Redactor replaces the values of named fields in a log entry map.
type Redactor struct {
	fields map[string]struct{}
	mask   string
}

// Option configures a Redactor.
type Option func(*Redactor)

// WithMask overrides the default redaction mask string.
func WithMask(mask string) Option {
	return func(r *Redactor) {
		if mask != "" {
			r.mask = mask
		}
	}
}

// New creates a Redactor that will redact the given field names.
// Field matching is case-insensitive.
func New(fields []string, opts ...Option) (*Redactor, error) {
	if len(fields) == 0 {
		return nil, errors.New("redact: at least one field name is required")
	}
	r := &Redactor{
		fields: make(map[string]struct{}, len(fields)),
		mask:   defaultMask,
	}
	for _, f := range fields {
		if f == "" {
			return nil, errors.New("redact: field name must not be empty")
		}
		r.fields[strings.ToLower(f)] = struct{}{}
	}
	for _, o := range opts {
		o(r)
	}
	return r, nil
}

// Apply redacts sensitive fields in-place on the provided entry map.
// Non-string values in targeted fields are replaced with the mask string.
func (r *Redactor) Apply(entry map[string]any) {
	for k := range entry {
		if _, ok := r.fields[strings.ToLower(k)]; ok {
			entry[k] = r.mask
		}
	}
}

// Fields returns the set of field names being redacted.
func (r *Redactor) Fields() []string {
	out := make([]string, 0, len(r.fields))
	for f := range r.fields {
		out = append(out, f)
	}
	return out
}
