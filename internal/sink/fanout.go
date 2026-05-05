package sink

import (
	"fmt"
	"strings"
	"sync"
)

// Fanout sends a log payload to multiple sinks concurrently.
type Fanout struct {
	sinks []*Sink
}

// NewFanout creates a Fanout from the provided sinks.
func NewFanout(sinks ...*Sink) *Fanout {
	return &Fanout{sinks: sinks}
}

// Send dispatches payload to all sinks in parallel and collects errors.
func (f *Fanout) Send(payload []byte) error {
	var (
		mu   sync.Mutex
		errs []string
		wg   sync.WaitGroup
	)

	for _, s := range f.sinks {
		wg.Add(1)
		go func(s *Sink) {
			defer wg.Done()
			if err := s.Send(payload); err != nil {
				mu.Lock()
				errs = append(errs, err.Error())
				mu.Unlock()
			}
		}(s)
	}

	wg.Wait()

	if len(errs) > 0 {
		return fmt.Errorf("fanout errors: %s", strings.Join(errs, "; "))
	}
	return nil
}

// Len returns the number of sinks in the fanout.
func (f *Fanout) Len() int {
	return len(f.sinks)
}
