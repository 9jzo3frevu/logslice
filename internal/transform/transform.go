// Package transform provides log entry field transformation utilities.
// It supports adding, removing, and renaming fields in structured log entries.
package transform

import "fmt"

// Rule defines a single transformation to apply to a log entry.
type Rule struct {
	// Op is the operation: "add", "remove", or "rename".
	Op string `yaml:"op"`
	// Field is the target field name.
	Field string `yaml:"field"`
	// Value is used for "add" operations.
	Value string `yaml:"value,omitempty"`
	// NewField is used for "rename" operations.
	NewField string `yaml:"new_field,omitempty"`
}

// Transformer applies a set of Rules to log entry maps.
type Transformer struct {
	rules []Rule
}

// New validates and returns a Transformer for the given rules.
// Returns an error if any rule is invalid.
func New(rules []Rule) (*Transformer, error) {
	for i, r := range rules {
		switch r.Op {
		case "add":
			if r.Field == "" {
				return nil, fmt.Errorf("rule[%d]: add requires a field name", i)
			}
		case "remove":
			if r.Field == "" {
				return nil, fmt.Errorf("rule[%d]: remove requires a field name", i)
			}
		case "rename":
			if r.Field == "" || r.NewField == "" {
				return nil, fmt.Errorf("rule[%d]: rename requires field and new_field", i)
			}
		default:
			return nil, fmt.Errorf("rule[%d]: unknown op %q", i, r.Op)
		}
	}
	return &Transformer{rules: rules}, nil
}

// Apply executes all rules against entry, returning the modified copy.
// The original map is not mutated.
func (t *Transformer) Apply(entry map[string]any) map[string]any {
	out := make(map[string]any, len(entry))
	for k, v := range entry {
		out[k] = v
	}
	for _, r := range t.rules {
		switch r.Op {
		case "add":
			out[r.Field] = r.Value
		case "remove":
			delete(out, r.Field)
		case "rename":
			if val, ok := out[r.Field]; ok {
				out[r.NewField] = val
				delete(out, r.Field)
			}
		}
	}
	return out
}
