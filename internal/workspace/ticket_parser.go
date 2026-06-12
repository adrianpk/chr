package workspace

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// TicketParser defines a package boundary interface.
type TicketParser interface {
	ParseTicket(ctx context.Context, opts ParseTicketOptions) (IndexedTicket, error)
}

// ParseTicketOptions describes a package boundary value.
type ParseTicketOptions struct {
	ProjectID    string
	ProjectDir   string
	TicketsDir   string
	Path         string
	IndexedAt    time.Time
	RelativeBase string
}

type ticketParser struct{}

type ticketFrontmatter struct {
	Metadata   map[string]string
	ReviewRefs []TicketReviewRef
}

// NewTicketParser constructs the corresponding package service or value.
func NewTicketParser() TicketParser {
	return &ticketParser{}
}

// ParseTicket executes the corresponding package operation.
func (p *ticketParser) ParseTicket(ctx context.Context, opts ParseTicketOptions) (IndexedTicket, error) {
	err := ctx.Err()
	if err != nil {
		return IndexedTicket{}, err
	}

	content, err := os.ReadFile(opts.Path)
	if err != nil {
		return IndexedTicket{}, fmt.Errorf("read ticket file: %w", err)
	}

	info, err := os.Stat(opts.Path)
	if err != nil {
		return IndexedTicket{}, fmt.Errorf("stat ticket file: %w", err)
	}

	frontmatter, err := parseTicketFrontmatter(string(content))
	if err != nil {
		return IndexedTicket{}, err
	}

	relativePath, err := relativeTicketPath(opts)
	if err != nil {
		return IndexedTicket{}, err
	}

	return indexedTicketFromMetadata(opts, relativePath, info, content, frontmatter)
}

func parseTicketFrontmatter(content string) (ticketFrontmatter, error) {
	if !strings.HasPrefix(content, "---\n") {
		return ticketFrontmatter{}, fmt.Errorf("ticket frontmatter opening delimiter is required")
	}

	remaining := strings.TrimPrefix(content, "---\n")

	end := strings.Index(remaining, "\n---")
	if end < 0 {
		return ticketFrontmatter{}, fmt.Errorf("ticket frontmatter closing delimiter is required")
	}

	frontmatter := remaining[:end]
	metadata := map[string]string{}
	reviewRefs := []TicketReviewRef{}

	lines := strings.Split(frontmatter, "\n")
	for index := 0; index < len(lines); index++ {
		line := lines[index]

		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}

		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)

		if key == "review_refs" && value != "[]" {
			parsedRefs, nextIndex := parseTicketReviewRefsBlock(lines, index+1)
			reviewRefs = parsedRefs
			index = nextIndex - 1

			continue
		}

		metadata[key] = value
	}

	return ticketFrontmatter{Metadata: metadata, ReviewRefs: reviewRefs}, nil
}

func parseTicketReviewRefsBlock(lines []string, start int) ([]TicketReviewRef, int) {
	refs := []TicketReviewRef{}
	current := TicketReviewRef{}
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

			current = TicketReviewRef{}
			hasCurrent = true
			trimmed = strings.TrimSpace(strings.TrimPrefix(trimmed, "- "))
		}

		key, value, ok := strings.Cut(trimmed, ":")
		if !ok || !hasCurrent {
			continue
		}

		assignTicketReviewRefField(&current, strings.TrimSpace(key), unquoteMetadataString(value))
	}

	if hasCurrent {
		refs = append(refs, current)
	}

	return refs, index
}

