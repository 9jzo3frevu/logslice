// Package replay implements a bounded in-memory dead-letter queue for logslice.
//
// When a sink fails to accept a log entry the payload is stored in a [Store].
// Queued entries can be inspected and re-sent via [Store.Replay] or through
// the HTTP [Handler].
//
// Usage:
//
//	store := replay.New(200)           // capacity of 200 entries
//	store.Add("my-sink", payload, 1)   // record a failure
//	store.Replay(ctx, senders)         // attempt re-delivery
//
// The HTTP handler exposes two endpoints:
//
//	GET  /replay  — list queued entries without consuming them
//	POST /replay  — trigger a replay attempt and return the remaining count
package replay
