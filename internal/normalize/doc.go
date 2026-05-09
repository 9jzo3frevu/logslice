// Package normalize provides field-name normalization for structured log entries.
//
// Many log producers use vendor-specific or abbreviated field names (e.g. "msg"
// instead of "message", "lvl" instead of "level"). The Normalizer rewrites
// those aliases to the canonical logslice field names before the entry is
// processed by the filter, transform, or sink pipeline stages.
//
// Usage:
//
//	n, err := normalize.New([]normalize.Mapping{
//		{From: "msg",  To: "message"},
//		{From: "lvl",  To: "level"},
//		{From: "ts",   To: "timestamp"},
//	})
//	if err != nil {
//		log.Fatal(err)
//	}
//	entry = n.Apply(entry)
package normalize
