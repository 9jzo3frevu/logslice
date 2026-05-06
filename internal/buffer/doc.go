// Package buffer provides a bounded, asynchronous in-memory buffer for
// log entries in the logslice pipeline.
//
// It is designed to absorb short write bursts without blocking the HTTP
// handler path. A background Worker drains the buffer and forwards
// entries to any Sender implementation (typically a sink.Fanout).
//
// Usage:
//
//	buf := buffer.New(512)
//	worker := buffer.NewWorker(buf, fanout)
//	go worker.Run(ctx)
//
//	// In your handler:
//	if err := buf.Write(buffer.Entry{Payload: raw}); err != nil {
//	    // handle ErrFull — drop or return 503
//	}
package buffer
