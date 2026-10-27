package metadata

import (
	"strings"
)

// YesNo renders the existing yes/no context label.
func YesNo(value bool) string {
	if value {
		return "yes"
	}

	return "no"
}

// ContextTableCell keeps the original context table escaping, distinct from tracker cells.
func ContextTableCell(value string) string {
	value = strings.ReplaceAll(value, "\n", " ")
	value = strings.ReplaceAll(value, "|", "\\|")

	return strings.TrimSpace(value)
}

// WriteMarkdownField writes the existing context field heading and spacing.
func WriteMarkdownField(output *strings.Builder, title, value string) {
	output.WriteString("## ")
	output.WriteString(title)
	output.WriteString("\n\n")
	output.WriteString(value)
	output.WriteString("\n\n")
}

// WriteMarkdownList writes the existing context list heading and spacing.
func WriteMarkdownList(output *strings.Builder, title string, values []string) {
	output.WriteString("## ")
	output.WriteString(title)
	output.WriteString("\n\n")

	for _, value := range values {
		output.WriteString("- ")
		output.WriteString(value)
		output.WriteString("\n")
	}

	output.WriteString("\n")
}
