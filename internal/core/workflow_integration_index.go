package core

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"forge.adrianpk.com/chr/internal/core/database"
)

// WorkflowIntegrationIndex extends the existing SQLite projection. Its rows are
// disposable; every query revalidates the authoritative repository snapshot.
type WorkflowIntegrationIndex interface {
	RebuildWorkflowIntegrationIndex(context.Context, *WorkflowStore) error
	QueryWorkflowIntegrationQueue(context.Context, *WorkflowStore, WorkflowIntegrationQueueQuery) (WorkflowIntegrationQueuePage, error)
}

// RebuildWorkflowIntegrationIndex replaces derived rows after validating records.
// Failed or interrupted cache work never repairs domain or publication journals.
func (i *sqliteTicketIndex) RebuildWorkflowIntegrationIndex(ctx context.Context, store *WorkflowStore) error {
	err := store.withRepositoryLock(ctx, store.requireNoWorkflowTransition, func() error {
		snapshot, err := store.readWorkflowIntegrationSnapshot(ctx)
		if err != nil {
			return err
		}

		coordinator, err := store.checkWorkflowCoordinator(ctx)
		if err != nil {
			return err
		}

		return i.replaceWorkflowIntegrationIndex(ctx, store.projectID, coordinator.ID, snapshot)
	})

	return integrationRecordError("integration index rebuild", err)
}

func (i *sqliteTicketIndex) replaceWorkflowIntegrationIndex(ctx context.Context, projectID, coordinatorID string, snapshot workflowIntegrationSnapshot) error {
	tx, err := i.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	for _, table := range []string{"workflow_integration_entries", "workflow_integration_snapshots"} {
		if _, err := tx.ExecContext(ctx, "DELETE FROM "+table+" WHERE project_id = ? AND coordinator_id = ?", projectID, coordinatorID); err != nil {
			return database.RollbackOnError(tx, err)
		}
	}

	for _, view := range snapshot.Views {
		if err := indexWorkflowIntegrationView(ctx, tx, view); err != nil {
			return database.RollbackOnError(tx, err)
		}
	}

	_, err = tx.ExecContext(ctx, "INSERT INTO workflow_integration_snapshots VALUES(?, ?, ?)", projectID, coordinatorID, snapshot.Hash)
	if err != nil {
		return database.RollbackOnError(tx, err)
	}

	return tx.Commit()
}

func workflowIntegrationViewHash(view WorkflowIntegrationView) (string, error) {
	data, err := json.Marshal(view)

	return workflowHash(data), err
}

func indexWorkflowIntegrationView(ctx context.Context, tx *sql.Tx, view WorkflowIntegrationView) error {
	hash, err := workflowIntegrationViewHash(view)
	if err != nil {
		return err
	}

	e := view.Entry
	// Decimal padding preserves unsigned ordering without SQLite signed overflow.
	sequence := fmt.Sprintf("%020d", e.QueueSequence)
	_, err = tx.ExecContext(ctx, "INSERT INTO workflow_integration_entries VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", e.ProjectID, e.CoordinatorID, e.Namespace, e.ID, e.InvocationID, e.Unit.Kind, e.Unit.ID, e.Repository, e.TargetRef, e.State, sequence, hash)

	return err
}

// QueryWorkflowIntegrationQueue rejects stale, missing or corrupted projection
// rows before returning the same domain page as the source-only query.
func (i *sqliteTicketIndex) QueryWorkflowIntegrationQueue(ctx context.Context, store *WorkflowStore, q WorkflowIntegrationQueueQuery) (page WorkflowIntegrationQueuePage, err error) {
	if err := validateWorkflowIntegrationQueueQuery(q); err != nil {
		return page, err
	}

	err = store.withRepositoryLock(ctx, store.requireNoWorkflowTransition, func() error {
		snapshot, readErr := store.readWorkflowIntegrationSnapshot(ctx)
		if readErr != nil {
			return readErr
		}

		coordinator, readErr := store.checkWorkflowCoordinator(ctx)
		if readErr != nil {
			return readErr
		}

		if readErr := i.checkWorkflowIntegrationIndex(ctx, store.projectID, coordinator.ID, snapshot); readErr != nil {
			return readErr
		}

		page, readErr = workflowIntegrationQueuePage(snapshot, q)

		return readErr
	})

	return page, integrationRecordError("integration index", err)
}

func (i *sqliteTicketIndex) checkWorkflowIntegrationIndex(ctx context.Context, projectID, coordinatorID string, snapshot workflowIntegrationSnapshot) (err error) {
	var hash string

	query := "SELECT source_hash FROM workflow_integration_snapshots WHERE project_id = ? AND coordinator_id = ?"
	if err := i.db.QueryRowContext(ctx, query, projectID, coordinatorID).Scan(&hash); err != nil {
		return err
	}

	if hash != snapshot.Hash {
		return ErrWorkflowConflict
	}

	rows, err := i.db.QueryContext(ctx, "SELECT entry_id, image_hash FROM workflow_integration_entries WHERE project_id = ? AND coordinator_id = ? ORDER BY queue_sequence, entry_id", projectID, coordinatorID)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, rows.Close()) }()

	return checkWorkflowIntegrationIndexRows(rows, snapshot.Views)
}

func checkWorkflowIntegrationIndexRows(rows *sql.Rows, views []WorkflowIntegrationView) error {
	count := 0

	for rows.Next() {
		var id, hash string
		if err := rows.Scan(&id, &hash); err != nil {
			return err
		}

		if count >= len(views) || id != views[count].Entry.ID {
			return ErrWorkflowConflict
		}

		expected, err := workflowIntegrationViewHash(views[count])
		if err != nil || hash != expected {
			return errorsOrConflict(err)
		}

		count++
	}

	if count != len(views) {
		return ErrWorkflowConflict
	}

	return rows.Err()
}
