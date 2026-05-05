package config_test

import (
	"os"
	"testing"

	"github.com/yourorg/logslice/internal/config"
)

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "logslice-*.yaml")
	if err != nil {
		t.Fatalf("creating temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("writing temp file: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestLoad_Valid(t *testing.T) {
	cfgYAML := `
server:
  addr: ":9090"
filters:
  - field: level
    match: error
    tag: critical
sinks:
  - name: stdout-sink
    type: stdout
`
	path := writeTempConfig(t, cfgYAML)
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Server.Addr != ":9090" {
		t.Errorf("expected addr :9090, got %s", cfg.Server.Addr)
	}
	if len(cfg.Filters) != 1 || cfg.Filters[0].Tag != "critical" {
		t.Errorf("unexpected filters: %+v", cfg.Filters)
	}
	if len(cfg.Sinks) != 1 || cfg.Sinks[0].Name != "stdout-sink" {
		t.Errorf("unexpected sinks: %+v", cfg.Sinks)
	}
}

func TestLoad_DefaultAddr(t *testing.T) {
	path := writeTempConfig(t, "sinks:\n  - name: s1\n    type: stdout\n")
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Server.Addr != ":8080" {
		t.Errorf("expected default addr :8080, got %s", cfg.Server.Addr)
	}
}

func TestLoad_MissingSinkName(t *testing.T) {
	path := writeTempConfig(t, "sinks:\n  - type: stdout\n")
	_, err := config.Load(path)
	if err == nil {
		t.Fatal("expected validation error for missing sink name")
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := config.Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}
