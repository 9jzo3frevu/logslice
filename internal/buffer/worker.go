package buffer

import (
	"context"
	"log"
)

// Sender is anything that can send a raw payload, e.g. a sink fanout.
type Sender interface {
	Send(payload []byte) error
}

// Worker drains a Buffer and forwards each entry to a Sender.
type Worker struct {
	buf    *Buffer
	sender Sender
}

// NewWorker creates a Worker that drains buf and forwards to sender.
func NewWorker(buf *Buffer, sender Sender) *Worker {
	return &Worker{buf: buf, sender: sender}
}

// Run starts the drain loop. It blocks until ctx is cancelled.
func (w *Worker) Run(ctx context.Context) {
	w.buf.Drain(ctx, func(e Entry) {
		if err := w.sender.Send(e.Payload); err != nil {
			log.Printf("buffer.Worker: send error: %v", err)
		}
	})
}
