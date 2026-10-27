package ticket

import (
	"context"
	"fmt"
	"time"
)

// IndexStatusChecker defines a package boundary interface.
type IndexStatusChecker interface {
	CheckIndex(ctx context.Context, opts CheckIndexOptions) (IndexStatus, error)
}

// CheckIndexOptions describes a package boundary value.
type CheckIndexOptions struct {
	ProjectID  string
	ProjectDir string
	IndexedAt  time.Time
}

// IndexStatus describes a package boundary value.
type IndexStatus struct {
	Missing  []IndexStatusFile
	Stale    []IndexStatusTicket
	Orphaned []Indexed
	Invalid  []RebuildParseError
}

// IndexStatusFile describes a package boundary value.
type IndexStatusFile struct {
	Path         string
	RelativePath string
}

// IndexStatusTicket describes a package boundary value.
type IndexStatusTicket struct {
	Indexed Indexed
	Parsed  Indexed
}

type indexStatusChecker struct {
	discoverer Discoverer
	parser     Parser
	index      Index
	now        func() time.Time
}

// IndexStatusOption defines a package boundary type.
type IndexStatusOption func(*indexStatusChecker)

// NewIndexStatusChecker constructs the corresponding package service or value.
func NewIndexStatusChecker(index Index, opts ...IndexStatusOption) IndexStatusChecker {
	checker := &indexStatusChecker{
		discoverer: NewDiscoverer(),
		parser:     NewParser(),
		index:      index,
		now:        time.Now,
	}

	for _, opt := range opts {
		opt(checker)
	}

	return checker
}

// WithIndexStatusClock executes the corresponding package operation.
func WithIndexStatusClock(now func() time.Time) IndexStatusOption {
	return func(checker *indexStatusChecker) {
		if now != nil {
			checker.now = now
		}
	}
}

// CheckIndex executes the corresponding package operation.
func (c *indexStatusChecker) CheckIndex(ctx context.Context, opts CheckIndexOptions) (IndexStatus, error) {
	err := validateCheckProjectTicketIndexOptions(opts)
	if err != nil {
		return IndexStatus{}, err
	}

	files, err := c.discoverer.Discover(ctx, DiscoverOptions{
		ProjectDir: opts.ProjectDir,
	})
	if err != nil {
		return IndexStatus{}, err
	}

	indexedTickets, err := c.index.ListProject(ctx, opts.ProjectID)
	if err != nil {
		return IndexStatus{}, err
	}

	return c.compareProjectTicketIndex(ctx, opts, files, indexedTickets), nil
}

func (c *indexStatusChecker) compareProjectTicketIndex(ctx context.Context, opts CheckIndexOptions, files []DiscoveredFile, indexedTickets []Indexed) IndexStatus {
	status := IndexStatus{}
	indexedByRelativePath := indexedTicketsByRelativePath(indexedTickets)
	seenRelativePaths := map[string]struct{}{}
	indexedAt := c.indexedAt(opts)

	for _, file := range files {
		seenRelativePaths[file.RelativePath] = struct{}{}

		parsed, err := c.parser.Parse(ctx, ParseOptions{
			ProjectID:    opts.ProjectID,
			ProjectDir:   opts.ProjectDir,
			Path:         file.Path,
			IndexedAt:    indexedAt,
			RelativeBase: Directory(opts.ProjectDir),
		})
		if err != nil {
			status.Invalid = append(status.Invalid, RebuildParseError{
				Path:         file.Path,
				RelativePath: file.RelativePath,
				Err:          err,
			})

			continue
		}

		indexed, found := indexedByRelativePath[file.RelativePath]
		if !found {
			status.Missing = append(status.Missing, IndexStatusFile(file))

			continue
		}

		if ticketFileStateChanged(indexed, parsed) {
			status.Stale = append(status.Stale, IndexStatusTicket{
				Indexed: indexed,
				Parsed:  parsed,
			})
		}
	}

	status.Orphaned = orphanedIndexedTickets(indexedTickets, seenRelativePaths)

	return status
}

func (c *indexStatusChecker) indexedAt(opts CheckIndexOptions) time.Time {
	indexedAt := opts.IndexedAt.UTC()
	if indexedAt.IsZero() {
		indexedAt = c.now().UTC()
	}

	return indexedAt
}

func validateCheckProjectTicketIndexOptions(opts CheckIndexOptions) error {
	if opts.ProjectID == "" {
		return fmt.Errorf("project id is required")
	}

	if opts.ProjectDir == "" {
		return fmt.Errorf("project directory is required")
	}

	return nil
}

func indexedTicketsByRelativePath(tickets []Indexed) map[string]Indexed {
	indexed := make(map[string]Indexed, len(tickets))

	for _, ticket := range tickets {
		indexed[ticket.RelativePath] = ticket
	}

	return indexed
}

func ticketFileStateChanged(indexed, parsed Indexed) bool {
	return indexed.FileSize != parsed.FileSize ||
		indexed.ContentHash != parsed.ContentHash
}

func orphanedIndexedTickets(tickets []Indexed, seenRelativePaths map[string]struct{}) []Indexed {
	orphaned := []Indexed{}

	for _, ticket := range tickets {
		_, found := seenRelativePaths[ticket.RelativePath]
		if !found {
			orphaned = append(orphaned, ticket)
		}
	}

	return orphaned
}
