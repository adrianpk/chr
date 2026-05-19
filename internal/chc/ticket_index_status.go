package chc

import (
	"context"
	"fmt"
	"time"
)

type TicketIndexStatusChecker interface {
	CheckProjectTicketIndex(ctx context.Context, opts CheckProjectTicketIndexOptions) (TicketIndexStatus, error)
}

type CheckProjectTicketIndexOptions struct {
	ProjectID  string
	ProjectDir string
	IndexedAt  time.Time
}

type TicketIndexStatus struct {
	Missing  []TicketIndexStatusFile
	Stale    []TicketIndexStatusTicket
	Orphaned []IndexedTicket
	Invalid  []TicketIndexRebuildParseError
}

type TicketIndexStatusFile struct {
	Path         string
	RelativePath string
}

type TicketIndexStatusTicket struct {
	Indexed IndexedTicket
	Parsed  IndexedTicket
}

type ticketIndexStatusChecker struct {
	discoverer TicketFileDiscoverer
	parser     TicketParser
	index      TicketIndex
	now        func() time.Time
}

type TicketIndexStatusCheckerOption func(*ticketIndexStatusChecker)

func NewTicketIndexStatusChecker(index TicketIndex, opts ...TicketIndexStatusCheckerOption) TicketIndexStatusChecker {
	checker := &ticketIndexStatusChecker{
		discoverer: NewTicketFileDiscoverer(),
		parser:     NewTicketParser(),
		index:      index,
		now:        time.Now,
	}

	for _, opt := range opts {
		opt(checker)
	}

	return checker
}

func WithTicketIndexStatusCheckerClock(now func() time.Time) TicketIndexStatusCheckerOption {
	return func(checker *ticketIndexStatusChecker) {
		if now != nil {
			checker.now = now
		}
	}
}

func (c *ticketIndexStatusChecker) CheckProjectTicketIndex(
	ctx context.Context,
	opts CheckProjectTicketIndexOptions,
) (TicketIndexStatus, error) {
	if err := validateCheckProjectTicketIndexOptions(opts); err != nil {
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
	indexedTickets []IndexedTicket,
) TicketIndexStatus {
	status := TicketIndexStatus{}
	indexedByRelativePath := indexedTicketsByRelativePath(indexedTickets)
	seenRelativePaths := map[string]struct{}{}
	indexedAt := c.indexedAt(opts)

	for _, file := range files {
		seenRelativePaths[file.RelativePath] = struct{}{}

		parsed, err := c.parser.ParseTicket(ctx, ParseTicketOptions{
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

func indexedTicketsByRelativePath(tickets []IndexedTicket) map[string]IndexedTicket {
	indexed := make(map[string]IndexedTicket, len(tickets))

	for _, ticket := range tickets {
		indexed[ticket.RelativePath] = ticket
	}

	return indexed
}

func ticketFileStateChanged(indexed, parsed IndexedTicket) bool {
	return indexed.FileSize != parsed.FileSize ||
		indexed.ContentHash != parsed.ContentHash
}

func orphanedIndexedTickets(tickets []IndexedTicket, seenRelativePaths map[string]struct{}) []IndexedTicket {
	orphaned := []IndexedTicket{}

	for _, ticket := range tickets {
		if _, found := seenRelativePaths[ticket.RelativePath]; !found {
			orphaned = append(orphaned, ticket)
		}
	}

	return orphaned
}
