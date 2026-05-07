// Package enrichment implements static field injection for log entries.
//
// An Enricher holds a fixed set of key-value string pairs (e.g. environment,
// region, application name) and merges them into every log entry it processes.
// Enricher fields always win over conflicting keys already present in the entry,
// ensuring that infrastructure-level metadata cannot be spoofed by the sender.
//
// # Usage
//
//	e, err := enrichment.New(map[string]string{
//		"env":    "production",
//		"region": "eu-west-1",
//	})
//
// The package also exposes Middleware, an HTTP middleware that transparently
// enriches JSON POST bodies before they reach the downstream handler.
package enrichment
