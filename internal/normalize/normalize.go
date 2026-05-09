// Package normalize provides field-name normalization for incoming log entries.
// It maps aliased or vendor-specific field names to canonical logslice fields.
package normalize

import "fmt"

// Mapping defines a single field alias → canonical name pair.
type Mapping struct {
	From string `yaml:"from"`
	To   string `yaml:"to"`
}

// Normalizer rewrites aliased keys in a log entry map.
type Normalizer struct {
	mappings []Mapping
}

// New constructs a Normalizer from the provided mappings.
// Returns an error if any mapping has an empty From or To field.
func New(mappings []Mapping) (*Normalizer, error) {
	for i, m := range mappings {
		if m.From == "" {
			return nil, fmt.Errorf("normalize: mapping[%d] has empty 'from' field", i)
		}
		if m.To == "" {
			return nil, fmt.Errorf("normalize: mapping[%d] has empty 'to' field", i)
		}
	}
	return &Normalizer{mappings: mappings}, nil
}

// Apply rewrites aliased keys in entry according to the configured mappings.
// If a source key is present and the destination key is absent, the value is
// moved; if the destination key already exists the source key is left intact.
// The original entry map is mutated in place and returned.
func (n *Normalizer) Apply(entry map[string]any) map[string]any {
	for _, m := range n.mappings {
		val, ok := entry[m.From]
		if !ok {
			continue
		}
		if _, exists := entry[m.To]; exists {
			continue
		}
		entry[m.To] = val
		delete(entry, m.From)
	}
	return entry
}

// Mappings returns a copy of the configured mapping slice.
func (n *Normalizer) Mappings() []Mapping {
	out := make([]Mapping, len(n.mappings))
	copy(out, n.mappings)
	return out
}
