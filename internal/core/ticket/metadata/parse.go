package metadata

import (
	"fmt"
	"strings"

	opsmetadata "forge.adrianpk.com/chr/internal/core/ops/metadata"
)

// Frontmatter retains scalar fields and the original ordered review references.
type Frontmatter struct {
	Metadata   map[string]string
	ReviewRefs []ReviewRef
}

// Parse reads the original legacy ticket frontmatter representation.
func Parse(content string) (Frontmatter, error) {
	if !strings.HasPrefix(content, "---\n") {
		return Frontmatter{}, fmt.Errorf("ticket frontmatter opening delimiter is required")
	}

	remaining := strings.TrimPrefix(content, "---\n")

	end := strings.Index(remaining, "\n---")
	if end < 0 {
		return Frontmatter{}, fmt.Errorf("ticket frontmatter closing delimiter is required")
	}

	frontmatter := remaining[:end]
	metadata := map[string]string{}
	reviewRefs := []ReviewRef{}

	lines := strings.Split(frontmatter, "\n")
	for index := 0; index < len(lines); index++ {
		field := ticketFrontmatterField(lines[index])
		if !field.found {
			continue
		}

		if field.key == "commits" && field.value == "" {
			block := parseTicketScalarList(lines, index+1)
			metadata[field.key], index = opsmetadata.QuoteList(block.values), block.next-1

			continue
		}

		if field.key == "review_refs" && field.value != "[]" {
			block := parseTicketReviewRefsBlock(lines, index+1)
			reviewRefs = block.refs
			index = block.next - 1

			continue
		}

		metadata[field.key] = field.value
	}

	return Frontmatter{Metadata: metadata, ReviewRefs: reviewRefs}, nil
}

type frontmatterField struct {
	key, value string
	found      bool
}

func ticketFrontmatterField(line string) frontmatterField {
	if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
		return frontmatterField{}
	}

	key, value, found := strings.Cut(line, ":")

	return frontmatterField{key: strings.TrimSpace(key), value: strings.TrimSpace(value), found: found}
}

type scalarListBlock struct {
	values []string
	next   int
}

func parseTicketScalarList(lines []string, start int) scalarListBlock {
	values := []string{}
	index := start

	for ; index < len(lines); index++ {
		line := lines[index]
		if !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
			break
		}

		if value, found := strings.CutPrefix(strings.TrimSpace(line), "- "); found {
			values = append(values, opsmetadata.Unquote(value))
		}
	}

	return scalarListBlock{values: values, next: index}
}

type reviewRefsBlock struct {
	refs []ReviewRef
	next int
}

func parseTicketReviewRefsBlock(lines []string, start int) reviewRefsBlock {
	refs := []ReviewRef{}
	current := ReviewRef{}
	hasCurrent := false
	index := start

	for ; index < len(lines); index++ {
		line := lines[index]
		if !strings.HasPrefix(line, "  ") {
			break
		}

		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "- ") {
			if hasCurrent {
				refs = append(refs, current)
			}

			current = ReviewRef{}
			hasCurrent = true
			trimmed = strings.TrimSpace(strings.TrimPrefix(trimmed, "- "))
		}

		key, value, ok := strings.Cut(trimmed, ":")
		if !ok || !hasCurrent {
			continue
		}

		assignTicketReviewRefField(&current, strings.TrimSpace(key), opsmetadata.Unquote(value))
	}

	if hasCurrent {
		refs = append(refs, current)
	}

	return reviewRefsBlock{refs: refs, next: index}
}

func assignTicketReviewRefField(ref *ReviewRef, key, value string) {
	switch key {
	case "type":
		ref.Type = value
	case "provider":
		ref.Provider = value
	case "url":
		ref.URL = value
	case "id":
		ref.ID = value
	case "status":
		ref.Status = value
	}
}

// ReviewRef describes a package boundary value.
type ReviewRef struct {
	Type     string
	Provider string
	URL      string
	ID       string
	Status   string
}
