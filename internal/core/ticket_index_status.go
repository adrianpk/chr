package core

import (
	"context"
	"fmt"
	"time"

	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
)

// TicketIndexStatusChecker defines a package boundary interface.
type TicketIndexStatusChecker interface {
	CheckProjectTicketIndex(ctx context.Context, opts CheckProjectTicketIndexOptions) (TicketIndexStatus, error)
}

// CheckProjectTicketIndexOptions describes a package boundary value.
type CheckProjectTicketIndexOptions struct {
	ProjectID  string
	ProjectDir string
	IndexedAt  time.Time
}

// TicketIndexStatus describes a package boundary value.
type TicketIndexStatus struct {
	Missing  []TicketIndexStatusFile
	Stale    []TicketIndexStatusTicket
	Orphaned []ticketcore.IndexedTicket
	Invalid  []TicketIndexRebuildParseError
}

// TicketIndexStatusFile describes a package boundary value.
type TicketIndexStatusFile struct {
	Path         string
	RelativePath string
}

// TicketIndexStatusTicket describes a package boundary value.
type TicketIndexStatusTicket struct {
	Indexed ticketcore.IndexedTicket
	Parsed  ticketcore.IndexedTicket
}

type ticketIndexStatusChecker struct {
	discoverer TicketFileDiscoverer
	parser     ticketcore.TicketParser
	index      TicketIndex
	now        func() time.Time
}

// TicketIndexStatusCheckerOption defines a package boundary type.
type TicketIndexStatusCheckerOption func(*ticketIndexStatusChecker)

// NewTicketIndexStatusChecker constructs the corresponding package service or value.
func NewTicketIndexStatusChecker(index TicketIndex, opts ...TicketIndexStatusCheckerOption) TicketIndexStatusChecker {
	checker := &ticketIndexStatusChecker{
		discoverer: NewTicketFileDiscoverer(),
		parser:     ticketcore.NewTicketParser(),
		index:      index,
		now:        time.Now,
	}

	for _, opt := range opts {
		opt(checker)
	}

	return checker
}

// WithTicketIndexStatusCheckerClock executes the corresponding package operation.
func WithTicketIndexStatusCheckerClock(now func() time.Time) TicketIndexStatusCheckerOption {
	return func(checker *ticketIndexStatusChecker) {
		if now != nil {
			checker.now = now
		}
	}
}

// CheckProjectTicketIndex executes the corresponding package operation.
func (c *ticketIndexStatusChecker) CheckProjectTicketIndex(
	ctx context.Context,
	opts CheckProjectTicketIndexOptions,
) (TicketIndexStatus, error) {
	err := validateCheckProjectTicketIndexOptions(opts)
	if err != nil {
		return TicketIndexStatus{}, err
	}

	files, err := c.discoverer.DiscoverTicketFiles(ctx, DiscoverTicketFilesOptions{
		ProjectDir: opts.ProjectDir,
	})
	if err != nil {
		return TicketIndexStatus{}, err
	}

	indexedTickets, err := c.index.ListProjectTickets(ctx, opts.ProjectID)
	if err != nil {
		return TicketIndexStatus{}, err
	}

	return c.compareProjectTicketIndex(ctx, opts, files, indexedTickets), nil
}

func (c *ticketIndexStatusChecker) compareProjectTicketIndex(
	ctx context.Context,
	opts CheckProjectTicketIndexOptions,
	files []DiscoveredTicketFile,
	indexedTickets []ticketcore.IndexedTicket,
) TicketIndexStatus {
	status := TicketIndexStatus{}
	indexedByRelativePath := indexedTicketsByRelativePath(indexedTickets)
	seenRelativePaths := map[string]struct{}{}
	indexedAt := c.indexedAt(opts)

	for _, file := range files {
		seenRelativePaths[file.RelativePath] = struct{}{}

		parsed, err := c.parser.ParseTicket(ctx, ticketcore.ParseTicketOptions{
			ProjectID:    opts.ProjectID,
			ProjectDir:   opts.ProjectDir,
			Path:         file.Path,
			IndexedAt:    indexedAt,
			RelativeBase: ticketsDirectory(opts.ProjectDir),
		})
		if err != nil {
			status.Invalid = append(status.Invalid, TicketIndexRebuildParseError{
				Path:         file.Path,
				RelativePath: file.RelativePath,
				Err:          err,
			})

			continue
		}

		indexed, found := indexedByRelativePath[file.RelativePath]
		if !found {
			status.Missing = append(status.Missing, TicketIndexStatusFile(file))

			continue
		}

		if ticketFileStateChanged(indexed, parsed) {
			status.Stale = append(status.Stale, TicketIndexStatusTicket{
				Indexed: indexed,
				Parsed:  parsed,
			})
		}
	}

	status.Orphaned = orphanedIndexedTickets(indexedTickets, seenRelativePaths)

	return status
}

func (c *ticketIndexStatusChecker) indexedAt(opts CheckProjectTicketIndexOptions) time.Time {
	indexedAt := opts.IndexedAt.UTC()
	if indexedAt.IsZero() {
		indexedAt = c.now().UTC()
	}

	return indexedAt
}

func validateCheckProjectTicketIndexOptions(opts CheckProjectTicketIndexOptions) error {
	if opts.ProjectID == "" {
		return fmt.Errorf("project id is required")
	}

	if opts.ProjectDir == "" {
		return fmt.Errorf("project directory is required")
	}

	return nil
}

func indexedTicketsByRelativePath(tickets []ticketcore.IndexedTicket) map[string]ticketcore.IndexedTicket {
	indexed := make(map[string]ticketcore.IndexedTicket, len(tickets))

	for _, ticket := range tickets {
		indexed[ticket.RelativePath] = ticket
	}

	return indexed
}

func ticketFileStateChanged(indexed, parsed ticketcore.IndexedTicket) bool {
	return indexed.FileSize != parsed.FileSize ||
		indexed.ContentHash != parsed.ContentHash
}

func orphanedIndexedTickets(tickets []ticketcore.IndexedTicket, seenRelativePaths map[string]struct{}) []ticketcore.IndexedTicket {
	orphaned := []ticketcore.IndexedTicket{}

	for _, ticket := range tickets {
		_, found := seenRelativePaths[ticket.RelativePath]
		if !found {
			orphaned = append(orphaned, ticket)
		}
	}

	return orphaned
}
