// Package pipeline wires together filtering, transformation, and fanout
// into a single processing step for incoming log entries.
package pipeline

import (
	"fmt"

	"github.com/yourorg/logslice/internal/filter"
	"github.com/yourorg/logslice/internal/metrics"
	"github.com/yourorg/logslice/internal/sink"
	"github.com/yourorg/logslice/internal/transform"
)

// Processor applies filter, transform, and fanout to a log entry.
type Processor struct {
	filter    *filter.Filter
	transform *transform.Transform
	fanout    *sink.Fanout
}

// Config holds the dependencies needed to build a Processor.
type Config struct {
	Filter    *filter.Filter
	Transform *transform.Transform
	Fanout    *sink.Fanout
}

// New creates a Processor from the provided Config.
// All fields in Config are required.
func New(cfg Config) (*Processor, error) {
	if cfg.Filter == nil {
		return nil, fmt.Errorf("pipeline: filter is required")
	}
	if cfg.Transform == nil {
		return nil, fmt.Errorf("pipeline: transform is required")
	}
	if cfg.Fanout == nil {
		return nil, fmt.Errorf("pipeline: fanout is required")
	}
	return &Processor{
		filter:    cfg.Filter,
		transform: cfg.Transform,
		fanout:    cfg.Fanout,
	}, nil
}

// Process filters, transforms, and forwards a single log entry.
// It returns true if the entry was forwarded, false if it was filtered out.
// Any forwarding errors are recorded in metrics but not returned.
func (p *Processor) Process(entry map[string]any) bool {
	if !p.filter.Allow(entry) {
		metrics.Global.Filtered.Add(1)
		return false
	}

	p.filter.Tag(entry)
	entry = p.transform.Apply(entry)

	if err := p.fanout.Send(entry); err != nil {
		metrics.Global.Errors.Add(1)
	}
	metrics.Global.Forwarded.Add(1)
	return true
}
