package ticket

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	opsmetadata "forge.adrianpk.com/chr/internal/core/ops/metadata"
	ticketmetadata "forge.adrianpk.com/chr/internal/core/ticket/metadata"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
)

// Parser defines a package boundary interface.
type Parser interface {
	Parse(ctx context.Context, opts ParseOptions) (Indexed, error)
}

// ParseOptions describes a package boundary value.
type ParseOptions struct {
	ProjectID    string
	ProjectDir   string
	TicketsDir   string
	Path         string
	IndexedAt    time.Time
	RelativeBase string
}

type parser struct{}

// NewParser constructs the corresponding package service or value.
func NewParser() Parser {
	return &parser{}
}

// Parse executes the corresponding package operation.
func (p *parser) Parse(ctx context.Context, opts ParseOptions) (Indexed, error) {
	err := ctx.Err()
	if err != nil {
		return Indexed{}, err
	}

	content, err := os.ReadFile(opts.Path)
	if err != nil {
		return Indexed{}, fmt.Errorf("read ticket file: %w", err)
	}

	info, err := os.Stat(opts.Path)
	if err != nil {
		return Indexed{}, fmt.Errorf("stat ticket file: %w", err)
	}

	frontmatter, err := ticketmetadata.Parse(string(content))
	if err != nil {
		return Indexed{}, err
	}

	relativePath, err := relativeTicketPath(opts)
	if err != nil {
		return Indexed{}, err
	}

	return indexedTicketFromMetadata(opts, relativePath, info, content, frontmatter)
}

func indexedTicketFromMetadata(
	opts ParseOptions,
	relativePath string,
	info os.FileInfo,
	content []byte,
	frontmatter ticketmetadata.Frontmatter,
) (Indexed, error) {
	metadata := frontmatter.Metadata

	createdAt, updatedAt, mergedAt, reviewRound, err := parseIndexedTicketMetadataTimes(metadata)
	if err != nil {
		return Indexed{}, err
	}

	ticket := Indexed{
		IndexedIdentity: IndexedIdentity{
			ProjectID:    opts.ProjectID,
			Path:         opts.Path,
			RelativePath: relativePath,
			Key:          opsmetadata.Unquote(metadata["key"]),
		}, IndexedSummary: indexedTicketSummaryFromMetadata(metadata), IndexedClassification: indexedTicketClassificationFromMetadata(metadata), IndexedSource: IndexedSource{
			SourceKind: opsmetadata.Unquote(metadata["source_kind"]),
			SourceRef:  opsmetadata.Unquote(metadata["source_ref"]),
			Slice:      opsmetadata.Unquote(metadata["slice"]),
			Branch:     opsmetadata.Unquote(metadata["branch"]),
		}, IndexedReview: IndexedReview{
			ReviewRefs:        frontmatter.ReviewRefs,
			ReviewState:       opsmetadata.Unquote(metadata["review_state"]),
			ReviewRound:       reviewRound,
			ReviewResult:      opsmetadata.Unquote(metadata["review_result"]),
			ReviewCommentsURL: opsmetadata.Unquote(metadata["review_comments_url"]),
			BlockingFindings:  opsmetadata.ParseList(metadata["blocking_findings"]),
		}, IndexedMerge: IndexedMerge{
			MergeState:   opsmetadata.Unquote(metadata["merge_state"]),
			MergeResult:  opsmetadata.Unquote(metadata["merge_result"]),
			MergeActor:   opsmetadata.Unquote(metadata["merge_actor"]),
			MergeCommit:  opsmetadata.Unquote(metadata["merge_commit"]),
			MergedAt:     mergedAt,
			MergeBlocker: opsmetadata.Unquote(metadata["merge_blocker"]),
		}, IndexedTimes: IndexedTimes{
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		}, IndexedProjection: IndexedProjection{
			FileMTime:   info.ModTime().UTC(),
			FileSize:    info.Size(),
			ContentHash: contenthash.Sum(content),
			IndexedAt:   opts.IndexedAt.UTC(),
		},
	}

	ticket.BoardColumn, _ = BoardStatus(ticket.Status)

	err = ValidateIndexed(ticket)
	if err != nil {
		return Indexed{}, err
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

	reviewRound, err := ticketmetadata.OptionalInteger(metadata["review_round"])
	if err != nil {
		return nil, nil, nil, 0, err
	}

	return createdAt, updatedAt, mergedAt, reviewRound, nil
}

func relativeTicketPath(opts ParseOptions) (string, error) {
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

func indexedTicketSummaryFromMetadata(metadata map[string]string) IndexedSummary {
	return IndexedSummary{
		Title:    opsmetadata.Unquote(metadata["title"]),
		Assignee: opsmetadata.Unquote(metadata["assignee"]),
		Synopsis: opsmetadata.Unquote(metadata["synopsis"]),
	}
}

func indexedTicketClassificationFromMetadata(metadata map[string]string) IndexedClassification {
	return IndexedClassification{
		Status:      opsmetadata.Unquote(metadata["status"]),
		Priority:    opsmetadata.Unquote(metadata["priority"]),
		Severity:    opsmetadata.Unquote(metadata["severity"]),
		Type:        opsmetadata.Unquote(metadata["type"]),
		Tags:        opsmetadata.ParseList(metadata["tags"]),
		UseCase:     opsmetadata.Unquote(metadata["use_case"]),
		ProjectArea: opsmetadata.Unquote(metadata["project_area"]),
	}
}
