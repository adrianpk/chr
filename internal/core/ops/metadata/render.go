package metadata

import (
	"strings"
)

// TableCell escapes the existing Markdown tracker cell representation.
func TableCell(value string) string {
	return strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", "|", "&#124;", "\n", " ", "\r", " ").Replace(value)
}

// ReplaceLines replaces the original half-open metadata line range.
func ReplaceLines(lines []string, start, end int, replacement []string) []string {
	updated := append([]string{}, lines[:start]...)
	updated = append(updated, replacement...)

	return append(updated, lines[end:]...)
}
