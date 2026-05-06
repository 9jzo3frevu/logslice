// Package buffer provides an in-memory ring buffer for log entries
// that can absorb bursts and drain to a sink asynchronously.
package buffer

import (
	"context"
	"errors"
	"sync"
)

// ErrFull is returned when the buffer has reached its capacity.
var ErrFull = errors.New("buffer: full")

// Entry represents a buffered log payload.
type Entry struct {
	Payload []byte
}

// Buffer is a bounded, thread-safe channel-backed ring buffer.
type Buffer struct {
	mu       sync.Mutex
	ch       chan Entry
	capacity int
}

// New creates a Buffer with the given capacity.
// capacity must be >= 1; if not, it defaults to 1.
func New(capacity int) *Buffer {
	if capacity < 1 {
		capacity = 1
	}
	return &Buffer{
		ch:       make(chan Entry, capacity),
		capacity: capacity,
	}
}

// Write enqueues an entry. Returns ErrFull if the buffer is at capacity.
func (b *Buffer) Write(e Entry) error {
	select {
	case b.ch <- e:
		return nil
	default:
		return ErrFull
	}
}

// Drain reads entries from the buffer and calls fn for each one until
// ctx is cancelled or the buffer is empty and closed.
func (b *Buffer) Drain(ctx context.Context, fn func(Entry)) {
	for {
		select {
		case <-ctx.Done():
			return
		case e, ok := <-b.ch:
			if !ok {
				return
			}
			fn(e)
		}
	}
}

// Len returns the current number of entries waiting in the buffer.
func (b *Buffer) Len() int {
	return len(b.ch)
}

// Cap returns the maximum capacity of the buffer.
func (b *Buffer) Cap() int {
	return b.capacity
}

// Close closes the underlying channel. After Close, Write will panic.
func (b *Buffer) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()
	close(b.ch)
}
