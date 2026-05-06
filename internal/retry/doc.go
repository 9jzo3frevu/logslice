// Package retry implements a simple exponential-backoff retry helper used
// by the logslice sink layer when forwarding log entries to remote endpoints.
//
// Usage:
//
//	p := retry.DefaultPolicy()
//	err := p.Do(ctx, func() error {
//		return sink.Send(entry)
//	})
package retry
