package ticket

import (
	"context"
	"fmt"
	"path/filepath"
	"time"
)

// Reconciler defines a package boundary interface.
type Reconciler interface {
	ReconcileProject(ctx context.Context, opts ReconcileProjectOptions) (ReconcileResult, error)
	ReconcileTicket(ctx context.Context, opts ReconcileTicketOptions) (ReconcileResult, error)
}

// ReconcileProjectOptions describes a package boundary value.
type ReconcileProjectOptions struct {
	ProjectID    string
	ProjectDir   string
	ReconciledAt time.Time
}

// ReconcileTicketOptions describes a package boundary value.
type ReconcileTicketOptions struct {
	ProjectID    string
	ProjectDir   string
	TicketPath   string
	ReconciledAt time.Time
}

// ReconcileResult describes a package boundary value.
type ReconcileResult struct {
	IndexedTickets int
	Blockers       []ReconcileBlocker
}

// ReconcileBlockerKind defines a package boundary type.
type ReconcileBlockerKind string

// ReconcileBlockerParseError defines a package boundary constant.
const ReconcileBlockerParseError ReconcileBlockerKind = "parse_error"

// ReconcileBlocker describes a package boundary value.
type ReconcileBlocker struct {
	Kind         ReconcileBlockerKind
	Path         string
	RelativePath string
	Err          error
}

// HasBlockers executes the corresponding package operation.
func (r ReconcileResult) HasBlockers() bool {
	return len(r.Blockers) > 0
}

// FirstBlocker executes the corresponding package operation.
func (r ReconcileResult) FirstBlocker() (ReconcileBlocker, bool) {
	if len(r.Blockers) == 0 {
		return ReconcileBlocker{}, false
	}

	return r.Blockers[0], true
}

type reconciler struct {
	rebuilder Rebuilder
	parser    Parser
	index     Index
}

// NewReconciler constructs the corresponding package service or value.
func NewReconciler(index Index) Reconciler {
	return &reconciler{
		rebuilder: NewRebuilder(index),
		parser:    NewParser(),
		index:     index,
	}
}

// ReconcileProject executes the corresponding package operation.
func (r *reconciler) ReconcileProject(ctx context.Context, opts ReconcileProjectOptions) (ReconcileResult, error) {
	err := validateReconcileProjectTicketsOptions(opts)
	if err != nil {
		return ReconcileResult{}, err
	}

	result, err := r.rebuilder.Rebuild(ctx, RebuildOptions{
		ProjectID:  opts.ProjectID,
		ProjectDir: opts.ProjectDir,
		IndexedAt:  opts.ReconciledAt,
	})
	if err != nil {
		return ReconcileResult{}, err
	}

	return ReconcileResult{
		IndexedTickets: result.IndexedTickets,
		Blockers:       ticketReconciliationBlockersFromParseErrors(result.ParseErrors),
	}, nil
}

// ReconcileTicket executes the corresponding package operation.
func (r *reconciler) ReconcileTicket(ctx context.Context, opts ReconcileTicketOptions) (ReconcileResult, error) {
	err := validateReconcileProjectTicketOptions(opts)
	if err != nil {
		return ReconcileResult{}, err
	}

	indexedAt := opts.ReconciledAt.UTC()
	if indexedAt.IsZero() {
		indexedAt = time.Now().UTC()
	}

	parsed := r.parseProjectTicket(ctx, opts, indexedAt)
	if parsed.Blocker != nil {
		return ReconcileResult{
			Blockers: []ReconcileBlocker{*parsed.Blocker},
		}, nil
	}

	err = r.index.Upsert(ctx, parsed.Ticket)
	if err != nil {
		return ReconcileResult{}, err
	}

	return ReconcileResult{IndexedTickets: 1}, nil
}

type reconcileParseResult struct {
	Ticket  Indexed
	Blocker *ReconcileBlocker
}

func (r *reconciler) parseProjectTicket(ctx context.Context, opts ReconcileTicketOptions, indexedAt time.Time) reconcileParseResult {
	ticket, err := r.parser.Parse(ctx, ParseOptions{
		ProjectID:    opts.ProjectID,
		ProjectDir:   opts.ProjectDir,
		Path:         opts.TicketPath,
		IndexedAt:    indexedAt,
		RelativeBase: Directory(opts.ProjectDir),
	})
	if err != nil {
		return reconcileParseResult{Blocker: &ReconcileBlocker{
			Kind:         ReconcileBlockerParseError,
			Path:         opts.TicketPath,
			RelativePath: relativeTicketPathForBlocker(opts.ProjectDir, opts.TicketPath),
			Err:          err,
		}}
	}

	return reconcileParseResult{Ticket: ticket}
}

func validateReconcileProjectTicketsOptions(opts ReconcileProjectOptions) error {
	if opts.ProjectID == "" {
		return fmt.Errorf("project id is required")
	}

	if opts.ProjectDir == "" {
		return fmt.Errorf("project directory is required")
	}

	return nil
}

func validateReconcileProjectTicketOptions(opts ReconcileTicketOptions) error {
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
	relativePath, err := filepath.Rel(Directory(projectDir), ticketPath)
	if err != nil {
		return ticketPath
	}

	return filepath.ToSlash(relativePath)
}

func ticketReconciliationBlockersFromParseErrors(parseErrors []RebuildParseError) []ReconcileBlocker {
	blockers := make([]ReconcileBlocker, 0, len(parseErrors))

	for _, parseErr := range parseErrors {
		blockers = append(blockers, ReconcileBlocker{
			Kind:         ReconcileBlockerParseError,
			Path:         parseErr.Path,
			RelativePath: parseErr.RelativePath,
			Err:          parseErr.Err,
		})
	}

	return blockers
}
