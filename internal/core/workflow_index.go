package core

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	"forge.adrianpk.com/chr/internal/core/database"
	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// WorkflowIndex is implemented by the existing SQLite ticket index.
type WorkflowIndex interface {
	RebuildWorkflowIndex(context.Context, *WorkflowStore) error
	QueryWorkflow(context.Context, *WorkflowStore, WorkflowQuery) (WorkflowPage, error)
}

// WorkflowQuery selects a bounded page from one exact shared snapshot.
type WorkflowQuery struct {
	Kind         string
	OwnerID      string
	UseCase      string
	LooseOnly    bool
	Offset       int
	Limit        int
	SnapshotHash string
}

// WorkflowPage returns domain records rather than SQL-owned lifecycle state.
type WorkflowPage struct {
	Records      []WorkflowRecord
	Slices       []workflowstate.WorkflowSlice
	SnapshotHash string
	NextOffset   int
}

// RebuildWorkflowIndex replaces only the requested namespace's derived rows.
func (i *workflowTicketIndex) RebuildWorkflowIndex(ctx context.Context, store *WorkflowStore) error {
	return store.withRepositoryLock(ctx, store.requireNoWorkflowTransition, func() error {
		snapshot, err := store.readSnapshot(ctx)
		if err != nil {
			return err
		}

		tx, err := i.db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}

		if err := clearWorkflowIndex(ctx, tx, store); err != nil {
			return database.RollbackOnError(tx, err)
		}

		sort.SliceStable(snapshot.Records, func(a, b int) bool { return snapshot.Records[a].Set != nil && snapshot.Records[b].Set == nil })

		for _, r := range snapshot.Records {
			if err := indexWorkflowRecord(ctx, tx, store, r); err != nil {
				return database.RollbackOnError(tx, err)
			}
		}

		_, err = tx.ExecContext(ctx, "INSERT INTO workflow_snapshots(project_id, namespace, source_hash) VALUES(?, ?, ?)", store.projectID, store.layout.Namespace(), workflowSnapshotHash(snapshot))
		if err != nil {
			return database.RollbackOnError(tx, err)
		}

		return tx.Commit()
	})
}

func clearWorkflowIndex(ctx context.Context, tx *sql.Tx, store *WorkflowStore) error {
	queries := []string{
		"DELETE FROM indexed_ticket_tags WHERE project_id = ? AND ticket_path IN (SELECT path FROM indexed_tickets WHERE namespace = ? AND workflow_revision > 0)",
		"DELETE FROM indexed_tickets WHERE project_id = ? AND namespace = ? AND workflow_revision > 0",
		"DELETE FROM workflow_slices WHERE project_id = ? AND namespace = ?",
		"DELETE FROM workflow_sets WHERE project_id = ? AND namespace = ?",
		"DELETE FROM workflow_receipts WHERE project_id = ? AND namespace = ?",
		"DELETE FROM workflow_snapshots WHERE project_id = ? AND namespace = ?",
	}
	for _, query := range queries {
		if _, err := tx.ExecContext(ctx, query, store.projectID, store.layout.Namespace()); err != nil {
			return err
		}
	}

	return nil
}

func indexWorkflowRecord(ctx context.Context, tx *sql.Tx, store *WorkflowStore, r WorkflowRecord) error {
	if r.Task != nil {
		return indexWorkflowTask(ctx, tx, store, r)
	}

	if r.Set != nil {
		return indexWorkflowSet(ctx, tx, r)
	}

	receipt := r.Receipt
	_, err := tx.ExecContext(ctx, "INSERT INTO workflow_receipts VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", r.ProjectID, r.Namespace, r.ID, r.Path, r.Revision, r.SourceHash, receipt.CandidateID, receipt.SubmissionID, receipt.Unit.Kind, receipt.Unit.ID, receipt.PullRequest, receipt.ConfirmedAt.Format(time.RFC3339Nano))

	return err
}

func indexWorkflowSet(ctx context.Context, tx *sql.Tx, r WorkflowRecord) error {
	set := r.Set

	_, err := tx.ExecContext(ctx, "INSERT INTO workflow_sets VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", r.ProjectID, r.Namespace, r.ID, r.Path, r.Revision, r.SourceHash, set.UseCase, set.UseCaseOrder, r.CreatedAt.Format(time.RFC3339Nano), set.Status, set.ClosureRef)
	if err != nil {
		return err
	}

	for _, slice := range set.Slices {
		_, err := tx.ExecContext(ctx, "INSERT INTO workflow_slices VALUES(?, ?, ?, ?, ?, ?, ?, ?)", r.ProjectID, r.Namespace, slice.ID, r.ID, slice.Order, slice.Phase, slice.PullRequest, slice.DeliveryRef)
		if err != nil {
			return err
		}
	}

	return nil
}