func assignTicketReviewRefField(ref *TicketReviewRef, key, value string) {
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

func indexedTicketFromMetadata(
	opts ParseTicketOptions,
	relativePath string,
	info os.FileInfo,
	content []byte,
	frontmatter ticketFrontmatter,
) (IndexedTicket, error) {
	metadata := frontmatter.Metadata

	createdAt, updatedAt, mergedAt, reviewRound, err := parseIndexedTicketMetadataTimes(metadata)
	if err != nil {
		return IndexedTicket{}, err
	}

	ticket := IndexedTicket{
		ProjectID:         opts.ProjectID,
		Path:              opts.Path,
		RelativePath:      relativePath,
		Key:               unquoteMetadataString(metadata["key"]),
		Title:             unquoteMetadataString(metadata["title"]),
		Synopsis:          unquoteMetadataString(metadata["synopsis"]),
		Status:            unquoteMetadataString(metadata["status"]),
		Priority:          unquoteMetadataString(metadata["priority"]),
		Severity:          unquoteMetadataString(metadata["severity"]),
		Type:              unquoteMetadataString(metadata["type"]),
		Tags:              parseMetadataStringList(metadata["tags"]),
		UseCase:           unquoteMetadataString(metadata["use_case"]),
		ProjectArea:       unquoteMetadataString(metadata["project_area"]),
		SourceKind:        unquoteMetadataString(metadata["source_kind"]),
		SourceRef:         unquoteMetadataString(metadata["source_ref"]),
		Slice:             unquoteMetadataString(metadata["slice"]),
		Branch:            unquoteMetadataString(metadata["branch"]),
		ReviewRefs:        frontmatter.ReviewRefs,
		ReviewState:       unquoteMetadataString(metadata["review_state"]),
		ReviewRound:       reviewRound,
		ReviewResult:      unquoteMetadataString(metadata["review_result"]),
		ReviewCommentsURL: unquoteMetadataString(metadata["review_comments_url"]),
		BlockingFindings:  parseMetadataStringList(metadata["blocking_findings"]),
		MergeState:        unquoteMetadataString(metadata["merge_state"]),
		MergeResult:       unquoteMetadataString(metadata["merge_result"]),
		MergeActor:        unquoteMetadataString(metadata["merge_actor"]),
		MergeCommit:       unquoteMetadataString(metadata["merge_commit"]),
		MergedAt:          mergedAt,
		MergeBlocker:      unquoteMetadataString(metadata["merge_blocker"]),
		CreatedAt:         createdAt,
		UpdatedAt:         updatedAt,
		FileMTime:         info.ModTime().UTC(),
		FileSize:          info.Size(),
		ContentHash:       contentHash(content),
		IndexedAt:         opts.IndexedAt.UTC(),
	}

	ticket.BoardColumn, _ = BoardTicketStatusFor(ticket.Status)

	err = ValidateIndexedTicket(ticket)
	if err != nil {
		return IndexedTicket{}, err
	}

	return ticket, nil
}

func parseIndexedTicketMetadataTimes(metadata map[string]string) (*time.Time, *time.Time, *time.Time, int, error) {
	createdAt, err := optionalTicketTime(metadata["created_at"])
	if err != nil {
		return nil, nil, nil, 0, err
	}

	updatedAt, err := optionalTicketTime(metadata["updated_at"])
	if err != nil {
		return nil, nil, nil, 0, err
	}

	mergedAt, err := optionalTicketTime(metadata["merged_at"])
	if err != nil {
		return nil, nil, nil, 0, err
	}

	reviewRound, err := optionalTicketInt(metadata["review_round"])
	if err != nil {
		return nil, nil, nil, 0, err
	}

	return createdAt, updatedAt, mergedAt, reviewRound, nil
}

func relativeTicketPath(opts ParseTicketOptions) (string, error) {
	base := opts.RelativeBase
	if base == "" {
		base = opts.TicketsDir
	}

	if base == "" {
		base = UserPath(opts.ProjectDir, "tickets")
	}

	relativePath, err := filepath.Rel(base, opts.Path)
	if err != nil {
		return "", fmt.Errorf("resolve relative ticket path: %w", err)
	}

	return filepath.ToSlash(relativePath), nil
}

func unquoteMetadataString(value string) string {
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

func parseMetadataStringList(value string) []string {
	value = strings.TrimSpace(value)
	if value == "" || value == "[]" {
		return nil
	}

	value = strings.TrimPrefix(value, "[")
	value = strings.TrimSuffix(value, "]")

	parts := strings.Split(value, ",")

	values := make([]string, 0, len(parts))
	for _, part := range parts {
		parsed := unquoteMetadataString(part)
		if parsed != "" {
			values = append(values, parsed)
		}
	}

	return values
}

func optionalTicketTime(value string) (*time.Time, error) {
	value = unquoteMetadataString(value)
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

func optionalTicketInt(value string) (int, error) {
	value = unquoteMetadataString(value)
	if value == "" {
		return 0, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("parse ticket integer: %w", err)
	}

	return parsed, nil
}

func contentHash(content []byte) string {
	sum := sha256.Sum256(content)

	return hex.EncodeToString(sum[:])
}
