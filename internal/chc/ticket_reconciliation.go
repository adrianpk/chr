package chc

import (
	"context"
	"fmt"
	"path/filepath"
	"time"
)

type TicketReconciler interface {
	ReconcileProjectTickets(ctx context.Context, opts ReconcileProjectTicketsOptions) (TicketReconciliationResult, error)
	ReconcileProjectTicket(ctx context.Context, opts ReconcileProjectTicketOptions) (TicketReconciliationResult, error)
}

type ReconcileProjectTicketsOptions struct {
	ProjectID    string
	ProjectDir   string
	ReconciledAt time.Time
}

type ReconcileProjectTicketOptions struct {
	ProjectID    string
	ProjectDir   string
	TicketPath   string
	ReconciledAt time.Time
}

type TicketReconciliationResult struct {
	IndexedTickets int
	Blockers       []TicketReconciliationBlocker
}

type TicketReconciliationBlockerKind string

const TicketReconciliationBlockerParseError TicketReconciliationBlockerKind = "parse_error"

type TicketReconciliationBlocker struct {
	Kind         TicketReconciliationBlockerKind
	Path         string
	RelativePath string
	Err          error
}

func (r TicketReconciliationResult) HasBlockers() bool {
	return len(r.Blockers) > 0
}

func (r TicketReconciliationResult) FirstBlocker() (TicketReconciliationBlocker, bool) {
	if len(r.Blockers) == 0 {
		return TicketReconciliationBlocker{}, false
	}

	return r.Blockers[0], true
}

type ticketReconciler struct {
	rebuilder TicketIndexRebuilder
	parser    TicketParser
	index     TicketIndex
}

func NewTicketReconciler(index TicketIndex) TicketReconciler {
	return &ticketReconciler{
		rebuilder: NewTicketIndexRebuilder(index),
		parser:    NewTicketParser(),
		index:     index,
	}
}

func (r *ticketReconciler) ReconcileProjectTickets(
	ctx context.Context,
	opts ReconcileProjectTicketsOptions,
) (TicketReconciliationResult, error) {
	err := validateReconcileProjectTicketsOptions(opts)
	if err != nil {
		return TicketReconciliationResult{}, err
	}

	result, err := r.rebuilder.RebuildProjectTickets(ctx, RebuildProjectTicketsOptions{
		ProjectID:  opts.ProjectID,
		ProjectDir: opts.ProjectDir,
		IndexedAt:  opts.ReconciledAt,
	})
	if err != nil {
		return TicketReconciliationResult{}, err
	}

	return TicketReconciliationResult{
		IndexedTickets: result.IndexedTickets,
		Blockers:       ticketReconciliationBlockersFromParseErrors(result.ParseErrors),
	}, nil
}

func (r *ticketReconciler) ReconcileProjectTicket(
	ctx context.Context,
	opts ReconcileProjectTicketOptions,
) (TicketReconciliationResult, error) {
	err := validateReconcileProjectTicketOptions(opts)
	if err != nil {
		return TicketReconciliationResult{}, err
	}

	indexedAt := opts.ReconciledAt.UTC()
	if indexedAt.IsZero() {
		indexedAt = time.Now().UTC()
	}

	ticket, blocker, blocked := r.parseProjectTicket(ctx, opts, indexedAt)
	if blocked {
		return TicketReconciliationResult{
			Blockers: []TicketReconciliationBlocker{blocker},
		}, nil
	}

	err = r.index.UpsertTicket(ctx, ticket)
	if err != nil {
		return TicketReconciliationResult{}, err
	}

	return TicketReconciliationResult{IndexedTickets: 1}, nil
}

func (r *ticketReconciler) parseProjectTicket(
	ctx context.Context,
	opts ReconcileProjectTicketOptions,
	indexedAt time.Time,
) (IndexedTicket, TicketReconciliationBlocker, bool) {
	ticket, err := r.parser.ParseTicket(ctx, ParseTicketOptions{
		ProjectID:    opts.ProjectID,
		ProjectDir:   opts.ProjectDir,
		Path:         opts.TicketPath,
		IndexedAt:    indexedAt,
		RelativeBase: ticketsDirectory(opts.ProjectDir),
	})
	if err != nil {
		return IndexedTicket{}, TicketReconciliationBlocker{
			Kind:         TicketReconciliationBlockerParseError,
			Path:         opts.TicketPath,
			RelativePath: relativeTicketPathForBlocker(opts.ProjectDir, opts.TicketPath),
			Err:          err,
		}, true
	}

	return ticket, TicketReconciliationBlocker{}, false
}

func validateReconcileProjectTicketsOptions(opts ReconcileProjectTicketsOptions) error {
	if opts.ProjectID == "" {
		return fmt.Errorf("project id is required")
	}

	if opts.ProjectDir == "" {
		return fmt.Errorf("project directory is required")
	}

	return nil
}

func validateReconcileProjectTicketOptions(opts ReconcileProjectTicketOptions) error {
	if opts.ProjectID == "" {
		return fmt.Errorf("project id is required")
	}

	if opts.ProjectDir == "" {
		return fmt.Errorf("project directory is required")
	}

	if opts.TicketPath == "" {
		return fmt.Errorf("ticket path is required")
	}

	return nil
}

func relativeTicketPathForBlocker(projectDir, ticketPath string) string {
	relativePath, err := filepath.Rel(ticketsDirectory(projectDir), ticketPath)
	if err != nil {
		return ticketPath
	}

	return filepath.ToSlash(relativePath)
}

func ticketReconciliationBlockersFromParseErrors(parseErrors []TicketIndexRebuildParseError) []TicketReconciliationBlocker {
	blockers := make([]TicketReconciliationBlocker, 0, len(parseErrors))

	for _, parseErr := range parseErrors {
		blockers = append(blockers, TicketReconciliationBlocker{
			Kind:         TicketReconciliationBlockerParseError,
			Path:         parseErr.Path,
			RelativePath: parseErr.RelativePath,
			Err:          parseErr.Err,
		})
	}

	return blockers
}
