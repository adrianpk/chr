package chc

import (
	"context"
	"fmt"
	"time"
)

type TicketIndexRebuilder interface {
	RebuildProjectTickets(ctx context.Context, opts RebuildProjectTicketsOptions) (RebuildProjectTicketsResult, error)
}

type RebuildProjectTicketsOptions struct {
	ProjectID  string
	ProjectDir string
	IndexedAt  time.Time
}

type RebuildProjectTicketsResult struct {
	IndexedTickets int
	ParseErrors    []TicketIndexRebuildParseError
}

type TicketIndexRebuildParseError struct {
	Path         string
	RelativePath string
	Err          error
}

type ticketIndexRebuilder struct {
	discoverer TicketFileDiscoverer
	parser     TicketParser
	index      TicketIndex
	now        func() time.Time
}

type TicketIndexRebuilderOption func(*ticketIndexRebuilder)

func NewTicketIndexRebuilder(index TicketIndex, opts ...TicketIndexRebuilderOption) TicketIndexRebuilder {
	rebuilder := &ticketIndexRebuilder{
		discoverer: NewTicketFileDiscoverer(),
		parser:     NewTicketParser(),
		index:      index,
		now:        time.Now,
	}

	for _, opt := range opts {
		opt(rebuilder)
	}

	return rebuilder
}

func WithTicketIndexRebuilderDiscoverer(discoverer TicketFileDiscoverer) TicketIndexRebuilderOption {
	return func(rebuilder *ticketIndexRebuilder) {
		if discoverer != nil {
			rebuilder.discoverer = discoverer
		}
	}
}

func WithTicketIndexRebuilderParser(parser TicketParser) TicketIndexRebuilderOption {
	return func(rebuilder *ticketIndexRebuilder) {
		if parser != nil {
			rebuilder.parser = parser
		}
	}
}

func WithTicketIndexRebuilderClock(now func() time.Time) TicketIndexRebuilderOption {
	return func(rebuilder *ticketIndexRebuilder) {
		if now != nil {
			rebuilder.now = now
		}
	}
}

func (r *ticketIndexRebuilder) RebuildProjectTickets(
	ctx context.Context,
	opts RebuildProjectTicketsOptions,
) (RebuildProjectTicketsResult, error) {
	if err := validateRebuildProjectTicketsOptions(opts); err != nil {
		return RebuildProjectTicketsResult{}, err
	}

	files, err := r.discoverer.DiscoverTicketFiles(ctx, DiscoverTicketFilesOptions{
		ProjectDir: opts.ProjectDir,
	})
	if err != nil {
		return RebuildProjectTicketsResult{}, err
	}

	result := RebuildProjectTicketsResult{}
	indexedAt := opts.IndexedAt.UTC()
	if indexedAt.IsZero() {
		indexedAt = r.now().UTC()
	}

	if err := r.index.DeleteProjectTickets(ctx, opts.ProjectID); err != nil {
		return RebuildProjectTicketsResult{}, err
	}

	for _, file := range files {
		ticket, err := r.parser.ParseTicket(ctx, ParseTicketOptions{
			ProjectID:    opts.ProjectID,
			ProjectDir:   opts.ProjectDir,
			Path:         file.Path,
			IndexedAt:    indexedAt,
			RelativeBase: ticketsDirectory(opts.ProjectDir),
		})
		if err != nil {
			result.ParseErrors = append(result.ParseErrors, TicketIndexRebuildParseError{
				Path:         file.Path,
				RelativePath: file.RelativePath,
				Err:          err,
			})

			continue
		}

		if err := r.index.UpsertTicket(ctx, ticket); err != nil {
			return RebuildProjectTicketsResult{}, err
		}

		result.IndexedTickets++
	}

	return result, nil
}

func validateRebuildProjectTicketsOptions(opts RebuildProjectTicketsOptions) error {
	if opts.ProjectID == "" {
		return fmt.Errorf("project id is required")
	}

	if opts.ProjectDir == "" {
		return fmt.Errorf("project directory is required")
	}

	return nil
}
