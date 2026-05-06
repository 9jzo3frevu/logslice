// Package pipeline wires together filter, transform, and sink components
// into a single http.Handler that processes incoming log entries end-to-end.
package pipeline

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/logslice/logslice/internal/config"
	"github.com/logslice/logslice/internal/filter"
	"github.com/logslice/logslice/internal/schema"
	"github.com/logslice/logslice/internal/sink"
)

// Pipeline holds the assembled components for log processing.
type Pipeline struct {
	filter *filter.Filter
	fanout *sink.Fanout
}

// New constructs a Pipeline from the provided configuration.
// It returns an error if any sink or filter cannot be initialised.
func New(cfg *config.Config) (*Pipeline, error) {
	f, err := filter.New(cfg.Filter.MinLevel, cfg.Filter.Tags)
	if err != nil {
		return nil, fmt.Errorf("pipeline: filter: %w", err)
	}

	sinks := make([]*sink.Sink, 0, len(cfg.Sinks))
	for _, sc := range cfg.Sinks {
		s, err := sink.New(sc.Name, sc.URL)
		if err != nil {
			return nil, fmt.Errorf("pipeline: sink %q: %w", sc.Name, err)
		}
		sinks = append(sinks, s)
	}

	return &Pipeline{
		filter: f,
		fanout: sink.NewFanout(sinks),
	}, nil
}

// ServeHTTP implements http.Handler so the pipeline can be mounted directly
// onto a router. It decodes, validates, filters, and forwards each entry.
func (p *Pipeline) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var entry map[string]any
	if err := json.NewDecoder(r.Body).Decode(&entry); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if err := schema.Validate(entry); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	entry = schema.Normalize(entry)

	if !p.filter.Allow(entry) {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	p.filter.Tag(entry)

	if err := p.fanout.Send(r.Context(), entry); err != nil {
		http.Error(w, "failed to forward log", http.StatusBadGateway)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}
