// Package slug owns the shared Unicode artifact and tag naming rule.
package slug

import (
	"strings"
	"unicode"
)

// Normalize retains letters and digits with single separators.
func Normalize(value string) string {
	value = strings.TrimSpace(value)

	var output strings.Builder

	lastDash := false

	for _, char := range value {
		if unicode.IsLetter(char) || unicode.IsDigit(char) {
			output.WriteRune(char)

			lastDash = false

			continue
		}

		if !lastDash && output.Len() > 0 {
			output.WriteByte('-')

			lastDash = true
		}
	}

	return strings.Trim(output.String(), "-")
}
