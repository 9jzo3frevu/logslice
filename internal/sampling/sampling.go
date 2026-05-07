// Package sampling provides probabilistic log sampling to reduce volume
// while preserving statistical representation across log levels.
package sampling

import (
	"math/rand"
	"strings"
	"sync"
)

// Sampler decides whether a log entry should be forwarded based on
// per-level sampling rates in the range (0.0, 1.0].
type Sampler struct {
	mu    sync.Mutex
	rates map[string]float64
	rng   *rand.Rand
}

// Config holds per-level sampling rates. A rate of 1.0 means keep all
// entries; 0.1 means keep roughly 10%.
type Config struct {
	// Rates maps lowercase level names ("debug", "info", etc.) to a
	// sampling probability in (0.0, 1.0].
	Rates map[string]float64
	// Default is the fallback rate for levels not listed in Rates.
	// If zero, it defaults to 1.0 (keep everything).
	Default float64
}

// New creates a Sampler from the provided Config.
// Returns an error if any rate is outside (0.0, 1.0].
func New(cfg Config, src rand.Source) (*Sampler, error) {
	rates := make(map[string]float64, len(cfg.Rates))
	for lvl, r := range cfg.Rates {
		if r <= 0 || r > 1 {
			return nil, &InvalidRateError{Level: lvl, Rate: r}
		}
		rates[strings.ToLower(lvl)] = r
	}
	if cfg.Default == 0 {
		cfg.Default = 1.0
	}
	if cfg.Default <= 0 || cfg.Default > 1 {
		return nil, &InvalidRateError{Level: "default", Rate: cfg.Default}
	}
	rates["default"] = cfg.Default
	if src == nil {
		src = rand.NewSource(42)
	}
	return &Sampler{
		rates: rates,
		rng:   rand.New(src),
	}, nil
}

// Allow returns true if the entry with the given level should be forwarded.
func (s *Sampler) Allow(level string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	rate, ok := s.rates[strings.ToLower(level)]
	if !ok {
		rate = s.rates["default"]
	}
	if rate >= 1.0 {
		return true
	}
	return s.rng.Float64() < rate
}

// InvalidRateError is returned when a sampling rate is out of range.
type InvalidRateError struct {
	Level string
	Rate  float64
}

func (e *InvalidRateError) Error() string {
	return "sampling: invalid rate " + e.Level + ": must be in (0.0, 1.0]"
}
