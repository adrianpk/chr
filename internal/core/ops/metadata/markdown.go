package metadata

import (
	"fmt"
	"strings"
)

// Markdown is the original frontmatter and body image, with explicit presence.
// Body retains its original slice and omission semantics.
type Markdown struct {
	Front   string
	Body    []byte
	Present bool
}

// SplitMarkdown preserves optional frontmatter for legacy authoring documents.
// Native callers require presence separately before decoding their schema.
func SplitMarkdown(data []byte) (Markdown, error) {
	text := string(data)
	if !strings.HasPrefix(text, "---\n") {
		return Markdown{Body: data}, nil
	}

	end := strings.Index(text[4:], "\n---\n")
	if end < 0 {
		return Markdown{}, fmt.Errorf("incomplete workflow frontmatter")
	}

	return Markdown{Front: text[4 : end+4], Body: data[end+9:], Present: true}, nil
}
