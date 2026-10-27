package ticket

import (
	"context"
	"fmt"
	"time"
)

// Rebuilder defines a package boundary interface.
type Rebuilder interface {
	Rebuild(ctx context.Context, opts RebuildOptions) (RebuildResult, error)
}

// RebuildOptions describes a package boundary value.
type RebuildOptions struct {
	ProjectID  string
	ProjectDir string
	IndexedAt  time.Time
}

// RebuildResult describes a package boundary value.
type RebuildResult struct {
	IndexedTickets int
	ParseErrors    []RebuildParseError
}

// RebuildParseError describes a package boundary value.
type RebuildParseError struct {
	Path         string
	RelativePath string
	Err          error
}

type rebuilder struct {
	discoverer Discoverer
	parser     Parser
	index      Index
	now        func() time.Time
}

// RebuildOption defines a package boundary type.
type RebuildOption func(*rebuilder)

// NewRebuilder constructs the corresponding package service or value.
func NewRebuilder(index Index, opts ...RebuildOption) Rebuilder {
	rebuilder := &rebuilder{
		discoverer: NewDiscoverer(),
		parser:     NewParser(),
		index:      index,
		now:        time.Now,
	}

	for _, opt := range opts {
		opt(rebuilder)
	}

	return rebuilder
}

// WithRebuildDiscoverer executes the corresponding package operation.
func WithRebuildDiscoverer(discoverer Discoverer) RebuildOption {
	return func(rebuilder *rebuilder) {
		if discoverer != nil {
			rebuilder.discoverer = discoverer
		}
	}
}

// WithRebuildParser executes the corresponding package operation.
func WithRebuildParser(parser Parser) RebuildOption {
	return func(rebuilder *rebuilder) {
		if parser != nil {
			rebuilder.parser = parser
		}
	}
}

// WithRebuildClock executes the corresponding package operation.
func WithRebuildClock(now func() time.Time) RebuildOption {
	return func(rebuilder *rebuilder) {
		if now != nil {
			rebuilder.now = now
		}
	}
}

// Rebuild executes the corresponding package operation.
func (r *rebuilder) Rebuild(ctx context.Context, opts RebuildOptions) (RebuildResult, error) {
	err := validateRebuildProjectTicketsOptions(opts)
	if err != nil {
		return RebuildResult{}, err
	}

	files, err := r.discoverer.Discover(ctx, DiscoverOptions{
		ProjectDir: opts.ProjectDir,
	})
	if err != nil {
		return RebuildResult{}, err
	}

	result := RebuildResult{}

	indexedAt := opts.IndexedAt.UTC()
	if indexedAt.IsZero() {
		indexedAt = r.now().UTC()
	}

	err = r.index.DeleteProject(ctx, opts.ProjectID)
	if err != nil {
		return RebuildResult{}, err
	}

	for _, file := range files {
		ticket, err := r.parser.Parse(ctx, ParseOptions{
			ProjectID:    opts.ProjectID,
			ProjectDir:   opts.ProjectDir,
			Path:         file.Path,
			IndexedAt:    indexedAt,
			RelativeBase: Directory(opts.ProjectDir),
		})
		if err != nil {
			result.ParseErrors = append(result.ParseErrors, RebuildParseError{
				Path:         file.Path,
				RelativePath: file.RelativePath,
				Err:          err,
			})

			continue
		}

		err = r.index.Upsert(ctx, ticket)
		if err != nil {
			return RebuildResult{}, err
		}

		result.IndexedTickets++
	}

	return result, nil
}

func validateRebuildProjectTicketsOptions(opts RebuildOptions) error {
	if opts.ProjectID == "" {
		return fmt.Errorf("project id is required")
	}

	if opts.ProjectDir == "" {
		return fmt.Errorf("project directory is required")
	}

	return nil
}
