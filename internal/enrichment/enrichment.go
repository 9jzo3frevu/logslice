// Package enrichment provides automatic field injection into log entries
// based on static key-value pairs defined at construction time.
package enrichment

import (
	"errors"
	"fmt"
)

// Enricher appends static fields to every log entry it processes.
type Enricher struct {
	fields map[string]string
}

// New constructs an Enricher from the provided fields map.
// Returns an error if any key is empty or if fields is nil.
func New(fields map[string]string) (*Enricher, error) {
	if fields == nil {
		return nil, errors.New("enrichment: fields map must not be nil")
	}
	for k := range fields {
		if k == "" {
			return nil, errors.New("enrichment: field key must not be empty")
		}
	}
	// defensive copy
	cp := make(map[string]string, len(fields))
	for k, v := range fields {
		cp[k] = v
	}
	return &Enricher{fields: cp}, nil
}

// Apply merges the static fields into entry, overwriting any existing keys
// that conflict. It returns a new map so the original is not mutated.
func (e *Enricher) Apply(entry map[string]any) (map[string]any, error) {
	if entry == nil {
		return nil, errors.New("enrichment: entry must not be nil")
	}
	out := make(map[string]any, len(entry)+len(e.fields))
	for k, v := range entry {
		out[k] = v
	}
	for k, v := range e.fields {
		out[k] = v
	}
	return out, nil
}

// Fields returns a copy of the static fields configured on this Enricher.
func (e *Enricher) Fields() map[string]string {
	cp := make(map[string]string, len(e.fields))
	for k, v := range e.fields {
		cp[k] = v
	}
	return cp
}

// String returns a human-readable summary of the enricher.
func (e *Enricher) String() string {
	return fmt.Sprintf("Enricher{fields: %v}", e.fields)
}
