// Package routing provides label-based log routing rules that direct
// log entries to named sink groups based on field matching.
package routing

import (
	"errors"
	"strings"
)

// Rule describes a single routing rule: if the log entry's field
// matches value, the entry is forwarded to the listed sinks.
type Rule struct {
	Field string   `yaml:"field"`
	Value string   `yaml:"value"`
	Sinks []string `yaml:"sinks"`
}

// Router holds an ordered list of rules and a default sink list used
// when no rule matches.
type Router struct {
	rules        []Rule
	defaultSinks []string
}

// New validates and returns a Router. defaultSinks must not be empty.
func New(rules []Rule, defaultSinks []string) (*Router, error) {
	if len(defaultSinks) == 0 {
		return nil, errors.New("routing: defaultSinks must not be empty")
	}
	for i, r := range rules {
		if strings.TrimSpace(r.Field) == "" {
			return nil, errors.New("routing: rule field must not be empty")
		}
		if len(r.Sinks) == 0 {
			return nil, errors.New("routing: rule sinks must not be empty")
		}
		rules[i].Field = strings.ToLower(strings.TrimSpace(r.Field))
	}
	return &Router{rules: rules, defaultSinks: defaultSinks}, nil
}

// Match evaluates the entry (represented as a flat string→string map)
// against each rule in order and returns the first matching sink list.
// If no rule matches, defaultSinks is returned.
func (r *Router) Match(entry map[string]string) []string {
	for _, rule := range r.rules {
		v, ok := entry[rule.Field]
		if !ok {
			v, ok = entry[strings.ToLower(rule.Field)]
		}
		if ok && strings.EqualFold(v, rule.Value) {
			return rule.Sinks
		}
	}
	return r.defaultSinks
}

// DefaultSinks returns the configured fallback sink list.
func (r *Router) DefaultSinks() []string { return r.defaultSinks }
