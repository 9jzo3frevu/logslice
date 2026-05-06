// Package schema defines the canonical log entry structure used throughout
// logslice. It provides:
//
//   - Entry: the core structured log type with timestamp, level, message,
//     service, and arbitrary string fields.
//   - Validate: ensures required fields are present.
//   - Normalize: fills in defaults (e.g. current UTC timestamp) and
//     initialises optional maps.
//   - Decoder: a streaming JSON decoder that validates and normalises each
//     entry as it is read from an io.Reader.
//
// All other packages (filter, transform, pipeline, …) operate on *Entry
// values produced by this package.
package schema
