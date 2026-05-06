package audit

import (
	"encoding/json"
	"io"
)

// jsonArrayEncoder writes a JSON array to w incrementally.
type jsonArrayEncoder struct {
	w     io.Writer
	first bool
	enc   *json.Encoder
}

func newJSONArrayEncoder(w io.Writer) *jsonArrayEncoder {
	_, _ = w.Write([]byte("["))
	e := json.NewEncoder(w)
	e.SetEscapeHTML(false)
	return &jsonArrayEncoder{w: w, first: true, enc: e}
}

// Encode appends a JSON-encoded value to the array.
func (a *jsonArrayEncoder) Encode(v any) error {
	if !a.first {
		_, _ = a.w.Write([]byte(","))
	}
	a.first = false
	return a.enc.Encode(v)
}

// Close writes the closing bracket.
func (a *jsonArrayEncoder) Close() {
	_, _ = a.w.Write([]byte("]"))
}
