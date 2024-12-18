package logger

import (
	"fmt"
)

// extractFields converts variadic arguments to Fields
func extractFields(args ...interface{}) Fields {
	fields := make(Fields)

	// Process arguments in pairs
	for i := 0; i < len(args); i += 2 {
		// Ensure we have a key-value pair
		if i+1 >= len(args) {
			break
		}

		// Convert key to string
		key := fmt.Sprintf("%v", args[i])
		value := args[i+1]

		fields[key] = value
	}

	return fields
}

// mergeFields combines multiple field sources
func mergeFields(base Fields, additional interface{}) Fields {
	// Create a new Fields map to avoid modifying the original
	merged := make(Fields)

	// Add base fields
	for k, v := range base {
		merged[k] = v
	}

	// Handle different types of additional fields
	switch t := additional.(type) {
	case Fields:
		// If additional is Fields, add its key-value pairs
		for k, v := range t {
			merged[k] = v
		}
	case []interface{}:
		// If additional is a slice of arguments, process in pairs
		for i := 0; i < len(t); i += 2 {
			if i+1 < len(t) {
				key := fmt.Sprintf("%v", t[i])
				merged[key] = t[i+1]
			}
		}
	}

	return merged
}

// convertFieldsToArgs converts Fields to a slice of arguments
func convertFieldsToArgs(fields Fields) []interface{} {
	args := make([]interface{}, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}
	return args
}

// slogAttrFromFields converts Fields to slog.Attr slice
func slogAttrFromFields(fields Fields) []any {
	attrs := make([]any, 0, len(fields)*2)
	for k, v := range fields {
		attrs = append(attrs, k, v)
	}
	return attrs
}
