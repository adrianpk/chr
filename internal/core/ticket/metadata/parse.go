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
		key, value, ok := ticketFrontmatterField(lines[index])
		if !ok {
			continue
		}

		if key == "commits" && value == "" {
			values, next := parseTicketScalarList(lines, index+1)
			metadata[key], index = opsmetadata.QuoteList(values), next-1

			continue
		}

		if key == "review_refs" && value != "[]" {
			parsedRefs, nextIndex := parseTicketReviewRefsBlock(lines, index+1)
			reviewRefs = parsedRefs
			index = nextIndex - 1

			continue
		}

		metadata[key] = value
	}

	return Frontmatter{Metadata: metadata, ReviewRefs: reviewRefs}, nil
}

func ticketFrontmatterField(line string) (string, string, bool) {
	if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
		return "", "", false
	}

	key, value, found := strings.Cut(line, ":")

	return strings.TrimSpace(key), strings.TrimSpace(value), found
}

func parseTicketScalarList(lines []string, start int) ([]string, int) {
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

	return values, index
}

func parseTicketReviewRefsBlock(lines []string, start int) ([]ReviewRef, int) {
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

	return refs, index
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
