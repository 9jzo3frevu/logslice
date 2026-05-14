// Package masking provides regex-based partial value masking for structured
// log entries flowing through the logslice proxy.
//
// Unlike the redact package which removes fields entirely, masking replaces
// portions of field values using regular expression substitution. This is
// useful for partially obscuring sensitive data such as credit card numbers,
// email addresses, or API tokens while retaining enough context for debugging.
//
// Usage:
//
//	m, err := masking.New([]masking.Config{
//		{Field: "email", Pattern: `[^@]+@`, Replacement: "***@"},
//		{Field: "card",  Pattern: `\d{12}(\d{4})`, Replacement: "************$1"},
//	})
//	// Wrap your HTTP handler:
//	http.Handle("/logs", masking.Middleware(m)(next))
package masking
