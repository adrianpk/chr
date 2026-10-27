// Package configvalue owns shared configuration literal rules.
package configvalue

import "strings"

// SetString retains the current value when the new value is empty.
func SetString(target *string, value string) {
	if value != "" {
		*target = value
	}
}

// String trims whitespace and supported quote delimiters.
func String(value string) string {
	return strings.Trim(strings.TrimSpace(value), `"'`)
}

// Bool recognizes the existing true spellings; other values are false.
func Bool(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "true", "yes", "1", "on":
		return true
	default:
		return false
	}
}
