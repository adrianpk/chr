package metadata

import (
	"fmt"
	"strings"

	opsmetadata "forge.adrianpk.com/chr/internal/core/ops/metadata"
)

// SetField replaces a scalar or inserts it before the first selected metadata key.
func SetField(content, key, value string, beforeKeys []string) (string, error) {
	lines, err := FrontmatterLines(content)
	if err != nil {
		return "", err
	}

	for index, line := range lines {
		lineKey, _, ok := strings.Cut(line, ":")
		if ok && lineKey == key {
			lines[index] = key + ": " + value

			return ReplaceFrontmatterLines(content, lines)
		}
	}

	insertAt := FrontmatterInsertIndex(lines, beforeKeys)
	lines = opsmetadata.ReplaceLines(lines, insertAt, insertAt, []string{key + ": " + value})

	return ReplaceFrontmatterLines(content, lines)
}

// FrontmatterLines returns the existing ticket metadata lines without changing their bytes.
func FrontmatterLines(content string) ([]string, error) {
	bounds, err := FrontmatterBounds(content)
	if err != nil {
		return nil, err
	}

	return strings.Split(content[bounds.Start:bounds.End], "\n"), nil
}

// TextRange identifies the original half-open ticket text or metadata-line range.
type TextRange struct {
	Start, End int
}

// FrontmatterBounds requires the original ticket opening and closing delimiters.
func FrontmatterBounds(content string) (TextRange, error) {
	if !strings.HasPrefix(content, "---\n") {
		return TextRange{}, fmt.Errorf("ticket frontmatter opening delimiter is required")
	}

	start := len("---\n")

	end := strings.Index(content[start:], "\n---")
	if end < 0 {
		return TextRange{}, fmt.Errorf("ticket frontmatter closing delimiter is required")
	}

	return TextRange{Start: start, End: start + end}, nil
}

// ReplaceFrontmatterLines preserves the exact original body and delimiter bytes.
func ReplaceFrontmatterLines(content string, lines []string) (string, error) {
	bounds, err := FrontmatterBounds(content)
	if err != nil {
		return "", err
	}

	return content[:bounds.Start] + strings.Join(lines, "\n") + content[bounds.End:], nil
}

// FrontmatterBlockRange locates an indented field block, or returns negative bounds.
func FrontmatterBlockRange(lines []string, key string) TextRange {
	for index, line := range lines {
		lineKey, _, ok := strings.Cut(line, ":")
		if !ok || lineKey != key {
			continue
		}

		end := index + 1
		for end < len(lines) && strings.HasPrefix(lines[end], "  ") {
			end++
		}

		return TextRange{Start: index, End: end}
	}

	return TextRange{Start: -1, End: -1}
}

// FrontmatterInsertIndex selects the first requested key in existing line order.
func FrontmatterInsertIndex(lines, beforeKeys []string) int {
	for index, line := range lines {
		lineKey, _, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}

		for _, beforeKey := range beforeKeys {
			if strings.TrimSpace(lineKey) == beforeKey {
				return index
			}
		}
	}

	return len(lines)
}
