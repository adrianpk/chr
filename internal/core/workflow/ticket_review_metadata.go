package workflow

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	opsmetadata "forge.adrianpk.com/chr/internal/core/ops/metadata"
	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
	ticketmetadata "forge.adrianpk.com/chr/internal/core/ticket/metadata"
)

// TicketReviewMetadataUpdater defines a package boundary interface.
type TicketReviewMetadataUpdater interface {
	UpdateReviewMetadata(ctx context.Context, opts UpdateTicketReviewMetadataOptions) error
}

// UpdateTicketReviewMetadataOptions describes a package boundary value.
type UpdateTicketReviewMetadataOptions struct {
	Path      string
	Branch    string
	ReviewRef *ticketmetadata.ReviewRef
	UpdatedAt time.Time
}

type ticketReviewMetadataUpdater struct{}

// NewTicketReviewMetadataUpdater constructs the corresponding package service or value.
func NewTicketReviewMetadataUpdater() TicketReviewMetadataUpdater {
	return &ticketReviewMetadataUpdater{}
}

// UpdateReviewMetadata executes the corresponding package operation.
func (u *ticketReviewMetadataUpdater) UpdateReviewMetadata(ctx context.Context, opts UpdateTicketReviewMetadataOptions) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if strings.TrimSpace(opts.Branch) == "" && opts.ReviewRef == nil {
		return fmt.Errorf("ticket review metadata update requires branch or review ref")
	}

	content, err := ReadLegacyTicketFile(opts.Path)
	if err != nil {
		return fmt.Errorf("read ticket file: %w", err)
	}

	updated, err := updateTicketReviewMetadataContent(string(content), opts)
	if err != nil {
		return err
	}

	err = ReplaceLegacyTicketFile(ctx, opts.Path, content, []byte(updated))
	if err != nil {
		return fmt.Errorf("write ticket file: %w", err)
	}

	return nil
}

func updateTicketReviewMetadataContent(content string, opts UpdateTicketReviewMetadataOptions) (string, error) {
	updated := content

	if strings.TrimSpace(opts.Branch) != "" {
		var err error

		updated, err = ticketmetadata.SetField(
			updated,
			"branch",
			opsmetadata.Quote(strings.TrimSpace(opts.Branch)),
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

	return ticketmetadata.SetField(updated, "updated_at", opsmetadata.Timestamp(updatedAt), nil)
}

func upsertTicketReviewRef(content string, ref ticketmetadata.ReviewRef) (string, error) {
	if err := ticketcore.ValidateReviewRef(ref); err != nil {
		return "", err
	}

	frontmatter, err := ticketmetadata.Parse(content)
	if err != nil {
		return "", err
	}

	refs := upsertReviewRef(frontmatter.ReviewRefs, ref)

	return setTicketReviewRefsField(content, refs)
}

func upsertReviewRef(refs []ticketmetadata.ReviewRef, ref ticketmetadata.ReviewRef) []ticketmetadata.ReviewRef {
	refs = append([]ticketmetadata.ReviewRef(nil), refs...)

	for index, existing := range refs {
		if reviewRefsMatch(existing, ref) {
			refs[index] = ref

			return refs
		}
	}

	return append(refs, ref)
}

func reviewRefsMatch(existing, next ticketmetadata.ReviewRef) bool {
	if existing.URL != "" && existing.URL == next.URL {
		return true
	}

	return existing.Type == next.Type &&
		existing.Provider == next.Provider &&
		existing.ID != "" &&
		existing.ID == next.ID
}

func setTicketReviewRefsField(content string, refs []ticketmetadata.ReviewRef) (string, error) {
	var rendered bytes.Buffer

	err := ticketcore.WriteReviewRefsField(&rendered, refs)
	if err != nil {
		return "", err
	}

	lines, err := ticketmetadata.FrontmatterLines(content)
	if err != nil {
		return "", err
	}

	replacement := strings.Split(strings.TrimRight(rendered.String(), "\n"), "\n")

	block := ticketmetadata.FrontmatterBlockRange(lines, "review_refs")
	if block.Start < 0 {
		insertAt := ticketmetadata.FrontmatterInsertIndex(lines, []string{
			"review_state",
			"review_result",
			"review_comments_url",
			"review_round",
			"blocking_findings",
			"created_at",
		})

		lines = opsmetadata.ReplaceLines(lines, insertAt, insertAt, replacement)

		return ticketmetadata.ReplaceFrontmatterLines(content, lines)
	}

	lines = opsmetadata.ReplaceLines(lines, block.Start, block.End, replacement)

	return ticketmetadata.ReplaceFrontmatterLines(content, lines)
}