func indexWorkflowTask(ctx context.Context, tx *sql.Tx, store *WorkflowStore, r WorkflowRecord) error {
	ticket, err := workflowIndexedTicket(ctx, store, r)
	if err != nil {
		return err
	}

	if r.Task.SetID != "" && ticket.UseCase == "" {
		if err := tx.QueryRowContext(ctx, "SELECT use_case FROM workflow_sets WHERE project_id = ? AND namespace = ? AND id = ?", r.ProjectID, r.Namespace, r.Task.SetID).Scan(&ticket.UseCase); err != nil {
			return err
		}
	}

	if err := ticketcore.UpsertInTransaction(ctx, tx, ticket); err != nil {
		return err
	}

	task := r.Task
	_, err = tx.ExecContext(ctx, `UPDATE indexed_tickets SET workflow_schema = ?, workflow_revision = ?, representation_ref = ?, body_ref = ?, execution_contract_hash = ?, slice_set_id = ?, slice_id = ?, task_order = ?, task_phase = ?, slice_delivery_ref = ? WHERE project_id = ? AND path = ?`, task.Schema, r.Revision, r.Path, ticket.RelativePath, task.ContractHash, task.SetID, task.SliceID, task.Order, task.Phase, task.DeliveryRef, r.ProjectID, ticket.Path)

	return err
}

func workflowIndexedTicket(ctx context.Context, store *WorkflowStore, r WorkflowRecord) (ticketcore.Indexed, error) {
	if strings.HasSuffix(r.Path, ".md") {
		parsed, err := ticketcore.NewParser().Parse(ctx, ticketcore.ParseOptions{ProjectID: r.ProjectID, ProjectDir: store.layout.ProjectDir(), Path: filepath.Join(store.layout.Root(), r.Path), RelativeBase: store.layout.Root(), IndexedAt: time.Now().UTC()})
		parsed.Namespace = r.Namespace

		return parsed, err
	}

	if strings.HasPrefix(string(r.Body), "---\n") {
		return workflowCanonicalTicket(ctx, store, r)
	}

	body := filepath.Join(filepath.Dir(r.Path), "artifact.md")

	display, err := workflowPlainTicketDisplay(r)
	if err != nil {
		return ticketcore.Indexed{}, err
	}

	display.ProjectID, display.Namespace, display.Key = r.ProjectID, r.Namespace, r.Task.Key
	display.Status, display.UseCase = ticketcore.Status(r.Task.Status), r.Task.UseCase
	display.Path, display.RelativePath = filepath.Join(store.layout.Root(), body), body
	display.CreatedAt, display.UpdatedAt, display.FileMTime = &r.CreatedAt, &r.UpdatedAt, r.UpdatedAt
	display.FileSize, display.ContentHash, display.IndexedAt = int64(len(r.Body)), r.SourceHash, time.Now().UTC()

	return display, nil
}

func workflowCanonicalTicket(ctx context.Context, store *WorkflowStore, r WorkflowRecord) (ticketcore.Indexed, error) {
	relative := filepath.Join(filepath.Dir(r.Path), "artifact.md")

	parsed, err := ticketcore.NewParser().Parse(ctx, ticketcore.ParseOptions{ProjectID: r.ProjectID, ProjectDir: store.layout.ProjectDir(), Path: filepath.Join(store.layout.Root(), relative), RelativeBase: store.layout.Root(), IndexedAt: time.Now().UTC()})
	if err != nil {
		return ticketcore.Indexed{}, err
	}

	parsed.Namespace, parsed.Key, parsed.Status = r.Namespace, r.Task.Key, ticketcore.Status(r.Task.Status)
	parsed.UseCase, parsed.ContentHash = r.Task.UseCase, r.SourceHash

	return parsed, nil
}

func workflowSnapshotHash(snapshot WorkflowSnapshot) string {
	images := []string{}
	for _, r := range snapshot.Records {
		images = append(images, r.Path+"\x00"+r.SourceHash)
	}

	sort.Strings(images)

	return contenthash.Sum([]byte(strings.Join(images, "\n")))
}

// QueryWorkflow rejects stale SQL or pagination snapshots before returning work.
func (i *workflowTicketIndex) QueryWorkflow(ctx context.Context, store *WorkflowStore, q WorkflowQuery) (page WorkflowPage, err error) {
	if q.Limit < 1 || q.Limit > workflowstate.PageLimit || q.Offset < 0 {
		return WorkflowPage{}, fmt.Errorf("workflow query page bounds exceeded")
	}

	err = store.withRepositoryLock(ctx, store.requireNoWorkflowTransition, func() error {
		snapshot, err := store.readSnapshot(ctx)
		if err != nil {
			return err
		}

		hash := workflowSnapshotHash(snapshot)

		var indexedHash string
		if err := i.db.QueryRowContext(ctx, "SELECT source_hash FROM workflow_snapshots WHERE project_id = ? AND namespace = ?", store.projectID, store.layout.Namespace()).Scan(&indexedHash); err != nil {
			return err
		}

		if indexedHash != hash || (q.SnapshotHash != "" && q.SnapshotHash != hash) {
			return fmt.Errorf("%w: rebuild workflow index or restart pagination", workflowstate.ErrConflict)
		}

		page, err = i.queryWorkflowPage(ctx, store, snapshot, q)
		page.SnapshotHash = hash

		return err
	})

	return page, err
}
