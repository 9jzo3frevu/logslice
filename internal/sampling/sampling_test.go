package sampling_test

import (
	"math/rand"
	"testing"

	"github.com/yourorg/logslice/internal/sampling"
)

// deterministicSource always returns the same value so tests are reproducible.
type deterministicSource struct{ val int64 }

func (d deterministicSource) Int63() int64 { return d.val }
func (d deterministicSource) Seed(_ int64)  {}

func newSampler(t *testing.T, cfg sampling.Config, src rand.Source) *sampling.Sampler {
	t.Helper()
	s, err := sampling.New(cfg, src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return s
}

func TestNew_Valid(t *testing.T) {
	_, err := sampling.New(sampling.Config{
		Rates:   map[string]float64{"debug": 0.1, "info": 0.5},
		Default: 1.0,
	}, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestNew_InvalidRate_Zero(t *testing.T) {
	_, err := sampling.New(sampling.Config{
		Rates: map[string]float64{"debug": 0},
	}, nil)
	if err == nil {
		t.Fatal("expected error for rate=0")
	}
}

func TestNew_InvalidRate_AboveOne(t *testing.T) {
	_, err := sampling.New(sampling.Config{
		Rates: map[string]float64{"info": 1.5},
	}, nil)
	if err == nil {
		t.Fatal("expected error for rate>1")
	}
}

func TestNew_InvalidDefault(t *testing.T) {
	_, err := sampling.New(sampling.Config{Default: -0.1}, nil)
	if err == nil {
		t.Fatal("expected error for negative default rate")
	}
}

func TestAllow_RateOne_AlwaysTrue(t *testing.T) {
	s := newSampler(t, sampling.Config{Default: 1.0}, nil)
	for i := 0; i < 100; i++ {
		if !s.Allow("info") {
			t.Fatal("rate=1.0 should always allow")
		}
	}
}

func TestAllow_LevelCaseInsensitive(t *testing.T) {
	s := newSampler(t, sampling.Config{
		Rates:   map[string]float64{"debug": 1.0},
		Default: 1.0,
	}, nil)
	if !s.Allow("DEBUG") {
		t.Fatal("level matching should be case-insensitive")
	}
}

func TestAllow_LowRate_Blocks(t *testing.T) {
	// Int63() returns math.MaxInt64/2, so Float64() ≈ 0.5; rate 0.1 should block.
	src := deterministicSource{val: 4611686018427387903}
	s := newSampler(t, sampling.Config{
		Rates:   map[string]float64{"debug": 0.1},
		Default: 1.0,
	}, src)
	if s.Allow("debug") {
		t.Fatal("expected debug entry to be blocked at rate 0.1 with rng≈0.5")
	}
}

func TestAllow_FallsBackToDefault(t *testing.T) {
	s := newSampler(t, sampling.Config{Default: 1.0}, nil)
	if !s.Allow("trace") {
		t.Fatal("unknown level should use default rate")
	}
}
