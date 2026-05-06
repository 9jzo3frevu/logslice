// Package transform provides log entry transformation pipelines.
//
// A Transform applies a sequence of operations to a log entry map,
// such as adding new fields, renaming existing fields, or deleting fields.
//
// Example usage:
//
//	t, err := transform.New([]config.TransformRule{
//		{Op: "add", Field: "env", Value: "production"},
//		{Op: "rename", Field: "msg", NewField: "message"},
//	})
//	if err != nil {
//		log.Fatal(err)
//	}
//	entry = t.Apply(entry)
package transform
