package core

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"time"
)

// TicketReviewMetadataUpdater defines a package boundary interface.
type TicketReviewMetadataUpdater interface {
	UpdateReviewMetadata(ctx context.Context, opts UpdateTicketReviewMetadataOptions) error
}

// UpdateTicketReviewMetadataOptions describes a package boundary value.
type UpdateTicketReviewMetadataOptions struct {
	Path      string
	Branch    string
	ReviewRef *TicketReviewRef
	UpdatedAt time.Time
}

type ticketReviewMetadataUpdater struct{}

// NewTicketReviewMetadataUpdater constructs the corresponding package service or value.
func NewTicketReviewMetadataUpdater() TicketReviewMetadataUpdater {
	return &ticketReviewMetadataUpdater{}
}

// UpdateReviewMetadata executes the corresponding package operation.
func (u *ticketReviewMetadataUpdater) UpdateReviewMetadata(
	ctx context.Context,
	opts UpdateTicketReviewMetadataOptions,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if strings.TrimSpace(opts.Branch) == "" && opts.ReviewRef == nil {
		return fmt.Errorf("ticket review metadata update requires branch or review ref")
	}

	content, err := readLegacyTicketFile(opts.Path)
	if err != nil {
		return fmt.Errorf("read ticket file: %w", err)
	}

	updated, err := updateTicketReviewMetadataContent(string(content), opts)
	if err != nil {
		return err
	}

	err = os.WriteFile(opts.Path, []byte(updated), 0o644)
	if err != nil {
		return fmt.Errorf("write ticket file: %w", err)
	}

	return nil
}

func updateTicketReviewMetadataContent(
	content string,
	opts UpdateTicketReviewMetadataOptions,
) (string, error) {
	updated := content

	if strings.TrimSpace(opts.Branch) != "" {
		var err error

		updated, err = setTicketFrontmatterScalarField(
			updated,
			"branch",
			yamlString(strings.TrimSpace(opts.Branch)),
			[]string{"review_refs", "review_state", "created_at"},
		)
		if err != nil {
			return "", err
		}
	}

	if opts.ReviewRef != nil {
		var err error

		updated, err = upsertTicketReviewRef(updated, *opts.ReviewRef)
		if err != nil {
			return "", err
		}
	}

	updatedAt := opts.UpdatedAt.UTC()
	if updatedAt.IsZero() {
		updatedAt = time.Now().UTC()
	}

	return setTicketFrontmatterScalarField(updated, "updated_at", yamlTimestamp(updatedAt), nil)
}

func upsertTicketReviewRef(content string, ref TicketReviewRef) (string, error) {
	if err := validateTicketReviewRef(ref); err != nil {
		return "", err
	}

	frontmatter, err := parseTicketFrontmatter(content)
	if err != nil {
		return "", err
	}

	refs := upsertReviewRef(frontmatter.ReviewRefs, ref)

	return setTicketReviewRefsField(content, refs)
}

func upsertReviewRef(refs []TicketReviewRef, ref TicketReviewRef) []TicketReviewRef {
	refs = append([]TicketReviewRef(nil), refs...)

	for index, existing := range refs {
		if reviewRefsMatch(existing, ref) {
			refs[index] = ref

			return refs
		}
	}

	return append(refs, ref)
}

func reviewRefsMatch(existing, next TicketReviewRef) bool {
	if existing.URL != "" && existing.URL == next.URL {
		return true
	}

	return existing.Type == next.Type &&
		existing.Provider == next.Provider &&
		existing.ID != "" &&
		existing.ID == next.ID
}

func setTicketReviewRefsField(content string, refs []TicketReviewRef) (string, error) {
	var rendered bytes.Buffer

	err := writeTicketReviewRefsField(&rendered, refs)
	if err != nil {
		return "", err
	}

	lines, err := ticketFrontmatterLines(content)
	if err != nil {
		return "", err
	}

	replacement := strings.Split(strings.TrimRight(rendered.String(), "\n"), "\n")

	start, end := ticketFrontmatterBlockRange(lines, "review_refs")
	if start < 0 {
		insertAt := ticketFrontmatterInsertIndex(lines, []string{
			"review_state",
			"review_result",
			"review_comments_url",
			"review_round",
			"blocking_findings",
			"created_at",
		})

		lines = insertStringLines(lines, insertAt, replacement)

		return replaceTicketFrontmatterLines(content, lines)
	}

	lines = replaceStringLines(lines, start, end, replacement)

	return replaceTicketFrontmatterLines(content, lines)
}

func setTicketFrontmatterScalarField(
	content string,
	key string,
	value string,
	beforeKeys []string,
) (string, error) {
	lines, err := ticketFrontmatterLines(content)
	if err != nil {
		return "", err
	}

	for index, line := range lines {
		lineKey, _, ok := strings.Cut(line, ":")
		if ok && strings.TrimSpace(lineKey) == key {
			lines[index] = key + ": " + value

			return replaceTicketFrontmatterLines(content, lines)
		}
	}

	insertAt := ticketFrontmatterInsertIndex(lines, beforeKeys)
	lines = insertStringLines(lines, insertAt, []string{key + ": " + value})

	return replaceTicketFrontmatterLines(content, lines)
}

func ticketFrontmatterLines(content string) ([]string, error) {
	start, end, err := ticketFrontmatterBounds(content)
	if err != nil {
		return nil, err
	}

	return strings.Split(content[start:end], "\n"), nil
}

func ticketFrontmatterBounds(content string) (int, int, error) {
	if !strings.HasPrefix(content, "---\n") {
		return 0, 0, fmt.Errorf("ticket frontmatter opening delimiter is required")
	}

	start := len("---\n")

	end := strings.Index(content[start:], "\n---")
	if end < 0 {
		return 0, 0, fmt.Errorf("ticket frontmatter closing delimiter is required")
	}

	return start, start + end, nil
}

func replaceTicketFrontmatterLines(content string, lines []string) (string, error) {
	start, end, err := ticketFrontmatterBounds(content)
	if err != nil {
		return "", err
	}

	return content[:start] + strings.Join(lines, "\n") + content[end:], nil
}

func ticketFrontmatterBlockRange(lines []string, key string) (int, int) {
	for index, line := range lines {
		lineKey, _, ok := strings.Cut(line, ":")
		if !ok || strings.TrimSpace(lineKey) != key {
			continue
		}

		end := index + 1
		for end < len(lines) && strings.HasPrefix(lines[end], "  ") {
			end++
		}

		return index, end
	}

	return -1, -1
}

func ticketFrontmatterInsertIndex(lines, beforeKeys []string) int {
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

func insertStringLines(lines []string, index int, inserted []string) []string {
	updated := append([]string{}, lines[:index]...)
	updated = append(updated, inserted...)

	return append(updated, lines[index:]...)
}

func replaceStringLines(lines []string, start, end int, replacement []string) []string {
	updated := append([]string{}, lines[:start]...)
	updated = append(updated, replacement...)

	return append(updated, lines[end:]...)
}
