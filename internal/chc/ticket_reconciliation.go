package chc

import (
	"context"
	"fmt"
	"time"
)

type TicketReconciler interface {
	ReconcileProjectTickets(ctx context.Context, opts ReconcileProjectTicketsOptions) (TicketReconciliationResult, error)
}

type ReconcileProjectTicketsOptions struct {
	ProjectID    string
	ProjectDir   string
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
}

func NewTicketReconciler(index TicketIndex) TicketReconciler {
	return &ticketReconciler{
		rebuilder: NewTicketIndexRebuilder(index),
	}
}

func (r *ticketReconciler) ReconcileProjectTickets(
	ctx context.Context,
	opts ReconcileProjectTicketsOptions,
) (TicketReconciliationResult, error) {
	if err := validateReconcileProjectTicketsOptions(opts); err != nil {
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

func validateReconcileProjectTicketsOptions(opts ReconcileProjectTicketsOptions) error {
	if opts.ProjectID == "" {
		return fmt.Errorf("project id is required")
	}

	if opts.ProjectDir == "" {
		return fmt.Errorf("project directory is required")
	}

	return nil
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
