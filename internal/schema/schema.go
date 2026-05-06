// Package schema provides JSON log entry validation and normalization.
package schema

import (
	"errors"
	"fmt"
	"time"
)

// Entry represents a structured log entry passed through the proxy.
type Entry struct {
	Timestamp time.Time         `json:"timestamp"`
	Level     string            `json:"level"`
	Message   string            `json:"message"`
	Service   string            `json:"service,omitempty"`
	Fields    map[string]string `json:"fields,omitempty"`
}

// Validate checks that the entry contains required fields.
func Validate(e *Entry) error {
	if e == nil {
		return errors.New("entry must not be nil")
	}
	if e.Message == "" {
		return errors.New("entry.message is required")
	}
	if e.Level == "" {
		return errors.New("entry.level is required")
	}
	return nil
}

// Normalize fills in default values for optional fields.
func Normalize(e *Entry) error {
	if err := Validate(e); err != nil {
		return fmt.Errorf("normalize: %w", err)
	}
	if e.Timestamp.IsZero() {
		e.Timestamp = time.Now().UTC()
	}
	if e.Fields == nil {
		e.Fields = make(map[string]string)
	}
	return nil
}
