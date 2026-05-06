package schema

import (
	"encoding/json"
	"fmt"
	"io"
)

// Decoder reads and validates log entries from a JSON stream.
type Decoder struct {
	dec *json.Decoder
}

// NewDecoder creates a Decoder that reads from r.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{dec: json.NewDecoder(r)}
}

// Decode reads the next JSON value from the stream, validates and normalizes it.
func (d *Decoder) Decode() (*Entry, error) {
	var e Entry
	if err := d.dec.Decode(&e); err != nil {
		return nil, fmt.Errorf("decode json: %w", err)
	}
	if err := Normalize(&e); err != nil {
		return nil, fmt.Errorf("invalid entry: %w", err)
	}
	return &e, nil
}
