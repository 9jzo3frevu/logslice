// Package retry implements a simple exponential-backoff retry helper used
// by the logslice sink layer when forwarding log entries to remote endpoints.
//
// The package exposes a [Policy] type that controls the maximum number of
// attempts, the initial backoff duration, and the multiplier applied on each
// successive failure. A zero-value Policy is not valid; use [DefaultPolicy] to
// obtain a ready-to-use configuration.
//
// Retries are halted early when the supplied [context.Context] is cancelled or
// its deadline is exceeded, in which case the context error is returned
// directly without wrapping.
//
// Usage:
//
//	p := retry.DefaultPolicy()
//	err := p.Do(ctx, func() error {
//		return sink.Send(entry)
//	})
package retry
