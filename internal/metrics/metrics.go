package metrics

import "sync/atomic"

// Counters holds atomic counters for proxy operation metrics.
type Counters struct {
	Received  atomic.Int64
	Forwarded atomic.Int64
	Filtered  atomic.Int64
	Errors    atomic.Int64
}

// Global is the package-level metrics instance used across the proxy.
var Global = &Counters{}

// Snapshot returns a point-in-time copy of the counters as plain int64 values.
type Snapshot struct {
	Received  int64 `json:"received"`
	Forwarded int64 `json:"forwarded"`
	Filtered  int64 `json:"filtered"`
	Errors    int64 `json:"errors"`
}

// Snap returns a Snapshot of the current counter values.
func (c *Counters) Snap() Snapshot {
	return Snapshot{
		Received:  c.Received.Load(),
		Forwarded: c.Forwarded.Load(),
		Filtered:  c.Filtered.Load(),
		Errors:    c.Errors.Load(),
	}
}

// Reset zeroes all counters.
func (c *Counters) Reset() {
	c.Received.Store(0)
	c.Forwarded.Store(0)
	c.Filtered.Store(0)
	c.Errors.Store(0)
}
