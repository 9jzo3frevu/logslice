package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Sink defines a downstream log destination.
type Sink struct {
	Name    string            `yaml:"name"`
	Type    string            `yaml:"type"` // e.g. "http", "stdout", "file"
	URL     string            `yaml:"url,omitempty"`
	Path    string            `yaml:"path,omitempty"`
	Headers map[string]string `yaml:"headers,omitempty"`
}

// Filter defines a tag-based log filter rule.
type Filter struct {
	Field string `yaml:"field"`
	Match string `yaml:"match"`
	Tag   string `yaml:"tag"`
}

// Server holds HTTP listener configuration.
type Server struct {
	Addr string `yaml:"addr"`
}

// Config is the top-level logslice configuration.
type Config struct {
	Server  Server   `yaml:"server"`
	Filters []Filter `yaml:"filters"`
	Sinks   []Sink   `yaml:"sinks"`
}

// Load reads and parses a YAML config file from the given path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config: reading file %q: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("config: parsing yaml: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("config: validation failed: %w", err)
	}

	return &cfg, nil
}

func (c *Config) validate() error {
	if c.Server.Addr == "" {
		c.Server.Addr = ":8080"
	}
	for i, sink := range c.Sinks {
		if sink.Name == "" {
			return fmt.Errorf("sink[%d] missing name", i)
		}
		if sink.Type == "" {
			return fmt.Errorf("sink %q missing type", sink.Name)
		}
	}
	return nil
}
