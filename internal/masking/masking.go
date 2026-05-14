// Package masking provides pattern-based value masking for structured log entries.
// Unlike redact (which removes fields), masking partially obscures values using
// configurable patterns such as regex substitution.
package masking

import (
	"errors"
	"fmt"
	"regexp"
)

// Rule describes a single masking rule: a compiled regex and its replacement.
type Rule struct {
	Field       string
	Pattern     *regexp.Regexp
	Replacement string
}

// Masker applies a set of rules to log entry maps.
type Masker struct {
	rules []Rule
}

// Config holds the configuration for a single masking rule.
type Config struct {
	Field       string
	Pattern     string
	Replacement string
}

// New constructs a Masker from the provided rule configs.
// Returns an error if any pattern fails to compile or a field name is empty.
func New(configs []Config) (*Masker, error) {
	if len(configs) == 0 {
		return nil, errors.New("masking: at least one rule is required")
	}
	rules := make([]Rule, 0, len(configs))
	for i, c := range configs {
		if c.Field == "" {
			return nil, fmt.Errorf("masking: rule %d: field name must not be empty", i)
		}
		if c.Pattern == "" {
			return nil, fmt.Errorf("masking: rule %d: pattern must not be empty", i)
		}
		re, err := regexp.Compile(c.Pattern)
		if err != nil {
			return nil, fmt.Errorf("masking: rule %d: invalid pattern: %w", i, err)
		}
		rules = append(rules, Rule{
			Field:       c.Field,
			Pattern:     re,
			Replacement: c.Replacement,
		})
	}
	return &Masker{rules: rules}, nil
}

// Apply iterates over all rules and masks matching field values in entry.
// Only string-typed field values are processed; other types are left unchanged.
func (m *Masker) Apply(entry map[string]any) map[string]any {
	for _, rule := range m.rules {
		v, ok := entry[rule.Field]
		if !ok {
			continue
		}
		s, ok := v.(string)
		if !ok {
			continue
		}
		entry[rule.Field] = rule.Pattern.ReplaceAllString(s, rule.Replacement)
	}
	return entry
}
