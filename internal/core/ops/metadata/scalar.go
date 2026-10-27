package metadata

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Quote encodes the existing double-quoted metadata scalar.
func Quote(value string) string {
	return strconv.Quote(value)
}

// QuoteList preserves the existing bracketed, comma-separated scalar list.
func QuoteList(values []string) string {
	if len(values) == 0 {
		return "[]"
	}

	quoted := make([]string, 0, len(values))
	for _, value := range values {
		quoted = append(quoted, Quote(value))
	}

	return "[" + strings.Join(quoted, ", ") + "]"
}

// Timestamp emits the existing quoted UTC RFC3339 value or empty zero-time scalar.
func Timestamp(value time.Time) string {
	if value.IsZero() {
		return Quote("")
	}

	return Quote(value.UTC().Format(time.RFC3339))
}

// Unquote preserves trimming and the existing permissive scalar fallback.
func Unquote(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}

	unquoted, err := strconv.Unquote(value)
	if err == nil {
		return unquoted
	}

	return value
}

// ParseList preserves the existing list splitting and empty-item omission.
func ParseList(value string) []string {
	value = strings.TrimSpace(value)
	if value == "" || value == "[]" {
		return nil
	}

	value = strings.TrimPrefix(value, "[")
	value = strings.TrimSuffix(value, "]")

	parts := strings.Split(value, ",")

	values := make([]string, 0, len(parts))
	for _, part := range parts {
		parsed := Unquote(part)
		if parsed != "" {
			values = append(values, parsed)
		}
	}

	return values
}

// OptionalTimestamp parses the existing optional RFC3339 metadata timestamp.
func OptionalTimestamp(value string) (*time.Time, error) {
	value = Unquote(value)
	if value == "" {
		return nil, nil
	}

	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil, fmt.Errorf("parse ticket timestamp: %w", err)
	}

	parsed = parsed.UTC()

	return &parsed, nil
}
