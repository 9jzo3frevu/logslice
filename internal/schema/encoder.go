package schema

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

// Encoder writes LogEntry values as newline-delimited JSON to an underlying
// writer. Each call to Encode serialises one entry and appends a newline.
type Encoder struct {
	w   io.Writer
	enc *json.Encoder
}

// NewEncoder returns an Encoder that writes to w.
func NewEncoder(w io.Writer) *Encoder {
	e := json.NewEncoder(w)
	e.SetEscapeHTML(false)
	return &Encoder{w: w, enc: e}
}

// Encode validates, normalises, and writes entry as a single JSON line.
// It returns an error if validation fails or the write fails.
func (e *Encoder) Encode(entry *LogEntry) error {
	if err := Validate(entry); err != nil {
		return fmt.Errorf("encoder: %w", err)
	}
	Normalize(entry)
	if err := e.enc.Encode(entry); err != nil {
		return fmt.Errorf("encoder: %w", err)
	}
	return nil
}

// EncodeToBytes validates, normalises, and serialises entry to a byte slice
// without a trailing newline. Useful when the caller needs a []byte payload.
func EncodeToBytes(entry *LogEntry) ([]byte, error) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	if err := enc.Encode(entry); err != nil {
		return nil, err
	}
	// json.Encoder always appends '\n'; trim it for a clean payload.
	return bytes.TrimRight(buf.Bytes(), "\n"), nil
}
