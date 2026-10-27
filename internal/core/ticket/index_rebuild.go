package ticket

import (
	"context"
	"fmt"
	"time"
)

// TicketIndexRebuilder defines a package boundary interface.
type TicketIndexRebuilder interface {
	RebuildProjectTickets(ctx context.Context, opts RebuildProjectTicketsOptions) (RebuildProjectTicketsResult, error)
}

// RebuildProjectTicketsOptions describes a package boundary value.
type RebuildProjectTicketsOptions struct {
	ProjectID  string
	ProjectDir string
	IndexedAt  time.Time
}

// RebuildProjectTicketsResult describes a package boundary value.
type RebuildProjectTicketsResult struct {
	IndexedTickets int
	ParseErrors    []TicketIndexRebuildParseError
}

// TicketIndexRebuildParseError describes a package boundary value.
type TicketIndexRebuildParseError struct {
	Path         string
	RelativePath string
	Err          error
}

type ticketIndexRebuilder struct {
	discoverer TicketFileDiscoverer
	parser     Parser
	index      TicketIndex
	now        func() time.Time
}

// TicketIndexRebuilderOption defines a package boundary type.
type TicketIndexRebuilderOption func(*ticketIndexRebuilder)

// NewTicketIndexRebuilder constructs the corresponding package service or value.
func NewTicketIndexRebuilder(index TicketIndex, opts ...TicketIndexRebuilderOption) TicketIndexRebuilder {
	rebuilder := &ticketIndexRebuilder{
		discoverer: NewTicketFileDiscoverer(),
		parser:     NewParser(),
		index:      index,
		now:        time.Now,
	}

	for _, opt := range opts {
		opt(rebuilder)
	}

	return rebuilder
}

// WithTicketIndexRebuilderDiscoverer executes the corresponding package operation.
func WithTicketIndexRebuilderDiscoverer(discoverer TicketFileDiscoverer) TicketIndexRebuilderOption {
	return func(rebuilder *ticketIndexRebuilder) {
		if discoverer != nil {
			rebuilder.discoverer = discoverer
		}
	}
}

// WithTicketIndexRebuilderParser executes the corresponding package operation.
func WithTicketIndexRebuilderParser(parser Parser) TicketIndexRebuilderOption {
	return func(rebuilder *ticketIndexRebuilder) {
		if parser != nil {
			rebuilder.parser = parser
		}
	}
}

// WithTicketIndexRebuilderClock executes the corresponding package operation.
func WithTicketIndexRebuilderClock(now func() time.Time) TicketIndexRebuilderOption {
	return func(rebuilder *ticketIndexRebuilder) {
		if now != nil {
			rebuilder.now = now
		}
	}
}

// RebuildProjectTickets executes the corresponding package operation.
func (r *ticketIndexRebuilder) RebuildProjectTickets(ctx context.Context, opts RebuildProjectTicketsOptions) (RebuildProjectTicketsResult, error) {
	err := validateRebuildProjectTicketsOptions(opts)
	if err != nil {
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

	err = r.index.DeleteProjectTickets(ctx, opts.ProjectID)
	if err != nil {
		return RebuildProjectTicketsResult{}, err
	}

	for _, file := range files {
		ticket, err := r.parser.Parse(ctx, ParseOptions{
			ProjectID:    opts.ProjectID,
			ProjectDir:   opts.ProjectDir,
			Path:         file.Path,
			IndexedAt:    indexedAt,
			RelativeBase: TicketsDirectory(opts.ProjectDir),
		})
		if err != nil {
			result.ParseErrors = append(result.ParseErrors, TicketIndexRebuildParseError{
				Path:         file.Path,
				RelativePath: file.RelativePath,
				Err:          err,
			})

			continue
		}

		err = r.index.UpsertTicket(ctx, ticket)
		if err != nil {
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
