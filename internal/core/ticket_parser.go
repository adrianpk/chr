package core

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	opsmetadata "forge.adrianpk.com/chr/internal/core/ops/metadata"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
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

	return ticketFrontmatter{Metadata: metadata, ReviewRefs: reviewRefs}, nil
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

		assignTicketReviewRefField(&current, strings.TrimSpace(key), opsmetadata.Unquote(value))
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
		IndexedTicketIdentity: IndexedTicketIdentity{
			ProjectID:    opts.ProjectID,
			Path:         opts.Path,
			RelativePath: relativePath,
			Key:          opsmetadata.Unquote(metadata["key"]),
		}, IndexedTicketSummary: indexedTicketSummaryFromMetadata(metadata), IndexedTicketClassification: indexedTicketClassificationFromMetadata(metadata), IndexedTicketSource: IndexedTicketSource{
			SourceKind: opsmetadata.Unquote(metadata["source_kind"]),
			SourceRef:  opsmetadata.Unquote(metadata["source_ref"]),
			Slice:      opsmetadata.Unquote(metadata["slice"]),
			Branch:     opsmetadata.Unquote(metadata["branch"]),
		}, IndexedTicketReview: IndexedTicketReview{
			ReviewRefs:        frontmatter.ReviewRefs,
			ReviewState:       opsmetadata.Unquote(metadata["review_state"]),
			ReviewRound:       reviewRound,
			ReviewResult:      opsmetadata.Unquote(metadata["review_result"]),
			ReviewCommentsURL: opsmetadata.Unquote(metadata["review_comments_url"]),
			BlockingFindings:  opsmetadata.ParseList(metadata["blocking_findings"]),
		}, IndexedTicketMerge: IndexedTicketMerge{
			MergeState:   opsmetadata.Unquote(metadata["merge_state"]),
			MergeResult:  opsmetadata.Unquote(metadata["merge_result"]),
			MergeActor:   opsmetadata.Unquote(metadata["merge_actor"]),
			MergeCommit:  opsmetadata.Unquote(metadata["merge_commit"]),
			MergedAt:     mergedAt,
			MergeBlocker: opsmetadata.Unquote(metadata["merge_blocker"]),
		}, IndexedTicketTimes: IndexedTicketTimes{
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		}, IndexedTicketProjection: IndexedTicketProjection{
			FileMTime:   info.ModTime().UTC(),
			FileSize:    info.Size(),
			ContentHash: contenthash.Sum(content),
			IndexedAt:   opts.IndexedAt.UTC(),
		},
	}

	ticket.BoardColumn, _ = BoardTicketStatusFor(ticket.Status)

	err = ValidateIndexedTicket(ticket)
	if err != nil {
		return IndexedTicket{}, err
	}

	return ticket, nil
}

func parseIndexedTicketMetadataTimes(metadata map[string]string) (*time.Time, *time.Time, *time.Time, int, error) {
	createdAt, err := opsmetadata.OptionalTimestamp(metadata["created_at"])
	if err != nil {
		return nil, nil, nil, 0, err
	}

	updatedAt, err := opsmetadata.OptionalTimestamp(metadata["updated_at"])
	if err != nil {
		return nil, nil, nil, 0, err
	}

	mergedAt, err := opsmetadata.OptionalTimestamp(metadata["merged_at"])
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
		base = workspacecore.UserPath(opts.ProjectDir, "tickets")
	}

	relativePath, err := filepath.Rel(base, opts.Path)
	if err != nil {
		return "", fmt.Errorf("resolve relative ticket path: %w", err)
	}

	return filepath.ToSlash(relativePath), nil
}

func optionalTicketInt(value string) (int, error) {
	value = opsmetadata.Unquote(value)
	if value == "" {
		return 0, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("parse ticket integer: %w", err)
	}

	return parsed, nil
}

func indexedTicketSummaryFromMetadata(metadata map[string]string) IndexedTicketSummary {
	return IndexedTicketSummary{
		Title:    opsmetadata.Unquote(metadata["title"]),
		Assignee: opsmetadata.Unquote(metadata["assignee"]),
		Synopsis: opsmetadata.Unquote(metadata["synopsis"]),
	}
}

func indexedTicketClassificationFromMetadata(metadata map[string]string) IndexedTicketClassification {
	return IndexedTicketClassification{
		Status:      opsmetadata.Unquote(metadata["status"]),
		Priority:    opsmetadata.Unquote(metadata["priority"]),
		Severity:    opsmetadata.Unquote(metadata["severity"]),
		Type:        opsmetadata.Unquote(metadata["type"]),
		Tags:        opsmetadata.ParseList(metadata["tags"]),
		UseCase:     opsmetadata.Unquote(metadata["use_case"]),
		ProjectArea: opsmetadata.Unquote(metadata["project_area"]),
	}
}
